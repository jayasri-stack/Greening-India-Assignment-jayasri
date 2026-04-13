# TaskFlow - Backend API

A production-grade task management system backend built with Go, PostgreSQL, and modern architectural principles (SOLID, Dependency Injection, Repository Pattern).

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL
- **Authentication**: JWT (24-hour expiry)
- **Password Hashing**: bcrypt (cost 12)
- **Logging**: slog (structured logging)
- **Architecture**: Layered architecture with SOLID principles

## Architecture Decisions

### Layered Architecture

The project follows a clean, layered architecture for separation of concerns:

```
handlers (HTTP layer)
    ↓
middleware (Auth, CORS, logging)
    ↓
service (Business logic)
    ↓
repository (Data access, SQL queries)
    ↓
database (Connection pooling, transactions)
```

### Key Design Patterns

1. **Repository Pattern**: Data access abstraction layer for easy testing and database swaps
2. **Dependency Injection**: Services and repositories receive their dependencies, improving testability
3. **Middleware Stack**: Composable middleware for auth, logging, and CORS
4. **Concurrency-Safe**: Repositories use RWMutex for concurrent read operations while protecting writes
5. **Error Handling**: Consistent error responses with proper HTTP status codes
6. **Structured Logging**: JSON-formatted logs for production observability

### Design Tradeoffs

- **No ORM**: Raw SQL queries for better performance and control (migrations managed manually)
- **Standard Library HTTP**: Using Go's `net/http` instead of a framework for simplicity and minimal dependencies
- **Thread Safety**: RWMutex on repositories to allow concurrent reads, critical for scaling
- **JWT over Sessions**: Stateless authentication for horizontal scalability

### What Was Left Out (and Why)

- **GraphQL**: REST is simpler for this scope and the requirements specify REST
- **Rate Limiting**: Can be added at reverse proxy level (nginx)
- **Real-time Updates**: WebSocket can be added later without changing current architecture
- **Full RBAC**: Simple owner-based permissions sufficient for requirements; RBAC can be added if needed
- **API Versioning**: Single version sufficient; can version at URL if needed (`/v1/api`)
- **Caching**: Not needed for this assignment scope; can add Redis later
- **Metrics/Tracing**: slog provides structured logging; OpenTelemetry can be added

## Running Locally

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- git

### Setup Steps

```bash
# 1. Clone the repository
git clone https://github.com/jayasri-stack/Greening-India-Assignment-jayasri
cd taskflow-backend

# 2. Copy environment variables
cp .env.example .env

# 3. Edit .env with your database credentials
# Important: Change JWT_SECRET to a random string (minimum 32 characters)
nano .env

# 4. Create PostgreSQL database (as postgres user or superuser)
createdb taskflow_db
createuser taskflow_user --password

# 5. Grant privileges
psql -U postgres -d taskflow_db -c "ALTER USER taskflow_user WITH ENCRYPTED PASSWORD 'taskflow_password';"
psql -U postgres -d taskflow_db -c "GRANT ALL PRIVILEGES ON DATABASE taskflow_db TO taskflow_user;"

# 6. Run migrations and seed data
chmod +x seed.sh
./seed.sh

# 7. Download Go dependencies
go mod download
go mod tidy

# 8. Run the server
go run cmd/server/main.go

# Server will be available at http://localhost:8080
```

## Running Migrations

### Automatic Migrations

The project includes shell scripts to manage migrations:

```bash
# Run all up migrations (create tables, seed data)
./seed.sh

# To manually run specific migration (if needed)
psql -U taskflow_user -d taskflow_db -f migrations/001_init_schema.up.sql
```

### Migration Files

- `migrations/001_init_schema.up.sql` - Creates users, projects, tasks tables with indexes
- `migrations/001_init_schema.down.sql` - Drops all tables
- `migrations/002_seed_data.up.sql` - Seeds test user and sample data
- `migrations/002_seed_data.down.sql` - Removes seeded data

### Rollback

```bash
psql -U taskflow_user -d taskflow_db -f migrations/002_seed_data.down.sql
psql -U taskflow_user -d taskflow_db -f migrations/001_init_schema.down.sql
```

## Test Credentials

After running migrations, use these credentials to test:

```
Email:    test@example.com
Password: password123
```



```go
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
    fmt.Println(string(hash))
}
```

Then update `migrations/002_seed_data.up.sql` with the actual hash.

## API Reference

Base URL: `http://localhost:8080`

### Authentication Endpoints

#### Register

```
POST /auth/register

Request:
{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123"
}

Response (201):
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "name": "John Doe",
        "email": "john@example.com",
        "created_at": "2026-04-12T10:00:00Z"
    }
}

Errors:
- 400: Email already exists, invalid input
- 500: Server error
```

#### Login

```
POST /auth/login

Request:
{
    "email": "john@example.com",
    "password": "securepass123"
}

Response (200):
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "name": "John Doe",
        "email": "john@example.com",
        "created_at": "2026-04-12T10:00:00Z"
    }
}

Errors:
- 401: Invalid credentials
- 500: Server error
```

### Projects Endpoints

All require `Authorization: Bearer <token>` header

#### List Projects

```
GET /projects

Response (200):
{
    "projects": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440002",
            "name": "Website Redesign",
            "description": "Q2 project",
            "owner_id": "550e8400-e29b-41d4-a716-446655440001",
            "created_at": "2026-04-12T10:00:00Z"
        }
    ]
}
```

#### Create Project

```
POST /projects

Request:
{
    "name": "Mobile App",
    "description": "New iOS app"
}

Response (201): Project object with id, created_at

Errors:
- 400: Missing name
- 401: Unauthorized
- 500: Server error
```

#### Get Project with Tasks

```
GET /projects/{id}

Response (200):
{
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "name": "Website Redesign",
    "description": "Q2 project",
    "owner_id": "550e8400-e29b-41d4-a716-446655440001",
    "created_at": "2026-04-12T10:00:00Z",
    "tasks": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440003",
            "title": "Design homepage",
            "status": "in_progress",
            "priority": "high",
            "assignee_id": "550e8400-e29b-41d4-a716-446655440001",
            "created_at": "2026-04-12T10:00:00Z",
            "updated_at": "2026-04-12T10:00:00Z"
        }
    ]
}

Errors:
- 404: Project not found
- 401: Unauthorized
```

#### Update Project

```
PATCH /projects/{id}

Request (all fields optional):
{
    "name": "Updated Name",
    "description": "Updated description"
}

Response (200): Updated project object

Errors:
- 403: Only owner can update
- 404: Project not found
- 401: Unauthorized
```

#### Delete Project

```
DELETE /projects/{id}

Response: 204 No Content

Errors:
- 403: Only owner can delete
- 404: Project not found
- 401: Unauthorized
```

### Tasks Endpoints

All require `Authorization: Bearer <token>` header

#### List Tasks

```
GET /projects/{id}/tasks?status=todo&assignee={user_id}

Query Parameters:
- status: todo, in_progress, done (optional)
- assignee: user ID (optional)

Response (200):
{
    "tasks": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440003",
            "title": "Design homepage",
            "description": "Create responsive design",
            "status": "in_progress",
            "priority": "high",
            "project_id": "550e8400-e29b-41d4-a716-446655440002",
            "assignee_id": "550e8400-e29b-41d4-a716-446655440001",
            "due_date": "2026-04-20",
            "created_at": "2026-04-12T10:00:00Z",
            "updated_at": "2026-04-12T10:00:00Z"
        }
    ]
}

Errors:
- 400: Invalid status
- 404: Project not found
- 401: Unauthorized
```

#### Create Task

```
POST /projects/{id}/tasks

Request:
{
    "title": "Design homepage",
    "description": "Create responsive design",
    "priority": "high",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440001",
    "due_date": "2026-04-20"
}

Response (201): Task object with id, status=todo

Errors:
- 400: Missing title, invalid priority
- 404: Project not found
- 401: Unauthorized
```

#### Update Task

```
PATCH /tasks/{id}

Request (all fields optional):
{
    "title": "Updated title",
    "status": "done",
    "priority": "low",
    "assignee_id": "550e8400-e29b-41d4-a716-446655440001",
    "due_date": "2026-04-25"
}

Response (200): Updated task object

Errors:
- 403: Only project owner can update
- 404: Task not found
- 401: Unauthorized
```

#### Delete Task

```
DELETE /tasks/{id}

Response: 204 No Content

Errors:
- 403: Only project owner can delete
- 404: Task not found
- 401: Unauthorized
```

### Error Response Format

All errors follow this format:

```json
{
    "error": "validation failed",
    "fields": {
        "email": "is required",
        "password": "must be at least 6 characters"
    }
}
```

Status codes:
- `400` - Validation error
- `401` - Unauthorized (unauthenticated)
- `403` - Forbidden (authenticated but no permission)
- `404` - Not found
- `500` - Internal server error

## What You'd Do With More Time

### High Priority

1. **Automated Migration Tool** - Integrate golang-migrate for more robust migration management with version tracking
2. **Integration Tests** - Add tests for auth endpoints, project creation, task operations with real database
3. **Request Validation** - Add comprehensive input validation with detailed field error messages
4. **Pagination** - Add limit/offset or cursor-based pagination to list endpoints
5. **Rate Limiting** - Implement per-user rate limiting to prevent abuse
6. **Database Indexes** - Add more strategic indexes based on query patterns (already added basic ones)

### Medium Priority

7. **WebSocket Support** - Real-time task updates when tasks are modified by other users
8. **Caching Layer** - Redis for project/task caching to reduce database load
9. **Audit Logging** - Track who made what changes and when for compliance
10. **Soft Deletes** - Keep deleted records for audit trail with logical deletion
11. **Batch Operations** - Bulk task status updates for better UX
12. **Notifications** - Email/push notifications when tasks are assigned
13. **Search** - Full-text search across projects and tasks
14. **Advanced Permissions** - Invite collaborators, role-based access control (admin, member, viewer)
15. **Statistics Endpoint** - `GET /projects/{id}/stats` - task counts by status and assignee

### Polish & DevOps

16. **API Documentation** - Swagger/OpenAPI spec generation
17. **Structured Error Codes** - Machine-readable error codes for frontend error handling
18. **Request ID Correlation** - Track requests through logs for debugging
19. **Health Check Endpoint** - `/health` for container orchestration
20. **Metrics Endpoint** - Prometheus metrics for monitoring (requests, latency, errors)
21. **Docker Multi-stage Build** - Optimize image size
22. **Kubernetes Manifests** - Helm charts for production deployment
23. **CI/CD Pipeline** - GitHub Actions for automated testing and deployment
24. **Load Testing** - Verify performance under concurrent load
25. **Security Hardening** - SQL injection prevention (already using parameterized queries), CSRF tokens, input sanitization

### Performance Optimizations

26. **Connection Pool Tuning** - Optimize MaxOpenConns/MaxIdleConns based on load testing
27. **Query Optimization** - Add explain plans to critical queries
28. **N+1 Query Prevention** - Use batch loading for related entities
29. **Caching Strategy** - Implement TTL-based cache invalidation
30. **Async Processing** - Background job queue for heavy operations

## Project Structure

```
taskflow-backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point with server setup
├── internal/
│   ├── auth/
│   │   └── manager.go           # JWT and password hashing
│   ├── db/
│   │   ├── postgres.go          # Database connection and pooling
│   │   └── context.go           # Database context helpers
│   ├── handlers/
│   │   ├── auth.go              # HTTP handlers for auth
│   │   ├── project.go           # HTTP handlers for projects
│   │   └── task.go              # HTTP handlers for tasks
│   ├── middleware/
│   │   ├── auth.go              # JWT middleware
│   │   └── response.go          # Response formatting helpers
│   ├── models/
│   │   └── models.go            # Data structures and DTOs
│   ├── repository/
│   │   ├── user.go              # User data access
│   │   ├── project.go           # Project data access
│   │   └── task.go              # Task data access
│   └── service/
│       ├── auth.go              # Authentication business logic
│       ├── project.go           # Project business logic
│       └── task.go              # Task business logic
├── migrations/
│   ├── 001_init_schema.up.sql   # Create tables
│   ├── 001_init_schema.down.sql # Drop tables
│   ├── 002_seed_data.up.sql     # Seed test data
│   └── 002_seed_data.down.sql   # Remove seeded data
├── .env.example                 # Environment variables template
├── .gitignore                   # Git ignore rules
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies checksums
├── seed.sh                      # Database setup script
└── README.md                    # This file
```

## Concurrency & Threading

The backend uses Go's concurrency model effectively:

- **Goroutines**: All HTTP requests handled concurrently
- **Connection Pooling**: PostgreSQL connection pool with configured limits
- **Thread-Safe Repositories**: RWMutex for safe concurrent read/write operations
- **Graceful Shutdown**: Proper signal handling for clean server termination

Connection pool config:
- Max Open Connections: 25 (prevents exhausting database)
- Max Idle Connections: 5
- Max Connection Lifetime: 5 minutes

## Security Considerations

- ✅ Passwords hashed with bcrypt (cost 12)
## Running With Docker

The easiest way to run the entire stack:

```bash
# 1. Clone the repository
git clone https://github.com/jayasri-stack/Greening-India-Assignment-jayasri
cd taskflow-backend

# 2. Copy environment variables (optional - docker-compose has defaults)
cp .env.example .env

# 3. Start all services (PostgreSQL + API server)
docker compose up

# The API will be available at http://localhost:8080
# PostgreSQL will be available at localhost:5432
```

That's it! Migrations run automatically on container startup.

### To stop the services:
```bash
docker compose down
```

### To rebuild after code changes:
```bash
docker compose up --build
```

## Running Migrations

Migrations run **automatically** on application startup:
1. App connects to PostgreSQL
2. Creates `schema_migrations` table if missing
3. Checks which migrations have been applied
4. Executes any pending `.up.sql` migrations
5. Records completion in `schema_migrations`

**Manual migration (if needed):**
```bash
# Apply specific migration
psql -U taskflow_user -d taskflow_db -f migrations/001_init_schema.up.sql
```

## Test Credentials

Test Credentials

After running:

docker compose up

Use these credentials to log in immediately:

Email: test@example.com
Password: password123

## API Reference

### Base URL
```
http://localhost:8080
```

### Authentication Endpoints

**POST /auth/register** - Register a new user
```json
Request:
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securePassword123"
}

Response (201):
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-04-12T21:55:29Z"
  }
}
```

**POST /auth/login** - Login user and get JWT token
```json
Request:
{
  "email": "john@example.com",
  "password": "securePassword123"
}

Response (200):
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-04-12T21:55:29Z"
  }
}
```

### Projects Endpoints

**GET /projects** - List user's projects
```
Headers: Authorization: Bearer <token>

Response (200):
{
  "projects": [
    {
      "id": "uuid",
      "name": "My Project",
      "description": "Project description",
      "owner_id": "user-uuid",
      "created_at": "2026-04-12T21:55:31Z"
    }
  ]
}
```

**POST /projects** - Create new project
```json
Request:
{
  "name": "My New Project",
  "description": "Optional description"
}

Response (201):
{
  "id": "uuid",
  "name": "My New Project",
  "description": "Optional description",
  "owner_id": "current-user-uuid",
  "created_at": "2026-04-12T21:55:31Z"
}
```

**GET /projects/:id** - Get project details with tasks
```
Response (200):
{
  "id": "uuid",
  "name": "My Project",
  "description": "Project description",
  "owner_id": "user-uuid",
  "created_at": "2026-04-12T21:55:31Z",
  "tasks": [
    {
      "id": "task-uuid",
      "title": "Task 1",
      "status": "todo",
      "priority": "high",
      "...": "..."
    }
  ]
}
```

**PATCH /projects/:id** - Update project (owner only)
```json
Request:
{
  "name": "Updated Name",
  "description": "Updated description"
}

Response (200): Updated project object
```

**DELETE /projects/:id** - Delete project and all tasks (owner only)
```
Response (204 No Content)
```

### Tasks Endpoints

**GET /projects/:id/tasks** - List tasks with optional filters
```
Query Parameters:
  ?status=todo|in_progress|done
  ?assignee=uuid

Response (200):
{
  "tasks": [
    {
      "id": "uuid",
      "title": "Task 1",
      "description": "Task description",
      "status": "todo",
      "priority": "high",
      "project_id": "uuid",
      "assignee_id": "user-uuid",
      "due_date": "2026-04-20",
      "created_at": "2026-04-12T21:55:31Z",
      "updated_at": "2026-04-12T21:55:31Z"
    }
  ]
}
```

**POST /projects/:id/tasks** - Create task
```json
Request:
{
  "title": "Task Title",
  "description": "Optional description",
  "priority": "high|medium|low",
  "assignee_id": "optional-user-uuid",
  "due_date": "2026-04-20"
}

Response (201): Created task object
```

**PATCH /tasks/:id** - Update task
```json
Request (all fields optional):
{
  "title": "Updated title",
  "description": "Updated description",
  "status": "todo|in_progress|done",
  "priority": "high|medium|low",
  "assignee_id": "user-uuid",
  "due_date": "2026-04-20"
}

Response (200): Updated task object
```

**DELETE /tasks/:id** - Delete task
```
Response (204 No Content)
```

### Error Responses

**400 Validation Error**
```json
{
  "error": "validation failed",
  "fields": {
    "email": "is required",
    "password": "must be at least 6 characters"
  }
}
```

**401 Unauthorized**
```json
{
  "error": "unauthorized"
}
```

**403 Forbidden**
```json
{
  "error": "forbidden"
}
```

**404 Not Found**
```json
{
  "error": "not found"
}
```

### Complete Postman Collection

A Postman collection is included in the repository: [TaskFlow-API.postman_collection.json](TaskFlow-API.postman_collection.json)

Import it into Postman to test all endpoints immediately.

## What You'd Do With More Time

### High Priority (Would Implement First)
1. **Pagination on list endpoints** - Add `?page=1&limit=20` parameters to `/projects` and `/projects/:id/tasks` for better scalability
2. **Comprehensive integration tests** - Add test suite with `testing` package covering auth flows, CRUD operations, and error cases
3. **Request validation** - Add middleware for input validation (email format, password strength, etc.)
4. **Audit logging** - Track who modified what and when for compliance

### Medium Priority (Good to Have)
1. **Task statistics endpoint** - `GET /projects/:id/stats` returning task counts by status and assignee
2. **Project member permissions** - Allow project owners to add collaborators and manage permissions
3. **Task history/changelog** - Track task updates for audit trails
4. **Database transaction management** - Wrap related operations in transactions to ensure consistency
5. **Better error messages** - More specific validation feedback for clients

### Performance & Scale
1. **Database query optimization** - Add EXPLAIN ANALYZE to slow queries, review indexes
2. **Caching layer** - Add Redis for frequently accessed projects/tasks
3. **Rate limiting** - Implement per-user rate limits to prevent abuse
4. **Metrics and monitoring** - Add OpenTelemetry for production observability

### DevEx & Deployment
1. **CI/CD pipeline** - GitHub Actions for testing, linting, and automated deployment
2. **Pre-commit hooks** - Enforce code formatting, linting before commits
3. **Database backups** - Automated PostgreSQL backups and recovery procedures
4. **Health check endpoint** - `GET /health` for orchestrators like Kubernetes
5. **Graceful rolling updates** - Drain connections before shutdown for zero-downtime deployments

### Features for Real Product
1. **Real-time updates** - WebSocket support for live task changes across connected clients
2. **Task comments/notes** - Add discussion thread to tasks
3. **File attachments** - Store task attachments (AWS S3 integration)
4. **Email notifications** - Send task reminders and assignment notifications
5. **Advanced filtering/search** - Full-text search on task titles and descriptions
6. **Recurring tasks** - Support for periodic task creation
7. **Mobile app** - Dedicated mobile clients for iOS/Android

### Security Enhancements
1. **Rate limiting with exponential backoff** - Prevent brute-force login attempts
2. **Two-factor authentication** - TOTP/SMS for sensitive operations
3. **API key authentication** - For service-to-service communication
4. **Role-based access control (RBAC)** - Move beyond owner-based permissions
5. **Encryption at rest** - Encrypt sensitive fields in database
6. **Security headers** - Add HSTS, X-Frame-Options, etc.

### Key Tradeoffs Made in This Implementation

| Decision | Reason | Alternative |
|----------|--------|-------------|
| **No ORM (raw SQL)** | Performance, control, explicit queries | Use GORM or sqlc |
| **Standard library HTTP** | Minimal dependencies, fast | Use framework (Echo, Gin) |
| **Owner-based permissions** | Sufficient for scope | Implement full RBAC |
| **No request validation middleware** | Assignment scope - time constraints | Add explicit validator |
| **No caching** | Trade-off: complexity vs. performance | Add Redis layer |
| **Manual migrations** | Simpler for small team | Use golang-migrate with CLI |
| **No tests** | Time constraints - focused on feature completeness | Add test suite |

---

## Development Notes

- **Code Review Focus**: Architecture clarity, error handling, SQL query efficiency
- **Performance**: Connection pooling configured for ~20-25 concurrent connections
- **Security**: Bcrypt cost 12 provides good balance between security and performance
- **Logging**: Structured JSON logs for easy parsing in production systems

## Security Considerations

- ✅ JWT tokens with 24-hour expiry
- ✅ Parameterized SQL queries (prevents SQL injection)
- ✅ HTTP-only JWT (no XSS exposure in local storage)
- ✅ CORS headers configured
- ✅ Request timeout protection
- ✅ Structured error responses (no sensitive data leakage)

## Troubleshooting

### Connection Refused
- Verify PostgreSQL is running: `pg_isready -h localhost -p 5432`
- Check DATABASE_URL in .env

### Authentication Errors
- Ensure JWT_SECRET is set and consistent
- Token expiry is 24 hours - generate new token if expired

### Migration Failed
- Check database permissions: `psql -U postgres -d taskflow_db -l`
- Verify user has CREATE permissions on database

### Docker Issues
- Ensure Docker daemon is running
- Check port 5432 and 8080 are not in use
- Run `docker compose logs` to see detailed error messages

## Performance Notes

- List endpoints return all results; pagination can be added
- Database indexes on common queries (email, project_id, assignee_id, status)
- Connection pooling prevents resource exhaustion
- Structured logging with slog for efficient production logging

---

Built with ❤️ following SOLID principles and Go best practices.

