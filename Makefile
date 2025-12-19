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
