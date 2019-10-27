package file

import (
	"errors"
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/diff"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FileHandle for file access / operations and metadata lookup
type FileHandle struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}

// NewFileHandle creates a new FileHandle
func NewFileHandle(path string) (FileHandle, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return FileHandle{}, err
	}

	name := filepath.Base(path)
	return FileHandle{name, path, fi.Size(), fi.ModTime()}, nil
}

// // UniqueName returns the unique file name
// //   * the file name if the file is in the actual filesystem
// //   * <FILE-NAME>-<SNAP-NAME>.<SUFFIX> if the file is from a snapshot
// func (fh *FileHandle) UniqueName() string {
//	ds := zfs.FindDatasetForFile(fh.Path)
//	if ds.PathIsInSnapshot(fh.Path) {
//		snapName := ds.ExtractSnapName(fh.Path)

//		// build unique-name
//		if strings.Contains(fh.Name, ".") {
//			f := strings.Split(fh.Name, ".")
//			return fmt.Sprintf("%s-%s.%s", f[0], snapName, f[1])
//		}
//		return fmt.Sprintf("%s-%s", fh.Name, snapName)
//	}

//	return fh.Name
// }

// MimeType of the file
func (fh *FileHandle) MimeType() (string, error) {
	f, err := os.Open(fh.Path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// read the first 512 bytes
	//  * http.DetectContentType considers at most 512 bytes
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer[:n]), nil
}

// ReadText returns the file content as string
func (fh *FileHandle) ReadText() (string, error) {
	b, err := ioutil.ReadFile(fh.Path)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("unable to read file content: %s", err.Error())
	}
	return string(b), nil
}

// Rename renames a file under the same directory
func (fh *FileHandle) Rename(newName string) error {
	newPath := fmt.Sprintf("%s/%s", filepath.Dir(fh.Path), newName)
	return fh.Move(newPath)
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

// CopyTo copies the file content to the given Writer
func (fh *FileHandle) CopyTo(w io.Writer) error {
	f, err := os.Open(fh.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

// CopyAs copies a file
func (fh *FileHandle) CopyAs(path string) (err error) {
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

// Patch applys the given deltas to the current file
//   * deleted entries are inserted
//   * inserted entries are removed
func (fh *FileHandle) Patch(deltas diff.Deltas) error {

	// verify the equal parts from the deltas are the same as in the given file
	// returns a error if not
	verifyDeltasAreApplicable := func() error {
		f, err := os.Open(fh.Path)
		if err != nil {
			return fmt.Errorf("open file: '%s' - %s", fh.Name, err.Error())
		}
		defer f.Close()

		for _, delta := range deltas {
			if delta.Type == diff.Eq {
				buffer := make([]byte, len(delta.Text))
				if _, err = f.ReadAt(buffer, delta.StartPosTarget-1); err != nil && err != io.EOF {
					return fmt.Errorf("read file: '%s' - %s", fh.Name, err.Error())
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
		// open src / dst
		src, err := os.Open(fh.Path)
		if err != nil {
			return fmt.Errorf("unable to open src-file: '%s' - %s", fh.Path, err.Error())
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("unable to open dst-file: '%s' - %s", dstPath, err.Error())
		}
		defer func() {
			dst.Close()
			dst.Sync()
		}()

		// apply deltas
		var srcPos, offset int64
		for _, delta := range deltas {
			if delta.Type == diff.Del {
				// copy unchanged
				bytesToRead := delta.StartPosTarget - 1 - srcPos
				if offset, err = io.CopyN(dst, src, bytesToRead); err != nil && err != io.EOF {
					return fmt.Errorf("copy unchanged text - %s", err.Error())
				}
				srcPos += offset

				// restore deleted text
				if _, err := dst.Write([]byte(delta.Text)); err != nil {
					return fmt.Errorf("restore deleted text - %s", err.Error())
				}
			}

			if delta.Type == diff.Ins {
				// copy unchanged
				bytesToRead := delta.StartPosTarget - 1 - srcPos
				if offset, err = io.CopyN(dst, src, bytesToRead); err != nil && err != io.EOF {
					return fmt.Errorf("copy unchanged text - %s", err.Error())
				}
				srcPos += offset

				// skip inserted text
				deletedTextLength := int64(len(delta.Text))
				if _, err = src.Seek(deletedTextLength, 1); err != nil {
					return fmt.Errorf("seek error - %s", err.Error())
				}

				srcPos += deletedTextLength

			}
		}
		// copy the rest
		if _, err = io.Copy(dst, src); err != nil && err != io.EOF {
			return fmt.Errorf("copy rest text - %s", err.Error())
		}

		return nil
	}

	// check
	if err := verifyDeltasAreApplicable(); err != nil {
		return fmt.Errorf("verify deltas: %s", err.Error())
	}

	// patch
	tsString := time.Now().Format("20060102_150405")
	patchWorkFilePath := fmt.Sprintf("%s/.zsd-patch-in-process-%s_%s", filepath.Dir(fh.Path), fh.Name, tsString)
	if err := applyDeltasTo(patchWorkFilePath); err != nil {
		// delete patch work file
		os.Remove(patchWorkFilePath)
		return fmt.Errorf("unable to apply deltas - keep file untouched - %s", err.Error())
	}

	if err := fh.MoveToBackup(); err != nil {
		return err
	}

	if err := os.Rename(patchWorkFilePath, fh.Path); err != nil {
		return fmt.Errorf("unable to rename patch file to original file - %s", err.Error())
	}

	return nil
}

// MoveToBackup moves the file in the backup location
func (fh *FileHandle) MoveToBackup() error {
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

	// move file, don't update Name / Path in FileHandle
	now := time.Now().Format("20060102_150405")
	backupFilePath := fmt.Sprintf("%s/%s_%s", backupDir, fh.Name, now)
	log.Infof("move actual file in backup directory: %s\n", backupFilePath)
	return os.Rename(fh.Path, backupFilePath)
}
