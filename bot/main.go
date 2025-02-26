package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/rest"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
	"github.com/ivgag/schedulr/tgbot"
	"github.com/ivgag/schedulr/utils"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create global context with SIGINT handling.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfgName, err := utils.GetenvOrError("CONFIG_NAME")
	if err != nil {
		log.Panic().Err(err).Msg("Failed to get CONFIG_PATH")
	}

	path, err := utils.GetenvOrError("CONFIG_PATH")
	if err != nil {
		log.Panic().Err(err).Msg("Failed to get CONFIG_PATH")
	}

	cfg, err := LoadConfig(path, cfgName)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to load config")
	}

	db := initDatabase(cfg.Database.URL)
	defer db.Close()

	// Initialize repositories.
	userRepo := storage.NewUserRepository(db)
	linkedAccountRepo := storage.NewLinkedAccountRepository(db)

	// Initialize AI services.
	aiSvc := initAIService(&cfg.AIConfig)

	// Initialize Google token and user services.
	googleTokenSvc := service.NewGoogleTokenService(&cfg.Google, linkedAccountRepo)
	tokenServices := map[model.Provider]service.TokenService{
		model.ProviderGoogle: googleTokenSvc,
	}
	timeZoneService := service.NewTimezoneService(&cfg.Google)
	userSvc := service.NewUserService(userRepo, tokenServices, *timeZoneService, linkedAccountRepo)

	// Initialize calendar and event services.
	googleCalendarSvc := service.NewGoogleCalendarService(googleTokenSvc)
	calendarServices := map[model.Provider]service.CalendarService{
		model.ProviderGoogle: googleCalendarSvc,
	}
	eventSvc := service.NewEventService(*aiSvc, *userSvc, calendarServices)

	// Start Telegram bot.
	bot := tgbot.NewBot(ctx, &cfg.TelegramBot, userSvc, eventSvc)
	go startTelegramBot(bot)

	// Initialize REST router and server.
	router := rest.NewRouter(&cfg.TelegramBot, userSvc)
	srv := initHTTPServer(cfg.Rest, router)
	go startHTTPServer(srv, cfg.Rest)

	log.Info().
		Int("port", cfg.Rest.Port).
		Msg("REST server is running")

	<-ctx.Done()
	shutdownServer(srv)
	bot.Stop()
}

func initDatabase(dbURL string) *sql.DB {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to connect to the database")
	}
	if err = db.Ping(); err != nil {
		log.Panic().Err(err).Msg("Failed to ping the database")
	}
	return db
}

func initAIService(aiConfig *service.AIConfig) *service.AIService {
	openAi := ai.NewOpenAI(&aiConfig.OpenAI)
	deepseek := ai.NewDeepSeekAI(&aiConfig.Deepseek)
	return service.NewAIService([]ai.AI{openAi, deepseek}, aiConfig)
}

func createAutocertManager(restCfg rest.RestConfig) autocert.Manager {
	return autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(restCfg.Domain),
		Cache:      autocert.DirCache("/certs"),
	}
}

func initHTTPServer(restCfg rest.RestConfig, router http.Handler) *http.Server {
	addr := ":" + strconv.Itoa(restCfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	if restCfg.TLS {
		m := createAutocertManager(restCfg)
		srv.TLSConfig = m.TLSConfig()
	}
	return srv
}

func startHTTPServer(srv *http.Server, restCfg rest.RestConfig) {
	if restCfg.TLS {
		m := createAutocertManager(restCfg)
		// Start HTTP to HTTPS redirection.
		go func() {
			redirectSrv := &http.Server{
				Addr:    ":80",
				Handler: m.HTTPHandler(nil),
			}
			if err := redirectSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("HTTP redirection server error")
			}
		}()
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("REST server TLS error")
		}
	} else {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("REST server error")
		}
	}
}

func startTelegramBot(bot *tgbot.Bot) {
	if err := bot.Start(); err != nil {
		log.Panic().Err(err).Msg("Telegram bot error")
	}
}

func shutdownServer(srv *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("REST server shutdown error")
	}
}

type AppConfig struct {
	TelegramBot tgbot.TelegramBotConfig `mapstructure:"telegram_bot"`
	AIConfig    service.AIConfig        `mapstructure:"ai"`
	Google      service.GoogleConfig    `mapstructure:"google"`
	Database    storage.DatabaseConfig  `mapstructure:"database"`
	Rest        rest.RestConfig         `mapstructure:"rest"`
}

func LoadConfig(
	path string,
	configName string,
) (*AppConfig, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Expand environment variables in string values
	for _, key := range viper.AllKeys() {
		val := viper.GetString(key)
		viper.Set(key, os.ExpandEnv(val))
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
