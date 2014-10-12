package main

import (
	"testing"
)

var snapshots = ZFSSnapshots{
	ZFSSnapshot{"snap1", "t1", "/path/.zfs/snapshot/snap1"},
	ZFSSnapshot{"snap2", "t2", "/path/.zfs/snapshot/snap1"},
	ZFSSnapshot{"SNAP3", "T3", "/path/.zfs/snapshot/snap1"},
}

func TestReverse(t *testing.T) {
	initLogHandlersForTest()
	if snapshots.Reverse()[0].Name != "SNAP3" {
		t.Error("SNAP3 expected")
	}
}

func TestFilter(t *testing.T) {
	initLogHandlersForTest()

	filtered := snapshots.Filter(func(snap ZFSSnapshot) bool {
		return snap.Name == "snap2"
	})

	if len(filtered) != 1 {
		t.Error("expected len(filtered) == 1")
	}
}

func TestScanSnapshots(t *testing.T) {
	initLogHandlersForTest()

	ds := ZFSDataset{"name", "used", "avail", "refer", "mount", execZFSMock("zfs-name@snap-name\t20140101", nil)}
	snaps, err := ds.ScanSnapshots()

	if err != nil {
		t.Error("unexpected err:", err)
	}

	if len(snaps) != 1 {
		t.Error("unexpected snaps length: ", len(snaps))
	}

	if snaps[0].Name != "snap-name" {
		t.Error("unexpected snap name: ", snaps[0].Name)
	}
}

func TestScanSnapshotsEmpty(t *testing.T) {
	initLogHandlersForTest()

	ds := ZFSDataset{"name", "used", "avail", "refer", "mount", execZFSMock("", nil)}

	snaps, err := ds.ScanSnapshots()
	if err != nil {
		t.Error("unexpected err:", err)
	}

	if len(snaps) != 0 {
		t.Error("unexpected snaps length: ", len(snaps))
	}

}
