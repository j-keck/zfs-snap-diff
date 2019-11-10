package scanner

import (
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"path"
	"strings"
)

var log = plog.GlobalLogger()


type Scanner struct {
	ActualFh   fs.FileHandle
	TimeRange  TimeRange
	Comparator Comparator
}

func NewScanner(tr TimeRange, compareMethod string, fh fs.FileHandle) (Scanner, error) {
	cmp, err := NewComparator(compareMethod, fh)
	if err != nil {
		return Scanner{}, err
	}

	return Scanner{ fh, tr, cmp }, nil
}

type FileVersion struct {
	File     fs.FileHandle `json:"file"`
	Snapshot zfs.Snapshot  `json:"snapshot"`
}

func (self *Scanner) FindFileVersions(ds zfs.Dataset) ([]FileVersion, error) {
	snaps, err := ds.ScanSnapshots()
	if err != nil {
		return nil, err
	}

	var versions = make([]FileVersion, 0)
	for _, snap := range snaps {
		if ! self.TimeRange.Contains(snap.Created) {
			log.Tracef("skip snapshot because it's not in the time range: %+v", snap)
			continue
		}

		relPath := strings.TrimPrefix(self.ActualFh.Path, ds.MountPoint.Path)
		fhInSnap, err := fs.NewFileHandle(path.Join(snap.Path, relPath))
		// not every snapshot has a version of the file - ignore errors
		if err != nil {
			continue
		}

		log.Tracef("check if file was changed under path: %s", fhInSnap.Path)
		if self.Comparator.HasChanged(fhInSnap) {
			log.Debugf("file was changed in snapshot: %s", fhInSnap.Path)
			versions = append(versions, FileVersion{fhInSnap, snap})
		}
	}
	log.Tracef("versions for file: %+v - %+v", self.ActualFh, versions)
	return versions, nil
}
