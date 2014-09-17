package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileHandle to access files / meta infos
type FileHandle struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}

// NewFileHandle creates a new FileHandle
func NewFileHandle(path string) (*FileHandle, error) {
	return newFileHandle(path)
}

// NewFileHandleInSnapshot creates a new FileHandle from a file in the given snapshot
func NewFileHandleInSnapshot(path, snapName string) (*FileHandle, error) {
	relativePath := strings.TrimPrefix(path, zfsMountPoint)
	pathInSnap := fmt.Sprintf("%s/.zfs/snapshot/%s%s", zfsMountPoint, snapName, relativePath)

	return newFileHandle(pathInSnap)
}

func newFileHandle(path string) (*FileHandle, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(path)
	return &FileHandle{name, path, fi.Size(), fi.ModTime()}, nil
}

func (fh *FileHandle) UniqueName() string {
	// file under a snapshot?
	if strings.HasPrefix(fh.Path, zfsMountPoint+"/.zfs/snapshot") {
		// extract snapshot-name
		s := strings.TrimPrefix(fh.Path, zfsMountPoint)
		s = strings.TrimPrefix(s, "/.zfs/snapshot/")
		snapName := strings.Split(s, "/")[0]

		// build unique-name
		if strings.Contains(fh.Name, ".") {
			f := strings.Split(fh.Name, ".")
			return fmt.Sprintf("%s-%s.%s", f[0], snapName, f[1])
		}
		return fmt.Sprintf("%s-%s", fh.Name, snapName)
	}

	return fh.Name
}

// MimeType of the file
func (fh *FileHandle) MimeType() (string, error) {
	f, err := os.Open(fh.Path)
	if err != nil {
		return "", err
	}

	buffer := make([]byte, 1024)
	n, err := f.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer[:n]), nil
}

// CopyTo copies the file content to the given Writer
func (fh *FileHandle) CopyTo(w io.Writer) error {
	f, err := os.Open(fh.Path)

	if err != nil {
		return err
	}

	_, err = io.Copy(w, f)
	return err
}

// HashChanged compares two FileHandles
//   * currently only per mod-time and size - performance reasons FIXME
func (fh *FileHandle) HasChanged(other *FileHandle) bool {
	timeChanged := !fh.ModTime.Equal(other.ModTime)
	sizeChanged := fh.Size != other.Size

	return timeChanged || sizeChanged
}
