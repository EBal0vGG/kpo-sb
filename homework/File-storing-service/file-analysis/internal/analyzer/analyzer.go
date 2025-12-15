package analyzer

import (
	"context"
	"fmt"
	"strings"
)

// Analyzer performs plagiarism detection.
type Analyzer struct {
	threshold float64
}

func New(threshold float64) *Analyzer {
	return &Analyzer{threshold: threshold}
}

// CheckPlagiarism compares file hash and computes similarity score.
// Returns plagiarism flag and similarity score (0.0 to 1.0).
func (a *Analyzer) CheckPlagiarism(ctx context.Context, hash string, existingHashes []string, text string, existingTexts []string) (bool, float64, error) {
	// Exact hash match = 100% plagiarism
	for _, h := range existingHashes {
		if h == hash {
			return true, 1.0, nil
		}
	}

	// Text similarity using Jaccard similarity on shingles
	if len(existingTexts) == 0 {
		return false, 0.0, nil
	}

	maxSim := 0.0
	for _, existing := range existingTexts {
		sim := jaccardSimilarity(text, existing)
		if sim > maxSim {
			maxSim = sim
		}
	}

	isPlagiarism := maxSim >= a.threshold
	return isPlagiarism, maxSim, nil
}

// jaccardSimilarity computes Jaccard similarity between two texts using 5-shingles.
func jaccardSimilarity(text1, text2 string) float64 {
	shingles1 := makeShingles(text1, 5)
	shingles2 := makeShingles(text2, 5)

	if len(shingles1) == 0 && len(shingles2) == 0 {
		return 1.0
	}
	if len(shingles1) == 0 || len(shingles2) == 0 {
		return 0.0
	}

	intersection := 0
	union := make(map[string]bool)

	for s := range shingles1 {
		union[s] = true
		if shingles2[s] {
			intersection++
		}
	}
	for s := range shingles2 {
		union[s] = true
	}

	if len(union) == 0 {
		return 0.0
	}

	return float64(intersection) / float64(len(union))
}

// makeShingles creates k-shingles (character n-grams) from text.
func makeShingles(text string, k int) map[string]bool {
	text = strings.ToLower(strings.TrimSpace(text))
	if len(text) < k {
		return map[string]bool{text: true}
	}

	shingles := make(map[string]bool)
	for i := 0; i <= len(text)-k; i++ {
		shingle := text[i : i+k]
		shingles[shingle] = true
	}
	return shingles
}

// ExtractText extracts plain text from file content (simplified for MVP).
// For TXT: return as-is.
// For DOCX: placeholder - would need docx parsing library.
func ExtractText(content []byte, mime string) (string, error) {
	switch mime {
	case "text/plain":
		return string(content), nil
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		// TODO: implement DOCX parsing (would need a library like github.com/lukasjarosch/go-docx)
		// For MVP, return empty string to skip text similarity check
		return "", fmt.Errorf("DOCX text extraction not implemented yet")
	default:
		return "", fmt.Errorf("unsupported mime type: %s", mime)
	}
}



