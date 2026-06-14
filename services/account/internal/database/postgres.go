package database

import (
	"database/sql"
	"os"
	"strconv"

	"github.com/lib/pq"
)

func NewPostgress() (*sql.DB, error) {
	cfg, err := pq.NewConfig("")
	if err != nil {
		return nil, err
	}
	cfg.SSLMode = "disable"

	cfg.Host = os.Getenv("DB_HOST")
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		cfg.Port = uint16(port)
	}
	cfg.Database = os.Getenv("DB_NAME")

	cfg.User = os.Getenv("DB_USER")
	cfg.Password = os.Getenv("DB_PASSWORD")

	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Create connection pool.
	db := sql.OpenDB(c)

	return db, nil

}

func Init(db *sql.DB) error {
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

	FOREIGN KEY(account_id) 
	REFERENCES accounts(id) 
	ON DELETE CASCADE
);
`
	_, err := db.Exec(query)
	return err
}
