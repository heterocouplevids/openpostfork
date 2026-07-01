import type { RequestHandler } from './$types';

export const prerender = true;

const siteUrl = 'https://openpost.social';
const routes = [
	{ path: '/', priority: '1.0' },
	{ path: '/tools', priority: '0.8' },
	{ path: '/blog', priority: '0.8' },
	{ path: '/blog/agentic-social-media-scheduling', priority: '0.8' },
	{ path: '/compare/social-media-schedulers', priority: '0.8' },
	{ path: '/tips/best-times-to-post', priority: '0.7' },
	{ path: '/tips/cross-posting-without-looking-spammy', priority: '0.7' }
];

function escapeXml(value: string) {
	return value
		.replaceAll('&', '&amp;')
		.replaceAll('<', '&lt;')
		.replaceAll('>', '&gt;')
		.replaceAll('"', '&quot;')
		.replaceAll("'", '&apos;');
}

export const GET: RequestHandler = () => {
	const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${routes
	.map(
		(route) => `  <url>
    <loc>${escapeXml(`${siteUrl}${route.path}`)}</loc>
    <priority>${route.priority}</priority>
  </url>`
	)
	.join('\n')}
</urlset>
`;

	return new Response(body, {
		headers: {
			'content-type': 'application/xml; charset=utf-8'
		}
	});
};
