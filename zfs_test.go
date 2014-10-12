package main

import (
	"testing"
)

func TestFindDatasetForFile(t *testing.T) {
	datasets := ZFSDatasets{
		ZFSDataset{"zp1", "1k", "2k", "3k", "/zp1", nil},
		ZFSDataset{"zp1/a", "1k", "2k", "3k", "/zp1/a", nil},
		ZFSDataset{"zp1/aa", "1k", "2k", "3k", "/", nil},
	}

	zfs := ZFS{datasets, nil}
	dataset := zfs.FindDatasetForFile("/zp1/aa/file")
	if dataset.Name != "zp1" {
		t.Fatal("unexpected dataset", dataset)
	}
}
