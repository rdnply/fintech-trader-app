package session

import "time"

type Session struct {
	SessionID string
	UserID string
	CreatedAt time.Time
	ValidUntil time.Time
}
