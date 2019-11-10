package scanner

import (
	"time"
	"errors"
)

type TimeRange struct {
	From time.Time `json:"from"`
	Till time.Time `json:"till"`
}

func TimeRangeFromLastNDays(nDays int) TimeRange {
	till := time.Now()
	from := time.Unix(till.Unix() - int64(nDays * 24 * 60 * 60), 0)
	return TimeRange { from, till }
}

func ParseTimeRange(layout string, fromStr string, tillStr string) (TimeRange, error) {
	from, err := time.Parse(layout, fromStr)
	if err != nil {
		return TimeRange{}, err
	}

	till, err := time.Parse(layout, tillStr)
	if err != nil {
		return TimeRange{}, err
	}

	if from.After(till) {
		return TimeRange{}, errors.New("invalid TimeRange - from is after till")
	}

	return TimeRange { from, till }, nil
}


func (self *TimeRange) Contains(other time.Time) bool {
	return other.After(self.From) && other.Before(self.Till)
}
