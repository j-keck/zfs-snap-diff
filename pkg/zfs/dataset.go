package zfs

import (
	"errors"
	"github.com/j-keck/zfs-snap-diff/pkg/file"
	"path"
	"strconv"
	"strings"
	"time"
)

// Dataset represents a zfs dataset (aka. zfs filesystem)
type Dataset struct {
	Name  string
	Used  uint64
	Avail uint64
	Refer uint64
	file.DirEntry
	cmd ZFSCmd
}

// ScanSnapshots returns a list of all snapshots for this dataset
func (self *Dataset) ScanSnapshots() (Snapshots, error) {
	out, err := self.cmd.exec("list -t snapshot -s creation -r -d 1 -o name,creation -Hp", self.Name)
	if err != nil {
		return nil, errors.New(out)
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
	for _, line := range strings.Split(out, "\n") {
		if snapName, creation, ok := parse(line); ok {
			// remove dataset name from snapshot
			fields := strings.Split(snapName, "@")
			snapName := fields[len(fields)-1]

			// path
			path := self.Path + "/.zfs/snapshot/" + snapName

			// append new snap to snapshots
			snapshots = append(snapshots, Snapshot{snapName, creation, path})
		}

	}
	return snapshots.Reverse(), nil
}

type FileVersion struct {
	File     file.FileHandle
	Snapshot Snapshot
}

func (self *Dataset) FindFileVersions(comparator file.Comparator, fh file.FileHandle) ([]FileVersion, error) {
	snaps, err := self.ScanSnapshots()
	if err != nil {
		return nil, err
	}

	var versions []FileVersion
	for _, snap := range snaps {
		relPath := strings.TrimPrefix(fh.Path, self.Path)
		fhInSnap, err := file.NewFileHandle(path.Join(snap.Path, relPath))
		// not ever snapshot has a version of the file - ignore errors
		if err != nil {
			continue
		}

		log.Tracef("check if file was changed under path: %s", fhInSnap.Path)
		if comparator.HasChanged(fhInSnap) {
			log.Tracef("file was changed")
			versions = append(versions, FileVersion{fhInSnap, snap})
		}
	}
	return versions, nil
}

// Datasets are a list of Dataset
type Datasets []Dataset

// Root returns the parent dataset
func (ds Datasets) Root() *Dataset {
	return &ds[0]
}
