package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// DirHandle represents a directory
type DirHandle struct {
	FSHandle
}

// GetDirHandle returns a handle to a existing directory.
// If the directory does not exists, a error is returned.
func GetDirHandle(path string) (DirHandle, error) {
	handle, err := GetFSHandle(path)

	if err != nil {
		return DirHandle{}, err
	}

	return handle.AsDirHandle()
}


// GetOrCreateDirHandle returns a handle to a existing or a new created directory.
func GetOrCreateDirHandle(path string, perm os.FileMode) (DirHandle, error) {
	if dir, err := GetDirHandle(path); err == nil {
		log.Tracef("GetOrCreateDirHandle - directory: %s exists", path)
		return dir, nil
	} else {
		if os.IsNotExist(err) {
			log.Infof("requested directory: %s does not exits - create it", path)
			if err := os.Mkdir(path, perm); err != nil {
				return DirHandle{}, err
			}
			return GetDirHandle(path)
		}
		return DirHandle{}, err
	}
}


// GetSubDirHandle returns a child-dir of the current dir-handle.
// This operation fails if the requested directory does not exists.
func (self *DirHandle) GetSubDirHandle(name string) (DirHandle, error) {
	path := filepath.Join(self.Path, name)
	return GetDirHandle(path)
}

// GetOrCreateSubDirHandle returns a handle to a existing or a new created child-dir of the current dir-handle.
func (self *DirHandle) GetOrCreateSubDirHandle(name string, perm os.FileMode) (DirHandle, error) {
	path := filepath.Join(self.Path, name)
	return GetOrCreateDirHandle(path, perm)
}


// ReadFile reads a child-file of the current dir-handle.
func (self *DirHandle) ReadFile(name string) ([]byte, error) {
	path := filepath.Join(self.Path, name)
	fh, err := GetFileHandle(path)
	if err != nil {
		return nil, err
	}

	return fh.Read()
}

// WriteFile creates or overrides a child-file of the current dir-handle
func (self *DirHandle) WriteFile(name string, data []byte, perm os.FileMode) (FileHandle, error) {
	path := filepath.Join(self.Path, name)
	if err := ioutil.WriteFile(path, data, perm); err != nil {
		return FileHandle{}, err
	}
	return GetFileHandle(path)
}

// Ls returns a directory listing of the current dir-handle.
// Directory and filenames are grouped and sorted by names.
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
			dirs = append(dirs, getFSHandle(self.Path, fileInfo))
		} else {
			files = append(files, getFSHandle(self.Path, fileInfo))
		}
	}
	return append(dirs, files...), nil
}

