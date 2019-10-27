package zfs

import (
	"testing"
	"time"
)

var snapshots = Snapshots{
	Snapshot{"snap1", time.Unix(0, 0), "/path/.zfs/snapshot/snap1"},
	Snapshot{"snap2", time.Unix(0, 0), "/path/.zfs/snapshot/snap1"},
	Snapshot{"SNAP3", time.Unix(0, 0), "/path/.zfs/snapshot/snap1"},
}

func TestReverse(t *testing.T) {
	if snapshots.Reverse()[0].Name != "SNAP3" {
		t.Error("SNAP3 expected")
	}
}

func TestFilter(t *testing.T) {
	filtered := snapshots.Filter(func(snap Snapshot) bool {
		return snap.Name == "snap2"
	})

	if len(filtered) != 1 {
		t.Error("expected len(filtered) == 1")
	}
}
