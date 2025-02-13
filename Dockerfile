# Stage 1: Build the application
FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY ./bot .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

# Stage 2: Build the final image
FROM alpine:3.21.2
WORKDIR /app

# Install ca-certificates and wget (for downloading migrate CLI)
RUN apk add --no-cache ca-certificates wget tar

# Download and install migrate CLI
RUN wget -qO- https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz | tar xvz -C /usr/local/bin

# Copy the built application
COPY --from=builder /app/app .

# Copy the database migrations
COPY ./db/migrations /migrations

# Copy the entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]