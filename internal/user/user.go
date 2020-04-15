package user

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const TimeLayout = "2006-01-02T15:04:05Z"

type BirthdayDate struct {
	Date time.Time
}

type JSONTime struct {
	time.Time
}

type User struct {
	ID        int          `json:"id,omitempty"`
	FirstName string       `json:"first_name,omitempty"`
	LastName  string       `json:"last_name,omitempty"`
	Birthday  BirthdayDate `json:"birthday,omitempty"`
	Email     string       `json:"email"`
	Password  string       `json:"password,omitempty"`
	UpdatedAt JSONTime     `json:"updated_at,omitempty"`
	CreatedAt JSONTime     `json:"created_at,omitempty"`
}

func (b *BirthdayDate) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		return nil
	}

	const layout = `"` + "2006-01-02" + `"`
	t, err := time.Parse(layout, s)
	if err != nil {
		return fmt.Errorf("can't parse birth date string: %v", err)
	}
	b.Date = t

	return nil
}

func (t *JSONTime) MarshalJSON() ([]byte, error) {
	s := t.Format(TimeLayout)

	return []byte(s), nil
}

func (jt JSONTime) Value() (driver.Value, error) {
	return jt.Time, nil
}

func (jt *JSONTime) Scan(value interface{}) error {
	jt.Time = value.(time.Time)
	str := jt.Time.Format(TimeLayout)
	t, err := time.Parse(TimeLayout, str)
	if err != nil {
		return fmt.Errorf("can't parse current time string: %v", err)
	}
	jt.Time = t
	return nil
}

