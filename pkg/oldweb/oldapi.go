package oldweb

import (
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"sort"
	"strings"
)

// FindDatasetForFile searches and returns the dataset where the given file lives
func findDatasetForFile(path string) (zfs.Dataset, error) {

	datasets := z.Datasets()

	// sort the datasets - longest path at first
	sort.Sort(SortByMountPointDesc(datasets))

	for _, ds := range datasets {
		if strings.HasPrefix(path, ds.MountPoint.Path) {
			return ds, nil
		}
	}
	return zfs.Dataset{}, fmt.Errorf("Path '%s' is not in a Dataset", path)
}

func PathIsInSnapshot(dsPath, filePath string) bool {
	return strings.HasPrefix(filePath, dsPath+"/.zfs/snapshot")
}

// ExtractSnapName extracts the snapshot name from a given path
func ExtractSnapName(dsPath, filePath string) string {
	s := strings.TrimPrefix(filePath, dsPath)
	s = strings.TrimPrefix(s, "/.zfs/snapshot/")
	fields := strings.Split(s, "/")
	return fields[0]
}

// UniqueName returns the unique file name
//   * the file name if the file is in the actual filesystem
//   * <FILE-NAME>-<SNAP-NAME>.<SUFFIX> if the file is from a snapshot
func UniqueName(fh fs.FileHandle) string {
	ds, _ := findDatasetForFile(fh.Path)
	if PathIsInSnapshot(ds.MountPoint.Path, fh.Path) {
		snapName := ExtractSnapName(ds.MountPoint.Path, fh.Path)

		// build unique-name
		if strings.Contains(fh.Name, ".") {
			f := strings.Split(fh.Name, ".")
			return fmt.Sprintf("%s-%s.%s", f[0], snapName, f[1])
		}
		return fmt.Sprintf("%s-%s", fh.Name, snapName)
	}

	return fh.Name
}

func NewFileHandleInSnapshot(path, snapName string) (fs.FileHandle, error) {
	ds, _ := findDatasetForFile(path)
	relativePath := strings.TrimPrefix(path, ds.MountPoint.Path)
	pathInSnap := fmt.Sprintf("%s/.zfs/snapshot/%s%s", ds.MountPoint.Path, snapName, relativePath)
	return fs.NewFileHandle(pathInSnap)
}

// SortByMountPointDesc implments sort.Interface for Datasets based on the mount point
type SortByMountPointDesc zfs.Datasets

func (s SortByMountPointDesc) Len() int {
	return len(s)
}

func (s SortByMountPointDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]

}

func (s SortByMountPointDesc) Less(i, j int) bool {
	return len(s[i].MountPoint.Path) > len(s[j].MountPoint.Path)
}
