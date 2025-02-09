module github.com/ivgag/schedulr/tgbot

go 1.23.1

require github.com/ivgag/schedulr/ai v0.0.0 // indirect

replace github.com/ivgag/schedulr/ai => ../ai

require github.com/ivgag/schedulr/storage v0.0.0

replace github.com/ivgag/schedulr/storage => ../storage

require github.com/ivgag/schedulr/service v0.0.0

replace github.com/ivgag/schedulr/service => ../service

require (
	github.com/go-telegram/bot v1.13.3
	github.com/sashabaranov/go-openai v1.37.0 // indirect
)
