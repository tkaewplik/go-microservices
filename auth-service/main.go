package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"google.golang.org/grpc"

	authgrpc "github.com/tkaewplik/go-microservices/auth-service/internal/grpc"
	"github.com/tkaewplik/go-microservices/auth-service/internal/handler"
	"github.com/tkaewplik/go-microservices/auth-service/internal/repository"
	"github.com/tkaewplik/go-microservices/auth-service/internal/service"
	"github.com/tkaewplik/go-microservices/pkg/database"
	pb "github.com/tkaewplik/go-microservices/proto/auth"
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
		DBName:   getEnv("DB_NAME", "authdb"),
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

	// Initialize layers
	userRepo := repository.NewPostgresUserRepository(db)
	secretKey := getEnv("JWT_SECRET", "your-secret-key")
	authService := service.NewAuthService(userRepo, secretKey)

	// Start gRPC server
	grpcPort := getEnv("GRPC_PORT", "50051")
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Error("failed to listen for gRPC", "error", err, "port", grpcPort)
			os.Exit(1)
		}

		grpcServer := grpc.NewServer()
		authGRPCServer := authgrpc.NewAuthServer(authService, secretKey)
		pb.RegisterAuthServiceServer(grpcServer, authGRPCServer)

		logger.Info("gRPC server starting", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
			os.Exit(1)
		}
	}()

	// HTTP server (for backwards compatibility and health checks)
	authHandler := handler.NewAuthHandler(authService, logger)
	mux := http.NewServeMux()
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Start HTTP server
	port := getEnv("PORT", "8081")
	logger.Info("HTTP server starting", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("HTTP server failed", "error", err)
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
