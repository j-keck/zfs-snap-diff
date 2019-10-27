package zfs

import (
	"github.com/j-keck/zfs-snap-diff/pkg/file"
	"sort"
	"testing"
)

func TestScanSnapshots(t *testing.T) {
	out := `tank/fs1@one	1
tank/fs1@two	2
tank/fs1@three	3`

	ds := new(Dataset)
	ds.Name = "tank"
	ds.Path = "/tank"
	ds.cmd = NewZFSCmdMock(out, nil)

	snaps, err := ds.ScanSnapshots()
	if err != nil {
		t.Error(err)
	}

	expected := 3
	if len(snaps) != expected {
		t.Errorf("%d snapshots found - expected %d", len(snaps), expected)
	}
}

func TestFindOtherFileVersions(t *testing.T) {
	out := `tank/@s01	1
tank@s02	2
tank@s03	3
tank@s04	4
tank@s05	5`

	actual, err := file.NewFileHandle("testdata/tank/testfile")
	if err != nil {
		t.Error(err)
		return
	}

	cmp, err := file.NewComparator("md5", actual)
	if err != nil {
		t.Error(err)
		return
	}

	ds := new(Dataset)
	ds.Name = "tank"
	ds.Path = "testdata/tank"
	ds.cmd = NewZFSCmdMock(out, nil)

	versions, err := ds.FindFileVersions(cmp, actual)
	if err != nil {
		t.Error(err)
		return
	}

	const expected = 2
	if len(versions) != expected {
		t.Errorf("found %d versions, expected: %d", len(versions), expected)
	}

	// extract the snapshot names and verify them
	var snapNames []string
	for _, v := range versions {
		snapNames = append(snapNames, v.Snapshot.Name)
	}
	sort.Strings(snapNames)

	if snapNames[0] != "s02" {
		t.Error("expected the first version in snapshot 's02'")
	}

	if snapNames[1] != "s05" {
		t.Error("expected the second version in snapshot 's05'")
	}
}
