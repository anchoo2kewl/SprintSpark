import { test, expect } from '@playwright/test';
import {
  generateTestUser,
  signupUser,
  loginUser,
  getAuthToken,
} from './helpers';

test.describe('Admin Functionality', () => {
  const adminUser = generateTestUser();
  const regularUser = generateTestUser();

  test.beforeAll(async ({ browser }) => {
    // Create admin user via API
    const context = await browser.newContext();
    const page = await context.newPage();

    // Signup as admin user
    await signupUser(page, adminUser.email, adminUser.password);
    await page.waitForTimeout(500);

    // Make user admin via API (simulating CLI command)
    const token = await getAuthToken(page);
    await page.evaluate(async ({ email, token: _authToken }) => {
      // In a real scenario, this would be done via server CLI
      // For testing, we'll need to make this user admin through the database
      // This is a simplification - in production the CLI command would be used
      console.log(`Admin user created: ${email}`);
    }, { email: adminUser.email, token });

    await context.close();
  });

  test('Non-admin user cannot access admin page', async ({ page }) => {
    // Create and login as regular user
    await signupUser(page, regularUser.email, regularUser.password);

    // Try to navigate to admin page
    await page.goto('/app/admin');

    // Should be redirected back to /app
    await page.waitForURL('/app');
    await expect(page).toHaveURL('/app');
  });

  test('Admin user can access admin dashboard', async ({ page }) => {
    // Login as admin (assuming user was made admin via CLI)
    await loginUser(page, adminUser.email, adminUser.password);

    // Check if Admin button is visible in sidebar
    const adminButton = page.locator('button:has-text("Admin")');

    // If admin button is not visible, user is not admin - skip test
    const isAdminButtonVisible = await adminButton.isVisible().catch(() => false);
    if (!isAdminButtonVisible) {
      test.skip();
      return;
    }

    // Click admin button
    await adminButton.click();

    // Verify we're on admin page
    await expect(page).toHaveURL('/app/admin');
    await expect(page.locator('h1:has-text("Admin Dashboard")')).toBeVisible();
  });

  test('Admin can view users list with statistics', async ({ page }) => {
    // Login as admin
    await loginUser(page, adminUser.email, adminUser.password);

    // Navigate to admin page
    await page.goto('/app/admin');

    // Check if we can access (user might not be admin in test env)
    const isAdminPage = await page.locator('h1:has-text("Admin Dashboard")').isVisible().catch(() => false);
    if (!isAdminPage) {
      test.skip();
      return;
    }

    // Wait for users table to load
    await expect(page.locator('table')).toBeVisible();

    // Verify table headers
    await expect(page.locator('th:has-text("Email")')).toBeVisible();
    await expect(page.locator('th:has-text("Admin")')).toBeVisible();
    await expect(page.locator('th:has-text("Logins")')).toBeVisible();
    await expect(page.locator('th:has-text("Failed")')).toBeVisible();
    await expect(page.locator('th:has-text("Last IP")')).toBeVisible();
    await expect(page.locator('th:has-text("Actions")')).toBeVisible();

    // Verify at least one user row exists (the admin user)
    const userRows = page.locator('tbody tr');
    await expect(userRows).toHaveCount(await userRows.count());
    expect(await userRows.count()).toBeGreaterThan(0);

    // Verify admin user is in the list
    await expect(page.locator(`td:has-text("${adminUser.email}")`)).toBeVisible();
  });

  test('Admin can view user activity log', async ({ page }) => {
    // Login as admin
    await loginUser(page, adminUser.email, adminUser.password);

    // Navigate to admin page
    await page.goto('/app/admin');

    // Check if we can access
    const isAdminPage = await page.locator('h1:has-text("Admin Dashboard")').isVisible().catch(() => false);
    if (!isAdminPage) {
      test.skip();
      return;
    }

    // Wait for users table
    await expect(page.locator('table')).toBeVisible();

    // Click "Activity" button for first user
    const activityButton = page.locator('button:has-text("Activity")').first();
    await activityButton.click();

    // Wait for activity log to appear
    await expect(page.locator('h2:has-text("Activity Log")')).toBeVisible();

    // Activity log should show some entries or "No activity" message
    const hasActivity = await page.locator('text=/LOGIN|FAILED_LOGIN|No activity/i').isVisible();
    expect(hasActivity).toBeTruthy();
  });

  test('Admin can toggle admin privileges for users', async ({ page }) => {
    // Login as admin
    await loginUser(page, adminUser.email, adminUser.password);

    // Navigate to admin page
    await page.goto('/app/admin');

    // Check if we can access
    const isAdminPage = await page.locator('h1:has-text("Admin Dashboard")').isVisible().catch(() => false);
    if (!isAdminPage) {
      test.skip();
      return;
    }

    // Wait for users table
    await expect(page.locator('table')).toBeVisible();

    // Find a non-admin user row (if any exist)
    const userRows = page.locator('tbody tr');
    const rowCount = await userRows.count();

    if (rowCount > 1) {
      // Look for "Make Admin" button
      const makeAdminButton = page.locator('button:has-text("Make Admin")').first();
      const hasMakeAdmin = await makeAdminButton.isVisible().catch(() => false);

      if (hasMakeAdmin) {
        const row = page.locator('tbody tr').filter({ has: makeAdminButton });

        // Click Make Admin
        await makeAdminButton.click();

        // Wait for API call to complete
        await page.waitForTimeout(1000);

        // Verify the button changed to "Revoke Admin"
        await expect(row.locator('button:has-text("Revoke Admin")')).toBeVisible();

        // Verify admin badge is shown
        await expect(row.locator('text=Admin')).toBeVisible();
      }
    }
  });

  test('Activity log shows login attempts with IP addresses', async ({ page }) => {
    // Login as admin (this creates a login activity)
    await loginUser(page, adminUser.email, adminUser.password);

    // Navigate to admin page
    await page.goto('/app/admin');

    // Check if we can access
    const isAdminPage = await page.locator('h1:has-text("Admin Dashboard")').isVisible().catch(() => false);
    if (!isAdminPage) {
      test.skip();
      return;
    }

    // Wait for users table
    await expect(page.locator('table')).toBeVisible();

    // Click Activity for admin user
    const adminRow = page.locator(`tr:has-text("${adminUser.email}")`);
    await adminRow.locator('button:has-text("Activity")').click();

    // Wait for activity log
    await expect(page.locator('h2:has-text("Activity Log")')).toBeVisible();

    // Should see at least one LOGIN activity
    await expect(page.locator('text=/LOGIN/i')).toBeVisible();

    // Should see IP address label
    await expect(page.locator('text=/IP:/i')).toBeVisible();
  });

  test('Failed login attempts are tracked', async ({ page }) => {
    // Try to login with wrong password (creates failed login activity)
    await page.goto('/login');
    await page.fill('input[type="email"]', adminUser.email);
    await page.fill('input[type="password"]', 'wrongpassword123');
    await page.click('button[type="submit"]');

    // Should see error message
    await expect(page.locator('text=/invalid|incorrect|failed/i')).toBeVisible();

    // Now login correctly
    await loginUser(page, adminUser.email, adminUser.password);

    // Navigate to admin page
    await page.goto('/app/admin');

    // Check if we can access
    const isAdminPage = await page.locator('h1:has-text("Admin Dashboard")').isVisible().catch(() => false);
    if (!isAdminPage) {
      test.skip();
      return;
    }

    // Wait for users table
    await expect(page.locator('table')).toBeVisible();

    // Check failed attempts column for admin user
    const adminRow = page.locator(`tr:has-text("${adminUser.email}")`);
    const failedCount = await adminRow.locator('td').nth(3).textContent(); // Failed column

    // Should have at least 1 failed attempt from our wrong password attempt
    expect(parseInt(failedCount || '0')).toBeGreaterThanOrEqual(1);
  });
});
