package ai

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
	constants "github.com/cohesion-org/deepseek-go/constants"
)

func NewDeepSeekAI() (*DeepSeekAI, error) {
	apiKey, err := utils.GetenvOrError("DEEPSEEK_API_KEY")
	if err != nil {
		return nil, err
	}

	client := deepseek.NewClient(apiKey)

	return &DeepSeekAI{
		client: client,
	}, nil
}

type DeepSeekAI struct {
	client *deepseek.Client
}

func (d *DeepSeekAI) Provider() AIProvider {
	return ProviderDeepSeek
}

func (d *DeepSeekAI) ExtractCalendarEvents(message *model.TextMessage) ([]model.Event, model.Error) {
	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleSystem, Content: extractCalendarEventsPrompt()},
			{Role: constants.ChatMessageRoleUser, Content: messagesToText([]model.TextMessage{*message})},
		},
		ResponseFormat: &deepseek.ResponseFormat{
			Type: "json_object",
		},
	}

	ctx := context.Background()
	response, err := d.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	e := &deepseek.APIError{}
	if errors.As(err, &e) {
		switch e.StatusCode {
		case 500, 503:
			return nil, ApiError{
				Message:      e.Message,
				ResponseCode: e.StatusCode,
				Retryable:    true,
			}
		default:
			return nil, ApiError{
				Message:      e.Message,
				ResponseCode: e.StatusCode,
				Retryable:    false,
			}
		}
	} else if err != nil {
		return nil, err
	} else {
		responseContent := response.Choices[0].Message.Content

		var events []model.Event
		err = json.Unmarshal([]byte(removeJsonFormattingMarkers(responseContent)), &events)
		if err != nil {
			return nil, err
		}

		return events, nil
	}
}
