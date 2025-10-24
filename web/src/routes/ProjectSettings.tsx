import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import TextInput from '../components/ui/TextInput'
import FormError from '../components/ui/FormError'
import { apiClient } from '../lib/api'

interface ProjectMember {
  id: number
  user_id: number
  email: string
  role: string
  created_at: string
}

interface GitHubSettings {
  github_repo_url: string
  github_owner: string
  github_repo_name: string
  github_branch: string
  github_sync_enabled: boolean
  github_last_sync: string | null
}

export default function ProjectSettings() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const projectId = parseInt(id || '0')

  // Members state
  const [members, setMembers] = useState<ProjectMember[]>([])
  const [newMemberEmail, setNewMemberEmail] = useState('')
  const [newMemberRole, setNewMemberRole] = useState('viewer')
  const [memberError, setMemberError] = useState('')
  const [memberSuccess, setMemberSuccess] = useState('')
  const [isAddingMember, setIsAddingMember] = useState(false)

  // GitHub state
  const [githubSettings, setGithubSettings] = useState<GitHubSettings>({
    github_repo_url: '',
    github_owner: '',
    github_repo_name: '',
    github_branch: 'main',
    github_sync_enabled: false,
    github_last_sync: null,
  })
  const [githubError, setGithubError] = useState('')
  const [githubSuccess, setGithubSuccess] = useState('')
  const [isSavingGitHub, setIsSavingGitHub] = useState(false)

  useEffect(() => {
    loadMembers()
    loadGitHubSettings()
  }, [projectId])

  const loadMembers = async () => {
    try {
      const data = await apiClient.getProjectMembers(projectId)
      setMembers(data)
    } catch (error: any) {
      console.error('Failed to load members:', error)
    }
  }

  const loadGitHubSettings = async () => {
    try {
      const data = await apiClient.getProjectGitHub(projectId)
      setGithubSettings(data)
    } catch (error: any) {
      console.error('Failed to load GitHub settings:', error)
    }
  }

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault()
    setMemberError('')
    setMemberSuccess('')

    if (!newMemberEmail) {
      setMemberError('Email is required')
      return
    }

    setIsAddingMember(true)

    try {
      await apiClient.addProjectMember(projectId, {
        email: newMemberEmail,
        role: newMemberRole,
      })

      setMemberSuccess('Member added successfully')
      setNewMemberEmail('')
      setNewMemberRole('viewer')
      loadMembers()
    } catch (error: any) {
      setMemberError(error.message || 'Failed to add member')
    } finally {
      setIsAddingMember(false)
    }
  }

  const handleUpdateMemberRole = async (memberId: number, role: string) => {
    try {
      await apiClient.updateProjectMember(projectId, memberId, { role })
      setMemberSuccess('Member role updated successfully')
      loadMembers()
    } catch (error: any) {
      setMemberError(error.message || 'Failed to update member role')
    }
  }

  const handleRemoveMember = async (memberId: number) => {
    if (!confirm('Are you sure you want to remove this member?')) {
      return
    }

    try {
      await apiClient.removeProjectMember(projectId, memberId)
      setMemberSuccess('Member removed successfully')
      loadMembers()
    } catch (error: any) {
      setMemberError(error.message || 'Failed to remove member')
    }
  }

  const handleSaveGitHub = async (e: React.FormEvent) => {
    e.preventDefault()
    setGithubError('')
    setGithubSuccess('')

    setIsSavingGitHub(true)

    try {
      await apiClient.updateProjectGitHub(projectId, {
        github_repo_url: githubSettings.github_repo_url,
        github_owner: githubSettings.github_owner,
        github_repo_name: githubSettings.github_repo_name,
        github_branch: githubSettings.github_branch,
        github_sync_enabled: githubSettings.github_sync_enabled,
      })

      setGithubSuccess('GitHub settings saved successfully')
    } catch (error: any) {
      setGithubError(error.message || 'Failed to save GitHub settings')
    } finally {
      setIsSavingGitHub(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Project Settings</h1>
              <p className="text-gray-600 mt-1">Manage team members and GitHub integration</p>
            </div>
            <Button onClick={() => navigate(`/app/projects/${projectId}`)} variant="secondary">
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back to Project
            </Button>
          </div>
        </div>

        <div className="space-y-6">
          {/* Team Members Section */}
          <Card className="shadow-md">
            <div className="p-6 sm:p-8">
              <div className="flex items-start gap-4 mb-6">
                <div className="flex-shrink-0 w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                  </svg>
                </div>
                <div className="flex-1">
                  <h2 className="text-xl font-semibold text-gray-900 mb-1">Team Members</h2>
                  <p className="text-sm text-gray-600">Share this project with other users</p>
                </div>
              </div>

              {memberSuccess && (
                <div className="mb-4 p-4 bg-green-50 border-l-4 border-green-400 rounded-r-lg">
                  <div className="flex items-center">
                    <svg className="w-5 h-5 text-green-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    <span className="text-green-800 font-medium">{memberSuccess}</span>
                  </div>
                </div>
              )}

              {memberError && <FormError message={memberError} className="mb-4" />}

              {/* Add Member Form */}
              <form onSubmit={handleAddMember} className="mb-6 p-4 bg-gray-50 border border-gray-200 rounded-lg">
                <h3 className="font-semibold text-gray-900 mb-4">Add Team Member</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                  <div className="md:col-span-2">
                    <TextInput
                      label="Email Address"
                      type="email"
                      value={newMemberEmail}
                      onChange={(e) => setNewMemberEmail(e.target.value)}
                      placeholder="colleague@example.com"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Role <span className="text-red-500">*</span>
                    </label>
                    <select
                      value={newMemberRole}
                      onChange={(e) => setNewMemberRole(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-colors"
                    >
                      <option value="viewer">Viewer</option>
                      <option value="editor">Editor</option>
                      <option value="admin">Admin</option>
                    </select>
                  </div>
                </div>
                <div className="mt-4">
                  <Button type="submit" disabled={isAddingMember} size="sm">
                    {isAddingMember ? 'Adding...' : 'Add Member'}
                  </Button>
                </div>
              </form>

              {/* Members List */}
              <div>
                <h3 className="font-semibold text-gray-900 mb-3">Current Members ({members.length})</h3>
                {members.length === 0 ? (
                  <div className="text-center py-8 text-gray-500">
                    <svg className="w-12 h-12 mx-auto mb-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                    <p>No members added yet</p>
                    <p className="text-sm mt-1">Add registered users to collaborate on this project</p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {members.map((member) => (
                      <div
                        key={member.id}
                        className="flex items-center justify-between p-4 bg-white border border-gray-200 rounded-lg hover:border-gray-300 transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-full flex items-center justify-center text-white font-semibold">
                            {member.email.charAt(0).toUpperCase()}
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">{member.email}</p>
                            <p className="text-xs text-gray-500">Added {new Date(member.created_at).toLocaleDateString()}</p>
                          </div>
                        </div>
                        <div className="flex items-center gap-3">
                          <select
                            value={member.role}
                            onChange={(e) => handleUpdateMemberRole(member.id, e.target.value)}
                            className="px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                          >
                            <option value="viewer">Viewer</option>
                            <option value="editor">Editor</option>
                            <option value="admin">Admin</option>
                          </select>
                          <button
                            onClick={() => handleRemoveMember(member.id)}
                            className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                            title="Remove member"
                          >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Role Descriptions */}
              <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                <h4 className="font-semibold text-blue-900 mb-2 text-sm">Role Permissions</h4>
                <ul className="text-sm text-blue-800 space-y-1">
                  <li><strong>Viewer:</strong> Can view project and tasks</li>
                  <li><strong>Editor:</strong> Can view, create, and edit tasks</li>
                  <li><strong>Admin:</strong> Full access including managing members and settings</li>
                </ul>
              </div>
            </div>
          </Card>

          {/* GitHub Integration Section */}
          <Card className="shadow-md">
            <div className="p-6 sm:p-8">
              <div className="flex items-start gap-4 mb-6">
                <div className="flex-shrink-0 w-10 h-10 bg-gray-900 rounded-lg flex items-center justify-center">
                  <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                    <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="flex-1">
                  <h2 className="text-xl font-semibold text-gray-900 mb-1">GitHub Integration</h2>
                  <p className="text-sm text-gray-600">Connect this project to a GitHub repository</p>
                </div>
              </div>

              {githubSuccess && (
                <div className="mb-4 p-4 bg-green-50 border-l-4 border-green-400 rounded-r-lg">
                  <div className="flex items-center">
                    <svg className="w-5 h-5 text-green-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    <span className="text-green-800 font-medium">{githubSuccess}</span>
                  </div>
                </div>
              )}

              {githubError && <FormError message={githubError} className="mb-4" />}

              <form onSubmit={handleSaveGitHub} className="space-y-4">
                <TextInput
                  label="Repository URL"
                  type="url"
                  value={githubSettings.github_repo_url}
                  onChange={(e) => setGithubSettings({ ...githubSettings, github_repo_url: e.target.value })}
                  placeholder="https://github.com/owner/repo"
                  helpText="Full URL to the GitHub repository"
                />

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <TextInput
                    label="Repository Owner"
                    type="text"
                    value={githubSettings.github_owner}
                    onChange={(e) => setGithubSettings({ ...githubSettings, github_owner: e.target.value })}
                    placeholder="username or organization"
                  />

                  <TextInput
                    label="Repository Name"
                    type="text"
                    value={githubSettings.github_repo_name}
                    onChange={(e) => setGithubSettings({ ...githubSettings, github_repo_name: e.target.value })}
                    placeholder="repository-name"
                  />
                </div>

                <TextInput
                  label="Default Branch"
                  type="text"
                  value={githubSettings.github_branch}
                  onChange={(e) => setGithubSettings({ ...githubSettings, github_branch: e.target.value })}
                  placeholder="main"
                  helpText="The default branch to track (e.g., main, master, develop)"
                />

                <div className="flex items-center gap-3 p-4 bg-gray-50 border border-gray-200 rounded-lg">
                  <input
                    type="checkbox"
                    id="sync-enabled"
                    checked={githubSettings.github_sync_enabled}
                    onChange={(e) => setGithubSettings({ ...githubSettings, github_sync_enabled: e.target.checked })}
                    className="w-4 h-4 text-primary-600 border-gray-300 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <label htmlFor="sync-enabled" className="flex-1">
                    <span className="font-medium text-gray-900">Enable GitHub Sync</span>
                    <p className="text-sm text-gray-600 mt-0.5">Automatically sync tasks with GitHub issues</p>
                  </label>
                </div>

                {githubSettings.github_last_sync && (
                  <div className="text-sm text-gray-600">
                    Last synced: {new Date(githubSettings.github_last_sync).toLocaleString()}
                  </div>
                )}

                <Button type="submit" disabled={isSavingGitHub}>
                  {isSavingGitHub ? 'Saving...' : 'Save GitHub Settings'}
                </Button>
              </form>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
