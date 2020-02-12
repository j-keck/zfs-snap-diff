//
// Package fs provides filesystem operations.
//
// To get a handle to existing fs-entry, you can use:
//
//   - `fs.GetFileHandle(path string)` for a file handle
//   - `fs.GetDirHandle(path string)` for a directory handle
//
// this operations fail if the request entry does not exists.
//
//
// To get or create a directory handle use
//
//   - `fs.GetOrCreateDirHandle(path string, perm os.FilePerm)`
//   - `fs.GetOrCreateSubDirHandle(name string, perm os.FilePerm)`
//
package fs

import (
	"github.com/j-keck/plog"
)

var log = plog.GlobalLogger()
