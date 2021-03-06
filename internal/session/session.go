package session

import (
	"time"

	"github.com/pkg/errors"
)

type Session struct {
	SessionID  string
	UserID     int64
	CreatedAt  time.Time
	ValidUntil time.Time
}

type Storage interface {
	Create(session *Session) error
	FindByID(id int64) (*Session, error)
	FindByToken(token string) (*Session, error)
}

func New(token string, userID int64) (*Session, error) {
	str := time.Now().Format(time.RFC3339)

	now, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse current time string")
	}

	const Deadline = 30
	until := now.Add(time.Minute * Deadline)

	return &Session{token, userID, now, until}, nil
}
