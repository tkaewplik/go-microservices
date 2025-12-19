package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tkaewplik/go-microservices/auth-service/internal/service"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// RegisterRequest represents the request body for registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode register request", "error", err)
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		h.respondError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	response, err := h.authService.Register(ctx, req.Username, req.Password)
	if err != nil {
		h.logger.Error("failed to register user", "error", err, "username", req.Username)

		if errors.Is(err, service.ErrUserAlreadyExists) {
			h.respondError(w, http.StatusConflict, "user already exists")
			return
		}

		h.respondError(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	h.logger.Info("user registered successfully", "user_id", response.ID, "username", response.Username)
	h.respondJSON(w, http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode login request", "error", err)
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		h.respondError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	response, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		h.logger.Warn("login failed", "error", err, "username", req.Username)

		if errors.Is(err, service.ErrInvalidCredentials) {
			h.respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		h.respondError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	h.logger.Info("user logged in successfully", "user_id", response.ID, "username", response.Username)
	h.respondJSON(w, http.StatusOK, response)
}

// respondJSON writes a JSON response
func (h *AuthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// respondError writes an error response
func (h *AuthHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{Error: message})
}
