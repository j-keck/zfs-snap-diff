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
	init(current fs.FileHandle)
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
	current   fs.FileHandle
	otherSize int64
}

func (self *CompareBySize) init(current fs.FileHandle) {
	self.current = current
}
func (self *CompareBySize) HasChanged(other fs.FileHandle) bool {
	// previous other size
	prevSize := self.otherSize

	// cache the size of the other file for the next run
	self.otherSize = other.Size

	return self.current.Size != other.Size &&
		prevSize != other.Size
}

// by modification time
type CompareByMTime struct {
	current    fs.FileHandle
	otherMTime time.Time
}

func (self *CompareByMTime) init(current fs.FileHandle) {
	self.current = current
}
func (self *CompareByMTime) HasChanged(other fs.FileHandle) bool {
	// previous other mtime
	prevMTime := self.otherMTime

	// cache the mtime of the other file for the next run
	self.otherMTime = other.MTime

	return !(self.current.MTime.Equal(other.MTime) || prevMTime.Equal(other.MTime))
}

//
// by size and modification time
type CompareBySizeAndModTime struct {
	bySize  Comparator
	byMTime Comparator
}

func (self *CompareBySizeAndModTime) init(current fs.FileHandle) {
	bySize := new(CompareBySize)
	bySize.init(current)
	self.bySize = bySize

	byMTime := new(CompareByMTime)
	byMTime.init(current)
	self.byMTime = byMTime
}
func (self *CompareBySizeAndModTime) HasChanged(other fs.FileHandle) bool {
	return self.bySize.HasChanged(other) || self.byMTime.HasChanged(other)
}

//
// compare by content
type CompareByContent struct {
	current        fs.FileHandle
	currentContent []byte
	otherContent   []byte
}

func (self *CompareByContent) init(current fs.FileHandle) {
	self.current = current

	buf, err := ioutil.ReadFile(self.current.Path)
	if err != nil {
		log.Warnf("unable to read the 'current' file: %s - err: %v", self.current.Path, err)
	}
	self.currentContent = buf
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

	return bytes.Compare(self.currentContent, buf) != 0 &&
		bytes.Compare(prevContent, buf) != 0
}

//
// compare by md5 hash
type CompareByMD5 struct {
	current     fs.FileHandle
	currentHash []byte
	otherHash   []byte
}

func (self *CompareByMD5) init(current fs.FileHandle) {
	self.current = current

	h, err := self.calculateMD5(current.Path)
	if err != nil {
		log.Warnf("unable to hash the 'current' file: %s - err: %v", self.current.Path, err)
	}
	self.currentHash = h
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

	// compare the current (newest) hash with the others file hash AND
	// the others file hash with the previous hash
	return bytes.Compare(self.currentHash, h) != 0 &&
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
