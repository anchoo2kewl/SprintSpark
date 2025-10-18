# API Usage Examples

Complete guide to using the SprintSpark API.

## Table of Contents

- [Authentication](#authentication)
- [Projects](#projects)
- [Tasks](#tasks)
- [Error Handling](#error-handling)
- [Pagination](#pagination)
- [Rate Limiting](#rate-limiting)

---

## Authentication

### Sign Up

Create a new user account.

**Request:**
```bash
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

**Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-10-18T00:00:00Z"
  }
}
```

**Password Requirements:**
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

### Login

Authenticate with existing credentials.

**Request:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-10-18T00:00:00Z"
  }
}
```

### Get Current User

Retrieve authenticated user information.

**Request:**
```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response (200 OK):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "created_at": "2025-10-18T00:00:00Z"
}
```

---

## Projects

### List Projects

Get all projects for the authenticated user.

**Request:**
```bash
curl http://localhost:8080/api/projects \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**With Pagination:**
```bash
curl "http://localhost:8080/api/projects?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "SprintSpark Development",
    "description": "Building the next-gen project management tool",
    "user_id": 1,
    "created_at": "2025-10-18T00:00:00Z",
    "updated_at": "2025-10-18T00:00:00Z"
  },
  {
    "id": 2,
    "name": "Website Redesign",
    "description": "Modernizing company website",
    "user_id": 1,
    "created_at": "2025-10-18T01:00:00Z",
    "updated_at": "2025-10-18T01:00:00Z"
  }
]
```

### Get Project

Retrieve a specific project by ID.

**Request:**
```bash
curl http://localhost:8080/api/projects/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "SprintSpark Development",
  "description": "Building the next-gen project management tool",
  "user_id": 1,
  "created_at": "2025-10-18T00:00:00Z",
  "updated_at": "2025-10-18T00:00:00Z"
}
```

### Create Project

Create a new project.

**Request:**
```bash
curl -X POST http://localhost:8080/api/projects \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mobile App Development",
    "description": "iOS and Android applications"
  }'
```

**Response (201 Created):**
```json
{
  "id": 3,
  "name": "Mobile App Development",
  "description": "iOS and Android applications",
  "user_id": 1,
  "created_at": "2025-10-18T02:00:00Z",
  "updated_at": "2025-10-18T02:00:00Z"
}
```

### Update Project

Update an existing project.

**Request:**
```bash
curl -X PATCH http://localhost:8080/api/projects/3 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mobile App - iOS First",
    "description": "Starting with iOS development"
  }'
```

**Response (200 OK):**
```json
{
  "id": 3,
  "name": "Mobile App - iOS First",
  "description": "Starting with iOS development",
  "user_id": 1,
  "created_at": "2025-10-18T02:00:00Z",
  "updated_at": "2025-10-18T02:30:00Z"
}
```

### Delete Project

Delete a project and all associated tasks.

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/projects/3 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response (204 No Content)**

---

## Tasks

### List Tasks

Get all tasks for a project.

**Request:**
```bash
curl http://localhost:8080/api/projects/1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**With Search:**
```bash
curl "http://localhost:8080/api/projects/1/tasks?query=authentication" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "title": "Implement JWT authentication",
    "description": "Add secure token-based auth",
    "status": "done",
    "priority": "high",
    "project_id": 1,
    "created_at": "2025-10-18T00:00:00Z",
    "updated_at": "2025-10-18T01:00:00Z"
  },
  {
    "id": 2,
    "title": "Add password reset flow",
    "description": "Email-based password recovery",
    "status": "in_progress",
    "priority": "medium",
    "project_id": 1,
    "created_at": "2025-10-18T00:30:00Z",
    "updated_at": "2025-10-18T01:30:00Z"
  }
]
```

### Create Task

Add a task to a project.

**Request:**
```bash
curl -X POST http://localhost:8080/api/projects/1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Write API documentation",
    "description": "Complete OpenAPI spec and examples",
    "status": "todo",
    "priority": "high"
  }'
```

**Response (201 Created):**
```json
{
  "id": 3,
  "title": "Write API documentation",
  "description": "Complete OpenAPI spec and examples",
  "status": "todo",
  "priority": "high",
  "project_id": 1,
  "created_at": "2025-10-18T02:00:00Z",
  "updated_at": "2025-10-18T02:00:00Z"
}
```

**Valid Status Values:**
- `todo` - Not started
- `in_progress` - Currently working
- `done` - Completed

**Valid Priority Values:**
- `low` - Can wait
- `medium` - Normal priority
- `high` - Important
- `urgent` - Critical

### Update Task

Update task details or status.

**Request:**
```bash
curl -X PATCH http://localhost:8080/api/tasks/3 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_progress",
    "description": "Started writing OpenAPI spec"
  }'
```

**Response (200 OK):**
```json
{
  "id": 3,
  "title": "Write API documentation",
  "description": "Started writing OpenAPI spec",
  "status": "in_progress",
  "priority": "high",
  "project_id": 1,
  "created_at": "2025-10-18T02:00:00Z",
  "updated_at": "2025-10-18T02:15:00Z"
}
```

### Delete Task

Remove a task from a project.

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/tasks/3 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response (204 No Content)**

---

## Error Handling

All errors return a consistent JSON format:

```json
{
  "error": "Human-readable error message",
  "code": "machine_readable_code"
}
```

### Common Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `invalid_request` | Malformed request body |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden` | User doesn't own the resource |
| 404 | `not_found` | Resource doesn't exist |
| 409 | `conflict` | Resource already exists |
| 422 | `validation_error` | Input validation failed |
| 429 | `rate_limit_exceeded` | Too many requests |
| 500 | `internal_error` | Server error |

### Example Error Responses

**Invalid Email (400):**
```json
{
  "error": "Invalid email format",
  "code": "validation_error"
}
```

**Unauthorized (401):**
```json
{
  "error": "Authentication required",
  "code": "unauthorized"
}
```

**Project Not Found (404):**
```json
{
  "error": "Project not found",
  "code": "not_found"
}
```

**Email Already Exists (409):**
```json
{
  "error": "Email already registered",
  "code": "conflict"
}
```

---

## Pagination

List endpoints support pagination via query parameters.

**Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

**Example:**
```bash
curl "http://localhost:8080/api/projects?page=2&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response Headers:**
```
X-Total-Count: 45
X-Page: 2
X-Per-Page: 10
X-Total-Pages: 5
```

---

## Rate Limiting

The API implements rate limiting to prevent abuse.

**Limits:**
- **Authentication endpoints:** 5 requests per minute
- **Other endpoints:** 100 requests per minute

**Response Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1634567890
```

**Rate Limit Exceeded (429):**
```json
{
  "error": "Rate limit exceeded. Please try again later.",
  "code": "rate_limit_exceeded"
}
```

---

## TypeScript/JavaScript Examples

### Using Fetch API

```typescript
const API_BASE = 'http://localhost:8080/api'

// Login
const login = async (email: string, password: string) => {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error)
  }

  const { token, user } = await response.json()
  localStorage.setItem('token', token)
  return user
}

// Create Project (authenticated)
const createProject = async (name: string, description: string) => {
  const token = localStorage.getItem('token')

  const response = await fetch(`${API_BASE}/projects`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ name, description })
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error)
  }

  return response.json()
}

// List Tasks with Search
const searchTasks = async (projectId: number, query: string) => {
  const token = localStorage.getItem('token')
  const url = new URL(`${API_BASE}/projects/${projectId}/tasks`)
  url.searchParams.set('query', query)

  const response = await fetch(url.toString(), {
    headers: { 'Authorization': `Bearer ${token}` }
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error)
  }

  return response.json()
}
```

### Using the Type-Safe Client

```typescript
import { api } from './lib/api'

// All methods are fully typed
const projects = await api.getProjects()
const project = await api.createProject({
  name: 'New Project',
  description: 'Description'
})

const tasks = await api.getTasks(project.id)
const task = await api.createTask(project.id, {
  title: 'New Task',
  status: 'todo',
  priority: 'high'
})
```

---

## Best Practices

1. **Always use HTTPS in production**
2. **Store tokens securely** (HttpOnly cookies or secure storage)
3. **Handle errors gracefully** - Show user-friendly messages
4. **Implement retry logic** for transient errors
5. **Respect rate limits** - Implement exponential backoff
6. **Validate input** before sending to API
7. **Use TypeScript** for type safety

---

**For more examples, see the [frontend source code](../web/src/lib/api.ts).**
