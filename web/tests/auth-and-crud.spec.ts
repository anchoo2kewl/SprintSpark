import { test, expect } from '@playwright/test';
import {
  generateTestUser,
  generateProject,
  signupUser,
  loginUser,
  logoutUser,
  createProjectUI,
  verifyProjectInSidebar,
  navigateToProject,
  verifyProjectsAPI,
  getAuthToken,
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
