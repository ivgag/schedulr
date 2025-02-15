package ai

import (
	"context"
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/utils"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
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

	var result []model.Event
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
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
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "extractedEvents",
					Schema: schema,
					Strict: true,
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
	}

	if err != nil {
		return nil, err
	}

	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
