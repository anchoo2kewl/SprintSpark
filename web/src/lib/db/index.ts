/**
 * RxDB Database Initialization
 * Local-first database using IndexedDB
 */

import { createRxDatabase, addRxPlugin } from 'rxdb'
import { getRxStorageDexie } from 'rxdb/plugins/storage-dexie'
import { RxDBDevModePlugin } from 'rxdb/plugins/dev-mode'
import { wrappedValidateAjvStorage } from 'rxdb/plugins/validate-ajv'
import type { RxDatabase, RxCollection } from 'rxdb'

import {
  userSchema,
  projectSchema,
  taskSchema,
  sprintSchema,
  tagSchema,
  syncQueueSchema,
  type UserDocument,
  type ProjectDocument,
  type TaskDocument,
  type SprintDocument,
  type TagDocument,
  type SyncQueueDocument,
} from './schema'

// Add dev mode plugin in development
if (import.meta.env.DEV) {
  addRxPlugin(RxDBDevModePlugin)
}

// Database collections type
export type SprintSparkCollections = {
  users: RxCollection<UserDocument>
  projects: RxCollection<ProjectDocument>
  tasks: RxCollection<TaskDocument>
  sprints: RxCollection<SprintDocument>
  tags: RxCollection<TagDocument>
  syncqueue: RxCollection<SyncQueueDocument>
}

export type SprintSparkDatabase = RxDatabase<SprintSparkCollections>

// Singleton database instance
let dbInstance: SprintSparkDatabase | null = null

/**
 * Initialize the RxDB database
 * Creates IndexedDB with all collections
 */
export async function initDatabase(userId: number): Promise<SprintSparkDatabase> {
  // Return existing instance if already initialized
  if (dbInstance) {
    return dbInstance
  }

  try {
    // Create database with AJV validator for dev mode
    const db = await createRxDatabase<SprintSparkCollections>({
      name: `sprintspark_${userId}`, // Separate DB per user
      storage: wrappedValidateAjvStorage({ storage: getRxStorageDexie() }),
      multiInstance: false, // Single tab support (can enable later)
      eventReduce: true, // Performance optimization
      cleanupPolicy: {
        minimumDeletedTime: 1000 * 60 * 60 * 24 * 7, // Keep deleted docs for 7 days
        minimumCollectionAge: 1000 * 60, // Wait 1 minute before cleanup
        runEach: 1000 * 60 * 5, // Run cleanup every 5 minutes
        awaitReplicationsInSync: true,
        waitForLeadership: true,
      },
    })

    // Add collections
    await db.addCollections({
      users: {
        schema: userSchema,
      },
      projects: {
        schema: projectSchema,
      },
      tasks: {
        schema: taskSchema,
      },
      sprints: {
        schema: sprintSchema,
      },
      tags: {
        schema: tagSchema,
      },
      syncqueue: {
        schema: syncQueueSchema,
      },
    })

    dbInstance = db

    console.log('[RxDB] Database initialized successfully')

    return db
  } catch (error) {
    console.error('[RxDB] Failed to initialize database:', error)
    throw error
  }
}

/**
 * Get the current database instance
 * Throws if database not initialized
 */
export function getDatabase(): SprintSparkDatabase {
  if (!dbInstance) {
    throw new Error('Database not initialized. Call initDatabase() first.')
  }
  return dbInstance
}

/**
 * Destroy the database instance
 * Useful for logout or switching users
 */
export async function destroyDatabase(): Promise<void> {
  if (dbInstance) {
    await dbInstance.remove()
    dbInstance = null
    console.log('[RxDB] Database destroyed')
  }
}

/**
 * Clear all data from the database
 * Keeps the database instance alive
 */
export async function clearDatabase(): Promise<void> {
  const db = getDatabase()

  await Promise.all([
    db.projects.find().remove(),
    db.tasks.find().remove(),
    db.sprints.find().remove(),
    db.tags.find().remove(),
    db.syncqueue.find().remove(),
  ])

  console.log('[RxDB] Database cleared')
}
