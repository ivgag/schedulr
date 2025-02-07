package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

var TELEGRAM_BOT_TOKEN_ENV = "TELEGRAM_BOT_TOKEN"
var OPEN_AI_CLEINT_TOKEN_ENV = "AI_SERVICE_CLIENT"

var aiServiceClient = createOpenAiClient()
var telegramChatBot = telegramBot()

func main() {

	// Set up to receive updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramChatBot.GetUpdatesChan(u)

	// Listen for updates
	for update := range updates {
		if update.Message != nil { // If we got a message
			handleMessage(update.Message)
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := message.Text

	resp, err := aiServiceClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)

	// Simple reply
	msg := tgbotapi.NewMessage(chatID, "You said: "+text)
	_, err = telegramChatBot.Send(msg)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

func telegramBot() *tgbotapi.BotAPI {
	// Replace with your actual Telegram bot token
	botToken := os.Getenv(TELEGRAM_BOT_TOKEN_ENV)
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN_ENV environment variable not set")
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

func createOpenAiClient() *openai.Client {
	openAiToken := os.Getenv(OPEN_AI_CLEINT_TOKEN_ENV)
	if openAiToken == "" {
		log.Fatal("OPEN_AI_CLEINT_TOKEN_ENV environment variable not set")
	}

	return openai.NewClient(openAiToken)
}
