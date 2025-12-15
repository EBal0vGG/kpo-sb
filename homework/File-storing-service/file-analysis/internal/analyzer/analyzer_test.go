package analyzer

import (
	"context"
	"testing"
)

func TestAnalyzer_CheckPlagiarism_ExactHashMatch(t *testing.T) {
	analyzer := New(0.7)

	hash := "abc123"
	existingHashes := []string{"def456", "abc123", "ghi789"}
	text := "some text"
	existingTexts := []string{}

	isPlagiarism, score, err := analyzer.CheckPlagiarism(context.Background(), hash, existingHashes, text, existingTexts)
	if err != nil {
		t.Fatalf("CheckPlagiarism failed: %v", err)
	}

	if !isPlagiarism {
		t.Error("Expected plagiarism flag to be true for exact hash match")
	}
	if score != 1.0 {
		t.Errorf("Expected score 1.0 for exact hash match, got %f", score)
	}
}

func TestAnalyzer_CheckPlagiarism_NoMatch(t *testing.T) {
	analyzer := New(0.7)

	hash := "abc123"
	existingHashes := []string{"def456", "ghi789"}
	text := "completely different text about programming"
	existingTexts := []string{"totally unrelated content about cooking recipes"}

	isPlagiarism, score, err := analyzer.CheckPlagiarism(context.Background(), hash, existingHashes, text, existingTexts)
	if err != nil {
		t.Fatalf("CheckPlagiarism failed: %v", err)
	}

	if isPlagiarism {
		t.Error("Expected plagiarism flag to be false for no match")
	}
	if score >= 0.7 {
		t.Errorf("Expected score < 0.7, got %f", score)
	}
}

func TestAnalyzer_CheckPlagiarism_HighSimilarity(t *testing.T) {
	analyzer := New(0.7)

	hash := "abc123"
	existingHashes := []string{"def456"}
	text := "this is a test document with some content"
	existingTexts := []string{"this is a test document with some content and more"}

	_, score, err := analyzer.CheckPlagiarism(context.Background(), hash, existingHashes, text, existingTexts)
	if err != nil {
		t.Fatalf("CheckPlagiarism failed: %v", err)
	}

	// Should detect high similarity
	if score < 0.5 {
		t.Errorf("Expected high similarity score, got %f", score)
	}
}

func TestJaccardSimilarity_Identical(t *testing.T) {
	text1 := "hello world"
	text2 := "hello world"

	similarity := jaccardSimilarity(text1, text2)
	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical texts, got %f", similarity)
	}
}

func TestJaccardSimilarity_Different(t *testing.T) {
	text1 := "hello world"
	text2 := "completely different text"

	similarity := jaccardSimilarity(text1, text2)
	if similarity >= 0.5 {
		t.Errorf("Expected low similarity for different texts, got %f", similarity)
	}
}

func TestMakeShingles(t *testing.T) {
	text := "hello"
	shingles := makeShingles(text, 5)

	if len(shingles) != 1 {
		t.Errorf("Expected 1 shingle for short text, got %d", len(shingles))
	}

	text2 := "hello world"
	shingles2 := makeShingles(text2, 5)
	if len(shingles2) == 0 {
		t.Error("Expected shingles for longer text")
	}
}

func TestExtractText_TXT(t *testing.T) {
	content := []byte("plain text content")
	mime := "text/plain"

	text, err := ExtractText(content, mime)
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}

	if text != string(content) {
		t.Errorf("Expected text %s, got %s", string(content), text)
	}
}

func TestExtractText_DOCX(t *testing.T) {
	content := []byte("fake docx content")
	mime := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	_, err := ExtractText(content, mime)
	if err == nil {
		t.Error("Expected error for DOCX (not implemented)")
	}
}

func TestExtractText_Unsupported(t *testing.T) {
	content := []byte("some content")
	mime := "application/pdf"

	_, err := ExtractText(content, mime)
	if err == nil {
		t.Error("Expected error for unsupported MIME type")
	}
}

