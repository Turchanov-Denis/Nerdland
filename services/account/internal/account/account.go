package account

import (
	"time"
)

type Account struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Profile struct {
	AccountID   int64
	Username    string
	DisplayName string
	Bio         string
	AvatarUrl   string
}

type PublicProfile struct {
	Username    string
	DisplayName string
	Bio         string
	AvatarUrl   string
}
