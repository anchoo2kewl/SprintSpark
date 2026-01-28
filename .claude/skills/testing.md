# Testing SprintSpark

Comprehensive testing guide for SprintSpark application.

## Quick Test Commands

```bash
# Run all tests
./script/server test

# Run API tests only
./script/server test-api

# Run frontend tests
./script/server test-web

# Run E2E tests
./script/server test-e2e
```

## API Testing

### Run All API Tests

```bash
cd api
go test -v ./...
```

### Run Specific Test

```bash
cd api
go test -v -run TestHandleCreateAPIKey ./internal/api
```

### Run Tests with Coverage

```bash
cd api
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Specific Package

```bash
cd api
go test -v ./internal/api
go test -v ./internal/db
go test -v ./internal/auth
```

## Frontend Testing

### Run Unit Tests

```bash
cd web
npm test
```

### Run E2E Tests

```bash
cd web
npx playwright test
```

### Run E2E Tests in UI Mode

```bash
cd web
npx playwright test --ui
```

### Run Specific E2E Test

```bash
cd web
npx playwright test tests/auth.spec.ts
```

### Debug E2E Tests

```bash
cd web
npx playwright test --debug
```

## Manual API Testing

### Authentication

```bash
# Signup
curl -X POST https://sprintspark.biswas.me/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Login
curl -X POST https://sprintspark.biswas.me/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Get current user
TOKEN="your_jwt_token"
curl -X GET https://sprintspark.biswas.me/api/me \
  -H "Authorization: Bearer $TOKEN"
```

### API Key Testing

```bash
# Create API key
curl -X POST https://sprintspark.biswas.me/api/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Test Key","expires_in":90}'

# List API keys
curl -X GET https://sprintspark.biswas.me/api/api-keys \
  -H "Authorization: Bearer $TOKEN"

# Use API key
API_KEY="your_api_key"
curl -X GET https://sprintspark.biswas.me/api/me \
  -H "Authorization: ApiKey $API_KEY"

# Delete API key
curl -X DELETE https://sprintspark.biswas.me/api/api-keys/1 \
  -H "Authorization: Bearer $TOKEN"
```

### Projects Testing

```bash
# Create project
curl -X POST https://sprintspark.biswas.me/api/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Test Project","description":"Test description"}'

# List projects
curl -X GET https://sprintspark.biswas.me/api/projects \
  -H "Authorization: Bearer $TOKEN"

# Get project
curl -X GET https://sprintspark.biswas.me/api/projects/1 \
  -H "Authorization: Bearer $TOKEN"

# Update project
curl -X PATCH https://sprintspark.biswas.me/api/projects/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Updated Name"}'

# Delete project
curl -X DELETE https://sprintspark.biswas.me/api/projects/1 \
  -H "Authorization: Bearer $TOKEN"
```

### Tasks Testing

```bash
# Create task
curl -X POST https://sprintspark.biswas.me/api/projects/1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Test Task","description":"Test description","status":"todo"}'

# List tasks
curl -X GET https://sprintspark.biswas.me/api/projects/1/tasks \
  -H "Authorization: Bearer $TOKEN"

# Update task
curl -X PATCH https://sprintspark.biswas.me/api/tasks/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"status":"done"}'

# Delete task
curl -X DELETE https://sprintspark.biswas.me/api/tasks/1 \
  -H "Authorization: Bearer $TOKEN"
```

## Load Testing

### Simple Load Test

```bash
# Install hey (HTTP load generator)
# macOS: brew install hey
# Linux: go install github.com/rakyll/hey@latest

# Test API health endpoint
hey -n 1000 -c 10 https://sprintspark.biswas.me/api/health

# Test authenticated endpoint
hey -n 1000 -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  https://sprintspark.biswas.me/api/projects
```

### Rate Limiting Test

```bash
# Test rate limiting (should get 429 after 20 requests/min on auth endpoints)
for i in {1..25}; do
  curl -X POST https://sprintspark.biswas.me/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"wrong"}' \
    -w "\nRequest $i: HTTP %{http_code}\n" \
    -o /dev/null
  sleep 1
done
```

## Test Coverage Goals

- **API**: 80%+ coverage for critical paths
- **Frontend**: User-centric testing (not implementation)
- **E2E**: Happy path coverage for main features

## Test Data

### Demo Users

```
demo.user1@sprintspark.app / DemoPass123!
demo.user2@sprintspark.app / DemoPass223!
demo.user3@sprintspark.app / DemoPass323!
```

### Populate Demo Data

```bash
cd script
API_URL=http://localhost:8083 TESTUSER_EMAIL=test@example.com TESTUSER_PASSWORD=test1234 ./populate_demo_data.sh
```

## CI/CD Testing

Tests run automatically on:
- Git push (local pre-push hook)
- Pull requests (GitHub Actions)
- Deployment (pre-deployment checks)

## Test Output

### Successful Test Run

```
=== Running All Tests ===

[INFO] Running API Tests
✅ API Tests - Passed (80% coverage)

[INFO] Running Frontend Tests
✅ Frontend Tests - Passed

[INFO] Running Playwright E2E Tests
✅ E2E Tests - Passed

✅ All critical tests passed!
```

### Failed Test Run

```
=== Running All Tests ===

[INFO] Running API Tests
❌ API Tests - Failed
  FAIL: TestHandleCreateAPIKey

❌ Some tests failed
```

## Debugging Tests

### Debug Go Tests

```bash
cd api
go test -v -run TestFailingTest ./internal/api
```

### Debug with Print Statements

```go
func TestSomething(t *testing.T) {
    t.Logf("Debug: value = %v", value)
    // test code
}
```

### Debug Playwright Tests

```bash
cd web
PWDEBUG=1 npx playwright test
```

## Best Practices

✅ **DO:**
- Write tests for new features
- Run tests before committing
- Test edge cases and error scenarios
- Use table-driven tests in Go
- Mock at network boundary only

❌ **DON'T:**
- Skip tests to "save time"
- Test implementation details
- Write flaky tests
- Ignore test failures
- Commit without running tests
