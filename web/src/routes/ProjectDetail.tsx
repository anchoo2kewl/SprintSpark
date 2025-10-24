import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { DndContext, DragEndEvent, DragOverlay, DragStartEvent, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { api, Project, Task } from '../lib/api'

export default function ProjectDetail() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // New task modal state
  const [showNewTaskModal, setShowNewTaskModal] = useState(false)
  const [newTaskTitle, setNewTaskTitle] = useState('')
  const [newTaskDescription, setNewTaskDescription] = useState('')
  const [newTaskDueDate, setNewTaskDueDate] = useState('')
  const [creating, setCreating] = useState(false)

  // Task detail modal state
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [editDescription, setEditDescription] = useState('')
  const [editStatus, setEditStatus] = useState<'todo' | 'in_progress' | 'done'>('todo')
  const [editDueDate, setEditDueDate] = useState('')
  const [updating, setUpdating] = useState(false)

  // Drag and drop state
  const [activeTask, setActiveTask] = useState<Task | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  )

  useEffect(() => {
    if (projectId) {
      loadProjectAndTasks()
    }
  }, [projectId])

  const loadProjectAndTasks = async () => {
    try {
      setLoading(true)
      setError(null)
      const [projectData, tasksData] = await Promise.all([
        api.getProject(Number(projectId)),
        api.getTasks(Number(projectId)),
      ])
      setProject(projectData)
      setTasks(tasksData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load project')
    } finally {
      setLoading(false)
    }
  }

  const handleCreateTask = async () => {
    if (!newTaskTitle.trim() || !projectId) return

    try {
      setCreating(true)
      const task = await api.createTask(Number(projectId), {
        title: newTaskTitle.trim(),
        description: newTaskDescription.trim() || undefined,
        status: 'todo',
        due_date: newTaskDueDate || undefined,
      })
      setTasks([...tasks, task])
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

  const handleTaskClick = (task: Task) => {
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
      const updated = await api.updateTask(selectedTask.id, {
        title: editTitle.trim(),
        description: editDescription.trim() || undefined,
        status: editStatus,
        due_date: editDueDate || null,
      })
      setTasks(tasks.map(t => t.id === updated.id ? updated : t))
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

    // Optimistically update UI
    setTasks(tasks.map(t =>
      t.id === taskId ? { ...t, status: newStatus } : t
    ))

    try {
      const updated = await api.updateTask(taskId, { status: newStatus })
      setTasks(tasks.map(t => t.id === updated.id ? updated : t))
    } catch (err) {
      // Revert on error
      await loadProjectAndTasks()
      alert(err instanceof Error ? err.message : 'Failed to update task status')
    }
  }

  if (loading) {
    return (
      <div className="p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
          <div className="h-4 bg-gray-100 rounded w-1/2"></div>
          <div className="space-y-3 mt-8">
            <div className="h-20 bg-gray-100 rounded"></div>
            <div className="h-20 bg-gray-100 rounded"></div>
            <div className="h-20 bg-gray-100 rounded"></div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-danger-50 border border-danger-200 text-danger-700 px-4 py-3 rounded">
          {error}
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
      <div className="h-full flex flex-col">
        {/* Project Header */}
        <div className="bg-white border-b border-gray-200 px-8 py-6">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">
                {project?.name}
              </h1>
              {project?.description && (
                <p className="mt-1 text-sm text-gray-600">{project.description}</p>
              )}
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={() => navigate(`/app/projects/${projectId}/settings`)}
                className="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-medium rounded-lg transition-colors duration-200"
                title="Project Settings"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                Settings
              </button>
              <button
                onClick={() => setShowNewTaskModal(true)}
                className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg transition-colors duration-200"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                New Task
              </button>
            </div>
          </div>

          {/* Task Stats */}
          <div className="flex gap-6 mt-4">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-gray-400"></div>
              <span className="text-sm text-gray-600">
                {tasksByStatus.todo.length} To Do
              </span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-blue-500"></div>
              <span className="text-sm text-gray-600">
                {tasksByStatus.in_progress.length} In Progress
              </span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
              <span className="text-sm text-gray-600">
                {tasksByStatus.done.length} Done
              </span>
            </div>
          </div>
        </div>

        {/* Tasks Board */}
        <div className="flex-1 overflow-y-auto p-8">
          {tasks.length === 0 ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-center">
                <svg
                  className="mx-auto h-12 w-12 text-gray-400"
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
                <h3 className="mt-2 text-sm font-medium text-gray-900">No tasks</h3>
                <p className="mt-1 text-sm text-gray-500">
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
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Create New Task</h2>

              <div className="space-y-4">
                <div>
                  <label htmlFor="task-title" className="block text-sm font-medium text-gray-700 mb-1">
                    Title *
                  </label>
                  <input
                    id="task-title"
                    type="text"
                    value={newTaskTitle}
                    onChange={(e) => setNewTaskTitle(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
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
                  <label htmlFor="task-description" className="block text-sm font-medium text-gray-700 mb-1">
                    Description
                  </label>
                  <textarea
                    id="task-description"
                    value={newTaskDescription}
                    onChange={(e) => setNewTaskDescription(e.target.value)}
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent resize-none"
                    placeholder="Enter task description (optional)"
                  />
                </div>

                <div>
                  <label htmlFor="task-due-date" className="block text-sm font-medium text-gray-700 mb-1">
                    Due Date
                  </label>
                  <input
                    id="task-due-date"
                    type="date"
                    value={newTaskDueDate}
                    onChange={(e) => setNewTaskDueDate(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={() => {
                    setShowNewTaskModal(false)
                    setNewTaskTitle('')
                    setNewTaskDescription('')
                    setNewTaskDueDate('')
                  }}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors duration-200"
                  disabled={creating}
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateTask}
                  disabled={!newTaskTitle.trim() || creating}
                  className="flex-1 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
                >
                  {creating ? 'Creating...' : 'Create Task'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Task Detail Modal */}
        {selectedTask && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Edit Task</h2>

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
  tasks: Task[]
  color: string
  projectId: string
  onTaskClick: (task: Task) => void
}) {
  const { setNodeRef, isOver } = useDroppable({ id })

  const colorClasses = {
    gray: 'bg-gray-400',
    blue: 'bg-blue-500',
    green: 'bg-green-500',
  }

  return (
    <div ref={setNodeRef} className={`min-h-[200px] ${isOver ? 'bg-gray-50 ring-2 ring-primary-300 rounded-lg' : ''}`}>
      <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${colorClasses[color as keyof typeof colorClasses]}`}></div>
        {title} ({count})
      </h3>
      <div className="space-y-3">
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
  task: Task
  projectId: string
  onTaskClick: (task: Task) => void
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
  task: Task
  projectId: string
  isDragging?: boolean
}) {
  const navigate = useNavigate()

  return (
    <div
      onClick={() => navigate(`/app/projects/${projectId}/tasks/${task.id}`)}
      className={`bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow duration-200 cursor-pointer ${
        isDragging ? 'shadow-xl rotate-2' : ''
      } ${task.status === 'done' ? 'opacity-75' : ''}`}
    >
      <h4 className="font-medium text-gray-900 hover:text-primary-600 mb-2">{task.title}</h4>
      {task.assignee_id && (
        <div className="flex items-center gap-2 text-xs text-gray-600">
          <div className="w-5 h-5 rounded-full bg-primary-100 flex items-center justify-center">
            <svg className="w-3 h-3 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <span>Assigned to user {task.assignee_id}</span>
        </div>
      )}
    </div>
  )
}
