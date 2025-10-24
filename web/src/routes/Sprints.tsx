import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import TextInput from '../components/ui/TextInput'
import FormError from '../components/ui/FormError'
import { apiClient } from '../lib/api'

interface Sprint {
  id: number
  name: string
  goal: string
  start_date: string
  end_date: string
  status: 'planned' | 'active' | 'completed'
  created_at: string
}

export default function Sprints() {
  const navigate = useNavigate()
  const [sprints, setSprints] = useState<Sprint[]>([])
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState<{
    name: string
    goal: string
    start_date: string
    end_date: string
    status: 'planned' | 'active' | 'completed'
  }>({
    name: '',
    goal: '',
    start_date: '',
    end_date: '',
    status: 'planned',
  })
  const [editingId, setEditingId] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => {
    loadSprints()
  }, [])

  const loadSprints = async () => {
    try {
      const data = await apiClient.getSprints()
      setSprints(data)
    } catch (error: any) {
      console.error('Failed to load sprints:', error)
    }
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')

    if (!formData.name.trim()) {
      setError('Sprint name is required')
      return
    }

    try {
      if (editingId) {
        await apiClient.updateSprint(editingId, formData)
        setSuccess('Sprint updated successfully')
      } else {
        await apiClient.createSprint(formData)
        setSuccess('Sprint created successfully')
      }

      setShowForm(false)
      setEditingId(null)
      setFormData({ name: '', goal: '', start_date: '', end_date: '', status: 'planned' })
      loadSprints()
    } catch (error: any) {
      setError(error.message || 'Failed to save sprint')
    }
  }

  const handleEdit = (sprint: Sprint) => {
    setEditingId(sprint.id)
    setFormData({
      name: sprint.name,
      goal: sprint.goal || '',
      start_date: sprint.start_date || '',
      end_date: sprint.end_date || '',
      status: sprint.status,
    })
    setShowForm(true)
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this sprint?')) {
      return
    }

    try {
      await apiClient.deleteSprint(id)
      setSuccess('Sprint deleted successfully')
      loadSprints()
    } catch (error: any) {
      setError(error.message || 'Failed to delete sprint')
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800 border-green-200'
      case 'planned':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      case 'completed':
        return 'bg-gray-100 text-gray-800 border-gray-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Sprints</h1>
              <p className="text-gray-600 mt-1">Organize tasks into time-boxed iterations</p>
            </div>
            <Button onClick={() => navigate('/app')} variant="secondary">
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back
            </Button>
          </div>
        </div>

        <Card className="shadow-md">
          <div className="p-6 sm:p-8">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <h2 className="text-xl font-semibold text-gray-900">Manage Sprints</h2>
              </div>
              <Button
                onClick={() => {
                  setShowForm(true)
                  setEditingId(null)
                  setFormData({ name: '', goal: '', start_date: '', end_date: '', status: 'planned' })
                }}
                size="sm"
              >
                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                New Sprint
              </Button>
            </div>

            {success && (
              <div className="mb-4 p-3 bg-green-50 border-l-4 border-green-400 rounded-r text-green-800 text-sm">
                {success}
              </div>
            )}

            {error && <FormError message={error} className="mb-4" />}

            {showForm && (
              <form onSubmit={handleSave} className="mb-6 p-4 bg-purple-50 border border-purple-200 rounded-lg">
                <h3 className="font-semibold text-gray-900 mb-4">
                  {editingId ? 'Edit Sprint' : 'New Sprint'}
                </h3>

                <div className="space-y-3">
                  <TextInput
                    label="Sprint Name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="Sprint 1"
                    required
                  />

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Goal (Optional)</label>
                    <textarea
                      value={formData.goal}
                      onChange={(e) => setFormData({ ...formData, goal: e.target.value })}
                      placeholder="What do you want to achieve?"
                      rows={2}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <TextInput
                      label="Start Date"
                      type="date"
                      value={formData.start_date}
                      onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                    />

                    <TextInput
                      label="End Date"
                      type="date"
                      value={formData.end_date}
                      onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                    <select
                      value={formData.status}
                      onChange={(e) => setFormData({ ...formData, status: e.target.value as any })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                    >
                      <option value="planned">Planned</option>
                      <option value="active">Active</option>
                      <option value="completed">Completed</option>
                    </select>
                  </div>
                </div>

                <div className="flex gap-2 mt-4">
                  <Button type="submit" size="sm">
                    {editingId ? 'Update' : 'Create'}
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={() => {
                      setShowForm(false)
                      setEditingId(null)
                    }}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            )}

            <div className="space-y-3">
              {sprints.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <svg className="w-12 h-12 mx-auto mb-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  <p>No sprints yet</p>
                  <p className="text-sm mt-1">Create your first sprint to get started</p>
                </div>
              ) : (
                sprints.map((sprint) => (
                  <div
                    key={sprint.id}
                    className="p-4 bg-white border border-gray-200 rounded-lg hover:border-gray-300 transition-colors"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="font-semibold text-gray-900">{sprint.name}</h3>
                          <span className={`px-2 py-0.5 text-xs font-medium rounded-full border ${getStatusColor(sprint.status)}`}>
                            {sprint.status}
                          </span>
                        </div>
                        {sprint.goal && <p className="text-sm text-gray-600 mb-2">{sprint.goal}</p>}
                        {(sprint.start_date || sprint.end_date) && (
                          <p className="text-xs text-gray-500">
                            {sprint.start_date && new Date(sprint.start_date).toLocaleDateString()}
                            {sprint.start_date && sprint.end_date && ' - '}
                            {sprint.end_date && new Date(sprint.end_date).toLocaleDateString()}
                          </p>
                        )}
                      </div>
                      <div className="flex gap-1">
                        <button
                          onClick={() => handleEdit(sprint)}
                          className="p-1.5 text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          title="Edit"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        </button>
                        <button
                          onClick={() => handleDelete(sprint.id)}
                          className="p-1.5 text-red-600 hover:bg-red-50 rounded transition-colors"
                          title="Delete"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}
