import { test, expect } from '@playwright/test';

test.describe('Task Detail Page', () => {
  let authToken: string;
  let projectId: string;
  let taskId: string;

  test.beforeEach(async ({ request }) => {
    // Create a unique user for this test
    const email = `test-${Date.now()}@example.com`;
    const password = 'password123';

    // Signup
    const signupResponse = await request.post('http://localhost:8083/api/auth/signup', {
      data: {
        email,
        password,
        name: 'Test User'
      }
    });
    const signupData = await signupResponse.json();
    authToken = signupData.token;

    // Create a project
    const projectResponse = await request.post('http://localhost:8083/api/projects', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      },
      data: {
        name: 'Test Project',
        description: 'Test Description'
      }
    });
    const projectData = await projectResponse.json();
    projectId = projectData.id;

    // Create a task with markdown description
    const taskResponse = await request.post(`http://localhost:8083/api/projects/${projectId}/tasks`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      },
      data: {
        title: 'Test Task',
        description: `# Task Description

## Overview
This is a **test task** with *markdown* formatting.

## Features
- Bullet point 1
- Bullet point 2
- Bullet point 3

## Code Example
\`\`\`javascript
console.log('Hello, World!');
\`\`\`

[Link to example](https://example.com)`,
        status: 'todo',
        priority: 'high',
        estimated_hours: 5,
        actual_hours: 3
      }
    });
    const taskData = await taskResponse.json();
    taskId = taskData.id;
  });

  test('should display task title and metadata', async ({ page }) => {
    // Set auth token in local storage
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail page
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Wait for page to load
    await page.waitForSelector('h1');

    // Check title
    await expect(page.locator('h1')).toContainText('Test Task');

    // Check status badge
    await expect(page.locator('text=To Do')).toBeVisible();

    // Check priority badge
    await expect(page.locator('text=high')).toBeVisible();
  });

  test('should render markdown description correctly', async ({ page }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Wait for description section
    await page.waitForSelector('.prose');

    // Check markdown rendering
    await expect(page.locator('.prose h1')).toContainText('Task Description');
    await expect(page.locator('.prose h2').first()).toContainText('Overview');

    // Check bold and italic
    await expect(page.locator('.prose strong')).toContainText('test task');
    await expect(page.locator('.prose em')).toContainText('markdown');

    // Check list items
    await expect(page.locator('.prose li')).toHaveCount(3);
    await expect(page.locator('.prose li').first()).toContainText('Bullet point 1');

    // Check code block
    await expect(page.locator('.prose pre')).toContainText("console.log('Hello, World!')");

    // Check link
    await expect(page.locator('.prose a[href="https://example.com"]')).toBeVisible();
  });

  test('should display metadata in right sidebar', async ({ page }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Wait for sidebar
    await page.waitForSelector('text=Estimated Hours');

    // Check metadata fields
    await expect(page.locator('text=5h')).toBeVisible(); // Estimated hours
    await expect(page.locator('text=3h')).toBeVisible(); // Actual hours
    await expect(page.locator('text=Status')).toBeVisible();
    await expect(page.locator('text=Priority')).toBeVisible();
  });

  test('should allow editing task', async ({ page }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Click Edit button
    await page.click('button:has-text("Edit")');

    // Edit title
    const titleInput = page.locator('input[placeholder="Task title"]');
    await titleInput.fill('Updated Task Title');

    // Edit description
    const descriptionTextarea = page.locator('textarea[placeholder*="markdown"]');
    await descriptionTextarea.fill('# Updated Description\n\nThis is the updated content.');

    // Change status
    await page.selectOption('select', { label: 'In Progress' });

    // Change priority
    await page.locator('select').nth(1).selectOption('urgent');

    // Save changes
    await page.click('button:has-text("Save")');

    // Wait for save to complete
    await page.waitForTimeout(1000);

    // Verify changes
    await expect(page.locator('h1')).toContainText('Updated Task Title');
    await expect(page.locator('.prose h1')).toContainText('Updated Description');
    await expect(page.locator('text=In Progress')).toBeVisible();
    await expect(page.locator('text=urgent')).toBeVisible();
  });

  test('should allow canceling edit', async ({ page }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Click Edit button
    await page.click('button:has-text("Edit")');

    // Make changes
    const titleInput = page.locator('input[placeholder="Task title"]');
    await titleInput.fill('This should not be saved');

    // Cancel
    await page.click('button:has-text("Cancel")');

    // Verify original content is preserved
    await expect(page.locator('h1')).toContainText('Test Task');
    await expect(page.locator('h1')).not.toContainText('This should not be saved');
  });

  test('should delete task and redirect to project', async ({ page, request }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Setup dialog handler
    page.on('dialog', dialog => dialog.accept());

    // Click Delete button
    await page.click('button:has-text("Delete")');

    // Wait for navigation
    await page.waitForURL(`**/app/projects/${projectId}`);

    // Verify we're back at project page
    expect(page.url()).toContain(`/app/projects/${projectId}`);

    // Verify task is deleted via API
    const response = await request.get(`http://localhost:8083/api/projects/${projectId}/tasks`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const tasks = await response.json();
    expect(tasks).toHaveLength(0);
  });

  test('should navigate back to project', async ({ page }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Click back button
    await page.click('button:has-text("Back to project")');

    // Wait for navigation
    await page.waitForURL(`**/app/projects/${projectId}`);

    // Verify we're at project page
    expect(page.url()).toContain(`/app/projects/${projectId}`);
  });

  test('should update time tracking fields', async ({ page }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to task detail
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/${taskId}`);

    // Click Edit
    await page.click('button:has-text("Edit")');

    // Update estimated hours
    const estimatedInput = page.locator('input[placeholder="0"]').first();
    await estimatedInput.fill('10');

    // Update actual hours
    const actualInput = page.locator('input[placeholder="0"]').nth(1);
    await actualInput.fill('8');

    // Save
    await page.click('button:has-text("Save")');

    // Wait for save
    await page.waitForTimeout(1000);

    // Verify updates
    await expect(page.locator('text=10h')).toBeVisible();
    await expect(page.locator('text=8h')).toBeVisible();
  });

  test('should handle task not found', async ({ page }) => {
    // Set auth token
    await page.goto('http://localhost:8084/');
    await page.evaluate((token) => {
      localStorage.setItem('auth_token', token);
    }, authToken);

    // Navigate to non-existent task
    await page.goto(`http://localhost:8084/app/projects/${projectId}/tasks/999999`);

    // Wait for error message
    await page.waitForSelector('text=Task not found');

    // Check error is displayed
    await expect(page.locator('text=Task not found')).toBeVisible();

    // Check back button is available
    await expect(page.locator('button:has-text("Back to Project")')).toBeVisible();
  });
});
