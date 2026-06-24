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
