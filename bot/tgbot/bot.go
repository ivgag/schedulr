package tgbot

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
)

// NewBot initializes the Telegram bot using the provided context.
func NewBot(
	ctx context.Context,
	userService *service.UserService,
	eventService *service.EventService,
) (*Bot, error) {
	return &Bot{
		ctx:          ctx,
		userService:  userService,
		eventService: eventService,
	}, nil
}

// Bot wraps the bot.Bot along with service dependencies.
type Bot struct {
	ctx          context.Context
	chatBot      *bot.Bot
	userService  *service.UserService
	eventService *service.EventService
}

// Start begins processing updates. This method blocks until the context is cancelled.
func (b *Bot) Start() error {
	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDefaultHandler(b.defaultHandler),
	}

	chatBot, err := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)
	if err != nil {
		return err
	}
	b.chatBot = chatBot

	b.chatBot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, b.startHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeMessageText, "/linkgoogle", bot.MatchTypeExact, b.linkGoogleAccountHandler)

	b.chatBot.Start(b.ctx)

	return nil
}

// Stop terminates the Telegram bot's update processing.
func (b *Bot) Stop() {
	b.chatBot.Close(b.ctx)
}

func (b *Bot) startHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	err := b.userService.CreateUser(&storage.User{
		TelegramID: update.Message.Chat.ID,
	})

	if err != nil {
		botAPI.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
	} else {
		botAPI.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Link you Calendar: /linkgoogle",
		})
	}
}

func (b *Bot) linkGoogleAccountHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	user, err := b.userService.GetUserByTelegramID(update.Message.Chat.ID)
	if err != nil {
		botAPI.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
		return
	}

	botAPI.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Link your Google Calendar: " + b.userService.GetOAuth2Url(user.ID),
	})
}

func (b *Bot) defaultHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	events, err := b.eventService.CreateEventsFromUserMessage(ai.UserMessage{
		Text:    update.Message.Text,
		Caption: update.Message.Caption,
	})

	if err != nil {
		botAPI.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
	} else {
		botAPI.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Events created: " + fmt.Sprintf("%v", events),
		})
	}
}
