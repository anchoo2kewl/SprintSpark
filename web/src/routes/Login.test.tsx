import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Login from './Login'

// Mock react-router-dom
const mockNavigate = vi.fn()
const mockSearchParams = new URLSearchParams()
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams],
  Link: ({ to, children }: { to: string; children: React.ReactNode }) => <a href={to}>{children}</a>,
}))

// Mock useAuth
const mockLogin = vi.fn()
const mockClearError = vi.fn()
let mockAuthState = {
  user: null as { id: number; email: string; is_admin: boolean; created_at: string } | null,
  error: null as string | null,
  loading: false,
  login: mockLogin,
  clearError: mockClearError,
  signup: vi.fn(),
  logout: vi.fn(),
}

vi.mock('../state/AuthContext', () => ({
  useAuth: () => mockAuthState,
}))

describe('Login', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAuthState = {
      user: null,
      error: null,
      loading: false,
      login: mockLogin,
      clearError: mockClearError,
      signup: vi.fn(),
      logout: vi.fn(),
    }
  })

  it('renders login form with email and password fields', () => {
    render(<Login />)

    expect(screen.getByLabelText(/email address/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
    expect(screen.getByText(/sign in to taskai/i)).toBeInTheDocument()
    expect(screen.getByText(/don't have an account/i)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /sign up/i })).toHaveAttribute('href', '/signup')
  })

  it('shows validation errors on empty submit', async () => {
    render(<Login />)

    // Use fireEvent.submit to bypass jsdom's native required validation
    const form = screen.getByRole('button', { name: /sign in/i }).closest('form')!
    fireEvent.submit(form)

    await waitFor(() => {
      expect(screen.getByText(/email is required/i)).toBeInTheDocument()
    })
    expect(screen.getByText(/password is required/i)).toBeInTheDocument()
    expect(mockLogin).not.toHaveBeenCalled()
  })

  it('calls login on valid form submit', async () => {
    const user = userEvent.setup()
    mockLogin.mockResolvedValue(undefined)
    render(<Login />)

    await user.type(screen.getByLabelText(/email address/i), 'test@example.com')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      })
    })
    expect(mockClearError).toHaveBeenCalled()
  })

  it('shows server error from auth context', () => {
    mockAuthState.error = 'Invalid credentials'
    render(<Login />)

    expect(screen.getByRole('alert')).toHaveTextContent('Invalid credentials')
  })

  it('redirects when user is already logged in', async () => {
    mockAuthState.user = {
      id: 1,
      email: 'test@example.com',
      is_admin: false,
      created_at: '2024-01-01T00:00:00Z',
    }

    render(<Login />)

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/app', { replace: true })
    })
  })

  it('shows loading state during submission', () => {
    mockAuthState.loading = true
    render(<Login />)

    expect(screen.getByRole('button', { name: /sign in/i })).toBeDisabled()
    expect(screen.getByLabelText(/email address/i)).toBeDisabled()
    expect(screen.getByLabelText(/password/i)).toBeDisabled()
  })
})
