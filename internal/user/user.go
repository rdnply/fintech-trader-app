package user

import (
	"cw1/internal/format"
)

type User struct {
	ID        int64       `json:"id,omitempty"`
	FirstName string      `json:"first_name,omitempty"`
	LastName  string      `json:"last_name,omitempty"`
	Birthday  format.Day  `json:"birthday,omitempty"`
	Email     string      `json:"email"`
	Password  string      `json:"password,omitempty"`
	UpdatedAt format.Time `json:"updated_at,omitempty"`
	CreatedAt format.Time `json:"created_at,omitempty"`
}

type Storage interface {
	Create(u *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id int64) (*User, error)
	Update(u *User) error
}

type Info struct {
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Birthday  format.Day `json:"birthday,omitempty"`
	Email     string     `json:"email"`
}

func NewInfo(u *User) Info {
	return Info{u.FirstName, u.LastName, u.Birthday, u.Email}
}
