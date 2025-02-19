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
	"encoding/json"
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/rs/zerolog/log"

	deepseek "github.com/cohesion-org/deepseek-go"
	constants "github.com/cohesion-org/deepseek-go/constants"
)

func NewDeepSeekAI(config *DeepseekConfig) *DeepSeekAI {
	client := deepseek.NewClient(config.APIKey)

	return &DeepSeekAI{
		client: client,
	}
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
	} else {
		responseContent := response.Choices[0].Message.Content

		var events []model.Event
		err = json.Unmarshal([]byte(removeJsonFormattingMarkers(responseContent)), &events)
		if err != nil {
			log.Error().
				Str("responseContent", responseContent).
				Err(err).Msg("Failed to unmarshal DeepSeek response")

			return nil, err
		}

		return events, nil
	}
}

type DeepseekConfig struct {
	APIKey string `mapstructure:"api_key"`
}
