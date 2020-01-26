package scanner

import (
	"testing"
	"time"
)

func TestTimeRangeContains(t *testing.T) {
	tr, err := ParseTimeRange("15:04", "10:00", "11:00")
	if err != nil {
		t.Error(err.Error())
	}

	if !tr.Contains(tr.From.Add(1 * time.Second)) {
		t.Error("expected time is in range")
	}

	if tr.Contains(tr.From.Add(10 * time.Hour)) {
		t.Error("expected time is NOT in range")
	}

}
