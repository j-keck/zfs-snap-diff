package file

import (
	"testing"
	"time"
)

func TestNewComparator(t *testing.T) {
	newComparator := func(method string, fh FileHandle) Comparator {
		c, err := NewComparator(method, fh)
		if err != nil {
			t.Error(err)
			return nil
		}
		return c
	}

	if _, ok := newComparator("size", *new(FileHandle)).(*CompareBySize); !ok {
		t.Errorf("CompareBySize expected")
	}

	if _, ok := newComparator("size+modTime", *new(FileHandle)).(*CompareBySizeAndModTime); !ok {
		t.Errorf("CompareBySizeAndModTime comparator expected")
	}

	if _, ok := newComparator("content", *new(FileHandle)).(*CompareByContent); !ok {
		t.Errorf("CompareByContent comparator expected")
	}

	if _, ok := newComparator("md5", *new(FileHandle)).(*CompareByMD5); !ok {
		t.Errorf("CompareByMD5 comparator expected")
	}

	// 'auto' should return CompareByMD5 for text files
	textFilePath := "testdata/text.txt"
	textFile, err := NewFileHandle(textFilePath)
	if err != nil {
		t.Errorf("%s not found - err: %v", textFilePath, err)
		return
	}
	if _, ok := newComparator("auto", textFile).(*CompareByMD5); !ok {
		t.Errorf("CompareByMD5 comparator expected")
	}

	// 'auto' should return CompareByMD5 for binary files
	pdfFilePath := "testdata/gospec.pdf"
	pdfFile, err := NewFileHandle(pdfFilePath)
	if err != nil {
		t.Errorf("%s not found - err: %v", pdfFilePath, err)
		return
	}
	if _, ok := newComparator("auto", pdfFile).(*CompareBySizeAndModTime); !ok {
		t.Errorf("CompareBySizeAndModTime comparator expected")
	}
}

func TestCompareByMTime(t *testing.T) {

	cmp, _ := NewComparator("mtime", FileHandle{ModTime: time.Unix(10, 0)})
	// no diff to actual
	if cmp.HasChanged(FileHandle{ModTime: time.Unix(10, 0)}) {
		t.Error("wrong change detected")
	}

	// diff to actual - should trigger
	if !cmp.HasChanged(FileHandle{ModTime: time.Unix(20, 0)}) {
		t.Error("wrong change detected")
	}

	// diff to actual - should NOT trigger, same time as before
	if cmp.HasChanged(FileHandle{ModTime: time.Unix(20, 0)}) {
		t.Error("wrong change detected")
	}

	// diff to actual and before - should trigger
	if !cmp.HasChanged(FileHandle{ModTime: time.Unix(30, 0)}) {
		t.Error("wrong change detected")
	}

}
