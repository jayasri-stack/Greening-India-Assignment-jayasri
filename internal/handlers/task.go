package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/middleware"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/models"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/service"
)

// TaskHandler handles task endpoints
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// ListTasks handles GET /projects/:id/tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodGet {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	status := r.URL.Query().Get("status")
	assigneeID := r.URL.Query().Get("assignee")

	tasks, err := h.taskService.ListTasksForProject(r.Context(), projectID, status, assigneeID)
	if err != nil {
		if err.Error() == "project not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "invalid status" {
			middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", map[string]string{"status": "must be todo, in_progress, or done"})
			return
		}
		slog.Error("failed to list tasks", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to list tasks", nil)
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	middleware.RespondJSON(w, http.StatusOK, models.TasksResponse{Tasks: tasks})
}

// CreateTask handles POST /projects/:id/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPost {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req models.CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		slog.Warn("failed to decode create task request", "error", err)
		fields := map[string]string{"request": "invalid JSON"}
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	if req.Title == "" {
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", map[string]string{"title": "is required"})
		return
	}

	task, err := h.taskService.CreateTask(r.Context(), userClaims.UserID, projectID, req.Title, &req)
	if err != nil {
		if err.Error() == "project not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "invalid priority" {
			middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", map[string]string{"priority": "must be low, medium, or high"})
			return
		}
		slog.Error("failed to create task", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to create task", nil)
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, task)
}

// UpdateTask handles PATCH /tasks/:id
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request, taskID string) {
	if r.Method != http.MethodPatch {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req models.UpdateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		slog.Warn("failed to decode update task request", "error", err)
		fields := map[string]string{"request": "invalid JSON"}
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	task, err := h.taskService.UpdateTask(r.Context(), taskID, userClaims.UserID, &req)
	if err != nil {
		if err.Error() == "task not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "project not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "only project owner can update tasks" {
			middleware.RespondErrorJSON(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		if err.Error() == "invalid status" || err.Error() == "invalid priority" {
			middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", map[string]string{err.Error(): "invalid"})
			return
		}
		slog.Error("failed to update task", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to update task", nil)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, task)
}

// DeleteTask handles DELETE /tasks/:id
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request, taskID string) {
	if r.Method != http.MethodDelete {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	err := h.taskService.DeleteTask(r.Context(), taskID, userClaims.UserID)
	if err != nil {
		if err.Error() == "task not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "project not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "only project owner can delete tasks" {
			middleware.RespondErrorJSON(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		slog.Error("failed to delete task", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to delete task", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
