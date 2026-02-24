package main

import (
	"encoding/json"
	"time"
)

const (
	MonthYearLayout = "01-2006"
)

type MonthYear struct {
	time.Time
}

func (my *MonthYear) Parse(s string) error {
	var err error
	my.Time, err = time.Parse(MonthYearLayout, s)
	return err
}

func TruncateTimeToMonth(t time.Time) MonthYear {
	return MonthYear{time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())}
}

func (my *MonthYear) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if err := my.Parse(s); err != nil {
		return err
	}
	return nil
}

func (my MonthYear) MarshalJSON() ([]byte, error) {
	s := my.Time.Format(MonthYearLayout)
	return json.Marshal(s)
}
