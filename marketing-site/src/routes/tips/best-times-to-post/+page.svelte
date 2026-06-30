<script module lang="ts">
	const appUrl = 'https://app.openpost.social';
	const tips = [
		[
			'Start from intent',
			'Match timing to the job of the post: launch, reply, teach, recap, or ask.'
		],
		[
			'Respect the room',
			'LinkedIn can take a morning narrative. X and Bluesky need sharper hooks. Mastodon rewards context.'
		],
		[
			'Stagger deliberately',
			'Publish the canonical announcement first, then adapt follow-ups instead of blasting every network at once.'
		],
		[
			'Review failures',
			'A good schedule includes time to catch provider errors, media processing delays, and last-minute edits.'
		]
	];
	const windows = [
		[
			'LinkedIn',
			'Tuesday to Thursday, local business morning',
			'Professional posts need review time and a clear opening line.'
		],
		[
			'X',
			'Weekday morning or early afternoon',
			'Use sharper hooks and split longer context into threads.'
		],
		['Bluesky', 'Late morning through evening', 'Conversation quality matters more than volume.'],
		['Mastodon', 'Instance-local daytime', 'Add context and avoid drive-by link drops.'],
		[
			'Threads',
			'Lunch breaks and early evening',
			'Conversational framing usually beats formal announcements.'
		]
	];
</script>

<svelte:head>
	<title>Best Times to Post on Social Media - OpenPost</title>
	<meta
		name="description"
		content="A practical guide to choosing social posting times across LinkedIn, X, Bluesky, Mastodon, and Threads without blasting every platform at once."
	/>
	<meta property="og:title" content="Best Times to Post on Social Media" />
	<meta
		property="og:description"
		content="Use timing windows, release choreography, and platform-native drafts instead of generic one-size-fits-all scheduling advice."
	/>
</svelte:head>

<main class="article shell">
	<nav aria-label="Breadcrumb">
		<a href="/">OpenPost</a>
		<span>/</span>
		<a href="/tools">Tools</a>
	</nav>
	<p class="eyebrow">Publishing tips</p>
	<h1>Best times to post are useful. Release choreography is better.</h1>
	<p class="lede">
		There is no universal perfect posting time. A useful schedule starts with audience context,
		platform behavior, and the kind of post you are shipping.
	</p>

	<section>
		<h2>Practical timing windows</h2>
		<div class="window-list">
			{#each windows as [platform, timing, note] (platform)}
				<article>
					<strong>{platform}</strong>
					<span>{timing}</span>
					<p>{note}</p>
				</article>
			{/each}
		</div>
	</section>

	<section>
		<h2>How OpenPost thinks about timing</h2>
		<div class="tip-grid">
			{#each tips as [title, text] (title)}
				<article>
					<h3>{title}</h3>
					<p>{text}</p>
				</article>
			{/each}
		</div>
	</section>

	<section class="callout">
		<h2>Turn timing into a release plan.</h2>
		<p>
			OpenPost Cloud is being built around publications, renditions, social sets, next-slot
			scheduling, and assistant workflows so each network gets the right post at the right time.
		</p>
		<a href={`${appUrl}/register?plan=creator`}>Start OpenPost Cloud</a>
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

	nav a {
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
	h3,
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
		margin-bottom: 24px;
		font-size: clamp(2.2rem, 5vw, 4.4rem);
		line-height: 0.94;
	}

	.lede {
		max-width: 68ch;
		margin-bottom: 72px;
		color: #d2cabe;
		font-size: 1.18rem;
		line-height: 1.65;
	}

	section {
		margin-top: 72px;
	}

	.window-list,
	.tip-grid {
		display: grid;
		gap: 14px;
	}

	.window-list article,
	.tip-grid article,
	.callout {
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 18px;
		background: #141210;
		padding: 20px;
	}

	.window-list article {
		display: grid;
		grid-template-columns: 130px minmax(180px, 0.72fr) 1fr;
		gap: 16px;
		align-items: start;
	}

	.window-list strong,
	.window-list span,
	.tip-grid h3 {
		color: #f5f1e8;
	}

	.window-list span {
		font-weight: 900;
	}

	.window-list p,
	.tip-grid p,
	.callout p {
		margin: 0;
		color: #d2cabe;
		line-height: 1.6;
	}

	.tip-grid {
		grid-template-columns: repeat(4, minmax(0, 1fr));
	}

	.tip-grid h3 {
		margin-bottom: 10px;
	}

	.callout {
		background: #f1eadb;
		color: #15120f;
	}

	.callout p {
		color: #332b24;
	}

	.callout a {
		display: inline-flex;
		margin-top: 22px;
		border-radius: 999px;
		background: #15120f;
		padding: 10px 16px;
		color: #f5f1e8;
		font-weight: 900;
		text-decoration: none;
	}

	@media (max-width: 780px) {
		.window-list article,
		.tip-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
