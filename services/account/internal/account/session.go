package account

import (
	"errors"
	"time"
)

type Session struct {
	SessionID        int64
	AccountID        int64
	RefreshTokenHash string
	UserAgent        string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	RevokedAt        *time.Time
}

var (
	ErrSessionNotFound = errors.New("session not found")
)
