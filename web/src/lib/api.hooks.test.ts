import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import {
  useApi,
  useProjects,
  useProject,
  useTasks,
  useCreateProject,
  useUpdateProject,
  useDeleteProject,
  useCreateTask,
  useUpdateTask,
  useDeleteTask,
} from './api.hooks'
import { api } from './api'
import type { Project, Task } from './api'

// Mock the api module
vi.mock('./api', () => {
  const mockApi = {
    getProjects: vi.fn(),
    getProject: vi.fn(),
    getTasks: vi.fn(),
    createProject: vi.fn(),
    updateProject: vi.fn(),
    deleteProject: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    deleteTask: vi.fn(),
  }
  return {
    api: mockApi,
    apiClient: mockApi,
    default: mockApi,
  }
})

const mockedApi = vi.mocked(api)

const fakeProject: Project = {
  id: 1,
  name: 'Test Project',
  description: 'A test project',
  user_id: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const fakeTask: Task = {
  id: 1,
  project_id: 1,
  title: 'Test Task',
  status: 'todo',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

describe('useApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches data on mount and returns it', async () => {
    const mockData = { message: 'hello' }
    const apiCall = vi.fn().mockResolvedValue(mockData)

    const { result } = renderHook(() => useApi(apiCall, []))

    // Initially loading
    expect(result.current.loading).toBe(true)
    expect(result.current.data).toBeNull()
    expect(result.current.error).toBeNull()

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toEqual(mockData)
    expect(result.current.error).toBeNull()
    expect(apiCall).toHaveBeenCalledTimes(1)
  })

  it('starts in loading state', () => {
    const apiCall = vi.fn().mockReturnValue(new Promise(() => {})) // never resolves

    const { result } = renderHook(() => useApi(apiCall, []))

    expect(result.current.loading).toBe(true)
    expect(result.current.data).toBeNull()
    expect(result.current.error).toBeNull()
  })

  it('handles errors from the API call', async () => {
    const apiCall = vi.fn().mockRejectedValue(new Error('Network failure'))

    const { result } = renderHook(() => useApi(apiCall, []))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toBeNull()
    expect(result.current.error).toBe('Network failure')
  })

  it('handles non-Error thrown values', async () => {
    const apiCall = vi.fn().mockRejectedValue('string error')

    const { result } = renderHook(() => useApi(apiCall, []))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBe('An error occurred')
  })

  it('refetches when dependencies change', async () => {
    const apiCall = vi.fn()
      .mockResolvedValueOnce({ id: 1 })
      .mockResolvedValueOnce({ id: 2 })

    let dep = 1
    const { result, rerender } = renderHook(() => useApi(apiCall, [dep]))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    expect(result.current.data).toEqual({ id: 1 })

    // Change dependency
    dep = 2
    rerender()

    await waitFor(() => {
      expect(result.current.data).toEqual({ id: 2 })
    })

    expect(apiCall).toHaveBeenCalledTimes(2)
  })

  it('refetch function triggers a new fetch', async () => {
    const apiCall = vi.fn()
      .mockResolvedValueOnce({ count: 1 })
      .mockResolvedValueOnce({ count: 2 })

    const { result } = renderHook(() => useApi(apiCall, []))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    expect(result.current.data).toEqual({ count: 1 })

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.data).toEqual({ count: 2 })
    expect(apiCall).toHaveBeenCalledTimes(2)
  })

  it('sets loading back to true during refetch', async () => {
    let resolveSecond: (value: unknown) => void
    const secondPromise = new Promise((resolve) => {
      resolveSecond = resolve
    })

    const apiCall = vi.fn()
      .mockResolvedValueOnce({ first: true })
      .mockReturnValueOnce(secondPromise)

    const { result } = renderHook(() => useApi(apiCall, []))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    // Start refetch but don't resolve yet
    act(() => {
      result.current.refetch()
    })

    await waitFor(() => {
      expect(result.current.loading).toBe(true)
    })

    // Resolve the second call
    await act(async () => {
      resolveSecond!({ second: true })
    })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    expect(result.current.data).toEqual({ second: true })
  })
})

describe('useProjects', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls api.getProjects and returns data', async () => {
    const projects = [fakeProject]
    mockedApi.getProjects.mockResolvedValue(projects)

    const { result } = renderHook(() => useProjects())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toEqual(projects)
    expect(mockedApi.getProjects).toHaveBeenCalledTimes(1)
  })

  it('handles getProjects error', async () => {
    mockedApi.getProjects.mockRejectedValue(new Error('Unauthorized'))

    const { result } = renderHook(() => useProjects())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBe('Unauthorized')
    expect(result.current.data).toBeNull()
  })
})

describe('useProject', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches project by id', async () => {
    mockedApi.getProject.mockResolvedValue(fakeProject)

    const { result } = renderHook(() => useProject(1))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toEqual(fakeProject)
    expect(mockedApi.getProject).toHaveBeenCalledWith(1)
  })

  it('returns error when id is undefined', async () => {
    const { result } = renderHook(() => useProject(undefined))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBe('Project ID is required')
    expect(result.current.data).toBeNull()
  })

  it('refetches when id changes', async () => {
    const project1 = { ...fakeProject, id: 1, name: 'Project 1' }
    const project2 = { ...fakeProject, id: 2, name: 'Project 2' }
    mockedApi.getProject
      .mockResolvedValueOnce(project1)
      .mockResolvedValueOnce(project2)

    let id: number | undefined = 1
    const { result, rerender } = renderHook(() => useProject(id))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
    expect(result.current.data).toEqual(project1)

    id = 2
    rerender()

    await waitFor(() => {
      expect(result.current.data).toEqual(project2)
    })
  })
})

describe('useTasks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches tasks by project id', async () => {
    const tasks = [fakeTask]
    mockedApi.getTasks.mockResolvedValue(tasks)

    const { result } = renderHook(() => useTasks(1))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toEqual(tasks)
    expect(mockedApi.getTasks).toHaveBeenCalledWith(1)
  })

  it('returns error when projectId is undefined', async () => {
    const { result } = renderHook(() => useTasks(undefined))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBe('Project ID is required')
    expect(result.current.data).toBeNull()
  })
})

describe('useCreateProject', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns a callable createProject function', () => {
    const { result } = renderHook(() => useCreateProject())

    expect(typeof result.current.createProject).toBe('function')
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('creates a project and returns it', async () => {
    mockedApi.createProject.mockResolvedValue(fakeProject)

    const { result } = renderHook(() => useCreateProject())

    let returnedProject: Project | undefined
    await act(async () => {
      returnedProject = await result.current.createProject({
        name: 'Test Project',
        description: 'A test project',
      })
    })

    expect(returnedProject).toEqual(fakeProject)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(mockedApi.createProject).toHaveBeenCalledWith({
      name: 'Test Project',
      description: 'A test project',
    })
  })

  it('handles loading state during creation', async () => {
    let resolveCreate: (value: Project) => void
    const createPromise = new Promise<Project>((resolve) => {
      resolveCreate = resolve
    })
    mockedApi.createProject.mockReturnValue(createPromise)

    const { result } = renderHook(() => useCreateProject())

    // Start creation
    let createPromiseResult: Promise<Project>
    act(() => {
      createPromiseResult = result.current.createProject({ name: 'New' })
    })

    await waitFor(() => {
      expect(result.current.loading).toBe(true)
    })

    // Resolve
    await act(async () => {
      resolveCreate!(fakeProject)
      await createPromiseResult!
    })

    expect(result.current.loading).toBe(false)
  })

  it('sets error on failure and re-throws', async () => {
    mockedApi.createProject.mockRejectedValue(new Error('Name required'))

    const { result } = renderHook(() => useCreateProject())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.createProject({ name: '' })
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBeInstanceOf(Error)
    expect((caughtError as Error).message).toBe('Name required')
    expect(result.current.error).toBe('Name required')
    expect(result.current.loading).toBe(false)
  })

  it('handles non-Error thrown values', async () => {
    mockedApi.createProject.mockRejectedValue('string error')

    const { result } = renderHook(() => useCreateProject())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.createProject({ name: '' })
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBe('string error')
    expect(result.current.error).toBe('Failed to create project')
  })
})

describe('useUpdateProject', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns a callable updateProject function', () => {
    const { result } = renderHook(() => useUpdateProject())

    expect(typeof result.current.updateProject).toBe('function')
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('updates a project successfully', async () => {
    const updated = { ...fakeProject, name: 'Updated Name' }
    mockedApi.updateProject.mockResolvedValue(updated)

    const { result } = renderHook(() => useUpdateProject())

    let returnedProject: Project | undefined
    await act(async () => {
      returnedProject = await result.current.updateProject(1, {
        name: 'Updated Name',
      })
    })

    expect(returnedProject).toEqual(updated)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(mockedApi.updateProject).toHaveBeenCalledWith(1, {
      name: 'Updated Name',
    })
  })

  it('sets error on failure and re-throws', async () => {
    mockedApi.updateProject.mockRejectedValue(new Error('Not found'))

    const { result } = renderHook(() => useUpdateProject())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.updateProject(999, { name: 'Bad' })
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBeInstanceOf(Error)
    expect((caughtError as Error).message).toBe('Not found')
    expect(result.current.error).toBe('Not found')
    expect(result.current.loading).toBe(false)
  })
})

describe('useDeleteProject', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns a callable deleteProject function', () => {
    const { result } = renderHook(() => useDeleteProject())

    expect(typeof result.current.deleteProject).toBe('function')
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('deletes a project successfully', async () => {
    mockedApi.deleteProject.mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteProject())

    await act(async () => {
      await result.current.deleteProject(1)
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(mockedApi.deleteProject).toHaveBeenCalledWith(1)
  })

  it('sets error on failure and re-throws', async () => {
    mockedApi.deleteProject.mockRejectedValue(new Error('Forbidden'))

    const { result } = renderHook(() => useDeleteProject())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.deleteProject(1)
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBeInstanceOf(Error)
    expect((caughtError as Error).message).toBe('Forbidden')
    expect(result.current.error).toBe('Forbidden')
    expect(result.current.loading).toBe(false)
  })
})

describe('useCreateTask', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns a callable createTask function', () => {
    const { result } = renderHook(() => useCreateTask())

    expect(typeof result.current.createTask).toBe('function')
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('creates a task successfully', async () => {
    mockedApi.createTask.mockResolvedValue(fakeTask)

    const { result } = renderHook(() => useCreateTask())

    let returnedTask: Task | undefined
    await act(async () => {
      returnedTask = await result.current.createTask(1, {
        title: 'Test Task',
        status: 'todo',
      })
    })

    expect(returnedTask).toEqual(fakeTask)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(mockedApi.createTask).toHaveBeenCalledWith(1, {
      title: 'Test Task',
      status: 'todo',
    })
  })

  it('sets error on failure and re-throws', async () => {
    mockedApi.createTask.mockRejectedValue(new Error('Validation error'))

    const { result } = renderHook(() => useCreateTask())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.createTask(1, { title: '' })
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBeInstanceOf(Error)
    expect((caughtError as Error).message).toBe('Validation error')
    expect(result.current.error).toBe('Validation error')
    expect(result.current.loading).toBe(false)
  })
})

describe('useUpdateTask', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns a callable updateTask function', () => {
    const { result } = renderHook(() => useUpdateTask())

    expect(typeof result.current.updateTask).toBe('function')
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('updates a task successfully', async () => {
    const updated = { ...fakeTask, title: 'Updated Task', status: 'done' as const }
    mockedApi.updateTask.mockResolvedValue(updated)

    const { result } = renderHook(() => useUpdateTask())

    let returnedTask: Task | undefined
    await act(async () => {
      returnedTask = await result.current.updateTask(1, {
        title: 'Updated Task',
        status: 'done',
      })
    })

    expect(returnedTask).toEqual(updated)
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(mockedApi.updateTask).toHaveBeenCalledWith(1, {
      title: 'Updated Task',
      status: 'done',
    })
  })

  it('sets error on failure and re-throws', async () => {
    mockedApi.updateTask.mockRejectedValue(new Error('Server error'))

    const { result } = renderHook(() => useUpdateTask())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.updateTask(1, { status: 'done' })
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBeInstanceOf(Error)
    expect((caughtError as Error).message).toBe('Server error')
    expect(result.current.error).toBe('Server error')
    expect(result.current.loading).toBe(false)
  })
})

describe('useDeleteTask', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns a callable deleteTask function', () => {
    const { result } = renderHook(() => useDeleteTask())

    expect(typeof result.current.deleteTask).toBe('function')
    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('deletes a task successfully', async () => {
    mockedApi.deleteTask.mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteTask())

    await act(async () => {
      await result.current.deleteTask(1)
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(mockedApi.deleteTask).toHaveBeenCalledWith(1)
  })

  it('sets error on failure and re-throws', async () => {
    mockedApi.deleteTask.mockRejectedValue(new Error('Not found'))

    const { result } = renderHook(() => useDeleteTask())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.deleteTask(999)
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBeInstanceOf(Error)
    expect((caughtError as Error).message).toBe('Not found')
    expect(result.current.error).toBe('Not found')
    expect(result.current.loading).toBe(false)
  })

  it('handles non-Error thrown values', async () => {
    mockedApi.deleteTask.mockRejectedValue(42)

    const { result } = renderHook(() => useDeleteTask())

    let caughtError: unknown
    await act(async () => {
      try {
        await result.current.deleteTask(1)
      } catch (err) {
        caughtError = err
      }
    })

    expect(caughtError).toBe(42)
    expect(result.current.error).toBe('Failed to delete task')
  })
})
