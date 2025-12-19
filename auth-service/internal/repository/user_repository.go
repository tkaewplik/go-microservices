package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tkaewplik/go-microservices/auth-service/internal/domain"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create creates a new user in the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id"

	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// FindByUsername finds a user by username
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := "SELECT id, username, password FROM users WHERE username = $1"

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	return user, nil
}

// FindByID finds a user by ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	query := "SELECT id, username, password FROM users WHERE id = $1"

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return user, nil
}
