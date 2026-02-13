import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useLocalTasks } from './useLocalTasks'

// Mock SyncContext
const mockDb = null as unknown
const mockSyncService = null as unknown

vi.mock('../state/SyncContext', () => ({
  useSync: () => ({
    db: mockDb,
    syncService: mockSyncService,
  }),
}))

// Mock API
const mockGetTasks = vi.fn()
const mockCreateTask = vi.fn()
const mockUpdateTask = vi.fn()

vi.mock('../lib/api', () => ({
  api: {
    getTasks: (...args: unknown[]) => mockGetTasks(...args),
    createTask: (...args: unknown[]) => mockCreateTask(...args),
    updateTask: (...args: unknown[]) => mockUpdateTask(...args),
  },
}))

describe('useLocalTasks (server fallback mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns loading true initially', () => {
    mockGetTasks.mockReturnValue(new Promise(() => {}))
    const { result } = renderHook(() => useLocalTasks(1))
    expect(result.current.loading).toBe(true)
    expect(result.current.tasks).toEqual([])
  })

  it('fetches tasks from server when db is null', async () => {
    const tasks = [
      { id: 1, project_id: 1, title: 'Task 1', status: 'todo', created_at: '2024-01-01', updated_at: '2024-01-01' },
      { id: 2, project_id: 1, title: 'Task 2', status: 'done', created_at: '2024-01-01', updated_at: '2024-01-01' },
    ]
    mockGetTasks.mockResolvedValue(tasks)

    const { result } = renderHook(() => useLocalTasks(1))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.tasks).toHaveLength(2)
    expect(result.current.error).toBeNull()
    expect(mockGetTasks).toHaveBeenCalledWith(1)
  })

  it('sets error on fetch failure', async () => {
    mockGetTasks.mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => useLocalTasks(1))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBe('Network error')
    expect(result.current.tasks).toEqual([])
  })

  it('createTask calls server API and updates local state', async () => {
    mockGetTasks.mockResolvedValue([])
    const newTask = {
      id: 10,
      project_id: 1,
      title: 'New task',
      status: 'todo',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    }
    mockCreateTask.mockResolvedValue(newTask)

    const { result } = renderHook(() => useLocalTasks(1))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    await act(async () => {
      await result.current.createTask({ title: 'New task' })
    })

    expect(mockCreateTask).toHaveBeenCalledWith(1, { title: 'New task' })
    expect(result.current.tasks).toHaveLength(1)
    expect(result.current.tasks[0].title).toBe('New task')
  })

  it('updateTask calls server API and updates local state', async () => {
    const initial = [
      { id: 1, project_id: 1, title: 'Task 1', status: 'todo', created_at: '2024-01-01', updated_at: '2024-01-01' },
    ]
    mockGetTasks.mockResolvedValue(initial)
    const updated = { ...initial[0], status: 'done' }
    mockUpdateTask.mockResolvedValue(updated)

    const { result } = renderHook(() => useLocalTasks(1))

    await waitFor(() => {
      expect(result.current.tasks).toHaveLength(1)
    })

    await act(async () => {
      await result.current.updateTask(1, { status: 'done' })
    })

    expect(mockUpdateTask).toHaveBeenCalledWith(1, { status: 'done' })
    expect(result.current.tasks[0].status).toBe('done')
  })

  it('refetches when projectId changes', async () => {
    mockGetTasks.mockResolvedValue([])

    const { rerender } = renderHook(
      ({ projectId }) => useLocalTasks(projectId),
      { initialProps: { projectId: 1 } }
    )

    await waitFor(() => {
      expect(mockGetTasks).toHaveBeenCalledWith(1)
    })

    rerender({ projectId: 2 })

    await waitFor(() => {
      expect(mockGetTasks).toHaveBeenCalledWith(2)
    })
  })
})
