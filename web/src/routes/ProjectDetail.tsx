import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { DndContext, DragEndEvent, DragOverlay, DragStartEvent, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { api, Project } from '../lib/api'
import { useLocalTasks } from '../hooks/useLocalTasks'
import type { TaskDocument } from '../lib/db/schema'

export default function ProjectDetail() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [loadingProject, setLoadingProject] = useState(true)
  const [projectError, setProjectError] = useState<string | null>(null)

  // Use local-first tasks hook
  const {
    tasks,
    loading: loadingTasks,
    error: tasksError,
    createTask,
    updateTask,
    updateTaskStatus,
  } = useLocalTasks(Number(projectId))

  // New task modal state
  const [showNewTaskModal, setShowNewTaskModal] = useState(false)
  const [newTaskTitle, setNewTaskTitle] = useState('')
  const [newTaskDescription, setNewTaskDescription] = useState('')
  const [newTaskDueDate, setNewTaskDueDate] = useState('')
  const [creating, setCreating] = useState(false)

  // Task detail modal state
  const [selectedTask, setSelectedTask] = useState<TaskDocument | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [editDescription, setEditDescription] = useState('')
  const [editStatus, setEditStatus] = useState<'todo' | 'in_progress' | 'done'>('todo')
  const [editDueDate, setEditDueDate] = useState('')
  const [updating, setUpdating] = useState(false)

  // Drag and drop state
  const [activeTask, setActiveTask] = useState<TaskDocument | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  )

  // Load project metadata (tasks are handled by useLocalTasks hook)
  useEffect(() => {
    if (projectId) {
      loadProject()
    }
  }, [projectId])

  const loadProject = async () => {
    try {
      setLoadingProject(true)
      setProjectError(null)
      const projectData = await api.getProject(Number(projectId))
      setProject(projectData)
    } catch (err) {
      setProjectError(err instanceof Error ? err.message : 'Failed to load project')
    } finally {
      setLoadingProject(false)
    }
  }

  const handleCreateTask = async () => {
    if (!newTaskTitle.trim() || !projectId) return

    try {
      setCreating(true)
      // Optimistic create - updates UI instantly and syncs in background
      await createTask({
        title: newTaskTitle.trim(),
        description: newTaskDescription.trim() || undefined,
        status: 'todo',
        due_date: newTaskDueDate || undefined,
      })
      setShowNewTaskModal(false)
      setNewTaskTitle('')
      setNewTaskDescription('')
      setNewTaskDueDate('')
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create task')
    } finally {
      setCreating(false)
    }
  }

  const handleTaskClick = (task: TaskDocument) => {
    setSelectedTask(task)
    setEditTitle(task.title || '')
    setEditDescription(task.description || '')
    setEditStatus(task.status as 'todo' | 'in_progress' | 'done')
    setEditDueDate(task.due_date ? task.due_date.split('T')[0] : '')
  }

  const handleUpdateTask = async () => {
    if (!selectedTask || !editTitle.trim() || selectedTask.id === undefined) return

    try {
      setUpdating(true)
      // Optimistic update - updates UI instantly
      await updateTask(selectedTask.id, {
        title: editTitle.trim(),
        description: editDescription.trim() || undefined,
        status: editStatus,
        due_date: editDueDate || null,
      })
      setSelectedTask(null)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update task')
    } finally {
      setUpdating(false)
    }
  }

  const handleDragStart = (event: DragStartEvent) => {
    const task = tasks.find(t => t.id === event.active.id)
    setActiveTask(task || null)
  }

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event
    setActiveTask(null)

    if (!over) return

    const taskId = active.id as number
    const newStatus = over.id as 'todo' | 'in_progress' | 'done'

    const task = tasks.find(t => t.id === taskId)
    if (!task || task.status === newStatus) return

    try {
      // Optimistic update - UI updates instantly, syncs in background
      await updateTaskStatus(taskId, newStatus)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update task status')
    }
  }

  if (loadingProject || loadingTasks) {
    return (
      <div className="p-6 bg-dark-bg-primary">
        <div className="animate-pulse space-y-3">
          <div className="h-6 bg-dark-bg-tertiary/40 rounded w-1/3"></div>
          <div className="h-3 bg-dark-bg-tertiary/30 rounded w-1/2"></div>
          <div className="space-y-2 mt-6">
            <div className="h-16 bg-dark-bg-tertiary/30 rounded"></div>
            <div className="h-16 bg-dark-bg-tertiary/30 rounded"></div>
            <div className="h-16 bg-dark-bg-tertiary/30 rounded"></div>
          </div>
        </div>
      </div>
    )
  }

  if (projectError || tasksError) {
    return (
      <div className="p-6 bg-dark-bg-primary">
        <div className="bg-danger-500/10 border border-danger-500/20 text-danger-400 px-4 py-3 rounded text-sm">
          {projectError || tasksError}
        </div>
      </div>
    )
  }

  const tasksByStatus = {
    todo: tasks.filter((t) => t.status === 'todo'),
    in_progress: tasks.filter((t) => t.status === 'in_progress'),
    done: tasks.filter((t) => t.status === 'done'),
  }

  return (
    <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <div className="h-full flex flex-col bg-dark-bg-primary">
        {/* Project Header */}
        <div className="bg-dark-bg-secondary border-b border-dark-bg-tertiary/20 px-6 py-4">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-lg font-semibold text-dark-text-primary">
                {project?.name}
              </h1>
              {project?.description && (
                <p className="mt-1 text-xs text-dark-text-secondary">{project.description}</p>
              )}
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => navigate(`/app/projects/${projectId}/settings`)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-dark-bg-tertiary/30 hover:bg-dark-bg-tertiary/50 text-dark-text-secondary hover:text-dark-text-primary text-xs font-medium rounded-md transition-colors duration-150"
                title="Project Settings"
              >
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                Settings
              </button>
              <button
                onClick={() => setShowNewTaskModal(true)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary-500 hover:bg-primary-600 text-white text-xs font-medium rounded-md transition-colors duration-150"
              >
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                New Task
              </button>
            </div>
          </div>

          {/* Task Stats */}
          <div className="flex gap-4 mt-3">
            <div className="flex items-center gap-1.5">
              <div className="w-2 h-2 rounded-full bg-dark-text-tertiary"></div>
              <span className="text-xs text-dark-text-secondary">
                {tasksByStatus.todo.length} To Do
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <div className="w-2 h-2 rounded-full bg-primary-400"></div>
              <span className="text-xs text-dark-text-secondary">
                {tasksByStatus.in_progress.length} In Progress
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <div className="w-2 h-2 rounded-full bg-success-400"></div>
              <span className="text-xs text-dark-text-secondary">
                {tasksByStatus.done.length} Done
              </span>
            </div>
          </div>
        </div>

        {/* Tasks Board */}
        <div className="flex-1 overflow-y-auto p-6 bg-dark-bg-primary">
          {tasks.length === 0 ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-center">
                <svg
                  className="mx-auto h-10 w-10 text-dark-text-tertiary"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                  />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-dark-text-primary">No tasks</h3>
                <p className="mt-1 text-xs text-dark-text-secondary">
                  Get started by creating a new task.
                </p>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-6">
              {/* To Do Column */}
              <TaskColumn
                id="todo"
                title="To Do"
                count={tasksByStatus.todo.length}
                tasks={tasksByStatus.todo}
                color="gray"
                projectId={projectId || ''}
                onTaskClick={handleTaskClick}
              />

              {/* In Progress Column */}
              <TaskColumn
                id="in_progress"
                title="In Progress"
                count={tasksByStatus.in_progress.length}
                tasks={tasksByStatus.in_progress}
                color="blue"
                projectId={projectId || ''}
                onTaskClick={handleTaskClick}
              />

              {/* Done Column */}
              <TaskColumn
                id="done"
                title="Done"
                count={tasksByStatus.done.length}
                tasks={tasksByStatus.done}
                color="green"
                projectId={projectId || ''}
                onTaskClick={handleTaskClick}
              />
            </div>
          )}
        </div>

        {/* Drag Overlay */}
        <DragOverlay>
          {activeTask ? (
            <TaskCard
              task={activeTask}
              projectId={projectId || ''}
              isDragging
            />
          ) : null}
        </DragOverlay>

        {/* New Task Modal */}
        {showNewTaskModal && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
            <div className="bg-dark-bg-secondary rounded-lg shadow-linear-lg max-w-md w-full p-5 border border-dark-bg-tertiary/30">
              <h2 className="text-base font-semibold text-dark-text-primary mb-4">Create New Task</h2>

              <div className="space-y-3">
                <div>
                  <label htmlFor="task-title" className="block text-xs font-medium text-dark-text-secondary mb-1">
                    Title *
                  </label>
                  <input
                    id="task-title"
                    type="text"
                    value={newTaskTitle}
                    onChange={(e) => setNewTaskTitle(e.target.value)}
                    className="w-full px-3 py-2 text-sm bg-dark-bg-primary border border-dark-bg-tertiary/30 text-dark-text-primary rounded-md focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                    placeholder="Enter task title"
                    autoFocus
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && newTaskTitle.trim()) {
                        handleCreateTask()
                      }
                    }}
                  />
                </div>

                <div>
                  <label htmlFor="task-description" className="block text-xs font-medium text-dark-text-secondary mb-1">
                    Description
                  </label>
                  <textarea
                    id="task-description"
                    value={newTaskDescription}
                    onChange={(e) => setNewTaskDescription(e.target.value)}
                    rows={3}
                    className="w-full px-3 py-2 text-sm bg-dark-bg-primary border border-dark-bg-tertiary/30 text-dark-text-primary rounded-md focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 resize-none"
                    placeholder="Enter task description (optional)"
                  />
                </div>

                <div>
                  <label htmlFor="task-due-date" className="block text-xs font-medium text-dark-text-secondary mb-1">
                    Due Date
                  </label>
                  <input
                    id="task-due-date"
                    type="date"
                    value={newTaskDueDate}
                    onChange={(e) => setNewTaskDueDate(e.target.value)}
                    className="w-full px-3 py-2 text-sm bg-dark-bg-primary border border-dark-bg-tertiary/30 text-dark-text-primary rounded-md focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>

              <div className="flex gap-2 mt-5">
                <button
                  onClick={() => {
                    setShowNewTaskModal(false)
                    setNewTaskTitle('')
                    setNewTaskDescription('')
                    setNewTaskDueDate('')
                  }}
                  className="flex-1 px-3 py-1.5 text-sm border border-dark-bg-tertiary/30 text-dark-text-secondary rounded-md hover:bg-dark-bg-tertiary/30 transition-colors duration-150"
                  disabled={creating}
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateTask}
                  disabled={!newTaskTitle.trim() || creating}
                  className="flex-1 px-3 py-1.5 text-sm bg-primary-500 text-white rounded-md hover:bg-primary-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-150"
                >
                  {creating ? 'Creating...' : 'Create Task'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Task Detail Modal */}
        {selectedTask && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
            <div className="bg-dark-bg-secondary rounded-lg shadow-linear-lg max-w-md w-full p-5 border border-dark-bg-tertiary/30">
              <h2 className="text-base font-semibold text-dark-text-primary mb-4">Edit Task</h2>

              <div className="space-y-4">
                <div>
                  <label htmlFor="edit-title" className="block text-sm font-medium text-gray-700 mb-1">
                    Title *
                  </label>
                  <input
                    id="edit-title"
                    type="text"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                </div>

                <div>
                  <label htmlFor="edit-description" className="block text-sm font-medium text-gray-700 mb-1">
                    Description
                  </label>
                  <textarea
                    id="edit-description"
                    value={editDescription}
                    onChange={(e) => setEditDescription(e.target.value)}
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
                  />
                </div>

                <div>
                  <label htmlFor="edit-status" className="block text-sm font-medium text-gray-700 mb-1">
                    Status
                  </label>
                  <select
                    id="edit-status"
                    value={editStatus}
                    onChange={(e) => setEditStatus(e.target.value as 'todo' | 'in_progress' | 'done')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  >
                    <option value="todo">To Do</option>
                    <option value="in_progress">In Progress</option>
                    <option value="done">Done</option>
                  </select>
                </div>

                <div>
                  <label htmlFor="edit-due-date" className="block text-sm font-medium text-gray-700 mb-1">
                    Due Date
                  </label>
                  <input
                    id="edit-due-date"
                    type="date"
                    value={editDueDate}
                    onChange={(e) => setEditDueDate(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={() => setSelectedTask(null)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors duration-200"
                  disabled={updating}
                >
                  Cancel
                </button>
                <button
                  onClick={handleUpdateTask}
                  disabled={!editTitle.trim() || updating}
                  className="flex-1 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
                >
                  {updating ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </DndContext>
  )
}

// Helper components
import { useDroppable } from '@dnd-kit/core'
import { useDraggable } from '@dnd-kit/core'

function TaskColumn({ id, title, count, tasks, color, projectId, onTaskClick }: {
  id: string
  title: string
  count: number
  tasks: TaskDocument[]
  color: string
  projectId: string
  onTaskClick: (task: TaskDocument) => void
}) {
  const { setNodeRef, isOver } = useDroppable({ id })

  const colorClasses = {
    gray: 'bg-dark-text-tertiary',
    blue: 'bg-primary-400',
    green: 'bg-success-400',
  }

  return (
    <div ref={setNodeRef} className={`min-h-[200px] ${isOver ? 'bg-dark-bg-tertiary/20 ring-1 ring-primary-500/30 rounded-md' : ''}`}>
      <h3 className="text-xs font-semibold text-dark-text-secondary mb-2 flex items-center gap-1.5">
        <div className={`w-1.5 h-1.5 rounded-full ${colorClasses[color as keyof typeof colorClasses]}`}></div>
        {title} ({count})
      </h3>
      <div className="space-y-2">
        {tasks.map((task) => (
          <DraggableTask
            key={task.id}
            task={task}
            projectId={projectId || ''}
            onTaskClick={onTaskClick}
          />
        ))}
      </div>
    </div>
  )
}

function DraggableTask({ task, projectId, onTaskClick }: {
  task: TaskDocument
  projectId: string
  onTaskClick: (task: TaskDocument) => void
}) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: task.id as number,
  })

  const style = transform ? {
    transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
    opacity: isDragging ? 0.5 : 1,
  } : undefined

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      onClick={() => onTaskClick(task)}
    >
      <TaskCard
        task={task}
        projectId={projectId || ''}
        isDragging={isDragging}
      />
    </div>
  )
}

function TaskCard({ task, projectId, isDragging }: {
  task: TaskDocument
  projectId: string
  isDragging?: boolean
}) {
  const navigate = useNavigate()

  return (
    <div
      onClick={() => navigate(`/app/projects/${projectId}/tasks/${task.id}`)}
      className={`bg-dark-bg-secondary border border-dark-bg-tertiary/30 rounded-md p-3 hover:border-dark-bg-tertiary/50 transition-all duration-150 cursor-pointer ${
        isDragging ? 'shadow-linear-lg rotate-1' : ''
      } ${task.status === 'done' ? 'opacity-60' : ''}`}
    >
      <h4 className="text-sm font-medium text-dark-text-primary hover:text-primary-400 transition-colors">{task.title}</h4>
      {task.assignee_id && (
        <div className="flex items-center gap-1.5 text-xs text-dark-text-tertiary mt-2">
          <div className="w-4 h-4 rounded-full bg-primary-500/10 flex items-center justify-center">
            <svg className="w-2.5 h-2.5 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <span>{task.assignee_name || `User ${task.assignee_id}`}</span>
        </div>
      )}
    </div>
  )
}
