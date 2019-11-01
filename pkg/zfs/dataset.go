package zfs

import (
	"errors"
	"github.com/j-keck/zfs-snap-diff/pkg/comparator"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"path"
	"strconv"
	"strings"
	"time"
)

// Dataset represents a zfs dataset (aka. zfs filesystem)
type Dataset struct {
	Name       string       `json:"name"`
	Used       uint64       `json:"used"`
	Avail      uint64       `json:"avail"`
	Refer      uint64       `json:"refer"`
	MountPoint fs.DirHandle `json:"mountPoint"`
	cmd        ZFSCmd
}

// ScanSnapshots returns a list of all snapshots for this dataset
func (self *Dataset) ScanSnapshots() (Snapshots, error) {
	stdout, stderr, err := self.cmd.Exec("list -t snapshot -s creation -r -d 1 -o name,creation -Hp", self.Name)
	if err != nil {
		return nil, errors.New(stderr)
	}

	parse := func(s string) (string, time.Time, bool) {
		const n = 2
		fields := strings.SplitN(s, "\t", n)
		if len(fields) == n {
			n, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				log.Errorf("unable to convert '%s' to a number: %s", fields[1], err.Error())
				return "", time.Unix(0, 0), false
			}
			return fields[0], time.Unix(n, 0), true
		} else {
			return "", time.Unix(0, 0), false
		}
	}

	snapshots := Snapshots{}
	for _, line := range strings.Split(stdout, "\n") {
		if snapName, creation, ok := parse(line); ok {
			// remove dataset name from snapshot
			fields := strings.Split(snapName, "@")
			snapName := fields[len(fields)-1]

			// path
			path := self.MountPoint.Path + "/.zfs/snapshot/" + snapName

			// append new snap to snapshots
			snapshots = append(snapshots, Snapshot{snapName, creation, path})
		}

	}
	return snapshots.Reverse(), nil
}

// func (self *Dataset) OpenFile(path string) (file.FileHandle, error) {
//	// FIXME: validate path is under the given mountpoint
//	return file.NewFileHandle(path)
// }

type FileVersion struct {
	File     fs.FileHandle
	Snapshot Snapshot
}

func (self *Dataset) FindFileVersions(comparator comparator.Comparator, fh fs.FileHandle) ([]FileVersion, error) {
	snaps, err := self.ScanSnapshots()
	if err != nil {
		return nil, err
	}

	var versions = make([]FileVersion, 0)
	for _, snap := range snaps {
		relPath := strings.TrimPrefix(fh.Path, self.MountPoint.Path)
		fhInSnap, err := fs.NewFileHandle(path.Join(snap.Path, relPath))
		// not ever snapshot has a version of the file - ignore errors
		if err != nil {
			continue
		}

		log.Tracef("check if file was changed under path: %s", fhInSnap.Path)
		if comparator.HasChanged(fhInSnap) {
			log.Debugf("file was changed in snapshot: %s", fhInSnap.Path)
			versions = append(versions, FileVersion{fhInSnap, snap})
		}
	}
	log.Tracef("versions for file: %+v - %+v", fh, versions)
	return versions, nil
}

// Datasets are a list of Dataset
type Datasets []Dataset

// Root returns the parent dataset
func (ds Datasets) Root() *Dataset {
	return &ds[0]
}
