# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY pkg/ ./pkg/
COPY gateway/ ./gateway/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./gateway

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /gateway .

EXPOSE 8080

CMD ["./gateway"]
