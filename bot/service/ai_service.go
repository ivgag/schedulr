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
		if err == nil {
			log.Debug().
				Str("provider", string(ai.Provider())).
				Msg("AI provider successfully extracted events from the message")
			return events, nil
		} else {
			log.Warn().
				Str("provider", string(ai.Provider())).
				Err(err).
				Msg("AI provider failed to extract events from the message")
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
