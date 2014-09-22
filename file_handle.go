package main

import (
	"bytes"
	"crypto/md5"
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

// Rename renames a file under the same directory
func (fh *FileHandle) Rename(newName string) error {
	newPath := fmt.Sprintf("%s/%s", filepath.Dir(fh.Path), newName)
	if err := os.Rename(fh.Path, newPath); err != nil {
		return err
	}

	// update file name / path
	fh.Name = newName
	fh.Path = newPath
	return nil
}

// Move moves / renames a file
func (fh *FileHandle) Move(newPath string) error {
	if err := os.Rename(fh.Path, newPath); err != nil {
		return err
	}

	// update file name / path
	fh.Name = filepath.Base(newPath)
	fh.Path = newPath
	return nil
}

// Copy copies a file
func (fh *FileHandle) Copy(path string) (err error) {
	// open src
	in, err := os.Open(fh.Path)
	if err != nil {
		return err
	}
	defer in.Close()

	// open dest
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := out.Close()
		if err == nil {
			err = closeErr

			if err == nil {
				// copy success - update file name / path
				fh.Name = filepath.Base(path)
				fh.Path = path
			}
		}
	}()

	// copy
	if _, err = io.Copy(out, in); err != nil {
		return
	}

	// sync
	err = out.Sync()
	return
}

// CompareFileFuncByName resturns a compare-file function by name
func CompareFileFuncByName(compareFileMethod string) (CompareFileFunc, error) {
	switch compareFileMethod {
	case "size+modTime":
		return CompareFileBySizeAndModTime(), nil
	case "size":
		return CompareFileBySize(), nil
	case "md5":
		return CompareFileByMD5(), nil
	default:
		return nil, fmt.Errorf("no such compare method: '%s' - avaliable: 'size+modTime', 'size' or 'md5'", compareFileMethod)
	}
}

// CompareFileFunc for different compare methods
type CompareFileFunc func(*FileHandle, *FileHandle) bool

// CompareFileBySize compares files per size
func CompareFileBySize() CompareFileFunc {
	return func(a *FileHandle, b *FileHandle) bool {
		return a.Size != b.Size
	}
}

// CompareFileBySizeAndModTime compares files per size+modTime
func CompareFileBySizeAndModTime() CompareFileFunc {
	return func(a *FileHandle, b *FileHandle) bool {
		timeChanged := !a.ModTime.Equal(b.ModTime)
		sizeChanged := a.Size != b.Size

		return timeChanged || sizeChanged
	}
}

// CompareFileByMD5 compares files per md5
func CompareFileByMD5() CompareFileFunc {
	calculateMD5 := func(fh *FileHandle) []byte {
		in, err := os.Open(fh.Path)
		if err != nil {
			panic(err)
		}
		defer in.Close()

		buf := make([]byte, 1024)
		hash := md5.New()
		for {
			n, err := in.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 0 {
				break
			}

			if _, err := io.WriteString(hash, string(buf[:n])); err != nil {
				panic(err)
			}
		}
		return hash.Sum(nil)
	}

	return func(a *FileHandle, b *FileHandle) bool {
		return bytes.Compare(calculateMD5(a), calculateMD5(b)) != 0
	}
}
