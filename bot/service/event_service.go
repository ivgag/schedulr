package service

import (
	"github.com/ivgag/schedulr/ai"
	"github.com/ivgag/schedulr/domain"
)

func NewEventService(
	ai ai.AI,
	userService UserService,
	calendarService CalendarService,
) *EventService {
	return &EventService{
		ai:              ai,
		userService:     userService,
		calendarService: calendarService,
	}
}

type EventService struct {
	ai              ai.AI
	userService     UserService
	calendarService CalendarService
}

func (s *EventService) CreateEventsFromUserMessage(telegramID int64, message domain.UserMessage) ([]domain.Event, error) {
	user, err := s.userService.GetUserByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}

	events, err := s.ai.GetEvents(&message)
	if err != nil {
		return nil, err
	}

	for i := range events {
		err = s.calendarService.CreateEvent(user.ID, domain.ProviderGoogle, events[i])
		if err != nil {
			return nil, err
		}
	}

	return events, nil
}
