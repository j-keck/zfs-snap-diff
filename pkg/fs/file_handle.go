package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type FileHandle struct {
	FSHandle
}

func NewFileHandle(path string) (FileHandle, error) {
	handle, err := NewFSHandle(path)
	if err != nil {
		return FileHandle{}, err
	}

	return handle.AsFileHandle()
}

func (self *FileHandle) MimeType() (string, error) {
	fh, err := os.Open(self.Path)
	if err != nil {
		return "", err
	}
	defer fh.Close()

	// read the first 512 bytes
	//  * http.DetectContentType considers at most 512 bytes
	buf := make([]byte, 512)
	n, err := fh.Read(buf)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}

func (self *FileHandle) Read() ([]byte, error) {
	buf, err := ioutil.ReadFile(self.Path)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to read file content: %v", err)
	}
	return buf, nil
}

func (self *FileHandle) ReadString() (string, error) {
	buf, err := self.Read()
	return string(buf), err
}

func (self *FileHandle) CopyTo(w io.Writer) error {
	fh, err := os.Open(self.Path)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(w, fh)
	return err
}

// Copy copies a file
func (fh *FileHandle) Copy(path string) (err error) {
	var src, dst *os.File

	// open src
	if src, err = os.Open(fh.Path); err != nil {
		return err
	}
	defer src.Close()

	// open dest
	if dst, err = os.Create(path); err != nil {
		return err
	}
	defer func() {
		closeErr := dst.Close()
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
	if _, err = io.Copy(dst, src); err != nil {
		return
	}

	// sync
	err = dst.Sync()
	return
}
