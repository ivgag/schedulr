package bot

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var telegramBotTokenEnv = "TELEGRAM_BOT_TOKEN"

func RunTelegramBot() (Bot, error) {

	// Set up to receive updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	var telegramChatBot = telegramBot()
	updates := telegramChatBot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.

		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand "
		case "start":
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := telegramChatBot.Send(msg); err != nil {
			log.Panic(err)
		}
	}

	// Listen for updates
	for update := range updates {
		if update.Message != nil { // If we got a message
			handleMessage(telegramChatBot, update.Message)
		}
	}

	return Bot{TelegramChatBot: telegramChatBot}, nil
}

func handleMessage(telegramChatBot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Simple reply
	msg := tgbotapi.NewMessage(chatID, "I'm a bot, please talk to me!")
	_, err := telegramChatBot.Send(msg)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

func telegramBot() *tgbotapi.BotAPI {
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

	return bot
}

type Bot struct {
	TelegramChatBot *tgbotapi.BotAPI
}

func (b *Bot) Stop() {
	b.TelegramChatBot.StopReceivingUpdates()
}
