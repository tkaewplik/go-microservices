package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tkaewplik/go-microservices/auth-service/internal/domain"
	"github.com/tkaewplik/go-microservices/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrHashingPassword    = errors.New("failed to hash password")
	ErrGeneratingToken    = errors.New("failed to generate token")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  domain.UserRepository
	secretKey string
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo domain.UserRepository, secretKey string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		secretKey: secretKey,
	}
}

// Register creates a new user and returns authentication response
func (s *AuthService) Register(ctx context.Context, username, password string) (*domain.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHashingPassword, err)
	}

	// Create user
	user := &domain.User{
		Username: username,
		Password: string(hashedPassword),
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, err := jwt.GenerateToken(createdUser.ID, createdUser.Username, s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGeneratingToken, err)
	}

	return &domain.AuthResponse{
		ID:       createdUser.ID,
		Username: createdUser.Username,
		Token:    token,
	}, nil
}

// Login authenticates a user and returns authentication response
func (s *AuthService) Login(ctx context.Context, username, password string) (*domain.AuthResponse, error) {
	// Find user
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := jwt.GenerateToken(user.ID, user.Username, s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGeneratingToken, err)
	}

	return &domain.AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	}, nil
}
