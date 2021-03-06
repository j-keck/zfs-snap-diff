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
				Name:  name,
				Path:  self.MountPoint.Path + "/.zfs/snapshot/" + name,
				Kind:  fs.DIR,
				Size:  0,
				MTime: creation,
			}}

			// append new snap to snapshots
			snapshots = append(snapshots, Snapshot{name, fullName, creation, dir})
		}

	}
	return snapshots.Reverse(), nil
}

func (self *Dataset) CreateSnapshot(name string) (string, error) {
	if len(name) == 0 {
		return "", errors.New("snapshot-name can't be empty")
	}

	if !strings.HasPrefix(name, self.Name) {
		name = self.Name + "@" + name
	}

	log.Debugf("create snapshot: %s", name)
	stdout, stderr, err := self.cmd.Exec("snapshot", name)
	log.Tracef("create snapshot stdout: %s", stdout)
	log.Tracef("create snapshot stderr: %s", stderr)
	return name, err
}

func (self *Dataset) CloneSnapshot(snapName, fsName string, flags []string) error {
	if len(fsName) == 0 {
		return errors.New("filesystem-name can't be empty")
	}

	if !strings.HasPrefix(snapName, self.Name) {
		snapName = self.Name + "@" + snapName
	}

	log.Debugf("clone snapshot: %s to %s", snapName, fsName)
	args := append(flags, snapName, fsName)
	stdout, stderr, err := self.cmd.Exec("clone", args...)
	log.Tracef("clone snapshot stdout: %s", stdout)
	log.Tracef("clone snapshot stderr: %s", stderr)
	return err
}

// FIXME: check if the given name is a snapshot name
func (self *Dataset) RenameSnapshot(oldName, newName string) error {
	if len(newName) == 0 {
		return errors.New("new snapshot-name can't be empty")
	}

	if !strings.HasPrefix(oldName, self.Name) {
		oldName = self.Name + "@" + oldName
	}

	if !strings.HasPrefix(newName, self.Name) {
		newName = self.Name + "@" + newName
	}

	log.Debugf("rename snapshot: %s -> %s", oldName, newName)
	stdout, stderr, err := self.cmd.Exec("rename", oldName, newName)
	log.Tracef("rename snapshot stdout: %s", stdout)
	log.Tracef("rename snapshot stderr: %s", stderr)
	return err
}

func (self *Dataset) DestroySnapshot(name string, flags []string) error {
	if !strings.HasPrefix(name, self.Name) {
		name = self.Name + "@" + name
	}

	log.Debugf("destroy snapshot: %s", name)
	args := append(flags, name)
	stdout, stderr, err := self.cmd.Exec("destroy", args...)
	log.Tracef("destroy snapshot stdout: %s", stdout)
	log.Tracef("destroy snapshot stderr: %s", stderr)
	return err
}

func (self *Dataset) RollbackSnapshot(name string, flags []string) error {
	if !strings.HasPrefix(name, self.Name) {
		name = self.Name + "@" + name
	}

	log.Debugf("rollback snapshot: %s", name)
	args := append(flags, name)
	stdout, stderr, err := self.cmd.Exec("rollback", args...)
	log.Tracef("rollback snapshot stdout: %s", stdout)
	log.Tracef("rollback snapshot stderr: %s", stderr)
	return err
}

// Datasets are a list of Dataset
type Datasets []Dataset

// Root returns the parent dataset
func (ds Datasets) Root() *Dataset {
	return &ds[0]
}
