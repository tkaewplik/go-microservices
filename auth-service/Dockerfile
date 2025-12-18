# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY pkg/ ./pkg/
COPY auth-service/ ./auth-service/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./auth-service

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /auth-service .

EXPOSE 8081

CMD ["./auth-service"]
