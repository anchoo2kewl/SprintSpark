import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthProvider, useAuth } from './AuthContext'
import { api } from '../lib/api'
import type { ReactNode } from 'react'

// Mock the api module
vi.mock('../lib/api', () => {
  const mockApi = {
    getToken: vi.fn(),
    setToken: vi.fn(),
    getCurrentUser: vi.fn(),
    login: vi.fn(),
    signup: vi.fn(),
    logout: vi.fn(),
  }
  return {
    api: mockApi,
    apiClient: mockApi,
    default: mockApi,
  }
})

const mockedApi = vi.mocked(api)

// Test consumer component that exposes auth context values
function TestConsumer({ onRender }: { onRender?: (ctx: ReturnType<typeof useAuth>) => void }) {
  const auth = useAuth()
  if (onRender) {
    onRender(auth)
  }
  return (
    <div>
      <span data-testid="user">{auth.user ? auth.user.email : 'null'}</span>
      <span data-testid="loading">{String(auth.loading)}</span>
      <span data-testid="error">{auth.error ?? 'null'}</span>
      <button onClick={() => auth.login({ email: 'test@example.com', password: 'password123' }).catch(() => {})}>
        Login
      </button>
      <button onClick={() => auth.signup({ email: 'new@example.com', password: 'password123' }).catch(() => {})}>
        Signup
      </button>
      <button onClick={auth.logout}>Logout</button>
      <button onClick={auth.clearError}>Clear Error</button>
    </div>
  )
}

function renderWithProvider(ui: ReactNode) {
  return render(<AuthProvider>{ui}</AuthProvider>)
}

describe('AuthProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Default: no stored token
    mockedApi.getToken.mockReturnValue(null)
  })

  it('provides auth context to children', async () => {
    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('user')).toHaveTextContent('null')
    expect(screen.getByTestId('error')).toHaveTextContent('null')
  })

  it('starts in loading state and loads user from existing token', async () => {
    mockedApi.getToken.mockReturnValue('existing-token')
    mockedApi.getCurrentUser.mockResolvedValue({
      id: 1,
      email: 'existing@example.com',
      is_admin: false,
      created_at: '2024-01-01T00:00:00Z',
    })

    renderWithProvider(<TestConsumer />)

    // Initially loading
    expect(screen.getByTestId('loading')).toHaveTextContent('true')

    // After loading completes
    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('user')).toHaveTextContent('existing@example.com')
    expect(mockedApi.getCurrentUser).toHaveBeenCalledTimes(1)
  })

  it('clears invalid token on mount when getCurrentUser fails', async () => {
    mockedApi.getToken.mockReturnValue('bad-token')
    mockedApi.getCurrentUser.mockRejectedValue(new Error('Unauthorized'))

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('user')).toHaveTextContent('null')
    expect(mockedApi.logout).toHaveBeenCalledTimes(1)
  })

  it('does not call getCurrentUser when no token exists', async () => {
    mockedApi.getToken.mockReturnValue(null)

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(mockedApi.getCurrentUser).not.toHaveBeenCalled()
  })
})

describe('login', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedApi.getToken.mockReturnValue(null)
  })

  it('logs in successfully and sets user', async () => {
    const user = userEvent.setup()
    mockedApi.login.mockResolvedValue({
      token: 'new-token',
      user: { id: 1, email: 'test@example.com', is_admin: false, created_at: '2024-01-01T00:00:00Z' },
    })

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await user.click(screen.getByText('Login'))

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
    expect(screen.getByTestId('error')).toHaveTextContent('null')
    expect(mockedApi.login).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'password123',
    })
  })

  it('sets error on login failure', async () => {
    const user = userEvent.setup()
    mockedApi.login.mockRejectedValue(new Error('Invalid credentials'))

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await user.click(screen.getByText('Login'))

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('error')).toHaveTextContent('Invalid credentials')
    expect(screen.getByTestId('user')).toHaveTextContent('null')
  })

  it('handles non-Error thrown values on login failure', async () => {
    const user = userEvent.setup()
    mockedApi.login.mockRejectedValue('string error')

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await user.click(screen.getByText('Login'))

    await waitFor(() => {
      expect(screen.getByTestId('error')).toHaveTextContent('Login failed')
    })
  })
})

describe('signup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedApi.getToken.mockReturnValue(null)
  })

  it('signs up successfully and sets user', async () => {
    const user = userEvent.setup()
    mockedApi.signup.mockResolvedValue({
      token: 'signup-token',
      user: { id: 2, email: 'new@example.com', is_admin: false, created_at: '2024-01-01T00:00:00Z' },
    })

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await user.click(screen.getByText('Signup'))

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('user')).toHaveTextContent('new@example.com')
    expect(screen.getByTestId('error')).toHaveTextContent('null')
    expect(mockedApi.signup).toHaveBeenCalledWith({
      email: 'new@example.com',
      password: 'password123',
    })
  })

  it('sets error on signup failure', async () => {
    const user = userEvent.setup()
    mockedApi.signup.mockRejectedValue(new Error('Email already taken'))

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await user.click(screen.getByText('Signup'))

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    expect(screen.getByTestId('error')).toHaveTextContent('Email already taken')
    expect(screen.getByTestId('user')).toHaveTextContent('null')
  })

  it('handles non-Error thrown values on signup failure', async () => {
    const user = userEvent.setup()
    mockedApi.signup.mockRejectedValue({ code: 'CONFLICT' })

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await user.click(screen.getByText('Signup'))

    await waitFor(() => {
      expect(screen.getByTestId('error')).toHaveTextContent('Signup failed')
    })
  })
})

describe('logout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedApi.getToken.mockReturnValue(null)
  })

  it('clears user and error on logout', async () => {
    const user = userEvent.setup()

    // First, log in
    mockedApi.login.mockResolvedValue({
      token: 'token',
      user: { id: 1, email: 'test@example.com', is_admin: false, created_at: '2024-01-01T00:00:00Z' },
    })

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    await user.click(screen.getByText('Login'))

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
    })

    // Now log out
    await user.click(screen.getByText('Logout'))

    expect(screen.getByTestId('user')).toHaveTextContent('null')
    expect(screen.getByTestId('error')).toHaveTextContent('null')
    expect(mockedApi.logout).toHaveBeenCalled()
  })
})

describe('clearError', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedApi.getToken.mockReturnValue(null)
  })

  it('clears the error state', async () => {
    const user = userEvent.setup()
    mockedApi.login.mockRejectedValue(new Error('Bad request'))

    renderWithProvider(<TestConsumer />)

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })

    // Trigger an error
    await user.click(screen.getByText('Login'))

    await waitFor(() => {
      expect(screen.getByTestId('error')).toHaveTextContent('Bad request')
    })

    // Clear the error
    await user.click(screen.getByText('Clear Error'))

    expect(screen.getByTestId('error')).toHaveTextContent('null')
  })
})

describe('useAuth', () => {
  it('throws when used outside AuthProvider', () => {
    // Suppress console.error for expected error
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => {
      render(<TestConsumer />)
    }).toThrow('useAuth must be used within an AuthProvider')

    consoleSpy.mockRestore()
  })

  it('returns all expected context properties', async () => {
    mockedApi.getToken.mockReturnValue(null)
    let capturedCtx: ReturnType<typeof useAuth> | null = null

    renderWithProvider(
      <TestConsumer onRender={(ctx) => { capturedCtx = ctx }} />
    )

    await waitFor(() => {
      expect(capturedCtx).not.toBeNull()
    })

    expect(capturedCtx).toHaveProperty('user')
    expect(capturedCtx).toHaveProperty('loading')
    expect(capturedCtx).toHaveProperty('error')
    expect(capturedCtx).toHaveProperty('login')
    expect(capturedCtx).toHaveProperty('signup')
    expect(capturedCtx).toHaveProperty('logout')
    expect(capturedCtx).toHaveProperty('clearError')
    expect(typeof capturedCtx!.login).toBe('function')
    expect(typeof capturedCtx!.signup).toBe('function')
    expect(typeof capturedCtx!.logout).toBe('function')
    expect(typeof capturedCtx!.clearError).toBe('function')
  })
})
