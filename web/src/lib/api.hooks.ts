import { useState, useEffect, useCallback } from 'react'
import { api } from './api'
import type {
  CreateProjectRequest,
  UpdateProjectRequest,
  CreateTaskRequest,
  UpdateTaskRequest
} from './api'

// Generic API hook state
interface ApiState<T> {
  data: T | null
  loading: boolean
  error: string | null
}

// Generic API hook for fetching data
export function useApi<T>(
  apiCall: () => Promise<T>,
  dependencies: any[] = []
): ApiState<T> & { refetch: () => Promise<void> } {
  const [state, setState] = useState<ApiState<T>>({
    data: null,
    loading: true,
    error: null,
  })

  const fetchData = useCallback(async () => {
    setState(prev => ({ ...prev, loading: true, error: null }))
    try {
      const data = await apiCall()
      setState({ data, loading: false, error: null })
    } catch (err) {
      setState({
        data: null,
        loading: false,
        error: err instanceof Error ? err.message : 'An error occurred',
      })
    }
  }, [apiCall])

  useEffect(() => {
    fetchData()
  }, dependencies)

  return { ...state, refetch: fetchData }
}

// Hook for fetching projects
export function useProjects() {
  return useApi(() => api.getProjects(), [])
}

// Hook for fetching a single project
export function useProject(id: number | undefined) {
  return useApi(
    () => {
      if (!id) throw new Error('Project ID is required')
      return api.getProject(id)
    },
    [id]
  )
}

// Hook for fetching tasks
export function useTasks(projectId: number | undefined) {
  return useApi(
    () => {
      if (!projectId) throw new Error('Project ID is required')
      return api.getTasks(projectId)
    },
    [projectId]
  )
}

// Hook for creating a project
export function useCreateProject() {
  const [state, setState] = useState<{
    loading: boolean
    error: string | null
  }>({
    loading: false,
    error: null,
  })

  const createProject = useCallback(async (data: CreateProjectRequest) => {
    setState({ loading: true, error: null })
    try {
      const project = await api.createProject(data)
      setState({ loading: false, error: null })
      return project
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create project'
      setState({ loading: false, error: errorMessage })
      throw err
    }
  }, [])

  return { createProject, ...state }
}

// Hook for updating a project
export function useUpdateProject() {
  const [state, setState] = useState<{
    loading: boolean
    error: string | null
  }>({
    loading: false,
    error: null,
  })

  const updateProject = useCallback(async (id: number, data: UpdateProjectRequest) => {
    setState({ loading: true, error: null })
    try {
      const project = await api.updateProject(id, data)
      setState({ loading: false, error: null })
      return project
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update project'
      setState({ loading: false, error: errorMessage })
      throw err
    }
  }, [])

  return { updateProject, ...state }
}

// Hook for deleting a project
export function useDeleteProject() {
  const [state, setState] = useState<{
    loading: boolean
    error: string | null
  }>({
    loading: false,
    error: null,
  })

  const deleteProject = useCallback(async (id: number) => {
    setState({ loading: true, error: null })
    try {
      await api.deleteProject(id)
      setState({ loading: false, error: null })
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete project'
      setState({ loading: false, error: errorMessage })
      throw err
    }
  }, [])

  return { deleteProject, ...state }
}

// Hook for creating a task
export function useCreateTask() {
  const [state, setState] = useState<{
    loading: boolean
    error: string | null
  }>({
    loading: false,
    error: null,
  })

  const createTask = useCallback(async (projectId: number, data: CreateTaskRequest) => {
    setState({ loading: true, error: null })
    try {
      const task = await api.createTask(projectId, data)
      setState({ loading: false, error: null })
      return task
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create task'
      setState({ loading: false, error: errorMessage })
      throw err
    }
  }, [])

  return { createTask, ...state }
}

// Hook for updating a task
export function useUpdateTask() {
  const [state, setState] = useState<{
    loading: boolean
    error: string | null
  }>({
    loading: false,
    error: null,
  })

  const updateTask = useCallback(async (id: number, data: UpdateTaskRequest) => {
    setState({ loading: true, error: null })
    try {
      const task = await api.updateTask(id, data)
      setState({ loading: false, error: null })
      return task
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update task'
      setState({ loading: false, error: errorMessage })
      throw err
    }
  }, [])

  return { updateTask, ...state }
}

// Hook for deleting a task
export function useDeleteTask() {
  const [state, setState] = useState<{
    loading: boolean
    error: string | null
  }>({
    loading: false,
    error: null,
  })

  const deleteTask = useCallback(async (id: number) => {
    setState({ loading: true, error: null })
    try {
      await api.deleteTask(id)
      setState({ loading: false, error: null })
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete task'
      setState({ loading: false, error: errorMessage })
      throw err
    }
  }, [])

  return { deleteTask, ...state }
}
