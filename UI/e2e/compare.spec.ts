import { test, expect } from '@playwright/test';

test.describe('Compare page', () => {
  test('allows selecting two projects and opening comparison', async ({ page }) => {
    await page.goto('/compare');

    await expect(page.getByRole('heading', { name: 'Сравнение', exact: true })).toBeVisible();

    const checkboxes = page.locator('input[type="checkbox"]');
    const count = await checkboxes.count();

    expect(count).toBeGreaterThanOrEqual(2);

    await checkboxes.nth(0).check();
    await checkboxes.nth(1).check();

    await page.getByRole('button', { name: /Сравнить/i }).click();

    await expect(page.getByText(/Минимальное число проектов для сравнения/i)).toHaveCount(0);
  });
});
