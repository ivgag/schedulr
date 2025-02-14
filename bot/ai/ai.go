package ai

import "github.com/ivgag/schedulr/model"

type AI interface {
	GetEvents(message *model.TextMessage) ([]model.Event, error)
}
