package main

import (
	"testing"
)

func TestFindDatasetForFile(t *testing.T) {
	datasets := ZFSDatasets{
		ZFSDataset{"zp1", "/zp1", nil},
		ZFSDataset{"zp1/a", "/zp1/a", nil},
		ZFSDataset{"zp1/aa", "/", nil},
	}

	zfs := ZFS{datasets, nil}
	dataset := zfs.FindDatasetForFile("/zp1/aa/file")
	if dataset.Name != "zp1" {
		t.Fatal("unexpected dataset", dataset)
	}
}
