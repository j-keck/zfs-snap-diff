package main

import (
	"io/ioutil"
	"time"
)

// DirEntries from a directory
type DirEntries []DirEntry

// ScanDirEntries scan a given directory
func ScanDirEntries(path string) (DirEntries, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirEntries DirEntries
	for _, fi := range files {
		_type := "F"
		if fi.IsDir() {
			_type = "D"
		}
		dirEntry := DirEntry{_type, fi.Name(), fi.Size(), fi.ModTime()}

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
