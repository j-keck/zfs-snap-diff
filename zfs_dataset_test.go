package main

import (
	"testing"
)

func TestScanDatasets(t *testing.T) {
	initLogHandlersForTest()

	out := zfsOutput("zp1,1k,1k,1k,/", "zp1/c1,1k,1k,1k,/zp1/c1", "zp1/c2,1k,1k,1k,/zp/c2", "zp1/tmp,1k,1k,1k,/tmp")
	datasets, _ := ScanDatasets("zp", execZFSMock(out, nil))
	if len(datasets) != 4 {
		t.Error("4 datasets expected - received:", len(datasets))
	}
}

func TestScanDatasetsShouldFilterLegacy(t *testing.T) {
	initLogHandlersForTest()

	out := zfsOutput("zp,1k,1k,1k,/zp", "zp1/a,1k,1k,1k,legacy", "zp1/b,1k,1k,1k,legacy")
	datasets, _ := ScanDatasets("zp", execZFSMock(out, nil))
	if len(datasets) != 1 {
		t.Error("1 datasets expected - received:", len(datasets))
	}
}
