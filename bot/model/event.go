package model

type Event struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Start       TimeStamp `json:"start"`
	End         TimeStamp `json:"end"`
	Location    string    `json:"location"`
	EventType   string    `json:"eventType"`
	Link        string    `json:"link"`
}

type TimeStamp struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}
