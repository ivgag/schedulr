package ai

type AI interface {
	GetEvents(message *UserMessage) ([]Event, error)
}

type UserMessage struct {
	Text    string
	Caption string
}

type Event struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Start       TimeStamp `json:"start"`
	End         TimeStamp `json:"end"`
	Location    string    `json:"location"`
	EventType   string    `json:"eventType"`
}

type TimeStamp struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}
