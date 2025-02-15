module github.com/ivgag/schedulr/ai

go 1.23.3

toolchain go1.23.6

replace (
	github.com/ivgag/schedulr/model => ../model
	github.com/ivgag/schedulr/utils => ../utils
)

require (
	github.com/ivgag/schedulr/model v0.0.0
	github.com/ivgag/schedulr/utils v0.0.0
	github.com/sashabaranov/go-openai v1.37.0
)

require (
	github.com/cohesion-org/deepseek-go v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
)
