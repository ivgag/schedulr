package tgbot

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
	"github.com/rs/zerolog/log"
)

// NewBot initializes the Telegram bot with the provided context and services.
func NewBot(ctx context.Context, userService *service.UserService, eventService *service.EventService) (*Bot, error) {
	return &Bot{
		ctx:          ctx,
		userService:  userService,
		eventService: eventService,
	}, nil
}

// Bot wraps bot.Bot with service dependencies.
type Bot struct {
	ctx          context.Context
	chatBot      *bot.Bot
	userService  *service.UserService
	eventService *service.EventService
}

// Start initializes the Telegram bot, registers handlers, and starts update processing.
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

// Stop terminates update processing.
func (b *Bot) Stop() {
	b.chatBot.Close(b.ctx)
}

// startHandler creates a new user and prompts to link calendar.
func (b *Bot) startHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	if err := b.userService.CreateUser(&storage.User{TelegramID: chatID}); err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}
	b.sendMessage(ctx, chatID, "Link your Calendar: /linkgoogle", "")
}

// linkGoogleAccountHandler returns the OAuth2 URL for linking a Google account.
func (b *Bot) linkGoogleAccountHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	link, err := b.userService.GetOAuth2Url(chatID, model.ProviderGoogle)
	if err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}
	b.sendMessage(ctx, chatID, "Link your Google Calendar: "+link, "")
}

// defaultHandler processes incoming messages (text or captions) and delegates handling.
func (b *Bot) defaultHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	if update.Message.Text != "" {
		b.handleIncomingText(ctx, update, update.Message.Text, update.Message.Entities, model.UserMessage, update.Message.From.Username)
	} else if update.Message.Caption != "" {
		from := getForwardOrigin(update)
		b.handleIncomingText(ctx, update, update.Message.Caption, update.Message.CaptionEntities, model.ForwardedMessage, from)
	}
	// Если сообщение не содержит текст или подпись, ничего не делаем.
}

func (b *Bot) handleIncomingText(
	ctx context.Context,
	update *models.Update,
	text string,
	entities []models.MessageEntity,
	msgType model.MessageType,
	from string,
) {
	matterEntities := make([]models.MessageEntity, 0)
	for _, entity := range entities {
		if entity.Type == models.MessageEntityTypeTextLink ||
			entity.Type == models.MessageEntityTypeTextMention ||
			entity.Type == models.MessageEntityTypeURL {
			matterEntities = append(matterEntities, entity)
		}
	}

	message := &model.TextMessage{
		From:        from,
		Text:        formatMessageText(text, matterEntities),
		MessageType: msgType,
	}
	b.handleTextMessage(ctx, update, message)
}

// handleTextMessage processes the text message and sends events as Telegram messages.
func (b *Bot) handleTextMessage(ctx context.Context, update *models.Update, textMessage *model.TextMessage) {
	chatID := update.Message.Chat.ID
	events, err := b.eventService.CreateEventsFromUserMessage(chatID, *textMessage)
	if err != nil {
		log.Error().
			Int64("chatID", chatID).
			Err(err).
			Msg("Failed to create events")

		b.sendMessage(ctx, chatID, "Failed to create events. Try later.", "")
		return
	}

	for _, event := range events {
		b.sendMessage(ctx, chatID, formatEventForTelegram(event), models.ParseModeMarkdownV1)
	}
}

// sendMessage wraps bot.SendMessage with common parameters.
func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string, parseMode models.ParseMode) {
	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	if parseMode != "" {
		params.ParseMode = parseMode
	}
	b.chatBot.SendMessage(ctx, params)
}

// formatEventForTelegram returns a formatted string representing the event.
func formatEventForTelegram(e model.Event) string {
	message := fmt.Sprintf("*%s*\n", e.Title)
	if e.Description != "" {
		message += fmt.Sprintf("%s\n", e.Description)
	}
	message += fmt.Sprintf("*When:* %s - %s (%s)\n",
		e.Start.DateTime.Format(time.DateTime),
		e.End.DateTime.Format(time.DateTime),
		e.Start.TimeZone,
	)
	if e.Location != "" {
		message += fmt.Sprintf("*Where:* %s\n", e.Location)
	}
	if e.Link != "" {
		message += fmt.Sprintf("[More details](%s)\n", e.Link)
	}
	return message
}

// getForwardOrigin extracts the origin information from a forwarded message.
func getForwardOrigin(update *models.Update) string {
	fwd := update.Message.ForwardOrigin
	switch {
	case fwd.MessageOriginUser != nil:
		return fwd.MessageOriginUser.SenderUser.Username
	case fwd.MessageOriginHiddenUser != nil:
		return fwd.MessageOriginHiddenUser.SenderUserName
	case fwd.MessageOriginChannel != nil:
		return fwd.MessageOriginChannel.Chat.Title
	default:
		return fwd.MessageOriginChat.SenderChat.Username
	}
}

func formatMessageText(text string, entities []models.MessageEntity) string {
	if len(entities) == 0 {
		return text
	}

	var sb strings.Builder

	sort.SliceStable(entities, func(i, j int) bool {
		return entities[i].Offset < entities[j].Offset
	})

	runeArray := []rune(text)
	generalOffset := 0

	for _, entiry := range entities {
		text1 := string(runeArray[generalOffset : entiry.Offset+entiry.Length])

		sb.WriteString(
			fmt.Sprintf(
				"%s[%s]",
				text1,
				entiry.URL,
			),
		)
		generalOffset = entiry.Offset + entiry.Length
	}

	return sb.String()
}
