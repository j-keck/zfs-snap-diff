package scanner

import (
	"fmt"
	"time"
)

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func TimeRangeFromLastNDays(nDays int) TimeRange {
	self := TimeRange{To: time.Now()}
	self.AdjustFromToNDaysBeforeTo(nDays)
	return self
}

func ParseTimeRange(layout string, fromStr string, toStr string) (TimeRange, error) {
	from, err := time.Parse(layout, fromStr)
	if err != nil {
		return TimeRange{}, err
	}

	to, err := time.Parse(layout, toStr)
	if err != nil {
		return TimeRange{}, err
	}

	if from.After(to) {
		return TimeRange{},
			fmt.Errorf("invalid TimeRange - from (%s) is after to (%s)", from, to)
	}

	return TimeRange{from, to}, nil
}

func (self *TimeRange) Contains(other time.Time) bool {
	return other.After(self.From) && other.Before(self.To)
}

func (self *TimeRange) AdjustFromToNDaysBeforeTo(nDays int) {
	self.From = time.Unix(self.To.Unix()-int64(nDays*24*60*60), 0)
}

func (self *TimeRange) FromIsAfterTo() bool {
	return self.From.After(self.To)
}

func (self *TimeRange) String() string {
	return fmt.Sprintf("between %s and %s", self.From, self.To)
}
