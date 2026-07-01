import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ComposePreviewPanel, { type PreviewGroup } from './compose-preview-panel.svelte';

const previewGroups: PreviewGroup[] = [
	{
		key: 'account-1:instagram',
		accountId: 'account-1',
		platformKey: 'instagram',
		platformName: 'Instagram',
		username: 'openpost_main',
		displayName: 'OpenPost Main',
		avatarUrl: 'https://cdn.example/main.jpg',
		isUnsynced: false,
		posts: [
			{
				key: 'post-1',
				content: 'Main channel launch',
				mediaIds: [],
				mediaMimeTypes: {}
			}
		]
	},
	{
		key: 'account-2:instagram',
		accountId: 'account-2',
		platformKey: 'instagram',
		platformName: 'Instagram',
		username: 'openpost_studio',
		displayName: 'OpenPost Studio',
		avatarUrl: 'https://cdn.example/studio.jpg',
		isUnsynced: true,
		posts: [
			{
				key: 'post-1',
				content: 'Studio-specific caption',
				mediaIds: [],
				mediaMimeTypes: {}
			}
		]
	}
];

describe('ComposePreviewPanel', () => {
	it('renders a separate custom preview per selected account', async () => {
		const screen = await render(ComposePreviewPanel, { groups: previewGroups });

		expect(screen.container.querySelectorAll('[data-testid="instagram-preview"]')).toHaveLength(2);
		expect(screen.container.textContent).toContain('@openpost_main');
		expect(screen.container.textContent).toContain('@openpost_studio');
		await expect.element(screen.getByText('Studio-specific caption')).toBeVisible();
		await expect.element(screen.getByText('Customized for Instagram')).toBeVisible();
		expect(screen.container.querySelector('img[src="https://cdn.example/main.jpg"]')).toBeTruthy();
	});
});
