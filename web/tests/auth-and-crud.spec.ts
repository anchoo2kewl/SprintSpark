import { test, expect } from '@playwright/test';
import {
  generateTestUser,
  generateProject,
  generateTask,
  signupUser,
  loginUser,
  logoutUser,
  createProjectUI,
  verifyProjectInSidebar,
  navigateToProject,
  verifyProjectsAPI,
  getAuthToken,
  createTaskUI,
  verifyTaskInBoard,
  verifyTasksAPI,
  dragTaskToStatus,
  openTaskDetail,
  updateTaskViaModal,
  verifyTaskStatus,
  verifyTaskDueDate,
} from './helpers';

test.describe('SprintSpark - Complete User Journey', () => {
  // Generate unique test user for each test run
  const testUser = generateTestUser();
  const project1 = generateProject();
  const project2 = generateProject();

  test('should complete full authentication and CRUD flow', async ({ page }) => {
    // ====================
    // 1. SIGNUP FLOW
    // ====================
    await test.step('User can sign up with new account', async () => {
      await signupUser(page, testUser.email, testUser.password);

      // Verify we're on the dashboard
      await expect(page).toHaveURL('/app');

      // Verify user email is displayed in header
      await expect(page.locator(`text=${testUser.email}`)).toBeVisible();

      // Verify auth token is stored
      const token = await getAuthToken(page);
      expect(token).toBeTruthy();
      expect(token).toMatch(/^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$/); // JWT format
    });

    // ====================
    // 2. PROJECT CREATION
    // ====================
    await test.step('User can create first project', async () => {
      // Create project via UI
      await createProjectUI(page, project1.name, project1.description);

      // Verify project appears in sidebar
      await verifyProjectInSidebar(page, project1.name);

      // Verify we're navigated to the project detail page
      await expect(page).toHaveURL(/\/app\/projects\/\d+/);

      // Verify project name is displayed in header
      await expect(page.locator(`h1:has-text("${project1.name}")`)).toBeVisible();

      // Verify project description is displayed
      if (project1.description) {
        await expect(page.locator(`text=${project1.description}`)).toBeVisible();
      }

      // Take screenshot of project detail page
      await page.screenshot({ path: 'test-results/project-detail.png', fullPage: true });
    });

    await test.step('User can create second project', async () => {
      await createProjectUI(page, project2.name, project2.description);

      // Verify both projects appear in sidebar
      await verifyProjectInSidebar(page, project1.name);
      await verifyProjectInSidebar(page, project2.name);
    });

    // ====================
    // 3. PROJECT NAVIGATION
    // ====================
    await test.step('User can navigate between projects', async () => {
      // Navigate to first project
      await navigateToProject(page, project1.name);
      await expect(page.locator(`h1:has-text("${project1.name}")`)).toBeVisible();

      // Navigate to second project
      await navigateToProject(page, project2.name);
      await expect(page.locator(`h1:has-text("${project2.name}")`)).toBeVisible();

      // Navigate back to first project
      await navigateToProject(page, project1.name);
      await expect(page.locator(`h1:has-text("${project1.name}")`)).toBeVisible();
    });

    // ====================
    // 4. API STATE VERIFICATION
    // ====================
    await test.step('Projects are persisted in API', async () => {
      const projects = await verifyProjectsAPI(page);

      // Verify we have at least 2 projects
      expect(projects.length).toBeGreaterThanOrEqual(2);

      // Verify our projects are in the API response
      const project1Data = projects.find((p: any) => p.name === project1.name);
      const project2Data = projects.find((p: any) => p.name === project2.name);

      expect(project1Data).toBeTruthy();
      expect(project2Data).toBeTruthy();

      // Verify project structure
      expect(project1Data).toMatchObject({
        name: project1.name,
        description: project1.description,
      });
      expect(project1Data).toHaveProperty('id');
      expect(project1Data).toHaveProperty('owner_id');
      expect(project1Data).toHaveProperty('created_at');
      expect(project1Data).toHaveProperty('updated_at');
    });

    // ====================
    // 5. EMPTY STATES
    // ====================
    await test.step('Project shows empty state when no tasks', async () => {
      await navigateToProject(page, project1.name);

      // Verify "No tasks" empty state
      await expect(page.locator('text=No tasks')).toBeVisible();
      await expect(page.locator('text=Get started by creating a new task')).toBeVisible();
    });

    // ====================
    // 5.5 TASK CREATION
    // ====================
    const task1 = generateTask();
    const task2 = generateTask();
    const task3 = generateTask();
    let project1Id: number;

    await test.step('User can create tasks in project', async () => {
      await navigateToProject(page, project1.name);

      // Get project ID from URL
      const url = page.url();
      const match = url.match(/\/projects\/(\d+)/);
      project1Id = match ? parseInt(match[1]) : 0;
      expect(project1Id).toBeGreaterThan(0);

      // Create first task
      await createTaskUI(page, task1.title, task1.description);

      // Verify task appears in the board
      await verifyTaskInBoard(page, task1.title);

      // Verify "To Do" column has 1 task
      await expect(page.locator('text=To Do (1)')).toBeVisible();

      // Create second and third tasks
      await createTaskUI(page, task2.title, task2.description);
      await createTaskUI(page, task3.title);

      // Verify all tasks appear
      await verifyTaskInBoard(page, task1.title);
      await verifyTaskInBoard(page, task2.title);
      await verifyTaskInBoard(page, task3.title);

      // Verify count updated
      await expect(page.locator('text=To Do (3)')).toBeVisible();

      // Take screenshot of tasks board
      await page.screenshot({ path: 'test-results/tasks-board.png', fullPage: true });
    });

    await test.step('Tasks are persisted in API', async () => {
      const tasks = await verifyTasksAPI(page, project1Id);

      // Verify we have 3 tasks
      expect(tasks.length).toBe(3);

      // Verify task data
      const task1Data = tasks.find((t: any) => t.title === task1.title);
      const task2Data = tasks.find((t: any) => t.title === task2.title);
      const task3Data = tasks.find((t: any) => t.title === task3.title);

      expect(task1Data).toBeTruthy();
      expect(task2Data).toBeTruthy();
      expect(task3Data).toBeTruthy();

      // Verify task structure
      expect(task1Data).toMatchObject({
        title: task1.title,
        description: task1.description,
        status: 'todo',
      });

      expect(task1Data).toHaveProperty('id');
      expect(task1Data).toHaveProperty('project_id', project1Id);
      expect(task1Data).toHaveProperty('created_at');
    });

    await test.step('Empty state disappears after tasks created', async () => {
      await navigateToProject(page, project1.name);

      // Verify "No tasks" message is NOT visible
      await expect(page.locator('text=No tasks')).not.toBeVisible();

      // Verify task board is visible
      await expect(page.locator('text=To Do (3)')).toBeVisible();
    });

    // ====================
    // 5.7 DRAG AND DROP STATUS CHANGES
    // ====================
    await test.step('User can drag tasks to change status', async () => {
      await navigateToProject(page, project1.name);

      // Drag first task from "To Do" to "In Progress"
      await dragTaskToStatus(page, task1.title, 'in_progress');

      // Verify task moved to "In Progress" column
      await expect(page.locator('text=To Do (2)')).toBeVisible();
      await expect(page.locator('text=In Progress (1)')).toBeVisible();

      // Verify task has "In Progress" badge
      await verifyTaskStatus(page, task1.title, 'In Progress');

      // Drag second task to "Done"
      await dragTaskToStatus(page, task2.title, 'done');

      // Verify counts updated
      await expect(page.locator('text=To Do (1)')).toBeVisible();
      await expect(page.locator('text=Done (1)')).toBeVisible();

      // Verify task has "Done" badge
      await verifyTaskStatus(page, task2.title, 'Done');
    });

    await test.step('Drag and drop changes persist in API', async () => {
      const tasks = await verifyTasksAPI(page, project1Id);

      const task1Data = tasks.find((t: any) => t.title === task1.title);
      const task2Data = tasks.find((t: any) => t.title === task2.title);

      expect(task1Data?.status).toBe('in_progress');
      expect(task2Data?.status).toBe('done');
    });

    // ====================
    // 5.8 TASK EDITING VIA MODAL
    // ====================
    await test.step('User can edit task via detail modal', async () => {
      await navigateToProject(page, project1.name);

      // Click on task3 to open detail modal
      await openTaskDetail(page, task3.title);

      // Verify modal shows correct data
      await expect(page.locator('#edit-title')).toHaveValue(task3.title);
      await expect(page.locator('#edit-description')).toHaveValue(task3.description || '');

      // Update task with new data
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      const tomorrowStr = tomorrow.toISOString().split('T')[0];

      await updateTaskViaModal(page, {
        title: task3.title + ' (Updated)',
        description: 'Updated description',
        status: 'in_progress',
        dueDate: tomorrowStr
      });

      // Verify updated task appears with new title
      await verifyTaskInBoard(page, task3.title + ' (Updated)');

      // Verify status changed to "In Progress"
      await verifyTaskStatus(page, task3.title + ' (Updated)', 'In Progress');

      // Verify due date is displayed
      await verifyTaskDueDate(page, task3.title + ' (Updated)');

      // Verify counts updated (now 2 in progress, 0 in todo)
      await expect(page.locator('text=To Do (0)')).toBeVisible();
      await expect(page.locator('text=In Progress (2)')).toBeVisible();
    });

    await test.step('Task edits persist in API', async () => {
      const tasks = await verifyTasksAPI(page, project1Id);

      const task3Data = tasks.find((t: any) => t.title === task3.title + ' (Updated)');

      expect(task3Data).toBeTruthy();
      expect(task3Data?.description).toBe('Updated description');
      expect(task3Data?.status).toBe('in_progress');
      expect(task3Data?.due_date).toBeTruthy();

      // Verify due date is correct (should be tomorrow)
      const dueDate = new Date(task3Data?.due_date);
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      expect(dueDate.getDate()).toBe(tomorrow.getDate());
    });

    await test.step('Can change task status via dropdown in modal', async () => {
      await navigateToProject(page, project1.name);

      // Open task1 detail
      await openTaskDetail(page, task1.title);

      // Change status from "In Progress" to "Done" via dropdown
      await updateTaskViaModal(page, {
        status: 'done'
      });

      // Verify task moved to "Done" column
      await expect(page.locator('text=In Progress (1)')).toBeVisible();
      await expect(page.locator('text=Done (2)')).toBeVisible();
      await verifyTaskStatus(page, task1.title, 'Done');
    });

    // ====================
    // 6. LOGOUT FLOW
    // ====================
    await test.step('User can logout successfully', async () => {
      await logoutUser(page);

      // Verify we're redirected to login page
      await expect(page).toHaveURL('/login');

      // Verify token is cleared
      const token = await getAuthToken(page);
      expect(token).toBeNull();
    });

    // ====================
    // 7. LOGIN FLOW
    // ====================
    await test.step('User can login with existing credentials', async () => {
      await loginUser(page, testUser.email, testUser.password);

      // Verify we're on the dashboard
      await expect(page).toHaveURL('/app');

      // Verify user email is displayed
      await expect(page.locator(`text=${testUser.email}`)).toBeVisible();

      // Verify projects are still there after re-login
      await verifyProjectInSidebar(page, project1.name);
      await verifyProjectInSidebar(page, project2.name);
    });

    // ====================
    // 8. PERSISTENCE VERIFICATION
    // ====================
    await test.step('Data persists across sessions', async () => {
      // Verify projects are still accessible via API
      const projects = await verifyProjectsAPI(page);
      expect(projects.length).toBeGreaterThanOrEqual(2);

      const project1Data = projects.find((p: any) => p.name === project1.name);
      expect(project1Data).toBeTruthy();
    });

    // Final screenshot of complete state
    await page.screenshot({ path: 'test-results/final-state.png', fullPage: true });
  });

  test('should handle invalid login gracefully', async ({ page }) => {
    await test.step('Shows error for invalid credentials', async () => {
      await page.goto('/login');
      await page.fill('input[name="email"]', 'nonexistent@example.com');
      await page.fill('input[name="password"]', 'wrongpassword');
      await page.click('button[type="submit"]');

      // Verify error message is displayed
      await expect(page.locator('text=/invalid credentials|authentication failed/i')).toBeVisible({
        timeout: 5000,
      });

      // Verify we're still on login page
      await expect(page).toHaveURL('/login');
    });
  });

  test('should handle signup validation', async ({ page }) => {
    await test.step('Shows validation errors for invalid signup', async () => {
      await page.goto('/signup');

      // Try to submit with empty fields
      await page.click('button[type="submit"]');

      // Verify validation errors appear (email and password required)
      const emailError = page.locator('text=/email.*required/i');
      const passwordError = page.locator('text=/password.*required/i');

      // At least one validation error should be visible
      await expect(
        emailError.or(passwordError)
      ).toBeVisible({ timeout: 5000 });
    });

    await test.step('Shows error for weak password', async () => {
      await page.goto('/signup');
      await page.fill('input[name="email"]', 'test@example.com');
      await page.fill('input[name="password"]', 'weak');
      await page.fill('input[name="confirm-password"]', 'weak');

      // Trigger validation (blur or submit)
      await page.click('button[type="submit"]');

      // Verify password strength error
      await expect(
        page.locator('text=/password must.*8 characters|password.*too short/i')
      ).toBeVisible({ timeout: 5000 });
    });

    await test.step('Shows error for password mismatch', async () => {
      await page.goto('/signup');
      await page.fill('input[name="email"]', 'test@example.com');
      await page.fill('input[name="password"]', 'ValidPassword123!');
      await page.fill('input[name="confirm-password"]', 'DifferentPassword123!');
      await page.click('button[type="submit"]');

      // Verify password mismatch error
      await expect(
        page.locator('text=/passwords.*match|passwords must be the same/i')
      ).toBeVisible({ timeout: 5000 });
    });
  });
});
