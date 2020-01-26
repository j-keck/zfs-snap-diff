package zfs

import (
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
	"time"
)

// Snapshot - zfs snapshot
type Snapshot struct {
	Name    string       `json:"name"`
	Created time.Time    `json:"created"`
	Dir     fs.DirHandle `json:"dir"`
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

// // FilterWhereFileWasModified finds all snapshots where the file was modified
// func (s *Snapshots) FilterWhereFileWasModified(path string, fileHasChangedFuncGen file.FileHasChangedFuncGen) Snapshots {
//	//	fh, _ := file.NewFileHandle(path)
//	//	fileHasChangedFunc := fileHasChangedFuncGen(fh)

//	return s.Filter(func(snap Snapshot) bool {
//		// ignore errors here if file not found (e.g. was deleted)
//		// if snapFileFh, err := file.NewFileHandleInSnapshot(path, snap.Name); err == nil {
//		//	if fileHasChangedFunc(fh, snapFileFh) {
//		//		// file changed in snapshot
//		//		fh = snapFileFh
//		//		return true
//		//	}
//		// }
//		return false
//	})
// }
