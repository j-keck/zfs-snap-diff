package main

import (
	"fmt"
	"strings"
)

// ZFSDataset represents a zfs dataset (aka. zfs filesystem)
type ZFSDataset struct {
	Name       string
	Used       string
	Avail      string
	Refer      string
	MountPoint string
	execZFS    execZFSFunc
}

// ConvertToActualPath converts a given path in the snapshot to a path in the actual dataset
func (d *ZFSDataset) ConvertToActualPath(path string) string {
	p := strings.TrimPrefix(path, d.MountPoint) // remove mount point
	p = strings.TrimPrefix(p, "/.zfs/snapshot") // remove zfs ctrl dir
	pathInActual := fmt.Sprintf("%s/%s", d.MountPoint, p)
	return pathInActual
}

// ConvertToSnapPath converts a path in the actual dataset to a path in a snapshot
func (d *ZFSDataset) ConvertToSnapPath(path, snapName string) string {
	relativePath := strings.TrimPrefix(path, d.MountPoint)
	pathInSnap := fmt.Sprintf("%s/.zfs/snapshot/%s%s", d.MountPoint, snapName, relativePath)
	return pathInSnap
}

// PathIsInSnapshot returns true if the given path is in a snapshot
func (d *ZFSDataset) PathIsInSnapshot(path string) bool {
	return strings.HasPrefix(path, d.MountPoint+"/.zfs/snapshot")
}

// ExtractSnapName extracts the snapshot name from a given path
func (d *ZFSDataset) ExtractSnapName(path string) string {
	s := strings.TrimPrefix(path, d.MountPoint)
	s = strings.TrimPrefix(s, "/.zfs/snapshot/")
	return firstElement(s, "/")
}

// ZFSDatasets are a list of ZFSDataset
type ZFSDatasets []ZFSDataset

// ScanDatasets returns all datasets under a given pool name
func ScanDatasets(name string, execZFS execZFSFunc) (ZFSDatasets, error) {
	logDebug.Printf("search datasets under zfs: %s\n", name)

	out, err := execZFS("list -H -o name,used,avail,refer,mountpoint -r -t filesystem", name)
	if err != nil {
		return nil, err
	}

	var datasets ZFSDatasets
	for _, line := range strings.Split(out, "\n") {
		if name, used, avail, refer, mountPoint, ok := split5(line, "\t"); ok {
			// don't add legacy datasets
			if mountPoint != "legacy" {
				datasets = append(datasets, ZFSDataset{name, used, avail, refer, mountPoint, execZFS})
			}
		}

	}
	logDebug.Printf("%d datasets found\n", len(datasets))
	return datasets, nil
}

// Root returns the parent dataset
func (ds ZFSDatasets) Root() *ZFSDataset {
	return &ds[0]
}

// SortByMountPointDesc implments sort.Interface for ZFSDatasets based on the mount point
type SortByMountPointDesc ZFSDatasets

func (s SortByMountPointDesc) Len() int {
	return len(s)
}

func (s SortByMountPointDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]

}

func (s SortByMountPointDesc) Less(i, j int) bool {
	return len(s[i].MountPoint) > len(s[j].MountPoint)
}
