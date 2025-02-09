package bot

import (
	"fmt"
	"log"
	"os"

	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var telegramBotTokenEnv = "TELEGRAM_BOT_TOKEN"

func RunTelegramBot(
	userService *service.UserService,
	eventService *service.EventService,
) (Bot, error) {
	// Replace with your actual Telegram bot token
	botToken := os.Getenv(telegramBotTokenEnv)
	if botToken == "" {
		log.Fatal(telegramBotTokenEnv + " environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// Debug mode can be turned on for more logs
	bot.Debug = false

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	return Bot{
		telegramChatBot: bot,
		userService:     userService,
		eventService:    eventService,
	}, nil
}

type Bot struct {
	telegramChatBot *tgbotapi.BotAPI
	userService     *service.UserService
	eventService    *service.EventService
}

func (b *Bot) Start() {
	go b.startReceivingUpdates()
}

func (b *Bot) Stop() {
	b.telegramChatBot.StopReceivingUpdates()
}

func (b *Bot) startReceivingUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.telegramChatBot.GetUpdatesChan(u)

	for update := range updates {
		msg, err := b.handleUpdate(update)
		if err != nil {
			log.Panic(err)
		}

		reply := tgbotapi.NewMessage(update.Message.Chat.ID, msg)

		if _, err := b.telegramChatBot.Send(reply); err != nil {
			log.Panic(err)
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) (string, error) {
	if update.Message == nil { // ignore any non-Message updates
		return "", nil
	}

	if update.Message.IsCommand() { // ignore any non-command Messages
		return b.handleCommand(update.Message)
	} else {
		return b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) (string, error) {
	events, err := b.eventService.CreateEventsFromUserMessage(
		ai.UserMessage{
			Text:    message.Text,
			Caption: message.Caption,
		},
	)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Created events: ", events), nil
}

func (b *Bot) handleCommand(message *tgbotapi.Message) (string, error) {
	switch command := message.Command(); command {
	case "help":
		return "I understand ", nil
	case "init":
		b.userService.CreateUser(&storage.User{
			TelegramId: message.From.ID,
		})

		return "registered", nil
	default:
		return "I don't know that command", nil
	}
}
