package tweets

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateTweet(ctx context.Context, req CreateTweetRequest) (Tweet, error)
	//GetTweetByID(ctx context.Context, id int64) (Tweet, error)
	//DeleteTweet(ctx context.Context, id int64) error
	//ListTweetsByAccountID(ctx context.Context, accountID int64) ([]Tweet, error)
}

type RepositoryPostgres struct {
	db *sql.DB
}

func (r *RepositoryPostgres) CreateTweet(ctx context.Context, req CreateTweetRequest) (Tweet, error) {
	const queryInsetTweet = `
	INSERT INTO tweets(account_id, text)
	VALUES($1, $2)
	RETURNING id
`
	var tweet Tweet
	row := r.db.QueryRowContext(ctx, queryInsetTweet, req.AccountID, req.Text)

	err := row.Scan(&tweet.ID)
	if err != nil {
		return Tweet{}, err
	}
	tweet.AccountID = req.AccountID
	tweet.Text = req.Text
	return tweet, nil
}

func NewRepositoryPostgres(db *sql.DB) Repository {
	if db == nil {
		panic("database connection cannot be nil")
	}
	return &RepositoryPostgres{db: db}
}
