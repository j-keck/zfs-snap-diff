package main

import (
	"errors"
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

func TestPanicOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("panic expected")
		}
	}()

	panicOnError(errors.New("dummy error"))
}
