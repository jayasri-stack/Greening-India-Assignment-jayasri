package service

import (
	"context"
	"fmt"

	"jayasri-stack/Greening-India-Assignment-jayasri/internal/models"
	"jayasri-stack/Greening-India-Assignment-jayasri/internal/repository"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo    *repository.TaskRepository
	projectRepo *repository.ProjectRepository
}

// NewTaskService creates a new task service
func NewTaskService(taskRepo *repository.TaskRepository, projectRepo *repository.ProjectRepository) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
	}
}

// CreateTask creates a new task in a project
func (s *TaskService) CreateTask(ctx context.Context, userID, projectID, title string, req *models.CreateTaskRequest) (*models.Task, error) {
	// Verify user has access to project
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	// Validate task title
	if title == "" {
		return nil, fmt.Errorf("task title is required")
	}

	// Validate priority
	if req.Priority != "" {
		if req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" {
			return nil, fmt.Errorf("invalid priority")
		}
	} else {
		req.Priority = "medium"
	}

	return s.taskRepo.CreateTask(ctx, title, projectID, req.Description, &req.Priority, req.AssigneeID, req.DueDate)
}

// GetTaskByID retrieves a task
func (s *TaskService) GetTaskByID(ctx context.Context, taskID string) (*models.Task, error) {
	return s.taskRepo.GetTaskByID(ctx, taskID)
}

// ListTasksForProject lists tasks with filters
func (s *TaskService) ListTasksForProject(ctx context.Context, projectID, status, assigneeID string) ([]models.Task, error) {
	// Verify project exists
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	// Validate status if provided
	if status != "" {
		if status != "todo" && status != "in_progress" && status != "done" {
			return nil, fmt.Errorf("invalid status")
		}
	}

	return s.taskRepo.ListTasksForProject(ctx, projectID, status, assigneeID)
}

// UpdateTask updates a task
func (s *TaskService) UpdateTask(ctx context.Context, taskID, userID string, req *models.UpdateTaskRequest) (*models.Task, error) {
	// Get task
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Get project to verify ownership
	project, err := s.projectRepo.GetProjectByID(ctx, task.ProjectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	// Verify user is project owner (can update any task)
	if project.OwnerID != userID {
		return nil, fmt.Errorf("only project owner can update tasks")
	}

	// Validate status if provided
	if req.Status != nil {
		if *req.Status != "todo" && *req.Status != "in_progress" && *req.Status != "done" {
			return nil, fmt.Errorf("invalid status")
		}
	}

	// Validate priority if provided
	if req.Priority != nil {
		if *req.Priority != "low" && *req.Priority != "medium" && *req.Priority != "high" {
			return nil, fmt.Errorf("invalid priority")
		}
	}

	return s.taskRepo.UpdateTask(ctx, taskID, req.Title, req.Description, req.Status, req.Priority, req.AssigneeID, req.DueDate)
}

// DeleteTask deletes a task
func (s *TaskService) DeleteTask(ctx context.Context, taskID, userID string) error {
	// Get task
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	if task == nil {
		return fmt.Errorf("task not found")
	}

	// Get project to check permissions
	project, err := s.projectRepo.GetProjectByID(ctx, task.ProjectID)
	if err != nil {
		return err
	}

	if project == nil {
		return fmt.Errorf("project not found")
	}

	// Only project owner can delete tasks
	if project.OwnerID != userID {
		return fmt.Errorf("only project owner can delete tasks")
	}

	return s.taskRepo.DeleteTask(ctx, taskID)
}
