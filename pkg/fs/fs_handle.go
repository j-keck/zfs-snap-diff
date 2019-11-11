package fs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

// FSHandle represents a handle to a filesystem entry
//
// This can be a file, a directory or anything else.
type FSHandle struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Kind    Kind      `json:"kind"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

func NewFSHandle(path string) (FSHandle, error) {
	if len(path) == 0 {
		return FSHandle{}, errors.New("the given path was empty")
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return FSHandle{}, err
	}

	return newFSHandle(filepath.Dir(path), fileInfo), nil
}

func newFSHandle(dirname string, fileInfo os.FileInfo) FSHandle {
	path := path.Join(dirname, fileInfo.Name())
	kind := KindFromFileInfo(fileInfo)

	return FSHandle{
		Name:    fileInfo.Name(),
		Path:    path,
		Kind:    kind,
		Size:    fileInfo.Size(),
		ModTime: fileInfo.ModTime(),
	}
}

func (self *FSHandle) AsFileHandle() (FileHandle, error) {
	if self.Kind == FILE {
		return FileHandle{*self}, nil
	}
	return FileHandle{}, fmt.Errorf("'%s' is not a file - it's a '%s'", self.Path, self.Kind)
}

func (self *FSHandle) AsDirHandle() (DirHandle, error) {
	if self.Kind == DIR {
		return DirHandle{*self}, nil
	}
	return DirHandle{}, fmt.Errorf("'%s' is not a dir - it's a '%s'", self.Path, self.Kind)
}

func (self *FSHandle) Dirname() string {
	return filepath.Dir(self.Path)
}

func (self *FSHandle) Move(path string) error {
	if err := os.Rename(self.Path, path); err != nil {
		return err
	}

	self.Name = filepath.Base(path)
	self.Path = path
	return nil
}

func (self *FSHandle) Rename(name string) error {
	path := filepath.Join(self.Dirname(), name)
	return self.Move(path)
}
