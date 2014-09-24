package main

import (
	"errors"
	"strings"
)

// ZFSSnapshots represents snapshots from a zfs filesystem
type ZFSSnapshots []ZFSSnapshot

// ScanSnapshots scan snapshots for the given zfs filesystem
func (zfs *ZFS) ScanSnapshots() (ZFSSnapshots, error) {
	out, err := zfs.execZFS("list -t snapshot -r -o name,creation -H", zfs.Name)
	if err != nil {
		return nil, errors.New(out)
	}

	snapshots := ZFSSnapshots{}
	for _, line := range strings.Split(out, "\n") {
		// extract fields
		fields := strings.SplitN(line, "\t", 2)
		snapName := lastElement(fields[0], "@")
		creation := fields[1]

		// path
		path := zfs.MountPoint + "/.zfs/snapshot/" + snapName

		// append new snap to snapshots
		snap := ZFSSnapshot{snapName, creation, path}
		snapshots = append(snapshots, snap)
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

func (s *ZFSSnapshots) FilterWhereFileWasModified(path string, compareFunc CompareFileFunc) ZFSSnapshots {
	fh, _ := NewFileHandle(path)
	return s.Filter(func(snap ZFSSnapshot) bool {
		// ignore errors here if file not found (e.g. was deleted)
		if snapFileFh, err := NewFileHandleInSnapshot(path, snap.Name); err == nil {
			if compareFunc(snapFileFh, fh) {
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
