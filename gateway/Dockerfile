# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy pkg module first (needed for replace directive)
COPY pkg/ ./pkg/
COPY proto/ ./proto/

# Copy service code
COPY gateway/ ./gateway/

# Build from the service directory
WORKDIR /app/gateway
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./*.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /gateway .

EXPOSE 8080

CMD ["./gateway"]
