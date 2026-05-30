import { test, expect } from '@playwright/test';
import { loginAsNewUser } from './auth';

type ProjectItem = {
  id: number;
  key?: string;
  title?: string;
  name?: string;
};

test.describe('Compare page', () => {
  test('allows opening comparison for two projects', async ({ page, request }) => {
    const session = await loginAsNewUser(page, request);

    const projectsResponse = await request.get('http://localhost:8000/api/v1/projects', {
      headers: {
        Authorization: `Bearer ${session.token}`
      }
    });

    expect(projectsResponse.ok()).toBeTruthy();

    const body = (await projectsResponse.json()) as { projects?: ProjectItem[] };
    const projects = body.projects ?? [];

    expect(projects.length).toBeGreaterThanOrEqual(2);

    const selected = projects.slice(0, 2);

    const params = new URLSearchParams();
    selected.forEach((project) => {
      params.append('keys', project.title || project.name || project.key || '');
    });
    selected.forEach((project) => {
      const id = String(project.id);
      params.append('value', id);
      params.append('projectIds', id);
    });

    await page.goto(`/compare-projects?${params.toString()}`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('heading', { name: /Сравнение/i })).toBeVisible({ timeout: 30000 });
    await expect(page.locator('table.tbl')).toBeVisible({ timeout: 30000 });
    await expect(page.getByText('Сухая статистика')).toBeVisible({ timeout: 30000 });
  });
});

