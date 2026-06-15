package account

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

type RegisterRequest struct {
	Email       string
	Password    string
	Username    string
	DisplayName string
}

type RegisterResponse struct {
	AccountID   int64
	Username    string
	DisplayName string

	AccessToken string
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AccountID int64
	Username  string
}

func (r *Repository) Register(
	req RegisterRequest,
) (RegisterResponse, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return RegisterResponse{}, err
	}
	defer tx.Rollback()

	// https://pkg.go.dev/golang.org/x/crypto/bcrypt
	bytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, err
	}
	passwordHash := string(bytes)

	// create account
	accountID, err := r.createAccount(tx, req.Email, passwordHash)
	if err != nil {
		return RegisterResponse{}, err
	}
	//create profile
	err = r.createProfile(tx, accountID, req.Username, req.DisplayName)
	if err != nil {
		return RegisterResponse{}, err
	}

	ans := RegisterResponse{
		AccountID:   accountID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
	}
	if err := tx.Commit(); err != nil {
		return RegisterResponse{}, err
	}
	return ans, nil
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
