package zfs

import (
	"errors"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
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
		if fullName, creation, ok := parse(line); ok {
			// remove dataset name from snapshot
			fields := strings.Split(fullName, "@")
			name := fields[len(fields)-1]

			// create the dir-handle per hand.
			// this prevents a unnecessary 'os.Stat' call for each snapshot
			// (which takes some time with thousands snapshots on spinning disks)
			dir := fs.DirHandle{fs.FSHandle{
				Name:    name,
				Path:    self.MountPoint.Path + "/.zfs/snapshot/" + name,
				Kind:    fs.DIR,
				Size:    0,
				ModTime: creation,
			}}

			// append new snap to snapshots
			snapshots = append(snapshots, Snapshot{name, fullName, creation, dir})
		}

	}
	return snapshots.Reverse(), nil
}

// Datasets are a list of Dataset
type Datasets []Dataset

// Root returns the parent dataset
func (ds Datasets) Root() *Dataset {
	return &ds[0]
}
