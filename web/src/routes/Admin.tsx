import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../state/AuthContext'
import { api } from '../lib/api'

// Types from backend
interface UserWithStats {
  id: number
  email: string
  is_admin: boolean
  created_at: string
  login_count: number
  last_login_at?: string | null
  last_login_ip?: string | null
  failed_attempts: number
}

interface UserActivity {
  id: number
  user_id: number
  activity_type: string
  ip_address?: string | null
  user_agent?: string | null
  created_at: string
}

export default function Admin() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const [users, setUsers] = useState<UserWithStats[]>([])
  const [selectedUser, setSelectedUser] = useState<UserWithStats | null>(null)
  const [activities, setActivities] = useState<UserActivity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activityLoading, setActivityLoading] = useState(false)

  useEffect(() => {
    // Check if user is admin
    if (!user?.is_admin) {
      navigate('/app')
      return
    }

    loadUsers()
  }, [user, navigate])

  const loadUsers = async () => {
    try {
      setLoading(true)
      setError(null)

      const data = await api.getUsers()
      setUsers(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users')
    } finally {
      setLoading(false)
    }
  }

  const loadUserActivity = async (userId: number) => {
    try {
      setActivityLoading(true)

      const data = await api.getUserActivity(userId)
      setActivities(data)
    } catch (err) {
      console.error('Failed to load activity:', err)
      setActivities([])
    } finally {
      setActivityLoading(false)
    }
  }

  const toggleAdminStatus = async (userId: number, currentStatus: boolean) => {
    try {
      await api.updateUserAdmin(userId, !currentStatus)

      // Reload users to get updated data
      await loadUsers()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update admin status')
    }
  }

  const handleViewActivity = (user: UserWithStats) => {
    setSelectedUser(user)
    loadUserActivity(user.id)
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    })
  }

  const getActivityTypeColor = (type: string) => {
    switch (type) {
      case 'login':
        return 'text-success-300 bg-success-500/10 border-success-500/30'
      case 'failed_login':
        return 'text-danger-300 bg-danger-500/10 border-danger-500/30'
      case 'logout':
        return 'text-dark-text-secondary bg-dark-bg-tertiary/30 border-dark-bg-tertiary/30'
      default:
        return 'text-primary-300 bg-primary-500/10 border-primary-500/30'
    }
  }

  if (loading) {
    return (
      <div className="p-8 bg-dark-bg-primary min-h-screen">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-dark-bg-secondary rounded w-1/3"></div>
          <div className="h-64 bg-dark-bg-secondary rounded"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-8 bg-dark-bg-primary min-h-screen">
        <div className="bg-danger-500/10 border border-danger-500/30 text-danger-300 px-4 py-3 rounded">
          {error}
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col bg-dark-bg-primary">
      {/* Header */}
      <div className="bg-dark-bg-secondary border-b border-dark-bg-tertiary/30 px-8 py-6">
        <h1 className="text-2xl font-bold text-dark-text-primary">Admin Dashboard</h1>
        <p className="mt-1 text-sm text-dark-text-secondary">Manage users and monitor activity</p>
      </div>

      <div className="flex-1 overflow-y-auto p-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Users Table */}
          <div className="bg-dark-bg-secondary rounded-lg shadow-md border border-dark-bg-tertiary/30">
            <div className="px-6 py-4 border-b border-dark-bg-tertiary/30">
              <h2 className="text-lg font-semibold text-dark-text-primary">Users ({users.length})</h2>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-dark-bg-tertiary/30">
                <thead className="bg-dark-bg-tertiary/20">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-dark-text-secondary uppercase">Email</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-dark-text-secondary uppercase">Admin</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-dark-text-secondary uppercase">Logins</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-dark-text-secondary uppercase">Failed</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-dark-text-secondary uppercase">Last IP</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-dark-text-secondary uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-dark-bg-secondary divide-y divide-dark-bg-tertiary/30">
                  {users.map((u) => (
                    <tr key={u.id} className="hover:bg-dark-bg-tertiary/20 transition-colors">
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-dark-text-primary">{u.email}</td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                          u.is_admin ? 'bg-purple-500/10 text-purple-400 border border-purple-500/30' : 'bg-dark-bg-tertiary/30 text-dark-text-secondary border border-dark-bg-tertiary/30'
                        }`}>
                          {u.is_admin ? 'Admin' : 'User'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-dark-text-primary">{u.login_count}</td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`text-sm ${u.failed_attempts > 0 ? 'text-danger-400 font-semibold' : 'text-dark-text-primary'}`}>
                          {u.failed_attempts}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-dark-text-secondary">
                        {u.last_login_ip || 'N/A'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                        <button
                          onClick={() => handleViewActivity(u)}
                          className="text-primary-400 hover:text-primary-300 font-medium transition-colors"
                        >
                          Activity
                        </button>
                        <button
                          onClick={() => toggleAdminStatus(u.id, u.is_admin)}
                          className="text-purple-400 hover:text-purple-300 font-medium transition-colors"
                        >
                          {u.is_admin ? 'Revoke Admin' : 'Make Admin'}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Activity Log */}
          <div className="bg-dark-bg-secondary rounded-lg shadow-md border border-dark-bg-tertiary/30">
            <div className="px-6 py-4 border-b border-dark-bg-tertiary/30">
              <h2 className="text-lg font-semibold text-dark-text-primary">
                Activity Log {selectedUser && `- ${selectedUser.email}`}
              </h2>
            </div>
            <div className="p-6">
              {!selectedUser ? (
                <div className="text-center py-12 text-dark-text-tertiary">
                  <svg
                    className="mx-auto h-12 w-12 text-dark-text-tertiary opacity-50"
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
                  <p className="mt-2">Select a user to view their activity</p>
                </div>
              ) : activityLoading ? (
                <div className="text-center py-12">
                  <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
                  <p className="mt-2 text-dark-text-secondary">Loading activity...</p>
                </div>
              ) : activities.length === 0 ? (
                <div className="text-center py-12 text-dark-text-tertiary">
                  <p>No activity recorded for this user</p>
                </div>
              ) : (
                <div className="space-y-3 max-h-[600px] overflow-y-auto">
                  {activities.map((activity) => (
                    <div key={activity.id} className="border border-dark-bg-tertiary/30 rounded-lg p-4 bg-dark-bg-primary hover:bg-dark-bg-tertiary/10 transition-colors">
                      <div className="flex items-center justify-between">
                        <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium border ${getActivityTypeColor(activity.activity_type)}`}>
                          {activity.activity_type.replace('_', ' ').toUpperCase()}
                        </span>
                        <span className="text-xs text-dark-text-tertiary">{formatDate(activity.created_at)}</span>
                      </div>
                      {activity.ip_address && (
                        <div className="mt-2 text-sm text-dark-text-secondary">
                          <span className="font-medium text-dark-text-primary">IP:</span> {activity.ip_address}
                        </div>
                      )}
                      {activity.user_agent && (
                        <div className="mt-1 text-xs text-dark-text-tertiary truncate">
                          {activity.user_agent}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
