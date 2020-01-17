package scanner

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewDateRange(t *testing.T) {
	expectedFrom := date(2019, 1, 1)
	to := date(2019, 1, 2)

	dr := NewDateRange(to, 1)
	if dr.From != expectedFrom {
		t.Errorf("Unexpected 'From': %s, expected: %s",
			dr.From, expectedFrom)
	}
}

func TestUnmarshalWithTo(t *testing.T) {
	var dr DateRange
	err := json.Unmarshal([]byte(`{"to": "2019-02-03", "days": 1}`), &dr)
	if err != nil {
		t.Error(err)
	}

	if dr.From != date(2019, 2, 2) {
		t.Errorf("Unexpected From: %s", dr.From)
	}
}

func TestUnmarshalWithFrom(t *testing.T) {
	var dr DateRange
	err := json.Unmarshal([]byte(`{"from": "2019-02-03", "days": 1}`), &dr)
	if err != nil {
		t.Error(err)
	}

	if dr.To != date(2019, 2, 4) {
		t.Errorf("Unexpected To: %s", dr.To)
	}
}


func TestJSON(t *testing.T) {
	dr := NewDateRange(time.Now(), 1)
	js, err := json.Marshal(dr)
	if err != nil {
		t.Error(err)
	}

	var dr2 DateRange
	err = json.Unmarshal(js, &dr2)
	if err != nil {
		t.Error(err)
	}

	if dr != dr2 {
		t.Errorf("instances were differnt:\n    %+v\n    %+v", dr, dr2)
	}
}
func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
