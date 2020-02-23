package scanner

import (
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
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
	zfs           zfs.ZFS
}

type ScanResult struct {
	FileVersions        []FileVersion `json:"fileVersions"`
	DateRange           DateRange     `json:"dateRange"`
	SnapsScanned        int           `json:"snapsScanned"`
	SnapsToScan         int           `json:"snapsToScan"`
	SnapsFileMissing    int           `json:"snapsFileMissing"`
	LastScannedSnapshot zfs.Snapshot  `json:"lastScannedSnapshot"`
	ScanDuration        time.Duration `json:"scanDuration"`
}

type FileVersion struct {
	Current  fs.FileHandle `json:"current"`
	Backup   fs.FileHandle `json:"backup"`
	Snapshot zfs.Snapshot  `json:"snapshot"`
}

func NewScanner(dateRange DateRange, compareMethod string, dataset zfs.Dataset, zfs zfs.ZFS) Scanner {
	return Scanner{dateRange, compareMethod, dataset, zfs}
}

func (self *Scanner) FindFileVersions(pathCurrentVersion string) (ScanResult, error) {
	sr := ScanResult{FileVersions: make([]FileVersion, 0), DateRange: self.dateRange}
	startTs := time.Now()

	currentVersionFh, err := fs.GetFileHandle(pathCurrentVersion)
	if err != nil {
		return ScanResult{}, err
	}

	snaps, err := self.dataset.ScanSnapshots()
	if err != nil {
		return ScanResult{}, err
	}

	log.Debugf("search for file versions for file: %s, in the date range: %s",
		pathCurrentVersion, self.dateRange.String())
	var cmp Comparator
	snapsSkipped := 0
	for idx, snap := range snaps {

		// search is data-range based - check if the current checked snapshot
		// was created in the given range
		if self.dateRange.IsBefore(snap.Created) {
			snapsSkipped = snapsSkipped + 1
			log.Tracef("skip snapshot - snapshot is younger (%s) than the time-range: %s",
				snap.Created, self.dateRange.String())

			continue
		}

		if self.dateRange.IsAfter(snap.Created) {
			log.Debugf("abort search - snapshot is older (%s) than the time-range %s",
				snap.Created, self.dateRange.String())
			break
		}

		// mount the snapshot if necessary
		if config.Get.ZFS.MountSnapshots {
			isMounted, err := snap.IsMounted()
			if err != nil {
				log.Errorf("unable to check if snapshot: %s is mounted - %v", snap.Name, err)
			}

			if !isMounted {
				if err := self.zfs.MountSnapshot(snap); err != nil {
					log.Errorf("unable to mount snapshot: %s - %v", snap.Name, err)

					// skip this snapshot
					continue
				}
			}
		}

		// initialize the file-content comparator
		if cmp == nil {

			var pathInitVersion string
			if p, ok := self.findLastPathInSnap(pathCurrentVersion, idx-1, snaps); ok {
				pathInitVersion = p
			} else {
				pathInitVersion = pathCurrentVersion
			}

			fh, err := fs.GetFileHandle(pathInitVersion)
			if err != nil {
				return sr, err
			}

			cmp, err = NewComparator(self.compareMethod, fh)
			if err != nil {
				return sr, err
			}
		}

		// get the file-handle to the backup version in the snapshot
		fh, err := fs.GetFileHandle(self.pathInSnapshot(pathCurrentVersion, snap))
		if err != nil {
			// not every snapshot MUST have a version of the file.
			// maybe the file was deleted and restored - so ignore the error
			sr.SnapsFileMissing = sr.SnapsFileMissing + 1
			continue
		}

		// compare the file content
		log.Tracef("check if file was changed under path: %s", fh.Path)
		if cmp.HasChanged(fh) {
			log.Debugf("file was changed in snapshot: %s", fh.Path)
			sr.FileVersions = append(sr.FileVersions, FileVersion{currentVersionFh, fh, snap})
		}

		// update stats
		sr.SnapsScanned = sr.SnapsScanned + 1
		sr.LastScannedSnapshot = snap
	}

	sr.ScanDuration = time.Now().Sub(startTs)
	sr.SnapsToScan = len(snaps) - snapsSkipped - sr.SnapsScanned

	log.Debugf("%d versions for file %s found - scan duration: %s",
		len(sr.FileVersions), pathCurrentVersion, sr.ScanDuration)
	return sr, nil
}

func (self *Scanner) pathInSnapshot(pathCurrentVersion string, snap zfs.Snapshot) string {
	p := strings.TrimPrefix(pathCurrentVersion, self.dataset.MountPoint.Path)
	return path.Join(snap.MountPoint.Path, p)
}

func (self *Scanner) findLastPathInSnap(p string, idx int, snaps []zfs.Snapshot) (string, bool) {
	for idx >= 0 {
		pathInSnap := self.pathInSnapshot(p, snaps[idx])
		if _, err := fs.GetFileHandle(pathInSnap); err == nil {
			return pathInSnap, true
		}
		idx = idx - 1
	}
	return "", false
}
