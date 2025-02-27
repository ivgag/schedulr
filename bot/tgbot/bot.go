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
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/service"
	"github.com/rs/zerolog/log"
)

const (
	startCommand    = "/start"
	settingsCommand = "/settings"

	shareLocationCallback  = "share_location"
	updateCalendarCallback = "update_calendar"

	selectCalendarPrefix = "select_calendar_"
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

	b.chatBot.RegisterHandler(bot.HandlerTypeMessageText, startCommand, bot.MatchTypeExact, b.startHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeMessageText, settingsCommand, bot.MatchTypeExact, b.settingsHandler)

	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, shareLocationCallback, bot.MatchTypeExact, b.shareLocationHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, updateCalendarCallback, bot.MatchTypeExact, b.updateCalendarHandler)
	b.chatBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, selectCalendarPrefix, bot.MatchTypePrefix, b.selectCalendarHandler)

	b.chatBot.Start(b.ctx)

	b.chatBot.SetMyCommands(
		b.ctx,
		&bot.SetMyCommandsParams{
			Commands: []models.BotCommand{
				{
					Command:     startCommand,
					Description: "Start the bot",
				},
				{
					Command:     settingsCommand,
					Description: "Show settings",
				},
			},
		},
	)

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
	b.sendSettingsKeyboard(ctx, botAPI, update.Message.From.ID)
}

// settingsHandler sends a message with an inline keyboard.
func (b *Bot) settingsHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	b.sendSettingsKeyboard(ctx, botAPI, update.Message.From.ID)
}

func (b *Bot) sendSettingsKeyboard(ctx context.Context, botAPI *bot.Bot, chatID int64) {

	user, err := b.userService.GetUserByTelegramID(chatID)
	if err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}

	var buttons [][]models.InlineKeyboardButton = make([][]models.InlineKeyboardButton, 0)
	buttons = append(buttons, []models.InlineKeyboardButton{
		{
			Text:         "Share Location",
			CallbackData: shareLocationCallback,
		},
		{
			Text:         "Update Preferred Calendar",
			CallbackData: updateCalendarCallback,
		},
	})

	inlineKeyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	sb := strings.Builder{}
	sb.WriteString("Settings:")
	sb.WriteString("\n - Preferred Calendar: ")
	sb.WriteString(string(user.PreferredCalendar))
	sb.WriteString("\n - Time Zone: ")
	sb.WriteString(user.TimeZone)

	params := &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        sb.String(),
		ReplyMarkup: inlineKeyboard,
	}

	botAPI.SendMessage(ctx, params)
}

func (b *Bot) shareLocationHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	btn := models.KeyboardButton{
		RequestLocation: true,
		Text:            "Share Location",
	}

	markup := models.ReplyKeyboardMarkup{
		Keyboard:        [][]models.KeyboardButton{{btn}},
		OneTimeKeyboard: true,
	}

	msg := &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.From.ID,
		Text:        "Share your location to set your time zone.",
		ReplyMarkup: markup,
	}

	botAPI.SendMessage(ctx, msg)
}

func (b *Bot) updateCalendarHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	var buttons [][]models.InlineKeyboardButton = make([][]models.InlineKeyboardButton, 0)
	for _, calendar := range model.CalendarTypes() {
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         string(calendar),
				CallbackData: string(selectCalendarPrefix + calendar),
			},
		})
	}

	inlineKeyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	params := &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Select your preferred calendar.",
		ReplyMarkup: inlineKeyboard,
	}

	botAPI.SendMessage(ctx, params)
}

func (b *Bot) selectCalendarHandler(ctx context.Context, botAPI *bot.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	calendar := strings.TrimPrefix(update.CallbackQuery.Data, selectCalendarPrefix)

	err := b.userService.UpdateUserPreferredCalendar(chatID, model.Calendar(calendar))
	if err != nil {
		b.sendMessage(ctx, chatID, err.Error(), "")
		return
	}

	b.sendMessage(ctx, chatID, "Preferred calendar set to "+calendar, "")
	b.sendSettingsKeyboard(ctx, botAPI, chatID)
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
	b.settingsHandler(ctx, botAPI, update)
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
		textMessages[i] = UpdateToMessage(msg)
	}

	createdEvents, err := b.eventService.CreateEventsFromUserMessage(chatID, &textMessages)
	if err != nil {
		log.Error().
			Int64("chatID", chatID).
			Err(err).
			Msg("Failed to create events")
		b.sendMessage(ctx, chatID, "Failed to create events. Try later.", "")
	} else if len(*createdEvents) == 0 {
		b.sendMessage(ctx, chatID, "No events found in forwarded messages.", "")
	} else {
		for _, scheduledEvent := range *createdEvents {
			b.sendMessage(ctx, chatID, FormatScheduledEvent(&scheduledEvent), models.ParseModeMarkdown)
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

func debugHandler(format string, args ...interface{}) {
	log.Debug().Msg(fmt.Sprintf(format, args...))
}

func errorsHandler(err error) {
	log.Error().Err(err).Msg("Telegram bot error")
}

func eventButtons(scheduledEvent *model.ScheduledEvent) []models.InlineKeyboardButton {
	var buttons []models.InlineKeyboardButton = make([]models.InlineKeyboardButton, 0)

	if scheduledEvent.Link != "" {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text: "Link",
			URL:  scheduledEvent.Link,
		})
	} else if scheduledEvent.Event.DeepLink != "" {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text: "Add to Calendar",
			URL:  scheduledEvent.Event.DeepLink,
		})
	}

	return buttons
}
