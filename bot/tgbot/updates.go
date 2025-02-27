package tgbot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-telegram/bot/models"
	"github.com/ivgag/schedulr/model"
)

func UpdateToMessage(update *models.Update) model.TextMessage {
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
		Text:        FormatTextWithEntities(text, matterEntities),
		MessageType: msgType,
	}
}

func FormatTextWithEntities(text string, entities []models.MessageEntity) string {
	if len(entities) == 0 {
		return text
	}

	var sb strings.Builder
	// Sort entities by starting offset.
	sort.SliceStable(entities, func(i, j int) bool {
		return entities[i].Offset < entities[j].Offset
	})

	runeArray := []rune(text)
	generalOffset := 0

	for _, entity := range entities {
		// Add text before the current entity.
		if entity.Offset > generalOffset {
			sb.WriteString(string(runeArray[generalOffset:entity.Offset]))
		}

		// Add formatted text for the entity.
		entityText := string(runeArray[entity.Offset : entity.Offset+entity.Length])
		sb.WriteString(fmt.Sprintf("%s[%s]", entityText, entity.URL))
		generalOffset = entity.Offset + entity.Length
	}

	// Append remaining text.
	sb.WriteString(string(runeArray[generalOffset:]))
	return sb.String()
}
