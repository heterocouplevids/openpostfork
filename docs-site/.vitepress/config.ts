import { defineConfig } from 'vitepress';

// Default to root-path hosting so custom-domain deployments work without extra config.
// Repository-path deployments (for example, GitHub Pages at /openpost/) should set OPENPOST_DOCS_BASE explicitly.
const docsBase = process.env.OPENPOST_DOCS_BASE?.trim() || '/';

const docsSidebar = [
	{
		text: 'Start Here',
		collapsed: false,
		items: [
			{ text: 'What is OpenPost?', link: '/guide/what-is-openpost' },
			{ text: 'Quickstart', link: '/guide/quickstart' },
			{ text: 'Concepts', link: '/guide/concepts' },
		],
	},
	{
		text: 'User Docs: Web App',
		collapsed: false,
		items: [
			{ text: 'Overview', link: '/usage/' },
			{ text: 'Workspaces', link: '/usage/workspaces' },
			{ text: 'Settings', link: '/usage/settings' },
			{ text: 'Accounts', link: '/usage/accounts' },
			{ text: 'Composing Posts', link: '/usage/composing-posts' },
			{ text: 'Threads', link: '/usage/threads' },
			{ text: 'Scheduling', link: '/usage/scheduling' },
			{ text: 'Media Library', link: '/usage/media-library' },
		],
	},
	{
		text: 'User Docs: CLI',
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
		text: 'User Docs: MCP',
		collapsed: false,
		items: [{ text: 'Assistant Scheduling', link: '/mcp/' }],
	},
	{
		text: 'Self-Hosting: Overview',
		collapsed: false,
		items: [{ text: 'Start Here', link: '/self-hosting/' }],
	},
	{
		text: 'Self-Hosting: Install',
		collapsed: false,
		items: [
			{ text: 'Why Self-Host?', link: '/guide/why-selfhost' },
			{ text: 'Docker Compose', link: '/installation/docker-compose' },
			{ text: 'Single Binary', link: '/installation/binary' },
			{ text: 'Nix Module', link: '/installation/nix-module' },
			{ text: 'Reverse Proxy', link: '/installation/reverse-proxy' },
			{ text: 'Build From Source', link: '/installation/build-from-source' },
			{ text: 'Docker Run', link: '/installation/docker-run' },
		],
	},
	{
		text: 'Self-Hosting: Configure',
		collapsed: false,
		items: [
			{ text: 'Overview', link: '/configuration/overview' },
			{ text: 'Environment Variables', link: '/configuration/environment-variables' },
			{ text: 'Database', link: '/configuration/database' },
			{ text: 'Media Storage', link: '/configuration/media-storage' },
			{ text: 'CORS and URLs', link: '/configuration/cors-and-urls' },
			{ text: 'Production Checklist', link: '/configuration/production-checklist' },
		],
	},
	{
		text: 'Self-Hosting: Providers',
		collapsed: false,
		items: [
			{ text: 'Overview', link: '/providers/overview' },
			{ text: 'Supported Platforms & Limits', link: '/providers/platform-limits' },
			{ text: 'Provider Troubleshooting', link: '/providers/troubleshooting' },
			{ text: 'Provider Roadmap', link: '/providers/roadmap' },
			{ text: 'X', link: '/providers/x' },
			{ text: 'Mastodon', link: '/providers/mastodon' },
			{ text: 'Bluesky', link: '/providers/bluesky' },
			{ text: 'LinkedIn', link: '/providers/linkedin' },
			{ text: 'Threads', link: '/providers/threads' },
			{ text: 'Facebook', link: '/providers/facebook' },
			{ text: 'Instagram', link: '/providers/instagram' },
			{ text: 'TikTok', link: '/providers/tiktok' },
			{ text: 'YouTube', link: '/providers/youtube' },
		],
	},
	{
		text: 'Self-Hosting: Operate',
		collapsed: false,
		items: [
			{ text: 'Backups', link: '/operations/backups' },
			{ text: 'Health Checks', link: '/operations/health-checks' },
			{ text: 'Logs', link: '/operations/logs' },
			{ text: 'Upgrades', link: '/operations/upgrades' },
			{ text: 'Troubleshooting', link: '/operations/troubleshooting' },
		],
	},
	{
		text: 'Apps',
		collapsed: false,
		items: [{ text: 'Android App', link: '/installation/android' }],
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
		text: 'Developer Docs',
		collapsed: false,
		items: [
			{ text: 'Overview', link: '/development/' },
			{ text: 'Setup', link: '/development/setup' },
			{ text: 'Architecture', link: '/development/architecture' },
			{ text: 'API Reference', link: '/development/api-reference' },
			{ text: 'Frontend', link: '/development/frontend' },
			{ text: 'Backend', link: '/development/backend' },
			{ text: 'Platform Adapters', link: '/development/platform-adapters' },
			{ text: 'Background Jobs', link: '/development/background-jobs' },
			{ text: 'Testing', link: '/development/testing' },
			{ text: 'MCP And ChatGPT App', link: '/development/mcp' },
			{ text: 'Billing And Usage', link: '/development/billing-and-usage' },
			{ text: 'Production Readiness', link: '/development/production-readiness' },
			{ text: 'Contributing', link: '/development/contributing' },
		],
	},
];

export default defineConfig({
	title: 'OpenPost',
	description: 'Self-hosted Buffer/Hootsuite alternative. Schedule posts to X, Mastodon, Bluesky, Threads, LinkedIn, Facebook Pages, Instagram Business, TikTok, and YouTube from your own server.',
	base: docsBase,
	cleanUrls: true,
	lastUpdated: true,
	head: [
		['link', { rel: 'icon', href: `${docsBase}assets/brand/icon.svg` }],
		['meta', { property: 'og:type', content: 'website' }],
		['meta', { property: 'og:title', content: 'OpenPost' }],
		['meta', { property: 'og:description', content: 'Self-hosted Buffer/Hootsuite alternative. Schedule posts to X, Mastodon, Bluesky, Threads, LinkedIn, Facebook Pages, Instagram Business, TikTok, and YouTube from your own server.' }],
		['meta', { property: 'og:image', content: `${docsBase}assets/brand/og-image.svg` }],
	],
	themeConfig: {
		logo: '/assets/brand/icon.svg',
		nav: [
			{ text: 'Home', link: 'https://openpost.social' },
			{ text: 'User Docs', link: '/usage/' },
			{ text: 'CLI', link: '/cli/' },
			{ text: 'MCP', link: '/mcp/' },
			{ text: 'Self-Hosting', link: '/self-hosting/' },
			{ text: 'Providers', link: '/providers/overview' },
			{ text: 'Developer Docs', link: '/development/' },
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
