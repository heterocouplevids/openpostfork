import { defineConfig } from 'vitepress';

// Default to root-path hosting so custom-domain deployments work without extra config.
// Repository-path deployments (for example, GitHub Pages at /openpost/) should set OPENPOST_DOCS_BASE explicitly.
const docsBase = process.env.OPENPOST_DOCS_BASE?.trim() || '/';

const docsSidebar = [
	{
		text: 'Getting Started',
		collapsed: false,
		items: [
			{ text: 'What is OpenPost?', link: '/guide/what-is-openpost' },
			{ text: 'Why self-host?', link: '/guide/why-selfhost' },
			{ text: 'Quickstart', link: '/guide/quickstart' },
			{ text: 'Concepts', link: '/guide/concepts' },
		],
	},
	{
		text: 'Deployment',
		collapsed: false,
		items: [
			{ text: 'Docker Compose', link: '/installation/docker-compose' },
			{ text: 'Single Binary', link: '/installation/binary' },
			{ text: 'Nix Module', link: '/installation/nix-module' },
			{ text: 'Reverse Proxy', link: '/installation/reverse-proxy' },
			{ text: 'Android App', link: '/installation/android' },
		],
	},
	{
		text: 'Configuration',
		collapsed: false,
		items: [
			{ text: 'Environment Variables', link: '/configuration/environment-variables' },
			{ text: 'Database', link: '/configuration/database' },
			{ text: 'Media Storage', link: '/configuration/media-storage' },
			{ text: 'CORS and URLs', link: '/configuration/cors-and-urls' },
		],
	},
	{
		text: 'Providers',
		collapsed: false,
		items: [
			{ text: 'Overview', link: '/providers/overview' },
			{ text: 'Supported Platforms & Limits', link: '/providers/platform-limits' },
			{ text: 'X', link: '/providers/x' },
			{ text: 'Mastodon', link: '/providers/mastodon' },
			{ text: 'Bluesky', link: '/providers/bluesky' },
			{ text: 'LinkedIn', link: '/providers/linkedin' },
			{ text: 'Threads', link: '/providers/threads' },
		],
	},
	{
		text: 'Using OpenPost',
		collapsed: false,
		items: [
			{ text: 'Accounts', link: '/usage/accounts' },
			{ text: 'Composing Posts', link: '/usage/composing-posts' },
			{ text: 'Scheduling', link: '/usage/scheduling' },
			{ text: 'Media Library', link: '/usage/media-library' },
		],
	},
	{
		text: 'CLI',
		collapsed: false,
		items: [
			{ text: 'Overview', link: '/cli/' },
			{ text: 'Installation', link: '/cli/installation' },
			{ text: 'Authentication', link: '/cli/authentication' },
			{ text: 'Posting', link: '/cli/posting' },
			{ text: 'Automation', link: '/cli/automation' },
			{ text: 'Command Reference', link: '/reference/cli' },
		],
	},
	{
		text: 'Operations',
		collapsed: false,
		items: [
			{ text: 'Backups', link: '/operations/backups' },
			{ text: 'Upgrades', link: '/operations/upgrades' },
			{ text: 'Troubleshooting', link: '/operations/troubleshooting' },
		],
	},
	{
		text: 'Reference',
		collapsed: false,
		items: [
			{ text: 'API', link: '/reference/api' },
			{ text: 'CLI', link: '/reference/cli' },
			{ text: 'Environment Variables', link: '/reference/env-vars' },
			{ text: 'Callback URLs', link: '/reference/callback-urls' },
			{ text: 'Docker Compose', link: '/reference/docker-compose' },
		],
	},
];

const developmentSidebar = [
	{
		text: 'Development',
		collapsed: false,
		items: [
			{ text: 'Setup', link: '/development/setup' },
			{ text: 'Architecture', link: '/development/architecture' },
			{ text: 'API Reference', link: '/development/api-reference' },
			{ text: 'Frontend', link: '/development/frontend' },
			{ text: 'Backend', link: '/development/backend' },
			{ text: 'Platform Adapters', link: '/development/platform-adapters' },
			{ text: 'Background Jobs', link: '/development/background-jobs' },
			{ text: 'Testing', link: '/development/testing' },
			{ text: 'Contributing', link: '/development/contributing' },
		],
	},
];

export default defineConfig({
	title: 'OpenPost',
	description: 'Self-hosted Buffer/Hootsuite alternative. Schedule posts to X, Mastodon, Bluesky, Threads, and LinkedIn from your own server.',
	base: docsBase,
	cleanUrls: true,
	lastUpdated: true,
	head: [
		['link', { rel: 'icon', href: `${docsBase}assets/brand/icon.svg` }],
		['meta', { property: 'og:type', content: 'website' }],
		['meta', { property: 'og:title', content: 'OpenPost' }],
		['meta', { property: 'og:description', content: 'Self-hosted Buffer/Hootsuite alternative. Schedule posts to X, Mastodon, Bluesky, Threads, and LinkedIn from your own server.' }],
		['meta', { property: 'og:image', content: `${docsBase}assets/brand/og-image.svg` }],
	],
	themeConfig: {
		logo: '/assets/brand/icon.svg',
		nav: [
			{ text: 'Home', link: 'https://openpost.social' },
			{ text: 'Guide', link: '/guide/quickstart' },
			{ text: 'Installation', link: '/installation/docker-compose' },
			{ text: 'CLI', link: '/cli/' },
			{ text: 'Providers', link: '/providers/overview' },
			{ text: 'Operations', link: '/operations/backups' },
			{ text: 'Development', link: '/development/setup' },
		],
		socialLinks: [{ icon: 'github', link: 'https://github.com/rodrgds/openpost' }],
		search: {
			provider: 'local',
		},
		editLink: {
			pattern: 'https://github.com/rodrgds/openpost/edit/main/docs-site/:path',
			text: 'Edit this page on GitHub',
		},
		footer: {
			message: 'Released under the MIT License.',
			copyright: 'Copyright © Rodrigo Dias',
		},
		sidebar: {
			'/development/': developmentSidebar,
			'/': docsSidebar,
		},
	},
});
