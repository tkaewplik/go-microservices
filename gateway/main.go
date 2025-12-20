package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tkaewplik/go-microservices/pkg/middleware"
	authpb "github.com/tkaewplik/go-microservices/proto/auth"
	paymentpb "github.com/tkaewplik/go-microservices/proto/payment"
)

type Gateway struct {
	authClient    authpb.AuthServiceClient
	paymentClient paymentpb.PaymentServiceClient
	logger        *slog.Logger
}

func NewGateway(authGRPCAddr, paymentGRPCAddr string, logger *slog.Logger) (*Gateway, error) {
	// Connect to auth service gRPC
	authConn, err := grpc.NewClient(authGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	// Connect to payment service gRPC
	paymentConn, err := grpc.NewClient(paymentGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		authClient:    authpb.NewAuthServiceClient(authConn),
		paymentClient: paymentpb.NewPaymentServiceClient(paymentConn),
		logger:        logger,
	}, nil
}

// Auth handlers
func (g *Gateway) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		g.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := g.authClient.Register(ctx, &authpb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		g.logger.Error("register failed", "error", err)
		g.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	g.respondJSON(w, http.StatusCreated, resp)
}

func (g *Gateway) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		g.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := g.authClient.Login(ctx, &authpb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		g.logger.Error("login failed", "error", err)
		g.respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	g.respondJSON(w, http.StatusOK, resp)
}

// Payment handlers with auth validation
func (g *Gateway) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		g.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Validate token
	userID, err := g.validateAuth(r)
	if err != nil {
		g.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := g.paymentClient.CreateTransaction(ctx, &paymentpb.CreateTransactionRequest{
		UserId:      int32(userID),
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		g.logger.Error("create transaction failed", "error", err)
		g.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	g.respondJSON(w, http.StatusCreated, resp)
}

func (g *Gateway) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		g.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := g.validateAuth(r)
	if err != nil {
		g.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := g.paymentClient.GetTransactions(ctx, &paymentpb.GetTransactionsRequest{
		UserId: int32(userID),
	})
	if err != nil {
		g.logger.Error("get transactions failed", "error", err)
		g.respondError(w, http.StatusInternalServerError, "failed to get transactions")
		return
	}

	g.respondJSON(w, http.StatusOK, resp)
}

func (g *Gateway) handlePayTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		g.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, err := g.validateAuth(r)
	if err != nil {
		g.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := g.paymentClient.PayAllTransactions(ctx, &paymentpb.PayRequest{
		UserId: int32(userID),
	})
	if err != nil {
		g.logger.Error("pay transactions failed", "error", err)
		g.respondError(w, http.StatusInternalServerError, "failed to pay transactions")
		return
	}

	g.respondJSON(w, http.StatusOK, resp)
}

// validateAuth validates the JWT token via gRPC call to auth service
func (g *Gateway) validateAuth(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, ErrUnauthorized
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, ErrUnauthorized
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp, err := g.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		Token: parts[1],
	})
	if err != nil || !resp.Valid {
		return 0, ErrUnauthorized
	}

	return int(resp.UserId), nil
}

var ErrUnauthorized = &Error{Message: "unauthorized"}

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func (g *Gateway) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		g.logger.Error("failed to encode response", "error", err)
	}
}

func (g *Gateway) respondError(w http.ResponseWriter, status int, message string) {
	g.respondJSON(w, status, map[string]string{"error": message})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	authGRPCAddr := getEnv("AUTH_GRPC_ADDR", "localhost:50051")
	paymentGRPCAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:50052")

	gateway, err := NewGateway(authGRPCAddr, paymentGRPCAddr, logger)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/auth/register", gateway.handleRegister)
	mux.HandleFunc("/auth/login", gateway.handleLogin)

	// Payment routes
	mux.HandleFunc("/payment/transactions", gateway.handleCreateTransaction)
	mux.HandleFunc("/payment/transactions/list", gateway.handleGetTransactions)
	mux.HandleFunc("/payment/transactions/pay", gateway.handlePayTransactions)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	handler := middleware.CORS(mux)

	port := getEnv("PORT", "8080")
	logger.Info("API Gateway starting",
		"port", port,
		"auth_grpc", authGRPCAddr,
		"payment_grpc", paymentGRPCAddr,
	)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
