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

package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/rs/zerolog/log"

	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/model"
)

func NewAIService(
	ais []ai.AI,
	config *AIConfig,
) *AIService {
	aisMap := make(map[string]ai.AI)
	for _, ai := range ais {
		aisMap[strings.ToLower(string(ai.Provider()))] = ai
	}

	return &AIService{
		aisMap: aisMap,
		config: config,
	}
}

type AIService struct {
	aisMap map[string]ai.AI
	config *AIConfig
}

func (s *AIService) ExtractCalendarEvents(
	timeZone string,
	messages *[]model.TextMessage,
) (*[]model.Event, model.Error) {
	for _, service := range s.config.Priority {
		ai, ok := s.aisMap[strings.ToLower(service)]
		if !ok {
			continue
		}

		log.Debug().
			Interface("messages", messages).
			Str("provider", string(ai.Provider())).
			Msg("Extracting events with AI provider")

		response, err := s.extractEventsWithRetires(timeZone, messages, ai)
		if err != nil {
			log.Warn().
				Interface("messages", messages).
				Str("provider", string(ai.Provider())).
				Err(err).
				Msg("AI provider failed to extract events from the message")
		}

		log.Debug().
			Interface("messages", messages).
			Interface("response", response).
			Str("provider", string(ai.Provider())).
			Msg("AI provider successfully extracted events from the message")

		return &response.Result, nil
	}

	return nil, model.ErrorForMessage("No AI provider was able to extract events from the message")
}

func (s *AIService) extractEventsWithRetires(
	timeZone string,
	messages *[]model.TextMessage,
	agent ai.AI,
) (*ai.AiResponse[[]model.Event], model.Error) {
	operation := func() (ai.AiResponse[[]model.Event], error) {
		var apiError = ai.ApiError{}
		response, err := agent.ExtractCalendarEvents(&ai.ExtractCalendarEventsRequest{
			Now:      nowInTimezone(timeZone),
			Calendar: model.CalendarGoogle,
		}, messages)
		if err == nil {
			return *response, nil
		} else if errors.As(err, &apiError) && apiError.Retryable {
			return ai.AiResponse[[]model.Event]{}, err
		} else {
			return *response, backoff.Permanent(err)
		}
	}

	response, err := backoff.Retry(
		context.Background(),
		operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(3),
	)

	if err != nil {
		return nil, err
	}
	return &response, nil
}

type AIConfig struct {
	Deepseek ai.DeepseekConfig `mapstructure:"deepseek"`
	OpenAI   ai.OpenAIConfig   `mapstructure:"openai"`
	Priority []string          `mapstructure:"priority"`
}

func nowInTimezone(timeZone string) time.Time {
	if timeZone == "" {
		return time.Now()
	}

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		log.Error().
			Str("timezone", timeZone).
			Err(err).
			Msg("Failed to load timezone")
		return time.Now()
	}
	return time.Now().In(loc)
}
