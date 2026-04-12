package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/yourname/taskflow-backend/internal/db"
	"github.com/yourname/taskflow-backend/internal/models"
)

// UserRepository handles user data operations (Repository Pattern - SOLID Principle)
type UserRepository struct {
	db *db.Database
	mu sync.RWMutex
}

// NewUserRepository creates a new user repository
func NewUserRepository(database *db.Database) *UserRepository {
	return &UserRepository{
		db: database,
	}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, name, email, hashedPassword string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userID := uuid.New().String()

	query := `
		INSERT INTO users (id, name, email, password, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, name, email, created_at
	`

	user := &models.User{}
	err := r.db.QueryRow(query, userID, name, email, hashedPassword).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt,
	)

	if err != nil {
		slog.Error("failed to create user", "error", err, "email", email)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user creation failed")
		}
		return nil, fmt.Errorf("user creation error: %w", err)
	}

	slog.Info("user created", "user_id", user.ID, "email", email)
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		slog.Error("failed to get user by email", "error", err, "email", email)
		return nil, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("failed to get user by ID", "error", err, "user_id", userID)
		return nil, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}
