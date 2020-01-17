package scanner

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)


// Comparator compares ...
type Comparator interface {
	init(actual fs.FileHandle)
	HasChanged(other fs.FileHandle) bool
}


func NewComparator(method string, fh fs.FileHandle) (Comparator, error) {

	var comparator Comparator
	switch method {
	case "size":
		comparator = new(CompareBySize)
	case "modTime", "mtime":
		comparator = new(CompareByMTime)
	case "size+modTime", "size+mtime":
		comparator = new(CompareBySizeAndModTime)
	case "content":
		comparator = new(CompareByContent)
	case "md5":
		comparator = new(CompareByMD5)
	case "auto":
		// use md5 for text files, size+modTime for others
		mimeType, _ := fh.MimeType()
		if strings.HasPrefix(mimeType, "text") {
			comparator = new(CompareByMD5)
		} else {
			comparator = new(CompareBySizeAndModTime)
		}
	default:
		return nil, fmt.Errorf("no such comparator: '%s'", method)
	}
	comparator.init(fh)

	return comparator, nil
}

//
// by size
type CompareBySize struct {
	actual    fs.FileHandle
	otherSize int64
}

func (self *CompareBySize) init(actual fs.FileHandle) {
	self.actual = actual
}
func (self *CompareBySize) HasChanged(other fs.FileHandle) bool {
	// previous other size
	prevSize := self.otherSize

	// cache the size of the other file for the next run
	self.otherSize = other.Size

	return self.actual.Size != other.Size &&
		prevSize != other.Size
}

// by modification time
type CompareByMTime struct {
	actual     fs.FileHandle
	otherMTime time.Time
}

func (self *CompareByMTime) init(actual fs.FileHandle) {
	self.actual = actual
}
func (self *CompareByMTime) HasChanged(other fs.FileHandle) bool {
	// previous other mtime
	prevMTime := self.otherMTime

	// cache the mtime of the other file for the next run
	self.otherMTime = other.ModTime

	return !(self.actual.ModTime.Equal(other.ModTime) || prevMTime.Equal(other.ModTime))
}

//
// by size and modification time
type CompareBySizeAndModTime struct {
	bySize  Comparator
	byMTime Comparator
}

func (self *CompareBySizeAndModTime) init(actual fs.FileHandle) {
	bySize := new(CompareBySize)
	bySize.init(actual)
	self.bySize = bySize

	byMTime := new(CompareByMTime)
	byMTime.init(actual)
	self.byMTime = byMTime
}
func (self *CompareBySizeAndModTime) HasChanged(other fs.FileHandle) bool {
	return self.bySize.HasChanged(other) || self.byMTime.HasChanged(other)
}

//
// compare by content
type CompareByContent struct {
	actual        fs.FileHandle
	actualContent []byte
	otherContent  []byte
}

func (self *CompareByContent) init(actual fs.FileHandle) {
	self.actual = actual

	buf, err := ioutil.ReadFile(self.actual.Path)
	if err != nil {
		log.Warnf("unable to read the 'actual' file: %s - err: %v", self.actual.Path, err)
	}
	self.actualContent = buf
}
func (self *CompareByContent) HasChanged(other fs.FileHandle) bool {
	buf, err := ioutil.ReadFile(other.Path)
	if err != nil {
		log.Warnf("unable to read the 'other' file: %s - err: %v", other.Path, err)
	}

	// previous other content
	prevContent := self.otherContent

	// cache the content of the other file for the next run
	self.otherContent = buf

	return bytes.Compare(self.actualContent, buf) != 0 &&
		bytes.Compare(prevContent, buf) != 0
}

//
// compare by md5 hash
type CompareByMD5 struct {
	actual     fs.FileHandle
	actualHash []byte
	otherHash  []byte
}

func (self *CompareByMD5) init(actual fs.FileHandle) {
	self.actual = actual

	h, err := self.calculateMD5(actual.Path)
	if err != nil {
		log.Warnf("unable to hash the 'actual' file: %s - err: %v", self.actual.Path, err)
	}
	self.actualHash = h
}
func (self *CompareByMD5) HasChanged(other fs.FileHandle) bool {
	h, err := self.calculateMD5(other.Path)
	if err != nil {
		log.Warnf("unable to hash the 'other' file: %s - err: %v", other.Path, err)
		return true
	}

	// previous other's hash
	prevHash := self.otherHash

	// cache the md5 of the other file for the next run
	self.otherHash = h

	// compare the actual (newest) hash with the others file hash AND
	// the others file hash with the previous hash
	return bytes.Compare(self.actualHash, h) != 0 &&
		bytes.Compare(prevHash, h) != 0

}
func (self *CompareByMD5) calculateMD5(path string) ([]byte, error) {

	fh, err := os.Open(path)
	if err != nil {
		log.Warnf("unable to open file: %s, err: %v", path, err)
		return nil, err
	}
	defer fh.Close()

	h := md5.New()
	if _, err := io.Copy(h, fh); err != nil {
		log.Warnf("error reading file: %s, err: %v", path, err)
		return nil, err
	}

	return h.Sum(nil), nil
}
