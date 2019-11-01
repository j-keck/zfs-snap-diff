/*
Package fs implements filesystem operations.

To get a handle, you can use:

  - `fs.NewFSHandle` for anything in a filesystem
  - `fs.NewFileHandle` for a file handle
  - `fs.NewDirHandle` for a directory handle


You can convert a `fs.FSHandle` to a:

  - `FileHandle` per `fh, err := fsHandle.AsFileHandle()`
  - `DirHandle` per `dh, err := fsHandle.AsDirHandle()`

*/
package fs

import (
	"github.com/j-keck/plog"
)

var log = plog.GlobalLogger()
