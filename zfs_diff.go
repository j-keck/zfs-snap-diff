package main

import (
	"errors"
	"fmt"
	"strings"
)

// ZFSDiffs a diffs from a zfs snapshot
type ZFSDiffs []ZFSDiff

// ScanZFSDiffs scan zfs differences from the given snapshot to the current filesystem state
func (zfs *ZFS) ScanDiffs(snapName string) (ZFSDiffs, error) {
	// HINT: process uid needs 'zfs allow -u <USER> diff <ZFS_NAME>'
	fullSnapName := fmt.Sprintf("%s@%s", zfs.Name, snapName)
	out, err := zfs.execZFS("diff -H -F", fullSnapName, zfs.Name)
	if err != nil {
		return nil, errors.New(out)
	}

	// init replacer to replace '\040' with ' ' in 'zfs diff' output
	//   see: https://www.illumos.org/issues/1912
	replacer := strings.NewReplacer("\\040", " ")

	diffs := ZFSDiffs{}
	for _, line := range strings.Split(out, "\n") {
		//FIXME: filter only files, directories?
		//FIXME: type rename: '/' -> 'D' ...
		fields := strings.Split(line, "\t")
		zfsDiff := ZFSDiff{fields[0], fields[1], replacer.Replace(fields[2])}

		diffs = append(diffs, zfsDiff)
	}
	return diffs, nil
}

// ZFSDiff is a single zfs differences entry
type ZFSDiff struct {
	Change string
	Type   string
	Path   string
}
