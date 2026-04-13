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

// ProjectHandler handles project endpoints
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// ListProjects handles GET /projects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	projects, err := h.projectService.ListProjectsForUser(r.Context(), userClaims.UserID)
	if err != nil {
		slog.Error("failed to list projects", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to list projects", nil)
		return
	}

	if projects == nil {
		projects = []models.Project{}
	}

	middleware.RespondJSON(w, http.StatusOK, models.ProjectsResponse{Projects: projects})
}

// CreateProject handles POST /projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req models.CreateProjectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		slog.Warn("failed to decode create project request", "error", err)
		fields := map[string]string{"request": "invalid JSON"}
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	if req.Name == "" {
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", map[string]string{"name": "is required"})
		return
	}

	project, err := h.projectService.CreateProject(r.Context(), userClaims.UserID, req.Name, req.Description)
	if err != nil {
		slog.Error("failed to create project", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to create project", nil)
		return
	}

	middleware.RespondJSON(w, http.StatusCreated, project)
}

// GetProject handles GET /projects/:id
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodGet {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		if err.Error() == "project not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		slog.Error("failed to get project", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to get project", nil)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, project)
}

// UpdateProject handles PATCH /projects/:id
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodPatch {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req models.UpdateProjectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err != io.EOF {
		slog.Warn("failed to decode update project request", "error", err)
		fields := map[string]string{"request": "invalid JSON"}
		middleware.RespondErrorJSON(w, http.StatusBadRequest, "validation failed", fields)
		return
	}

	project, err := h.projectService.UpdateProject(r.Context(), projectID, userClaims.UserID, req.Name, req.Description)
	if err != nil {
		if err.Error() == "project not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "only project owner can update" {
			middleware.RespondErrorJSON(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		slog.Error("failed to update project", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to update project", nil)
		return
	}

	middleware.RespondJSON(w, http.StatusOK, project)
}

// DeleteProject handles DELETE /projects/:id
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request, projectID string) {
	if r.Method != http.MethodDelete {
		middleware.RespondErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		middleware.RespondErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	err := h.projectService.DeleteProject(r.Context(), projectID, userClaims.UserID)
	if err != nil {
		if err.Error() == "project not found" {
			middleware.RespondErrorJSON(w, http.StatusNotFound, "not found", nil)
			return
		}
		if err.Error() == "only project owner can delete" {
			middleware.RespondErrorJSON(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		slog.Error("failed to delete project", "error", err)
		middleware.RespondErrorJSON(w, http.StatusInternalServerError, "failed to delete project", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
