/**
 * Sync Context
 * React context for managing sync state and database access
 */

import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
// Temporarily disabled due to RxDB schema validation errors
// import { initDatabase, destroyDatabase, type SprintSparkDatabase } from '../lib/db'
import { destroyDatabase, type SprintSparkDatabase } from '../lib/db'
import { SyncService, type SyncState } from '../lib/sync/syncService'
import { useAuth } from './AuthContext'

interface SyncContextValue {
  db: SprintSparkDatabase | null
  syncService: SyncService | null
  syncState: SyncState
  isInitialized: boolean
  initializeSync: () => Promise<void>
  destroySync: () => Promise<void>
  triggerSync: () => Promise<void>
}

const SyncContext = createContext<SyncContextValue | undefined>(undefined)

export function SyncProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const [db, setDb] = useState<SprintSparkDatabase | null>(null)
  const [syncService, setSyncService] = useState<SyncService | null>(null)
  const [syncState, setSyncState] = useState<SyncState>({
    status: 'idle',
    lastSyncTime: null,
    error: null,
    pendingOperations: 0,
  })
  const [isInitialized, setIsInitialized] = useState(false)

  // Initialize database when user logs in
  const initializeSync = async () => {
    if (!user?.id || db) return

    try {
      console.log('[SyncContext] RxDB temporarily disabled - using server-only mode')

      // TODO: Fix RxDB schema validation issue
      // For now, just mark as initialized without local DB
      setIsInitialized(true)
      setSyncState({
        status: 'synced',
        lastSyncTime: Date.now(),
        error: null,
        pendingOperations: 0,
      })
      return

      // DISABLED FOR NOW:
      // console.log('[SyncContext] Initializing database for user:', user.id)
      // const database = await initDatabase(user.id)
      // setDb(database)
      // const service = new SyncService(database)
      // setSyncService(service)
      // service.subscribe(setSyncState)
      // service.startAutoSync(30000)
      // setIsInitialized(true)
      // console.log('[SyncContext] Database initialized successfully')
    } catch (error) {
      console.error('[SyncContext] Failed to initialize database:', error)
      setSyncState({
        status: 'error',
        lastSyncTime: null,
        error: error instanceof Error ? error.message : 'Database initialization failed',
        pendingOperations: 0,
      })
    }
  }

  // Destroy database when user logs out
  const destroySync = async () => {
    if (syncService) {
      syncService.stopAutoSync()
    }

    if (db) {
      await destroyDatabase()
      setDb(null)
      setSyncService(null)
      setIsInitialized(false)
      setSyncState({
        status: 'idle',
        lastSyncTime: null,
        error: null,
        pendingOperations: 0,
      })
      console.log('[SyncContext] Database destroyed')
    }
  }

  // Manual sync trigger
  const triggerSync = async () => {
    if (syncService) {
      await syncService.performFullSync()
    }
  }

  // Auto-initialize when user logs in
  useEffect(() => {
    if (user && !isInitialized) {
      initializeSync()
    } else if (!user && isInitialized) {
      destroySync()
    }
  }, [user?.id])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (syncService) {
        syncService.stopAutoSync()
      }
    }
  }, [syncService])

  const value: SyncContextValue = {
    db,
    syncService,
    syncState,
    isInitialized,
    initializeSync,
    destroySync,
    triggerSync,
  }

  return <SyncContext.Provider value={value}>{children}</SyncContext.Provider>
}

export function useSync() {
  const context = useContext(SyncContext)
  if (context === undefined) {
    throw new Error('useSync must be used within a SyncProvider')
  }
  return context
}
