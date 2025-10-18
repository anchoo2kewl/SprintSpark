import { useEffect, useState } from 'react'
import { useNavigate, useParams, useLocation } from 'react-router-dom'
import { api, Project } from '../lib/api'
import { useAuth } from '../state/AuthContext'

interface SidebarProps {
  onCreateProject: () => void
}

export default function Sidebar({ onCreateProject }: SidebarProps) {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()
  const location = useLocation()
  const { user } = useAuth()
  const { id: selectedProjectId } = useParams<{ id: string }>()

  useEffect(() => {
    loadProjects()
  }, [])

  const loadProjects = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await api.getProjects()
      setProjects(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects')
    } finally {
      setLoading(false)
    }
  }

  const handleProjectClick = (projectId: number | undefined) => {
    if (projectId) {
      navigate(`/app/projects/${projectId}`)
    }
  }

  const addProject = (project: Project) => {
    setProjects([project, ...projects])
  }

  // Expose addProject method to parent via callback
  useEffect(() => {
    ;(window as any).__addProject = addProject
    return () => {
      delete (window as any).__addProject
    }
  }, [projects])

  if (loading) {
    return (
      <div className="w-64 bg-white border-r border-gray-200 p-4">
        <div className="animate-pulse space-y-3">
          <div className="h-10 bg-gray-200 rounded"></div>
          <div className="h-16 bg-gray-100 rounded"></div>
          <div className="h-16 bg-gray-100 rounded"></div>
          <div className="h-16 bg-gray-100 rounded"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="w-64 bg-white border-r border-gray-200 flex flex-col">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 space-y-2">
        <button
          onClick={onCreateProject}
          className="w-full bg-primary-600 hover:bg-primary-700 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 flex items-center justify-center gap-2"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          New Project
        </button>

        {/* Admin Link */}
        {user?.is_admin && (
          <button
            onClick={() => navigate('/app/admin')}
            className={`w-full font-medium py-2 px-4 rounded-lg transition-colors duration-200 flex items-center justify-center gap-2 ${
              location.pathname === '/app/admin'
                ? 'bg-purple-600 text-white'
                : 'bg-purple-50 text-purple-700 hover:bg-purple-100'
            }`}
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            Admin
          </button>
        )}
      </div>

      {/* Projects List */}
      <div className="flex-1 overflow-y-auto p-4">
        {error && (
          <div className="bg-danger-50 border border-danger-200 text-danger-700 px-3 py-2 rounded mb-4 text-sm">
            {error}
          </div>
        )}

        {projects.length === 0 && !error && (
          <div className="text-center py-8">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 13h6m-3-3v6m5 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <p className="mt-2 text-sm text-gray-600">No projects yet</p>
            <p className="text-xs text-gray-500 mt-1">
              Create your first project to get started
            </p>
          </div>
        )}

        <div className="space-y-2">
          {projects.map((project) => {
            const isSelected = selectedProjectId === String(project.id)
            return (
              <button
                key={project.id}
                onClick={() => handleProjectClick(project.id)}
                className={`w-full text-left p-3 rounded-lg transition-all duration-200 ${
                  isSelected
                    ? 'bg-primary-50 border-2 border-primary-500 shadow-sm'
                    : 'bg-gray-50 hover:bg-gray-100 border-2 border-transparent'
                }`}
              >
                <h3
                  className={`font-medium text-sm truncate ${
                    isSelected ? 'text-primary-900' : 'text-gray-900'
                  }`}
                >
                  {project.name}
                </h3>
                {project.description && (
                  <p className="text-xs text-gray-500 truncate mt-1">
                    {project.description}
                  </p>
                )}
              </button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
