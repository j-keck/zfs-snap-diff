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
