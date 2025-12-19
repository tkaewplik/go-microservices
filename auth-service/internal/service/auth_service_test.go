package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tkaewplik/go-microservices/auth-service/internal/domain"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	users     map[string]*domain.User
	createErr error
	findErr   error
	nextID    int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[string]*domain.User),
		nextID: 1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.Username] = user
	return user, nil
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.users[username], nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewAuthService(repo, "test-secret")

	resp, err := svc.Register(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.ID != 1 {
		t.Errorf("expected user ID 1, got %d", resp.ID)
	}
	if resp.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", resp.Username)
	}
	if resp.Token == "" {
		t.Error("expected token to be generated")
	}
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewAuthService(repo, "test-secret")

	// First registration
	_, err := svc.Register(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("first registration should succeed: %v", err)
	}

	// Second registration with same username
	_, err = svc.Register(context.Background(), "testuser", "password456")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewAuthService(repo, "test-secret")

	// Register first
	_, err := svc.Register(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("registration should succeed: %v", err)
	}

	// Login
	resp, err := svc.Login(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", resp.Username)
	}
	if resp.Token == "" {
		t.Error("expected token to be generated")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewAuthService(repo, "test-secret")

	// Register first
	_, err := svc.Register(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("registration should succeed: %v", err)
	}

	// Login with wrong password
	_, err = svc.Login(context.Background(), "testuser", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewAuthService(repo, "test-secret")

	// Login without registering
	_, err := svc.Login(context.Background(), "nonexistent", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
