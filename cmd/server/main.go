package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/auth"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/db"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/handlers"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/middleware"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/repository"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/service"
)

type routeParamsKey struct{}

type routeEntry struct {
	method     string
	pattern    *regexp.Regexp
	paramNames []string
	handler    http.Handler
}

type Router struct {
	routes []routeEntry
}

func NewRouter() *Router {
	return &Router{}
}

func compileRoutePattern(pattern string) (*regexp.Regexp, []string) {
	paramNames := []string{}
	re := regexp.MustCompile(`\{([^/]+?)\}`)
	regexPattern := re.ReplaceAllStringFunc(pattern, func(m string) string {
		name := m[1 : len(m)-1]
		paramNames = append(paramNames, name)
		return "([^/]+)"
	})
	return regexp.MustCompile("^" + regexPattern + "$"), paramNames
}

func (r *Router) Handle(method, pattern string, handler http.Handler) {
	regex, paramNames := compileRoutePattern(pattern)
	r.routes = append(r.routes, routeEntry{method: method, pattern: regex, paramNames: paramNames, handler: handler})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if route.method != req.Method {
			continue
		}

		matches := route.pattern.FindStringSubmatch(req.URL.Path)
		if matches == nil {
			continue
		}

		params := map[string]string{}
		for i, name := range route.paramNames {
			params[name] = matches[i+1]
		}

		ctx := context.WithValue(req.Context(), routeParamsKey{}, params)
		route.handler.ServeHTTP(w, req.WithContext(ctx))
		return
	}

	http.NotFound(w, req)
}

func pathValue(r *http.Request, name string) string {
	params, ok := r.Context().Value(routeParamsKey{}).(map[string]string)
	if !ok {
		return ""
	}
	return params[name]
}

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL environment variable not set")
		os.Exit(1)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET environment variable not set")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	slog.Info("connecting to database", "dsn", maskDSN(dsn))
	database, err := db.New(dsn)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	// defer database.Close()

	 // Run migrations automatically on startup
	// if err := runMigrations(database); err != nil {
	 // 	slog.Error("failed to run migrations", "error", err)
	 // 	os.Exit(1)
	// }

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	taskRepo := repository.NewTaskRepository(database)

	// Initialize auth manager
	authMgr := auth.NewManager(jwtSecret, 24*time.Hour)

	// Initialize services
	authService := service.NewAuthService(userRepo, authMgr)
	projectService := service.NewProjectService(projectRepo, taskRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	projectHandler := handlers.NewProjectHandler(projectService)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Setup router
	router := NewRouter()

	// Auth endpoints (no authentication required)
	router.Handle("POST", "/auth/register", http.HandlerFunc(authHandler.Register))
	router.Handle("POST", "/auth/login", http.HandlerFunc(authHandler.Login))

	// Protected routes with auth middleware
	authMiddleware := middleware.AuthMiddleware(authMgr)

	// Projects endpoints
	router.Handle("GET", "/projects", authMiddleware(http.HandlerFunc(projectHandler.ListProjects)))
	router.Handle("POST", "/projects", authMiddleware(http.HandlerFunc(projectHandler.CreateProject)))
	router.Handle("GET", "/projects/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := pathValue(r, "id")
		projectHandler.GetProject(w, r, projectID)
	})))
	router.Handle("PATCH", "/projects/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := pathValue(r, "id")
		projectHandler.UpdateProject(w, r, projectID)
	})))
	router.Handle("DELETE", "/projects/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := pathValue(r, "id")
		projectHandler.DeleteProject(w, r, projectID)
	})))

	// Tasks endpoints
	router.Handle("GET", "/projects/{id}/tasks", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := pathValue(r, "id")
		taskHandler.ListTasks(w, r, projectID)
	})))
	router.Handle("POST", "/projects/{id}/tasks", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := pathValue(r, "id")
		taskHandler.CreateTask(w, r, projectID)
	})))
	router.Handle("PATCH", "/tasks/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID := pathValue(r, "id")
		taskHandler.UpdateTask(w, r, taskID)
	})))
	router.Handle("DELETE", "/tasks/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID := pathValue(r, "id")
		taskHandler.DeleteTask(w, r, taskID)
	})))

	// Add middleware stack
	var handler http.Handler = router
	handler = middleware.LoggingMiddleware()(handler)
	handler = middleware.CORSMiddleware()(handler)

	// Create HTTP server with graceful shutdown support
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	done := make(chan bool, 1)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan

		slog.Info("shutdown signal received, gracefully shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}

		done <- true
	}()

	// Start server
	slog.Info("starting server", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}

	<-done
	slog.Info("server stopped")
}


// maskDSN masks the password in the DSN for logging
func maskDSN(dsn string) string {
	// Find password pattern and mask it
	parts := strings.Split(dsn, "@")
	if len(parts) == 2 {
		userPart := strings.Split(parts[0], "://")[1]
		if strings.Contains(userPart, ":") {
			userParts := strings.Split(userPart, ":")
			return fmt.Sprintf("postgresql://%s:***@%s", userParts[0], parts[1])
		}
	}
	return dsn
}
