import { expect, test } from '@playwright/test';

test('settings shows billing plan controls for an authenticated workspace', async ({ page, request }) => {
	const unique = Date.now().toString(36);
	const email = `billing-${unique}@example.com`;
	const password = 'password-1234';

	const register = await request.post('/api/v1/auth/register', {
		data: { email, password }
	});
	expect(register.ok()).toBeTruthy();
	const auth = await register.json();
	expect(auth.token).toBeTruthy();

	const workspace = await request.post('/api/v1/workspaces', {
		headers: { Authorization: `Bearer ${auth.token}` },
		data: { name: 'Billing E2E' }
	});
	expect(workspace.ok()).toBeTruthy();

	await page.addInitScript((token) => {
		window.localStorage.setItem('token', token);
	}, auth.token);
	await page.goto('/settings');

	await expect(page.getByRole('heading', { name: 'Billing' })).toBeVisible();
	await expect(page.getByText('No active plan')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Customer Portal' })).toBeVisible();
	await expect(page.getByRole('button', { name: 'Start Checkout' })).toHaveCount(3);
	await expect(page.getByRole('heading', { name: 'Starter' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Creator' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Pro' })).toBeVisible();
});
