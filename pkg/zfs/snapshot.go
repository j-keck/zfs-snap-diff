package zfs

import (
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"os"
	"time"
)

// Snapshot - zfs snapshot
type Snapshot struct {
	Name       string       `json:"name"`
	FullName   string       `json:"fullName"`
	Created    time.Time    `json:"created"`
	MountPoint fs.DirHandle `json:"mountPoint"`
}

// Check if the snaphot is mounted
func (s *Snapshot) IsMounted() (bool, error) {
	log.Tracef("check if snapshot: %s is mounted", s.Name)

	// to check if the snapshot is mounted, list
	// the directory content. use 'File.Readdirnames'
	// for the directrory listing, because it has less overhead
	fh, err := os.Open(s.MountPoint.Path)
	if err != nil {
		return false, err
	}

	names, _ := fh.Readdirnames(10)
	isMounted := len(names) > 0
	log.Tracef("snapshot: %s is mounted = %v", s.Name, isMounted)
	return isMounted, nil
}

// Snapshots represents snapshots from a zfs dataset
type Snapshots []Snapshot

// Reverse reverse the snapshot list
func (s Snapshots) Reverse() Snapshots {
	reversed := Snapshots{}
	for i := len(s) - 1; i >= 0; i-- {
		reversed = append(reversed, s[i])
	}
	return reversed
}

// Filter filters snapshots per given filter function
func (s *Snapshots) Filter(f func(Snapshot) bool) Snapshots {
	newS := Snapshots{}
	for _, snap := range *s {
		if f(snap) {
			newS = append(newS, snap)
		}
	}
	return newS
}
