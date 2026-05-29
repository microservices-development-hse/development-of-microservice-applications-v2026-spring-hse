import { test, expect } from '@playwright/test';
import { loginAsNewUser } from './auth';

test.describe('Projects page', () => {
  test('sends add request for a project', async ({ page, request }) => {
    await loginAsNewUser(page, request);

    const backendResponsePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/connector/projects') && resp.ok()
    );

    await page.goto('/projects');
    await backendResponsePromise;

    const rows = page.locator('table tbody tr');
    await expect(rows.first()).toBeVisible({ timeout: 30000 });

    const addButton = rows.first().locator('button');
    await expect(addButton).toBeVisible({ timeout: 30000 });

    const requestPromise = page.waitForRequest(req =>
      req.method() === 'POST' &&
      (
        req.url().includes('/api/v1/connector/import') ||
        req.url().includes('/connector/updateProject') ||
        req.url().includes('/import')
      )
    );

    await addButton.click();

    const req = await requestPromise;
    expect(req.method()).toBe('POST');
    expect(req.postData() || '').toContain('project_key');
  });
});

