package ai

type AI interface {
	GetEvents(message *UserMessage) ([]Event, error)
}

type UserMessage struct {
	Text    string
	Caption string
}

type Event struct {
	Title       string
	Description string
}
