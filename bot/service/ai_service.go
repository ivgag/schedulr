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

	"github.com/cenkalti/backoff/v5"
	"github.com/rs/zerolog/log"

	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/model"
)

func NewAIService(ais []ai.AI) *AIService {

	aisMap := make(map[ai.AIProvider]ai.AI)
	for _, ai := range ais {
		aisMap[ai.Provider()] = ai
	}

	return &AIService{
		aisMap: aisMap,
	}
}

type AIService struct {
	aisMap map[ai.AIProvider]ai.AI
}

func (s *AIService) ExtractCalendarEvents(message *model.TextMessage) ([]model.Event, model.Error) {
	for _, ai := range s.aisMap {
		log.Debug().
			Str("provider", string(ai.Provider())).
			Msg("Extracting events with AI provider")

		events, err := s.extractEventsWithRetires(message, ai)
		if err != nil {
			log.Warn().
				Str("provider", string(ai.Provider())).
				Err(err).
				Msg("AI provider failed to extract events from the message")
		} else if len(events) == 0 {
			log.Warn().
				Str("provider", string(ai.Provider())).
				Interface("message", message).
				Msg("AI provider extracted no events from the message")

			return nil, model.ErrorForMessage("No events were extracted from the message")
		} else {
			log.Debug().
				Str("provider", string(ai.Provider())).
				Msg("AI provider successfully extracted events from the message")

			return events, nil
		}
	}

	return nil, model.ErrorForMessage("No AI provider was able to extract events from the message")
}

func (s *AIService) extractEventsWithRetires(message *model.TextMessage, agent ai.AI) ([]model.Event, model.Error) {
	operation := func() ([]model.Event, error) {
		var apiError = ai.ApiError{}
		events, err := agent.ExtractCalendarEvents(message)
		if err == nil {
			return events, nil
		} else if errors.As(err, &apiError) && apiError.Retryable {
			return nil, err
		} else {
			return events, backoff.Permanent(err)
		}
	}

	events, err := backoff.Retry(
		context.Background(),
		operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(3),
	)

	if err != nil {
		return nil, err
	}
	return events, nil
}
