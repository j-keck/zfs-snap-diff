package scanner

import (
	"fmt"
	"time"
)

// DateRange with from and to dates.
//   - time is ignored
//   - dates are inclusive
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func NewDateRange(from time.Time, to time.Time) (DateRange, error) {
	if from.After(to) {
		return DateRange{from, to}, fmt.Errorf("invalid DateRange - from: %v is AFTER to: %v", from , to)
	}
	return DateRange { from, to }, nil
}

func NDaysBack(n int, to time.Time) DateRange {
	self := DateRange{To: truncateTime(to.UTC())}
	dur := time.Duration(-n*24) * time.Hour
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
