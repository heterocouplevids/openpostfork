<script module lang="ts">
	const appUrl = 'https://app.openpost.social';
	const rows = [
		[
			'Core unit',
			'OpenPost starts from a source publication with goal, audience, media, and release context.',
			'Traditional schedulers usually start from individual posts or calendar slots.'
		],
		[
			'Assistant workflows',
			'Remote MCP and CLI workflows can inspect workspaces, create drafts, attach media, and schedule through scoped tokens.',
			'Automation often lives outside the scheduler or requires custom glue.'
		],
		[
			'Platform adaptation',
			'Destination-specific renditions keep each account version tied to the same source idea.',
			'Cross-posting can push teams toward identical copy unless the process is manually enforced.'
		],
		[
			'Self-hosting',
			'The core remains self-hostable while OpenPost Cloud provides managed provider apps, storage, backups, and workers.',
			'Most hosted schedulers are managed-only products.'
		]
	];
	const fit = [
		'You want a cloud scheduler, but still care that the core product is inspectable.',
		'You publish from launch notes, changelogs, docs, videos, and product updates.',
		'You want ChatGPT-style assistants and CLI automation to work through the same permission model as the app.',
		'You would rather verify platform-specific drafts than blast identical copy everywhere.'
	];
</script>

<svelte:head>
	<title>OpenPost vs Traditional Social Media Schedulers</title>
	<meta
		name="description"
		content="Compare OpenPost Cloud with traditional social media schedulers for source publications, agentic workflows, platform-native renditions, and self-hostable trust."
	/>
	<meta property="og:title" content="OpenPost vs Traditional Social Media Schedulers" />
	<meta
		property="og:description"
		content="A practical comparison for teams that want agentic social scheduling, CLI/MCP workflows, and publication-first planning."
	/>
</svelte:head>

<main class="compare shell">
	<nav aria-label="Breadcrumb">
		<a href="/">OpenPost</a>
		<span>/</span>
		<span>Compare</span>
	</nav>

	<section class="hero">
		<p class="eyebrow">Comparison</p>
		<h1>OpenPost is for teams that schedule releases, not just posts.</h1>
		<p>
			Traditional schedulers are useful calendars. OpenPost Cloud is being built around source
			publications, platform-native renditions, assistant-safe actions, and a self-hostable core.
		</p>
		<a href={`${appUrl}/register?plan=creator`}>Start OpenPost Cloud</a>
	</section>

	<section class="table" aria-label="OpenPost comparison table">
		<div class="table-head">
			<span>Decision point</span>
			<span>OpenPost Cloud</span>
			<span>Traditional scheduler</span>
		</div>
		{#each rows as [label, openpost, traditional] (label)}
			<article>
				<strong>{label}</strong>
				<p>{openpost}</p>
				<p>{traditional}</p>
			</article>
		{/each}
	</section>

	<section class="fit">
		<div>
			<p class="eyebrow">Good fit</p>
			<h2>Choose OpenPost when the workflow starts before the calendar.</h2>
		</div>
		<ul>
			{#each fit as item (item)}
				<li>{item}</li>
			{/each}
		</ul>
	</section>

	<section class="cta">
		<h2>Build a repeatable publishing system.</h2>
		<p>
			Create a source publication, adapt account-specific drafts, review the queue, and schedule
			with the dashboard, CLI, or MCP clients.
		</p>
		<a href="/blog/agentic-social-media-scheduling">Read the agentic scheduling model</a>
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
		width: min(1080px, calc(100% - 32px));
		margin: 0 auto;
	}

	a {
		color: inherit;
	}

	.compare {
		padding: 36px 0 80px;
	}

	nav {
		display: flex;
		gap: 10px;
		margin-bottom: clamp(56px, 10vw, 112px);
		color: #b9b1a6;
		font-weight: 800;
	}

	nav a,
	.hero a,
	.cta a {
		text-decoration: none;
	}

	nav a:hover,
	.cta a:hover {
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

	h2 {
		margin-bottom: 0;
		font-size: clamp(2.2rem, 5vw, 4.6rem);
		line-height: 0.92;
	}

	.hero {
		display: grid;
		gap: 22px;
		max-width: 820px;
	}

	.hero p,
	.table p,
	.cta p {
		color: #d2cabe;
		line-height: 1.65;
	}

	.hero p {
		max-width: 68ch;
		font-size: 1.16rem;
	}

	.hero a {
		display: inline-flex;
		width: fit-content;
		border-radius: 999px;
		background: #f1eadb;
		padding: 10px 16px;
		color: #15120f;
		font-weight: 900;
	}

	.table {
		display: grid;
		gap: 10px;
		margin-top: clamp(56px, 10vw, 96px);
	}

	.table-head,
	.table article {
		display: grid;
		grid-template-columns: 180px 1fr 1fr;
		gap: 18px;
	}

	.table-head {
		color: #75d69c;
		font-size: 0.78rem;
		font-weight: 900;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	.table article {
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 18px;
		background: #141210;
		padding: 18px;
	}

	.table strong {
		color: #f5f1e8;
	}

	.table p {
		margin: 0;
	}

	.fit {
		display: grid;
		grid-template-columns: minmax(0, 0.8fr) minmax(320px, 1fr);
		gap: 28px;
		align-items: start;
		margin-top: 72px;
	}

	.fit ul {
		display: grid;
		gap: 12px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.fit li {
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 14px;
		background: #141210;
		padding: 16px;
		color: #d2cabe;
		line-height: 1.55;
	}

	.cta {
		margin-top: 72px;
		border-radius: 22px;
		background: #f1eadb;
		padding: clamp(24px, 5vw, 46px);
		color: #15120f;
	}

	.cta p {
		max-width: 68ch;
		color: #332b24;
	}

	.cta a {
		display: inline-flex;
		margin-top: 20px;
		border-radius: 999px;
		background: #15120f;
		padding: 10px 16px;
		color: #f5f1e8;
		font-weight: 900;
	}

	@media (max-width: 820px) {
		.table-head {
			display: none;
		}

		.table article,
		.fit {
			grid-template-columns: 1fr;
		}
	}
</style>
