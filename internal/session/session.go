package session

import (
	"fmt"
	"time"
)

type Session struct {
	SessionID  string
	UserID     int
	CreatedAt  time.Time
	ValidUntil time.Time
}

const (
	Deadline = 30
)

var sessions []Session

func AddSession(token string, id int) error {
	str := time.Now().Format(time.RFC3339)

	now, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return fmt.Errorf("can't parse current time string: %v", err)
	}

	until := now.Add(time.Minute * Deadline)
	s := Session{token, id, now, until}

	sessions = append(sessions, s)

	return nil
}
