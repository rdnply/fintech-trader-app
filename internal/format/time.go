package format

import (
	"database/sql/driver"
	"time"
)

type Time struct {
	time.Time
}

func NewTime() Time {
	return Time{time.Now()}
}

func (t *Time) MarshalJSON() ([]byte, error) {
	s := t.Format(time.RFC3339)

	return []byte(s), nil
}

func (t Time) Value() (driver.Value, error) {
	return t.Time, nil
}

func (t *Time) Scan(value interface{}) error {
	t.Time = value.(time.Time)

	return nil
}
