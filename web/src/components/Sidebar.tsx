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
      <div className="w-60 bg-dark-bg-secondary border-r border-dark-bg-tertiary/20 p-3">
        <div className="animate-pulse space-y-2">
          <div className="h-8 bg-dark-bg-tertiary/40 rounded"></div>
          <div className="h-12 bg-dark-bg-tertiary/30 rounded"></div>
          <div className="h-12 bg-dark-bg-tertiary/30 rounded"></div>
          <div className="h-12 bg-dark-bg-tertiary/30 rounded"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="w-60 bg-dark-bg-secondary border-r border-dark-bg-tertiary/20 flex flex-col">
      {/* Header */}
      <div className="p-3 border-b border-dark-bg-tertiary/20 space-y-1.5">
        <button
          onClick={onCreateProject}
          className="w-full bg-primary-500 hover:bg-primary-600 text-white font-medium py-1.5 px-3 rounded-md transition-colors duration-150 flex items-center justify-center gap-2 text-sm"
        >
          <svg
            className="w-4 h-4"
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

        {/* Sprints Link */}
        <button
          onClick={() => navigate('/app/sprints')}
          className={`w-full font-medium py-1.5 px-3 rounded-md transition-colors duration-150 flex items-center gap-2 text-sm ${
            location.pathname === '/app/sprints'
              ? 'bg-dark-bg-tertiary/50 text-dark-text-primary'
              : 'text-dark-text-secondary hover:bg-dark-bg-tertiary/30 hover:text-dark-text-primary'
          }`}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          Sprints
        </button>

        {/* Tags Link */}
        <button
          onClick={() => navigate('/app/tags')}
          className={`w-full font-medium py-1.5 px-3 rounded-md transition-colors duration-150 flex items-center gap-2 text-sm ${
            location.pathname === '/app/tags'
              ? 'bg-dark-bg-tertiary/50 text-dark-text-primary'
              : 'text-dark-text-secondary hover:bg-dark-bg-tertiary/30 hover:text-dark-text-primary'
          }`}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
          </svg>
          Tags
        </button>

        {/* Settings Link */}
        <button
          onClick={() => navigate('/app/settings')}
          className={`w-full font-medium py-1.5 px-3 rounded-md transition-colors duration-150 flex items-center gap-2 text-sm ${
            location.pathname === '/app/settings'
              ? 'bg-dark-bg-tertiary/50 text-dark-text-primary'
              : 'text-dark-text-secondary hover:bg-dark-bg-tertiary/30 hover:text-dark-text-primary'
          }`}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Settings
        </button>

        {/* Admin Link */}
        {user?.is_admin && (
          <button
            onClick={() => navigate('/app/admin')}
            className={`w-full font-medium py-1.5 px-3 rounded-md transition-colors duration-150 flex items-center gap-2 text-sm ${
              location.pathname === '/app/admin'
                ? 'bg-primary-500/20 text-primary-400'
                : 'text-dark-text-secondary hover:bg-dark-bg-tertiary/30 hover:text-dark-text-primary'
            }`}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            Admin
          </button>
        )}
      </div>

      {/* Projects List */}
      <div className="flex-1 overflow-y-auto p-3">
        {error && (
          <div className="bg-danger-500/10 border border-danger-500/20 text-danger-400 px-3 py-2 rounded text-xs">
            {error}
          </div>
        )}

        {projects.length === 0 && !error && (
          <div className="text-center py-8">
            <svg
              className="mx-auto h-10 w-10 text-dark-text-tertiary"
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
            <p className="mt-2 text-xs text-dark-text-secondary">No projects yet</p>
            <p className="text-xs text-dark-text-tertiary mt-1">
              Create your first project to get started
            </p>
          </div>
        )}

        <div className="space-y-1">
          {projects.map((project) => {
            const isSelected = selectedProjectId === String(project.id)
            return (
              <button
                key={project.id}
                onClick={() => handleProjectClick(project.id)}
                className={`w-full text-left py-2 px-3 rounded-md transition-all duration-150 ${
                  isSelected
                    ? 'bg-primary-500/10 text-dark-text-primary border border-primary-500/30'
                    : 'text-dark-text-secondary hover:bg-dark-bg-tertiary/30 hover:text-dark-text-primary border border-transparent'
                }`}
              >
                <h3
                  className={`font-medium text-xs truncate ${
                    isSelected ? 'text-dark-text-primary' : 'text-dark-text-secondary'
                  }`}
                >
                  {project.name}
                </h3>
                {project.description && (
                  <p className="text-xs text-dark-text-tertiary truncate mt-0.5">
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
