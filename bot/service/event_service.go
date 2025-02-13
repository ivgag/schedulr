package service

import (
	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/model"
)

func NewEventService(
	ai ai.AI,
	userService UserService,
	clanedarServices map[model.Provider]CalendarService,
) *EventService {
	return &EventService{
		ai:               ai,
		userService:      userService,
		calendarServices: clanedarServices,
	}
}

type EventService struct {
	ai               ai.AI
	userService      UserService
	calendarServices map[model.Provider]CalendarService
}

func (s *EventService) CreateEventsFromUserMessage(telegramID int64, message model.UserMessage) ([]model.Event, error) {
	user, err := s.userService.GetUserByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}

	events, err := s.ai.GetEvents(&message)
	if err != nil {
		return nil, err
	}

	for i := range events {
		_, err := s.calendarServices[model.ProviderGoogle].CreateEvent(user.ID, &events[i])
		if err != nil {
			return nil, err
		}
	}

	return events, nil
}
