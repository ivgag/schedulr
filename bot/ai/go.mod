module github.com/ivgag/schedulr/ai

go 1.23.1

require github.com/ivgag/schedulr/domain v0.0.0
replace github.com/ivgag/schedulr/domain => ../domain

require (
	github.com/sashabaranov/go-openai v1.37.0
)
