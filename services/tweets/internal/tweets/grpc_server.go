package tweets

import (
	"context"

	pb "tweets/internal/api/tweets"
)

type GRPCServer struct {
	pb.UnimplementedTweetServiceServer
	service *TweetService
}

func NewGRPCServer(service *TweetService) *GRPCServer {
	return &GRPCServer{
		service: service,
	}
}

func (s *GRPCServer) CreateTweet(
	ctx context.Context,
	req *pb.CreateTweetRequest,
) (*pb.TweetResponse, error) {

	resp, err := s.service.CreateTweet(ctx, CreateTweetRequest{
		AccountID: req.AccountId,
		Text:      req.Text,
	})
	if err != nil {
		return nil, err
	}

	return &pb.TweetResponse{
		Id:        resp.ID,
		AccountId: resp.AccountID,
		Text:      resp.Text,
	}, nil
}
