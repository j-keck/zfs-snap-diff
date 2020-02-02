package scanner

import (
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"testing"
	"time"
)

func TestNewComparator(t *testing.T) {
	newComparator := func(method string, file fs.FileHandle) Comparator {
		c, err := NewComparator(method, file)
		if err != nil {
			t.Error(err)
			return nil
		}
		return c
	}

	if _, ok := newComparator("size", *new(fs.FileHandle)).(*CompareBySize); !ok {
		t.Errorf("CompareBySize expected")
	}

	if _, ok := newComparator("size+modTime", *new(fs.FileHandle)).(*CompareBySizeAndModTime); !ok {
		t.Errorf("CompareBySizeAndModTime comparator expected")
	}

	if _, ok := newComparator("content", *new(fs.FileHandle)).(*CompareByContent); !ok {
		t.Errorf("CompareByContent comparator expected")
	}

	if _, ok := newComparator("md5", *new(fs.FileHandle)).(*CompareByMD5); !ok {
		t.Errorf("CompareByMD5 comparator expected")
	}

	// 'auto' should return CompareByMD5 for text files
	textFilePath := "testdata/text.txt"
	textFile, err := fs.NewFileHandle(textFilePath)
	if err != nil {
		t.Errorf("%s not found - err: %v", textFilePath, err)
		return
	}
	if _, ok := newComparator("auto", textFile).(*CompareByMD5); !ok {
		t.Errorf("CompareByMD5 comparator expected")
	}

	// 'auto' should return CompareByMD5 for binary files
	pdfFilePath := "testdata/gospec.pdf"
	pdfFile, err := fs.NewFileHandle(pdfFilePath)
	if err != nil {
		t.Errorf("%s not found - err: %v", pdfFilePath, err)
		return
	}
	if _, ok := newComparator("auto", pdfFile).(*CompareBySizeAndModTime); !ok {
		t.Errorf("CompareBySizeAndModTime comparator expected")
	}
}

func TestCompareByMTime(t *testing.T) {
	fileHandleWithMTime := func(mtime time.Time) fs.FileHandle {
		fsHandle := fs.FSHandle{}
		fsHandle.MTime = mtime
		return fs.FileHandle{fsHandle}
	}

	cmp, _ := NewComparator("mtime", fileHandleWithMTime(time.Unix(10, 0)))
	// no diff to actual
	if cmp.HasChanged(fileHandleWithMTime(time.Unix(10, 0))) {
		t.Error("wrong change detected")
	}

	// diff to actual - should trigger
	if !cmp.HasChanged(fileHandleWithMTime(time.Unix(20, 0))) {
		t.Error("wrong change detected")
	}

	// diff to actual - should NOT trigger, same time as before
	if cmp.HasChanged(fileHandleWithMTime(time.Unix(20, 0))) {
		t.Error("wrong change detected")
	}

	// diff to actual and before - should trigger
	if !cmp.HasChanged(fileHandleWithMTime(time.Unix(30, 0))) {
		t.Error("wrong change detected")
	}

}
