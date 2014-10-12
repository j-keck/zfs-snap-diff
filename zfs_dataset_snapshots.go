package main

import (
	"errors"
	"strings"
)

// ZFSSnapshots represents snapshots from a zfs dataset
type ZFSSnapshots []ZFSSnapshot

// ScanSnapshots scan snapshots for the given zfs dataset
func (ds *ZFSDataset) ScanSnapshots() (ZFSSnapshots, error) {
	out, err := ds.execZFS("list -t snapshot -r -d 1 -o name,creation -H", ds.Name)
	if err != nil {
		return nil, errors.New(out)
	}

	snapshots := ZFSSnapshots{}
	for _, line := range strings.Split(out, "\n") {
		if snapName, creation, ok := split2(line, "\t"); ok {
			// remove dataset name from snapshot
			snapName := lastElement(snapName, "@")

			// path
			path := ds.MountPoint + "/.zfs/snapshot/" + snapName

			// append new snap to snapshots
			snapshots = append(snapshots, ZFSSnapshot{snapName, creation, path})
		}

	}
	return snapshots.Reverse(), nil
}

// Reverse reverse the snapshot list
func (s ZFSSnapshots) Reverse() ZFSSnapshots {
	reversed := ZFSSnapshots{}
	for i := len(s) - 1; i >= 0; i-- {
		reversed = append(reversed, s[i])
	}
	return reversed
}

// Filter filters snapshots per given filter function
func (s *ZFSSnapshots) Filter(f func(ZFSSnapshot) bool) ZFSSnapshots {
	newS := ZFSSnapshots{}
	for _, snap := range *s {
		if f(snap) {
			newS = append(newS, snap)
		}
	}
	return newS
}

func (s *ZFSSnapshots) FilterWhereFileWasModified(path string, fileHasChangedFuncGen FileHasChangedFuncGen) ZFSSnapshots {
	fh, _ := NewFileHandle(path)
	fileHasChangedFunc := fileHasChangedFuncGen(fh)

	return s.Filter(func(snap ZFSSnapshot) bool {
		// ignore errors here if file not found (e.g. was deleted)
		if snapFileFh, err := NewFileHandleInSnapshot(path, snap.Name); err == nil {
			if fileHasChangedFunc(fh, snapFileFh) {
				// file changed in snapshot
				fh = snapFileFh
				return true
			}
		}
		return false
	})
}

// ZFSSnapshot - zfs snapshot
type ZFSSnapshot struct {
	Name     string
	Creation string
	Path     string
}
