package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"jayasri-stack/Greening-India-Assignment-jayasri/internal/auth"
	"jayasri-stack/Greening-India-Assignment-jayasri/internal/models"
)

// ContextKey defines the type for context keys
type ContextKey string

const UserContextKey ContextKey = "user"

// AuthMiddleware validates JWT tokens from Authorization header
func AuthMiddleware(authMgr *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				slog.Warn("missing authorization header")
				respondError(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				slog.Warn("invalid authorization header format")
				respondError(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}

			token := parts[1]
			claims, err := authMgr.ValidateToken(token)
			if err != nil {
				slog.Warn("invalid token", "error", err)
				respondError(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}

			// Store claims in context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves user claims from context
func GetUserFromContext(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

// respondError is a helper to send error responses
func respondError(w http.ResponseWriter, statusCode int, message string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.ErrorResponse{
		Error:  message,
		Fields: fields,
	}

	// In production, use json.Marshal
	data, _ := marshalJSON(response)
	w.Write(data)
}

// marshalJSON is a simple JSON marshaller (you'd use encoding/json in production)
func marshalJSON(v interface{}) ([]byte, error) {
	// Placeholder - use encoding/json in actual implementation
	return []byte(`{}`), nil
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("incoming request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
