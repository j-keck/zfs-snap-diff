package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FIXME: make this a function from FileHandle
func Backup(fh FileHandle) error {
	backupDir := fmt.Sprintf("%s/.zsd", filepath.Dir(fh.Path))

	// ensure backupDir exists
	if fi, err := os.Stat(backupDir); os.IsNotExist(err) {
		log.Infof("create backup directory under: %s\n", backupDir)
		if err := os.Mkdir(backupDir, 0770); err != nil {
			log.Warnf("unable to create backup-dir: %s\n", err.Error())
			return err
		}
	} else if !fi.Mode().IsDir() {
		msg := fmt.Sprintf("backup directory exists (%s)- but is not a directory\n", backupDir)
		log.Warn(msg)
		return errors.New(msg)
	}

	// FIXME: make a copy instead of a move?
	// move the file in the backup location
	now := time.Now().Format("20060102_150405")
	backupFilePath := fmt.Sprintf("%s/%s_%s", backupDir, fh.Name, now)
	log.Infof("move actual file in backup directory: %s\n", backupFilePath)
	return os.Rename(fh.Path, backupFilePath)
}
