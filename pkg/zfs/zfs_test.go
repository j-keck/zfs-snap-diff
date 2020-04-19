package zfs

import (
	"testing"
)

func TestScanDatasets(t *testing.T) {
	out := `tank	1	2	3	testdata/tank
tank/sub	1	2	3	testdata/tank/subpool
`
	zfs := new(ZFS)
	zfs.cmd = NewZFSCmdMock(out, "", nil)
	ds, _, err := zfs.scanDatasets("tank")
	if err != nil {
		t.Error(err)
	}

	expected := 2
	if len(ds) != expected {
		t.Errorf("%d datasets found - expected %d", len(ds), expected)
	}
}
