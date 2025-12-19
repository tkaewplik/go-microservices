package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/tkaewplik/go-microservices/payment-service/internal/handler"
	"github.com/tkaewplik/go-microservices/payment-service/internal/kafka"
	"github.com/tkaewplik/go-microservices/payment-service/internal/repository"
	"github.com/tkaewplik/go-microservices/payment-service/internal/service"
	"github.com/tkaewplik/go-microservices/pkg/database"
	"github.com/tkaewplik/go-microservices/pkg/middleware"
)

func main() {
	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Database configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "paymentdb"),
	}

	// Connect to database
	db, err := database.Connect(dbConfig)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}()

	// Initialize Kafka publisher
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "transactions")

	kafkaCfg := kafka.Config{
		Brokers: strings.Split(kafkaBrokers, ","),
		Topic:   kafkaTopic,
	}

	publisher := kafka.NewPublisher(kafkaCfg, logger)
	defer func() {
		if err := publisher.Close(); err != nil {
			logger.Error("failed to close Kafka publisher", "error", err)
		}
	}()

	// Initialize layers
	txRepo := repository.NewPostgresTransactionRepository(db)
	paymentService := service.NewPaymentService(txRepo, publisher)
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	// Setup middleware
	secretKey := getEnv("JWT_SECRET", "your-secret-key")
	authMiddleware := middleware.NewAuthMiddleware(secretKey)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/transactions", authMiddleware.Authenticate(paymentHandler.CreateTransaction))
	mux.HandleFunc("/transactions/list", authMiddleware.Authenticate(paymentHandler.GetTransactions))
	mux.HandleFunc("/transactions/pay", authMiddleware.Authenticate(paymentHandler.PayAllTransactions))

	// Start server
	port := getEnv("PORT", "8082")
	logger.Info("payment service starting",
		"port", port,
		"kafka_brokers", kafkaBrokers,
		"kafka_topic", kafkaTopic,
	)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
