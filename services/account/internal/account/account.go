package account

import (
	"database/sql"
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
	AccountID   int64
	Email       string
	Username    string
	DisplayName string
	Bio         string
	AvatarUrl   string
	CreatedAt   time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	if db == nil {
		panic("database connection cannot be nil")
	}

	return &Repository{
		db: db,
	}
}

func (r *Repository) Init() error {
	query := `
CREATE TABLE IF NOT EXISTS accounts (
	id BIGSERIAL PRIMARY KEY,
	email TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    
    CONSTRAINT accounts_email_unique UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS profiles (
	account_id BIGINT PRIMARY KEY,
	username TEXT NOT NULL,
	display_name TEXT NOT NULL,
	bio TEXT NOT NULL DEFAULT '',
	avatar_url TEXT NOT NULL DEFAULT '',
	FOREIGN KEY (account_id)
	REFERENCES accounts(id)
	ON DELETE CASCADE
    
    CONSTRAINT profiles_username_unique UNIQUE (username)
);
`

	_, err := r.db.Exec(query)
	return err
}
