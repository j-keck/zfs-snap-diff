package scanner

import (
	"time"
	"fmt"
	"encoding/json"
)


func (self DateRange) MarshalJSON() ([]byte, error) {
	type J struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	return json.Marshal(
		J{self.From.Format("2006-01-02"), self.To.Format("2006-01-02")},
	)
}



func (self *DateRange) UnmarshalJSON(b []byte) error {
	// unmarshal in a map
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}


	// extract values
	daysI, daysOk := m["days"]
	fromI, fromOk := m["from"]
	toI, toOk := m["to"]

	if fromOk && toOk {
		if daysOk {
			log.Warnf("Unmarshal DateRange: fields 'days' AND 'from' AND 'to' found - ignore 'days'")
		}

		// from
		if from, err := readDate(fromI); err == nil {
			self.From = from
		} else {
			return err
		}

		// to
		if to, err := readDate(toI); err == nil {
			self.To = to
		} else {
			return err
		}
	} else if fromOk {
		from, err := readDate(fromI)
		if err != nil {
			return err
		}
		self.From = from

		days, err := readDays(daysI)
		if err != nil {
			return err
		}
		dur := time.Duration(int(days)*24) * time.Hour
		self.To = from.Add(dur)

	} else if toOk {
		to, err := readDate(toI)
		if err != nil {
			return err
		}
		self.To = to

		days, err := readDays(daysI)
		if err != nil {
			return err
		}
		dur := time.Duration(-int(days)*24) * time.Hour
		self.From = to.Add(dur)
	} else {
		msg := "invalid json for DateRange - expected fields: [from, to|from, days|to, days]"
		return fmt.Errorf(msg)
	}

	return nil
}

func readDate(dateI interface{}) (time.Time, error) {
	dateS, ok := dateI.(string)
	if ! ok {
		return time.Time{}, fmt.Errorf("date in json was not a string: %v", dateI)
	}

	date, err := time.Parse("2006-1-2", dateS)
	if err != nil {
		return time.Time{}, fmt.Errorf("unparsable date: %s - %v", dateS, err)
	}

	return truncateTime(date), nil
}

func readDays(daysI interface{}) (int, error) {
	daysF, ok := daysI.(float64)
	if ! ok {
		return 0, fmt.Errorf("days in json was not numeric: %v", daysI)
	}

	return int(daysF), nil
}
