package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ivgag/schedulr/model"
	openai "github.com/sashabaranov/go-openai"
)

var openApiClientTokenEnv = "AI_SERVICE_CLIENT"

type OpenAI struct {
	client *openai.Client
}

func (o *OpenAI) GetEvents(message *model.TextMessage) ([]model.Event, error) {
	prompt := fmt.Sprintf(`
		You are a calendar assistant that extracts event details from user input 
		(which may include announcements, tickets, ads, and other related content) and 
		converts them into JSON for creating calendar events (e.g., in Google, Microsoft, Yandex).

		Your Tasks
		1. Extract key event details:
			• Title
			• Description (including critical details like price, links, host’s name, etc.)
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

		Output format
		Your output must be a JSON array. Each event is represented as an object of the form:

		[
		{
			"title": "Event Title",
			"description": "A well-formatted brief description that includes all critical details (price, links, host’s name, etc.).",
			"start": {
			"dateTime": "YYYY-MM-DDTHH:MM:SSZ"
			},
			"end": {
			"dateTime": "YYYY-MM-DDTHH:MM:SSZ"
			},
			"location": "Event Location",
			"eventType": "announcement"
		}
		]
		`,
		time.Now().Format(time.DateTime),
	)

	userInput := messagesToText([]model.TextMessage{*message})

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userInput,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	responseContent := resp.Choices[0].Message.Content

	var events []model.Event
	err = json.Unmarshal([]byte(removeJsonFormattingMarkers(responseContent)), &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func NewOpenAI() (*OpenAI, error) {
	openAiToken := os.Getenv(openApiClientTokenEnv)
	if openAiToken == "" {
		return nil, errors.New(openApiClientTokenEnv + " is not set")
	}

	client := openai.NewClient(openAiToken)
	return &OpenAI{client: client}, nil
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
