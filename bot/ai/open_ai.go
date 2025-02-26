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

package ai

import (
	"context"
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/utils"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func NewOpenAI(config *OpenAIConfig) *OpenAI {
	client := openai.NewClient(config.APIKey)
	return &OpenAI{
		config: config,
		client: client,
	}
}

type OpenAI struct {
	config *OpenAIConfig
	client *openai.Client
}

func (o *OpenAI) Provider() AIProvider {
	return ProviderOpenAI
}

func (o *OpenAI) ExtractCalendarEvents(timeZone string, messages *[]model.TextMessage) (*AiResponse[[]model.Event], model.Error) {
	var response AiResponse[[]model.Event]
	var schema AiResponse[[]EventSchema]
	responseSchema, err := jsonschema.GenerateSchemaForType(schema)

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: o.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: extractCalendarEventsPrompt(utils.Now(timeZone)),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: messagesToText(messages),
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "extracted_events",
					Schema: responseSchema,
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
	} else {
		responseContent := resp.Choices[0].Message.Content

		err = responseSchema.Unmarshal(responseContent, &response)
		if err != nil {
			log.Error().
				Interface("messages", messages).
				Str("responseContent", responseContent).
				Err(err).Msg("Failed to unmarshal OpenAI response")
		}

		return &response, err
	}
}

type OpenAIConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}
