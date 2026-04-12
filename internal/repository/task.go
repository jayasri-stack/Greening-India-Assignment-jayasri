package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/yourname/taskflow-backend/internal/db"
	"github.com/yourname/taskflow-backend/internal/models"
)

// TaskRepository handles task data operations
type TaskRepository struct {
	db *db.Database
	mu sync.RWMutex
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(database *db.Database) *TaskRepository {
	return &TaskRepository{
		db: database,
	}
}

// CreateTask inserts a new task
func (r *TaskRepository) CreateTask(ctx context.Context, title, projectID string, description *string, priority *string, assigneeID *string, dueDate *string) (*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	taskID := uuid.New().String()

	query := `
		INSERT INTO tasks (id, title, description, priority, project_id, assignee_id, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at
	`

	var dueDateValue interface{}
	if dueDate != nil {
		dueDateValue = *dueDate
	}

	var task models.Task
	err := r.db.QueryRow(query, taskID, title, description, *priority, projectID, assigneeID, dueDateValue).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&task.AssigneeID,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		slog.Error("failed to create task", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("task creation error: %w", err)
	}

	return &task, nil
}

// GetTaskByID retrieves a task by ID
func (r *TaskRepository) GetTaskByID(ctx context.Context, taskID string) (*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at FROM tasks WHERE id = $1`

	var task models.Task
	err := r.db.QueryRow(query, taskID).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&task.AssigneeID,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("failed to get task", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &task, nil
}

// ListTasksForProject returns tasks for a project with optional filtering
func (r *TaskRepository) ListTasksForProject(ctx context.Context, projectID, status, assigneeID string) ([]models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	baseQuery := strings.Builder{}
	baseQuery.WriteString(`SELECT id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at FROM tasks WHERE project_id = $1`)
	args := []interface{}{projectID}
	argIndex := 2

	if status != "" {
		baseQuery.WriteString(fmt.Sprintf(" AND status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if assigneeID != "" {
		baseQuery.WriteString(fmt.Sprintf(" AND assignee_id = $%d", argIndex))
		args = append(args, assigneeID)
		argIndex++
	}

	baseQuery.WriteString(" ORDER BY created_at DESC")

	rows, err := r.db.Query(baseQuery.String(), args...)
	if err != nil {
		slog.Error("failed to list tasks", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.ProjectID,
			&task.AssigneeID,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			slog.Error("failed to scan task", "error", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// UpdateTask updates task fields
func (r *TaskRepository) UpdateTask(ctx context.Context, taskID string, title, description *string, status, priority, assigneeID, dueDate *string) (*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *title)
		argIndex++
	}

	if description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *description)
		argIndex++
	}

	if status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}

	if priority != nil {
		updates = append(updates, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *priority)
		argIndex++
	}

	if assigneeID != nil {
		updates = append(updates, fmt.Sprintf("assignee_id = $%d", argIndex))
		args = append(args, assigneeID)
		argIndex++
	}

	if dueDate != nil {
		updates = append(updates, fmt.Sprintf("due_date = $%d", argIndex))
		args = append(args, *dueDate)
		argIndex++
	}

	if len(updates) == 0 {
		return r.GetTaskByID(ctx, taskID)
	}

	args = append(args, taskID)
	query := fmt.Sprintf("UPDATE tasks SET %s, updated_at = NOW() WHERE id = $%d RETURNING id, title, description, status, priority, project_id, assignee_id, due_date, created_at, updated_at", strings.Join(updates, ", "), argIndex)

	var task models.Task
	err := r.db.QueryRow(query, args...).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&task.AssigneeID,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("failed to update task", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &task, nil
}

// DeleteTask deletes a task by ID
func (r *TaskRepository) DeleteTask(ctx context.Context, taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, taskID)
	if err != nil {
		slog.Error("failed to delete task", "error", err, "task_id", taskID)
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}
