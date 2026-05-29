import { test, expect } from '@playwright/test';
import { loginAsNewUser } from './auth';

test.describe('Compare page', () => {
  test('allows selecting two projects and opening comparison', async ({ page, request }) => {
    await loginAsNewUser(page, request);

    await page.goto('/compare');

    const checkboxes = page.locator('input[type="checkbox"]');
    await expect(checkboxes.first()).toBeVisible({ timeout: 30000 });

    const count = await checkboxes.count();
    expect(count).toBeGreaterThanOrEqual(2);

    await checkboxes.nth(0).check();
    await checkboxes.nth(1).check();

    await page.getByRole('button', { name: /Сравнить/i }).click();

    await expect(page).toHaveURL(/\/compare-projects/, { timeout: 30000 });
    await expect(page.getByText('Сухая статистика')).toBeVisible({ timeout: 30000 });
    await expect(page.locator('table.tbl')).toBeVisible({ timeout: 30000 });
  });
});

