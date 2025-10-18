# SprintSpark

[![CI](https://github.com/yourusername/SprintSpark/actions/workflows/ci.yml/badge.svg)](https://github.com/yourusername/SprintSpark/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/yourusername/SprintSpark/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/SprintSpark)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/SprintSpark)](https://goreportcard.com/report/github.com/yourusername/SprintSpark)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> A lightweight, production-grade project management system

**Stack:** Go + SQLite + React + TypeScript
**Philosophy:** Small, perfect commits > large, broken features

---

## Features

- **Project Management** - Create, organize, and track projects
- **Task Tracking** - Search, filter, and manage tasks within projects
- **Authentication** - Secure JWT-based auth with bcrypt password hashing
- **RESTful API** - OpenAPI-documented endpoints
- **Modern UI** - React 18 with TypeScript and Tailwind CSS

---

## Quick Start

### Prerequisites

- Go 1.21+
- Node 18+
- Docker & Docker Compose (optional)

### Local Development

```bash
# Clone repository
git clone <repo-url>
cd SprintSpark

# Start backend
cd api
cp .env.example .env
make migrate
make run

# In another terminal, start frontend
cd web
cp .env.example .env
npm install
npm run dev
```

Visit [http://localhost:5173](http://localhost:5173)

### Docker Development

```bash
docker-compose up
```

Visit [http://localhost:5173](http://localhost:5173)

---

## Project Structure

```
SprintSpark/
├── api/                    # Go backend
│   ├── cmd/api/           # Application entry point
│   ├── internal/
│   │   ├── api/          # HTTP handlers
│   │   ├── db/           # Database layer
│   │   └── config/       # Configuration
│   ├── data/             # SQLite database (gitignored)
│   ├── go.mod
│   └── Makefile
│
├── web/                   # React frontend
│   ├── src/
│   │   ├── components/   # Shared UI components
│   │   ├── routes/       # Page components
│   │   ├── lib/          # Utilities (API client)
│   │   └── state/        # Global state management
│   ├── package.json
│   └── vite.config.ts
│
└── .github/workflows/     # CI/CD
```

---

## Available Commands

### Backend (api/)

```bash
make run        # Start development server
make test       # Run tests with coverage
make lint       # Run linters
make fmt        # Format code
make migrate    # Run database migrations
make db-reset   # Reset database
```

### Frontend (web/)

```bash
npm run dev           # Start dev server
npm run build         # Production build
npm run preview       # Preview production build
npm run test          # Run tests
npx playwright test   # Run E2E tests
```

---

## API Documentation

Interactive API docs available at: [http://localhost:8080/api/openapi](http://localhost:8080/api/openapi)

### Authentication

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secure123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secure123"}'
```

### Projects

```bash
# List projects (requires auth)
curl http://localhost:8080/api/projects \
  -H "Authorization: Bearer <token>"

# Create project
curl -X POST http://localhost:8080/api/projects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Project","description":"Project description"}'
```

---

## Testing

```bash
# Backend tests
cd api && make test

# Frontend tests
cd web && npm run test

# E2E tests (requires running app)
cd web && npx playwright test
```

**Coverage target:** 80% for critical paths

---

## Contributing

See [CLAUDE.md](CLAUDE.md) for detailed development guidelines including:

- Coding standards (Go & TypeScript)
- Security checklist
- Definition of Done
- Commit message format
- Testing philosophy

---

## License

MIT

# Webhook configured
Auto-deployment configured ✓
