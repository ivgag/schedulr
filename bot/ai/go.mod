module github.com/ivgag/schedulr/ai

go 1.23.1

replace github.com/ivgag/schedulr/model => ../model

require (
	github.com/ivgag/schedulr/model v0.0.0
	github.com/sashabaranov/go-openai v1.37.0
)
