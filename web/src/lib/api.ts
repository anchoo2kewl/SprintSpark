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
  due_date?: string
  sprint_id?: number
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  assignee_id?: number
  estimated_hours?: number
  actual_hours?: number
  tag_ids?: number[]
}

export interface UpdateTaskRequest {
  title?: string
  description?: string
  status?: 'todo' | 'in_progress' | 'done'
  due_date?: string | null
  sprint_id?: number | null
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  assignee_id?: number | null
  estimated_hours?: number | null
  actual_hours?: number | null
  tag_ids?: number[]
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

  // Project settings - Members
  async getProjectMembers(projectId: number): Promise<any[]> {
    return this.request<any[]>(`/api/projects/${projectId}/members`)
  }

  async addProjectMember(projectId: number, data: { email: string; role: string }): Promise<any> {
    return this.request<any>(`/api/projects/${projectId}/members`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateProjectMember(projectId: number, memberId: number, data: { role: string }): Promise<any> {
    return this.request<any>(`/api/projects/${projectId}/members/${memberId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async removeProjectMember(projectId: number, memberId: number): Promise<void> {
    return this.request<void>(`/api/projects/${projectId}/members/${memberId}`, {
      method: 'DELETE',
    })
  }

  // Project settings - GitHub
  async getProjectGitHub(projectId: number): Promise<any> {
    return this.request<any>(`/api/projects/${projectId}/github`)
  }

  async updateProjectGitHub(projectId: number, data: any): Promise<any> {
    return this.request<any>(`/api/projects/${projectId}/github`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  // Admin endpoints
  async getUsers(): Promise<any[]> {
    return this.request<any[]>('/api/admin/users')
  }

  async getUserActivity(userId: number): Promise<any[]> {
    return this.request<any[]>(`/api/admin/users/${userId}/activity`)
  }

  async updateUserAdmin(userId: number, isAdmin: boolean): Promise<any> {
    return this.request<any>(`/api/admin/users/${userId}/admin`, {
      method: 'PATCH',
      body: JSON.stringify({ is_admin: isAdmin }),
    })
  }

  // Security/Settings endpoints
  async changePassword(data: { current_password: string; new_password: string }): Promise<{ message: string }> {
    return this.request<{ message: string }>('/api/settings/password', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async get2FAStatus(): Promise<{ enabled: boolean }> {
    return this.request<{ enabled: boolean }>('/api/settings/2fa/status')
  }

  async setup2FA(): Promise<{ secret: string; qr_code_url: string; qr_code_svg: string }> {
    return this.request<{ secret: string; qr_code_url: string; qr_code_svg: string }>('/api/settings/2fa/setup', {
      method: 'POST',
    })
  }

  async enable2FA(data: { code: string }): Promise<{ backup_codes: string[] }> {
    return this.request<{ backup_codes: string[] }>('/api/settings/2fa/enable', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async disable2FA(data: { password: string }): Promise<{ message: string }> {
    return this.request<{ message: string }>('/api/settings/2fa/disable', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  // Sprint endpoints
  async getSprints(): Promise<any[]> {
    return this.request<any[]>('/api/sprints')
  }

  async createSprint(data: any): Promise<any> {
    return this.request<any>('/api/sprints', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateSprint(id: number, data: any): Promise<any> {
    return this.request<any>(`/api/sprints/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteSprint(id: number): Promise<void> {
    return this.request<void>(`/api/sprints/${id}`, {
      method: 'DELETE',
    })
  }

  // Tag endpoints
  async getTags(): Promise<any[]> {
    return this.request<any[]>('/api/tags')
  }

  async createTag(data: any): Promise<any> {
    return this.request<any>('/api/tags', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateTag(id: number, data: any): Promise<any> {
    return this.request<any>(`/api/tags/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  async deleteTag(id: number): Promise<void> {
    return this.request<void>(`/api/tags/${id}`, {
      method: 'DELETE',
    })
  }
}

// Export a singleton instance
export const api = new ApiClient(API_BASE_URL)
export const apiClient = api // Alias for consistency
