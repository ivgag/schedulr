package ai

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/utils"
	openai "github.com/sashabaranov/go-openai"
)

func NewOpenAI() (*OpenAI, error) {
	openAiToken, err := utils.GetenvOrError("OPEN_AI_API_KEY")
	if err != nil {
		return nil, err
	}

	client := openai.NewClient(openAiToken)
	return &OpenAI{client: client}, nil
}

type OpenAI struct {
	client *openai.Client
}

func (o *OpenAI) Provider() AIProvider {
	return ProviderOpenAI
}

func (o *OpenAI) ExtractCalendarEvents(message *model.TextMessage) ([]model.Event, model.Error) {
	userInput := messagesToText([]model.TextMessage{*message})

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: extractCalendarEventsPrompt(),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userInput,
				},
			},
		},
	)

	e := &openai.APIError{}
	if errors.As(err, &e) {
		switch e.HTTPStatusCode {
		case 500, 503:
			return nil, ApiError{
				Message:      e.Message,
				ResponseCode: e.HTTPStatusCode,
				Retryable:    true,
			}
		default:
			return nil, ApiError{
				Message:      e.Message,
				ResponseCode: e.HTTPStatusCode,
				Retryable:    false,
			}
		}
	} else if err != nil {
		return nil, err
	} else {
		responseContent := resp.Choices[0].Message.Content

		var events []model.Event
		err = json.Unmarshal([]byte(removeJsonFormattingMarkers(responseContent)), &events)
		if err != nil {
			return nil, err
		}

		return events, nil
	}
}
