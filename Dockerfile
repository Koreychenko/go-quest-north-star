# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main/main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests (needed for Telegram API and Gemini AI)
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .

# Copy required resources
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/prompts ./prompts
COPY --from=builder /app/photos ./photos

# Command to run the application
CMD ["./main"]
