package ai

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ivgag/schedulr/domain"
	openai "github.com/sashabaranov/go-openai"
)

var openApiClientTokenEnv = "AI_SERVICE_CLIENT"

type OpenAI struct {
	client *openai.Client
}

func (o *OpenAI) GetEvents(message *domain.UserMessage) ([]domain.Event, error) {
	// Use the OpenAI API to generate events based on the user message.

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: `
						You are a calendar assistant that extracts event details from user input 
						and converts them into JSON for creating calendar events (Google, Microsoft, Yandex, etc.). 
						The input may include announcements, tickets, ads, and similar content.

						Tasks:
						 - Extract key event details: title, description (including critical details like price, 
						 	links, host name, etc.), start time, end time, location, and event type.
						 - Resolve relative dates using the reference date: today is ` + time.Now().String() +
						`
						 - If an event’s details cannot be fully extracted, ignore that event.
						 - If no event details are found, return an empty JSON array.

						Output format:

						[
						{
							"title": "Event Title",
							"description": "A brief description including all critical details (price, links, host's name, etc.)",
							"start": {
							"dateTime": "YYYY-MM-DD'T'HH:mm:ss"
							},
							"end": {
							"dateTime": "YYYY-MM-DD'T'HH:mm:ss"
							},
							"location": "Event Location",
							"eventType": "announcement"
						}
						]
						`,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message.Text + " " + message.Caption,
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	responseContent := resp.Choices[0].Message.Content
	log.Println("OpenAI response:", responseContent)

	var events []domain.Event
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
	// Remove formatting markers (` ```json` and trailing triple backticks)

	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")

	return text
}
