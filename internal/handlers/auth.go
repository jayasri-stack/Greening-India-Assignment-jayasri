package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"jayasri-stack/Greening-India-Assignment-jayasri/internal/middleware"
	"jayasri-stack/Greening-India-Assignment-jayasri/internal/models"
	"jayasri-stack/Greening-India-Assignment-jayasri/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	var req models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		slog.Warn("failed to decode register request", "error", err)
		fields := map[string]string{"request": "invalid JSON"}
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	// Validate
	fields := make(map[string]string)
	if req.Name == "" {
		fields["name"] = "is required"
	}
	if req.Email == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	// Register
	result, err := h.authService.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		slog.Warn("registration failed", "error", err, "email", req.Email)
		fields := map[string]string{"email": err.Error()}
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, result)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		slog.Warn("failed to decode login request", "error", err)
		fields := map[string]string{"request": "invalid JSON"}
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	// Validate
	fields := make(map[string]string)
	if req.Email == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	// Login
	result, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		slog.Warn("login failed", "error", err, "email", req.Email)
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, result)
}
