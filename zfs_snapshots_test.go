package main

import (
	"testing"
)

var snapshots = ZFSSnapshots{
	ZFSSnapshot{"snap1", "t1"},
	ZFSSnapshot{"snap2", "t2"},
	ZFSSnapshot{"SNAP3", "T3"},
}

func TestReverse(t *testing.T) {
	if snapshots.Reverse()[0].Name != "SNAP3" {
		t.Error("SNAP3 expected")
	}
}

func TestFilter(t *testing.T) {
	filtered := snapshots.Filter(func(snap ZFSSnapshot) bool {
		return snap.Name == "snap2"
	})

	if len(filtered) != 1 {
		t.Error("expected len(filtered) == 1")
	}
}
