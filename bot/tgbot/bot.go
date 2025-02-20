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

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/storage"
	"github.com/rs/zerolog/log"
)

// NewBot initializes the Telegram bot with the provided context and services.
func NewBot(
	ctx context.Context,
	cfg *TelegramBotConfig,
	userService *service.UserService,
	eventService *service.EventService,
) *Bot {
	return &Bot{
		ctx:          ctx,
		cfg:          cfg,
		userService:  userService,
		eventService: eventService,
	}
}

// Bot wraps bot.Bot with service dependencies.
type Bot struct {
	ctx          context.Context
	cfg          *TelegramBotConfig
	chatBot      *bot.Bot
	userService  *service.UserService
	eventService *service.EventService
}

// Start initializes the Telegram bot, registers handlers, and starts update processing.
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

	if err := b.userService.CreateUser(&storage.User{TelegramID: chatID}); err != nil {
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
func formatEventForTelegram(scheduledEvent model.ScheduledEvent) string {
	event := scheduledEvent.Event

	message := fmt.Sprintf("*%s*\n", event.Title)
	if event.Description != "" {
		message += fmt.Sprintf("%s\n", event.Description)
	}
	message += fmt.Sprintf("*When:* %s - %s\n",
		event.Start.String(),
		event.End.String(),
	)
	if event.Location != "" {
		message += fmt.Sprintf("*Where:* %s\n", event.Location)
	}
	if scheduledEvent.Link != "" {
		message += fmt.Sprintf("[More details](%s)\n", scheduledEvent.Link)
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

func debugHandler(format string, args ...interface{}) {
	log.Debug().Msg(fmt.Sprintf(format, args...))
}

func errorsHandler(err error) {
	log.Error().Err(err).Msg("Telegram bot error")
}

type TelegramBotConfig struct {
	Token string `mapstructure:"token"`
	URL   string `mapstructure:"url"`
}
