<script module lang="ts">
	const posts = [
		{
			href: '/blog/agentic-social-media-scheduling',
			label: 'Product strategy',
			title: 'Agentic social media scheduling needs a source of truth.',
			description:
				'Why OpenPost treats a publication as the canonical idea, then lets assistants and humans shape platform-native renditions from there.',
			meta: '8 min read'
		},
		{
			href: '/tips/best-times-to-post',
			label: 'Publishing tips',
			title: 'Best posting times are useful. Release choreography is better.',
			description:
				'Replace universal timing advice with platform context, local schedules, and staged releases.',
			meta: '6 min read'
		},
		{
			href: '/tips/cross-posting-without-looking-spammy',
			label: 'Workflow',
			title: 'Cross-post without making every account look automated.',
			description:
				'Turn one source idea into several network-native drafts instead of copy-pasting the same post everywhere.',
			meta: '5 min read'
		}
	];
</script>

<svelte:head>
	<title>OpenPost Blog - Agentic Social Media Scheduling</title>
	<meta
		name="description"
		content="OpenPost articles about agentic social media scheduling, source publications, platform-native renditions, and reliable release workflows."
	/>
	<meta property="og:title" content="OpenPost Blog" />
	<meta
		property="og:description"
		content="Practical writing on agentic scheduling, publication workflows, and social release operations."
	/>
</svelte:head>

<main class="blog shell">
	<nav aria-label="Breadcrumb">
		<a href="/">OpenPost</a>
		<span>/</span>
		<span>Blog</span>
	</nav>

	<section class="hero">
		<div>
			<p class="eyebrow">OpenPost blog</p>
			<h1>Publishing systems, not posting hacks.</h1>
		</div>
		<p>
			Notes on building a scheduler around source publications, platform-native renditions,
			assistant workflows, and reliable release operations.
		</p>
	</section>

	<section class="featured" aria-labelledby="featured-heading">
		<div>
			<p class="eyebrow">Featured</p>
			<h2 id="featured-heading">Agentic scheduling starts before the caption.</h2>
			<p>
				The useful unit is not a single post. It is the launch note, announcement, video idea,
				product update, or source brief that needs to become several destination-specific pieces.
			</p>
			<a href="/blog/agentic-social-media-scheduling">Read the article</a>
		</div>
		<img src="/assets/screenshots/main-dark.png" alt="OpenPost dashboard showing scheduled posts" />
	</section>

	<section class="post-list" aria-label="Latest articles">
		{#each posts as post (post.href)}
			<a href={post.href}>
				<span>{post.label}</span>
				<h2>{post.title}</h2>
				<p>{post.description}</p>
				<small>{post.meta}</small>
			</a>
		{/each}
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

	.blog {
		padding: 36px 0 80px;
	}

	nav {
		display: flex;
		gap: 10px;
		margin-bottom: clamp(56px, 10vw, 110px);
		color: #b9b1a6;
		font-weight: 800;
	}

	nav a,
	.featured a,
	.post-list a {
		text-decoration: none;
	}

	nav a:hover,
	.featured a:hover {
		color: #75d69c;
	}

	.eyebrow,
	.post-list span {
		margin: 0;
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

	.hero {
		display: grid;
		grid-template-columns: minmax(0, 0.95fr) minmax(280px, 0.62fr);
		gap: clamp(24px, 6vw, 72px);
		align-items: end;
	}

	h1 {
		max-width: 10ch;
		margin-bottom: 0;
		font-size: clamp(4rem, 9vw, 7.6rem);
		line-height: 0.84;
	}

	.hero > p,
	.featured p,
	.post-list p {
		color: #d2cabe;
		line-height: 1.65;
	}

	.hero > p {
		max-width: 54ch;
		font-size: 1.15rem;
	}

	.featured {
		display: grid;
		grid-template-columns: minmax(0, 0.72fr) minmax(360px, 1fr);
		gap: 28px;
		align-items: center;
		margin-top: clamp(56px, 10vw, 96px);
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 18px;
		background: #141210;
		padding: clamp(18px, 4vw, 30px);
	}

	.featured h2 {
		margin-bottom: 16px;
		font-size: clamp(2.2rem, 5vw, 4.8rem);
		line-height: 0.92;
	}

	.featured a {
		display: inline-flex;
		margin-top: 22px;
		border-radius: 999px;
		background: #f1eadb;
		padding: 10px 16px;
		color: #15120f;
		font-weight: 900;
	}

	.featured img {
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 14px;
		background: #0b0a09;
	}

	.post-list {
		display: grid;
		gap: 14px;
		margin-top: 28px;
	}

	.post-list a {
		display: grid;
		grid-template-columns: 160px minmax(220px, 0.72fr) 1fr auto;
		gap: 18px;
		align-items: start;
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 18px;
		background: #141210;
		padding: 18px;
		transition:
			border-color 160ms ease,
			transform 160ms ease;
	}

	.post-list a:hover {
		border-color: rgba(117, 214, 156, 0.45);
		transform: translateY(-1px);
	}

	.post-list h2 {
		margin-bottom: 0;
		font-size: clamp(1.6rem, 3vw, 2.6rem);
		line-height: 1;
	}

	.post-list p,
	.post-list small {
		margin: 0;
	}

	.post-list small {
		color: #b9b1a6;
		font-weight: 900;
		white-space: nowrap;
	}

	@media (max-width: 820px) {
		.hero,
		.featured,
		.post-list a {
			grid-template-columns: 1fr;
		}

		.post-list small {
			white-space: normal;
		}
	}
</style>
