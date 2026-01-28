/**
 * RxDB Schema Definitions
 * Local-first database schema mirroring the server-side SQLite schema
 */

import type { RxJsonSchema } from 'rxdb'

// User Schema
export type UserDocument = {
  id: number
  email: string
  is_admin: boolean
  created_at: string
  updated_at: string
  // Client-side only fields
  _lastSynced?: number
}

export const userSchema: RxJsonSchema<UserDocument> = {
  version: 0,
  primaryKey: 'id',
  type: 'object',
  properties: {
    id: {
      type: 'number',
      maxLength: 100,
    },
    email: {
      type: 'string',
    },
    is_admin: {
      type: 'boolean',
    },
    created_at: {
      type: 'string',
      format: 'date-time',
    },
    updated_at: {
      type: 'string',
      format: 'date-time',
    },
    _lastSynced: {
      type: 'number',
    },
  },
  required: ['id', 'email', 'is_admin', 'created_at', 'updated_at'],
}

// Project Schema
export type ProjectDocument = {
  id: number
  owner_id: number
  name: string
  description?: string
  created_at: string
  updated_at: string
  // Client-side only fields
  _deleted?: boolean
  _lastSynced?: number
}

export const projectSchema: RxJsonSchema<ProjectDocument> = {
  version: 0,
  primaryKey: 'id',
  type: 'object',
  properties: {
    id: {
      type: 'number',
      maxLength: 100,
    },
    owner_id: {
      type: 'number',
    },
    name: {
      type: 'string',
    },
    description: {
      type: 'string',
    },
    created_at: {
      type: 'string',
      format: 'date-time',
    },
    updated_at: {
      type: 'string',
      format: 'date-time',
    },
    _deleted: {
      type: 'boolean',
    },
    _lastSynced: {
      type: 'number',
    },
  },
  required: ['id', 'owner_id', 'name', 'created_at', 'updated_at'],
  indexes: ['owner_id', '_lastSynced'],
}

// Task (Issue) Schema
export type TaskDocument = {
  id: number
  project_id: number
  title: string
  description?: string
  status: 'todo' | 'in_progress' | 'done'
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  assignee_id?: number
  assignee_name?: string
  sprint_id?: number
  sprint_name?: string
  due_date?: string
  estimated_hours?: number
  actual_hours?: number
  created_at: string
  updated_at: string
  // Client-side only fields
  _deleted?: boolean
  _lastSynced?: number
  _pendingSync?: boolean
}

export const taskSchema: RxJsonSchema<TaskDocument> = {
  version: 0,
  primaryKey: 'id',
  type: 'object',
  properties: {
    id: {
      type: 'number',
      maxLength: 100,
    },
    project_id: {
      type: 'number',
    },
    title: {
      type: 'string',
    },
    description: {
      type: 'string',
    },
    status: {
      type: 'string',
      enum: ['todo', 'in_progress', 'done'],
    },
    priority: {
      type: 'string',
      enum: ['low', 'medium', 'high', 'urgent'],
    },
    assignee_id: {
      type: 'number',
    },
    assignee_name: {
      type: 'string',
    },
    sprint_id: {
      type: 'number',
    },
    sprint_name: {
      type: 'string',
    },
    due_date: {
      type: 'string',
      format: 'date-time',
    },
    estimated_hours: {
      type: 'number',
    },
    actual_hours: {
      type: 'number',
    },
    created_at: {
      type: 'string',
      format: 'date-time',
    },
    updated_at: {
      type: 'string',
      format: 'date-time',
    },
    _deleted: {
      type: 'boolean',
    },
    _lastSynced: {
      type: 'number',
    },
    _pendingSync: {
      type: 'boolean',
    },
  },
  required: ['id', 'project_id', 'title', 'status', 'created_at', 'updated_at'],
  indexes: ['project_id', 'status', 'assignee_id', '_lastSynced', '_pendingSync'],
}

// Sprint (Cycle) Schema
export type SprintDocument = {
  id: number
  user_id: number
  name: string
  goal?: string
  start_date?: string
  end_date?: string
  status: 'planned' | 'active' | 'completed'
  created_at: string
  updated_at: string
  // Client-side only fields
  _deleted?: boolean
  _lastSynced?: number
}

export const sprintSchema: RxJsonSchema<SprintDocument> = {
  version: 0,
  primaryKey: 'id',
  type: 'object',
  properties: {
    id: {
      type: 'number',
      maxLength: 100,
    },
    user_id: {
      type: 'number',
    },
    name: {
      type: 'string',
    },
    goal: {
      type: 'string',
    },
    start_date: {
      type: 'string',
      format: 'date-time',
    },
    end_date: {
      type: 'string',
      format: 'date-time',
    },
    status: {
      type: 'string',
      enum: ['planned', 'active', 'completed'],
    },
    created_at: {
      type: 'string',
      format: 'date-time',
    },
    updated_at: {
      type: 'string',
      format: 'date-time',
    },
    _deleted: {
      type: 'boolean',
    },
    _lastSynced: {
      type: 'number',
    },
  },
  required: ['id', 'user_id', 'name', 'status', 'created_at', 'updated_at'],
  indexes: ['user_id', 'status', '_lastSynced'],
}

// Tag Schema
export type TagDocument = {
  id: number
  user_id: number
  name: string
  color: string
  created_at: string
  // Client-side only fields
  _deleted?: boolean
  _lastSynced?: number
}

export const tagSchema: RxJsonSchema<TagDocument> = {
  version: 0,
  primaryKey: 'id',
  type: 'object',
  properties: {
    id: {
      type: 'number',
      maxLength: 100,
    },
    user_id: {
      type: 'number',
    },
    name: {
      type: 'string',
    },
    color: {
      type: 'string',
    },
    created_at: {
      type: 'string',
      format: 'date-time',
    },
    _deleted: {
      type: 'boolean',
    },
    _lastSynced: {
      type: 'number',
    },
  },
  required: ['id', 'user_id', 'name', 'color', 'created_at'],
  indexes: ['user_id', '_lastSynced'],
}

// Sync Queue Schema (for tracking pending operations)
export type SyncQueueDocument = {
  id: string
  operation: 'create' | 'update' | 'delete'
  collection: 'projects' | 'tasks' | 'sprints' | 'tags'
  documentId: number
  data: any
  timestamp: number
  retryCount: number
  error?: string
}

export const syncQueueSchema: RxJsonSchema<SyncQueueDocument> = {
  version: 0,
  primaryKey: 'id',
  type: 'object',
  properties: {
    id: {
      type: 'string',
      maxLength: 100,
    },
    operation: {
      type: 'string',
      enum: ['create', 'update', 'delete'],
    },
    collection: {
      type: 'string',
      enum: ['projects', 'tasks', 'sprints', 'tags'],
    },
    documentId: {
      type: 'number',
    },
    data: {
      type: 'object',
    },
    timestamp: {
      type: 'number',
    },
    retryCount: {
      type: 'number',
    },
    error: {
      type: 'string',
    },
  },
  required: ['id', 'operation', 'collection', 'documentId', 'data', 'timestamp', 'retryCount'],
  indexes: ['timestamp', 'collection'],
}
