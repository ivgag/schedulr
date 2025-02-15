package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivgag/schedulr/model"
)

type AIProvider string

const (
	ProviderOpenAI   AIProvider = "OpenAI"
	ProviderDeepSeek AIProvider = "DeepSeek"
)

type AI interface {
	ExtractCalendarEvents(message *model.TextMessage) ([]model.Event, model.Error)
	Provider() AIProvider
}

func extractCalendarEventsPrompt() string {
	return fmt.Sprintf(`
	You are a calendar assistant that extracts event details from user input 
	(which may include announcements, tickets, ads, and other related content) and 
	converts them into JSON for creating calendar events (e.g., in Google, Microsoft, Yandex).

	Your Tasks
	1. Extract key event details:
		• Title
		• Well-formatted brief description (including critical details like price, links, host’s name, etc.)
		• Start date/time
		• End date/time
		• Location
		• Event type
	2. Resolve relative dates using the reference date:
		> "Today is %s"
	3. If an event’s details cannot be fully extracted, ignore that event.
	4. If no event details are found, return an empty JSON array.

	Input Format
	• The user input may include Telegram messages, either single or multiple messages.
	• Messages may be:
		• Forwarded from an events channel in the user’s city.
		• A forwarded conversation between users.
		• Forwarded messages plus a command to the bot.
	• You need to parse all incoming text to find any event-related information.
	`,
		time.Now().Format(time.DateTime),
	)
}

func removeJsonFormattingMarkers(text string) string {
	// Remove formatting markers (```json and trailing backticks)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")
	return text
}

// messagesToText converts an array of TextMessages into a single string.
// It uses formatMessageText to include entity markers correctly.
func messagesToText(messages []model.TextMessage) string {
	var sb strings.Builder

	for _, msg := range messages {
		switch msg.MessageType {
		case model.UserMessage:
			sb.WriteString(fmt.Sprintf("%s: %s\n", msg.From, msg.Text))
		case model.ForwardedMessage:
			sb.WriteString(fmt.Sprintf("Forwarded from %s: %s\n", msg.From, msg.Text))
		}
	}

	return sb.String()
}
