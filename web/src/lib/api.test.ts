import { describe, it, expect, beforeEach, vi } from 'vitest'

// We need to mock fetch before importing the API client
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
Object.defineProperty(global, 'localStorage', { value: localStorageMock })

// Dynamic import type
type ApiModule = typeof import('./api')
let apiClient: ApiModule['apiClient']

beforeEach(async () => {
  vi.resetModules()
  mockFetch.mockReset()
  localStorageMock.getItem.mockReturnValue('test-token')

  // Re-import to get fresh instance
  const mod: ApiModule = await import('./api')
  apiClient = mod.apiClient
})

function mockResponse(data: unknown, status = 200) {
  mockFetch.mockResolvedValueOnce({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(data),
  })
}

function mockErrorResponse(error: string, status = 400) {
  mockFetch.mockResolvedValueOnce({
    ok: false,
    status,
    json: () => Promise.resolve({ error }),
  })
}

describe('ApiClient', () => {
  describe('getAssets', () => {
    it('calls correct URL with no params', async () => {
      mockResponse([])
      await apiClient.getAssets()

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/assets'),
        expect.any(Object)
      )
      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toMatch(/\/api\/assets$/)
    })

    it('passes query params correctly', async () => {
      mockResponse([])
      await apiClient.getAssets({ q: 'photo', type: 'image', limit: 20 })

      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toContain('q=photo')
      expect(url).toContain('type=image')
      expect(url).toContain('limit=20')
    })

    it('URL encodes search query', async () => {
      mockResponse([])
      await apiClient.getAssets({ q: 'hello world' })

      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toContain('q=hello+world')
    })

    it('passes offset param', async () => {
      mockResponse([])
      await apiClient.getAssets({ offset: 10 })

      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toContain('offset=10')
    })
  })

  describe('deleteAttachment', () => {
    it('sends DELETE to correct path', async () => {
      mockResponse(undefined, 204)
      // deleteAttachment returns void (204), but our client handles it
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: () => Promise.resolve({}),
      })
      mockFetch.mockReset()
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: () => Promise.resolve({}),
      })

      await apiClient.deleteAttachment(5)

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/attachments/5'),
        expect.objectContaining({ method: 'DELETE' })
      )
    })
  })

  describe('updateAttachment', () => {
    it('sends PATCH with alt_name data', async () => {
      mockResponse({ id: 5, alt_name: 'new name' })
      await apiClient.updateAttachment(5, { alt_name: 'new name' })

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/attachments/5'),
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify({ alt_name: 'new name' }),
        })
      )
    })
  })

  describe('getStorageUsage', () => {
    it('calls correct project storage URL', async () => {
      mockResponse([{ user_id: 1, file_count: 5, total_size: 1024 }])
      await apiClient.getStorageUsage(42)

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/projects/42/storage'),
        expect.any(Object)
      )
    })
  })

  describe('getImages', () => {
    it('calls /api/images with no query', async () => {
      mockResponse([])
      await apiClient.getImages()

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/images'),
        expect.any(Object)
      )
      const url = mockFetch.mock.calls[0][0] as string
      expect(url).not.toContain('?')
    })

    it('passes search query', async () => {
      mockResponse([])
      await apiClient.getImages('sunset')

      const url = mockFetch.mock.calls[0][0] as string
      expect(url).toContain('?q=sunset')
    })
  })

  describe('error handling', () => {
    it('throws on non-ok response', async () => {
      mockErrorResponse('Not authorized', 401)

      await expect(apiClient.getAssets()).rejects.toThrow('Not authorized')
    })

    it('throws generic error for non-json errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: () => Promise.resolve({ error: 'Internal server error' }),
      })

      await expect(apiClient.getAssets()).rejects.toThrow('Internal server error')
    })
  })

  describe('authorization', () => {
    it('includes Bearer token in requests', async () => {
      mockResponse([])
      await apiClient.getAssets()

      const config = mockFetch.mock.calls[0][1] as RequestInit
      expect((config.headers as Record<string, string>)['Authorization']).toContain('Bearer')
    })
  })
})
