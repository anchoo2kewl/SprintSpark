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
  const [editSprintId, setEditSprintId] = useState('')
  const [editAssigneeId, setEditAssigneeId] = useState('')
  const [editEstimatedHours, setEditEstimatedHours] = useState('')
  const [editActualHours, setEditActualHours] = useState('')

  // Sprints for selector
  const [sprints, setSprints] = useState<any[]>([])

  // Comments state
  const [comments, setComments] = useState<any[]>([])
  const [newComment, setNewComment] = useState('')
  const [postingComment, setPostingComment] = useState(false)

  useEffect(() => {
    loadTask()
    loadSprints()
    loadComments()
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
        setEditSprintId(foundTask.sprint_id?.toString() || '')
        setEditAssigneeId(foundTask.assignee_id?.toString() || '')
        setEditEstimatedHours(foundTask.estimated_hours?.toString() || '0')
        setEditActualHours(foundTask.actual_hours?.toString() || '0')
      } else {
        setError('Task not found')
      }
    } catch (error: any) {
      setError(error.message || 'Failed to load task')
    } finally {
      setLoading(false)
    }
  }

  const loadSprints = async () => {
    try {
      const sprintsData = await apiClient.getSprints()
      setSprints(sprintsData)
    } catch (error: any) {
      console.error('Failed to load sprints:', error)
    }
  }

  const loadComments = async () => {
    try {
      const commentsData = await apiClient.getTaskComments(Number(taskId))
      setComments(commentsData)
    } catch (error: any) {
      console.error('Failed to load comments:', error)
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
        sprint_id: editSprintId ? parseInt(editSprintId) : null,
        assignee_id: editAssigneeId ? parseInt(editAssigneeId) : undefined,
        estimated_hours: editEstimatedHours ? parseFloat(editEstimatedHours) : 0,
        actual_hours: editActualHours ? parseFloat(editActualHours) : 0,
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

  const handlePostComment = async () => {
    if (!newComment.trim()) return

    try {
      setPostingComment(true)
      await apiClient.createTaskComment(Number(taskId), newComment.trim())
      setNewComment('')
      await loadComments()
    } catch (error: any) {
      setError(error.message || 'Failed to post comment')
    } finally {
      setPostingComment(false)
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
      <div className="min-h-screen bg-dark-bg-primary flex items-center justify-center">
        <div className="text-dark-text-secondary">Loading task...</div>
      </div>
    )
  }

  if (error && !task) {
    return (
      <div className="min-h-screen bg-dark-bg-primary py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-dark-bg-secondary border border-dark-bg-tertiary/30 rounded-lg p-8 text-center">
            <p className="text-danger-400 mb-4">{error}</p>
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
    <div className="min-h-screen bg-dark-bg-primary">
      {/* Header */}
      <div className="border-b border-dark-bg-tertiary/20 bg-dark-bg-secondary">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between mb-4">
            <button
              onClick={() => navigate(`/app/projects/${projectId}`)}
              className="inline-flex items-center text-sm text-dark-text-secondary hover:text-dark-text-primary"
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
                    className="px-3 py-1.5 text-sm text-danger-400 hover:bg-danger-500/10 rounded-lg transition-colors"
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
                      setEditSprintId(task.sprint_id?.toString() || '')
                      setEditAssigneeId(task.assignee_id?.toString() || '')
                      setEditEstimatedHours(task.estimated_hours?.toString() || '0')
                      setEditActualHours(task.actual_hours?.toString() || '0')
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
                className="w-full text-3xl font-bold mb-2 px-3 py-2 border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                placeholder="Task title"
              />
            ) : (
              <h1 className="text-3xl font-bold text-dark-text-primary mb-2">{task.title}</h1>
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
                  className="px-2.5 py-1 text-xs font-semibold rounded-full border border-dark-bg-tertiary/30"
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
          <div className="flex-1 space-y-6">
            {/* Description */}
            <div className="bg-dark-bg-secondary border border-dark-bg-tertiary/30 rounded-lg p-6">
              <h2 className="text-sm font-semibold text-dark-text-primary mb-3">Description</h2>
              {isEditing ? (
                <div>
                  <textarea
                    value={editDescription}
                    onChange={(e) => setEditDescription(e.target.value)}
                    rows={15}
                    className="w-full px-3 py-2 border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none font-mono text-sm placeholder-dark-text-tertiary"
                    placeholder="Add a description in markdown format...

## Example
- Use **bold** and *italic*
- Create lists
- Add [links](https://example.com)
- ```code blocks```"
                  />
                  <p className="text-xs text-dark-text-tertiary mt-2">
                    <strong>Supports GitHub Flavored Markdown:</strong> headings, lists, links, code blocks, tables, and more
                  </p>
                </div>
              ) : (
                <>
                  {task.description ? (
                    <div className="prose prose-sm max-w-none prose-headings:text-dark-text-primary prose-p:text-dark-text-secondary prose-a:text-primary-400 prose-code:text-primary-400 prose-code:bg-primary-500/10 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-pre:bg-dark-bg-primary prose-pre:border prose-pre:border-dark-bg-tertiary/30">
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {task.description}
                      </ReactMarkdown>
                    </div>
                  ) : (
                    <p className="text-sm text-dark-text-tertiary italic">No description provided</p>
                  )}
                </>
              )}
            </div>

            {/* Comments Section */}
            <div className="bg-dark-bg-secondary border border-dark-bg-tertiary/30 rounded-lg p-6">
              <h2 className="text-sm font-semibold text-dark-text-primary mb-4">Comments</h2>

              {/* Comment input */}
              <div className="mb-4">
                <textarea
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none text-sm placeholder-dark-text-tertiary"
                  placeholder="Add a comment..."
                />
                <div className="flex justify-end mt-2">
                  <Button
                    onClick={handlePostComment}
                    size="sm"
                    disabled={!newComment.trim() || postingComment}
                  >
                    {postingComment ? 'Posting...' : 'Post Comment'}
                  </Button>
                </div>
              </div>

              {/* Comments list */}
              <div className="space-y-4">
                {comments.length === 0 ? (
                  <p className="text-sm text-dark-text-tertiary italic">No comments yet</p>
                ) : (
                  comments.map((comment) => (
                    <div key={comment.id} className="border-t border-dark-bg-tertiary/20 pt-4 first:border-t-0 first:pt-0">
                      <div className="flex items-start gap-3">
                        <div className="w-8 h-8 rounded-full bg-primary-500/10 flex items-center justify-center flex-shrink-0">
                          <svg className="w-4 h-4 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                          </svg>
                        </div>
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="text-sm font-medium text-dark-text-primary">
                              {comment.user_name || `User ${comment.user_id}`}
                            </span>
                            <span className="text-xs text-dark-text-tertiary">
                              {new Date(comment.created_at).toLocaleString()}
                            </span>
                          </div>
                          <p className="text-sm text-dark-text-secondary whitespace-pre-wrap">{comment.comment}</p>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>

          {/* Right Column - Metadata */}
          <div className="w-80 flex-shrink-0">
            <div className="space-y-6">
              {/* Status */}
              <div>
                <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Status</label>
                {isEditing ? (
                  <select
                    value={editStatus}
                    onChange={(e) => setEditStatus(e.target.value as any)}
                    className="w-full px-3 py-2 text-sm border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
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
                <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Priority</label>
                {isEditing ? (
                  <select
                    value={editPriority}
                    onChange={(e) => setEditPriority(e.target.value as 'low' | 'medium' | 'high' | 'urgent')}
                    className="w-full px-3 py-2 text-sm border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
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

              {/* Sprint */}
              <div>
                <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Sprint</label>
                {isEditing ? (
                  <select
                    value={editSprintId}
                    onChange={(e) => setEditSprintId(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                  >
                    <option value="">No Sprint</option>
                    {sprints.map((sprint) => (
                      <option key={sprint.id} value={sprint.id}>
                        {sprint.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <p className="text-sm text-dark-text-primary">
                    {task.sprint_name || 'No sprint assigned'}
                  </p>
                )}
              </div>

              {/* Assignee */}
              <div>
                <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Assignee</label>
                {isEditing ? (
                  <input
                    type="number"
                    value={editAssigneeId}
                    onChange={(e) => setEditAssigneeId(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none placeholder-dark-text-tertiary"
                    placeholder="User ID (leave empty to unassign)"
                  />
                ) : (
                  <div className="flex items-center gap-2">
                    <div className="w-6 h-6 rounded-full bg-primary-500/10 flex items-center justify-center flex-shrink-0">
                      <svg className="w-3.5 h-3.5 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </div>
                    <p className="text-sm text-dark-text-primary">
                      {task.assignee_name || (task.assignee_id ? `User ${task.assignee_id}` : 'Unassigned')}
                    </p>
                  </div>
                )}
              </div>

              {/* Due Date */}
              <div>
                <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Due Date</label>
                {isEditing ? (
                  <input
                    type="date"
                    value={editDueDate}
                    onChange={(e) => setEditDueDate(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                  />
                ) : (
                  <p className="text-sm text-dark-text-primary">
                    {task.due_date ? new Date(task.due_date).toLocaleDateString() : 'No due date'}
                  </p>
                )}
              </div>

              {/* Estimated Hours */}
              <div>
                <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Estimated Hours</label>
                {isEditing ? (
                  <input
                    type="number"
                    step="0.5"
                    value={editEstimatedHours}
                    onChange={(e) => setEditEstimatedHours(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none placeholder-dark-text-tertiary"
                    placeholder="0"
                  />
                ) : (
                  <p className="text-sm text-dark-text-primary">
                    {task.estimated_hours ?? 0}h
                  </p>
                )}
              </div>

              {/* Actual Hours */}
              <div>
                <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Actual Hours</label>
                {isEditing ? (
                  <input
                    type="number"
                    step="0.5"
                    value={editActualHours}
                    onChange={(e) => setEditActualHours(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-dark-bg-tertiary/30 bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none placeholder-dark-text-tertiary"
                    placeholder="0"
                  />
                ) : (
                  <p className="text-sm text-dark-text-primary">
                    {task.actual_hours ?? 0}h
                  </p>
                )}
              </div>

              {/* Tags */}
              {task.tags && task.tags.length > 0 && (
                <div>
                  <label className="block text-xs font-semibold text-dark-text-secondary mb-2">Tags</label>
                  <div className="flex flex-wrap gap-1.5">
                    {task.tags.map((tag) => (
                      <span
                        key={tag.id}
                        className="inline-flex items-center gap-1.5 px-2 py-1 text-xs font-medium rounded-md border border-dark-bg-tertiary/30"
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
              <div className="pt-6 border-t border-dark-bg-tertiary/20">
                <div className="space-y-3 text-xs">
                  {task.created_at && (
                    <div>
                      <span className="text-dark-text-secondary">Created </span>
                      <span className="text-dark-text-primary font-medium">
                        {new Date(task.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  )}
                  {task.updated_at && (
                    <div>
                      <span className="text-dark-text-secondary">Updated </span>
                      <span className="text-dark-text-primary font-medium">
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
        <div className="fixed bottom-4 right-4 bg-danger-500/10 border border-danger-500/30 text-danger-400 px-4 py-3 rounded-lg shadow-lg">
          {error}
        </div>
      )}
    </div>
  )
}
