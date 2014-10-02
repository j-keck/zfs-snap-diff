package main

import (
	"testing"
)

func TestLastElement(t *testing.T) {
	if lastElement("a/b/c/d", "/") != "d" {
		t.Error("d expected")
	}

	if lastElement("a.b.c.d", ".") != "d" {
		t.Error("d expected")
	}
}

func TestSplitText(t *testing.T) {
	tests := []struct {
		text        string
		expectedLen int
	}{
		{"a\nb", 2},
		{"a\nb\n", 2},
	}

	for _, test := range tests {
		lines := splitText(test.text)
		linesLen := len(lines)
		if linesLen != test.expectedLen {
			t.Errorf("lines-len: %d, expected-len: %d\n'%s'", linesLen, test.expectedLen, lines)
		}
	}
}
