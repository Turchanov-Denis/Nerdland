package account

import (
	"database/sql"
	"errors"
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

func (r *Repository) Init() error {
	query := `
CREATE TABLE IF NOT EXISTS accounts (
	id BIGSERIAL PRIMARY KEY,
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
`

	_, err := r.db.Exec(query)
	return err
}

// public profile
func (r *Repository) GetProfileByUsername(username string) (*PublicProfile, error) {
	var profile PublicProfile

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
