package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"path/filepath"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
)

// FileHandle represents a file
type FileHandle struct {
	FSHandle
}

// GetFileHandle returns a handle to a existing file.
// If the file does not exists, a error is returned.
// To create a file, use 'DirHandle.WriteFile'.
func GetFileHandle(path string) (FileHandle, error) {
	handle, err := GetFSHandle(path)
	if err != nil {
		return FileHandle{}, err
	}

	return handle.AsFileHandle()
}

// MimeType returns file mime-type.
// This functions makes io-operations to read a part of the file.
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

// Read returns the whole file content as a byte array.
func (self *FileHandle) Read() ([]byte, error) {
	buf, err := ioutil.ReadFile(self.Path)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to read file content: %v", err)
	}
	return buf, nil
}


// ReadString returns the whole file content as a string.
func (self *FileHandle) ReadString() (string, error) {
	buf, err := self.Read()
	return string(buf), err
}


// CopyTo copies the whole file content into a given writer.
func (self *FileHandle) CopyTo(w io.Writer) error {
	fh, err := os.Open(self.Path)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(w, fh)
	return err
}

// Copy copies a file.
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
	defer dst.Close()

	// copy
	if _, err = io.Copy(dst, src); err != nil {
		return
	}

	// sync
	err = dst.Sync()
	return
}

// Backup create a backup of the file in the backup location.
func (self *FileHandle) Backup() (string, error) {

	var backupPath string
	if config.Get.UseCacheDirForBackups {
		cacheDir, err := CacheDir()
		if err != nil {
			return "", err
		}

		backupDir, err := cacheDir.GetOrCreateSubDirHandle("backups", 0700)
		if err != nil {
			return "", err
		}
		backupPath = filepath.Join(backupDir.Path, self.Dirname())

		// create the backup directory hierarchy
		os.MkdirAll(backupPath, 0700)

	} else {
		dir, err := self.Dir()
		if err != nil {
			return "", err
		}

		backupDir, err := dir.GetOrCreateSubDirHandle(".zsd", 0770)
		if err != nil {
			return "", err
		}
		backupPath = backupDir.Path
	}


	// copy the file in the backup location
	now := time.Now().Format("20060102_150405")
	backupFilePath := fmt.Sprintf("%s/%s_%s", backupPath, self.Name, now)
	log.Debugf("copy actual file: %s in backup directory: %s", self.Name, backupFilePath)
	return backupFilePath, self.Copy(backupFilePath)
}
