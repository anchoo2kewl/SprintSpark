# Contributing to SprintSpark

Thank you for your interest in contributing to SprintSpark! This guide will help you get started.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)

---

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the community
- Show empathy towards other contributors

---

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 20+
- Docker & Docker Compose
- Git

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/SprintSpark.git
cd SprintSpark

# Add upstream remote
git remote add upstream https://github.com/anchoo2kewl/SprintSpark.git
```

### Setup Development Environment

```bash
# Start development servers
make dev

# Or manually:
# Terminal 1: API
cd api && make run

# Terminal 2: Web
cd web && npm run dev
```

---

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feat/your-feature-name
```

Branch naming conventions:
- `feat/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `test/` - Test additions/changes
- `refactor/` - Code refactoring
- `chore/` - Maintenance tasks

### 2. Make Your Changes

Follow the [coding standards](#coding-standards) and ensure all tests pass.

### 3. Test Your Changes

```bash
# Backend
cd api
make test
make lint

# Frontend
cd web
npm run test
npm run type-check
npx playwright test
```

### 4. Commit Your Changes

Follow the [commit guidelines](#commit-guidelines).

### 5. Push and Create PR

```bash
git push origin feat/your-feature-name
```

Then open a Pull Request on GitHub.

---

## Coding Standards

### Go Backend

#### Style Guidelines

```go
// ✅ GOOD
func GetUserProjects(ctx context.Context, userID int) ([]Project, error) {
    // Clear function name, proper context, explicit error handling
    projects, err := db.QueryProjects(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query projects: %w", err)
    }
    return projects, nil
}

// ❌ BAD
func get(id int) []Project {
    // Unclear name, no context, no error handling, panic
    projects := db.Query(id)
    if projects == nil {
        panic("failed")
    }
    return projects
}
```

#### Best Practices

**DO:**
- ✅ Use prepared statements for all SQL queries
- ✅ Pass `context.Context` as first parameter
- ✅ Use structured logging (never log secrets)
- ✅ Wrap errors with context: `fmt.Errorf("action failed: %w", err)`
- ✅ Write table-driven tests
- ✅ Use meaningful variable names
- ✅ Add comments for complex logic

**DON'T:**
- ❌ Use raw SQL string concatenation
- ❌ Use `panic()` in HTTP handlers
- ❌ Ignore errors (`_ = someFunc()`)
- ❌ Use global mutable state
- ❌ Log passwords, tokens, or sensitive data

#### Example: HTTP Handler

```go
func (s *Server) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
    // 1. Parse and validate input
    var req CreateProjectRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // 2. Get user from context
    userID := r.Context().Value("user_id").(int)

    // 3. Execute business logic
    project, err := s.db.CreateProject(r.Context(), userID, req.Name, req.Description)
    if err != nil {
        s.logger.Error("failed to create project", "error", err, "user_id", userID)
        s.respondError(w, http.StatusInternalServerError, "Failed to create project")
        return
    }

    // 4. Respond with success
    s.respondJSON(w, http.StatusCreated, project)
}
```

### TypeScript Frontend

#### Style Guidelines

```typescript
// ✅ GOOD
interface CreateProjectData {
  name: string
  description?: string
}

async function createProject(data: CreateProjectData): Promise<Project> {
  const response = await fetch('/api/projects', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })

  if (!response.ok) {
    throw new Error('Failed to create project')
  }

  return response.json()
}

// ❌ BAD
async function create(d: any) {
  const r = await fetch('/api/projects', {
    method: 'POST',
    body: JSON.stringify(d)
  })
  return r.json() // No error handling
}
```

#### Best Practices

**DO:**
- ✅ Use TypeScript strict mode
- ✅ Define interfaces for all data structures
- ✅ Handle loading, error, and empty states in UI
- ✅ Use React hooks properly (dependencies array)
- ✅ Add ARIA attributes for accessibility
- ✅ Use error boundaries for component errors

**DON'T:**
- ❌ Use `any` type
- ❌ Use inline styles (use Tailwind classes)
- ❌ Directly manipulate DOM
- ❌ Leave promises unhandled
- ❌ Use `console.log` in production code

#### Example: React Component

```typescript
export default function ProjectList() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadProjects()
  }, [])

  const loadProjects = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await api.getProjects()
      setProjects(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects')
    } finally {
      setLoading(false)
    }
  }

  if (loading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={error} onRetry={loadProjects} />
  if (projects.length === 0) return <EmptyState />

  return (
    <div className="grid gap-4">
      {projects.map(project => (
        <ProjectCard key={project.id} project={project} />
      ))}
    </div>
  )
}
```

---

## Testing Requirements

### Definition of Done

A feature is **DONE** when:

- [ ] Code passes linting (`make lint` for Go, `npm run type-check` for TS)
- [ ] All tests pass (`make test` for Go, `npm test` for TS)
- [ ] Unit tests added/updated for new code
- [ ] E2E tests cover happy path (if UI-facing)
- [ ] Error cases are tested
- [ ] Edge cases are handled
- [ ] OpenAPI spec updated (if API changed)
- [ ] No commented-out code
- [ ] Clear variable names
- [ ] Complex logic has comments

### Test Coverage Goals

- **Backend:** 80%+ for critical paths (auth, database, handlers)
- **Frontend:** All user journeys covered in E2E tests
- **API:** All endpoints tested in E2E suite

### Writing Tests

**Go (table-driven tests):**

```go
func TestCreateProject(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateProjectRequest
        want    *Project
        wantErr bool
    }{
        {
            name:  "valid project",
            input: CreateProjectRequest{Name: "Test Project"},
            want:  &Project{ID: 1, Name: "Test Project"},
            wantErr: false,
        },
        {
            name:  "empty name",
            input: CreateProjectRequest{Name: ""},
            want:  nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CreateProject(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateProject() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateProject() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Playwright (user-centric):**

```typescript
test('user can create and view project', async ({ page }) => {
  // Signup
  await page.goto('/signup')
  await page.fill('[name="email"]', 'test@example.com')
  await page.fill('[name="password"]', 'SecurePass123!')
  await page.click('button[type="submit"]')

  // Create project
  await page.click('text=New Project')
  await page.fill('[name="name"]', 'My Project')
  await page.click('button[type="submit"]')

  // Verify
  await expect(page.locator('text=My Project')).toBeVisible()
})
```

---

## Commit Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, no logic change)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

### Scopes

- `api` - Backend API changes
- `web` - Frontend changes
- `db` - Database changes
- `auth` - Authentication/authorization
- `tasks` - Task-related features
- `projects` - Project-related features
- `ci` - CI/CD changes

### Examples

```bash
feat(api): add rate limiting to auth endpoints

- Implement token bucket algorithm
- Configure via RATE_LIMIT env var
- Add tests for limit exceeded scenarios

Closes #123
```

```bash
fix(web): resolve infinite loop in project list

The useEffect was missing dependencies, causing constant re-renders.
Added proper dependency array to fix the issue.

Fixes #456
```

### Rules

- Use present tense ("add feature" not "added feature")
- Use imperative mood ("move cursor" not "moves cursor")
- Capitalize first letter of subject
- No period at the end of subject line
- Subject line max 72 characters
- Wrap body at 72 characters
- Reference issues and PRs in footer

---

## Pull Request Process

### Before Submitting

1. **Sync with upstream:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all tests:**
   ```bash
   make test        # In api/
   npm test         # In web/
   npx playwright test  # In web/
   ```

3. **Check code quality:**
   ```bash
   make lint        # In api/
   npm run type-check  # In web/
   ```

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] E2E tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-reviewed code
- [ ] Commented complex logic
- [ ] Updated documentation
- [ ] No console logs or debugging code
- [ ] Tests pass locally
```

### Review Process

1. Automated checks must pass (CI, linting, tests)
2. At least one approval from maintainer
3. No unresolved conversations
4. Branch up to date with main

### After Approval

- Maintainer will merge using "Squash and merge"
- Delete your branch after merge

---

## Project Structure

```
SprintSpark/
├── api/                      # Go backend
│   ├── cmd/api/             # Application entry point
│   │   └── main.go
│   ├── internal/            # Private application code
│   │   ├── api/            # HTTP handlers & server
│   │   ├── auth/           # Authentication logic
│   │   ├── config/         # Configuration
│   │   ├── db/             # Database layer & migrations
│   │   └── e2e/            # E2E tests
│   ├── data/               # SQLite database (gitignored)
│   ├── openapi.yaml        # API specification
│   ├── go.mod              # Go dependencies
│   └── Makefile            # Common tasks
│
├── web/                     # React frontend
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── routes/         # Page components
│   │   ├── lib/           # Utilities & API client
│   │   ├── state/         # Global state (Context API)
│   │   └── main.tsx       # Application entry
│   ├── tests/             # E2E tests
│   ├── package.json       # NPM dependencies
│   └── vite.config.ts     # Vite configuration
│
├── webhook/               # Auto-deployment system
├── .github/workflows/     # CI/CD pipelines
├── docker-compose.yml     # Production deployment
├── README.md             # Project documentation
├── CONTRIBUTING.md       # This file
└── CLAUDE.md            # AI development guide
```

---

## Getting Help

- **Questions?** Open a [Discussion](https://github.com/anchoo2kewl/SprintSpark/discussions)
- **Bug?** Open an [Issue](https://github.com/anchoo2kewl/SprintSpark/issues)
- **Feature Request?** Open an [Issue](https://github.com/anchoo2kewl/SprintSpark/issues) with `feature` label

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to SprintSpark! 🚀**
