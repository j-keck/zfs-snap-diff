package fs

import (
	"io/ioutil"
)

type DirHandle struct {
	FSHandle
}

func NewDirHandle(path string) (DirHandle, error) {
	handle, err := NewFSHandle(path)
	if err != nil {
		return DirHandle{}, err
	}

	return handle.AsDirHandle()
}

func (self *DirHandle) Ls() ([]FSHandle, error) {
	log.Debugf("scan directory under: %s", self.Path)
	ls, err := ioutil.ReadDir(self.Path)
	if err != nil {
		return nil, err
	}

	entries := []FSHandle{}
	for _, fileInfo := range ls {
		entries = append(entries, newFSHandle(self.Path, fileInfo))
	}
	return entries, nil
}
