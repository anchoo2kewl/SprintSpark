import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useEffect } from 'react'
import { SyncProvider, useSync } from './SyncContext'

let mockUser: { id: number; email: string } | null = null

vi.mock('./AuthContext', () => ({
  useAuth: () => ({ user: mockUser }),
}))

// Test consumer component that exposes sync context values
function TestConsumer() {
  const { syncState, isInitialized, triggerSync } = useSync()
  return (
    <div>
      <span data-testid="status">{syncState.status}</span>
      <span data-testid="initialized">{String(isInitialized)}</span>
      <span data-testid="error">{syncState.error ?? 'null'}</span>
      <span data-testid="pending">{syncState.pendingOperations}</span>
      <span data-testid="lastSync">{syncState.lastSyncTime ? 'has-time' : 'null'}</span>
      <button onClick={triggerSync}>Sync</button>
    </div>
  )
}

function renderWithProvider() {
  return render(
    <SyncProvider>
      <TestConsumer />
    </SyncProvider>
  )
}

const mockRefreshTask = vi.fn()

function RegisteredTaskConsumer() {
  const { syncState, registerSyncTask, triggerSync } = useSync()

  useEffect(() => {
    return registerSyncTask('test-refresh', mockRefreshTask)
  }, [registerSyncTask])

  return (
    <div>
      <span data-testid="status">{syncState.status}</span>
      <span data-testid="pending">{syncState.pendingOperations}</span>
      <button onClick={triggerSync}>Sync</button>
    </div>
  )
}

function renderRegisteredTaskProvider() {
  return render(
    <SyncProvider>
      <RegisteredTaskConsumer />
    </SyncProvider>
  )
}

describe('SyncProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUser = null
    mockRefreshTask.mockResolvedValue(undefined)
  })

  it('provides context to children', () => {
    renderWithProvider()

    expect(screen.getByTestId('status')).toBeInTheDocument()
    expect(screen.getByTestId('initialized')).toBeInTheDocument()
    expect(screen.getByTestId('error')).toBeInTheDocument()
    expect(screen.getByTestId('pending')).toBeInTheDocument()
  })

  it('shows idle state when no user is logged in', () => {
    mockUser = null
    renderWithProvider()

    expect(screen.getByTestId('status')).toHaveTextContent('idle')
    expect(screen.getByTestId('initialized')).toHaveTextContent('false')
    expect(screen.getByTestId('error')).toHaveTextContent('null')
    expect(screen.getByTestId('pending')).toHaveTextContent('0')
    expect(screen.getByTestId('lastSync')).toHaveTextContent('null')
  })

  it('initializes when user logs in (auto-init)', async () => {
    mockUser = { id: 1, email: 'test@example.com' }
    renderWithProvider()

    // The useEffect auto-initializes when user is present
    await waitFor(() => {
      expect(screen.getByTestId('initialized')).toHaveTextContent('true')
    })

    expect(screen.getByTestId('status')).toHaveTextContent('synced')
    expect(screen.getByTestId('error')).toHaveTextContent('null')
    expect(screen.getByTestId('pending')).toHaveTextContent('0')
    expect(screen.getByTestId('lastSync')).toHaveTextContent('has-time')
  })

  it('triggerSync stays safe when no refresh tasks are registered', async () => {
    const user = userEvent.setup()
    mockUser = { id: 1, email: 'test@example.com' }
    renderWithProvider()

    await waitFor(() => {
      expect(screen.getByTestId('initialized')).toHaveTextContent('true')
    })

    await user.click(screen.getByText('Sync'))

    expect(screen.getByTestId('status')).toHaveTextContent('synced')
  })

  it('runs registered refresh tasks when triggered', async () => {
    const user = userEvent.setup()
    mockUser = { id: 1, email: 'test@example.com' }
    renderRegisteredTaskProvider()

    await waitFor(() => {
      expect(screen.getByTestId('status')).toHaveTextContent('synced')
    })

    await user.click(screen.getByText('Sync'))

    await waitFor(() => {
      expect(mockRefreshTask).toHaveBeenCalledTimes(1)
      expect(screen.getByTestId('status')).toHaveTextContent('synced')
      expect(screen.getByTestId('pending')).toHaveTextContent('0')
    })
  })
})

describe('useSync', () => {
  it('throws when used outside SyncProvider', () => {
    // Suppress console.error for expected error
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => {
      render(<TestConsumer />)
    }).toThrow('useSync must be used within a SyncProvider')

    consoleSpy.mockRestore()
  })
})
