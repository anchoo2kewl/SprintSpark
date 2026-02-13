import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import ProtectedRoute from './ProtectedRoute'

// Mock useAuth from AuthContext
const mockUseAuth = vi.fn()
vi.mock('../state/AuthContext', () => ({
  useAuth: () => mockUseAuth(),
}))

// Mock Navigate from react-router-dom
vi.mock('react-router-dom', () => ({
  Navigate: ({ to, replace }: { to: string; replace?: boolean }) => (
    <div data-testid="navigate" data-to={to} data-replace={String(replace)} />
  ),
}))

describe('ProtectedRoute', () => {
  it('shows loading spinner when loading is true', () => {
    mockUseAuth.mockReturnValue({ user: null, loading: true })

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    )

    expect(screen.getByText('Loading...')).toBeInTheDocument()
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
  })

  it('redirects to /login when there is no user', () => {
    mockUseAuth.mockReturnValue({ user: null, loading: false })

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    )

    const navigate = screen.getByTestId('navigate')
    expect(navigate).toBeInTheDocument()
    expect(navigate).toHaveAttribute('data-to', '/login')
    expect(navigate).toHaveAttribute('data-replace', 'true')
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
  })

  it('renders children when user exists', () => {
    mockUseAuth.mockReturnValue({
      user: { id: 1, name: 'Test User', email: 'test@example.com' },
      loading: false,
    })

    render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>
    )

    expect(screen.getByText('Protected Content')).toBeInTheDocument()
    expect(screen.queryByTestId('navigate')).not.toBeInTheDocument()
    expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
  })
})
