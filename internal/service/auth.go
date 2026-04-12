package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yourname/taskflow-backend/internal/auth"
	"github.com/yourname/taskflow-backend/internal/models"
	"github.com/yourname/taskflow-backend/internal/repository"
)

// AuthService handles authentication business logic (Dependency Injection)
type AuthService struct {
	userRepo *repository.UserRepository
	authMgr  *auth.Manager
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, authMgr *auth.Manager) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		authMgr:  authMgr,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*models.AuthResponse, error) {
	// Validate inputs
	if name == "" || email == "" || password == "" {
		return nil, fmt.Errorf("name, email, and password are required")
	}

	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		slog.Error("failed to check existing user", "error", err, "email", email)
		return nil, fmt.Errorf("database error")
	}

	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := s.authMgr.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	// Create user
	user, err := s.userRepo.CreateUser(ctx, name, email, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.authMgr.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	// Validate inputs
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Get user
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		slog.Error("failed to get user", "error", err, "email", email)
		return nil, fmt.Errorf("database error")
	}

	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if !s.authMgr.VerifyPassword(user.Password, password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate token
	token, err := s.authMgr.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}

	// Return user without password
	user.Password = ""
	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}
