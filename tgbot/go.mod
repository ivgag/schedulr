module github.com/ivgag/schedulr/tgbot

go 1.23.1

replace github.com/ivgag/schedulr/ai => ../ai

require github.com/ivgag/schedulr/storage v0.0.0

replace github.com/ivgag/schedulr/storage => ../storage

require github.com/ivgag/schedulr/service v0.0.0

replace github.com/ivgag/schedulr/service => ../service

require github.com/ivgag/schedulr/google v0.0.0 // indirect

replace github.com/ivgag/schedulr/google => ../google

require github.com/ivgag/schedulr/utils v0.0.0 // indirect

replace github.com/ivgag/schedulr/utils => ../utils

require (
	github.com/go-telegram/bot v1.13.3
	github.com/sashabaranov/go-openai v1.37.0 // indirect
	cloud.google.com/go/auth v0.14.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/ivgag/schedulr/ai v0.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/oauth2 v0.26.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/api v0.220.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250127172529-29210b9bc287 // indirect
	google.golang.org/grpc v1.70.0 // indirect
	google.golang.org/protobuf v1.36.4 // indirect
)
