package main

import (
	"fmt"
	"strings"
)

type ZFSDataset struct {
	Name       string
	MountPoint string
	execZFS    execZFSFunc
}

func (d *ZFSDataset) ConvertToActualPath(path string) string {
	p := strings.TrimPrefix(path, d.MountPoint) // remove mount point
	p = strings.TrimPrefix(p, "/.zfs/snapshot") // remove zfs ctrl dir
	pathInActual := fmt.Sprintf("%s/%s", d.MountPoint, p)
	return pathInActual
}
func (d *ZFSDataset) ConvertToSnapPath(path, snapName string) string {
	relativePath := strings.TrimPrefix(path, d.MountPoint)
	pathInSnap := fmt.Sprintf("%s/.zfs/snapshot/%s%s", d.MountPoint, snapName, relativePath)
	return pathInSnap
}
func (d *ZFSDataset) PathIsInSnapshot(path string) bool {
	return strings.HasPrefix(path, d.MountPoint+"/.zfs/snapshot")
}

func (d *ZFSDataset) ExtractSnapName(path string) string {
	s := strings.TrimPrefix(path, d.MountPoint)
	s = strings.TrimPrefix(s, "/.zfs/snapshot/")
	return firstElement(s, "/")
}

type ZFSDatasets []ZFSDataset

func ScanDatasets(name string, execZFS execZFSFunc) (ZFSDatasets, error) {
	logDebug.Printf("search datasets under zfs: %s\n", name)
	var datasets ZFSDatasets
	if out, err := execZFS("list -H -o name,mountpoint -r -t filesystem", name); err != nil {
		return nil, err
	} else {
		for _, line := range strings.Split(out, "\n") {
			if name, mountPoint, ok := split2(line, "\t"); ok {
				// don't add legacy datasets
				if mountPoint != "legacy" {
					datasets = append(datasets, ZFSDataset{name, mountPoint, execZFS})
				}
			}
		}
	}
	logDebug.Printf("%d datasets found\n", len(datasets))
	return datasets, nil
}

func (ds ZFSDatasets) Root() *ZFSDataset {
	return &ds[0]
}

// Sort
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
