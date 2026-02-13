import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
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

function App() {
  return (
    <AuthProvider>
      <SyncProvider>
        <BrowserRouter>
        <Routes>
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
      </BrowserRouter>
      </SyncProvider>
    </AuthProvider>
  )
}

export default App
