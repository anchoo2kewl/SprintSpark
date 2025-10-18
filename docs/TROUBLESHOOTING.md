# Troubleshooting Guide

Common issues and solutions for SprintSpark.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Database Issues](#database-issues)
- [API Issues](#api-issues)
- [Frontend Issues](#frontend-issues)
- [Docker Issues](#docker-issues)
- [Authentication Issues](#authentication-issues)
- [Performance Issues](#performance-issues)

---

## Installation Issues

### Go Version Mismatch

**Problem:** `go.mod requires go >= 1.24.0`

**Solution:**
```bash
# Check your Go version
go version

# Install Go 1.24+
# macOS
brew upgrade go

# Linux
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
```

### Node Version Mismatch

**Problem:** npm install fails or shows version warnings

**Solution:**
```bash
# Check your Node version
node --version

# Install Node 20+ using nvm
nvm install 20
nvm use 20

# Or update via package manager
brew upgrade node  # macOS
```

### Cannot Find Make

**Problem:** `make: command not found`

**Solution:**
```bash
# macOS
xcode-select --install

# Linux (Ubuntu/Debian)
sudo apt-get install build-essential

# Linux (CentOS/RHEL)
sudo yum groupinstall "Development Tools"
```

---

## Database Issues

### Database Locked Error

**Problem:** `database is locked` when starting API

**Cause:** Multiple processes trying to access SQLite simultaneously or unclean shutdown.

**Solution:**
```bash
# Option 1: Stop all processes
pkill -f sprintspark
docker-compose down

# Option 2: Remove lock files
rm api/data/sprintspark.db-shm
rm api/data/sprintspark.db-wal

# Option 3: Fresh database
rm -rf api/data/sprintspark.db*
cd api && make migrate
```

### Migration Failed

**Problem:** Migrations don't run or fail partway

**Solution:**
```bash
# Check migration status
cd api
sqlite3 data/sprintspark.db "SELECT * FROM schema_migrations;"

# Reset database completely
rm data/sprintspark.db*
make migrate

# Check migrations ran
sqlite3 data/sprintspark.db ".schema"
```

### Cannot Create Database Directory

**Problem:** Permission denied when creating `api/data/`

**Solution:**
```bash
# Create directory with proper permissions
mkdir -p api/data
chmod 755 api/data

# If in Docker, check volume mounts
docker-compose down -v
docker-compose up --build
```

---

## API Issues

### Port Already in Use

**Problem:** `bind: address already in use` on port 8080

**Solution:**
```bash
# Find and kill process using port 8080
lsof -ti:8080 | xargs kill -9

# Or change port in api/.env
echo "PORT=8081" >> api/.env

# Update web/.env to match
echo "VITE_API_URL=http://localhost:8081" >> web/.env
```

### CORS Errors

**Problem:** Browser shows CORS policy errors

**Solution:**
```bash
# Check CORS_ALLOWED_ORIGINS in api/.env
cat api/.env | grep CORS

# Add your frontend URL
echo "CORS_ALLOWED_ORIGINS=http://localhost:5173" >> api/.env

# For production
echo "CORS_ALLOWED_ORIGINS=https://yourdomain.com" >> api/.env

# Restart API
cd api && make run
```

### JWT Token Validation Failed

**Problem:** API returns 401 Unauthorized with valid token

**Causes:**
1. JWT_SECRET changed between token creation and validation
2. Token expired
3. Token malformed

**Solution:**
```bash
# Check JWT_SECRET is consistent
cat api/.env | grep JWT_SECRET

# If changed, users need to re-login
# Tokens are valid for 24 hours by default

# For development, use a consistent secret
echo "JWT_SECRET=dev-secret-dont-use-in-production" >> api/.env
```

### Cannot Connect to API

**Problem:** Frontend can't reach backend

**Solution:**
```bash
# Check API is running
curl http://localhost:8080/api/health

# Check API URL in web/.env
cat web/.env | grep VITE_API_URL

# Should match where API is running
echo "VITE_API_URL=http://localhost:8080" > web/.env

# Restart frontend
cd web && npm run dev
```

---

## Frontend Issues

### TypeScript Errors After API Changes

**Problem:** Type errors in components after updating API

**Solution:**
```bash
cd web

# Regenerate types from OpenAPI spec
npm run generate:types

# Check for errors
npm run type-check

# Clear build cache if needed
rm -rf dist/ node_modules/.vite
npm install
```

### Vite Dev Server Won't Start

**Problem:** `Error: listen EADDRINUSE: address already in use`

**Solution:**
```bash
# Kill process on port 5173
lsof -ti:5173 | xargs kill -9

# Or use different port
# In web/vite.config.ts:
server: {
  port: 3000
}
```

### Module Not Found Errors

**Problem:** Cannot find module '@/components/...'

**Solution:**
```bash
cd web

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install

# Check path aliases in tsconfig.json
cat tsconfig.json | grep paths
```

### Blank Page / White Screen

**Problem:** App loads but shows blank page

**Causes:**
1. JavaScript errors (check console)
2. API not accessible
3. Authentication redirect loop

**Solution:**
```bash
# Check browser console for errors
# Check API is accessible
curl http://localhost:8080/api/health

# Clear localStorage
# In browser console:
localStorage.clear()

# Restart with clean state
cd web
rm -rf dist/
npm run dev
```

---

## Docker Issues

### Docker Build Fails

**Problem:** `ERROR: failed to solve` during build

**Solution:**
```bash
# Clear Docker cache
docker system prune -a

# Rebuild without cache
docker-compose build --no-cache

# Check Docker daemon is running
docker info
```

### Container Exits Immediately

**Problem:** Container starts then immediately stops

**Solution:**
```bash
# Check container logs
docker-compose logs api
docker-compose logs web

# Run in foreground to see errors
docker-compose up

# Check container status
docker-compose ps -a
```

### Cannot Access Container from Host

**Problem:** localhost:8080 or localhost:80 not accessible

**Solution:**
```bash
# Check containers are running
docker-compose ps

# Check port bindings
docker-compose ps

# Verify ports in docker-compose.yml
cat docker-compose.yml | grep ports

# Check firewall isn't blocking
sudo ufw status  # Linux
# Add rule if needed:
sudo ufw allow 8080
```

### Volume Permission Issues

**Problem:** Permission denied when accessing volumes

**Solution:**
```bash
# Stop containers
docker-compose down

# Remove volumes
docker-compose down -v

# Start fresh
docker-compose up --build

# For Linux, may need to set ownership
sudo chown -R $USER:$USER api/data
```

---

## Authentication Issues

### Cannot Sign Up

**Problem:** Signup fails with validation errors

**Password Requirements:**
- Minimum 8 characters
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one number (0-9)
- At least one special character (!@#$%^&*)

**Solution:**
```bash
# Valid password example
SecurePass123!

# Test signup directly
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!"}'
```

### Logged Out After Refresh

**Problem:** User logged out when refreshing page

**Causes:**
1. Token not persisted to localStorage
2. Token expired
3. JWT_SECRET changed

**Solution:**
```bash
# Check token in browser dev tools
localStorage.getItem('token')

# Check token expiry (default 24h)
# In api/internal/config/config.go

# For development, extend expiry
# Or disable auto-logout in frontend
```

### Email Already Exists

**Problem:** Cannot signup with email that should be available

**Cause:** Email exists from previous testing

**Solution:**
```bash
# Check users in database
cd api
sqlite3 data/sprintspark.db "SELECT * FROM users;"

# Delete test user
sqlite3 data/sprintspark.db "DELETE FROM users WHERE email='test@example.com';"

# Or reset entire database
rm data/sprintspark.db*
make migrate
```

---

## Performance Issues

### Slow API Response

**Problem:** API endpoints take too long to respond

**Solution:**
```bash
# Check database indexes
cd api
sqlite3 data/sprintspark.db ".schema"

# Should see indexes on:
# - users(email)
# - projects(user_id)
# - tasks(project_id)

# Analyze slow queries
# Enable query logging in api/internal/db/db.go

# Optimize database
sqlite3 data/sprintspark.db "VACUUM; ANALYZE;"
```

### High Memory Usage

**Problem:** Application using excessive memory

**Solution:**
```bash
# Check container resource usage
docker stats

# Limit container resources in docker-compose.yml:
services:
  api:
    deploy:
      resources:
        limits:
          memory: 512M

# For development, may need to increase Node memory:
export NODE_OPTIONS="--max-old-space-size=4096"
npm run dev
```

### Slow Docker Builds

**Problem:** `docker-compose up --build` takes very long

**Solution:**
```bash
# Use layer caching effectively
# dependencies change less than code

# Build only changed service
docker-compose build api
docker-compose up -d api

# Clean up unused images
docker image prune -a
```

---

## Testing Issues

### E2E Tests Fail

**Problem:** Playwright tests timeout or fail

**Solution:**
```bash
# Install Playwright browsers
cd web
npx playwright install --with-deps

# Run in headed mode to see what's happening
npx playwright test --headed

# Check test database is clean
cd ../api
rm data/sprintspark.db*
make migrate
cd ../web
npx playwright test

# Increase timeout for slow CI
npx playwright test --timeout=60000
```

### Go Tests Fail

**Problem:** Backend tests failing

**Solution:**
```bash
cd api

# Run specific test
go test -run TestCreateProject ./internal/api

# Run with verbose output
go test -v ./...

# Check for race conditions
go test -race ./...

# Clear test cache
go clean -testcache
```

---

## Still Having Issues?

1. **Check Logs:**
   ```bash
   # API logs
   cd api && make run  # See output

   # Docker logs
   docker-compose logs -f

   # Web dev server logs
   cd web && npm run dev
   ```

2. **Verify Environment:**
   ```bash
   go version      # Should be 1.24+
   node --version  # Should be 20+
   docker --version
   ```

3. **Try Fresh Install:**
   ```bash
   # Clean everything
   docker-compose down -v
   rm -rf api/data/ web/node_modules/ web/dist/

   # Reinstall
   cd web && npm install
   cd ../api && go mod download

   # Start fresh
   cd .. && docker-compose up --build
   ```

4. **Get Help:**
   - Open an issue: https://github.com/anchoo2kewl/SprintSpark/issues
   - Include: OS, versions, error messages, logs

---

**Remember:** Most issues are environment-related. When in doubt, try the Docker setup for a clean, isolated environment!
