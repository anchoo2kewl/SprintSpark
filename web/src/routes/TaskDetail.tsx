import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import Button from '../components/ui/Button'
import { apiClient, Task, type SwimLane } from '../lib/api'

interface TaskDetailProps {
  isModal?: boolean
  onClose?: () => void
}

export default function TaskDetail({ isModal, onClose }: TaskDetailProps) {
  const { projectId, taskId } = useParams()
  const navigate = useNavigate()
  const [task, setTask] = useState<Task | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  // Inline editing
  const [editingField, setEditingField] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const titleRef = useRef<HTMLInputElement>(null)
  const descRef = useRef<HTMLTextAreaElement>(null)

  // Reference data
  const [sprints, setSprints] = useState<any[]>([])
  const [swimLanes, setSwimLanes] = useState<SwimLane[]>([])
  const [members, setMembers] = useState<any[]>([])

  // Attachments
  const [attachments, setAttachments] = useState<any[]>([])
  const [uploading, setUploading] = useState(false)

  // Comments
  const [comments, setComments] = useState<any[]>([])
  const [newComment, setNewComment] = useState('')
  const [postingComment, setPostingComment] = useState(false)

  useEffect(() => {
    loadTask()
    loadSprints()
    loadSwimLanes()
    loadComments()
    loadMembers()
    loadAttachments()
  }, [projectId, taskId])

  useEffect(() => {
    if (editingField === 'title') titleRef.current?.focus()
    if (editingField === 'description') {
      const el = descRef.current
      if (el) {
        el.focus()
        el.setSelectionRange(el.value.length, el.value.length)
      }
    }
  }, [editingField])

  const loadTask = async () => {
    try {
      setLoading(true)
      const tasks = await apiClient.getTasks(Number(projectId))
      const found = tasks.find((t: Task) => t.id === Number(taskId))
      if (found) setTask(found)
      else setError('Task not found')
    } catch (err: any) {
      setError(err.message || 'Failed to load task')
    } finally {
      setLoading(false)
    }
  }

  const loadSprints = async () => {
    try { setSprints(await apiClient.getSprints()) } catch { /* ignore */ }
  }

  const loadSwimLanes = async () => {
    try {
      const lanes = await apiClient.getSwimLanes(Number(projectId))
      setSwimLanes(lanes.sort((a, b) => a.position - b.position))
    } catch { /* ignore */ }
  }

  const loadComments = async () => {
    try { setComments(await apiClient.getTaskComments(Number(taskId))) } catch { /* ignore */ }
  }

  const loadMembers = async () => {
    try { setMembers(await apiClient.getProjectMembers(Number(projectId))) } catch { /* ignore */ }
  }

  const loadAttachments = async () => {
    try { setAttachments(await apiClient.getTaskAttachments(Number(taskId))) } catch { /* ignore */ }
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file type
    const allowedTypes = ['image/', 'video/', 'application/pdf']
    if (!allowedTypes.some(t => file.type.startsWith(t))) {
      setError('Only images, videos, and PDFs are allowed')
      return
    }

    try {
      setUploading(true)

      // Get upload signature from backend
      const sig = await apiClient.getUploadSignature()

      // Upload directly to Cloudinary
      const formData = new FormData()
      formData.append('file', file)
      formData.append('api_key', sig.api_key)
      formData.append('timestamp', String(sig.timestamp))
      formData.append('signature', sig.signature)

      const uploadRes = await fetch(
        `https://api.cloudinary.com/v1_1/${sig.cloud_name}/auto/upload`,
        { method: 'POST', body: formData }
      )

      if (!uploadRes.ok) throw new Error('Upload to Cloudinary failed')
      const uploadData = await uploadRes.json()

      // Determine file type category
      let fileType = 'image'
      if (file.type.startsWith('video/')) fileType = 'video'
      else if (file.type === 'application/pdf') fileType = 'pdf'

      // Save attachment metadata to backend
      await apiClient.createTaskAttachment(Number(taskId), {
        filename: file.name,
        file_type: fileType,
        content_type: file.type,
        file_size: file.size,
        cloudinary_url: uploadData.secure_url,
        cloudinary_public_id: uploadData.public_id,
      })

      await loadAttachments()
    } catch (err: any) {
      setError(err.message || 'Failed to upload file')
    } finally {
      setUploading(false)
      e.target.value = ''
    }
  }

  const handleDeleteAttachment = async (attachmentId: number) => {
    if (!confirm('Delete this attachment?')) return
    try {
      await apiClient.deleteTaskAttachment(Number(taskId), attachmentId)
      await loadAttachments()
    } catch (err: any) {
      setError(err.message || 'Failed to delete attachment')
    }
  }

  const saveField = async (field: string, value: any) => {
    if (!task) return
    try {
      setSaving(true)
      const update: Record<string, any> = {}

      switch (field) {
        case 'title':
          if (!value?.trim()) return
          update.title = value.trim()
          break
        case 'description':
          update.description = value?.trim() || ''
          break
        case 'swim_lane_id': {
          const lid = Number(value)
          update.swim_lane_id = lid
          const lane = swimLanes.find(l => l.id === lid)
          if (lane) {
            if (lane.name === 'To Do') update.status = 'todo'
            else if (lane.name === 'In Progress') update.status = 'in_progress'
            else if (lane.name === 'Done') update.status = 'done'
          }
          break
        }
        case 'priority':
          update.priority = value
          break
        case 'sprint_id':
          update.sprint_id = value ? parseInt(value) : null
          break
        case 'assignee_id':
          update.assignee_id = value ? parseInt(value) : null
          break
        case 'due_date':
          update.due_date = value || null
          break
        case 'estimated_hours':
          update.estimated_hours = value ? parseFloat(value) : 0
          break
        case 'actual_hours':
          update.actual_hours = value ? parseFloat(value) : 0
          break
      }

      await apiClient.updateTask(task.id!, update)
      await loadTask()
    } catch (err: any) {
      setError(err.message || 'Failed to update')
    } finally {
      setSaving(false)
      setEditingField(null)
    }
  }

  const startEdit = (field: string, currentValue: string) => {
    setEditingField(field)
    setEditValue(currentValue)
  }

  const cancelEdit = () => {
    setEditingField(null)
    setEditValue('')
  }

  const handleKeyDown = (e: React.KeyboardEvent, field: string) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      saveField(field, editValue)
    } else if (e.key === 'Escape') {
      cancelEdit()
    }
  }

  const handlePostComment = async () => {
    if (!newComment.trim()) return
    try {
      setPostingComment(true)
      await apiClient.createTaskComment(Number(taskId), newComment.trim())
      setNewComment('')
      await loadComments()
    } catch (err: any) {
      setError(err.message || 'Failed to post comment')
    } finally {
      setPostingComment(false)
    }
  }

  const handleDelete = async () => {
    if (!task || !confirm('Are you sure you want to delete this task?')) return
    try {
      await apiClient.deleteTask(task.id!)
      handleClose()
    } catch (err: any) {
      setError(err.message || 'Failed to delete task')
    }
  }

  const handleClose = () => {
    if (onClose) onClose()
    else navigate(`/app/projects/${projectId}`)
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent': return 'bg-danger-500/10 text-danger-400 border-danger-500/30'
      case 'high': return 'bg-warning-500/10 text-warning-400 border-warning-500/30'
      case 'medium': return 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30'
      case 'low': return 'bg-success-500/10 text-success-400 border-success-500/30'
      default: return 'bg-dark-bg-tertiary text-dark-text-tertiary border-dark-border-subtle'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'done': return 'bg-success-500/10 text-success-400 border-success-500/30'
      case 'in_progress': return 'bg-primary-500/10 text-primary-400 border-primary-500/30'
      case 'todo': return 'bg-dark-bg-tertiary text-dark-text-tertiary border-dark-border-subtle'
      default: return 'bg-dark-bg-tertiary text-dark-text-tertiary border-dark-border-subtle'
    }
  }

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'done': return 'Done'
      case 'in_progress': return 'In Progress'
      case 'todo': return 'To Do'
      default: return status
    }
  }

  if (loading) {
    return (
      <div className={`${isModal ? 'p-8' : 'min-h-screen bg-dark-bg-primary'} flex items-center justify-center`}>
        <div className="text-dark-text-secondary">Loading task...</div>
      </div>
    )
  }

  if (error && !task) {
    return (
      <div className={`${isModal ? 'p-8' : 'min-h-screen bg-dark-bg-primary py-8'}`}>
        <div className={`${isModal ? '' : 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'}`}>
          <div className="bg-dark-bg-secondary border border-dark-border-subtle rounded-lg p-8 text-center">
            <p className="text-danger-400 mb-4">{error}</p>
            <Button onClick={handleClose}>
              {isModal ? 'Close' : 'Back to Project'}
            </Button>
          </div>
        </div>
      </div>
    )
  }

  if (!task) return null

  const containerClass = isModal ? '' : 'min-h-screen bg-dark-bg-primary'
  const innerClass = isModal ? 'px-6' : 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'

  return (
    <div className={containerClass}>
      {/* Header */}
      <div className="border-b border-dark-border-subtle bg-dark-bg-secondary sticky top-0 z-10">
        <div className={`${innerClass} py-4`}>
          <div className="flex items-center justify-between mb-3">
            <button
              onClick={handleClose}
              className="inline-flex items-center text-sm text-dark-text-secondary hover:text-dark-text-primary transition-colors"
            >
              {isModal ? (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              ) : (
                <>
                  <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                  </svg>
                  Back to project
                </>
              )}
            </button>
            <button
              onClick={handleDelete}
              className="px-3 py-1.5 text-sm text-danger-400 hover:bg-danger-500/10 rounded-lg transition-colors"
            >
              Delete
            </button>
          </div>

          {/* Title - inline editable */}
          {editingField === 'title' ? (
            <input
              ref={titleRef}
              type="text"
              value={editValue}
              onChange={(e) => setEditValue(e.target.value)}
              onBlur={() => saveField('title', editValue)}
              onKeyDown={(e) => handleKeyDown(e, 'title')}
              className="w-full text-2xl font-bold px-2 py-1 -ml-2 border border-dark-border-subtle bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
            />
          ) : (
            <h1
              onClick={() => startEdit('title', task.title || '')}
              className="text-2xl font-bold text-dark-text-primary cursor-text hover:bg-dark-bg-tertiary/50 px-2 py-1 -ml-2 rounded-lg transition-colors"
              title="Click to edit title"
            >
              {task.title}
            </h1>
          )}

          {/* Status badges */}
          <div className="flex flex-wrap gap-2 mt-2">
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
                className="px-2.5 py-1 text-xs font-semibold rounded-full border border-dark-border-subtle"
                style={{ backgroundColor: tag.color + '20', color: tag.color }}
              >
                {tag.name}
              </span>
            ))}
            {saving && (
              <span className="px-2.5 py-1 text-xs text-dark-text-tertiary animate-pulse">
                Saving...
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className={`${innerClass} py-6`}>
        <div className="flex gap-6">
          {/* Left Column - Description & Comments */}
          <div className="flex-1 min-w-0 space-y-6">
            {/* Description */}
            <div className="bg-dark-bg-secondary border border-dark-border-subtle rounded-lg p-6">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-sm font-semibold text-dark-text-primary">Description</h2>
                {editingField !== 'description' && (
                  <button
                    onClick={() => startEdit('description', task.description || '')}
                    className="p-1.5 text-dark-text-tertiary hover:text-dark-text-primary hover:bg-dark-bg-tertiary rounded-md transition-colors"
                    title="Edit description"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                    </svg>
                  </button>
                )}
              </div>

              {editingField === 'description' ? (
                <div>
                  <textarea
                    ref={descRef}
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    rows={12}
                    className="w-full px-3 py-2 border border-dark-border-subtle bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none font-mono text-sm placeholder-dark-text-tertiary"
                    placeholder="Add a description in markdown format..."
                    onKeyDown={(e) => {
                      if (e.key === 'Escape') cancelEdit()
                      if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                        e.preventDefault()
                        saveField('description', editValue)
                      }
                    }}
                  />
                  <div className="flex items-center gap-2 mt-2">
                    <Button size="sm" onClick={() => saveField('description', editValue)} disabled={saving}>
                      {saving ? 'Saving...' : 'Save'}
                    </Button>
                    <Button size="sm" variant="secondary" onClick={cancelEdit}>
                      Cancel
                    </Button>
                    <span className="text-xs text-dark-text-tertiary ml-auto">
                      Markdown supported &middot; {navigator.platform.includes('Mac') ? 'Cmd' : 'Ctrl'}+Enter to save
                    </span>
                  </div>
                </div>
              ) : task.description ? (
                <div className="prose prose-sm max-w-none prose-headings:text-dark-text-primary prose-p:text-dark-text-secondary prose-a:text-primary-400 prose-code:text-primary-400 prose-code:bg-primary-500/10 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-pre:bg-dark-bg-primary prose-pre:border prose-pre:border-dark-border-subtle prose-strong:text-dark-text-primary prose-li:text-dark-text-secondary">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {task.description}
                  </ReactMarkdown>
                </div>
              ) : (
                <p
                  className="text-sm text-dark-text-tertiary italic cursor-pointer hover:text-dark-text-secondary transition-colors"
                  onClick={() => startEdit('description', '')}
                >
                  Click to add a description...
                </p>
              )}
            </div>

            {/* Attachments Section */}
            <div className="bg-dark-bg-secondary border border-dark-border-subtle rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-semibold text-dark-text-primary">
                  Attachments {attachments.length > 0 && `(${attachments.length})`}
                </h2>
                <label className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md cursor-pointer transition-colors ${
                  uploading
                    ? 'bg-dark-bg-tertiary text-dark-text-tertiary cursor-not-allowed'
                    : 'bg-primary-500/10 text-primary-400 hover:bg-primary-500/20'
                }`}>
                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                  </svg>
                  {uploading ? 'Uploading...' : 'Upload'}
                  <input
                    type="file"
                    className="hidden"
                    accept="image/*,video/*,.pdf"
                    onChange={handleFileUpload}
                    disabled={uploading}
                  />
                </label>
              </div>

              {attachments.length === 0 ? (
                <p className="text-sm text-dark-text-tertiary italic">No attachments</p>
              ) : (
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  {attachments.map((att: any) => (
                    <div key={att.id} className="group relative border border-dark-border-subtle rounded-lg overflow-hidden bg-dark-bg-primary">
                      {att.file_type === 'image' ? (
                        <a href={att.cloudinary_url} target="_blank" rel="noopener noreferrer">
                          <img
                            src={att.cloudinary_url}
                            alt={att.filename}
                            className="w-full h-24 object-cover"
                          />
                        </a>
                      ) : att.file_type === 'video' ? (
                        <a href={att.cloudinary_url} target="_blank" rel="noopener noreferrer" className="flex items-center justify-center h-24 bg-dark-bg-tertiary">
                          <svg className="w-8 h-8 text-dark-text-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                        </a>
                      ) : (
                        <a href={att.cloudinary_url} target="_blank" rel="noopener noreferrer" className="flex items-center justify-center h-24 bg-dark-bg-tertiary">
                          <svg className="w-8 h-8 text-dark-text-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                          </svg>
                        </a>
                      )}
                      <div className="p-2">
                        <p className="text-xs text-dark-text-primary truncate" title={att.filename}>{att.filename}</p>
                        <p className="text-[10px] text-dark-text-tertiary">
                          {(att.file_size / 1024 / 1024).toFixed(1)} MB
                        </p>
                      </div>
                      <button
                        onClick={() => handleDeleteAttachment(att.id)}
                        className="absolute top-1 right-1 p-1 bg-dark-bg-primary/80 rounded-md opacity-0 group-hover:opacity-100 transition-opacity text-danger-400 hover:bg-danger-500/10"
                        title="Delete attachment"
                      >
                        <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Comments Section */}
            <div className="bg-dark-bg-secondary border border-dark-border-subtle rounded-lg p-6">
              <h2 className="text-sm font-semibold text-dark-text-primary mb-4">
                Comments {comments.length > 0 && `(${comments.length})`}
              </h2>

              <div className="mb-4">
                <textarea
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 border border-dark-border-subtle bg-dark-bg-primary text-dark-text-primary rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none text-sm placeholder-dark-text-tertiary"
                  placeholder="Add a comment..."
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey) && newComment.trim()) {
                      e.preventDefault()
                      handlePostComment()
                    }
                  }}
                />
                <div className="flex justify-between items-center mt-2">
                  <span className="text-xs text-dark-text-tertiary">
                    {navigator.platform.includes('Mac') ? 'Cmd' : 'Ctrl'}+Enter to post
                  </span>
                  <Button
                    onClick={handlePostComment}
                    size="sm"
                    disabled={!newComment.trim() || postingComment}
                  >
                    {postingComment ? 'Posting...' : 'Post Comment'}
                  </Button>
                </div>
              </div>

              <div className="space-y-4">
                {comments.length === 0 ? (
                  <p className="text-sm text-dark-text-tertiary italic">No comments yet</p>
                ) : (
                  comments.map((comment) => (
                    <div key={comment.id} className="border-t border-dark-border-subtle pt-4 first:border-t-0 first:pt-0">
                      <div className="flex items-start gap-3">
                        <div className="w-8 h-8 rounded-full bg-primary-500/10 flex items-center justify-center flex-shrink-0">
                          <span className="text-xs font-medium text-primary-400">
                            {(comment.user_name || 'U').charAt(0).toUpperCase()}
                          </span>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="text-sm font-medium text-dark-text-primary">
                              {comment.user_name || `User ${comment.user_id}`}
                            </span>
                            <span className="text-xs text-dark-text-tertiary">
                              {new Date(comment.created_at).toLocaleString()}
                            </span>
                          </div>
                          <div className="text-sm text-dark-text-secondary whitespace-pre-wrap break-words">
                            {comment.comment}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>

          {/* Right Column - Sidebar */}
          <div className="w-72 flex-shrink-0">
            <div className="bg-dark-bg-secondary border border-dark-border-subtle rounded-lg divide-y divide-dark-border-subtle">
              {/* Swim Lane */}
              <SidebarField label="Swim Lane">
                <InlineSelect
                  value={String(task.swim_lane_id ?? '')}
                  onChange={(v) => saveField('swim_lane_id', v)}
                  options={swimLanes.map(l => ({ value: String(l.id), label: l.name }))}
                />
              </SidebarField>

              {/* Priority */}
              <SidebarField label="Priority">
                <InlineSelect
                  value={task.priority || 'medium'}
                  onChange={(v) => saveField('priority', v)}
                  options={[
                    { value: 'low', label: 'Low' },
                    { value: 'medium', label: 'Medium' },
                    { value: 'high', label: 'High' },
                    { value: 'urgent', label: 'Urgent' },
                  ]}
                />
              </SidebarField>

              {/* Sprint */}
              <SidebarField label="Sprint">
                <InlineSelect
                  value={task.sprint_id?.toString() || ''}
                  onChange={(v) => saveField('sprint_id', v)}
                  options={[
                    { value: '', label: 'No sprint' },
                    ...sprints.map(s => ({ value: String(s.id), label: s.name })),
                  ]}
                />
              </SidebarField>

              {/* Assignee */}
              <SidebarField label="Assignee">
                {members.length > 0 ? (
                  <InlineSelect
                    value={task.assignee_id?.toString() || ''}
                    onChange={(v) => saveField('assignee_id', v)}
                    options={[
                      { value: '', label: 'Unassigned' },
                      ...members.map(m => ({ value: String(m.user_id || m.id), label: m.user_name || m.email || `User ${m.user_id || m.id}` })),
                    ]}
                  />
                ) : (
                  <div className="flex items-center gap-2 px-3 py-1.5">
                    <div className="w-5 h-5 rounded-full bg-primary-500/10 flex items-center justify-center flex-shrink-0">
                      <svg className="w-3 h-3 text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </div>
                    <span className="text-sm text-dark-text-primary">
                      {task.assignee_name || (task.assignee_id ? `User ${task.assignee_id}` : 'Unassigned')}
                    </span>
                  </div>
                )}
              </SidebarField>

              {/* Due Date */}
              <SidebarField label="Due Date">
                {editingField === 'due_date' ? (
                  <input
                    type="date"
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    onBlur={() => saveField('due_date', editValue)}
                    onKeyDown={(e) => handleKeyDown(e, 'due_date')}
                    className="w-full text-sm bg-dark-bg-primary border border-dark-border-subtle text-dark-text-primary rounded-md px-3 py-1.5 focus:ring-1 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    autoFocus
                  />
                ) : (
                  <button
                    onClick={() => startEdit('due_date', task.due_date?.split('T')[0] || '')}
                    className="text-sm text-dark-text-primary hover:bg-dark-bg-tertiary/50 px-3 py-1.5 rounded-md w-full text-left transition-colors"
                  >
                    {task.due_date ? new Date(task.due_date).toLocaleDateString() : 'None'}
                  </button>
                )}
              </SidebarField>

              {/* Estimated Hours */}
              <SidebarField label="Estimated Hours">
                {editingField === 'estimated_hours' ? (
                  <input
                    type="number"
                    step="0.5"
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    onBlur={() => saveField('estimated_hours', editValue)}
                    onKeyDown={(e) => handleKeyDown(e, 'estimated_hours')}
                    className="w-full text-sm bg-dark-bg-primary border border-dark-border-subtle text-dark-text-primary rounded-md px-3 py-1.5 focus:ring-1 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    autoFocus
                  />
                ) : (
                  <button
                    onClick={() => startEdit('estimated_hours', String(task.estimated_hours ?? 0))}
                    className="text-sm text-dark-text-primary hover:bg-dark-bg-tertiary/50 px-3 py-1.5 rounded-md w-full text-left transition-colors"
                  >
                    {task.estimated_hours ?? 0}h
                  </button>
                )}
              </SidebarField>

              {/* Actual Hours */}
              <SidebarField label="Actual Hours">
                {editingField === 'actual_hours' ? (
                  <input
                    type="number"
                    step="0.5"
                    value={editValue}
                    onChange={(e) => setEditValue(e.target.value)}
                    onBlur={() => saveField('actual_hours', editValue)}
                    onKeyDown={(e) => handleKeyDown(e, 'actual_hours')}
                    className="w-full text-sm bg-dark-bg-primary border border-dark-border-subtle text-dark-text-primary rounded-md px-3 py-1.5 focus:ring-1 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    autoFocus
                  />
                ) : (
                  <button
                    onClick={() => startEdit('actual_hours', String(task.actual_hours ?? 0))}
                    className="text-sm text-dark-text-primary hover:bg-dark-bg-tertiary/50 px-3 py-1.5 rounded-md w-full text-left transition-colors"
                  >
                    {task.actual_hours ?? 0}h
                  </button>
                )}
              </SidebarField>

              {/* Tags */}
              {task.tags && task.tags.length > 0 && (
                <SidebarField label="Tags">
                  <div className="flex flex-wrap gap-1.5 px-3 py-1.5">
                    {task.tags.map((tag) => (
                      <span
                        key={tag.id}
                        className="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-md border border-dark-border-subtle"
                        style={{ backgroundColor: tag.color + '20', color: tag.color }}
                      >
                        <div className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: tag.color }} />
                        {tag.name}
                      </span>
                    ))}
                  </div>
                </SidebarField>
              )}
            </div>

            {/* Timestamps */}
            <div className="mt-4 space-y-2 text-xs text-dark-text-tertiary px-1">
              {task.created_at && (
                <div>
                  Created <span className="text-dark-text-secondary">{new Date(task.created_at).toLocaleDateString()}</span>
                </div>
              )}
              {task.updated_at && (
                <div>
                  Updated <span className="text-dark-text-secondary">{new Date(task.updated_at).toLocaleDateString()}</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Error notification */}
      {error && task && (
        <div className="fixed bottom-4 right-4 bg-danger-500/10 border border-danger-500/30 text-danger-400 px-4 py-3 rounded-lg shadow-lg z-50">
          <div className="flex items-center gap-2">
            <span>{error}</span>
            <button onClick={() => setError('')} className="text-danger-400 hover:text-danger-300">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

/* Sidebar helper components */

function SidebarField({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="px-4 py-3">
      <label className="block text-[11px] font-medium text-dark-text-tertiary uppercase tracking-wide mb-1">
        {label}
      </label>
      {children}
    </div>
  )
}

function InlineSelect({ value, onChange, options }: {
  value: string
  onChange: (value: string) => void
  options: { value: string; label: string }[]
}) {
  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full appearance-none bg-transparent cursor-pointer text-sm text-dark-text-primary hover:bg-dark-bg-tertiary/50 pl-3 pr-7 py-1.5 rounded-md border border-transparent hover:border-dark-border-subtle focus:border-primary-500 focus:ring-1 focus:ring-primary-500/30 outline-none transition-colors"
      >
        {options.map(opt => (
          <option key={opt.value} value={opt.value} className="bg-dark-bg-secondary text-dark-text-primary">
            {opt.label}
          </option>
        ))}
      </select>
      <svg className="absolute right-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-dark-text-tertiary pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
      </svg>
    </div>
  )
}
