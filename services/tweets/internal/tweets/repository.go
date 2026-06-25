package tweets

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateTweet(ctx context.Context, tweet Tweet) (Tweet, error)
	//GetTweetByID(ctx context.Context, id int64) (Tweet, error)
	//DeleteTweet(ctx context.Context, id int64) error
	//ListTweetsByAccountID(ctx context.Context, accountID int64) ([]Tweet, error)
}

type RepositoryPostgres struct {
	db *sql.DB
}

func (r *RepositoryPostgres) CreateTweet(ctx context.Context, tweet Tweet) (Tweet, error) {
	return Tweet{}, nil
}

func NewRepositoryPostgres(db *sql.DB) Repository {
	if db == nil {
		panic("database connection cannot be nil")
	}
	return &RepositoryPostgres{db: db}
}
