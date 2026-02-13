import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { SyncService, type SyncState } from './syncService'

// Mock the API module
const mockApi = vi.hoisted(() => ({
  getProjects: vi.fn(),
  getTasks: vi.fn(),
  getSprints: vi.fn(),
  getTags: vi.fn(),
  createProject: vi.fn(),
  updateProject: vi.fn(),
  deleteProject: vi.fn(),
  createTask: vi.fn(),
  updateTask: vi.fn(),
  deleteTask: vi.fn(),
  createSprint: vi.fn(),
  updateSprint: vi.fn(),
  deleteSprint: vi.fn(),
  createTag: vi.fn(),
  updateTag: vi.fn(),
  deleteTag: vi.fn(),
}))

vi.mock('../api', () => ({
  api: mockApi,
}))

// Helper to create a mock RxDB database
function createMockDb() {
  const createMockCollection = () => ({
    find: vi.fn().mockReturnValue({
      remove: vi.fn().mockResolvedValue(undefined),
      sort: vi.fn().mockReturnValue({
        exec: vi.fn().mockResolvedValue([]),
      }),
      exec: vi.fn().mockResolvedValue([]),
    }),
    upsert: vi.fn().mockResolvedValue(undefined),
    insert: vi.fn().mockResolvedValue(undefined),
    count: vi.fn().mockReturnValue({
      exec: vi.fn().mockResolvedValue(0),
    }),
  })

  return {
    projects: createMockCollection(),
    tasks: createMockCollection(),
    sprints: createMockCollection(),
    tags: createMockCollection(),
    syncqueue: createMockCollection(),
    users: createMockCollection(),
  }
}

describe('SyncService', () => {
  let db: ReturnType<typeof createMockDb>
  let service: SyncService

  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    db = createMockDb()
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    service = new SyncService(db as any)

    // Default mocks for pullFromServer
    mockApi.getProjects.mockResolvedValue([])
    mockApi.getSprints.mockResolvedValue([])
    mockApi.getTags.mockResolvedValue([])
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('getState', () => {
    it('returns initial idle state', () => {
      const state = service.getState()
      expect(state.status).toBe('idle')
      expect(state.lastSyncTime).toBeNull()
      expect(state.error).toBeNull()
      expect(state.pendingOperations).toBe(0)
    })
  })

  describe('subscribe', () => {
    it('calls listener immediately with current state', () => {
      const listener = vi.fn()
      service.subscribe(listener)
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({ status: 'idle' })
      )
    })

    it('returns unsubscribe function', () => {
      const listener = vi.fn()
      const unsubscribe = service.subscribe(listener)
      expect(typeof unsubscribe).toBe('function')
    })

    it('stops receiving updates after unsubscribe', async () => {
      const listener = vi.fn()
      const unsubscribe = service.subscribe(listener)

      listener.mockClear()
      unsubscribe()

      await service.performFullSync()
      // Should not have been called after unsubscribe
      expect(listener).not.toHaveBeenCalled()
    })
  })

  describe('performFullSync', () => {
    it('sets status to offline when navigator.onLine is false', async () => {
      const originalOnLine = navigator.onLine
      Object.defineProperty(navigator, 'onLine', { value: false, writable: true })

      await service.performFullSync()

      const state = service.getState()
      expect(state.status).toBe('offline')
      expect(state.error).toBe('No internet connection')

      Object.defineProperty(navigator, 'onLine', { value: originalOnLine, writable: true })
    })

    it('transitions through syncing to synced on success', async () => {
      const states: SyncState[] = []
      service.subscribe((state) => states.push({ ...state }))

      await service.performFullSync()

      const statuses = states.map(s => s.status)
      expect(statuses).toContain('syncing')
      expect(statuses[statuses.length - 1]).toBe('synced')
    })

    it('sets error status on failure', async () => {
      mockApi.getProjects.mockRejectedValue(new Error('API down'))

      await service.performFullSync()

      const state = service.getState()
      expect(state.status).toBe('error')
      expect(state.error).toBe('API down')
    })

    it('updates lastSyncTime on success', async () => {
      await service.performFullSync()

      const state = service.getState()
      expect(state.lastSyncTime).not.toBeNull()
    })

    it('pulls projects from server and upserts locally', async () => {
      mockApi.getProjects.mockResolvedValue([
        { id: 1, owner_id: 1, name: 'Project 1', created_at: '2024-01-01', updated_at: '2024-01-01' },
      ])
      mockApi.getTasks.mockResolvedValue([])

      await service.performFullSync()

      expect(db.projects.upsert).toHaveBeenCalledWith(
        expect.objectContaining({ id: 1, name: 'Project 1' })
      )
    })

    it('fetches tasks for each project', async () => {
      mockApi.getProjects.mockResolvedValue([
        { id: 1, owner_id: 1, name: 'P1', created_at: '', updated_at: '' },
        { id: 2, owner_id: 1, name: 'P2', created_at: '', updated_at: '' },
      ])
      mockApi.getTasks.mockResolvedValue([])

      await service.performFullSync()

      expect(mockApi.getTasks).toHaveBeenCalledWith(1)
      expect(mockApi.getTasks).toHaveBeenCalledWith(2)
    })
  })

  describe('queueOperation', () => {
    it('inserts operation into sync queue', async () => {
      await service.queueOperation('create', 'tasks', 1, { title: 'Test' })

      expect(db.syncqueue.insert).toHaveBeenCalledWith(
        expect.objectContaining({
          operation: 'create',
          collection: 'tasks',
          documentId: 1,
          data: { title: 'Test' },
          retryCount: 0,
        })
      )
    })

    it('updates pending operations count', async () => {
      db.syncqueue.count.mockReturnValue({
        exec: vi.fn().mockResolvedValue(3),
      })

      await service.queueOperation('update', 'projects', 1, { name: 'Updated' })

      const state = service.getState()
      expect(state.pendingOperations).toBe(3)
    })

    it('generates unique IDs for queued operations', async () => {
      await service.queueOperation('create', 'tasks', 1, { title: 'A' })
      await service.queueOperation('create', 'tasks', 2, { title: 'B' })

      const calls = db.syncqueue.insert.mock.calls
      expect(calls[0][0].id).not.toBe(calls[1][0].id)
    })
  })

  describe('startAutoSync / stopAutoSync', () => {
    it('starts periodic sync', async () => {
      const syncSpy = vi.spyOn(service, 'performFullSync').mockResolvedValue()

      service.startAutoSync(5000)

      // Initial sync call
      expect(syncSpy).toHaveBeenCalledTimes(1)

      // Advance timers
      vi.advanceTimersByTime(5000)
      expect(syncSpy).toHaveBeenCalledTimes(2)

      vi.advanceTimersByTime(5000)
      expect(syncSpy).toHaveBeenCalledTimes(3)

      service.stopAutoSync()
    })

    it('stops periodic sync', async () => {
      const syncSpy = vi.spyOn(service, 'performFullSync').mockResolvedValue()

      service.startAutoSync(5000)
      service.stopAutoSync()

      syncSpy.mockClear()
      vi.advanceTimersByTime(10000)

      expect(syncSpy).not.toHaveBeenCalled()
    })

    it('calling startAutoSync twice stops previous interval', async () => {
      const syncSpy = vi.spyOn(service, 'performFullSync').mockResolvedValue()

      service.startAutoSync(5000)
      service.startAutoSync(5000)

      // Only 2 initial syncs (one per startAutoSync call), not duplicated intervals
      expect(syncSpy).toHaveBeenCalledTimes(2)

      service.stopAutoSync()
    })
  })
})
