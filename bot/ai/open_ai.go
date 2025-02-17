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
