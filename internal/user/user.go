package user

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

type BirthDay struct {
	time.Time
}

type JSONTime struct {
	time.Time
}

type User struct {
	ID        int      `json:"id,omitempty"`
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	Birthday  BirthDay `json:"birthday,omitempty"`
	Email     string   `json:"email"`
	Password  string   `json:"password,omitempty"`
	UpdatedAt JSONTime `json:"updated_at,omitempty"`
	CreatedAt JSONTime `json:"created_at,omitempty"`
}

func (b *BirthDay) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		return nil
	}

	const layout = `"` + DateLayout + `"`
	t, err := time.Parse(layout, s)
	if err != nil {
		return fmt.Errorf("can't parse birth date string: %v", err)
	}
	b.Time = t

	return nil
}

func (bd *BirthDay) MarshalJSON() ([]byte, error) {
	s := bd.Format(DateLayout)

	return []byte(s), nil
}

func (bd BirthDay) Value() (driver.Value, error) {
	return bd.Time, nil
}

func (bd *BirthDay) Scan(value interface{}) error {
	t := value.(time.Time)
	str := t.Format(DateLayout)
	t, err := time.Parse(DateLayout, str)
	if err != nil {
		return fmt.Errorf("can't parse birth date in scanner for database: %v", err)
	}

	bd.Time = t
	return nil
}

func (t *JSONTime) MarshalJSON() ([]byte, error) {
	s := t.Format(time.RFC3339)

	return []byte(s), nil
}

func (jt JSONTime) Value() (driver.Value, error) {
	return jt.Time, nil
}

func (jt *JSONTime) Scan(value interface{}) error {
	jt.Time = value.(time.Time)

	return nil
}

