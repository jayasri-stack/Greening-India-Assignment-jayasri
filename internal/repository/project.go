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

// ProjectRepository handles project data operations
type ProjectRepository struct {
	db *db.Database
	mu sync.RWMutex
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(database *db.Database) *ProjectRepository {
	return &ProjectRepository{
		db: database,
	}
}

// CreateProject creates a new project
func (r *ProjectRepository) CreateProject(ctx context.Context, name, ownerID string, description *string) (*models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	projectID := uuid.New().String()

	query := `
		INSERT INTO projects (id, name, description, owner_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, name, description, owner_id, created_at
	`

	project := &models.Project{}
	err := r.db.QueryRow(query, projectID, name, description, ownerID).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt,
	)

	if err != nil {
		slog.Error("failed to create project", "error", err, "owner_id", ownerID)
		return nil, fmt.Errorf("project creation error: %w", err)
	}

	slog.Info("project created", "project_id", project.ID, "owner_id", ownerID)
	return project, nil
}

// GetProjectByID retrieves a project by ID
func (r *ProjectRepository) GetProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`

	project := &models.Project{}
	err := r.db.QueryRow(query, projectID).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("failed to get project", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("database error: %w", err)
	}

	return project, nil
}

// ListProjectsForUser returns projects the user owns or has tasks in (concurrent-safe)
func (r *ProjectRepository) ListProjectsForUser(ctx context.Context, userID string) ([]models.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		slog.Error("failed to list projects", "error", err, "user_id", userID)
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			slog.Error("failed to scan project", "error", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, rows.Err()
}

// UpdateProject updates project details (owner only)
func (r *ProjectRepository) UpdateProject(ctx context.Context, projectID string, name, description *string) (*models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *name)
		argIndex++
	}

	if description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *description)
		argIndex++
	}

	if len(updates) == 0 {
		project, err := r.GetProjectByID(ctx, projectID)
		return project, err
	}

	args = append(args, projectID)

	query := fmt.Sprintf(`
		UPDATE projects
		SET %s
		WHERE id = $%d
		RETURNING id, name, description, owner_id, created_at
	`, strings.Join(updates, ", "), argIndex)

	project := &models.Project{}
	err := r.db.QueryRow(query, args...).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt,
	)

	if err != nil {
		slog.Error("failed to update project", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("update error: %w", err)
	}

	return project, nil
}

// DeleteProject deletes a project and its tasks
func (r *ProjectRepository) DeleteProject(ctx context.Context, projectID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Delete tasks first (foreign key constraint)
	if _, err := r.db.Exec("DELETE FROM tasks WHERE project_id = $1", projectID); err != nil {
		slog.Error("failed to delete tasks", "error", err, "project_id", projectID)
		return fmt.Errorf("delete tasks error: %w", err)
	}

	// Delete project
	result, err := r.db.Exec("DELETE FROM projects WHERE id = $1", projectID)
	if err != nil {
		slog.Error("failed to delete project", "error", err, "project_id", projectID)
		return fmt.Errorf("delete error: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("project not found")
	}

	slog.Info("project deleted", "project_id", projectID)
	return nil
}
