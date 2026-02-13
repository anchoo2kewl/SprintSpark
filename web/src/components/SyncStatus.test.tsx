import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import SyncStatus from './SyncStatus'

const mockTriggerSync = vi.fn()
let mockSyncState = {
  status: 'idle' as string,
  lastSyncTime: null as number | null,
  error: null as string | null,
  pendingOperations: 0,
}

vi.mock('../state/SyncContext', () => ({
  useSync: () => ({ syncState: mockSyncState, triggerSync: mockTriggerSync }),
}))

describe('SyncStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockSyncState = {
      status: 'idle',
      lastSyncTime: null,
      error: null,
      pendingOperations: 0,
    }
  })

  it('shows "Syncing..." text when syncing', () => {
    mockSyncState.status = 'syncing'
    render(<SyncStatus />)
    expect(screen.getByText('Syncing...')).toBeInTheDocument()
  })

  it('shows "Synced" text when synced', () => {
    mockSyncState.status = 'synced'
    render(<SyncStatus />)
    expect(screen.getByText('Synced')).toBeInTheDocument()
  })

  it('shows "Sync Error" text when error', () => {
    mockSyncState.status = 'error'
    mockSyncState.error = 'Connection failed'
    render(<SyncStatus />)
    expect(screen.getByText('Sync Error')).toBeInTheDocument()
  })

  it('shows "Offline" text when offline', () => {
    mockSyncState.status = 'offline'
    render(<SyncStatus />)
    expect(screen.getByText('Offline')).toBeInTheDocument()
  })

  it('shows "Ready" text for idle/default status', () => {
    mockSyncState.status = 'idle'
    render(<SyncStatus />)
    expect(screen.getByText('Ready')).toBeInTheDocument()
  })

  it('shows pending operations badge when pendingOperations > 0', () => {
    mockSyncState.pendingOperations = 5
    render(<SyncStatus />)
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('does not show pending operations badge when pendingOperations is 0', () => {
    mockSyncState.pendingOperations = 0
    render(<SyncStatus />)
    // The badge should not be rendered at all
    const button = screen.getByRole('button')
    const badge = button.querySelector('.rounded-full')
    expect(badge).not.toBeInTheDocument()
  })

  it('shows "just now" when synced and lastSyncTime is very recent', () => {
    mockSyncState.status = 'synced'
    mockSyncState.lastSyncTime = Date.now()
    render(<SyncStatus />)
    expect(screen.getByText('just now')).toBeInTheDocument()
  })

  it('shows seconds ago when synced and lastSyncTime is seconds old', () => {
    mockSyncState.status = 'synced'
    mockSyncState.lastSyncTime = Date.now() - 30_000 // 30 seconds ago
    render(<SyncStatus />)
    expect(screen.getByText('30s ago')).toBeInTheDocument()
  })

  it('shows minutes ago when synced and lastSyncTime is minutes old', () => {
    mockSyncState.status = 'synced'
    mockSyncState.lastSyncTime = Date.now() - 180_000 // 3 minutes ago
    render(<SyncStatus />)
    expect(screen.getByText('3m ago')).toBeInTheDocument()
  })

  it('does not show last sync time when status is not synced', () => {
    mockSyncState.status = 'idle'
    mockSyncState.lastSyncTime = Date.now() - 60_000
    render(<SyncStatus />)
    // The formatted time span is only rendered when status === 'synced'
    expect(screen.queryByText(/ago$/)).not.toBeInTheDocument()
    expect(screen.queryByText('just now')).not.toBeInTheDocument()
  })

  it('disables button when syncing', () => {
    mockSyncState.status = 'syncing'
    render(<SyncStatus />)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('enables button when not syncing', () => {
    mockSyncState.status = 'synced'
    render(<SyncStatus />)
    expect(screen.getByRole('button')).toBeEnabled()
  })

  it('calls triggerSync on click', async () => {
    const user = userEvent.setup()
    mockSyncState.status = 'synced'
    render(<SyncStatus />)

    await user.click(screen.getByRole('button'))
    expect(mockTriggerSync).toHaveBeenCalledTimes(1)
  })

  it('does not call triggerSync when button is disabled (syncing)', async () => {
    const user = userEvent.setup()
    mockSyncState.status = 'syncing'
    render(<SyncStatus />)

    await user.click(screen.getByRole('button'))
    expect(mockTriggerSync).not.toHaveBeenCalled()
  })

  it('shows error message in title when error exists', () => {
    mockSyncState.status = 'error'
    mockSyncState.error = 'Connection timeout'
    render(<SyncStatus />)
    expect(screen.getByRole('button')).toHaveAttribute('title', 'Connection timeout')
  })

  it('shows last sync info in title when lastSyncTime exists and no error', () => {
    mockSyncState.status = 'synced'
    mockSyncState.lastSyncTime = Date.now()
    render(<SyncStatus />)
    expect(screen.getByRole('button')).toHaveAttribute('title', 'Last synced just now')
  })

  it('shows "Click to sync" in title when no error and no lastSyncTime', () => {
    mockSyncState.status = 'idle'
    render(<SyncStatus />)
    expect(screen.getByRole('button')).toHaveAttribute('title', 'Click to sync')
  })
})
