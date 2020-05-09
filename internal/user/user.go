package user

import (
	"cw1/internal/format"
	"encoding/json"
	"fmt"
)

type User struct {
	ID        int64           `json:"id,omitempty"`
	FirstName string          `json:"first_name,omitempty"`
	LastName  string          `json:"last_name,omitempty"`
	Birthday  *format.Day     `json:"birthday,omitempty"`
	Email     string          `json:"email"`
	Password  string          `json:"password,omitempty"`
	UpdatedAt format.NullTime `json:"updated_at,omitempty"`
	CreatedAt format.NullTime `json:"created_at,omitempty"`
}

type Storage interface {
	Create(u *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id int64) (*User, error)
	Update(u *User) error
}

func (u *User) MarshalJSON() ([]byte, error) {
	var birthday *string
	if u.Birthday.V.Valid {
		t := u.Birthday.V.Time
		b := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
		birthday = &b
	}

	return json.Marshal(&struct {
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Birthday  *string `json:"birthday,omitempty"`
		Email     string  `json:"email"`
	}{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Birthday:  birthday,
		Email:     u.Email,
	})
}
