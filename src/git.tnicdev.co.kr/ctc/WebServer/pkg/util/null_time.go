package util

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/blackss2/utility/convert"
)

type NullTime struct {
	Time  time.Time
	Valid bool
}

func NullTimeNow() NullTime {
	return NewNullTime(time.Now())
}

func NullTimeInvalid() NullTime {
	return NullTime{
		Valid: false,
	}
}

func NewNullTime(t time.Time) NullTime {
	return NullTime{
		Time:  t,
		Valid: true,
	}
}

func (nt *NullTime) String() string {
	if nt.Valid {
		return nt.Time.String()
	} else {
		return "nil"
	}
}

func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return []byte(fmt.Sprintf(`"%s"`, convert.String(nt.Time))), nil
	} else {
		return []byte("null"), nil
	}
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	nt.Valid = false
	if len(data) > 0 && string(data) != "null" {
		tm := convert.Time(data[1 : len(data)-1])
		if tm != nil {
			nt.Time = (*tm)
		}
		if nt.Time.Unix() != (time.Time{}).Unix() {
			nt.Valid = true
		}
	}
	return nil
}

func (nt *NullTime) MarshalRQL() (interface{}, error) {
	t := nt.Time

	timeVal := float64(t.UnixNano()) / float64(time.Second)

	// use seconds-since-epoch precision if time.Time `t`
	// is before the oldest nanosecond time
	if t.Before(time.Unix(0, math.MinInt64)) {
		timeVal = float64(t.Unix())
	}

	return map[string]interface{}{
		"$reql_type$": "TIME",
		"epoch_time":  timeVal,
		"timezone":    t.Format("-07:00"),
	}, nil
}

func (nt *NullTime) UnmarshalRQL(value interface{}) error {
	nt.Valid = false
	if value != nil {
		if t, is := value.(time.Time); is {
			if !t.IsZero() {
				nt.Time = t
				nt.Valid = true
			}
		} else {
			return errors.New("type error")
		}
	}
	return nil
}
