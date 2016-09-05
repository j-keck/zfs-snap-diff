package main

import (
	"io/ioutil"
	"os"
	"time"
)

// DirEntries from a directory
type DirEntries []DirEntry

// ScanDirEntries scan a given directory
func ScanDirEntries(path string) (DirEntries, error) {
	logDebug.Printf("scan directory under: %s\n", path)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dirEntries := DirEntries{}
	for _, fileInfo := range files {
		var fileType string

		// determine the file-type
		if fileInfo.IsDir() {
			fileType = "DIR"
		} else if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			fileType = "LINK"
		} else if fileInfo.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
			fileType = "PIPE"
		} else if fileInfo.Mode()&os.ModeSocket == os.ModeSocket {
			fileType = "SOCKET"
		} else if fileInfo.Mode()&os.ModeDevice == os.ModeDevice {
			fileType = "DEV"
		} else {
			fileType = "FILE"
		}

		dirEntry := DirEntry{fileType, fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime()}
		dirEntries = append(dirEntries, dirEntry)
	}
	return dirEntries, nil
}

// DirEntry is a file / directory
type DirEntry struct {
	Type    string
	Path    string
	Size    int64
	ModTime time.Time
}
