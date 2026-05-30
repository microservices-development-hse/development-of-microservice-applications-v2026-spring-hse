import { APIRequestContext, expect, Page } from '@playwright/test';
import { randomUUID } from 'crypto';

const AUTH_URL = 'http://localhost:8083';
const PASSWORD = 'admin';

export type AuthSession = {
  token: string;
  email: string;
  expiresAt: number;
};

export async function loginAsNewUser(page: Page, request: APIRequestContext): Promise<AuthSession> {
  const email = `e2e-${randomUUID()}@local`;

  const registerResponse = await request.post(`${AUTH_URL}/register`, {
    data: { email, password: PASSWORD }
  });

  expect([201, 409]).toContain(registerResponse.status());

  const loginResponse = await request.post(`${AUTH_URL}/login`, {
    data: { email, password: PASSWORD }
  });

  expect(loginResponse.ok()).toBeTruthy();

  const session = await loginResponse.json() as AuthSession;

  const normalizedSession: AuthSession = {
    ...session,
    expiresAt:
      typeof session.expiresAt === 'number' && session.expiresAt < 1_000_000_000_000
        ? session.expiresAt * 1000
        : session.expiresAt
  };

  await page.addInitScript((value) => {
    localStorage.setItem('jira-analyzer-session', JSON.stringify(value));
  }, normalizedSession);

  return normalizedSession;
}

