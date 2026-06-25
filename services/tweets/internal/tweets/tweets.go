package tweets

import (
	"context"
	"time"
)

type TweetService struct {
	r Repository
}
type Tweet struct {
	ID        int64
	AccountID int64
	Text      string
	CreatedAt time.Time
}
type CreateTweetRequest struct {
	AccountID int64
	Text      string
}

type TweetResponse struct {
	ID        int64
	AccountID int64
	Text      string
}

func NewTweetService(r Repository) *TweetService {
	return &TweetService{r: r}
}

func (t *TweetService) CreateTweet(ctx context.Context, req CreateTweetRequest) (TweetResponse, error) {
	// сохранить в БД
	return TweetResponse{}, nil
}
