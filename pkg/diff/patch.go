package diff

import (
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"io"
	"os"
	"path/filepath"
	"time"
)

func PatchPath(path string, deltas Deltas) error {
	fh, err := fs.NewFileHandle(path)
	if err != nil {
		return err
	}

	return Patch(fh, deltas)
}

// Patch applys the given deltas to the current file
//   * deleted entries are inserted
//   * inserted entries are removed
func Patch(fh fs.FileHandle, deltas Deltas) error {

	// verify the equal parts from the deltas are the same as in the given file
	// returns a error if not
	verifyDeltasAreApplicable := func() error {
		f, err := os.Open(fh.Path)
		if err != nil {
			return fmt.Errorf("open file: '%s' - %s", fh.Name, err.Error())
		}
		defer f.Close()

		for _, delta := range deltas {
			if delta.Type == Eq {
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
			if delta.Type == Del {
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

			if delta.Type == Ins {
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

	if err := fs.Backup(fh); err != nil {
		return err
	}

	if err := os.Rename(patchWorkFilePath, fh.Path); err != nil {
		return fmt.Errorf("unable to rename patch file to original file - %s", err.Error())
	}

	return nil
}
