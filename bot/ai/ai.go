package ai

import "github.com/ivgag/schedulr/domain"

type AI interface {
	GetEvents(message *domain.UserMessage) ([]domain.Event, error)
}
