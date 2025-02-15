package service

import (
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
		events, err := ai.ExtractCalendarEvents(message)
		if err == nil {
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
