# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy pkg module first (needed for replace directive)
COPY pkg/ ./pkg/
COPY proto/ ./proto/

# Copy service code
COPY payment-service/ ./payment-service/

# Build from the service directory
WORKDIR /app/payment-service
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-service ./*.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /payment-service .

EXPOSE 8082

CMD ["./payment-service"]
