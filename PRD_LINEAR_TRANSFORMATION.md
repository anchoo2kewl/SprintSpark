# Product Requirements Document: SprintSpark Linear Transformation

**Version:** 1.0
**Date:** January 22, 2026
**Status:** In Progress
**Owner:** Product Team

---

## Executive Summary

Transform SprintSpark from a traditional client-server project management tool into a modern, local-first, Linear-inspired application with offline capabilities, real-time sync, and a premium dark-themed user experience.

---

## 1. Vision & Objectives

### Vision
Build a fast, offline-capable project management system that matches Linear's polish and user experience while maintaining simplicity and production-quality code.

### Key Objectives
1. **Local-First Architecture** - Enable offline work with automatic background sync
2. **Linear-Inspired UI** - Implement dark gray color palette and premium feel
3. **Developer Experience** - Hot-reload development with sub-second feedback
4. **Production Quality** - Every commit deployable, all features tested

### Success Metrics
- **Performance**: Task list load < 100ms (local-first)
- **Offline Support**: 100% CRUD operations work offline
- **Sync Latency**: Server sync < 500ms when online
- **User Experience**: Dark theme consistency across all pages
- **Developer Velocity**: Changes visible < 2s (hot-reload)

---

## 2. Current State Analysis

### Technology Stack
- **Backend**: Go 1.21+ with SQLite
- **Frontend**: React 18 + TypeScript + Vite
- **Database**: SQLite (server), IndexedDB (client)
- **Deployment**: Docker + reverse proxy
- **Dev Tools**: Air (Go hot-reload), Vite HMR (React)

### Completed Features
✅ Dark theme system with semantic tokens
✅ Authentication (JWT + bcrypt)
✅ CRUD for Projects, Tasks, Sprints, Tags
✅ Hot-reload development mode
✅ Demo data population script
✅ Command palette (Cmd+K)
✅ Server-side API with OpenAPI docs

### Known Issues
❌ RxDB schema validation errors
❌ Local-first sync disabled (temporary fallback to server-only)
❌ No offline indicator in UI
❌ No conflict resolution for concurrent edits

---

## 3. Feature Requirements

### 3.1 Local-First Architecture

#### 3.1.1 Client-Side Database (RxDB + IndexedDB)

**Status**: 🟡 In Progress (Currently Disabled)

**Requirements**:
- Use RxDB v15+ with Dexie storage adapter
- Separate database per user: `sprintspark_{userId}`
- Collections: `tasks`, `projects`, `sprints`, `tags`, `sync_queue`
- Schema validation with proper AJV integration
- Automatic migrations for schema changes

**Schemas**:
```typescript
// Task Schema
{
  id: number (primary)
  title: string (required)
  description: string?
  status: 'todo' | 'in_progress' | 'done'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  project_id: number (indexed)
  assignee_id: number? (indexed)
  due_date: string? (ISO 8601)
  estimated_hours: number?
  actual_hours: number?
  created_at: string (ISO 8601, indexed)
  updated_at: string (ISO 8601, indexed)
  tags: Tag[]
  _deleted: boolean
  _synced: boolean
}

// Sync Queue Schema
{
  id: string (UUID, primary)
  operation: 'create' | 'update' | 'delete'
  collection: string
  doc_id: number
  data: object
  timestamp: number (indexed)
  retries: number
  status: 'pending' | 'syncing' | 'synced' | 'error'
}
```

**Current Blocker**:
```
Error: fieldNames do not match the regex
```

**Resolution Plan**:
1. Update syncQueueSchema to use strict field definitions instead of generic `data: object`
2. Add `wrappedValidateAjvStorage` with proper AJV configuration
3. Test schema validation in development mode
4. Add migration path from server-only to local-first

#### 3.1.2 Optimistic Updates

**Requirements**:
- All CRUD operations update local DB first
- UI updates immediately (< 16ms)
- Background sync to server with retry logic
- Visual indicators for sync status per entity

**User Experience**:
```
User Action → Local DB Update → UI Update (instant)
                      ↓
              Background Sync → Server
                      ↓
              Sync Status Update
```

**Error Handling**:
- Conflict detection (last-write-wins or custom resolution)
- Rollback UI if server rejects change
- Toast notifications for sync errors
- Retry exponential backoff: 1s, 2s, 4s, 8s, 16s (max)

#### 3.1.3 Background Sync Service

**Requirements**:
- Auto-sync every 30 seconds when online
- Pause sync when offline (detect via `navigator.onLine`)
- Process sync queue in order (FIFO)
- Batch sync operations (max 50 per batch)
- WebSocket for real-time updates (future enhancement)

**SyncService API**:
```typescript
class SyncService {
  startAutoSync(intervalMs: number): void
  stopAutoSync(): void
  syncNow(): Promise<SyncResult>
  getSyncStatus(): SyncState
  onSyncStateChange(callback: (state: SyncState) => void): void
}

interface SyncState {
  status: 'offline' | 'syncing' | 'synced' | 'error'
  lastSyncTime: number
  pendingOperations: number
  error: string | null
}
```

### 3.2 Dark Theme System

#### 3.2.1 Color Palette

**Status**: ✅ Complete

**Semantic Tokens** (Tailwind config):
```css
--dark-bg-primary: #0D0D0D      /* Main background */
--dark-bg-secondary: #1A1A1A    /* Cards, modals */
--dark-bg-tertiary: #2D2D2D     /* Hover states, borders */

--dark-text-primary: #EDEDED    /* Main text */
--dark-text-secondary: #A0A0A0  /* Secondary text */
--dark-text-tertiary: #6B6B6B   /* Disabled, placeholders */

--primary-400: #6B9EFF          /* Primary actions (hover) */
--primary-500: #5B8EEF          /* Primary actions */
--primary-600: #4B7EDF          /* Primary actions (pressed) */

--success-400: #5FD68A          /* Success states */
--success-500: #4FC676          /* Success (default) */

--danger-400: #F87171           /* Error/delete (hover) */
--danger-500: #EF4444           /* Error/delete */
```

**Components Styled**:
- ✅ Login / Signup
- ✅ Dashboard
- ✅ Project List
- ✅ Project Detail
- ✅ Task Cards
- ✅ Task Detail
- ✅ Sprints
- ✅ Tags
- ✅ Settings
- ✅ Sidebar Navigation
- ✅ Command Palette
- ✅ Forms (TextInput, Button, Card)

#### 3.2.2 Component Library

**Status**: ✅ Complete

**Dark-Themed Components**:
- `<Card>` - Background with subtle border
- `<Button>` - Primary, secondary, danger, outline variants
- `<TextInput>` - Dark inputs with focus states
- `<FormError>` - Error alerts with danger colors
- Status badges - Success, warning, error states
- Modal overlays - Dark backgrounds with proper contrast

### 3.3 Command Palette (Cmd+K)

**Status**: ✅ Complete

**Features**:
- Quick navigation to Projects, Sprints, Tags, Settings
- Keyboard shortcuts (⌘K on Mac, Ctrl+K on Windows)
- Fuzzy search across entities
- Recent items prioritization
- Close on Escape or click outside

**Future Enhancements**:
- Create tasks/projects from command palette
- Quick actions (mark task done, assign to me)
- Search task content (requires full-text search)

### 3.4 Development Experience

#### 3.4.1 Hot-Reload Development

**Status**: ✅ Complete

**Server Script** (`./script/server local dev`):
```bash
# Backend: Air watches Go files, recompiles on change
# Frontend: Vite HMR for instant React updates
# Logs: Tail both API and web logs in real-time

Commands:
  ./script/server local dev    # Start hot-reload mode
  ./script/server health        # Check API health
  ./script/server deploy        # Build & deploy to production
```

**Performance**:
- Go recompile: ~2-3s
- React HMR: < 100ms
- Full page reload: < 500ms

#### 3.4.2 Demo Data

**Status**: ✅ Complete

**Population Script** (`script/populate_demo_data.sh`):
- 6 users (1 admin, 5 regular)
- 3 projects
- 90 tasks (30 per project)
- 9 sprints (3 per project)
- 12 tags

**Credentials**:
- Admin: `testuser2@example.com` / `test1234`
- Regular: `test@example.com` / `test1234`

---

## 4. User Flows

### 4.1 Offline Task Creation

```
1. User opens app (offline)
2. UI shows "Offline - Changes will sync when online"
3. User creates task "Fix login bug"
4. Task appears instantly in list (optimistic update)
5. Task marked with sync indicator (cloud icon with clock)
6. User goes online
7. Background sync pushes task to server
8. Sync indicator changes to checkmark
9. Toast: "All changes synced"
```

### 4.2 Conflict Resolution (Future)

```
1. User A edits task title offline: "Fix bug" → "Fix critical bug"
2. User B edits same task online: "Fix bug" → "Fix login issue"
3. User A comes online, sync detects conflict
4. Last-write-wins: User A's change overwrites (they edited later)
5. (Future) Custom resolution UI: Show both versions, let user choose
```

### 4.3 Dark Theme Experience

```
1. User logs in → Dark login page
2. Navigate to dashboard → Dark sidebar + dark project cards
3. Click project → Dark task list with subtle borders
4. Open task detail → Dark modal with markdown rendering
5. Edit task → Dark form inputs with proper contrast
6. Go to Settings → Dark 2FA setup UI
7. All transitions smooth, no FOUC (Flash of Unstyled Content)
```

---

## 5. Technical Architecture

### 5.1 Data Flow (Local-First)

```
┌─────────────────┐
│   React UI      │
│  (Optimistic)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐       ┌──────────────┐
│   useLocalTasks │◄──────┤   RxDB       │
│   React Hook    │       │  IndexedDB   │
└────────┬────────┘       └──────┬───────┘
         │                       │
         ▼                       ▼
┌─────────────────┐       ┌──────────────┐
│  SyncService    │──────▶│  Sync Queue  │
│  (Background)   │       │  (Pending)   │
└────────┬────────┘       └──────────────┘
         │
         ▼
┌─────────────────┐
│   API Client    │
│   (Fetch)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Go API        │
│   (SQLite)      │
└─────────────────┘
```

### 5.2 Deployment Architecture

```
┌───────────────────┐
│   Users/Browser   │
└────────┬──────────┘
         │ HTTPS
         ▼
┌───────────────────┐
│  Nginx (Reverse)  │
│  sprintspark.     │
│  biswas.me:443    │
└────────┬──────────┘
         │
         ├─────────────────┐
         │                 │
         ▼                 ▼
┌────────────────┐  ┌────────────────┐
│  Vite Static   │  │   Go API       │
│  (React Build) │  │   :8083        │
│  /app/*        │  │   /api/*       │
└────────────────┘  └────────┬───────┘
                             │
                             ▼
                    ┌────────────────┐
                    │  SQLite DB     │
                    │  sprintspark.db│
                    └────────────────┘
```

### 5.3 File Structure

```
SprintSpark/
├── api/                      # Go backend
│   ├── cmd/api/main.go      # Entry point
│   ├── internal/
│   │   ├── api/             # HTTP handlers
│   │   ├── db/              # Database layer
│   │   └── config/          # Configuration
│   └── data/
│       └── sprintspark.db   # SQLite database
├── web/                      # React frontend
│   ├── src/
│   │   ├── routes/          # Page components
│   │   ├── components/      # Shared UI components
│   │   ├── hooks/           # React hooks (useLocalTasks)
│   │   ├── lib/
│   │   │   ├── api.ts       # API client
│   │   │   ├── db/          # RxDB setup
│   │   │   └── sync/        # Sync service
│   │   └── state/           # Global state (SyncContext)
│   ├── tailwind.config.js   # Dark theme tokens
│   └── index.html
├── script/
│   ├── server               # Dev/deploy management
│   └── populate_demo_data.sh# Demo data generator
├── .github/workflows/ci.yml # CI/CD pipeline
└── CLAUDE.md                # Development guidelines
```

---

## 6. Implementation Phases

### Phase 1: Foundation ✅ COMPLETE
- [x] Dark theme system (Tailwind tokens)
- [x] Component library (Card, Button, TextInput)
- [x] Login/Signup dark styling
- [x] Dashboard dark styling
- [x] Hot-reload development mode
- [x] Demo data population script

### Phase 2: UI Polish ✅ COMPLETE
- [x] Task detail page dark mode
- [x] Sprints page dark mode
- [x] Tags page dark mode
- [x] Settings page dark mode
- [x] Command palette (Cmd+K)
- [x] Sidebar navigation styling
- [x] Status badges and indicators

### Phase 3: Local-First (IN PROGRESS) 🟡
- [ ] Fix RxDB schema validation
- [ ] Implement optimistic updates
- [ ] Background sync service
- [ ] Offline indicator in UI
- [ ] Sync status per entity
- [ ] Error handling & retry logic

### Phase 4: Sync & Offline 🔜 PLANNED
- [ ] Conflict detection
- [ ] Last-write-wins resolution
- [ ] Batch sync operations
- [ ] WebSocket real-time updates (optional)
- [ ] Network status monitoring
- [ ] Sync queue management UI

### Phase 5: Polish & Performance 🔜 PLANNED
- [ ] Virtual scrolling for long lists
- [ ] Skeleton loaders
- [ ] Debounced search (300ms)
- [ ] Code splitting by route
- [ ] Image optimization
- [ ] Lazy loading components

---

## 7. Testing Strategy

### 7.1 Unit Tests (Target: 80% coverage)

**Backend (Go)**:
```go
// Table-driven tests
func TestHandlerName(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        // Test cases
    }
}
```

**Frontend (React Testing Library)**:
- User-centric tests (what user sees/does)
- Mock at network boundary only
- Test accessibility (ARIA)

### 7.2 E2E Tests (Playwright)

**Critical Paths**:
1. Auth flow (signup → login → logout)
2. Task CRUD (create → edit → delete)
3. Offline work (disconnect → create task → reconnect → verify sync)
4. Command palette (Cmd+K → search → navigate)

**Test Data**:
- Use demo data script for consistent state
- Clean DB between test runs

### 7.3 Manual QA Checklist

**Dark Theme**:
- [ ] All pages use dark background
- [ ] Text has minimum 4.5:1 contrast ratio
- [ ] Focus states visible
- [ ] No light theme leaks

**Offline**:
- [ ] Offline indicator visible
- [ ] CRUD works offline
- [ ] Sync happens on reconnect
- [ ] No data loss

**Performance**:
- [ ] Task list loads < 100ms
- [ ] Hot-reload < 2s
- [ ] Page transitions smooth (60 FPS)

---

## 8. Known Issues & Limitations

### 8.1 Current Blockers

#### RxDB Schema Validation Error
**Issue**: `fieldNames do not match the regex`
**Impact**: Local-first sync disabled
**Status**: 🔴 Critical
**Owner**: Engineering
**ETA**: Week of Jan 27, 2026

**Root Cause**:
```typescript
// syncQueueSchema has generic data field
data: {
  type: 'object'  // ❌ RxDB requires explicit field definitions
}
```

**Fix Required**:
```typescript
// Define strict schema or use JSONSchema validator
data: {
  type: 'object',
  properties: {
    title: { type: 'string' },
    description: { type: 'string' },
    // ... all possible fields
  },
  additionalProperties: false
}
```

### 8.2 Technical Debt

1. **Server-Only Fallback**: Current useLocalTasks hook falls back to direct API calls when RxDB fails
2. **No Conflict Resolution**: Last-write-wins not implemented
3. **No Real-Time Updates**: Changes from other users not pushed automatically
4. **No Virtual Scrolling**: Performance degrades with > 1000 tasks
5. **No Full-Text Search**: Task search limited to title/description substring match

### 8.3 Future Enhancements

- **Attachments**: Upload files to tasks
- **Comments**: Thread discussions per task
- **Activity Log**: Audit trail of changes
- **Custom Fields**: User-defined task metadata
- **Email Notifications**: Mention, assignment, due date alerts
- **Mobile App**: React Native port
- **AI Assistant**: Auto-categorize tasks, suggest assignees

---

## 9. Success Criteria

### 9.1 User Acceptance

**Local-First Experience**:
- [ ] User can create 10 tasks offline
- [ ] User sees all tasks instantly (< 100ms)
- [ ] User reconnects, all tasks sync to server
- [ ] No data loss or duplication

**Dark Theme**:
- [ ] All pages use consistent dark palette
- [ ] Text is readable (WCAG AA compliant)
- [ ] No white flashes during navigation

**Developer Experience**:
- [ ] Code changes visible in < 2s (hot-reload)
- [ ] Test suite passes in < 30s
- [ ] Deployment takes < 5 minutes

### 9.2 Performance Benchmarks

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Task List Load | < 100ms | 50ms (server) | ⚠️ Local-first pending |
| Optimistic Update | < 16ms | N/A | ⚠️ Not implemented |
| Background Sync | < 500ms | N/A | ⚠️ Not implemented |
| Hot-Reload (Go) | < 3s | 2s | ✅ |
| Hot-Reload (React) | < 100ms | 50ms | ✅ |
| First Contentful Paint | < 1s | 800ms | ✅ |

### 9.3 Code Quality

- **Linting**: 0 errors, 0 warnings
- **Test Coverage**: > 80% critical paths
- **Bundle Size**: < 500KB (gzipped)
- **Lighthouse Score**: > 90 (Performance, Accessibility)

---

## 10. Rollout Plan

### 10.1 Beta Testing (Week 1-2)

**Participants**: 5 internal users
**Focus Areas**:
- Offline functionality
- Dark theme consistency
- Sync performance

**Feedback Collection**:
- Daily standup check-ins
- Bug tracker (GitHub Issues)
- User satisfaction survey (1-10 scale)

### 10.2 Staged Rollout (Week 3-4)

**Phase 1**: 10% of users (read-only local cache)
**Phase 2**: 25% of users (optimistic updates)
**Phase 3**: 50% of users (background sync)
**Phase 4**: 100% of users (full local-first)

**Rollback Plan**:
- Feature flag: `ENABLE_LOCAL_FIRST=false`
- Fallback to server-only mode
- No data loss (local DB preserved)

### 10.3 Post-Launch Monitoring

**Key Metrics**:
- Sync error rate (target: < 0.1%)
- Offline usage % (expected: 20-30%)
- Sync queue size (target: < 50 pending ops)
- User-reported bugs (target: < 5/week)

**Dashboards**:
- Grafana: API latency, error rates
- Sentry: Client-side exceptions
- Google Analytics: User behavior

---

## 11. Appendix

### 11.1 Related Documents
- `CLAUDE.md` - Development guidelines
- `README.md` - Setup instructions
- `openapi.yaml` - API specification

### 11.2 References
- [Linear Design System](https://linear.app/design)
- [RxDB Documentation](https://rxdb.info/)
- [IndexedDB API](https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API)
- [WCAG Contrast Guidelines](https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html)

### 11.3 Glossary

| Term | Definition |
|------|------------|
| **Local-First** | Architecture where client-side database is source of truth, server is backup |
| **Optimistic Update** | UI updates immediately before server confirms change |
| **RxDB** | Reactive database built on top of IndexedDB with observables |
| **Sync Queue** | Ordered list of pending operations to push to server |
| **Hot-Reload** | Development feature that updates running app without full restart |
| **HMR** | Hot Module Replacement - Vite's instant update mechanism |
| **Semantic Tokens** | CSS variables that represent meaning (e.g., `--primary`) not raw values (e.g., `#5B8EEF`) |

---

**Document Control**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-22 | Product Team | Initial PRD created |

**Approvals**

- [ ] Product Manager
- [ ] Engineering Lead
- [ ] Design Lead
- [ ] QA Lead

---

**Next Steps**

1. **Immediate** (This Week):
   - Fix RxDB schema validation
   - Implement optimistic updates for tasks
   - Add offline indicator to UI

2. **Short-Term** (Next 2 Weeks):
   - Complete background sync service
   - Add conflict resolution
   - E2E tests for offline flow

3. **Long-Term** (Next Month):
   - Performance optimization (virtual scrolling)
   - Mobile-responsive dark theme
   - WebSocket real-time updates

---

*End of Document*
