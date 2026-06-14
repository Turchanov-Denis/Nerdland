package main

import (
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
		// Не блокируем выполнение, если env прокинут через Docker напрямую
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

	// db init
	err = database.Init(db)
	if err != nil {
		log.Error(err.Error())
	}
	// test register and test read

}
