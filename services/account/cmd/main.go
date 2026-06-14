package main

import (
	"account/internal/account"
	"account/internal/database"
	"account/internal/logger"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	log := logger.New()
	log.Info("The server is running, but it still attracts nerds.")

	err := godotenv.Load(".env", "./.env")
	if err != nil {
		log.Info("No .env found")
	}

	db, err := database.NewPostgress()
	if err != nil {
		log.Error(err.Error())
	}
	defer db.Close()

	// Make sure it works.
	err = db.Ping()
	if err != nil {
		log.Error(err.Error())
	}

	accountRepo := account.NewRepository(db)
	// db init
	accountRepo.Init()
	// test register and test read
	err = accountRepo.RegisterUser("alex@gmail.com",
		"hash123",
		"alex",
		"Alex")
	if err != nil {
		log.Error(err.Error())
	}
	profile, err := accountRepo.GetProfileByUsername("alex")
	if err != nil {
		log.Error(err.Error())
	}
	log.Info("Get", "profile",
		profile)
}
