import { expect, test } from '@playwright/test';

test('marketing page presents the cloud product and demo slot', async ({ page }) => {
	await page.goto('/');

	await expect(page).toHaveTitle(/OpenPost Cloud/);
	await expect(page.getByRole('heading', { name: 'Agentic social media scheduling' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Start free trial' }).first()).toBeVisible();
	await expect(page.getByRole('link', { name: 'Start free trial' }).first()).toHaveAttribute(
		'href',
		'https://app.openpost.social/register?plan=creator'
	);
	await expect(page.getByRole('link', { name: 'Start Starter' })).toHaveAttribute(
		'href',
		'https://app.openpost.social/register?plan=starter'
	);
	await expect(page.getByRole('link', { name: 'Start Creator' })).toHaveAttribute(
		'href',
		'https://app.openpost.social/register?plan=creator'
	);
	await expect(page.getByRole('link', { name: 'Start Pro' })).toHaveAttribute(
		'href',
		'https://app.openpost.social/register?plan=pro'
	);
	await expect(page.getByLabel('OpenPost product demo placeholder')).toBeVisible();
	await expect(page.getByText('Replace this with the recorded walkthrough.')).toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'Managed social infrastructure without enterprise theatre.' })
	).toBeVisible();
	await expect(page.getByRole('link', { name: 'View GitHub' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Tools' }).first()).toHaveAttribute('href', '/tools');
	await expect(page.getByRole('link', { name: 'Tips' }).first()).toHaveAttribute(
		'href',
		'/tips/best-times-to-post'
	);
});

test('marketing page has no horizontal overflow', async ({ page }) => {
	await page.goto('/');

	const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
	expect(overflow).toBeLessThanOrEqual(1);
});

test('marketing tools page counts, splits, and previews source copy', async ({ page }) => {
	await page.goto('/tools');

	await expect(page).toHaveTitle(/Free Social Post Tools/);
	await expect(
		page.getByRole('heading', { name: 'Check the post before it hits the queue.' })
	).toBeVisible();

	const input = page.getByTestId('post-tool-input');
	await input.fill('OpenPost ships a focused publication workflow for social releases.');

	await expect(page.getByTestId('remaining-X')).toContainText('214 left');
	await expect(page.getByTestId('thread-parts')).toContainText('1/1');
	await expect(page.getByText('OpenPost ships a focused publication workflow')).toHaveCount(5);
});

test('marketing tips pages are reachable SEO articles', async ({ page }) => {
	await page.goto('/tips/best-times-to-post');
	await expect(page).toHaveTitle(/Best Times to Post/);
	await expect(page.getByRole('heading', { name: /Best times to post/ })).toBeVisible();

	await page.goto('/tips/cross-posting-without-looking-spammy');
	await expect(page).toHaveTitle(/Cross-Post Without Looking Spammy/);
	await expect(page.getByRole('heading', { name: /Cross-posting works/ })).toBeVisible();
});

test('marketing SEO routes expose crawlable resources', async ({ request }) => {
	const robots = await request.get('/robots.txt');
	expect(robots.ok()).toBeTruthy();
	const robotsText = await robots.text();
	expect(robotsText).toContain('Sitemap: https://openpost.social/sitemap.xml');

	const sitemap = await request.get('/sitemap.xml');
	expect(sitemap.ok()).toBeTruthy();
	const xml = await sitemap.text();
	expect(xml).toContain('<loc>https://openpost.social/</loc>');
	expect(xml).toContain('<loc>https://openpost.social/tools</loc>');
	expect(xml).toContain('<loc>https://openpost.social/tips/best-times-to-post</loc>');
});
