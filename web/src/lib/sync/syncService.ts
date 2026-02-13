/**
 * Sync Service
 * Handles full sync between local RxDB and remote API
 */

import { api } from '../api'
import type { TaskAIDatabase } from '../db'
import type { ProjectDocument, TaskDocument, SprintDocument, TagDocument, SyncQueueDocument } from '../db/schema'

export type SyncStatus = 'idle' | 'syncing' | 'synced' | 'error' | 'offline'

export interface SyncState {
  status: SyncStatus
  lastSyncTime: number | null
  error: string | null
  pendingOperations: number
}

export class SyncService {
  private db: TaskAIDatabase
  private syncInterval: number | null = null
  private listeners: Array<(state: SyncState) => void> = []
  private state: SyncState = {
    status: 'idle',
    lastSyncTime: null,
    error: null,
    pendingOperations: 0,
  }

  constructor(db: TaskAIDatabase) {
    this.db = db
  }

  /**
   * Start automatic syncing at specified interval (ms)
   */
  startAutoSync(intervalMs: number = 30000): void {
    this.stopAutoSync()

    // Initial sync
    this.performFullSync()

    // Set up interval
    this.syncInterval = window.setInterval(() => {
      this.performFullSync()
    }, intervalMs)

    console.log(`[Sync] Auto-sync started (interval: ${intervalMs}ms)`)
  }

  /**
   * Stop automatic syncing
   */
  stopAutoSync(): void {
    if (this.syncInterval) {
      clearInterval(this.syncInterval)
      this.syncInterval = null
      console.log('[Sync] Auto-sync stopped')
    }
  }

  /**
   * Subscribe to sync state changes
   */
  subscribe(listener: (state: SyncState) => void): () => void {
    this.listeners.push(listener)
    // Immediately call with current state
    listener(this.state)

    // Return unsubscribe function
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener)
    }
  }

  /**
   * Update sync state and notify listeners
   */
  private updateState(updates: Partial<SyncState>): void {
    this.state = { ...this.state, ...updates }
    this.listeners.forEach(listener => listener(this.state))
  }

  /**
   * Perform full sync with the server
   * Downloads all data and replaces local data
   */
  async performFullSync(): Promise<void> {
    // Check online status
    if (!navigator.onLine) {
      this.updateState({ status: 'offline', error: 'No internet connection' })
      return
    }

    this.updateState({ status: 'syncing', error: null })

    try {
      // Count pending operations
      const pendingCount = await this.db.syncqueue.count().exec()
      this.updateState({ pendingOperations: pendingCount })

      // Process pending operations first
      await this.processSyncQueue()

      // Pull data from server
      await this.pullFromServer()

      // Update state
      this.updateState({
        status: 'synced',
        lastSyncTime: Date.now(),
        error: null,
        pendingOperations: 0,
      })

      console.log('[Sync] Full sync completed successfully')
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown sync error'
      console.error('[Sync] Full sync failed:', errorMessage)
      this.updateState({
        status: 'error',
        error: errorMessage,
      })
    }
  }

  /**
   * Pull all data from server and update local database
   */
  private async pullFromServer(): Promise<void> {
    console.log('[Sync] Pulling data from server...')

    // Fetch all data in parallel
    const [projects, sprints, tags] = await Promise.all([
      api.getProjects(),
      api.getSprints(),
      api.getTags(),
    ])

    // Upsert projects
    for (const project of projects) {
      const doc: ProjectDocument = {
        id: project.id!,
        owner_id: project.owner_id!,
        name: project.name || '',
        description: project.description || undefined,
        created_at: project.created_at || new Date().toISOString(),
        updated_at: project.updated_at || new Date().toISOString(),
        _lastSynced: Date.now(),
      }

      await this.db.projects.upsert(doc)
    }

    // Fetch tasks for each project
    for (const project of projects) {
      if (project.id) {
        const tasks = await api.getTasks(project.id)

        for (const task of tasks) {
          const doc: TaskDocument = {
            id: task.id!,
            project_id: task.project_id!,
            title: task.title || '',
            description: task.description || undefined,
            status: task.status as 'todo' | 'in_progress' | 'done',
            priority: task.priority as 'low' | 'medium' | 'high' | 'urgent' | undefined,
            assignee_id: task.assignee_id || undefined,
            sprint_id: task.sprint_id || undefined,
            due_date: task.due_date || undefined,
            estimated_hours: task.estimated_hours || undefined,
            actual_hours: task.actual_hours || undefined,
            created_at: task.created_at || new Date().toISOString(),
            updated_at: task.updated_at || new Date().toISOString(),
            _lastSynced: Date.now(),
            _pendingSync: false,
          }

          await this.db.tasks.upsert(doc)
        }
      }
    }

    // Upsert sprints
    for (const sprint of sprints) {
      const doc: SprintDocument = {
        id: sprint.id!,
        user_id: sprint.user_id!,
        name: sprint.name!,
        goal: sprint.goal,
        start_date: sprint.start_date,
        end_date: sprint.end_date,
        status: sprint.status as 'planned' | 'active' | 'completed',
        created_at: sprint.created_at || new Date().toISOString(),
        updated_at: sprint.updated_at || new Date().toISOString(),
        _lastSynced: Date.now(),
      }

      await this.db.sprints.upsert(doc)
    }

    // Upsert tags
    for (const tag of tags) {
      const doc: TagDocument = {
        id: tag.id!,
        user_id: tag.user_id!,
        name: tag.name!,
        color: tag.color!,
        created_at: tag.created_at || new Date().toISOString(),
        _lastSynced: Date.now(),
      }

      await this.db.tags.upsert(doc)
    }

    console.log('[Sync] Data pulled successfully:', {
      projects: projects.length,
      sprints: sprints.length,
      tags: tags.length,
    })
  }

  /**
   * Process pending operations in the sync queue
   * Sends local changes to the server
   */
  private async processSyncQueue(): Promise<void> {
    const queue = await this.db.syncqueue
      .find()
      .sort({ timestamp: 'asc' })
      .exec()

    console.log(`[Sync] Processing ${queue.length} pending operations`)

    for (const item of queue) {
      try {
        await this.executeSyncOperation(item.toJSON())
        await item.remove()
      } catch (error) {
        console.error('[Sync] Failed to process queue item:', item.id, error)

        // Increment retry count
        await item.update({
          $inc: { retryCount: 1 },
          $set: { error: error instanceof Error ? error.message : 'Unknown error' },
        })

        // Remove if too many retries
        if (item.retryCount >= 3) {
          console.error('[Sync] Max retries reached, removing item:', item.id)
          await item.remove()
        }
      }
    }
  }

  /**
   * Execute a single sync operation
   */
  private async executeSyncOperation(item: SyncQueueDocument): Promise<void> {
    const { operation, collection, documentId, data } = item

    switch (collection) {
      case 'projects':
        if (operation === 'create') {
          await api.createProject(data as { name: string; description?: string })
        } else if (operation === 'update') {
          await api.updateProject(documentId, data)
        } else if (operation === 'delete') {
          await api.deleteProject(documentId)
        }
        break

      case 'tasks':
        if (operation === 'create') {
          await api.createTask(data.project_id as number, data as { title: string })
        } else if (operation === 'update') {
          await api.updateTask(documentId, data)
        } else if (operation === 'delete') {
          await api.deleteTask(documentId)
        }
        break

      case 'sprints':
        if (operation === 'create') {
          await api.createSprint(data)
        } else if (operation === 'update') {
          await api.updateSprint(documentId, data)
        } else if (operation === 'delete') {
          await api.deleteSprint(documentId)
        }
        break

      case 'tags':
        if (operation === 'create') {
          await api.createTag(data)
        } else if (operation === 'update') {
          await api.updateTag(documentId, data)
        } else if (operation === 'delete') {
          await api.deleteTag(documentId)
        }
        break
    }
  }

  /**
   * Queue an operation for background sync
   */
  async queueOperation(
    operation: 'create' | 'update' | 'delete',
    collection: 'projects' | 'tasks' | 'sprints' | 'tags',
    documentId: number,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    data: Record<string, any>
  ): Promise<void> {
    await this.db.syncqueue.insert({
      id: `${collection}_${operation}_${documentId}_${Date.now()}`,
      operation,
      collection,
      documentId,
      data,
      timestamp: Date.now(),
      retryCount: 0,
    })

    // Update pending count
    const pendingCount = await this.db.syncqueue.count().exec()
    this.updateState({ pendingOperations: pendingCount })

    console.log('[Sync] Queued operation:', { operation, collection, documentId })
  }

  /**
   * Get current sync state
   */
  getState(): SyncState {
    return this.state
  }
}
