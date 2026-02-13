import { describe, it, expect } from 'vitest'
import {
  userSchema,
  projectSchema,
  taskSchema,
  sprintSchema,
  tagSchema,
  syncQueueSchema,
} from './schema'

describe('RxDB Schemas', () => {
  describe('userSchema', () => {
    it('has correct primary key', () => {
      expect(userSchema.primaryKey).toBe('id')
    })

    it('requires correct fields', () => {
      expect(userSchema.required).toContain('id')
      expect(userSchema.required).toContain('email')
      expect(userSchema.required).toContain('is_admin')
      expect(userSchema.required).toContain('created_at')
      expect(userSchema.required).toContain('updated_at')
    })

    it('has version 0', () => {
      expect(userSchema.version).toBe(0)
    })

    it('defines all expected properties', () => {
      const props = Object.keys(userSchema.properties)
      expect(props).toContain('id')
      expect(props).toContain('email')
      expect(props).toContain('is_admin')
      expect(props).toContain('_lastSynced')
    })
  })

  describe('projectSchema', () => {
    it('has correct primary key', () => {
      expect(projectSchema.primaryKey).toBe('id')
    })

    it('requires correct fields', () => {
      expect(projectSchema.required).toContain('id')
      expect(projectSchema.required).toContain('owner_id')
      expect(projectSchema.required).toContain('name')
      expect(projectSchema.required).toContain('created_at')
      expect(projectSchema.required).toContain('updated_at')
    })

    it('has indexes on owner_id and _lastSynced', () => {
      expect(projectSchema.indexes).toContain('owner_id')
      expect(projectSchema.indexes).toContain('_lastSynced')
    })

    it('includes optional description field', () => {
      expect(projectSchema.properties.description).toBeDefined()
      expect(projectSchema.properties.description.type).toBe('string')
    })

    it('includes client-side _deleted field', () => {
      expect(projectSchema.properties._deleted).toBeDefined()
      expect(projectSchema.properties._deleted.type).toBe('boolean')
    })
  })

  describe('taskSchema', () => {
    it('has correct primary key', () => {
      expect(taskSchema.primaryKey).toBe('id')
    })

    it('requires correct fields', () => {
      expect(taskSchema.required).toEqual(
        expect.arrayContaining(['id', 'project_id', 'title', 'status', 'created_at', 'updated_at'])
      )
    })

    it('defines status enum', () => {
      expect(taskSchema.properties.status.enum).toEqual(['todo', 'in_progress', 'done'])
    })

    it('defines priority enum', () => {
      expect(taskSchema.properties.priority.enum).toEqual(['low', 'medium', 'high', 'urgent'])
    })

    it('has indexes for query performance', () => {
      expect(taskSchema.indexes).toContain('project_id')
      expect(taskSchema.indexes).toContain('status')
      expect(taskSchema.indexes).toContain('assignee_id')
      expect(taskSchema.indexes).toContain('_pendingSync')
    })

    it('includes all task properties', () => {
      const props = Object.keys(taskSchema.properties)
      expect(props).toContain('title')
      expect(props).toContain('description')
      expect(props).toContain('swim_lane_id')
      expect(props).toContain('priority')
      expect(props).toContain('assignee_id')
      expect(props).toContain('sprint_id')
      expect(props).toContain('due_date')
      expect(props).toContain('estimated_hours')
      expect(props).toContain('actual_hours')
    })
  })

  describe('sprintSchema', () => {
    it('has correct primary key', () => {
      expect(sprintSchema.primaryKey).toBe('id')
    })

    it('requires correct fields', () => {
      expect(sprintSchema.required).toEqual(
        expect.arrayContaining(['id', 'user_id', 'name', 'status', 'created_at', 'updated_at'])
      )
    })

    it('defines status enum', () => {
      expect(sprintSchema.properties.status.enum).toEqual(['planned', 'active', 'completed'])
    })

    it('has indexes for query performance', () => {
      expect(sprintSchema.indexes).toContain('user_id')
      expect(sprintSchema.indexes).toContain('status')
    })
  })

  describe('tagSchema', () => {
    it('has correct primary key', () => {
      expect(tagSchema.primaryKey).toBe('id')
    })

    it('requires correct fields', () => {
      expect(tagSchema.required).toEqual(
        expect.arrayContaining(['id', 'user_id', 'name', 'color', 'created_at'])
      )
    })

    it('has indexes', () => {
      expect(tagSchema.indexes).toContain('user_id')
      expect(tagSchema.indexes).toContain('_lastSynced')
    })
  })

  describe('syncQueueSchema', () => {
    it('has string primary key', () => {
      expect(syncQueueSchema.primaryKey).toBe('id')
      expect(syncQueueSchema.properties.id.type).toBe('string')
    })

    it('defines operation enum', () => {
      expect(syncQueueSchema.properties.operation.enum).toEqual(['create', 'update', 'delete'])
    })

    it('defines collection enum', () => {
      expect(syncQueueSchema.properties.collection.enum).toEqual(['projects', 'tasks', 'sprints', 'tags'])
    })

    it('requires all sync queue fields', () => {
      expect(syncQueueSchema.required).toEqual(
        expect.arrayContaining(['id', 'operation', 'collection', 'documentId', 'data', 'timestamp', 'retryCount'])
      )
    })

    it('has indexes on timestamp and collection', () => {
      expect(syncQueueSchema.indexes).toContain('timestamp')
      expect(syncQueueSchema.indexes).toContain('collection')
    })
  })

  describe('Cross-schema consistency', () => {
    it('all schemas are version 0', () => {
      const schemas = [userSchema, projectSchema, taskSchema, sprintSchema, tagSchema, syncQueueSchema]
      schemas.forEach(schema => {
        expect(schema.version).toBe(0)
      })
    })

    it('all schemas are object type', () => {
      const schemas = [userSchema, projectSchema, taskSchema, sprintSchema, tagSchema, syncQueueSchema]
      schemas.forEach(schema => {
        expect(schema.type).toBe('object')
      })
    })
  })
})
