import { expect, test } from '@playwright/test';

test('marketing page presents the cloud product and demo slot', async ({ page }) => {
	await page.goto('/');

	await expect(page).toHaveTitle(/OpenPost Cloud/);
	await expect(page.getByRole('heading', { name: 'Agentic social media scheduling' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Start free trial' }).first()).toBeVisible();
	await expect(page.getByLabel('OpenPost product demo placeholder')).toBeVisible();
	await expect(page.getByText('Replace this with the recorded walkthrough.')).toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'Managed social infrastructure without enterprise theatre.' })
	).toBeVisible();
	await expect(page.getByRole('link', { name: 'View GitHub' })).toBeVisible();
});

test('marketing page has no horizontal overflow', async ({ page }) => {
	await page.goto('/');

	const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
	expect(overflow).toBeLessThanOrEqual(1);
});
