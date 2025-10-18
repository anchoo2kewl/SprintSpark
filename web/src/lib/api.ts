// Import generated types from OpenAPI spec
import type { components, operations } from './api.types'

// Re-export types for convenience
export type User = components['schemas']['User']
export type AuthResponse = components['schemas']['AuthResponse']
export type SignupRequest = components['schemas']['SignupRequest']
export type LoginRequest = components['schemas']['LoginRequest']
export type Project = components['schemas']['Project']
export type Task = components['schemas']['Task']
export type ApiError = components['schemas']['Error']

// Request types (not in OpenAPI spec yet, so define them)
export interface CreateProjectRequest {
  name: string
  description?: string
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
}

export interface CreateTaskRequest {
  title: string
  description?: string
  status?: 'todo' | 'in_progress' | 'done'
}

export interface UpdateTaskRequest {
  title?: string
  description?: string
  status?: 'todo' | 'in_progress' | 'done'
}

// Helper types for API responses (using available operations)
type SignupResponse = operations['signup']['responses']['201']['content']['application/json']
type LoginResponse = operations['login']['responses']['200']['content']['application/json']
type GetCurrentUserResponse = operations['getCurrentUser']['responses']['200']['content']['application/json']

// API Client Configuration
// Use relative URL in production (served behind nginx proxy)
// or VITE_API_URL for development override
const API_BASE_URL = import.meta.env.VITE_API_URL || (
  import.meta.env.PROD ? '' : 'http://localhost:8080'
)

class ApiClient {
  private baseURL: string
  private token: string | null = null

  constructor(baseURL: string) {
    this.baseURL = baseURL
    // Load token from localStorage on initialization
    this.token = localStorage.getItem('auth_token')
  }

  setToken(token: string | null) {
    this.token = token
    if (token) {
      localStorage.setItem('auth_token', token)
    } else {
      localStorage.removeItem('auth_token')
    }
  }

  getToken(): string | null {
    return this.token
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    // Add existing headers
    if (options.headers) {
      Object.assign(headers, options.headers)
    }

    // Add authorization header if token exists
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const config: RequestInit = {
      ...options,
      headers,
    }

    try {
      const response = await fetch(url, config)

      // Handle non-JSON responses (like 204 No Content)
      if (response.status === 204) {
        return {} as T
      }

      const data = await response.json()

      if (!response.ok) {
        // Handle API errors
        const error = data as ApiError
        throw new Error(error.error || `HTTP ${response.status}`)
      }

      return data as T
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('An unexpected error occurred')
    }
  }

  // Auth endpoints
  async signup(data: SignupRequest): Promise<SignupResponse> {
    const response = await this.request<SignupResponse>('/api/auth/signup', {
      method: 'POST',
      body: JSON.stringify(data),
    })
    if (response.token) {
      this.setToken(response.token)
    }
    return response
  }

  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    })
    if (response.token) {
      this.setToken(response.token)
    }
    return response
  }

  logout(): void {
    this.setToken(null)
  }

  async getCurrentUser(): Promise<GetCurrentUserResponse> {
    return this.request<GetCurrentUserResponse>('/api/me')
  }

  // Health check
  async healthCheck(): Promise<{ status: string; database?: string }> {
    return this.request('/healthz')
  }

  // Project endpoints
  async getProjects(): Promise<Project[]> {
    return this.request<Project[]>('/api/projects')
  }

  async getProject(id: number): Promise<Project> {
    return this.request<Project>('/api/projects/' + id)
  }

  async createProject(data: CreateProjectRequest): Promise<Project> {
    return this.request<Project>('/api/projects', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateProject(id: number, data: UpdateProjectRequest): Promise<Project> {
    return this.request<Project>('/api/projects/' + id, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteProject(id: number): Promise<void> {
    return this.request<void>('/api/projects/' + id, {
      method: 'DELETE',
    })
  }

  // Task endpoints
  async getTasks(projectId: number): Promise<Task[]> {
    return this.request<Task[]>('/api/projects/' + projectId + '/tasks')
  }

  async createTask(projectId: number, data: CreateTaskRequest): Promise<Task> {
    return this.request<Task>('/api/projects/' + projectId + '/tasks', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateTask(id: number, data: UpdateTaskRequest): Promise<Task> {
    return this.request<Task>('/api/tasks/' + id, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteTask(id: number): Promise<void> {
    return this.request<void>('/api/tasks/' + id, {
      method: 'DELETE',
    })
  }
}

// Export a singleton instance
export const api = new ApiClient(API_BASE_URL)
