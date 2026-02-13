import { BrowserRouter, Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { AuthProvider } from './state/AuthContext'
import { SyncProvider } from './state/SyncContext'
import ProtectedRoute from './components/ProtectedRoute'
import Landing from './routes/Landing'
import Login from './routes/Login'
import Signup from './routes/Signup'
import Dashboard from './routes/Dashboard'
import Projects from './routes/Projects'
import ProjectDetail from './routes/ProjectDetail'
import ProjectSettings from './routes/ProjectSettings'
import TaskDetail from './routes/TaskDetail'
import Sprints from './routes/Sprints'
import Tags from './routes/Tags'
import Admin from './routes/Admin'
import Settings from './routes/Settings'

function AppRoutes() {
  const location = useLocation()
  const bgLocation = (location.state as any)?.backgroundLocation as Location | undefined

  return (
    <>
      <Routes location={bgLocation || location}>
        {/* Public routes */}
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />

        {/* Protected routes */}
        <Route
          path="/app"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        >
          <Route index element={<Projects />} />
          <Route path="projects/:projectId" element={<ProjectDetail />} />
          <Route path="projects/:projectId/settings" element={<ProjectSettings />} />
          <Route path="projects/:projectId/tasks/:taskId" element={<TaskDetail />} />
          <Route path="sprints" element={<Sprints />} />
          <Route path="tags" element={<Tags />} />
          <Route path="admin" element={<Admin />} />
          <Route path="settings" element={<Settings />} />
        </Route>

        {/* Catch-all redirect to landing */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>

      {/* Task detail modal overlay when opened from project board */}
      {bgLocation && (
        <Routes>
          <Route
            path="/app/projects/:projectId/tasks/:taskId"
            element={<TaskDetailModal />}
          />
        </Routes>
      )}
    </>
  )
}

function TaskDetailModal() {
  const navigate = useNavigate()

  const handleClose = () => {
    navigate(-1)
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/60 backdrop-blur-sm overflow-y-auto"
      onClick={handleClose}
    >
      <div
        className="w-full max-w-5xl my-0 md:my-8 mx-0 md:mx-4 bg-dark-bg-primary md:rounded-xl overflow-hidden shadow-2xl border-0 md:border border-dark-border-subtle min-h-screen md:min-h-0"
        onClick={(e) => e.stopPropagation()}
      >
        <TaskDetail isModal onClose={handleClose} />
      </div>
    </div>
  )
}

function App() {
  return (
    <AuthProvider>
      <SyncProvider>
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
      </SyncProvider>
    </AuthProvider>
  )
}

export default App
