package main

import (
	"testing"
)

func TestScanDatasets(t *testing.T) {
	initLogHandlersForTest()

	datasets, _ := ScanDatasets("zp", execZFSMock("zp\t/zp\nzp1/c1\t/zp1/c1\nzp1/c2\t/zp/c2\nzp1/tmp\t/tmp\n", nil))
	if len(datasets) != 4 {
		t.Error("4 datasets expected - received:", len(datasets))
	}
}

func TestScanDatasetsShouldFilterLegacy(t *testing.T) {
	initLogHandlersForTest()

	datasets, _ := ScanDatasets("zp", execZFSMock("zp\t/zp\nzp1/a\tlegacy\nzp1/b\tlegacy\n", nil))
	if len(datasets) != 1 {
		t.Error("1 datasets expected - received:", len(datasets))
	}
}
