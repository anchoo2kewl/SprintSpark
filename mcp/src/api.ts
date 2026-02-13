/**
 * TaskAI REST API client.
 * Wraps fetch calls with Authorization: ApiKey header.
 */

export interface Project {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  [key: string]: unknown;
}

export interface Task {
  id: string;
  project_id: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assigned_to: string;
  created_at: string;
  updated_at: string;
  [key: string]: unknown;
}

export interface Comment {
  id: string;
  task_id: string;
  content: string;
  author_id: string;
  created_at: string;
  [key: string]: unknown;
}

export interface User {
  id: string;
  email: string;
  is_admin: boolean;
  [key: string]: unknown;
}

export class TaskAIClient {
  private baseURL: string;
  private apiKey: string;

  constructor(baseURL: string, apiKey: string) {
    // Strip trailing slash
    this.baseURL = baseURL.replace(/\/+$/, "");
    this.apiKey = apiKey;
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const res = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        Authorization: `ApiKey ${this.apiKey}`,
        ...options.headers,
      },
    });

    if (!res.ok) {
      const body = await res.text();
      throw new Error(`TaskAI API error ${res.status}: ${body}`);
    }

    return res.json() as Promise<T>;
  }

  async getMe(): Promise<User> {
    return this.request<User>("/api/me");
  }

  async listProjects(page = 1, limit = 20): Promise<{ projects: Project[]; total: number }> {
    return this.request(`/api/projects?page=${page}&limit=${limit}`);
  }

  async getProject(id: string): Promise<Project> {
    return this.request<Project>(`/api/projects/${encodeURIComponent(id)}`);
  }

  async listTasks(
    projectId: string,
    params?: { query?: string; status?: string; page?: number; limit?: number }
  ): Promise<{ tasks: Task[]; total: number }> {
    const qs = new URLSearchParams();
    if (params?.query) qs.set("query", params.query);
    if (params?.status) qs.set("status", params.status);
    if (params?.page) qs.set("page", String(params.page));
    if (params?.limit) qs.set("limit", String(params.limit));
    const suffix = qs.toString() ? `?${qs}` : "";
    return this.request(`/api/projects/${encodeURIComponent(projectId)}/tasks${suffix}`);
  }

  async createTask(
    projectId: string,
    data: { title: string; description?: string; status?: string; priority?: string; assigned_to?: string }
  ): Promise<Task> {
    return this.request<Task>(`/api/projects/${encodeURIComponent(projectId)}/tasks`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateTask(
    taskId: string,
    data: { title?: string; description?: string; status?: string; priority?: string; assigned_to?: string }
  ): Promise<Task> {
    return this.request<Task>(`/api/tasks/${encodeURIComponent(taskId)}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  }

  async listComments(taskId: string): Promise<{ comments: Comment[] }> {
    return this.request(`/api/tasks/${encodeURIComponent(taskId)}/comments`);
  }

  async addComment(taskId: string, content: string): Promise<Comment> {
    return this.request<Comment>(`/api/tasks/${encodeURIComponent(taskId)}/comments`, {
      method: "POST",
      body: JSON.stringify({ content }),
    });
  }
}
