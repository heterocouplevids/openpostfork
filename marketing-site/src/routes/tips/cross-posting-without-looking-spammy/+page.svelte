<script module lang="ts">
	const appUrl = 'https://app.openpost.social';
	const rules = [
		[
			'Start with a source publication',
			'Keep the canonical idea separate from the platform drafts.'
		],
		['Adapt the first line', 'Each network has a different hook style and tolerance for context.'],
		[
			'Change the format',
			'A LinkedIn post may become an X thread, a Mastodon note, and a Bluesky conversation starter.'
		],
		[
			'Reuse media carefully',
			'The same image can work everywhere, but alt text, cropping, and order may not.'
		],
		[
			'Stagger the release',
			'Give each post room to breathe and avoid making every account look automated.'
		]
	];
</script>

<svelte:head>
	<title>Cross-Post Without Looking Spammy - OpenPost</title>
	<meta
		name="description"
		content="A practical guide to cross-posting across social platforms without publishing identical copy everywhere."
	/>
	<meta property="og:title" content="Cross-Post Without Looking Spammy" />
	<meta
		property="og:description"
		content="Turn one source idea into platform-native renditions, not a copy-paste blast."
	/>
</svelte:head>

<main class="article shell">
	<nav aria-label="Breadcrumb">
		<a href="/">OpenPost</a>
		<span>/</span>
		<a href="/tools">Tools</a>
	</nav>
	<p class="eyebrow">Publishing tips</p>
	<h1>Cross-posting works when the copy does not look copied.</h1>
	<p class="lede">
		The mistake is not publishing one idea to several networks. The mistake is pretending every
		network is the same room.
	</p>

	<section class="rules">
		{#each rules as [title, text], index (title)}
			<article>
				<span>{index + 1}</span>
				<div>
					<h2>{title}</h2>
					<p>{text}</p>
				</div>
			</article>
		{/each}
	</section>

	<section class="closing">
		<h2>OpenPost is built around renditions, not copy-paste.</h2>
		<p>
			Create a publication once, adapt destination-specific drafts, review previews, then schedule
			the release intentionally from the dashboard, CLI, or MCP workflows.
		</p>
		<a href={`${appUrl}/register?plan=creator`}>Try OpenPost Cloud</a>
	</section>
</main>

<style>
	:global(*) {
		box-sizing: border-box;
	}

	:global(body) {
		margin: 0;
		background: #0b0a09;
		color: #f5f1e8;
		font-family:
			Avenir Next,
			Aptos,
			ui-sans-serif,
			sans-serif;
	}

	.shell {
		width: min(980px, calc(100% - 32px));
		margin: 0 auto;
	}

	a {
		color: inherit;
	}

	.article {
		padding: 36px 0 72px;
	}

	nav {
		display: flex;
		gap: 10px;
		margin-bottom: clamp(56px, 10vw, 112px);
		color: #b9b1a6;
		font-weight: 800;
	}

	nav a,
	.closing a {
		text-decoration: none;
	}

	nav a:hover {
		color: #75d69c;
	}

	.eyebrow {
		margin: 0 0 12px;
		color: #75d69c;
		font-size: 0.78rem;
		font-weight: 900;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	h1,
	h2,
	p {
		margin-top: 0;
	}

	h1,
	h2 {
		font-family:
			Charter,
			Iowan Old Style,
			Georgia,
			serif;
		letter-spacing: 0;
	}

	h1 {
		max-width: 12ch;
		margin-bottom: 24px;
		font-size: clamp(4rem, 9vw, 7.5rem);
		line-height: 0.86;
	}

	.lede {
		max-width: 66ch;
		margin-bottom: 72px;
		color: #d2cabe;
		font-size: 1.18rem;
		line-height: 1.65;
	}

	.rules {
		display: grid;
		gap: 14px;
	}

	.rules article {
		display: grid;
		grid-template-columns: 64px 1fr;
		gap: 20px;
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 18px;
		background: #141210;
		padding: 20px;
	}

	.rules span {
		color: #75d69c;
		font-size: 1.5rem;
		font-weight: 900;
	}

	.rules h2,
	.closing h2 {
		margin-bottom: 10px;
		font-size: clamp(1.8rem, 3vw, 3rem);
		line-height: 0.98;
	}

	.rules p,
	.closing p {
		margin: 0;
		color: #d2cabe;
		line-height: 1.62;
	}

	.closing {
		margin-top: 72px;
		border-radius: 22px;
		background: #f1eadb;
		padding: clamp(24px, 5vw, 46px);
		color: #15120f;
	}

	.closing p {
		max-width: 68ch;
		color: #332b24;
	}

	.closing a {
		display: inline-flex;
		margin-top: 22px;
		border-radius: 999px;
		background: #15120f;
		padding: 10px 16px;
		color: #f5f1e8;
		font-weight: 900;
	}

	@media (max-width: 640px) {
		.rules article {
			grid-template-columns: 1fr;
		}
	}
</style>
