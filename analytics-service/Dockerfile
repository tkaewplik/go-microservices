# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy pkg module first (needed for replace directive)
COPY pkg/ ./pkg/

# Copy service code
COPY analytics-service/ ./analytics-service/

# Build from the service directory
WORKDIR /app/analytics-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics-service ./*.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /analytics-service .

EXPOSE 8083

CMD ["./analytics-service"]
