package main

import (
	"database/sql"
	"log"

	"github.com/ivgag/schedulr/bot"
	"github.com/ivgag/schedulr/rest"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
)

func main() {
	// Set up the PostgreSQL connection.
	connStr := "user=youruser password=yourpassword dbname=yourdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	userRepo := storage.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepo)

	bot.RunTelegramBot(userService)
	rest.NewRouter()
}
