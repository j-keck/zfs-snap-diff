package fs

import (
	"archive/zip"
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

// GetFileHandle returns a handle to a existing file in the current dir-handle
func (self *DirHandle) GetFileHandle(name string) (FileHandle, error) {
	path := filepath.Join(self.Path, name)
	return GetFileHandle(path)
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
	log.Tracef("list directory content under: %s", self.Path)
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

func (self *DirHandle) CreateArchive(name string) (FileHandle, error) {
	maxSizeMB := config.Get.MaxArchiveUnpackedSizeMB
	log.Debugf("create archive from: %s as: %s (max unpacked size: %dMB)", self.Path, name, maxSizeMB)

	var add func(*FSHandle, string, *zip.Writer, *int) error
	add = func(h *FSHandle, basePath string, w *zip.Writer, archiveSize *int) error {
		if maxSizeMB > 0 && *archiveSize/1024/1024 > maxSizeMB {
			return fmt.Errorf("abort - maximum configured archive size reached (%dMB > %dMB)",
				*archiveSize/1024/1024,
				maxSizeMB)
		}
		switch h.Kind {
		case FILE:
			relPath := strings.TrimPrefix(h.Path, basePath+"/")
			log.Tracef("add %s to archive", relPath)
			f, err := w.Create(relPath)
			if err != nil {
				return err
			}

			b, err := (&FileHandle{*h}).Read()
			if err != nil {
				return err
			}

			n, err := f.Write(b)
			if err != nil {
				return err
			}
			*archiveSize += n
		case DIR:
			ls, err := (&DirHandle{*h}).Ls()
			if err != nil {
				return err
			}
			for _, e := range ls {
				err = add(&e, basePath, w, archiveSize)
				if err != nil {
					return err
				}
			}

		default:
			log.Debugf("skip %s (kind: %s) from archive", h.Name, h.Kind)
		}

		return nil
	}

	archivePath := os.TempDir() + "/" + name
	log.Debugf("create temporary archive at: %s", archivePath)
	archive, err := os.Create(archivePath)
	if err != nil {
		return FileHandle{}, err
	}
	defer archive.Close()

	w := zip.NewWriter(archive)

	archiveSize := 0
	err = add(&self.FSHandle, self.Dirname(), w, &archiveSize)
	if err != nil {
		return FileHandle{}, err
	}

	if err = w.Close(); err != nil {
		return FileHandle{}, err
	}
	log.Debugf("archive written - size: %dbytes", archiveSize)

	return GetFileHandle(archivePath)
}
