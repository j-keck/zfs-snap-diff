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
	log.Debugf("list directory content under: %s", self.Path)
	ls, err := ioutil.ReadDir(self.Path)
	if err != nil {
		return nil, err
	}

	dirs := []FSHandle{}
	files := []FSHandle{}
	for _, fileInfo := range ls {
		if fileInfo.IsDir() {
			dirs = append(dirs, newFSHandle(self.Path, fileInfo))
		} else {
			files = append(files, newFSHandle(self.Path, fileInfo))
		}
	}
	return append(dirs, files...), nil
}
