MIGRATE_CMD=migrate

AUTH_DB_URL=postgres://postgres:postgres@localhost:5433/authdb?sslmode=disable
PAYMENT_DB_URL=postgres://postgres:postgres@localhost:5434/paymentdb?sslmode=disable

.PHONY: migrate-auth-up migrate-auth-down migrate-payment-up migrate-payment-down migrate-up migrate-down

migrate-auth-up:
	$(MIGRATE_CMD) -path migrations/auth -database "$(AUTH_DB_URL)" up

migrate-auth-down:
	$(MIGRATE_CMD) -path migrations/auth -database "$(AUTH_DB_URL)" down

migrate-payment-up:
	$(MIGRATE_CMD) -path migrations/payment -database "$(PAYMENT_DB_URL)" up

migrate-payment-down:
	$(MIGRATE_CMD) -path migrations/payment -database "$(PAYMENT_DB_URL)" down

migrate-up: migrate-auth-up migrate-payment-up

migrate-down: migrate-auth-down migrate-payment-down

.PHONY: lint lint-auth lint-payment lint-gateway lint-pkg

lint: lint-auth lint-payment lint-gateway lint-pkg
	@echo "All lint checks passed"

lint-auth:
	cd auth-service && golangci-lint run

lint-payment:
	cd payment-service && golangci-lint run

lint-gateway:
	cd gateway && golangci-lint run

lint-pkg:
	cd pkg && golangci-lint run

test:
	cd auth-service && go test ./... -v
	cd payment-service && go test ./... -v
	cd gateway && go test ./... -v
	cd analytics-service && go test ./... -v

# Proto generation
.PHONY: proto-gen proto-gen-auth proto-gen-payment proto-install

# Install protoc plugins (run once)
proto-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate all proto files
proto-gen: proto-gen-auth proto-gen-payment
	@echo "Proto generation complete"

# Generate auth service proto
proto-gen-auth:
	PATH=$$PATH:$$(go env GOPATH)/bin protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/auth/auth.proto

# Generate payment service proto
proto-gen-payment:
	PATH=$$PATH:$$(go env GOPATH)/bin protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/payment/payment.proto