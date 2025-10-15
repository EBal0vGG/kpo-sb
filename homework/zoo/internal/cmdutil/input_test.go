package cmdutil

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestAskIntMin_and_Range(t *testing.T) {
	// Simulate inputs: -1 (invalid for min 0), then 5 (valid), then 11 (invalid range), then 7 (valid)
	in := bytes.NewBufferString(strings.Join([]string{"-1", "5", "11", "7"}, "\n") + "\n")
	r := bufio.NewReader(in)

	gotMin := AskIntMin(r, "Food: ", 0)
	if gotMin != 5 {
		t.Fatalf("AskIntMin => %d, want 5", gotMin)
	}

	gotRange := AskIntRange(r, "Kindness: ", 0, 10)
	if gotRange != 7 {
		t.Fatalf("AskIntRange => %d, want 7", gotRange)
	}
}
