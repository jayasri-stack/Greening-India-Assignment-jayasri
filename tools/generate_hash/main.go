package main

import (
	"log/slog"
	"os"

	"github.com/yourname/taskflow-backend/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

// This tool generates bcrypt hashes for seeding test users
// Usage: go run tools/generate_hash/main.go "password123"
func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: go run tools/generate_hash/main.go <password>")
		os.Exit(1)
	}

	password := os.Args[1]
	manager := auth.NewManager("dummy-secret", 0)

	hash, err := manager.HashPassword(password)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		os.Exit(1)
	}

	slog.Info("bcrypt hash generated", "hash", hash)
	slog.Info("update migrations/002_seed_data.up.sql with this hash")

	// Verify
	if manager.VerifyPassword(hash, password) {
		slog.Info("verification successful")
	}
}
