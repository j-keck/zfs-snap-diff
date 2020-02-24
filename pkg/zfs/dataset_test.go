package zfs

import (
	"testing"
)

func TestScanSnapshots(t *testing.T) {
	out := `tank/fs1@one	1
tank/fs1@two	2
tank/fs1@three	3`

	ds := new(Dataset)
	ds.Name = "tank"
	ds.cmd = NewZFSCmdMock(out, "", nil)

	snaps, err := ds.ScanSnapshots()
	if err != nil {
		t.Error(err)
	}

	expected := 3
	if len(snaps) != expected {
		t.Errorf("%d snapshots found - expected %d", len(snaps), expected)
	}
}
