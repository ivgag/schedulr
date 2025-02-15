package ai

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/utils"
	"github.com/sashabaranov/go-openai/jsonschema"

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
	var result []model.Event
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	outputSchema := "Output format: " + string(jsonBytes)

	request := &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleSystem, Content: extractCalendarEventsPrompt()},
			{Role: constants.ChatMessageRoleSystem, Content: outputSchema},
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
	}

	err = schema.Unmarshal(response.Choices[0].Message.Content, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
