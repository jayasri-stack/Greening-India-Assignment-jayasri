package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/auth"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/db"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/handlers"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/middleware"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/models"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/repository"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/service"
)

// TestServer holds test server components
type TestServer struct {
	server          *http.Server
	mux             *http.ServeMux
	database        *db.Database
	authManager     *auth.Manager
	authHandler     *handlers.AuthHandler
	projectHandler  *handlers.ProjectHandler
	taskHandler     *handlers.TaskHandler
}

// Setup creates a test server with in-memory or test database
func setupTestServer(t *testing.T) *TestServer {
	// Use test database or mock
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://taskflow_user:taskflow_password@localhost:5432/taskflow_test"
	}

	database, err := db.New(dsn)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	taskRepo := repository.NewTaskRepository(database)

	// Initialize auth manager
	authMgr := auth.NewManager("test-secret-key", 24*time.Hour)

	// Initialize services
	authService := service.NewAuthService(userRepo, authMgr)
	projectService := service.NewProjectService(projectRepo, taskRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	projectHandler := handlers.NewProjectHandler(projectService)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Setup router
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protected routes
	authMiddleware := middleware.AuthMiddleware(authMgr)

	mux.HandleFunc("GET /projects", authMiddleware(http.HandlerFunc(projectHandler.ListProjects)).ServeHTTP)
	mux.HandleFunc("POST /projects", authMiddleware(http.HandlerFunc(projectHandler.CreateProject)).ServeHTTP)
	mux.HandleFunc("GET /projects/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		projectHandler.GetProject(w, r, projectID)
	})).ServeHTTP)
	mux.HandleFunc("PATCH /projects/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		projectHandler.UpdateProject(w, r, projectID)
	})).ServeHTTP)
	mux.HandleFunc("DELETE /projects/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		projectHandler.DeleteProject(w, r, projectID)
	})).ServeHTTP)

	// Tasks endpoints
	mux.HandleFunc("GET /projects/{id}/tasks", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		taskHandler.ListTasks(w, r, projectID)
	})).ServeHTTP)
	mux.HandleFunc("POST /projects/{id}/tasks", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		taskHandler.CreateTask(w, r, projectID)
	})).ServeHTTP)
	mux.HandleFunc("PATCH /tasks/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("id")
		taskHandler.UpdateTask(w, r, taskID)
	})).ServeHTTP)
	mux.HandleFunc("DELETE /tasks/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID := r.PathValue("id")
		taskHandler.DeleteTask(w, r, taskID)
	})).ServeHTTP)

	// Add middleware stack
	var handler http.Handler = mux
	handler = middleware.LoggingMiddleware()(handler)
	handler = middleware.CORSMiddleware()(handler)

	return &TestServer{
		mux:            mux,
		database:       database,
		authManager:    authMgr,
		authHandler:    authHandler,
		projectHandler: projectHandler,
		taskHandler:    taskHandler,
	}
}

// Cleanup closes the database
func (ts *TestServer) cleanup() {
	if ts.database != nil {
		ts.database.Close()
	}
}

// Helper function to make requests
func makeRequest(t *testing.T, method, path string, body interface{}, token string, mux *http.ServeMux) (*http.Response, []byte) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, "http://localhost:8080"+path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var respBody []byte
	_, err = resp.Body.Read(respBody[:])
	if err != nil && err.Error() != "EOF" {
		respBody, _ = json.Marshal(w.Body.Bytes())
	} else {
		respBody = w.Body.Bytes()
	}

	return resp, respBody
}

// ==================== AUTH TESTS ====================

func TestRegisterSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	req := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	resp, bodyBytes := makeRequest(t, "POST", "/auth/register", req, "", ts.mux)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusCreated, resp.StatusCode, string(bodyBytes))
	}

	var result models.AuthResponse
	err := json.Unmarshal(bodyBytes, &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if result.Token == "" {
		t.Errorf("expected token in response, got empty")
	}
}

func TestRegisterMissingEmail(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	req := models.RegisterRequest{
		Name:     "John Doe",
		Password: "password123",
	}

	resp, _ := makeRequest(t, "POST", "/auth/register", req, "", ts.mux)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	req1 := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	// First registration
	resp1, _ := makeRequest(t, "POST", "/auth/register", req1, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Errorf("first registration failed with status %d", resp1.StatusCode)
	}

	// Duplicate registration
	resp2, _ := makeRequest(t, "POST", "/auth/register", req1, "", ts.mux)
	if resp2.StatusCode != http.StatusBadRequest {
		t.Errorf("duplicate registration should fail with status %d, got %d", http.StatusBadRequest, resp2.StatusCode)
	}
}

func TestLoginSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register first
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)

	// Login
	loginReq := models.LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	}

	resp, bodyBytes := makeRequest(t, "POST", "/auth/login", loginReq, "", ts.mux)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}

	var result models.AuthResponse
	err := json.Unmarshal(bodyBytes, &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if result.Token == "" {
		t.Errorf("expected token in response")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	loginReq := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	resp, _ := makeRequest(t, "POST", "/auth/login", loginReq, "", ts.mux)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

// ==================== PROJECT TESTS ====================

func TestListProjectsUnauthorized(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	resp, _ := makeRequest(t, "GET", "/projects", nil, "", ts.mux)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestListProjectsEmpty(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// List projects
	resp, bodyBytes := makeRequest(t, "GET", "/projects", nil, token, ts.mux)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}

	var result models.ProjectsResponse
	err := json.Unmarshal(bodyBytes, &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if result.Projects == nil {
		t.Errorf("projects should not be nil")
	}
}

func TestCreateProjectSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name:        "My Project",
		Description: nil,
	}

	resp, bodyBytes := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusCreated, resp.StatusCode, string(bodyBytes))
	}

	var project models.Project
	err := json.Unmarshal(bodyBytes, &project)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if project.Name != "My Project" {
		t.Errorf("expected project name 'My Project', got '%s'", project.Name)
	}
}

func TestCreateProjectMissingName(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project without name
	projectReq := models.CreateProjectRequest{
		Description: nil,
	}

	resp, _ := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGetProjectSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Get project
	resp, bodyBytes := makeRequest(t, "GET", "/projects/"+project.ID, nil, token, ts.mux)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}
}

func TestGetProjectNotFound(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Get non-existent project
	resp, _ := makeRequest(t, "GET", "/projects/nonexistent", nil, token, ts.mux)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestUpdateProjectSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Update project
	newName := "Updated Project"
	updateReq := models.UpdateProjectRequest{
		Name: &newName,
	}

	resp, bodyBytes := makeRequest(t, "PATCH", "/projects/"+project.ID, updateReq, token, ts.mux)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}

	var updatedProject models.Project
	json.Unmarshal(bodyBytes, &updatedProject)

	if updatedProject.Name != "Updated Project" {
		t.Errorf("expected project name 'Updated Project', got '%s'", updatedProject.Name)
	}
}

func TestDeleteProjectSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Delete project
	resp, _ := makeRequest(t, "DELETE", "/projects/"+project.ID, nil, token, ts.mux)

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

// ==================== TASK TESTS ====================

func TestCreateTaskSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Create task
	taskReq := models.CreateTaskRequest{
		Title:    "Task 1",
		Priority: "high",
	}

	resp, bodyBytes := makeRequest(t, "POST", "/projects/"+project.ID+"/tasks", taskReq, token, ts.mux)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusCreated, resp.StatusCode, string(bodyBytes))
	}

	var task models.Task
	err := json.Unmarshal(bodyBytes, &task)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if task.Title != "Task 1" {
		t.Errorf("expected task title 'Task 1', got '%s'", task.Title)
	}
}

func TestCreateTaskMissingTitle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Create task without title
	taskReq := models.CreateTaskRequest{
		Priority: "high",
	}

	resp, _ := makeRequest(t, "POST", "/projects/"+project.ID+"/tasks", taskReq, token, ts.mux)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestListTasksSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// List tasks
	resp, bodyBytes := makeRequest(t, "GET", "/projects/"+project.ID+"/tasks", nil, token, ts.mux)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}

	var result models.TasksResponse
	err := json.Unmarshal(bodyBytes, &result)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if result.Tasks == nil {
		t.Errorf("tasks should not be nil")
	}
}

func TestUpdateTaskSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Create task
	taskReq := models.CreateTaskRequest{
		Title:    "Task 1",
		Priority: "high",
	}
	resp3, bodyBytes3 := makeRequest(t, "POST", "/projects/"+project.ID+"/tasks", taskReq, token, ts.mux)
	if resp3.StatusCode != http.StatusCreated {
		t.Fatalf("task creation failed")
	}

	var task models.Task
	json.Unmarshal(bodyBytes3, &task)

	// Update task
	newStatus := "in_progress"
	updateReq := models.UpdateTaskRequest{
		Status: &newStatus,
	}

	resp, bodyBytes := makeRequest(t, "PATCH", "/tasks/"+task.ID, updateReq, token, ts.mux)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}

	var updatedTask models.Task
	json.Unmarshal(bodyBytes, &updatedTask)

	if updatedTask.Status != "in_progress" {
		t.Errorf("expected task status 'in_progress', got '%s'", updatedTask.Status)
	}
}

func TestDeleteTaskSuccess(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Create task
	taskReq := models.CreateTaskRequest{
		Title:    "Task 1",
		Priority: "high",
	}
	resp3, bodyBytes3 := makeRequest(t, "POST", "/projects/"+project.ID+"/tasks", taskReq, token, ts.mux)
	if resp3.StatusCode != http.StatusCreated {
		t.Fatalf("task creation failed")
	}

	var task models.Task
	json.Unmarshal(bodyBytes3, &task)

	// Delete task
	resp, _ := makeRequest(t, "DELETE", "/tasks/"+task.ID, nil, token, ts.mux)

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestListTasksWithFilters(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	// Register and login
	registerReq := models.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	resp1, bodyBytes1 := makeRequest(t, "POST", "/auth/register", registerReq, "", ts.mux)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("registration failed")
	}

	var authResp models.AuthResponse
	json.Unmarshal(bodyBytes1, &authResp)
	token := authResp.Token

	// Create project
	projectReq := models.CreateProjectRequest{
		Name: "My Project",
	}
	resp2, bodyBytes2 := makeRequest(t, "POST", "/projects", projectReq, token, ts.mux)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("project creation failed")
	}

	var project models.Project
	json.Unmarshal(bodyBytes2, &project)

	// Create tasks with different statuses
	taskReq1 := models.CreateTaskRequest{
		Title:    "Task 1",
		Priority: "high",
	}
	makeRequest(t, "POST", "/projects/"+project.ID+"/tasks", taskReq1, token, ts.mux)

	// List tasks with status filter
	resp, bodyBytes := makeRequest(t, "GET", "/projects/"+project.ID+"/tasks?status=todo", nil, token, ts.mux)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, string(bodyBytes))
	}
}
