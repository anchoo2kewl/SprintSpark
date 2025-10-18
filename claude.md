> **Mission:** Build SprintSpark incrementally with production-quality code that always runs and passes tests.
> **Philosophy:** Small, perfect commits > large, broken features

---

## 🎯 Project Identity

**SprintSpark** - A lightweight project management system  
**Stack:** Go + SQLite + React + TypeScript  
**Quality Bar:** Every commit deployable to production

---

## 📁 Critical Paths

```
/api/
  cmd/api/main.go         # Entry point
  internal/
    api/                  # HTTP layer
    db/                   # Database layer
    config/              # Configuration
  data/sprintspark.db    # SQLite (gitignored)

/web/
  src/
    lib/api.ts          # API client
    routes/             # Page components
    components/         # Shared UI
    state/              # Global state

/.github/workflows/ci.yml
```

---

## 🚦 Before You Code - STOP Protocol

**S**ummarize - State the task in one sentence  
**T**hink - List affected files and dependencies  
**O**utline - Break into 3-5 atomic steps  
**P**redict - Identify which tests need updating

Example:
```
Task: "Add rate limiting to login"
S: Implement token bucket rate limiting on /api/auth/login
T: Files: handlers.go, middleware.go, handlers_test.go
O: 1) Add rate limit middleware 2) Apply to login 3) Add tests
P: Update auth tests, add rate limit specific tests
```

---

## ⚡ Coding Standards

### Go Backend
```go
// ✅ ALWAYS
- Prepared statements for SQL
- Context with timeouts
- Structured logging (no secrets)
- Error wrapping with context
- Table-driven tests

// ❌ NEVER
- Raw SQL concatenation
- Panic in handlers
- Log passwords/tokens
- Ignore errors
- Global mutable state
```

### React Frontend
```typescript
// ✅ ALWAYS
- TypeScript strict mode
- Loading/error/empty states
- Optimistic updates
- Accessibility (ARIA)
- Error boundaries

// ❌ NEVER
- Any type
- Inline styles
- Direct DOM manipulation
- Unhandled promises
- Console.log in production
```

---

## 🔒 Security Checklist

Every feature MUST:
- [ ] Validate all inputs
- [ ] Check authorization (user owns resource)
- [ ] Use prepared statements
- [ ] Hash sensitive data (bcrypt for passwords)
- [ ] Set appropriate timeouts
- [ ] Rate limit where needed
- [ ] Never log secrets

---

## 📡 API Contract

### Authentication Required Endpoints
```
Bearer token required for all except /api/auth/* and /api/openapi

POST   /api/projects              Create project
GET    /api/projects?page=&limit= List projects
GET    /api/projects/:id          Get project
PATCH  /api/projects/:id          Update project
DELETE /api/projects/:id          Delete project

POST   /api/projects/:id/tasks    Create task
GET    /api/projects/:id/tasks    List tasks (?query= for search)
PATCH  /api/tasks/:id             Update task
DELETE /api/tasks/:id             Delete task
```

### Error Response Format
```json
{
  "error": "Human readable message",
  "code": "machine_readable_code"
}
```

---

## ✅ Definition of Done

A feature is DONE when:

1. **Code Quality**
   - [ ] Passes linting (`make lint`)
   - [ ] Passes tests (`make test`)
   - [ ] No commented code
   - [ ] Clear variable names

2. **Testing**
   - [ ] Unit tests added/updated
   - [ ] E2E test covers happy path
   - [ ] Error cases tested
   - [ ] Edge cases handled

3. **Documentation**
   - [ ] OpenAPI updated (if API changed)
   - [ ] README accurate
   - [ ] Comments on complex logic

4. **Manual Verification**
   - [ ] Feature works locally
   - [ ] No console errors
   - [ ] Graceful error handling
   - [ ] Accessibility checked

---

## 🧪 Test Philosophy

**Coverage Target:** 80% for critical paths

**Go Tests:**
```go
func TestHandlerName(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**React Tests:**
- User-centric (what user sees/does)
- No implementation details
- Mock at network boundary only

---

## 🎭 Agent Behavioral Contract

### When Asked to Build
1. **Plan** - Show 3-5 bullet plan
2. **Code** - Provide complete files
3. **Test** - Include test updates
4. **Verify** - Give exact commands to run

### When Unsure
- Ask ONE clarifying question
- Otherwise, choose the simpler option
- Add TODO comment for ambiguity

### When to Refuse
- Adding secrets to code
- Removing tests without replacement
- Making breaking schema changes
- Creating 500+ line commits

---

## 🚀 Speed Optimizations

### Database
- Index foreign keys and search columns
- Use transactions for multi-step operations
- Batch inserts when possible
- Connection pooling configured

### API
- Pagination defaults (limit=20, max=100)
- Request timeout (30s)
- Body size limit (1MB)
- Gzip compression enabled

### Frontend
- Code splitting by route
- Debounced search (300ms)
- Virtual scrolling for long lists
- Optimistic updates everywhere

---

## 🛠️ Common Tasks Reference

### Add New Endpoint
1. Define handler in `internal/api/handlers.go`
2. Add route in router setup
3. Update OpenAPI spec
4. Add handler test
5. Update Playwright test if UI-facing

### Add Database Migration
1. Create `internal/db/migrations/XXX_description.sql`
2. Never modify existing migrations
3. Test rollback plan
4. Update seed data if needed

### Add UI Feature
1. Create component with TypeScript
2. Add to route or parent
3. Update API client if needed
4. Add loading/error states
5. Update Playwright test

---

## 📋 Quick Commands

```bash
# Development
make dev                 # Start everything
make test               # Run all tests
make lint               # Check code quality
make fmt                # Format code

# Database
make migrate            # Run migrations
make db-reset          # Fresh database

# Frontend
npm run dev            # Start dev server
npm run build          # Production build
npm run test           # Run tests
npx playwright test    # E2E tests

# Docker
docker-compose up      # Start all services
docker-compose down -v # Stop and cleanup
```

---

## 🚨 Emergency Procedures

### If Build Breaks
1. Check recent commits: `git log --oneline -5`
2. Revert last commit: `git revert HEAD`
3. Run tests: `make test`
4. Fix forward if simple, otherwise stay reverted

### If Database Corrupted
1. Stop application
2. Backup current: `cp data/sprintspark.db data/backup.db`
3. Reset: `rm data/sprintspark.db && make migrate`
4. Restore from backup if needed

### If Tests Fail
1. Run single test: `go test -run TestName`
2. Check for race conditions: `go test -race`
3. Verify database state: `sqlite3 data/sprintspark.db`
4. Clear test cache: `go clean -testcache`

---

## 📝 Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: feat, fix, docs, style, refactor, test, chore  
Scope: api, web, db, auth, tasks, projects  

Example:
```
feat(api): add rate limiting to auth endpoints

- Implement token bucket algorithm
- Configure via RATE_LIMIT env var
- Add tests for limit exceeded

Closes #123
```

---

## 🎯 North Star Metrics

Every change should improve:
1. **Reliability** - Uptime, error rate
2. **Performance** - Response time, throughput
3. **Security** - Vulnerability count, auth strength
4. **Developer Experience** - Setup time, test speed
5. **User Experience** - Load time, error clarity

---

## 🤝 Working Agreement

I will:
- Keep changes small and focused
- Write tests for new code
- Handle errors gracefully
- Document complex logic
- Never break the build

You (human) will:
- Provide clear requirements
- Run verification commands
- Report issues with full context
- Review changes before deploying

Together we build reliable software, one perfect commit at a time. 🚀
