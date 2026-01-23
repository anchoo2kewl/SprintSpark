/**
 * useLocalTasks Hook
 * React hook for managing tasks with local-first optimistic updates
 */

import { useState, useEffect } from 'react'
import { useSync } from '../state/SyncContext'
import { api } from '../lib/api'
import type { TaskDocument } from '../lib/db/schema'
import type { UpdateTaskRequest } from '../lib/api'

export function useLocalTasks(projectId: number) {
  const { db, syncService } = useSync()
  const [tasks, setTasks] = useState<TaskDocument[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Subscribe to tasks from local database OR fetch from server
  useEffect(() => {
    if (!db) {
      // Fallback to server-side fetch when RxDB is disabled
      const fetchTasks = async () => {
        try {
          setLoading(true)
          const serverTasks = await api.getTasks(projectId)
          setTasks(serverTasks as any) // Type conversion for now
          setLoading(false)
        } catch (err) {
          console.error('[useLocalTasks] Server fetch error:', err)
          setError(err instanceof Error ? err.message : 'Failed to load tasks')
          setLoading(false)
        }
      }
      fetchTasks()
      return
    }

    setLoading(true)

    const subscription = db.tasks
      .find({
        selector: {
          project_id: projectId,
          _deleted: { $ne: true },
        },
        sort: [{ created_at: 'desc' }],
      })
      .$.subscribe({
        next: (docs) => {
          setTasks(docs.map(doc => doc.toJSON()))
          setLoading(false)
          setError(null)
        },
        error: (err) => {
          console.error('[useLocalTasks] Subscription error:', err)
          setError(err instanceof Error ? err.message : 'Failed to load tasks')
          setLoading(false)
        },
      })

    return () => {
      subscription.unsubscribe()
    }
  }, [db, projectId])

  /**
   * Create a new task with optimistic update
   */
  const createTask = async (data: {
    title: string
    description?: string
    status?: 'todo' | 'in_progress' | 'done'
    priority?: 'low' | 'medium' | 'high' | 'urgent'
    assignee_id?: number
    due_date?: string
  }) => {
    // Fallback to server API when RxDB is disabled
    if (!db || !syncService) {
      const newTask = await api.createTask(projectId, data)
      setTasks(prev => [newTask as any, ...prev])
      return
    }

    // Generate temporary ID (will be replaced by server ID after sync)
    const tempId = Date.now()

    const newTask: TaskDocument = {
      id: tempId,
      project_id: projectId,
      title: data.title,
      description: data.description,
      status: data.status || 'todo',
      priority: data.priority,
      assignee_id: data.assignee_id,
      due_date: data.due_date,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      _pendingSync: true,
      _lastSynced: Date.now(),
    }

    try {
      // 1. Optimistic update: Insert into local DB immediately
      await db.tasks.insert(newTask)

      // 2. Queue for background sync
      await syncService.queueOperation('create', 'tasks', tempId, {
        ...data,
        project_id: projectId,
      })

      console.log('[useLocalTasks] Task created optimistically:', tempId)
    } catch (error) {
      console.error('[useLocalTasks] Failed to create task:', error)
      throw error
    }
  }

  /**
   * Update a task with optimistic update
   */
  const updateTask = async (taskId: number, updates: UpdateTaskRequest) => {
    // Fallback to server API when RxDB is disabled
    if (!db || !syncService) {
      const updatedTask = await api.updateTask(taskId, updates)
      setTasks(prev => prev.map(t => t.id === taskId ? updatedTask as any : t))
      return
    }

    try {
      // Find the task
      const task = await db.tasks.findOne({ selector: { id: taskId } }).exec()
      if (!task) throw new Error('Task not found')

      // 1. Optimistic update: Update local DB immediately
      await task.update({
        $set: {
          ...updates,
          updated_at: new Date().toISOString(),
          _pendingSync: true,
        },
      })

      // 2. Queue for background sync
      await syncService.queueOperation('update', 'tasks', taskId, updates)

      console.log('[useLocalTasks] Task updated optimistically:', taskId)
    } catch (error) {
      console.error('[useLocalTasks] Failed to update task:', error)
      throw error
    }
  }

  /**
   * Delete a task with optimistic update
   */
  const deleteTask = async (taskId: number) => {
    if (!db || !syncService) throw new Error('Database not initialized')

    try {
      // Find the task
      const task = await db.tasks.findOne({ selector: { id: taskId } }).exec()
      if (!task) throw new Error('Task not found')

      // 1. Optimistic update: Mark as deleted in local DB
      await task.update({
        $set: {
          _deleted: true,
          _pendingSync: true,
        },
      })

      // 2. Queue for background sync
      await syncService.queueOperation('delete', 'tasks', taskId, {})

      console.log('[useLocalTasks] Task deleted optimistically:', taskId)
    } catch (error) {
      console.error('[useLocalTasks] Failed to delete task:', error)
      throw error
    }
  }

  /**
   * Update task status (common operation, optimized)
   */
  const updateTaskStatus = async (
    taskId: number,
    newStatus: 'todo' | 'in_progress' | 'done'
  ) => {
    await updateTask(taskId, { status: newStatus })
  }

  return {
    tasks,
    loading,
    error,
    createTask,
    updateTask,
    deleteTask,
    updateTaskStatus,
  }
}
