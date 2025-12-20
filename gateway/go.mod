module github.com/tkaewplik/go-microservices/gateway

go 1.25.5

replace github.com/tkaewplik/go-microservices/pkg => ../pkg

replace github.com/tkaewplik/go-microservices/proto => ../proto

require (
	github.com/tkaewplik/go-microservices/pkg v0.0.0-00010101000000-000000000000
	github.com/tkaewplik/go-microservices/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.77.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
