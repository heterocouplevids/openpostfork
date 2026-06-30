<script lang="ts">
	const appUrl = 'https://app.openpost.social';
	const docsUrl = 'https://docs.openpost.social';
	const samplePost =
		'OpenPost turns one launch note into platform-native drafts, then lets you review, schedule, and monitor the release from one dashboard.';
	const platformLimits = [
		{ name: 'X', limit: 280, note: 'Short hooks and threads' },
		{ name: 'Bluesky', limit: 300, note: 'Open-web conversations' },
		{ name: 'Threads', limit: 500, note: 'Conversational posts' },
		{ name: 'Mastodon', limit: 500, note: 'Instance-friendly context' },
		{ name: 'LinkedIn', limit: 3000, note: 'Longer professional copy' }
	];
	const previewPlatforms = ['X', 'Bluesky', 'LinkedIn', 'Mastodon'];

	let postText = $state(samplePost);
	let threadLimit = $state(280);

	const characterCount = $derived(Array.from(postText).length);
	const wordCount = $derived(postText.trim().split(/\s+/).filter(Boolean).length);
	const threadParts = $derived(splitThread(postText, threadLimit));

	function remaining(limit: number) {
		return limit - characterCount;
	}

	function splitThread(input: string, limit: number) {
		const normalized = input.replace(/\s+/g, ' ').trim();

		if (!normalized) {
			return [];
		}

		const words = normalized.split(' ');
		const chunks: string[] = [];
		let current = '';

		for (const word of words) {
			const candidate = current ? `${current} ${word}` : word;

			if (Array.from(candidate).length <= limit) {
				current = candidate;
				continue;
			}

			if (current) {
				chunks.push(current);
			}

			current = Array.from(word).length > limit ? Array.from(word).slice(0, limit).join('') : word;
		}

		if (current) {
			chunks.push(current);
		}

		return chunks;
	}
</script>

<svelte:head>
	<title>Free Social Post Tools - OpenPost Cloud</title>
	<meta
		name="description"
		content="Free social media tools from OpenPost: count characters, split threads, and preview platform-native posts before scheduling."
	/>
	<meta property="og:title" content="Free Social Post Tools" />
	<meta
		property="og:description"
		content="Draft once, check platform limits, split threads, and preview how a post will read across networks."
	/>
</svelte:head>

<header class="nav shell">
	<a class="brand" href="/" aria-label="OpenPost home">
		<img src="/icon.svg" alt="" />
		<span>OpenPost</span>
	</a>
	<nav aria-label="Tools navigation">
		<a href="/#pricing">Pricing</a>
		<a href="/tips/best-times-to-post">Tips</a>
		<a href={docsUrl}>Docs</a>
	</nav>
	<a class="login" href={appUrl}>Open app</a>
</header>

<main>
	<section class="hero shell">
		<p class="eyebrow">Free publishing tools</p>
		<h1>Check the post before it hits the queue.</h1>
		<p>
			Draft once, see platform limits, split a thread, and preview the same source idea across
			networks. OpenPost Cloud turns these checks into scheduled publication workflows.
		</p>
	</section>

	<section class="workbench shell" aria-label="Social post tool workbench">
		<div class="editor-panel">
			<label for="post-text">Source post</label>
			<textarea id="post-text" bind:value={postText} data-testid="post-tool-input"></textarea>
			<div class="editor-meta" aria-label="Post statistics">
				<span>{characterCount} chars</span>
				<span>{wordCount} words</span>
			</div>
		</div>

		<div class="limits-panel">
			<div class="panel-heading">
				<p class="eyebrow">Character counter</p>
				<h2>Platform fit</h2>
			</div>
			<div class="limit-list">
				{#each platformLimits as platform (platform.name)}
					<article class:over={remaining(platform.limit) < 0}>
						<div>
							<strong>{platform.name}</strong>
							<span>{platform.note}</span>
						</div>
						<p data-testid={`remaining-${platform.name}`}>
							{remaining(platform.limit) >= 0
								? `${remaining(platform.limit)} left`
								: `${Math.abs(remaining(platform.limit))} over`}
						</p>
					</article>
				{/each}
			</div>
		</div>
	</section>

	<section class="thread-tool shell" aria-labelledby="thread-heading">
		<div>
			<p class="eyebrow">Thread splitter</p>
			<h2 id="thread-heading">Turn a long note into clean parts.</h2>
			<p>
				Use this for launch notes, changelog summaries, release recaps, and rough posts that need to
				become a thread.
			</p>
			<label class="range-label" for="thread-limit">Characters per part: {threadLimit}</label>
			<input
				id="thread-limit"
				type="range"
				min="180"
				max="500"
				step="10"
				bind:value={threadLimit}
			/>
		</div>
		<div class="thread-parts" data-testid="thread-parts">
			{#if threadParts.length}
				{#each threadParts as part, index (part)}
					<article>
						<span>{index + 1}/{threadParts.length}</span>
						<p>{part}</p>
					</article>
				{/each}
			{:else}
				<p class="empty">Paste a draft to split it into thread parts.</p>
			{/if}
		</div>
	</section>

	<section class="previews shell" aria-labelledby="preview-heading">
		<div class="section-heading">
			<div>
				<p class="eyebrow">Social post preview</p>
				<h2 id="preview-heading">See how one source reads in different rooms.</h2>
			</div>
			<p>
				The real product keeps destination-specific renditions attached to a source publication so
				every network can get the version it deserves.
			</p>
		</div>
		<div class="preview-grid">
			{#each previewPlatforms as platform (platform)}
				<article>
					<div class="preview-top">
						<strong>{platform}</strong>
						<span>{characterCount} chars</span>
					</div>
					<p>{postText || 'Your post preview will appear here.'}</p>
				</article>
			{/each}
		</div>
	</section>

	<section class="cta">
		<div class="shell">
			<h2>Ready to move from checking copy to scheduling releases?</h2>
			<a class="button" href={`${appUrl}/register?plan=creator`}>Start OpenPost Cloud</a>
		</div>
	</section>
</main>

<style>
	:global(*) {
		box-sizing: border-box;
	}

	:global(html) {
		background: #0b0a09;
		color: #f5f1e8;
	}

	:global(body) {
		margin: 0;
		font-family:
			Avenir Next,
			Aptos,
			ui-sans-serif,
			sans-serif;
		background: #0b0a09;
		color: #f5f1e8;
	}

	.shell {
		width: min(1120px, calc(100% - 32px));
		margin: 0 auto;
	}

	a {
		color: inherit;
	}

	img {
		display: block;
		max-width: 100%;
	}

	.nav {
		position: sticky;
		top: 0;
		z-index: 10;
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 24px;
		padding: 18px 0;
		background: rgba(11, 10, 9, 0.86);
		backdrop-filter: blur(18px);
	}

	.brand,
	.nav nav,
	.login {
		display: inline-flex;
		align-items: center;
	}

	.brand {
		gap: 10px;
		text-decoration: none;
		font-weight: 900;
	}

	.brand img {
		width: 32px;
		height: 32px;
	}

	.nav nav {
		justify-content: center;
		gap: 22px;
		color: #b9b1a6;
		font-size: 0.92rem;
	}

	.nav a {
		text-decoration: none;
	}

	.nav nav a:hover {
		color: #75d69c;
	}

	.login,
	.button {
		justify-content: center;
		border: 1px solid rgba(245, 241, 232, 0.18);
		border-radius: 999px;
		padding: 8px 14px;
		color: #f5f1e8;
		font-weight: 800;
		text-decoration: none;
	}

	.hero {
		padding: clamp(72px, 12vw, 132px) 0 48px;
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
		max-width: 11ch;
		margin-bottom: 24px;
		font-size: clamp(4rem, 9vw, 7.8rem);
		line-height: 0.85;
	}

	h2 {
		margin-bottom: 18px;
		font-size: clamp(2.2rem, 5vw, 4.8rem);
		line-height: 0.92;
	}

	.hero p:not(.eyebrow),
	.section-heading > p,
	.thread-tool > div:first-child p {
		max-width: 68ch;
		color: #d2cabe;
		font-size: 1.08rem;
		line-height: 1.65;
	}

	.workbench,
	.thread-tool,
	.previews {
		display: grid;
		gap: 24px;
		padding: 48px 0;
	}

	.workbench {
		grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.8fr);
		align-items: start;
	}

	.editor-panel,
	.limits-panel,
	.thread-parts,
	.preview-grid article {
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 18px;
		background: #141210;
	}

	.editor-panel {
		padding: 18px;
	}

	label,
	.range-label {
		display: block;
		margin-bottom: 10px;
		color: #efe6d6;
		font-weight: 900;
	}

	textarea {
		width: 100%;
		min-height: 330px;
		resize: vertical;
		border: 1px solid rgba(245, 241, 232, 0.16);
		border-radius: 12px;
		background: #0b0a09;
		color: #f5f1e8;
		font: inherit;
		font-size: 1.05rem;
		line-height: 1.6;
		padding: 16px;
	}

	textarea:focus-visible,
	input:focus-visible,
	.button:focus-visible,
	.login:focus-visible,
	.nav a:focus-visible {
		outline: 2px solid #75d69c;
		outline-offset: 3px;
	}

	.editor-meta,
	.preview-top {
		display: flex;
		justify-content: space-between;
		gap: 16px;
		color: #a69d91;
		font-variant-numeric: tabular-nums;
		font-weight: 800;
	}

	.editor-meta {
		margin-top: 14px;
	}

	.limits-panel {
		padding: 20px;
	}

	.panel-heading h2 {
		font-size: 2.4rem;
	}

	.limit-list {
		display: grid;
		gap: 10px;
	}

	.limit-list article {
		display: flex;
		justify-content: space-between;
		gap: 16px;
		border: 1px solid rgba(245, 241, 232, 0.12);
		border-radius: 12px;
		padding: 14px;
	}

	.limit-list strong,
	.preview-top strong {
		display: block;
		color: #f5f1e8;
	}

	.limit-list span {
		display: block;
		margin-top: 4px;
		color: #a69d91;
		font-size: 0.9rem;
	}

	.limit-list p {
		margin: 0;
		color: #75d69c;
		white-space: nowrap;
		font-weight: 900;
	}

	.limit-list article.over p {
		color: #ff9f6f;
	}

	.thread-tool {
		grid-template-columns: minmax(260px, 0.72fr) minmax(0, 1.28fr);
		align-items: start;
	}

	input[type='range'] {
		width: 100%;
		accent-color: #75d69c;
	}

	.thread-parts {
		display: grid;
		gap: 12px;
		padding: 16px;
	}

	.thread-parts article {
		display: grid;
		grid-template-columns: 56px 1fr;
		gap: 14px;
		border: 1px solid rgba(245, 241, 232, 0.1);
		border-radius: 12px;
		padding: 14px;
	}

	.thread-parts span {
		color: #75d69c;
		font-weight: 900;
	}

	.thread-parts p,
	.preview-grid p,
	.empty {
		margin: 0;
		color: #d2cabe;
		line-height: 1.55;
	}

	.section-heading {
		display: grid;
		grid-template-columns: minmax(0, 0.95fr) minmax(260px, 0.65fr);
		gap: 24px;
		align-items: end;
	}

	.preview-grid {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 14px;
	}

	.preview-grid article {
		min-height: 230px;
		padding: 16px;
	}

	.preview-top {
		margin-bottom: 18px;
	}

	.cta {
		margin-top: 48px;
		background: #f1eadb;
		color: #15120f;
	}

	.cta .shell {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		padding: 56px 0;
	}

	.cta h2 {
		max-width: 780px;
		margin: 0;
	}

	.cta .button {
		background: #15120f;
		color: #f5f1e8;
		white-space: nowrap;
	}

	@media (max-width: 860px) {
		.nav {
			position: static;
			grid-template-columns: 1fr;
		}

		.nav nav {
			justify-content: flex-start;
			flex-wrap: wrap;
		}

		.workbench,
		.thread-tool,
		.section-heading,
		.preview-grid,
		.cta .shell {
			grid-template-columns: 1fr;
		}

		.preview-grid {
			display: grid;
		}

		.cta .shell {
			display: grid;
		}
	}
</style>
