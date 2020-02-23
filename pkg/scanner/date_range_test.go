package scanner

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewDate(t *testing.T) {
	dr, err := NewDateRange(date(2020, 1, 1), date(2019, 1, 1))
	if err == nil {
		t.Errorf("invalid DateRange constructed: %v", dr)
	}
}

func Test0DaysBack(t *testing.T) {
	expectedFrom := date(2020, 1, 1)
	dr := NDaysBack(0, date(2020, 1, 1))
	if dr.From != expectedFrom {
		t.Errorf("Unexpected 'From': %s, expected: %s",
			dr.From, expectedFrom)
	}
}

func Test1DaysBack(t *testing.T) {
	expectedFrom := date(2020, 1, 1)
	dr := NDaysBack(1, date(2020, 1, 2))
	if dr.From != expectedFrom {
		t.Errorf("Unexpected 'From': %s, expected: %s",
			dr.From, expectedFrom)
	}
}

func TestIsBefore(t *testing.T) {
	from := date(2020, 1, 1)
	to := date(2020, 1, 2)
	dr, err := NewDateRange(from, to)
	if err != nil {
		t.Errorf("%v", err)
	}

	if dr.IsBefore(from) {
		t.Errorf("Date IS in DateRange - range: %v, date: %v", dr, to)
	}

	if !dr.IsBefore(date(2020, 1, 3)) {
		t.Errorf("01-01-2020->02-01-2020 is before 03-01-2020 - but was not detected")
	}
}

func TestIsAfter(t *testing.T) {
	from := date(2020, 1, 2)
	to := date(2020, 1, 3)
	dr, err := NewDateRange(from, to)
	if err != nil {
		t.Errorf("%v", err)
	}

	if dr.IsAfter(to) {
		t.Errorf("Date IS in DateRange - range: %v, date: %v", dr, to)
	}

	if !dr.IsAfter(date(2020, 1, 1)) {
		t.Errorf("02-01-2020->03-01-2020 is after 01-01-2020 - but was not detected")
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

func TestJSONMarshalUnmarshal(t *testing.T) {
	dr := NDaysBack(1, time.Now())
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
