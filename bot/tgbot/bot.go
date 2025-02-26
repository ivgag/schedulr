/*
 * Created on Mon Feb 17 2025
 *
 *  Copyright (c) 2025 Ivan Gagarkin
 * SPDX-License-Identifier: EPL-2.0
 *
 * Licensed under the Eclipse Public License - v 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.eclipse.org/legal/epl-2.0/
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	b.chatBot.RegisterHandler(bot.HandlerTypeMessageText, "/settings", bot.MatchTypeExact, b.settingsHandler)

	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "linkg_google", bot.MatchTypeExact, b.linkGoogleAccountHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "unlink_google", bot.MatchTypeExact, b.unlinkGoogleAccountHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "share_location", bot.MatchTypeExact, b.shareLocationHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "set_timezone", bot.MatchTypeExact, b.shareLocationHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypeContains, b.locationHandler)

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

	user := model.User{
		TelegramID: chatID,
		Username:   update.Message.From.Username,
	}

	if err := b.userService.CreateUser(&user); err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}

	b.sendMessage(ctx, chatID, "Welcome to Schedulr! \nSet up your profile.", "")
	b.sendSettingsKeyboard(ctx, botAPI, update)
}

// settingsHandler sends a message with an inline keyboard.
func (b *Bot) settingsHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	b.sendSettingsKeyboard(ctx, botAPI, update)
}

func (b *Bot) sendSettingsKeyboard(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	userProfile, err := b.userService.GetUserProfileByTelegramID(chatID)
	if err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}

	var buttons [][]models.InlineKeyboardButton = make([][]models.InlineKeyboardButton, 0)

	googleAccount := userProfile.LinkedAccount(model.ProviderGoogle)
	if googleAccount != nil {
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         "Unlink Google Account",
				CallbackData: "linkg_google",
			},
		})
	} else {
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         "Link Google Account",
				CallbackData: "unlink_google",
			},
		})
	}

	buttons = append(buttons, []models.InlineKeyboardButton{
		{
			Text:         "Share Location",
			CallbackData: "share_location",
		},
		{
			Text:         "Set Timezone",
			CallbackData: "set_timezone",
		},
	})

	// Create an inline keyboard with three buttons.
	inlineKeyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text: `Settings:

		- Google Account: ` + googleAccount.AccountName + `
		- Timezone: ` + userProfile.User.TimeZone + `
		`,
		ReplyMarkup: inlineKeyboard,
	}

	botAPI.SendMessage(ctx, params)
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

func (b *Bot) unlinkGoogleAccountHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	deleted, err := b.userService.UnlinkAccount(chatID, model.ProviderGoogle)
	if err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}

	if deleted {
		b.sendMessage(ctx, chatID, "Google account unlinked successfully.", "")
	} else {
		b.sendMessage(ctx, chatID, "Google account not linked.", "")

	}
}

func (b *Bot) shareLocationHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	btn := models.KeyboardButton{
		RequestLocation: true,
		Text:            "Share Location to set Timezone. Works only on mobile.",
	}

	markup := models.ReplyKeyboardMarkup{
		Keyboard:        [][]models.KeyboardButton{{btn}},
		OneTimeKeyboard: true,
	}

	msg := &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.From.ID,
		Text:        "Share your location",
		ReplyMarkup: markup,
	}

	botAPI.SendMessage(ctx, msg)
}

func (b *Bot) locationHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	timeZone, err := b.userService.UpdateUserTimeZone(
		update.Message.Chat.ID,
		update.Message.Location.Latitude,
		update.Message.Location.Longitude,
	)

	if err != nil {
		b.sendMessage(ctx, update.Message.Chat.ID, err.Error(), "")
		return
	}

	b.sendMessage(ctx, update.Message.Chat.ID, "Timezone set to "+timeZone, "")
}

// defaultHandler now checks if the message is forwarded and buffers it.
func (b *Bot) defaultHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	if update.Message.Location != nil {
		b.locationHandler(ctx, botAPI, update)
	} else if update.Message.Caption != "" || update.Message.Text != "" {
		b.bufferUpdate(ctx, update)
	}
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
