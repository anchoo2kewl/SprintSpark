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
  adminBoostInvites: vi.fn(),
  getEmailProvider: vi.fn(),
  saveEmailProvider: vi.fn(),
  deleteEmailProvider: vi.fn(),
  testEmailProvider: vi.fn(),
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
    invite_count: 5,
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
    invite_count: 3,
  },
]

describe('Admin', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    authState.user = { email: 'admin@test.com', is_admin: true }
    apiMocks.getUsers.mockResolvedValue(users)
    apiMocks.getUserActivity.mockResolvedValue([])
    apiMocks.getEmailProvider.mockResolvedValue(null)
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

  it('renders user list after loading', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Admin Dashboard')).toBeInTheDocument()
    })

    expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    expect(screen.getByText('user@test.com')).toBeInTheDocument()
    expect(screen.getByText('Users (2)')).toBeInTheDocument()
  })

  it('shows admin toggle switches for each user', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    // Each user row has an admin toggle switch
    const switches = screen.getAllByRole('switch')
    expect(switches.length).toBe(2)

    // First user (admin) should have checked toggle
    expect(switches[0]).toHaveAttribute('aria-checked', 'true')
    // Second user (non-admin) should have unchecked toggle
    expect(switches[1]).toHaveAttribute('aria-checked', 'false')
  })

  it('shows invite count in user rows', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('5')).toBeInTheDocument()
      expect(screen.getByText('3')).toBeInTheDocument()
    })
  })

  it('shows error state on API failure', async () => {
    apiMocks.getUsers.mockRejectedValue(new Error('Server error'))

    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Server error')).toBeInTheDocument()
    })
  })

  it('expands user row to show details when clicked', async () => {
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

    // Click the user row to expand
    await user.click(screen.getByText('admin@test.com'))

    await waitFor(() => {
      expect(apiMocks.getUserActivity).toHaveBeenCalledWith(1)
    })

    // Expanded panel shows stats and activity
    await waitFor(() => {
      expect(screen.getByText('Logins')).toBeInTheDocument()
      expect(screen.getByText('10')).toBeInTheDocument()
      expect(screen.getByText('Failed Attempts')).toBeInTheDocument()
      expect(screen.getByText('login')).toBeInTheDocument()
      expect(screen.getByText('10.0.0.1')).toBeInTheDocument()
    })
  })

  it('shows empty activity message in expanded panel', async () => {
    apiMocks.getUserActivity.mockResolvedValue([])

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    // Click to expand
    await user.click(screen.getByText('admin@test.com'))

    await waitFor(() => {
      expect(screen.getByText('No activity recorded')).toBeInTheDocument()
    })
  })

  it('shows N/A for missing IP address in expanded panel', async () => {
    apiMocks.getUserActivity.mockResolvedValue([])

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('user@test.com')).toBeInTheDocument()
    })

    // Click the second user (has no IP)
    await user.click(screen.getByText('user@test.com'))

    await waitFor(() => {
      expect(screen.getByText('N/A')).toBeInTheDocument()
      expect(screen.getByText('Never')).toBeInTheDocument()
    })
  })

  it('toggles admin status via toggle switch', async () => {
    apiMocks.updateUserAdmin.mockResolvedValue(undefined)

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    // Click the toggle for the non-admin user (second switch)
    const switches = screen.getAllByRole('switch')
    await user.click(switches[1])

    expect(apiMocks.updateUserAdmin).toHaveBeenCalledWith(2, true)
  })

  it('shows tabs for Users and Email Provider', async () => {
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Users (2)')).toBeInTheDocument()
      expect(screen.getByText('Email Provider')).toBeInTheDocument()
    })
  })

  it('shows invite count editor in expanded panel', async () => {
    apiMocks.getUserActivity.mockResolvedValue([])

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    // Expand first user
    await user.click(screen.getByText('admin@test.com'))

    await waitFor(() => {
      expect(screen.getByText('Invite count:')).toBeInTheDocument()
    })
  })

  it('shows 192.168.1.1 in expanded panel for admin user', async () => {
    apiMocks.getUserActivity.mockResolvedValue([])

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('admin@test.com')).toBeInTheDocument()
    })

    await user.click(screen.getByText('admin@test.com'))

    await waitFor(() => {
      expect(screen.getByText('192.168.1.1')).toBeInTheDocument()
    })
  })
})

describe('Admin - Email Provider Tab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    authState.user = { email: 'admin@test.com', is_admin: true }
    apiMocks.getUsers.mockResolvedValue(users)
    apiMocks.getUserActivity.mockResolvedValue([])
    apiMocks.getEmailProvider.mockResolvedValue(null)
  })

  it('shows email provider form when tab is clicked', async () => {
    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Email Provider')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Email Provider'))

    await waitFor(() => {
      expect(screen.getByText('Email Provider (Brevo)')).toBeInTheDocument()
      expect(screen.getByText('Save Provider')).toBeInTheDocument()
    })
  })

  it('shows existing provider status when configured', async () => {
    apiMocks.getEmailProvider.mockResolvedValue({
      id: 1,
      provider: 'brevo',
      api_key: 'xkey****7890',
      sender_email: 'noreply@taskai.cc',
      sender_name: 'TaskAI',
      status: 'connected',
      last_checked_at: null,
      last_error: '',
      consecutive_failures: 0,
    })

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Email Provider')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Email Provider'))

    await waitFor(() => {
      expect(screen.getByText('connected')).toBeInTheDocument()
      expect(screen.getByText('TaskAI')).toBeInTheDocument()
      expect(screen.getByText('Test Connection')).toBeInTheDocument()
      expect(screen.getByText('Remove')).toBeInTheDocument()
      expect(screen.getByText('Update Provider')).toBeInTheDocument()
    })
  })

  it('handles test connection click', async () => {
    apiMocks.getEmailProvider.mockResolvedValue({
      id: 1,
      provider: 'brevo',
      api_key: 'xkey****7890',
      sender_email: 'noreply@taskai.cc',
      sender_name: 'TaskAI',
      status: 'connected',
      last_checked_at: null,
      last_error: '',
      consecutive_failures: 0,
    })
    apiMocks.testEmailProvider.mockResolvedValue({
      id: 1,
      provider: 'brevo',
      api_key: 'xkey****7890',
      sender_email: 'noreply@taskai.cc',
      sender_name: 'TaskAI',
      status: 'connected',
      last_checked_at: '2024-06-01T12:00:00Z',
      last_error: '',
      consecutive_failures: 0,
    })

    const user = userEvent.setup()
    render(<Admin />)

    await waitFor(() => {
      expect(screen.getByText('Email Provider')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Email Provider'))

    await waitFor(() => {
      expect(screen.getByText('Test Connection')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Test Connection'))

    await waitFor(() => {
      expect(apiMocks.testEmailProvider).toHaveBeenCalled()
      expect(screen.getByText('Connection successful')).toBeInTheDocument()
    })
  })
})
