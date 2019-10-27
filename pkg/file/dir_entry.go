package file

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

// DirEntries from a directory
type DirEntries []DirEntry

// DirEntry is a file / directory
type DirEntry struct {
	Name    string
	Path    string
	Type    string
	Size    uint64
	ModTime time.Time
}

func NewDirEntry(p string) (*DirEntry, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return new(DirEntry), err
	}
	return newDirEntry(filepath.Dir(p), fi), nil

}

func (self *DirEntry) Ls() (DirEntries, error) {
	log.Debugf("scan directory under: %s", self.Path)
	files, err := ioutil.ReadDir(self.Path)
	if err != nil {
		return nil, err
	}

	dirEntries := DirEntries{}
	for _, fi := range files {
		dirEntries = append(dirEntries, *newDirEntry(self.Path, fi))
	}
	return dirEntries, nil
}

func newDirEntry(parent string, fi os.FileInfo) *DirEntry {
	var fileType string

	// determine the file-type
	if fi.IsDir() {
		fileType = "DIR"
	} else if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		fileType = "LINK"
	} else if fi.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
		fileType = "PIPE"
	} else if fi.Mode()&os.ModeSocket == os.ModeSocket {
		fileType = "SOCKET"
	} else if fi.Mode()&os.ModeDevice == os.ModeDevice {
		fileType = "DEV"
	} else {
		fileType = "FILE"
	}

	p := path.Join(parent, fi.Name())
	return &DirEntry{fi.Name(), p, fileType, uint64(fi.Size()), fi.ModTime()}
}
