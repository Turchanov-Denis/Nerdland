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

type UserProfile struct {
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
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS profiles (
	account_id BIGINT PRIMARY KEY,
	username TEXT UNIQUE NOT NULL,
	display_name TEXT NOT NULL,
	bio TEXT NOT NULL DEFAULT '',
	avatar_url TEXT NOT NULL DEFAULT '',
	FOREIGN KEY (account_id)
	REFERENCES accounts(id)
	ON DELETE CASCADE
);
`

	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) RegisterUser(
	email string,
	passwordHash string,
	username string,
	displayName string,
) (err error) {

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var accountID int64

	err = tx.QueryRow(`
		INSERT INTO accounts (email, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`,
		email,
		passwordHash,
	).Scan(&accountID)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO profiles (
			account_id,
			username,
			display_name
		)
		VALUES ($1, $2, $3)
	`,
		accountID,
		username,
		displayName,
	)

	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (r *Repository) GetProfileByUsername(username string) (*UserProfile, error) {
	var profile UserProfile

	err := r.db.QueryRow(`
		SELECT
			p.account_id,
			a.email,
			p.username,
			p.display_name,
			p.bio,
			p.avatar_url,
			a.created_at
		FROM profiles p
		INNER JOIN accounts a
			ON a.id = p.account_id
		WHERE p.username = $1
	`,
		username,
	).Scan(
		&profile.AccountID,
		&profile.Email,
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarUrl,
		&profile.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &profile, nil
}
