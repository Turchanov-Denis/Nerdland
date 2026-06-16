package account

import (
	"database/sql"
	"errors"
	"time"
)

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
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	email TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    CONSTRAINT accounts_email_unique UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS profiles (
	account_id BIGINT PRIMARY KEY,
	username TEXT NOT NULL,
	display_name TEXT NOT NULL,
	bio TEXT NOT NULL DEFAULT '',
	avatar_url TEXT NOT NULL DEFAULT '',
	FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    
    CONSTRAINT profiles_username_unique UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS sessions (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	account_id BIGINT NOT NULL,
	refresh_token_hash TEXT NOT NULL,
	user_agent TEXT DEFAULT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	expires_at TIMESTAMPTZ NOT NULL,
	revoked_at TIMESTAMPTZ DEFAULT NULL,
	
	FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
	CONSTRAINT session_refresh_token_hash_unique UNIQUE (refresh_token_hash)
);
`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) CreateSession(s Session) error {
	const queryInsertSession = `
INSERT INTO sessions (account_id, refresh_token_hash, user_agent, expires_at)
VALUES ($1,$2,$3,$4)
`
	_, err := r.db.Exec(queryInsertSession, s.AccountID, s.RefreshTokenHash, s.UserAgent, s.ExpiresAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) RevokeSessionByRefreshTokenHash(RefreshTokenHash string) error {
	const queryRevokeSession = `
	UPDATE sessions SET revoked_at = $1 WHERE refresh_token_hash = $2 AND revoked_at IS NULL;
`
	res, err := r.db.Exec(queryRevokeSession, time.Now().UTC(), RefreshTokenHash)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *Repository) FindSessionByRefreshTokenHash(RefreshTokenHash string) (Session, error) {
	const queryFindSession = `
		SELECT id, account_id, refresh_token_hash, user_agent, created_at, expires_at, revoked_at
		FROM sessions
		WHERE refresh_token_hash = $1
		AND revoked_at IS NULL
		AND expires_at > now()
	`
	var s Session
	err := r.db.QueryRow(queryFindSession, RefreshTokenHash).Scan(
		&s.SessionID,
		&s.AccountID,
		&s.RefreshTokenHash,
		&s.UserAgent,
		&s.CreatedAt,
		&s.ExpiresAt,
		&s.RevokedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, err
	}

	return s, nil
}

func (r *Repository) createAccount(
	tx *sql.Tx,
	email string,
	passwordHash string,
) (int64, error) {
	var accountID int64

	const queryInsertAccount = `
	INSERT INTO accounts (email, password_hash)
	VALUES ($1, $2)
	RETURNING id
`
	err := tx.QueryRow(queryInsertAccount,
		email,
		passwordHash,
	).Scan(&accountID)

	if err != nil {
		return -1, err
	}
	return accountID, nil
}

func (r *Repository) createProfile(
	tx *sql.Tx,
	accountID int64,
	username string,
	displayName string,
) error {

	const queryInsertProfile = `
    INSERT INTO profiles (account_id, username, display_name)
    VALUES ($1, $2, $3);
`
	_, err := tx.Exec(queryInsertProfile,
		accountID,
		username,
		displayName)

	if err != nil {
		return err
	}

	return nil
}

// public profile
func (r *Repository) GetProfileByUsername(username string) (*PublicProfile, error) {
	var profile PublicProfile
	const selectProfileByUsernameQuery = `
	SELECT
		p.username,
		p.display_name,
		p.bio,
		p.avatar_url
	FROM profiles p
	WHERE p.username = $1
`
	err := r.db.QueryRow(selectProfileByUsernameQuery,
		username,
	).Scan(
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarUrl,
	)

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *Repository) getAccountByEmail(
	email string,
) (Account, error) {
	acc := Account{}
	query := "SELECT id, email, password_hash, created_at FROM accounts WHERE email = $1 Limit 1"
	err := r.db.QueryRow(query, email).Scan(&acc.ID, &acc.Email, &acc.PasswordHash, &acc.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Account{}, ErrAccountByEmailNotFound
		}
		return Account{}, err
	}
	return acc, nil
}
