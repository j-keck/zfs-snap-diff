package scanner

import (
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"path"
	"strings"
	"time"
)

var log = plog.GlobalLogger()

type Scanner struct {
	dateRange     DateRange
	compareMethod string
	dataset       zfs.Dataset
}


type ScanResult struct {
	FileVersions             []FileVersion  `json:"fileVersions"`
	ScannedDateRange         DateRange      `json:"scannedDateRange"`
	SkippedSnapshots         int            `json:"skippedSnapshots"`
	ScannedSnapshots         int            `json:"scannedSnapshots"`
	SnapshotsToScan          int            `json:"snapshotsToScan"`
	FileMissingSnapshots     int            `json:"fileMissingSnapshots"`
	LastScannedSnapshot      zfs.Snapshot   `json:"lastScannedSnapshot"`
	ScanDuration             time.Duration  `json:"scanDuration"`
}

type FileVersion struct {
	File     fs.FileHandle `json:"file"`
	Snapshot zfs.Snapshot  `json:"snapshot"`
}

func NewScanner(dateRange DateRange, compareMethod string, dataset zfs.Dataset) Scanner {
	return Scanner { dateRange, compareMethod, dataset }
}


func (self *Scanner) FindFileVersions(pathActualVersion string) (ScanResult, error) {
	sr := ScanResult{FileVersions: make([]FileVersion, 0), ScannedDateRange: self.dateRange}
	startTs := time.Now()

	snaps, err := self.dataset.ScanSnapshots()
	if err != nil {
		return ScanResult{}, err
	}

	log.Debugf("search for file versions for file: %s, in the date range: %s",
		pathActualVersion, self.dateRange.String())
	var cmp Comparator
	for idx, snap := range snaps {
        if self.dateRange.IsBefore(snap.Created) {
			sr.SkippedSnapshots = sr.SkippedSnapshots + 1
			log.Debugf("skip snapshot - snapshot is younger (%s) than the time-range: %s",
				snap.Created, self.dateRange.String())

			continue
		}

		if cmp == nil {
			// init comparator

			var pathInitVersion string
			if p, ok := self.findLastPathInSnap(pathActualVersion, idx - 1, snaps); ok {
				pathInitVersion = p
			} else {
				pathInitVersion = pathActualVersion
			}

			fh, err := fs.NewFileHandle(pathInitVersion)
			if err != nil {
				return sr, err
			}

			cmp, err = NewComparator(self.compareMethod, fh)
			if err != nil {
				return sr, err
			}
		}

		if self.dateRange.IsAfter(snap.Created) {
			log.Debugf("abort search - snapshot is older (%s) than the time-range %s",
				snap.Created, self.dateRange.String())
			break
		}

		sr.ScannedSnapshots = sr.ScannedSnapshots + 1
		sr.LastScannedSnapshot = snap

		fh, err := fs.NewFileHandle(self.pathInSnapshot(pathActualVersion, snap))
		if err != nil {
			// not every snapshot has a version of the file - ignore errors
            sr.FileMissingSnapshots = sr.FileMissingSnapshots + 1
			continue
		}

		log.Tracef("check if file was changed under path: %s", fh.Path)
		if cmp.HasChanged(fh) {
			log.Debugf("file was changed in snapshot: %s", fh.Path)
			sr.FileVersions = append(sr.FileVersions, FileVersion{fh, snap})
		}
	}

	sr.ScanDuration = time.Now().Sub(startTs)
	sr.SnapshotsToScan = len(snaps) - sr.SkippedSnapshots - sr.ScannedSnapshots

	log.Debugf("%d versions for file %s found - scan duration: %s",
		len(sr.FileVersions), pathActualVersion, sr.ScanDuration)
	return sr, nil
}

func (self *Scanner) pathInSnapshot(pathActualVersion string, snap zfs.Snapshot) string {
	p := strings.TrimPrefix(pathActualVersion, self.dataset.MountPoint.Path)
	return path.Join(snap.Path, p)
}

func (self *Scanner) findLastPathInSnap(p string, idx int, snaps []zfs.Snapshot) (string, bool) {
	for idx >= 0 {
	pathInSnap := self.pathInSnapshot(p, snaps[idx])
		if _, err := fs.NewFileHandle(pathInSnap); err == nil {
			return pathInSnap, true
		}
		idx = idx - 1
	}
	return "", false
}
