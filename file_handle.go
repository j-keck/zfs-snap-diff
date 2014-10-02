package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
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
	relativePath := strings.TrimPrefix(path, zfs.MountPoint)
	pathInSnap := fmt.Sprintf("%s/.zfs/snapshot/%s%s", zfs.MountPoint, snapName, relativePath)

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
	if strings.HasPrefix(fh.Path, zfs.MountPoint+"/.zfs/snapshot") {
		// extract snapshot-name
		s := strings.TrimPrefix(fh.Path, zfs.MountPoint)
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

func (fh *FileHandle) ReadText() (string, error) {
	b, err := ioutil.ReadFile(fh.Path)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("unable to read file content: %s", err.Error())
	}
	return string(b), nil
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

func (fh *FileHandle) Patch(deltas Deltas) error {
	var err error

	// verify the equal parts from the deltas are the same as in the given file
	// returns a error if not
	verifyDeltasAreApplicable := func() error {
		var f *os.File
		if f, err = os.Open(fh.Path); err != nil {
			return fmt.Errorf("unable to open file: '%s' - %s", fh.Name, err.Error())
		}
		defer f.Close()

		for _, delta := range deltas {
			if delta.Type == Eq {
				buffer := make([]byte, len(delta.Text))
				if _, err = f.ReadAt(buffer, delta.StartPosTarget-1); err != nil && err != io.EOF {
					return fmt.Errorf("unable to read file: '%s' - %s", fh.Name, err.Error())
				}

				if string(buffer) != delta.Text {
					msg := "unexpected content in file: '%s' - pos=%d, expected='%s', found='%s' - file changed?"
					return fmt.Errorf(msg, fh.Name, delta.StartPosTarget, delta.Text, string(buffer))
				}
			}
		}
		return nil
	}

	// apply the deltas to a given file
	applyDeltasTo := func(dstPath string) error {
		var src, dst *os.File
		// open src / dst
		if src, err = os.Open(fh.Path); err != nil {
			return fmt.Errorf("unable to open src-file: '%s' - %s", fh.Path, err.Error())
		}
		defer src.Close()

		if dst, err = os.Create(dstPath); err != nil {
			return fmt.Errorf("unable to open dst-file: '%s' - %s", dstPath, err.Error())
		}
		defer func() {
			dst.Close()
			dst.Sync()
		}()

		// apply deltas
		var srcPos, offset int64
		for _, delta := range deltas {
			if delta.Type == Del {
				// copy unchanged
				bytesToRead := delta.StartPosTarget - 1 - srcPos
				if offset, err = io.CopyN(dst, src, bytesToRead); err != nil && err != io.EOF {
					return fmt.Errorf("unable to copy unchanged text - %s", err.Error())
				}
				srcPos += offset

				// restore deleted text
				if _, err := dst.Write([]byte(delta.Text)); err != nil {
					return fmt.Errorf("unable to restore deleted text - %s", err.Error())
				}
			}

			if delta.Type == Ins {
				// copy unchanged
				bytesToRead := delta.StartPosTarget - 1 - srcPos
				if offset, err = io.CopyN(dst, src, bytesToRead); err != nil && err != io.EOF {
					return fmt.Errorf("unable to copy unchanged text - %s", err.Error())
				}
				srcPos += offset

				// skip inserted text
				deletedTextLength := int64(len(delta.Text))
				if _, err = src.Seek(deletedTextLength, 1); err != nil {
					return fmt.Errorf("unable to seek - %s", err.Error())
				}

				srcPos += deletedTextLength

			}
		}
		// copy the rest
		if _, err = io.Copy(dst, src); err != nil && err != io.EOF {
			return fmt.Errorf("unable to copy rest text - %s", err.Error())
		}

		return nil
	}

	// check
	if err := verifyDeltasAreApplicable(); err != nil {
		return fmt.Errorf("unable to verify deltas: %s", err.Error())
	}

	// patch
	tsString := time.Now().Format("20060102_150405")
	patchWorkFilePath := fmt.Sprintf("%s/.zsd-patch-in-process-%s_%s", filepath.Dir(fh.Path), fh.Name, tsString)
	if applyDeltasErr := applyDeltasTo(patchWorkFilePath); applyDeltasErr != nil {
		// delete patch work file
		os.Remove(patchWorkFilePath)
		return fmt.Errorf("unable to apply deltas - keep file untouched - %s", applyDeltasErr.Error())
	}

	backupFilePath := fmt.Sprintf("%s/%s_%s", filepath.Dir(fh.Path), fh.Name, tsString)
	if err := os.Rename(fh.Path, backupFilePath); err != nil {
		return fmt.Errorf("unable to rename orginal file - %s", err.Error())
	}

	if err := os.Rename(patchWorkFilePath, fh.Path); err != nil {
		return fmt.Errorf("unable to rename patch file to orginal file - %s", err.Error())
	}

	return nil
}

// FileHasChangedFunGen to create a FileHasChangedFunc
type FileHasChangedFuncGen func(*FileHandle) FileHasChangedFunc

// FileHasChangedFunc to detect if a file has changed
type FileHasChangedFunc func(*FileHandle, *FileHandle) bool

// NewFileHasChangedFuncGenByName returns a generator for a FileHasChangedFunc
// which implements a file changes algorithm by the given name
//   example:
//     hasChangedGen := NewFileHasChangedFuncGenByName("md5")
//     hasChanged := hasChangedGen(actualFH)
//     if hasChanged(snap1FH, snap2FH) {
//       -> actualFH content is different from snap1FH and snap2FH
//     }
func NewFileHasChangedFuncGenByName(method string) (FileHasChangedFuncGen, error) {
	switch method {

	case "size+modTime":
		return CompareFileBySizeAndModTime, nil
	case "size":
		return CompareFileBySize, nil
	case "md5":
		return CompareFileByMD5, nil
	case "auto":
		// use md5 for text files, size+modTime for others
		return func(actual *FileHandle) FileHasChangedFunc {
			mimeType, _ := actual.MimeType()
			if strings.HasPrefix(mimeType, "text") {
				return CompareFileByMD5(actual)
			}
			return CompareFileBySizeAndModTime(actual)
		}, nil
	default:
		return nil, fmt.Errorf("no such compare method: '%s' - avaliable: 'size+modTime', 'size', 'md5'", method)
	}
}

// CompareFileBySize returns a FileHasChangedFunc which compares files per their size
func CompareFileBySize(actual *FileHandle) FileHasChangedFunc {
	return func(a, b *FileHandle) bool {
		return a.Size != b.Size
	}
}

// CompareFileBySizeAndModTime returns a FileHasChangedFunc which compares files per their size+modTime
func CompareFileBySizeAndModTime(actual *FileHandle) FileHasChangedFunc {
	hasChanged := func(a, b *FileHandle) bool {
		timeChanged := !a.ModTime.Equal(b.ModTime)
		sizeChanged := a.Size != b.Size

		return timeChanged || sizeChanged
	}

	return func(a, b *FileHandle) bool {
		return hasChanged(a, b)
	}
}

// CompareFileByMD5 returns a FileHasChangedFunc which compares files per their md5 sum
func CompareFileByMD5(actual *FileHandle) FileHasChangedFunc {
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

	actualMD5 := calculateMD5(actual)

	var cached []byte
	return func(a, b *FileHandle) bool {
		var aMD5 []byte

		// use the cached value for aMD5 if it's not empty
		if cached == nil {
			aMD5 = calculateMD5(a)
		} else {
			aMD5 = cached
		}

		bMD5 := calculateMD5(b)

		// cache the current bMD5 for the next aMD5
		cached = bMD5
		return bytes.Compare(actualMD5, bMD5) != 0 && bytes.Compare(aMD5, bMD5) != 0
	}
}
