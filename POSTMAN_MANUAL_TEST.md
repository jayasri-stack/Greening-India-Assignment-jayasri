# TaskFlow API - Manual Testing Guide

## Base URL
```
http://localhost:8080
```

## Authentication
All endpoints except `/auth/register` and `/auth/login` require a Bearer token in the `Authorization` header.

**Header Format:**
```
Authorization: Bearer <token>
```

---

## ✅ Endpoints Summary (ALL TESTED AND WORKING)

| # | Method | Endpoint | Auth | Status | Purpose |
|---|--------|----------|------|--------|---------|
| 1 | POST | /auth/register | No | ✅ | Register new user |
| 2 | POST | /auth/login | No | ✅ | Login user and get token |
| 3 | GET | /projects | Yes | ✅ | List user's projects |
| 4 | POST | /projects | Yes | ✅ | Create new project |
| 5 | GET | /projects/{id} | Yes | ✅ | Get project details with tasks |
| 6 | PATCH | /projects/{id} | Yes | ✅ | Update project |
| 7 | DELETE | /projects/{id} | Yes | ✅ | Delete project |
| 8 | GET | /projects/{id}/tasks | Yes | ✅ | List tasks in project |
| 9 | POST | /projects/{id}/tasks | Yes | ✅ | Create task in project |
| 10 | PATCH | /tasks/{id} | Yes | ✅ | Update task |
| 11 | DELETE | /tasks/{id} | Yes | ✅ | Delete task |

---

## 📝 Detailed Endpoint Documentation

### 1. POST /auth/register
**Purpose:** Register a new user

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securePassword123"
}
```

**Success Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-here",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-04-12T21:55:29Z"
  }
}
```

**Test Data:**
```
Name: John Doe
Email: john@example.com
Password: securePassword123
```

---

### 2. POST /auth/login
**Purpose:** Login user and get JWT token

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "securePassword123"
}
```

**Success Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-here",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-04-12T21:55:29Z"
  }
}
```

**Test Data:**
```
Email: john@example.com
Password: securePassword123
```

---

### 3. GET /projects
**Purpose:** List all projects for the authenticated user

**Request Headers:**
```
Authorization: Bearer <your_token>
```

**Success Response (200):**
```json
{
  "projects": [
    {
      "id": "uuid-here",
      "name": "My Project",
      "description": "Project description",
      "owner_id": "user-uuid",
      "created_at": "2026-04-12T21:55:31Z"
    }
  ]
}
```

---

### 4. POST /projects
**Purpose:** Create a new project

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <your_token>
```

**Request Body:**
```json
{
  "name": "My New Project",
  "description": "This is a test project"
}
```

**Success Response (201):**
```json
{
  "id": "uuid-here",
  "name": "My New Project",
  "description": "This is a test project",
  "owner_id": "user-uuid",
  "created_at": "2026-04-12T21:55:31Z"
}
```

**Test Data:**
```
Name: Development Project
Description: A project for development tasks
```

---

### 5. GET /projects/{id}
**Purpose:** Get project details with all tasks

**Request Headers:**
```
Authorization: Bearer <your_token>
```

**URL:**
```
/projects/16b0222b-9ad9-441a-ba8c-08a8b21aff3a
```

**Success Response (200):**
```json
{
  "id": "uuid-here",
  "name": "My Project",
  "description": "Project description",
  "owner_id": "user-uuid",
  "created_at": "2026-04-12T21:55:31Z",
  "tasks": [
    {
      "id": "task-uuid",
      "title": "Task 1",
      "description": "Task description",
      "status": "todo",
      "priority": "high",
      "project_id": "project-uuid",
      "assignee_id": "user-uuid",
      "due_date": "2026-04-20",
      "created_at": "2026-04-12T21:55:31Z",
      "updated_at": "2026-04-12T21:55:31Z"
    }
  ]
}
```

---

### 6. PATCH /projects/{id}
**Purpose:** Update project details

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <your_token>
```

**Request Body:**
```json
{
  "name": "Updated Project Name",
  "description": "Updated description"
}
```

**Success Response (200):**
```json
{
  "id": "uuid-here",
  "name": "Updated Project Name",
  "description": "Updated description",
  "owner_id": "user-uuid",
  "created_at": "2026-04-12T21:55:31Z"
}
```

**Test Data:**
```
Name: Enterprise Project
Description: Updated with new requirements
```

---

### 7. DELETE /projects/{id}
**Purpose:** Delete a project and all its tasks

**Request Headers:**
```
Authorization: Bearer <your_token>
```

**Success Response (204 No Content)**
```
Empty body
```

---

### 8. GET /projects/{id}/tasks
**Purpose:** List all tasks in a project with optional filtering

**Request Headers:**
```
Authorization: Bearer <your_token>
```

**Query Parameters (Optional):**
- `status`: Filter by status (todo, in_progress, done)
- `assignee`: Filter by assignee ID

**Examples:**
```
/projects/{id}/tasks?status=in_progress
/projects/{id}/tasks?assignee=user-uuid
/projects/{id}/tasks?status=done&assignee=user-uuid
```

**Success Response (200):**
```json
{
  "tasks": [
    {
      "id": "task-uuid",
      "title": "Task 1",
      "description": "Task description",
      "status": "todo",
      "priority": "high",
      "project_id": "project-uuid",
      "assignee_id": "user-uuid",
      "due_date": "2026-04-20",
      "created_at": "2026-04-12T21:55:31Z",
      "updated_at": "2026-04-12T21:55:31Z"
    }
  ]
}
```

---

### 9. POST /projects/{id}/tasks
**Purpose:** Create a new task in a project

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <your_token>
```

**Request Body:**
```json
{
  "title": "Task Title",
  "description": "Task description",
  "priority": "high",
  "assignee_id": "user-uuid",
  "due_date": "2026-04-20"
}
```

**Priority Options:** low, medium, high
**Status Defaults To:** todo

**Success Response (201):**
```json
{
  "id": "task-uuid",
  "title": "Task Title",
  "description": "Task description",
  "status": "todo",
  "priority": "high",
  "project_id": "project-uuid",
  "assignee_id": "user-uuid",
  "due_date": "2026-04-20",
  "created_at": "2026-04-12T21:55:31Z",
  "updated_at": "2026-04-12T21:55:31Z"
}
```

**Test Data:**
```
Title: Implement Authentication
Description: Add JWT authentication to the system
Priority: high
Due Date: 2026-04-25
```

---

### 10. PATCH /tasks/{id}
**Purpose:** Update task details

**Request Headers:**
```
Content-Type: application/json
Authorization: Bearer <your_token>
```

**Request Body (All fields optional):**
```json
{
  "title": "Updated Task Title",
  "description": "Updated description",
  "status": "in_progress",
  "priority": "medium",
  "assignee_id": "user-uuid",
  "due_date": "2026-04-25"
}
```

**Status Options:** todo, in_progress, done
**Priority Options:** low, medium, high

**Success Response (200):**
```json
{
  "id": "task-uuid",
  "title": "Updated Task Title",
  "description": "Updated description",
  "status": "in_progress",
  "priority": "medium",
  "project_id": "project-uuid",
  "assignee_id": "user-uuid",
  "due_date": "2026-04-25",
  "created_at": "2026-04-12T21:55:31Z",
  "updated_at": "2026-04-12T21:55:32Z"
}
```

**Test Data:**
```
Status: in_progress
Priority: medium
```

---

### 11. DELETE /tasks/{id}
**Purpose:** Delete a task

**Request Headers:**
```
Authorization: Bearer <your_token>
```

**Success Response (204 No Content)**
```
Empty body
```

---

## 🧪 Complete Postman Test Workflow

Follow this sequence to test all endpoints:

### Step 1: Register User
```
POST http://localhost:8080/auth/register
Content-Type: application/json

{
  "name": "Test User",
  "email": "test@example.com",
  "password": "test123"
}
```
**Save the token from response**

### Step 2: Login
```
POST http://localhost:8080/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "test123"
}
```
**Save the token from response**

### Step 3: Create Project
```
POST http://localhost:8080/projects
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "Test Project",
  "description": "A project for testing"
}
```
**Save the project ID from response**

### Step 4: Get Projects
```
GET http://localhost:8080/projects
Authorization: Bearer <token>
```

### Step 5: Create Task
```
POST http://localhost:8080/projects/<project-id>/tasks
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Test Task",
  "description": "A task for testing",
  "priority": "high",
  "due_date": "2026-04-25"
}
```
**Save the task ID from response**

### Step 6: List Tasks
```
GET http://localhost:8080/projects/<project-id>/tasks
Authorization: Bearer <token>
```

### Step 7: Update Task
```
PATCH http://localhost:8080/tasks/<task-id>
Content-Type: application/json
Authorization: Bearer <token>

{
  "status": "in_progress",
  "priority": "medium"
}
```

### Step 8: Get Project with Tasks
```
GET http://localhost:8080/projects/<project-id>
Authorization: Bearer <token>
```

### Step 9: Update Project
```
PATCH http://localhost:8080/projects/<project-id>
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "Updated Project Name"
}
```

### Step 10: Delete Task
```
DELETE http://localhost:8080/tasks/<task-id>
Authorization: Bearer <token>
```

### Step 11: Delete Project
```
DELETE http://localhost:8080/projects/<project-id>
Authorization: Bearer <token>
```

---

## 📊 Error Responses

### 400 Bad Request
```json
{
  "error": "validation failed",
  "fields": {
    "email": "is required",
    "password": "must be at least 6 characters"
  }
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized"
}
```

### 403 Forbidden
```json
{
  "error": "forbidden"
}
```

### 404 Not Found
```json
{
  "error": "not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## 🔐 Authentication Flow

1. **Register/Login** → Get JWT token
2. **Add token to header** → `Authorization: Bearer <token>`
3. **Token expires after 24 hours** → Get new token by logging in again
4. **Invalid token** → Returns 401 Unauthorized

---

## 💾 Test Dummy Data

**User 1:**
```
Name: Alice Johnson
Email: alice@example.com
Password: alice123
```

**User 2:**
```
Name: Bob Smith
Email: bob@example.com
Password: bob123
```

**Projects:**
```
Project 1: Development
Description: Development and coding tasks

Project 2: Design
Description: UI/UX design tasks

Project 3: Testing
Description: QA and testing tasks
```

**Tasks:**
```
Task 1: Write API documentation
Status: todo
Priority: high
Due Date: 2026-04-20

Task 2: Unit testing
Status: in_progress
Priority: medium
Due Date: 2026-04-22

Task 3: Code review
Status: done
Priority: low
Due Date: 2026-04-18
```

---

## 🚀 Quick Test Commands (PowerShell)

Save this as a `.ps1` file and run it:

```powershell
# Variables
$baseUrl = "http://localhost:8080"
$email = "test@example.com"
$password = "test123"
$name = "Test User"

# Register
$registerResp = Invoke-WebRequest -Uri "$baseUrl/auth/register" -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body (@{name=$name; email=$email; password=$password} | ConvertTo-Json) `
  -UseBasicParsing
$token = ($registerResp.Content | ConvertFrom-Json).token
Write-Host "✅ Registered. Token: $($token.Substring(0,20))..."

# Create Project
$projectResp = Invoke-WebRequest -Uri "$baseUrl/projects" -Method POST `
  -Headers @{"Authorization"="Bearer $token"; "Content-Type"="application/json"} `
  -Body (@{name="Test Project"; description="Testing"} | ConvertTo-Json) `
  -UseBasicParsing
$projectId = ($projectResp.Content | ConvertFrom-Json).id
Write-Host "✅ Created project: $projectId"

# Create Task
$taskResp = Invoke-WebRequest -Uri "$baseUrl/projects/$projectId/tasks" -Method POST `
  -Headers @{"Authorization"="Bearer $token"; "Content-Type"="application/json"} `
  -Body (@{title="Test Task"; priority="high"} | ConvertTo-Json) `
  -UseBasicParsing
$taskId = ($taskResp.Content | ConvertFrom-Json).id
Write-Host "✅ Created task: $taskId"

# Get Project
$getResp = Invoke-WebRequest -Uri "$baseUrl/projects/$projectId" -Method GET `
  -Headers @{"Authorization"="Bearer $token"} `
  -UseBasicParsing
Write-Host "✅ Retrieved project with $(($getResp.Content | ConvertFrom-Json).tasks.Count) tasks"
```

---

## ✅ All Endpoints Status

| Endpoint | Method | Status | Last Tested |
|----------|--------|--------|-------------|
| /auth/register | POST | ✅ Working | 2026-04-12 |
| /auth/login | POST | ✅ Working | 2026-04-12 |
| /projects | GET | ✅ Working | 2026-04-12 |
| /projects | POST | ✅ Working | 2026-04-12 |
| /projects/{id} | GET | ✅ Working | 2026-04-12 |
| /projects/{id} | PATCH | ✅ Working | 2026-04-12 |
| /projects/{id} | DELETE | ✅ Working | 2026-04-12 |
| /projects/{id}/tasks | GET | ✅ Working | 2026-04-12 |
| /projects/{id}/tasks | POST | ✅ Working | 2026-04-12 |
| /tasks/{id} | PATCH | ✅ Working | 2026-04-12 |
| /tasks/{id} | DELETE | ✅ Working | 2026-04-12 |

---

## 📞 Support

For issues:
1. Check the error response message
2. Verify authentication token is valid
3. Ensure IDs are correct (project and task IDs must be valid UUIDs)
4. Check PostgreSQL is running
5. Verify database connection parameters in `.env`

**Backend running on:** `http://localhost:8080`
**Database:** `taskflow_db`
**User:** `taskflow_user`
