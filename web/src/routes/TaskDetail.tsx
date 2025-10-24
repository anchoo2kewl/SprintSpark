import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import Button from '../components/ui/Button'
import { apiClient, Task } from '../lib/api'

export default function TaskDetail() {
  const { projectId, taskId } = useParams()
  const navigate = useNavigate()
  const [task, setTask] = useState<Task | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)

  // Edit form state
  const [editTitle, setEditTitle] = useState('')
  const [editDescription, setEditDescription] = useState('')
  const [editStatus, setEditStatus] = useState<'todo' | 'in_progress' | 'done'>('todo')
  const [editPriority, setEditPriority] = useState<'low' | 'medium' | 'high' | 'urgent'>('medium')
  const [editDueDate, setEditDueDate] = useState('')
  const [editAssigneeId, setEditAssigneeId] = useState('')
  const [editEstimatedHours, setEditEstimatedHours] = useState('')
  const [editActualHours, setEditActualHours] = useState('')

  useEffect(() => {
    loadTask()
  }, [projectId, taskId])

  const loadTask = async () => {
    try {
      setLoading(true)
      const tasks = await apiClient.getTasks(Number(projectId))
      const foundTask = tasks.find((t: Task) => t.id === Number(taskId))

      if (foundTask) {
        setTask(foundTask)
        // Initialize edit form
        setEditTitle(foundTask.title || '')
        setEditDescription(foundTask.description || '')
        setEditStatus(foundTask.status || 'todo')
        setEditPriority(foundTask.priority || 'medium')
        setEditDueDate(foundTask.due_date || '')
        setEditAssigneeId(foundTask.assignee_id?.toString() || '')
        setEditEstimatedHours(foundTask.estimated_hours?.toString() || '')
        setEditActualHours(foundTask.actual_hours?.toString() || '')
      } else {
        setError('Task not found')
      }
    } catch (error: any) {
      setError(error.message || 'Failed to load task')
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    if (!task || !editTitle.trim()) return

    try {
      setSaving(true)
      await apiClient.updateTask(task.id!, {
        title: editTitle.trim(),
        description: editDescription.trim() || undefined,
        status: editStatus,
        priority: editPriority,
        due_date: editDueDate || undefined,
        assignee_id: editAssigneeId ? parseInt(editAssigneeId) : undefined,
        estimated_hours: editEstimatedHours ? parseFloat(editEstimatedHours) : undefined,
        actual_hours: editActualHours ? parseFloat(editActualHours) : undefined,
      })

      await loadTask()
      setIsEditing(false)
    } catch (error: any) {
      setError(error.message || 'Failed to update task')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!task || !confirm('Are you sure you want to delete this task?')) return

    try {
      await apiClient.deleteTask(task.id!)
      navigate(`/app/projects/${projectId}`)
    } catch (error: any) {
      setError(error.message || 'Failed to delete task')
    }
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent':
        return 'bg-red-100 text-red-800 border-red-200'
      case 'high':
        return 'bg-orange-100 text-orange-800 border-orange-200'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200'
      case 'low':
        return 'bg-green-100 text-green-800 border-green-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'done':
        return 'bg-green-100 text-green-800 border-green-200'
      case 'in_progress':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      case 'todo':
        return 'bg-gray-100 text-gray-800 border-gray-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'done':
        return 'Done'
      case 'in_progress':
        return 'In Progress'
      case 'todo':
        return 'To Do'
      default:
        return status
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">Loading task...</div>
      </div>
    )
  }

  if (error && !task) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-white border border-gray-200 rounded-lg p-8 text-center">
            <p className="text-red-600 mb-4">{error}</p>
            <Button onClick={() => navigate(`/app/projects/${projectId}`)}>
              Back to Project
            </Button>
          </div>
        </div>
      </div>
    )
  }

  if (!task) return null

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between mb-4">
            <button
              onClick={() => navigate(`/app/projects/${projectId}`)}
              className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900"
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back to project
            </button>
            <div className="flex items-center gap-2">
              {!isEditing ? (
                <>
                  <Button onClick={() => setIsEditing(true)} variant="secondary" size="sm">
                    Edit
                  </Button>
                  <button
                    onClick={handleDelete}
                    className="px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  >
                    Delete
                  </button>
                </>
              ) : (
                <>
                  <Button onClick={handleSave} size="sm" disabled={saving}>
                    {saving ? 'Saving...' : 'Save'}
                  </Button>
                  <Button
                    onClick={() => {
                      setIsEditing(false)
                      // Reset form
                      setEditTitle(task.title || '')
                      setEditDescription(task.description || '')
                      setEditStatus(task.status || 'todo')
                      setEditPriority(task.priority || 'medium')
                      setEditDueDate(task.due_date || '')
                      setEditAssigneeId(task.assignee_id?.toString() || '')
                      setEditEstimatedHours(task.estimated_hours?.toString() || '')
                      setEditActualHours(task.actual_hours?.toString() || '')
                    }}
                    variant="secondary"
                    size="sm"
                  >
                    Cancel
                  </Button>
                </>
              )}
            </div>
          </div>

          {/* Title Section */}
          <div>
            {isEditing ? (
              <input
                type="text"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                className="w-full text-3xl font-bold mb-2 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                placeholder="Task title"
              />
            ) : (
              <h1 className="text-3xl font-bold text-gray-900 mb-2">{task.title}</h1>
            )}

            {/* Status badges */}
            <div className="flex flex-wrap gap-2">
              {task.status && (
                <span className={`px-2.5 py-1 text-xs font-semibold rounded-full border ${getStatusColor(task.status)}`}>
                  {getStatusLabel(task.status)}
                </span>
              )}
              {task.priority && (
                <span className={`px-2.5 py-1 text-xs font-semibold rounded-full border ${getPriorityColor(task.priority)}`}>
                  {task.priority}
                </span>
              )}
              {task.tags && task.tags.length > 0 && task.tags.map((tag) => (
                <span
                  key={tag.id}
                  className="px-2.5 py-1 text-xs font-semibold rounded-full border border-gray-200"
                  style={{ backgroundColor: tag.color + '20', color: tag.color }}
                >
                  {tag.name}
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="flex gap-6">
          {/* Left Column - Main Content */}
          <div className="flex-1">
            {/* Description */}
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h2 className="text-sm font-semibold text-gray-900 mb-3">Description</h2>
              {isEditing ? (
                <div>
                  <textarea
                    value={editDescription}
                    onChange={(e) => setEditDescription(e.target.value)}
                    rows={15}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none font-mono text-sm"
                    placeholder="Add a description in markdown format...

## Example
- Use **bold** and *italic*
- Create lists
- Add [links](https://example.com)
- ```code blocks```"
                  />
                  <p className="text-xs text-gray-500 mt-2">
                    <strong>Supports GitHub Flavored Markdown:</strong> headings, lists, links, code blocks, tables, and more
                  </p>
                </div>
              ) : (
                <>
                  {task.description ? (
                    <div className="prose prose-sm max-w-none prose-headings:text-gray-900 prose-p:text-gray-700 prose-a:text-primary-600 prose-code:text-primary-600 prose-code:bg-primary-50 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-pre:bg-gray-50 prose-pre:border prose-pre:border-gray-200">
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {task.description}
                      </ReactMarkdown>
                    </div>
                  ) : (
                    <p className="text-sm text-gray-500 italic">No description provided</p>
                  )}
                </>
              )}
            </div>
          </div>

          {/* Right Column - Metadata */}
          <div className="w-80 flex-shrink-0">
            <div className="space-y-6">
              {/* Status */}
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-2">Status</label>
                {isEditing ? (
                  <select
                    value={editStatus}
                    onChange={(e) => setEditStatus(e.target.value as any)}
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                  >
                    <option value="todo">To Do</option>
                    <option value="in_progress">In Progress</option>
                    <option value="done">Done</option>
                  </select>
                ) : (
                  <span className={`inline-block px-3 py-1 text-sm font-medium rounded-full border ${getStatusColor(task.status || '')}`}>
                    {getStatusLabel(task.status || '')}
                  </span>
                )}
              </div>

              {/* Priority */}
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-2">Priority</label>
                {isEditing ? (
                  <select
                    value={editPriority}
                    onChange={(e) => setEditPriority(e.target.value as 'low' | 'medium' | 'high' | 'urgent')}
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                  >
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                    <option value="urgent">Urgent</option>
                  </select>
                ) : (
                  <span className={`inline-block px-3 py-1 text-sm font-medium rounded-full border ${getPriorityColor(task.priority || '')}`}>
                    {task.priority}
                  </span>
                )}
              </div>

              {/* Assignee */}
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-2">Assignee</label>
                {isEditing ? (
                  <input
                    type="number"
                    value={editAssigneeId}
                    onChange={(e) => setEditAssigneeId(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    placeholder="User ID (leave empty to unassign)"
                  />
                ) : (
                  <div className="flex items-center gap-2">
                    <div className="w-6 h-6 rounded-full bg-primary-100 flex items-center justify-center flex-shrink-0">
                      <svg className="w-3.5 h-3.5 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </div>
                    <p className="text-sm text-gray-900">
                      {task.assignee_id ? `User ${task.assignee_id}` : 'Unassigned'}
                    </p>
                  </div>
                )}
              </div>

              {/* Due Date */}
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-2">Due Date</label>
                {isEditing ? (
                  <input
                    type="date"
                    value={editDueDate}
                    onChange={(e) => setEditDueDate(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                  />
                ) : (
                  <p className="text-sm text-gray-900">
                    {task.due_date ? new Date(task.due_date).toLocaleDateString() : 'No due date'}
                  </p>
                )}
              </div>

              {/* Estimated Hours */}
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-2">Estimated Hours</label>
                {isEditing ? (
                  <input
                    type="number"
                    step="0.5"
                    value={editEstimatedHours}
                    onChange={(e) => setEditEstimatedHours(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    placeholder="0"
                  />
                ) : (
                  <p className="text-sm text-gray-900">
                    {task.estimated_hours !== null ? `${task.estimated_hours}h` : 'Not estimated'}
                  </p>
                )}
              </div>

              {/* Actual Hours */}
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-2">Actual Hours</label>
                {isEditing ? (
                  <input
                    type="number"
                    step="0.5"
                    value={editActualHours}
                    onChange={(e) => setEditActualHours(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    placeholder="0"
                  />
                ) : (
                  <p className="text-sm text-gray-900">
                    {task.actual_hours !== null ? `${task.actual_hours}h` : 'Not tracked'}
                  </p>
                )}
              </div>

              {/* Tags */}
              {task.tags && task.tags.length > 0 && (
                <div>
                  <label className="block text-xs font-semibold text-gray-700 mb-2">Tags</label>
                  <div className="flex flex-wrap gap-1.5">
                    {task.tags.map((tag) => (
                      <span
                        key={tag.id}
                        className="inline-flex items-center gap-1.5 px-2 py-1 text-xs font-medium rounded-md border border-gray-200"
                        style={{ backgroundColor: tag.color + '20', color: tag.color }}
                      >
                        <div className="w-2 h-2 rounded-full" style={{ backgroundColor: tag.color }} />
                        {tag.name}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Timestamps */}
              <div className="pt-6 border-t border-gray-200">
                <div className="space-y-3 text-xs">
                  {task.created_at && (
                    <div>
                      <span className="text-gray-600">Created </span>
                      <span className="text-gray-900 font-medium">
                        {new Date(task.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  )}
                  {task.updated_at && (
                    <div>
                      <span className="text-gray-600">Updated </span>
                      <span className="text-gray-900 font-medium">
                        {new Date(task.updated_at).toLocaleDateString()}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Error notification */}
      {error && task && (
        <div className="fixed bottom-4 right-4 bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg shadow-lg">
          {error}
        </div>
      )}
    </div>
  )
}
