package scanner

import (
	"fmt"
	"time"
)

type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func NewDateRange(to time.Time, rangeInDays int) DateRange {
	self := DateRange{To: truncateTime(to.UTC())}
	dur := time.Duration(-rangeInDays*24) * time.Hour
	self.From = truncateTime(to.Add(dur).UTC())
	return self
}

func (self *DateRange) IsAfter(other time.Time) bool {
	return self.From.After(truncateTime(other))
}

func (self *DateRange) IsBefore(other time.Time) bool {
	return self.To.Before(truncateTime(other))
}

func (self *DateRange) String() string {
	return fmt.Sprintf("between %s and %s",
		self.From.Format("Mon Jan 2 2006"), self.To.Format("Mon Jan 2 2006"))
}

func truncateTime(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}
