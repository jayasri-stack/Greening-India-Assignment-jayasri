package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/models"
	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/repository"
)

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo *repository.ProjectRepository
	taskRepo    *repository.TaskRepository
}

// NewProjectService creates a new project service
func NewProjectService(projectRepo *repository.ProjectRepository, taskRepo *repository.TaskRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
	}
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(ctx context.Context, userID, name string, description *string) (*models.Project, error) {
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	return s.projectRepo.CreateProject(ctx, name, userID, description)
}

// GetProjectByID retrieves a project with its tasks
func (s *ProjectService) GetProjectByID(ctx context.Context, projectID string) (*models.ProjectDetailResponse, error) {
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	// Fetch tasks for this project (concurrent-safe)
	tasks, err := s.taskRepo.ListTasksForProject(ctx, projectID, "", "")
	if err != nil {
		slog.Error("failed to fetch tasks", "error", err, "project_id", projectID)
		return nil, err
	}

	return &models.ProjectDetailResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
		Tasks:       tasks,
	}, nil
}

// ListProjectsForUser lists projects for a user
func (s *ProjectService) ListProjectsForUser(ctx context.Context, userID string) ([]models.Project, error) {
	return s.projectRepo.ListProjectsForUser(ctx, userID)
}

// UpdateProject updates project details
func (s *ProjectService) UpdateProject(ctx context.Context, projectID, userID string, name, description *string) (*models.Project, error) {
	// Check if user is owner
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	if project.OwnerID != userID {
		return nil, fmt.Errorf("only project owner can update")
	}

	return s.projectRepo.UpdateProject(ctx, projectID, name, description)
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID string) error {
	// Check if user is owner
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project == nil {
		return fmt.Errorf("project not found")
	}

	if project.OwnerID != userID {
		return fmt.Errorf("only project owner can delete")
	}

	return s.projectRepo.DeleteProject(ctx, projectID)
}
