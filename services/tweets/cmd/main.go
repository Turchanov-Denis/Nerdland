package main

import "tweets/logger"

func main() {
	log := logger.New()
	log.Info("tweets")
}
