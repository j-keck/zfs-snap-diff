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

func TestSplit2(t *testing.T) {
	a, b, ok := split2("a/b", "/")
	if !ok {
		t.Fatal("expected true for flag")
	}

	if a != "a" {
		t.Fatal("exepcted a as first element")
	}

	if b != "b" {
		t.Fatal("expected b as second element")
	}
}

func TestSplit2WithMissingElement(t *testing.T) {
	if _, _, ok := split2("aaa", "/"); ok {
		t.Fatal("expected false for flag")
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
