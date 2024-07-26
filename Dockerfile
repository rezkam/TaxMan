# Build the Go application
FROM golang:1.22.5 AS builder

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code and build the application
COPY . .
RUN go build -o server ./cmd/server/main.go

# Test the application and using the PostgreSQL database for testing
FROM golang:1.22.5 AS tester

WORKDIR /app

# Copy only the necessary files for testing from the builder stage
COPY --from=builder /app .
COPY go.mod go.sum ./
COPY . .

CMD ["go", "test", "-p", "1", "./...", "-v"]

# Production image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built application from the builder stage
COPY --from=builder /app/server .

CMD ["./server"]
