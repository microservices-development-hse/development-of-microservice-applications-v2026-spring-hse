import { test, expect } from '@playwright/test';

test.describe('Projects page', () => {
  test('sends add request for a project', async ({ page }) => {
    await page.goto('/projects');

    await expect(page.getByRole('heading', { name: 'Проекты', exact: true })).toBeVisible();

    const addButton = page.getByRole('button', { name: /Добавить/i }).first();
    await expect(addButton).toBeVisible();

    const requestPromise = page.waitForRequest(request =>
      request.method() === 'POST' &&
      (
        request.url().includes('/api/v1/connector/import') ||
        request.url().includes('/connector/updateProject') ||
        request.url().includes('/import')
      )
    );

    await addButton.click();

    const request = await requestPromise;
    expect(request.method()).toBe('POST');

    const postData = request.postData() || '';
    expect(postData).toContain('project_key');
  });
});
