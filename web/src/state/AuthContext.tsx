import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { api, User, SignupRequest, LoginRequest } from '../lib/api'

interface AuthContextType {
  user: User | null
  loading: boolean
  error: string | null
  login: (data: LoginRequest) => Promise<void>
  signup: (data: SignupRequest & { invite_code?: string }) => Promise<void>
  logout: () => void
  clearError: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Load user on mount if token exists
  useEffect(() => {
    const loadUser = async () => {
      const token = api.getToken()
      if (token) {
        try {
          const currentUser = await api.getCurrentUser()
          setUser(currentUser)
        } catch (err) {
          // Token is invalid, clear it
          api.logout()
        }
      }
      setLoading(false)
    }

    loadUser()
  }, [])

  const login = async (data: LoginRequest) => {
    try {
      setError(null)
      setLoading(true)
      const response = await api.login(data)
      setUser(response.user || null)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed'
      setError(message)
      throw err
    } finally {
      setLoading(false)
    }
  }

  const signup = async (data: SignupRequest & { invite_code?: string }) => {
    try {
      setError(null)
      setLoading(true)
      const response = await api.signup(data)
      setUser(response.user || null)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Signup failed'
      setError(message)
      throw err
    } finally {
      setLoading(false)
    }
  }

  const logout = () => {
    api.logout()
    setUser(null)
    setError(null)
  }

  const clearError = () => {
    setError(null)
  }

  const value = {
    user,
    loading,
    error,
    login,
    signup,
    logout,
    clearError,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
