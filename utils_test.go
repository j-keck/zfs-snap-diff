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
