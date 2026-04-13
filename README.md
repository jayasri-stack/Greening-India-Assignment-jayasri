---

## Architecture Overview

The project follows a clean layered architecture for separation of concerns:

```text
handlers (HTTP layer)
    ↓
middleware (Auth, CORS, logging)
    ↓
service (Business logic)
    ↓
repository (Data access, SQL queries)
    ↓
database (Connection pooling, transactions)
Design Patterns Used
Repository Pattern for data access abstraction
Dependency Injection for better testability
Middleware Stack for authentication, logging, and CORS
Structured Logging for production-friendly observability
Consistent Error Handling with proper HTTP status codes
Key Tradeoffs
Raw SQL instead of ORM for performance and query control
Standard library net/http instead of a framework to keep dependencies minimal
JWT authentication instead of sessions for stateless scalability
Owner-based permissions instead of full RBAC for assignment simplicity
Project Structure
Greening-India-Assignment-jayasri/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── db/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repository/
│   └── service/
├── migrations/
│   ├── 001_init_schema.up.sql
│   ├── 001_init_schema.down.sql
│   ├── 002_seed_data.up.sql
│   └── 002_seed_data.down.sql
├── tools/
│   └── generate_hash/
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── seed.sh
└── README.md
Features
User registration and login
JWT-based authentication
Project CRUD operations
Task CRUD operations
Task filtering by status and assignee
PostgreSQL persistence
Docker support
Seed data for testing
Prerequisites

For local setup without Docker:

Go 1.21 or higher
PostgreSQL 12 or higher
Git

For Docker setup:

Docker Desktop
Docker Compose
Run with Docker

This is the easiest way to run the full stack.

1. Clone the repository
git clone https://github.com/jayasri-stack/Greening-India-Assignment-jayasri.git
cd Greening-India-Assignment-jayasri
2. Copy environment variables
cp .env.example .env
3. Start the application
docker compose up --build

This starts:

PostgreSQL on port 5432
API server on port 8080
4. Access the API
http://localhost:8080
Stop containers
docker compose down
Run in background
docker compose up -d --build
Run Locally Without Docker
1. Clone the repository
git clone https://github.com/jayasri-stack/Greening-India-Assignment-jayasri.git
cd Greening-India-Assignment-jayasri
2. Copy environment variables
cp .env.example .env
3. Update .env

Make sure your database credentials and JWT secret are set correctly.

Example:

DB_NAME=taskflow_db
DB_USER=taskflow_user
DB_PASSWORD=taskflow_password
JWT_SECRET=your_super_secret_jwt_key_change_this
PORT=8080
ENV=development
4. Create PostgreSQL database
createdb taskflow_db
createuser taskflow_user --password
5. Grant privileges
psql -U postgres -d taskflow_db -c "ALTER USER taskflow_user WITH ENCRYPTED PASSWORD 'taskflow_password';"
psql -U postgres -d taskflow_db -c "GRANT ALL PRIVILEGES ON DATABASE taskflow_db TO taskflow_user;"
6. Run migrations and seed data
chmod +x seed.sh
./seed.sh
7. Install dependencies
go mod tidy
8. Start the server
go run cmd/server/main.go

The API will be available at:

http://localhost:8080
Database Migrations
Migration Files
migrations/001_init_schema.up.sql - create tables and indexes
migrations/001_init_schema.down.sql - drop tables
migrations/002_seed_data.up.sql - insert test user and sample data
migrations/002_seed_data.down.sql - remove seeded data
Run Migrations Manually
psql -U taskflow_user -d taskflow_db -f migrations/001_init_schema.up.sql
psql -U taskflow_user -d taskflow_db -f migrations/002_seed_data.up.sql
Rollback
psql -U taskflow_user -d taskflow_db -f migrations/002_seed_data.down.sql
psql -U taskflow_user -d taskflow_db -f migrations/001_init_schema.down.sql
Test Credentials

After running migrations or starting with Docker, use:

Email:    test@example.com
Password: password123
API Base URL
http://localhost:8080
Authentication Endpoints
Register

POST /auth/register

Request
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securePassword123"
}
Response 201
{
  "token": "jwt-token",
  "user": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-04-12T21:55:29Z"
  }
}
Login

POST /auth/login

Request
{
  "email": "john@example.com",
  "password": "securePassword123"
}
Response 200
{
  "token": "jwt-token",
  "user": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-04-12T21:55:29Z"
  }
}
Project Endpoints

All project routes require:

Authorization: Bearer <token>
List Projects

GET /projects

Response 200
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
Create Project

POST /projects

Request
{
  "name": "My New Project",
  "description": "Optional description"
}
Response 201
{
  "id": "uuid",
  "name": "My New Project",
  "description": "Optional description",
  "owner_id": "current-user-uuid",
  "created_at": "2026-04-12T21:55:31Z"
}
Get Project by ID

GET /projects/{id}

Response 200
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
      "priority": "high"
    }
  ]
}
Update Project

PATCH /projects/{id}

Request
{
  "name": "Updated Name",
  "description": "Updated description"
}
Response 200

Updated project object

Delete Project

DELETE /projects/{id}

Response
204 No Content
Task Endpoints

All task routes require:

Authorization: Bearer <token>
List Tasks

GET /projects/{id}/tasks?status=todo&assignee={user_id}

Response 200
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
Create Task

POST /projects/{id}/tasks

Request
{
  "title": "Task Title",
  "description": "Optional description",
  "priority": "high",
  "assignee_id": "optional-user-uuid",
  "due_date": "2026-04-20"
}
Response 201

Created task object

Update Task

PATCH /tasks/{id}

Request
{
  "title": "Updated title",
  "description": "Updated description",
  "status": "done",
  "priority": "low",
  "assignee_id": "user-uuid",
  "due_date": "2026-04-20"
}
Response 200

Updated task object

Delete Task

DELETE /tasks/{id}

Response
204 No Content
Error Response Format
{
  "error": "validation failed",
  "fields": {
    "email": "is required",
    "password": "must be at least 6 characters"
  }
}
Common Status Codes
400 - Validation error
401 - Unauthorized
403 - Forbidden
404 - Not found
500 - Internal server error
Security Considerations
Passwords hashed with bcrypt
JWT authentication used for protected routes
Parameterized SQL queries help prevent SQL injection
Structured error responses avoid leaking sensitive internals
CORS support configured through middleware
Concurrency Notes
Go handles HTTP requests concurrently
PostgreSQL connection pooling is used
Repository layer is designed to be concurrency-safe
Graceful shutdown support can be added for cleaner deployments
What I Would Add With More Time
Integration tests
Pagination for list endpoints
Request validation middleware
Health check endpoint
Metrics and tracing
Rate limiting
Role-based access control
Redis caching
CI/CD pipeline
OpenAPI / Swagger docs
Troubleshooting
Docker does not start
Make sure Docker Desktop is running
Make sure WSL2 is enabled on Windows
Check whether ports 5432 and 8080 are already in use
Database connection issues
Verify PostgreSQL is running
Check .env values
Confirm database and user exist
Migration issues
Check user permissions in PostgreSQL
Confirm migration files exist in the migrations/ folder
API not reachable
Confirm the app is listening on port 8080
Check container logs:
docker compose logs

Built with Go, PostgreSQL, and clean backend design principles.


Use this as your new `README.md`.

One important correction: if your project root is actually `taskflow-backend/` inside the repo, then only the folder name in the structure and `cd` commands should be changed. The uploaded README had inconsistent path references, which is why I normalized it here. :contentReference[oaicite:1]{index=1}
