import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Admin from './Admin'

// Mock react-router-dom
const mockNavigate = vi.fn()
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}))

// Mock API with vi.hoisted
const apiMocks = vi.hoisted(() => ({
  getUsers: vi.fn(),
  getUserActivity: vi.fn(),
  updateUserAdmin: vi.fn(),
}))

vi.mock('../lib/api', () => ({
  api: apiMocks,
}))

// Mock useAuth with vi.hoisted
const authState = vi.hoisted(() => ({
  user: null as { email: string; is_admin: boolean } | null,
}))

vi.mock('../state/AuthContext', () => ({
  useAuth: () => ({ user: authState.user }),
}))

const users = [
  {
    id: 1,
    email: 'admin@test.com',
    is_admin: true,
    created_at: '2024-01-01T00:00:00Z',
    login_count: 10,
    last_login_at: '2024-06-01T12:00:00Z',
    last_login_ip: '192.168.1.1',
    failed_attempts: 0,
  },
  {
    id: 2,
    email: 'user@test.com',
    is_admin: false,
    created_at: '2024-02-01T00:00:00Z',
    login_count: 3,
    last_login_at: null,
    last_login_ip: null,
    failed_attempts: 5,
  },
]

describe('Admin', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    authState.user = { email: 'admin@test.com', is_admin: true }
    apiMocks.getUsers.mockResolvedValue(users)
    apiMocks.getUserActivity.mockResolvedValue([])
  })

  it('redirects non-admin users to /app', () => {
    authState.user = { email: 'user@test.com', is_admin: false }
    apiMocks.getUsers.mockReturnValue(new Promise(() => {}))
    render(<Admin />)
    expect(mockNavigate).toHaveBeenCalledWith('/app')
  })

  it('redirects when user is null', () => {
    authState.user = null
    apiMocks.getUsers.mockReturnValue(new Promise(() => {}))
    render(<Admin />)
    expect(mockNavigate).toHaveBeenCalledWith('/app')
  })

  it('shows loading skeleton initially', () => {
    apiMocks.getUsers.mockReturnValue(new Promise(() => {}))
    render(<Admin />)
    expect(screen.queryByText('Admin Dashboard')).not.toBeInTheDocument()
  })

  it('renders user table after loading', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Admin Dashboard')).toBeInTheDocument()
    })

    expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    expect(screen.getByText('user@test.com')).toBeInTheDocument()
    expect(screen.getByText('Users (2)')).toBeInTheDocument()
  })

  it('shows admin/user badges correctly', async () => {
    render(<Admin />)

    await waitFor(() => {
      // "Admin" appears in table header and badge — verify both badge types exist
      const adminTexts = screen.getAllByText('Admin')
      expect(adminTexts.length).toBeGreaterThanOrEqual(2) // header + badge
      expect(screen.getByText('User')).toBeInTheDocument()
    })
  })

  it('displays login count and failed attempts', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('10')).toBeInTheDocument()
      expect(screen.getByText('5')).toBeInTheDocument()
    })
  })

  it('shows N/A for missing IP address', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('192.168.1.1')).toBeInTheDocument()
      expect(screen.getByText('N/A')).toBeInTheDocument()
    })
  })

  it('shows error state on API failure', async () => {
    apiMocks.getUsers.mockRejectedValue(new Error('Server error'))

    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Server error')).toBeInTheDocument()
    })
  })

  it('loads user activity when clicking Activity button', async () => {
    const activities = [
      {
        id: 1,
        user_id: 1,
        activity_type: 'login',
        ip_address: '10.0.0.1',
        user_agent: 'Mozilla/5.0',
        created_at: '2024-06-01T12:00:00Z',
      },
    ]
    apiMocks.getUserActivity.mockResolvedValue(activities)

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    const activityButtons = screen.getAllByText('Activity')
    await user.click(activityButtons[0])

    await waitFor(() => {
      expect(apiMocks.getUserActivity).toHaveBeenCalledWith(1)
      expect(screen.getByText('LOGIN')).toBeInTheDocument()
      expect(screen.getByText(/10\.0\.0\.1/)).toBeInTheDocument()
    })
  })

  it('shows empty activity message', async () => {
    apiMocks.getUserActivity.mockResolvedValue([])

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    const activityButtons = screen.getAllByText('Activity')
    await user.click(activityButtons[0])

    await waitFor(() => {
      expect(screen.getByText('No activity recorded for this user')).toBeInTheDocument()
    })
  })

  it('shows select user prompt before any activity selected', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Select a user to view their activity')).toBeInTheDocument()
    })
  })

  it('toggles admin status', async () => {
    apiMocks.updateUserAdmin.mockResolvedValue(undefined)

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    const makeAdminButton = screen.getByText('Make Admin')
    await user.click(makeAdminButton)

    expect(apiMocks.updateUserAdmin).toHaveBeenCalledWith(2, true)
  })

  it('shows Revoke Admin for admin users', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Revoke Admin')).toBeInTheDocument()
      expect(screen.getByText('Make Admin')).toBeInTheDocument()
    })
  })
})
