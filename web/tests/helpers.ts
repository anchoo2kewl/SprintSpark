import { Page, expect } from '@playwright/test';

// API base URL
const API_BASE_URL = process.env.VITE_API_URL || 'http://localhost:8080';

/**
 * Generate random test user credentials
 */
export function generateTestUser() {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 10000);
  return {
    email: 'test.user.' + timestamp + '.' + random + '@example.com',
    password: 'TestPass123!' + random,
  };
}

/**
 * Generate random project data
 */
export function generateProject() {
  const timestamp = Date.now();
  const isoDate = new Date().toISOString();
  return {
    name: 'Test Project ' + timestamp,
    description: 'This is a test project created at ' + isoDate,
  };
}

/**
 * Generate random task data
 */
export function generateTask() {
  const timestamp = Date.now();
  const isoDate = new Date().toISOString();
  return {
    title: 'Test Task ' + timestamp,
    description: 'Task created at ' + isoDate,
    status: 'todo' as const,
  };
}

/**
 * Sign up a new user
 */
export async function signupUser(page: Page, email: string, password: string) {
  await page.goto('/signup');
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);
  await page.fill('input[name="confirm-password"]', password);
  await page.click('button[type="submit"]');

  // Wait for redirect to dashboard
  await page.waitForURL('/app', { timeout: 10000 });
}

/**
 * Login with existing user
 */
export async function loginUser(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);
  await page.click('button[type="submit"]');

  // Wait for redirect to dashboard
  await page.waitForURL('/app', { timeout: 10000 });
}

/**
 * Logout current user
 */
export async function logoutUser(page: Page) {
  await page.click('button:has-text("Logout")');
  await page.waitForURL('/login', { timeout: 5000 });
}

/**
 * Create a project via UI
 */
export async function createProjectUI(page: Page, name: string, description?: string) {
  // Click "New Project" button in sidebar
  await page.click('button:has-text("New Project")');

  // Wait for modal
  await page.waitForSelector('input[name="name"]', { timeout: 5000 });

  // Fill form
  await page.fill('input[name="name"]', name);
  if (description) {
    await page.fill('textarea[name="description"]', description);
  }

  // Submit
  await page.click('button[type="submit"]:has-text("Create Project")');

  // Wait for modal to close
  await page.waitForSelector('input[name="name"]', { state: 'hidden', timeout: 5000 });
}

/**
 * Verify project appears in sidebar
 */
export async function verifyProjectInSidebar(page: Page, projectName: string) {
  const selector = 'button:has-text("' + projectName + '")';
  const projectButton = page.locator(selector).first();
  await expect(projectButton).toBeVisible({ timeout: 5000 });
  return projectButton;
}

/**
 * Click project in sidebar to navigate to it
 */
export async function navigateToProject(page: Page, projectName: string) {
  const projectButton = await verifyProjectInSidebar(page, projectName);
  await projectButton.click();

  // Wait for project detail page to load
  const headerSelector = 'h1:has-text("' + projectName + '")';
  await page.waitForSelector(headerSelector, { timeout: 5000 });
}

/**
 * Verify API state by making direct API call
 */
export async function verifyAPIState(request: any, endpoint: string, _expectedData: any) {
  const token = await request.storageState().then((state: any) => {
    return state.cookies.find((c: any) => c.name === 'auth_token')?.value;
  });

  const response = await request.get(`${API_BASE_URL}${endpoint}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });

  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  return data;
}

/**
 * Get auth token from localStorage via page
 */
export async function getAuthToken(page: Page): Promise<string | null> {
  return await page.evaluate(() => localStorage.getItem('auth_token'));
}

/**
 * Make authenticated API request from browser context
 */
export async function apiRequest(
  page: Page,
  method: string,
  endpoint: string,
  body?: any
): Promise<any> {
  const result = await page.evaluate(
    async ({ method, endpoint, body, apiUrl }) => {
      const token = localStorage.getItem('auth_token');

      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(`${apiUrl}${endpoint}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`API request failed: ${response.status} ${error}`);
      }

      if (response.status === 204) {
        return null;
      }

      return await response.json();
    },
    { method, endpoint, body, apiUrl: API_BASE_URL }
  );

  return result;
}

/**
 * Verify projects exist via API
 */
export async function verifyProjectsAPI(page: Page): Promise<any[]> {
  return await apiRequest(page, 'GET', '/api/projects');
}

/**
 * Verify tasks for a project via API
 */
export async function verifyTasksAPI(page: Page, projectId: number): Promise<any[]> {
  const endpoint = '/api/projects/' + projectId + '/tasks';
  return await apiRequest(page, 'GET', endpoint);
}

/**
 * Create a task via UI
 */
export async function createTaskUI(page: Page, title: string, description?: string) {
  // Click "New Task" button
  await page.click('button:has-text("New Task")');

  // Wait for modal to appear
  await page.waitForSelector('text=Create New Task');

  // Fill in task title
  await page.fill('#task-title', title);

  // Fill in description if provided
  if (description) {
    await page.fill('#task-description', description);
  }

  // Click create button
  await page.click('button:has-text("Create Task")');

  // Wait for modal to close
  await page.waitForSelector('text=Create New Task', { state: 'hidden', timeout: 5000 });
}

/**
 * Verify task appears in the board
 */
export async function verifyTaskInBoard(page: Page, taskTitle: string) {
  await expect(page.locator(`h4:has-text("${taskTitle}")`)).toBeVisible();
}

/**
 * Drag a task to a new status column (simulates drag and drop)
 */
export async function dragTaskToStatus(page: Page, taskTitle: string, targetStatus: 'todo' | 'in_progress' | 'done') {
  // Find the task card
  const taskCard = page.locator(`h4:has-text("${taskTitle}")`).locator('..')

  // Get task card position
  const taskBox = await taskCard.boundingBox()
  if (!taskBox) throw new Error('Task card not found')

  // Find target column
  const statusColumnMap = {
    'todo': 'To Do',
    'in_progress': 'In Progress',
    'done': 'Done'
  }
  const targetColumn = page.locator(`h3:has-text("${statusColumnMap[targetStatus]}")`).locator('..')
  const columnBox = await targetColumn.boundingBox()
  if (!columnBox) throw new Error('Target column not found')

  // Perform drag
  await taskCard.hover()
  await page.mouse.down()
  await page.mouse.move(columnBox.x + columnBox.width / 2, columnBox.y + 100, { steps: 10 })
  await page.mouse.up()

  // Wait for animation/API call
  await page.waitForTimeout(500)
}

/**
 * Click on a task to open detail modal
 */
export async function openTaskDetail(page: Page, taskTitle: string) {
  await page.locator(`h4:has-text("${taskTitle}")`).click()
  await page.waitForSelector('text=Edit Task')
}

/**
 * Update task via detail modal
 */
export async function updateTaskViaModal(
  page: Page,
  updates: { title?: string, description?: string, status?: string, dueDate?: string }
) {
  if (updates.title !== undefined) {
    await page.fill('#edit-title', updates.title)
  }
  if (updates.description !== undefined) {
    await page.fill('#edit-description', updates.description)
  }
  if (updates.status !== undefined) {
    await page.selectOption('#edit-status', updates.status)
  }
  if (updates.dueDate !== undefined) {
    await page.fill('#edit-due-date', updates.dueDate)
  }

  await page.click('button:has-text("Save Changes")')
  await page.waitForSelector('text=Edit Task', { state: 'hidden', timeout: 5000 })
}

/**
 * Verify task status badge
 */
export async function verifyTaskStatus(page: Page, taskTitle: string, expectedStatus: 'To Do' | 'In Progress' | 'Done') {
  const taskCard = page.locator(`h4:has-text("${taskTitle}")`).locator('..')
  await expect(taskCard.locator(`text="${expectedStatus}"`)).toBeVisible()
}

/**
 * Verify task has due date
 */
export async function verifyTaskDueDate(page: Page, taskTitle: string) {
  const taskCard = page.locator(`h4:has-text("${taskTitle}")`).locator('..')
  await expect(taskCard.locator('text=/📅/')).toBeVisible()
}

/**
 * Wait for element and take screenshot for debugging
 */
export async function waitAndScreenshot(
  page: Page,
  selector: string,
  name: string,
  options?: { timeout?: number }
) {
  try {
    await page.waitForSelector(selector, { timeout: options?.timeout || 5000 });
  } catch (error) {
    const screenshotPath = 'test-results/' + name + '-error.png';
    await page.screenshot({ path: screenshotPath, fullPage: true });
    throw error;
  }
}
