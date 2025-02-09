package service

import (
	"github.com/ivgag/schedulr/ai"
)

func NewEventService(ai ai.AI) *EventService {
	return &EventService{
		ai: ai,
	}
}

type EventService struct {
	ai ai.AI
}

func (s *EventService) CreateEventsFromUserMessage(message ai.UserMessage) ([]ai.Event, error) {
	return s.ai.GetEvents(&message)
}
