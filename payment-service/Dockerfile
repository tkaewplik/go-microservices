# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY pkg/ ./pkg/
COPY payment-service/ ./payment-service/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-service ./payment-service

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /payment-service .

EXPOSE 8082

CMD ["./payment-service"]
