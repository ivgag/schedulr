package ai

import (
	"context"
	"errors"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

var openApiClientTokenEnv = "AI_SERVICE_CLIENT"

type OpenAI struct {
	client *openai.Client
}

func (o *OpenAI) GetEvents(message *UserMessage) ([]Event, error) {
	// Use the OpenAI API to generate events based on the user message.

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: `
						You are a trained calendar assistant.
						The user will send you information about one or more events.
						The information might include announcements, tickets, ads, etc.
						Extract details from the user's input and transform them into JSON to 
						later create an appointment on a calendar (Google, Microsoft, Yandex, etc.).
						If you cannot extract details for any event, return an empty string.

						The output structure is:

						{
						"events": [
							{
							"title": "Event Title",
							"description": "A brief description of the event",
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
						}
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

	log.Println("OpenAI response:", resp.Choices[0].Message.Content)

	return []Event{}, nil
}

func NewOpenAI() (*OpenAI, error) {
	openAiToken := os.Getenv(openApiClientTokenEnv)
	if openAiToken == "" {
		return nil, errors.New(openApiClientTokenEnv + " is not set")
	}

	client := openai.NewClient(openAiToken)

	return &OpenAI{client: client}, nil
}
