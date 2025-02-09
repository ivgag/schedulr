package main

import (
	"database/sql"
	"log"

	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/bot"
	"github.com/ivgag/schedulr/rest"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=schedulr sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	userRepo := storage.NewPostgresUserRepository(db)

	ai, err := ai.NewOpenAI()
	if err != nil {
		panic(err)
	}

	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(ai)

	bot, err := bot.RunTelegramBot(userService, eventService)
	if err != nil {
		panic(err)
	}
	defer bot.Stop()
	bot.Start()

	gin := rest.NewRouter(userService)
	gin.Run(":8080")

	log.Default().Println("Server is running on port 8080")

}
