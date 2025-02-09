package tgbot

import (
	"context"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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

	b.chatBot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, b.helpHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeMessageText, "/register", bot.MatchTypeExact, b.registerHandler)

	b.chatBot.Start(b.ctx)

	return nil
}

// Stop terminates the Telegram bot's update processing.
func (b *Bot) Stop() {
	b.chatBot.Close(b.ctx)
}

func (b *Bot) helpHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	botAPI.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Help",
	})
}

func (b *Bot) registerHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	err := b.userService.CreateUser(&storage.User{
		TelegramId: update.Message.Chat.ID,
	})

	if err != nil {
		botAPI.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
		})
	}

	botAPI.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Registered",
	})
}

func (b *Bot) defaultHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	botAPI.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Say /hello",
	})
}
