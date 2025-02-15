package service

import (
	"github.com/ivgag/schedulr/model"
)

func NewEventService(
	aiService AIService,
	userService UserService,
	clanedarServices map[model.Provider]CalendarService,
) *EventService {
	return &EventService{
		aiService:        aiService,
		userService:      userService,
		calendarServices: clanedarServices,
	}
}

type EventService struct {
	aiService        AIService
	userService      UserService
	calendarServices map[model.Provider]CalendarService
}

func (s *EventService) CreateEventsFromUserMessage(telegramID int64, message model.TextMessage) ([]model.Event, error) {
	user, err := s.userService.GetUserByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}

	events, err := s.aiService.ExtractCalendarEvents(&message)
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
