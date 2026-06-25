package main

import (
	"net"
	pb "tweets/internal/api/tweets"
	"tweets/internal/database"
	"tweets/internal/tweets"
	"tweets/logger"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	log := logger.New()
	log.Info("tweets server start, nerd time gone")

	err := godotenv.Load(".env", "./.env")
	if err != nil {
		log.Error(".env not found, using environment variables")
	}

	db, err := database.NewPostgress()
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
	defer db.Close()

	// Make sure it works.
	err = db.Ping()
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}

	tweetRepository := tweets.NewRepositoryPostgres(db)
	tweetService := tweets.NewTweetService(tweetRepository)

	grpcServer := grpc.NewServer()

	pb.RegisterTweetServiceServer(
		grpcServer,
		tweets.NewGRPCServer(tweetService),
	)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}

	log.Info("gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Error(err.Error())
		panic(err)
	}
}
