import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Signup from './Signup'

// Mock react-router-dom
const mockNavigate = vi.fn()
let mockSearchParams = new URLSearchParams()
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams],
  Link: ({ to, children }: { to: string; children: React.ReactNode }) => <a href={to}>{children}</a>,
}))

// Mock the api module
vi.mock('../lib/api', () => ({
  apiClient: { validateInvite: vi.fn() },
  api: { validateInvite: vi.fn() },
}))

// Mock useAuth
const mockSignup = vi.fn()
const mockClearError = vi.fn()
let mockAuthState = {
  user: null as { id: number; email: string; is_admin: boolean; created_at: string } | null,
  error: null as string | null,
  loading: false,
  login: vi.fn(),
  clearError: mockClearError,
  signup: mockSignup,
  logout: vi.fn(),
}

vi.mock('../state/AuthContext', () => ({
  useAuth: () => mockAuthState,
}))

describe('Signup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockSearchParams = new URLSearchParams()
    mockAuthState = {
      user: null,
      error: null,
      loading: false,
      login: vi.fn(),
      clearError: mockClearError,
      signup: mockSignup,
      logout: vi.fn(),
    }
  })

  it('renders signup form with all fields', () => {
    render(<Signup />)

    expect(screen.getByLabelText(/invite code/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/email address/i)).toBeInTheDocument()
    // Password fields share similar label text; use the id-based placeholder to distinguish
    expect(screen.getByPlaceholderText('Paste your invite code')).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
    expect(screen.getByText(/create your account/i)).toBeInTheDocument()
    expect(screen.getByText(/already have an account/i)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /sign in/i })).toHaveAttribute('href', '/login')
  })

  it('shows invite code from URL search params', async () => {
    mockSearchParams = new URLSearchParams('code=ABC123')
    render(<Signup />)

    await waitFor(() => {
      expect(screen.getByLabelText(/invite code/i)).toHaveValue('ABC123')
    })
  })

  it('shows validation errors on empty submit', async () => {
    render(<Signup />)

    // Use fireEvent.submit to bypass jsdom's native required validation
    const form = screen.getByRole('button', { name: /create account/i }).closest('form')!
    fireEvent.submit(form)

    await waitFor(() => {
      expect(screen.getByText(/email is required/i)).toBeInTheDocument()
    })
    expect(screen.getByText(/password is required/i)).toBeInTheDocument()
    expect(mockSignup).not.toHaveBeenCalled()
  })

  it('calls signup on valid form submit', async () => {
    const user = userEvent.setup()
    mockSignup.mockResolvedValue(undefined)
    render(<Signup />)

    const inviteInput = screen.getByLabelText(/invite code/i)
    const emailInput = screen.getByLabelText(/email address/i)
    // Use the specific input ids to target password fields
    const passwordInput = document.getElementById('password')!
    const confirmPasswordInput = document.getElementById('confirm-password')!

    await user.type(inviteInput, 'VALID-CODE')
    await user.type(emailInput, 'new@example.com')
    await user.type(passwordInput, 'password1')
    await user.type(confirmPasswordInput, 'password1')
    await user.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => {
      expect(mockSignup).toHaveBeenCalledWith({
        email: 'new@example.com',
        password: 'password1',
        invite_code: 'VALID-CODE',
      })
    })
    expect(mockClearError).toHaveBeenCalled()
  })

  it('shows invite code required error when missing', async () => {
    const user = userEvent.setup()
    render(<Signup />)

    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = document.getElementById('password')!
    const confirmPasswordInput = document.getElementById('confirm-password')!

    // Fill in valid form fields but leave invite code empty
    await user.type(emailInput, 'new@example.com')
    await user.type(passwordInput, 'password1')
    await user.type(confirmPasswordInput, 'password1')

    // Use fireEvent.submit to bypass jsdom's native required validation on invite code
    const form = screen.getByRole('button', { name: /create account/i }).closest('form')!
    fireEvent.submit(form)

    await waitFor(() => {
      expect(screen.getByText(/invite code is required/i)).toBeInTheDocument()
    })
    expect(mockSignup).not.toHaveBeenCalled()
  })

  it('redirects when user is already logged in', async () => {
    mockAuthState.user = {
      id: 1,
      email: 'test@example.com',
      is_admin: false,
      created_at: '2024-01-01T00:00:00Z',
    }

    render(<Signup />)

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/app', { replace: true })
    })
  })
})
