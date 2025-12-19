package domain

import "context"

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user and returns the created user with ID
	Create(ctx context.Context, user *User) (*User, error)
	// FindByUsername finds a user by username
	FindByUsername(ctx context.Context, username string) (*User, error)
	// FindByID finds a user by ID
	FindByID(ctx context.Context, id int) (*User, error)
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}
