import { useState } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../state/AuthContext'
import Sidebar from '../components/Sidebar'
import ProjectModal from '../components/ProjectModal'
import SyncStatus from '../components/SyncStatus'
import CommandPalette from '../components/CommandPalette'
import { Project } from '../lib/api'

export default function Dashboard() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [isProjectModalOpen, setIsProjectModalOpen] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const handleProjectCreated = (project: Project) => {
    // Add project to sidebar via window callback
    if ((window as any).__addProject) {
      ;(window as any).__addProject(project)
    }
    // Navigate to the new project
    navigate(`/app/projects/${project.id}`)
  }

  return (
    <div className="min-h-screen bg-dark-bg-primary flex flex-col">
      {/* Header - Linear style */}
      <header className="bg-dark-bg-secondary border-b border-dark-bg-tertiary/20">
        <div className="flex items-center justify-between h-12 px-4">
          <div className="flex items-center gap-3">
            <img
              src="/logo.svg"
              alt="SprintSpark"
              className="w-6 h-6"
            />
            <h1 className="text-sm font-semibold text-dark-text-primary">SprintSpark</h1>
          </div>

          <div className="flex items-center gap-3">
            <SyncStatus />
            <div className="flex items-center gap-2">
              <div className="w-6 h-6 bg-primary-500/10 rounded-full flex items-center justify-center">
                <span className="text-xs font-medium text-primary-400">
                  {user?.email?.charAt(0).toUpperCase()}
                </span>
              </div>
              <span className="text-xs text-dark-text-secondary">{user?.email}</span>
            </div>
            <button
              onClick={handleLogout}
              className="inline-flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium text-dark-text-secondary hover:text-dark-text-primary hover:bg-dark-bg-tertiary/30 rounded-md transition-colors duration-150"
            >
              <svg
                className="w-3.5 h-3.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
                />
              </svg>
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Main Layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <Sidebar onCreateProject={() => setIsProjectModalOpen(true)} />

        {/* Main Content */}
        <main className="flex-1 overflow-y-auto bg-dark-bg-primary">
          <Outlet />
        </main>
      </div>

      {/* Project Modal */}
      <ProjectModal
        isOpen={isProjectModalOpen}
        onClose={() => setIsProjectModalOpen(false)}
        onProjectCreated={handleProjectCreated}
      />

      {/* Command Palette (Cmd+K) */}
      <CommandPalette />
    </div>
  )
}
