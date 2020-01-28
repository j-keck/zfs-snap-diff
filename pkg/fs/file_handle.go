package fs

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
	defer dst.Close()

	// copy
	if _, err = io.Copy(dst, src); err != nil {
		return
	}

	// sync
	err = dst.Sync()
	return
}

func (self *FileHandle) Backup() (string, error) {
	backupDir := fmt.Sprintf("%s/.zsd", filepath.Dir(self.Path))

	// ensure backupDir exists
	if fi, err := os.Stat(backupDir); os.IsNotExist(err) {
		log.Infof("create backup directory under: %s", backupDir)
		if err := os.Mkdir(backupDir, 0770); err != nil {
			log.Warnf("unable to create backup-dir: %s", err.Error())
			return "", err
		}
	} else if !fi.Mode().IsDir() {
		msg := fmt.Sprintf("backup directory exists (%s)- but is not a directory", backupDir)
		log.Warn(msg)
		return "", errors.New(msg)
	}

	// copy the file in the backup location
	now := time.Now().Format("20060102_150405")
	backupFilePath := fmt.Sprintf("%s/%s_%s", backupDir, self.Name, now)
	log.Infof("copy actual file: %s in backup directory: %s", self.Name, backupFilePath)
	return backupFilePath, self.Copy(backupFilePath)
}
