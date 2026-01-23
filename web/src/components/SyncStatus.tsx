/**
 * Sync Status Component
 * Shows real-time sync status in the UI
 */

import { useSync } from '../state/SyncContext'

export default function SyncStatus() {
  const { syncState, triggerSync } = useSync()

  const getStatusIcon = () => {
    switch (syncState.status) {
      case 'syncing':
        return (
          <svg className="animate-spin h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )
      case 'synced':
        return (
          <svg className="h-4 w-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        )
      case 'error':
        return (
          <svg className="h-4 w-4 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        )
      case 'offline':
        return (
          <svg className="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414"
            />
          </svg>
        )
      default:
        return (
          <svg className="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
        )
    }
  }

  const getStatusText = () => {
    switch (syncState.status) {
      case 'syncing':
        return 'Syncing...'
      case 'synced':
        return 'Synced'
      case 'error':
        return 'Sync Error'
      case 'offline':
        return 'Offline'
      default:
        return 'Ready'
    }
  }

  const getStatusColor = () => {
    switch (syncState.status) {
      case 'syncing':
        return 'text-blue-600 bg-blue-50 hover:bg-blue-100'
      case 'synced':
        return 'text-green-600 bg-green-50 hover:bg-green-100'
      case 'error':
        return 'text-red-600 bg-red-50 hover:bg-red-100'
      case 'offline':
        return 'text-gray-600 bg-gray-50 hover:bg-gray-100'
      default:
        return 'text-gray-600 bg-gray-50 hover:bg-gray-100'
    }
  }

  const formatLastSync = () => {
    if (!syncState.lastSyncTime) return null

    const elapsed = Date.now() - syncState.lastSyncTime
    const seconds = Math.floor(elapsed / 1000)
    const minutes = Math.floor(seconds / 60)

    if (minutes > 0) {
      return `${minutes}m ago`
    } else if (seconds > 0) {
      return `${seconds}s ago`
    } else {
      return 'just now'
    }
  }

  return (
    <button
      onClick={triggerSync}
      disabled={syncState.status === 'syncing'}
      className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors duration-200 ${getStatusColor()} ${
        syncState.status === 'syncing' ? 'cursor-not-allowed' : 'cursor-pointer'
      }`}
      title={
        syncState.error
          ? syncState.error
          : syncState.lastSyncTime
          ? `Last synced ${formatLastSync()}`
          : 'Click to sync'
      }
    >
      {getStatusIcon()}
      <span>{getStatusText()}</span>
      {syncState.pendingOperations > 0 && (
        <span className="inline-flex items-center justify-center px-1.5 py-0.5 text-xs font-bold leading-none text-white bg-blue-600 rounded-full">
          {syncState.pendingOperations}
        </span>
      )}
      {syncState.lastSyncTime && syncState.status === 'synced' && (
        <span className="text-gray-500">{formatLastSync()}</span>
      )}
    </button>
  )
}
