module github.com/ivgag/schedulr/service

go 1.23.1

require github.com/ivgag/schedulr/ai v0.0.0

replace github.com/ivgag/schedulr/ai v0.0.0 => ../ai

require github.com/ivgag/schedulr/storage v0.0.0

replace github.com/ivgag/schedulr/storage v0.0.0 => ../storage

require github.com/sashabaranov/go-openai v1.37.0 // indirect
