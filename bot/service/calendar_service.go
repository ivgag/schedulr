package service

import "github.com/ivgag/schedulr/model"

type CalendarService interface {
	CreateEvent(userID int, event *model.Event) (*model.Event, error)
}
