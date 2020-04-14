package user

import (
	"fmt"
	"time"
)

type BirthdayDate struct {
	Date time.Time
}

type User struct {
	ID        int
	FirstName string       `json:"first_name,omitempty"`
	LastName  string       `json:"last_name,omitempty"`
	Birthday  BirthdayDate `json:"birthday,omitempty"`
	Email     string       `json:"email"`
	Password  string       `json:"password"`
	UpdatedAt time.Time
	CreatedAt time.Time
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
