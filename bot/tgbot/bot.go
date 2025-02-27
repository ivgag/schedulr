package tgbot

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
	"github.com/rs/zerolog/log"
)

// TelegramBotConfig remains unchanged.
type TelegramBotConfig struct {
	Token string `mapstructure:"token"`
	URL   string `mapstructure:"url"`
}

// Bot wraps bot.Bot with service dependencies and buffering fields.
type Bot struct {
	ctx             context.Context
	cfg             *TelegramBotConfig
	chatBot         *bot.Bot
	userService     *service.UserService
	eventService    *service.EventService
	bufferedUpdates map[int64][]*models.Update // Keyed by chat ID.
	bufferTimers    map[int64]*time.Timer      // Timers per chat.
	bufferMutex     sync.Mutex                 // Mutex for bufferedUpdates and bufferTimers.
}

// NewBot initializes the bot and its buffers.
func NewBot(
	ctx context.Context,
	cfg *TelegramBotConfig,
	userService *service.UserService,
	eventService *service.EventService,
) *Bot {
	return &Bot{
		ctx:             ctx,
		cfg:             cfg,
		userService:     userService,
		eventService:    eventService,
		bufferedUpdates: make(map[int64][]*models.Update),
		bufferTimers:    make(map[int64]*time.Timer),
	}
}

// Start and other methods remain unchanged.
func (b *Bot) Start() error {
	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDebugHandler(debugHandler),
		bot.WithErrorsHandler(errorsHandler),
		bot.WithDefaultHandler(b.defaultHandler),
	}

	chatBot, err := bot.New(b.cfg.Token, opts...)
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

	user := storage.User{
		TelegramID: chatID,
		Username:   update.Message.From.Username,
	}

	if err := b.userService.CreateUser(&user); err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}
	b.sendMessage(ctx, chatID, "Link your Calendar: /linkgoogle", "")
}

// linkGoogleAccountHandler returns the OAuth2 URL for linking a Google account.
func (b *Bot) linkGoogleAccountHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	link, err := b.userService.GetOAuth2Url(
		chatID,
		func(err error) {
			if err != nil {
				b.sendMessage(ctx, chatID, err.Error(), "")
			} else {
				b.sendMessage(ctx, chatID, "Account linked successfully :)", "")
			}
		},
		model.ProviderGoogle,
	)
	if err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}
	b.sendMessage(ctx, chatID, "Link your Google Calendar: "+link, "")
}

// defaultHandler now checks if the message is forwarded and buffers it.
func (b *Bot) defaultHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	b.bufferUpdate(ctx, update)
	// If the message has no text or caption, do nothing.
}

// bufferForwardedText buffers forwarded text messages by chat.
func (b *Bot) bufferUpdate(ctx context.Context, update *models.Update) {
	chatID := update.Message.Chat.ID
	b.bufferMutex.Lock()
	defer b.bufferMutex.Unlock()

	// Append the message to the buffer.
	b.bufferedUpdates[chatID] = append(b.bufferedUpdates[chatID], update)

	// Reset the timer if it already exists.
	if timer, exists := b.bufferTimers[chatID]; exists {
		timer.Stop()
	}
	// Start a new timer. After 3 seconds of inactivity, process the buffered messages.
	b.bufferTimers[chatID] = time.AfterFunc(3*time.Second, func() {
		b.processBufferedMessages(ctx, chatID)
	})
}

// processBufferedMessages processes all buffered messages for a chat.
func (b *Bot) processBufferedMessages(ctx context.Context, chatID int64) {
	b.bufferMutex.Lock()

	messages := b.bufferedUpdates[chatID]

	delete(b.bufferedUpdates, chatID)
	delete(b.bufferTimers, chatID)
	b.bufferMutex.Unlock()

	if len(messages) == 0 {
		return
	}

	textMessages := make([]model.TextMessage, len(messages))
	for i, msg := range messages {
		textMessages[i] = updateToMessage(msg)
	}

	createdEvents, err := b.eventService.CreateEventsFromUserMessage(chatID, textMessages)
	if err != nil {
		log.Error().
			Int64("chatID", chatID).
			Err(err).
			Msg("Failed to create events")
		b.sendMessage(ctx, chatID, "Failed to create events. Try later.", "")
	} else if len(createdEvents) == 0 {
		b.sendMessage(ctx, chatID, "No events found in forwarded messages.", "")
	} else {
		for _, event := range createdEvents {
			b.sendMessage(ctx, chatID, formatEventForTelegram(event), models.ParseModeMarkdownV1)
		}
	}
}

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

func formatEventForTelegram(scheduledEvent model.ScheduledEvent) string {
	event := scheduledEvent.Event

	message := fmt.Sprintf("*%s*\n", event.Title)
	if event.Description != "" {
		message += fmt.Sprintf("%s\n", event.Description)
	}
	message += fmt.Sprintf("*When:* %s - %s\n",
		event.Start.Format(time.DateTime),
		event.End.Format(time.DateTime),
	)
	if event.Location != "" {
		message += fmt.Sprintf("*Where:* %s\n", event.Location)
	}
	if scheduledEvent.Link != "" {
		message += fmt.Sprintf("[More details](%s)\n", scheduledEvent.Link)
	}
	return message
}

func debugHandler(format string, args ...interface{}) {
	log.Debug().Msg(fmt.Sprintf(format, args...))
}

func errorsHandler(err error) {
	log.Error().Err(err).Msg("Telegram bot error")
}

func updateToMessage(update *models.Update) model.TextMessage {
	var msgType model.MessageType
	if update.Message.Text != "" {
		msgType = model.UserMessage
	} else if update.Message.Caption != "" {
		msgType = model.ForwardedMessage
	} else {
		return model.TextMessage{}
	}

	var from string
	var text string
	var entities []models.MessageEntity

	switch msgType {
	case model.UserMessage:
		from = update.Message.From.Username
		text = update.Message.Text
		entities = update.Message.Entities
	case model.ForwardedMessage:
		from = update.Message.From.Username
		text = update.Message.Caption
		entities = update.Message.CaptionEntities
	}

	var matterEntities = make([]models.MessageEntity, 0)
	for _, entity := range entities {
		if entity.Type == models.MessageEntityTypeTextLink ||
			entity.Type == models.MessageEntityTypeTextMention ||
			entity.Type == models.MessageEntityTypeURL {
			matterEntities = append(matterEntities, entity)
		}
	}

	return model.TextMessage{
		From:        from,
		Text:        FormatMessageText(text, matterEntities),
		MessageType: msgType,
	}
}

func FormatMessageText(text string, entities []models.MessageEntity) string {
	if len(entities) == 0 {
		return text
	}

	var sb strings.Builder
	sort.SliceStable(entities, func(i, j int) bool {
		return entities[i].Offset < entities[j].Offset
	})

	runeArray := []rune(text)
	generalOffset := 0

	for _, entity := range entities {
		part := string(runeArray[generalOffset : entity.Offset+entity.Length])
		sb.WriteString(fmt.Sprintf("%s[%s]", part, entity.URL))
		generalOffset = entity.Offset + entity.Length
	}

	sb.WriteString(string(runeArray[generalOffset:]))
	return sb.String()
}
