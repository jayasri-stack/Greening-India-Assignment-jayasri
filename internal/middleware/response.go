package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"jayasri-stack/Greening-India-Assignment-jayasri/internal/models"
)

// ErrorHandler wraps handlers to catch and handle errors gracefully
func ErrorHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(models.ErrorResponse{
					Error: "internal server error",
				})
			}
		}()

		handler(w, r)
	}
}

// RespondJSON writes JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// RespondError writes error response
func RespondErrorJSON(w http.ResponseWriter, statusCode int, message string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:  message,
		Fields: fields,
	})
}
