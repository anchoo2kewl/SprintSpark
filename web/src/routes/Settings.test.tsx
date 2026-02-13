import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Settings from './Settings'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}))

vi.mock('../components/ui/FormError', () => ({
  default: ({ message }: { message: string }) => message ? <div role="alert">{message}</div> : null,
}))

// Must use vi.hoisted to define mocks that are referenced in vi.mock factories
const mocks = vi.hoisted(() => ({
  get2FAStatus: vi.fn(),
  setup2FA: vi.fn(),
  enable2FA: vi.fn(),
  disable2FA: vi.fn(),
  changePassword: vi.fn(),
  getAPIKeys: vi.fn(),
  createAPIKey: vi.fn(),
  deleteAPIKey: vi.fn(),
  getCloudinaryCredential: vi.fn(),
  saveCloudinaryCredential: vi.fn(),
  testCloudinaryConnection: vi.fn(),
  deleteCloudinaryCredential: vi.fn(),
  getMyTeam: vi.fn(),
  getTeamMembers: vi.fn(),
  getMyInvitations: vi.fn(),
  inviteTeamMember: vi.fn(),
  removeTeamMember: vi.fn(),
  acceptInvitation: vi.fn(),
  rejectInvitation: vi.fn(),
  getInvites: vi.fn(),
  createInvite: vi.fn(),
}))

vi.mock('../lib/api', () => ({
  apiClient: mocks,
}))

describe('Settings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.get2FAStatus.mockResolvedValue({ enabled: false })
    mocks.getAPIKeys.mockResolvedValue([])
    mocks.getCloudinaryCredential.mockRejectedValue(new Error('not found'))
    mocks.getMyTeam.mockResolvedValue(null)
    mocks.getTeamMembers.mockResolvedValue([])
    mocks.getMyInvitations.mockResolvedValue([])
    mocks.getInvites.mockResolvedValue({ invites: [], invite_count: 0, is_admin: false })
  })

  it('renders the settings page with password section', async () => {
    render(<Settings />)
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Change Password' })).toBeInTheDocument()
    })
  })

  describe('Password Change', () => {
    it('validates password mismatch', async () => {
      const user = userEvent.setup()
      render(<Settings />)

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Change Password' })).toBeInTheDocument()
      })

      // Use autocomplete attributes to find password fields
      const currentPw = document.querySelector('input[autocomplete="current-password"]') as HTMLInputElement
      const newPwInputs = document.querySelectorAll('input[autocomplete="new-password"]')
      const newPw = newPwInputs[0] as HTMLInputElement
      const confirmPw = newPwInputs[1] as HTMLInputElement

      await user.type(currentPw, 'oldpassword')
      await user.type(newPw, 'newpassword1')
      await user.type(confirmPw, 'differentpassword')

      const changeBtn = screen.getAllByRole('button').find(
        btn => btn.textContent?.includes('Change Password')
      )!
      await user.click(changeBtn)

      expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
      expect(mocks.changePassword).not.toHaveBeenCalled()
    })

    it('validates minimum password length', async () => {
      const user = userEvent.setup()
      render(<Settings />)

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Change Password' })).toBeInTheDocument()
      })

      const currentPw = document.querySelector('input[autocomplete="current-password"]') as HTMLInputElement
      const newPwInputs = document.querySelectorAll('input[autocomplete="new-password"]')
      const newPw = newPwInputs[0] as HTMLInputElement
      const confirmPw = newPwInputs[1] as HTMLInputElement

      await user.type(currentPw, 'old12345')
      await user.type(newPw, 'short')
      await user.type(confirmPw, 'short')

      const changeBtn = screen.getAllByRole('button').find(
        btn => btn.textContent?.includes('Change Password')
      )!
      await user.click(changeBtn)

      expect(screen.getByText('Password must be at least 8 characters')).toBeInTheDocument()
    })

    it('successfully changes password', async () => {
      mocks.changePassword.mockResolvedValue(undefined)
      const user = userEvent.setup()
      render(<Settings />)

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Change Password' })).toBeInTheDocument()
      })

      const currentPw = document.querySelector('input[autocomplete="current-password"]') as HTMLInputElement
      const newPwInputs = document.querySelectorAll('input[autocomplete="new-password"]')
      const newPw = newPwInputs[0] as HTMLInputElement
      const confirmPw = newPwInputs[1] as HTMLInputElement

      await user.type(currentPw, 'oldpassword123')
      await user.type(newPw, 'newpassword123')
      await user.type(confirmPw, 'newpassword123')

      const changeBtn = screen.getAllByRole('button').find(
        btn => btn.textContent?.includes('Change Password')
      )!
      await user.click(changeBtn)

      await waitFor(() => {
        expect(mocks.changePassword).toHaveBeenCalledWith({
          current_password: 'oldpassword123',
          new_password: 'newpassword123',
        })
        expect(screen.getByText('Password changed successfully')).toBeInTheDocument()
      })
    })

    it('shows error on password change failure', async () => {
      mocks.changePassword.mockRejectedValue(new Error('Incorrect password'))
      const user = userEvent.setup()
      render(<Settings />)

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Change Password' })).toBeInTheDocument()
      })

      const currentPw = document.querySelector('input[autocomplete="current-password"]') as HTMLInputElement
      const newPwInputs = document.querySelectorAll('input[autocomplete="new-password"]')
      const newPw = newPwInputs[0] as HTMLInputElement
      const confirmPw = newPwInputs[1] as HTMLInputElement

      await user.type(currentPw, 'wrongpassword')
      await user.type(newPw, 'newpassword123')
      await user.type(confirmPw, 'newpassword123')

      const changeBtn = screen.getAllByRole('button').find(
        btn => btn.textContent?.includes('Change Password')
      )!
      await user.click(changeBtn)

      await waitFor(() => {
        expect(screen.getByText('Incorrect password')).toBeInTheDocument()
      })
    })
  })

  describe('Two-Factor Authentication', () => {
    it('shows setup button when 2FA is disabled', async () => {
      render(<Settings />)
      await waitFor(() => {
        expect(screen.getByText('Enable 2FA')).toBeInTheDocument()
      })
    })

    it('shows disable option when 2FA is enabled', async () => {
      mocks.get2FAStatus.mockResolvedValue({ enabled: true })
      render(<Settings />)
      await waitFor(() => {
        expect(screen.getByText('Disable 2FA')).toBeInTheDocument()
      })
    })
  })

  describe('API Keys', () => {
    it('renders API keys section', async () => {
      render(<Settings />)
      await waitFor(() => {
        expect(screen.getByText('API Keys')).toBeInTheDocument()
      })
    })

    it('shows existing API keys', async () => {
      mocks.getAPIKeys.mockResolvedValue([
        { id: 1, name: 'My Key', prefix: 'sk_test', created_at: '2024-01-01T00:00:00Z', expires_at: '2024-04-01T00:00:00Z' },
      ])

      render(<Settings />)
      await waitFor(() => {
        expect(screen.getByText('My Key')).toBeInTheDocument()
      })
    })
  })
})
