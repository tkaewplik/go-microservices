module github.com/tkaewplik/go-microservices/payment-service

go 1.25.5

replace github.com/tkaewplik/go-microservices/pkg => ../pkg

require (
	github.com/segmentio/kafka-go v0.4.49
	github.com/tkaewplik/go-microservices/pkg v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)
