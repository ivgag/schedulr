package model

type MessageType string

const (
	UserMessage      MessageType = "user"
	ForwardedMessage MessageType = "forwarded"
)

type TextMessage struct {
	From        string
	Text        string
	MessageType MessageType
}
