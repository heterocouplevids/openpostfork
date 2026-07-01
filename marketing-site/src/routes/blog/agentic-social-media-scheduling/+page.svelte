<script module lang="ts">
	const appUrl = 'https://app.openpost.social';
	const principles = [
		[
			'Keep the source intact',
			'Store the original idea, goal, audience, links, and media before making platform drafts.'
		],
		[
			'Adapt by destination',
			'Let LinkedIn, X, Bluesky, Mastodon, Instagram, YouTube, and TikTok get different copy when the format demands it.'
		],
		[
			'Log assistant actions',
			'Agents should create drafts, suggest timings, and schedule with scoped tokens, audit trails, and human review.'
		],
		[
			'Schedule as a release',
			'One announcement may need a canonical launch, later follow-ups, a video post, and a thread.'
		]
	];
</script>

<svelte:head>
	<title>Agentic Social Media Scheduling - OpenPost</title>
	<meta
		name="description"
		content="A practical model for agentic social media scheduling built around source publications, platform-native renditions, scoped assistant access, and human review."
	/>
	<meta property="og:title" content="Agentic Social Media Scheduling" />
	<meta
		property="og:description"
		content="Why OpenPost treats publications as the source of truth before agents draft platform-native posts."
	/>
</svelte:head>

<main class="article shell">
	<nav aria-label="Breadcrumb">
		<a href="/">OpenPost</a>
		<span>/</span>
		<a href="/blog">Blog</a>
	</nav>
	<p class="eyebrow">Product strategy</p>
	<h1>Agentic social media scheduling needs a source of truth.</h1>
	<p class="lede">
		An assistant can write a caption quickly. A production scheduler needs something sturdier: a
		source publication that keeps the intent, media, audience, and release plan together.
	</p>

	<section class="intro">
		<p>
			Most schedulers start with a blank post. That works for manual posting, but it becomes thin
			when an agent is asked to turn a changelog, launch note, video idea, or product update into
			several network-specific drafts.
		</p>
		<p>
			OpenPost is moving the workflow one level up. The publication is the canonical brief. Posts
			and renditions are outputs. The dashboard, CLI, and MCP tools all work from that same object.
		</p>
	</section>

	<section>
		<h2>The operating model</h2>
		<div class="principles">
			{#each principles as [title, text], index (title)}
				<article>
					<span>{index + 1}</span>
					<div>
						<h3>{title}</h3>
						<p>{text}</p>
					</div>
				</article>
			{/each}
		</div>
	</section>

	<section class="split">
		<div>
			<h2>Assistants should not be a second product surface.</h2>
			<p>
				When ChatGPT, a desktop MCP client, or a CLI command creates a draft, the result should
				appear in the same queue, with the same workspace permissions, same media rules, and same
				schedule limits as the web app.
			</p>
		</div>
		<img
			src="/assets/screenshots/settings-dark.png"
			alt="OpenPost settings with tokens and controls"
		/>
	</section>

	<section class="callout">
		<h2>Try the workflow from the managed cloud app.</h2>
		<p>
			Start with a source publication, adapt the drafts, review the account-specific previews, and
			schedule the release from one place.
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

	img {
		display: block;
		max-width: 100%;
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
	.callout a {
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

	.lede,
	.intro p,
	.principles p,
	.split p,
	.callout p {
		color: #d2cabe;
		line-height: 1.65;
	}

	.lede {
		max-width: 68ch;
		margin-bottom: 72px;
		font-size: 1.18rem;
	}

	.intro {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 28px;
		border-top: 1px solid rgba(245, 241, 232, 0.14);
		padding-top: 28px;
	}

	section {
		margin-top: 72px;
	}

	.principles {
		display: grid;
		gap: 14px;
	}

	.principles article {
		display: grid;
		grid-template-columns: 56px 1fr;
		gap: 18px;
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 18px;
		background: #141210;
		padding: 18px;
	}

	.principles span {
		color: #75d69c;
		font-size: 1.4rem;
		font-weight: 900;
	}

	.principles h3 {
		margin-bottom: 8px;
		color: #f5f1e8;
		font-size: 1.1rem;
	}

	.principles p,
	.split p,
	.callout p {
		margin-bottom: 0;
	}

	.split {
		display: grid;
		grid-template-columns: minmax(0, 0.72fr) minmax(320px, 1fr);
		gap: 28px;
		align-items: center;
	}

	.split img {
		border: 1px solid rgba(245, 241, 232, 0.14);
		border-radius: 16px;
		background: #141210;
	}

	.callout {
		border-radius: 22px;
		background: #f1eadb;
		padding: clamp(24px, 5vw, 46px);
		color: #15120f;
	}

	.callout p {
		max-width: 68ch;
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
	}

	@media (max-width: 760px) {
		.intro,
		.split {
			grid-template-columns: 1fr;
		}
	}
</style>
