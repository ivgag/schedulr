module github.com/ivgag/schedulr/service

go 1.23.1

require github.com/ivgag/schedulr/ai v0.0.0

replace github.com/ivgag/schedulr/ai v0.0.0 => ../ai

require github.com/ivgag/schedulr/storage v0.0.0

replace github.com/ivgag/schedulr/storage v0.0.0 => ../storage

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/sashabaranov/go-openai v1.37.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)
