module github.com/tkaewplik/go-microservices/gateway

go 1.25.5

replace github.com/tkaewplik/go-microservices/pkg => ../pkg

require github.com/tkaewplik/go-microservices/pkg v0.0.0-00010101000000-000000000000

require github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
