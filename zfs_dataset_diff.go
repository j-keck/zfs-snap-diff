package main

import (
	"errors"
	"fmt"
	"strings"
)

// ZFSDiffs a diffs from a zfs snapshot
type ZFSDiffs []ZFSDiff

// ScanDiffs scan zfs differences from the given snapshot to the current filesystem state
func (ds *ZFSDataset) ScanDiffs(snapName string) (ZFSDiffs, error) {
	// HINT: process uid needs 'zfs allow -u <USER> diff <ZFS_NAME>'
	fullSnapName := fmt.Sprintf("%s@%s", ds.Name, snapName)
	out, err := ds.execZFS("diff -H -F", fullSnapName, ds.Name)
	if err != nil {
		return nil, errors.New(out)
	}

	diffs := ZFSDiffs{}
	for _, line := range strings.Split(out, "\n") {
		//FIXME: filter only files, directories?
		//FIXME: type rename: '/' -> 'D' ...
		if change, changeType, path, ok := split3(line, "\t"); ok {
			// replace '\040' with ' ' in 'zfs diff' output
			//   see: https://www.illumos.org/issues/1912
			path = strings.Replace(path, "\\040", " ", -1)

			diffs = append(diffs, ZFSDiff{change, changeType, path})
		}
	}
	return diffs, nil
}

// ZFSDiff is a single zfs differences entry
type ZFSDiff struct {
	Change string
	Type   string
	Path   string
}
