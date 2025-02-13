package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/rest"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
	"github.com/ivgag/schedulr/tgbot"
	"github.com/ivgag/schedulr/utils"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	// Create a global context with SIGINT signal handling.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	connStr := utils.GetenvOrPanic("DATABASE_URL")
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	userRepo := storage.NewUserRepository(db)
	linkedAccountRepo := storage.NewLinkedAccountRepository(db)

	aiClient, err := ai.NewOpenAI()
	if err != nil {
		panic(err)
	}

	googleTokenService, err := service.NewGoogleTokenService(linkedAccountRepo)
	if err != nil {
		panic(err)
	}

	tokenServices := map[model.Provider]service.TokenService{
		model.ProviderGoogle: googleTokenService,
	}

	userService := service.NewUserService(
		userRepo,
		tokenServices,
	)

	googleCalendarService := service.NewGoogleCalendarService(googleTokenService)
	calendarServices := map[model.Provider]service.CalendarService{
		model.ProviderGoogle: googleCalendarService,
	}

	eventService := service.NewEventService(aiClient, *userService, calendarServices)

	// Initialize the Telegram bot with the global context.
	bot, err := tgbot.NewBot(ctx, userService, eventService)
	if err != nil {
		panic(err)
	}
	// Run the bot in a separate goroutine.
	go func() {
		if err := bot.Start(); err != nil {
			log.Printf("Telegram bot error: %v", err)
		}
	}()

	// Initialize the REST server.
	router := rest.NewRouter(userService)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	// Run the REST server in a separate goroutine.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("REST server error: %v", err)
		}
	}()
	log.Println("Server is running on port 8080")

	// Wait for the termination signal.
	<-ctx.Done()

	// Graceful shutdown of the REST server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("REST server shutdown error: %v", err)
	}

	bot.Stop()
}
