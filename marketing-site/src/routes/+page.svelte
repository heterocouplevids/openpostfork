<script module lang="ts">
  const appUrl = "https://app.openpost.social";
  const docsUrl = "https://docs.openpost.social";
  const githubUrl = "https://github.com/rodrgds/openpost";
  const signupUrl = (plan: string) =>
    `${appUrl}/register?plan=${encodeURIComponent(plan)}`;

  const platforms = [
    "X",
    "Instagram",
    "LinkedIn",
    "Facebook",
    "TikTok",
    "YouTube",
    "Bluesky",
    "Mastodon",
    "Threads",
  ];
  const agents = [
    {
      name: "ChatGPT and MCP",
      text: "Create drafts, adapt posts, and schedule through a hosted MCP endpoint built for OpenPost Cloud.",
      tone: "blue",
    },
    {
      name: "CLI workflows",
      text: "Ship launch notes, threads, and scheduled posts from terminals, scripts, and release automation.",
      tone: "green",
    },
    {
      name: "Human review",
      text: "Keep final approval, account selection, previews, and publish status inside one dashboard.",
      tone: "orange",
    },
  ];
  const features = [
    [
      "Platform-native versions",
      "Turn one source idea into tailored renditions for each account instead of cross-posting identical copy.",
    ],
    [
      "AI-assisted drafting",
      "Use agent workflows to turn docs, links, launch notes, and media into editable scheduled posts.",
    ],
    [
      "Release choreography",
      "Publish everything together, stagger networks, send LinkedIn first, or follow up with threads later.",
    ],
    [
      "Media-aware publishing",
      "Manage reusable media, provider limits, and public upload requirements from one workflow.",
    ],
    [
      "Operational confidence",
      "See failures, retries, queue state, and provider-specific errors before they become a support mystery.",
    ],
    [
      "Open core trust",
      "The cloud service funds a self-hostable product whose core stays inspectable and portable.",
    ],
  ];
  const proof = [
    ["2 min", "from launch note to scheduled draft"],
    ["9+", "platforms on the roadmap"],
    ["24/7", "hosted workers and media storage"],
    ["0", "servers to maintain on Cloud"],
  ];
  const steps = [
    [
      "Brief",
      "Paste a changelog, link, launch note, video idea, or rough thought.",
    ],
    [
      "Adapt",
      "OpenPost shapes account-specific versions with constraints and previews.",
    ],
    [
      "Approve",
      "Review the dashboard, adjust tone, attach media, and choose a release plan.",
    ],
    [
      "Publish",
      "Workers handle schedule timing, retries, provider APIs, and status reporting.",
    ],
  ];
  const faqs = [
    [
      "Is OpenPost Cloud only for AI posting?",
      "No. The dashboard remains the primary control surface. Agentic workflows help create, adapt, schedule, and inspect posts faster.",
    ],
    [
      "Will there still be a self-hosted version?",
      "Yes. Self-hosting stays a core trust signal, but this site is focused on the managed cloud product.",
    ],
    [
      "Where does the demo go?",
      "The hero has a dedicated demo stage that can be replaced with the recorded product walkthrough.",
    ],
    [
      "Why not just use a generic scheduler?",
      "OpenPost is being designed around drafts, renditions, CLI automation, MCP tools, and provider-aware publishing.",
    ],
    [
      "Can agents publish directly?",
      "The default direction is draft and schedule first. Consequential actions like publish-now should stay scoped, logged, and confirmable.",
    ],
  ];
  const plans = [
    {
      id: "starter",
      name: "Starter",
      price: "$6",
      description: "Open-web posting for small projects and founders.",
      limits: [
        "1 workspace",
        "3 social accounts",
        "100 scheduled posts/mo",
        "1 GB media",
      ],
    },
    {
      id: "creator",
      name: "Creator",
      price: "$12",
      description:
        "Agentic scheduling across mainstream platforms with more accounts and media.",
      limits: [
        "3 workspaces",
        "6 social accounts",
        "500 scheduled posts/mo",
        "5 GB media",
      ],
      featured: true,
    },
    {
      id: "pro",
      name: "Pro",
      price: "$24",
      description:
        "Higher limits, more workspaces, team workflows, and deeper operations visibility.",
      limits: [
        "10 workspaces",
        "15 social accounts",
        "2,500 scheduled posts/mo",
        "25 GB media",
      ],
    },
  ];
</script>

<svelte:head>
  <title>OpenPost Cloud - Agentic social media scheduling</title>
  <meta
    name="description"
    content="OpenPost Cloud is an agentic social media scheduler for creating, adapting, reviewing, scheduling, and monitoring platform-native posts."
  />
  <meta property="og:title" content="OpenPost Cloud" />
  <meta
    property="og:description"
    content="Create once, adapt per platform, and schedule with a dashboard, CLI, and MCP workflows."
  />
</svelte:head>

<header id="top" class="nav shell">
  <a class="brand" href="#top" aria-label="OpenPost home">
    <img src="/icon.svg" alt="" />
    <span>OpenPost</span>
  </a>
  <nav aria-label="Main navigation">
    <a href="#agentic">Agentic</a>
    <a href="#features">Features</a>
    <a href="#demo">Demo</a>
    <a href="#pricing">Pricing</a>
    <a href="/tools">Tools</a>
    <a href="/blog">Blog</a>
    <a href="/compare/social-media-schedulers">Compare</a>
    <a href="/tips/best-times-to-post">Tips</a>
  </nav>
  <a class="login" href={appUrl}>Login</a>
</header>

<main>
  <section class="hero shell">
    <div class="hero-copy">
      <div
        class="platform-cloud"
        aria-label="Supported and planned social platforms"
      >
        {#each platforms as platform (platform)}
          <span>{platform}</span>
        {/each}
      </div>
      <p class="eyebrow">OpenPost Cloud</p>
      <h1>Agentic social media scheduling</h1>
      <p class="lede">
        Start with one base draft, let agents prepare platform-native versions,
        approve the final copy, and schedule the release from one calm
        dashboard.
      </p>
      <div class="hero-actions">
        <a class="button primary" href={signupUrl("creator")}
          >Start free trial</a
        >
        <a class="button secondary" href="#demo">Watch demo</a>
      </div>
      <p class="microcopy">
        No credit card during beta. Built for creators, founders, agencies, and
        open-source teams.
      </p>
    </div>

    <div
      id="demo"
      class="demo-stage"
      aria-label="OpenPost product demo placeholder"
    >
      <div class="demo-toolbar">
        <span></span><span></span><span></span>
        <strong>Demo recording slot</strong>
      </div>
      <div class="demo-screen">
        <img
          src="/assets/screenshots/main-dark.png"
          alt="OpenPost dashboard preview"
        />
        <div class="play-button" aria-hidden="true">▶</div>
      </div>
      <div class="demo-caption">
        <strong>Replace this with the recorded walkthrough.</strong>
        <span
          >Show idea input, AI renditions, review, scheduling, CLI, and MCP
          handoff.</span
        >
      </div>
    </div>
  </section>

  <section class="proof shell" aria-label="OpenPost Cloud snapshot">
    {#each proof as [value, label] (value)}
      <div>
        <strong>{value}</strong>
        <span>{label}</span>
      </div>
    {/each}
  </section>

  <section id="agentic" class="agentic shell">
    <div class="section-kicker">Agentic workflows</div>
    <h2>Built for dashboards, CLIs, and assistants working together.</h2>
    <div class="agent-grid">
      {#each agents as agent (agent.name)}
        <article class={`agent-card ${agent.tone}`}>
          <p>{agent.name}</p>
          <h3>{agent.text}</h3>
        </article>
      {/each}
    </div>
  </section>

  <section class="workflow shell">
    <div class="section-heading">
      <div>
        <div class="section-kicker">How it works</div>
        <h2>From rough idea to scheduled release.</h2>
      </div>
      <p>
        OpenPost should feel less like a blank text box and more like a
        publishing operator: it understands destinations, formats, timing,
        media, and review.
      </p>
    </div>
    <div class="steps">
      {#each steps as [title, text], index (title)}
        <article>
          <span>{index + 1}</span>
          <h3>{title}</h3>
          <p>{text}</p>
        </article>
      {/each}
    </div>
  </section>

  <section id="features" class="features shell">
    <div class="section-heading">
      <div>
        <div class="section-kicker">Product surface</div>
        <h2>Everything a serious scheduler needs before analytics bloat.</h2>
      </div>
      <p>
        The cloud product focuses on reliable creation, adaptation, scheduling,
        media handling, and provider visibility first.
      </p>
    </div>
    <div class="feature-grid">
      {#each features as [title, text] (title)}
        <article>
          <h3>{title}</h3>
          <p>{text}</p>
        </article>
      {/each}
    </div>
  </section>

  <section class="product-shot shell">
    <div>
      <div class="section-kicker">Review before anything ships</div>
      <h2>Agents draft. Humans approve. Workers publish.</h2>
      <p>
        The dashboard stays the source of truth for connected accounts,
        scheduled posts, media, failed jobs, API tokens, and assistant sessions.
      </p>
      <a class="text-link" href={appUrl}>Open the hosted app</a>
    </div>
    <div class="shot-stack">
      <img
        src="/assets/screenshots/settings-dark.png"
        alt="OpenPost settings and automation controls"
      />
      <img
        src="/assets/screenshots/accounts-dark.png"
        alt="OpenPost social accounts dashboard"
      />
    </div>
  </section>

  <section class="platforms shell">
    <div class="section-kicker">Channels</div>
    <h2>One release plan for the places your audience already lives.</h2>
    <div class="platform-grid">
      {#each platforms as platform (platform)}
        <div>{platform}</div>
      {/each}
    </div>
    <p>
      Provider availability depends on API approval, media requirements, and
      platform-specific limits. OpenPost Cloud will prioritize reliable,
      verified publishing over logo-count theater.
    </p>
  </section>

  <section id="pricing" class="pricing shell">
    <div class="section-heading">
      <div>
        <div class="section-kicker">Pricing direction</div>
        <h2>Managed social infrastructure without enterprise theatre.</h2>
      </div>
      <p>
        Hosted pricing should cover provider approvals, worker reliability,
        media storage, support, and the edge cases that come with social APIs.
      </p>
    </div>
    <div class="plans">
      {#each plans as plan (plan.id)}
        <article class:featured={plan.featured}>
          <span>{plan.name}</span>
          <strong>{plan.price}</strong>
          <p>{plan.description}</p>
          <ul>
            {#each plan.limits as limit (limit)}
              <li>{limit}</li>
            {/each}
          </ul>
          <a class="plan-cta" href={signupUrl(plan.id)}>Start {plan.name}</a>
        </article>
      {/each}
    </div>
  </section>

  <section class="open-core shell">
    <div>
      <div class="section-kicker">Open core</div>
      <h2>Transparent enough to trust. Hosted when you want convenience.</h2>
    </div>
    <p>
      OpenPost remains self-hostable and inspectable. The official cloud service
      exists for teams that prefer managed provider apps, backups, updates,
      media hosting, and support. <a href={githubUrl}>View GitHub</a>
      or <a href={docsUrl}>read the docs</a>.
    </p>
  </section>

  <section class="faq shell">
    <div class="section-kicker">FAQ</div>
    <h2>Questions before you hand it your posting workflow.</h2>
    <div class="faq-list">
      {#each faqs as [question, answer] (question)}
        <details>
          <summary>{question}</summary>
          <p>{answer}</p>
        </details>
      {/each}
    </div>
  </section>

  <section class="final-cta">
    <div class="shell">
      <h2>Ready to turn one idea into a planned release?</h2>
      <a class="button primary" href={signupUrl("creator")}>Start free trial</a>
    </div>
  </section>
</main>

<footer class="shell footer">
  <div>
    <img src="/logo.svg" alt="OpenPost" />
    <p>Agentic social media scheduling for the cloud, CLI, and open web.</p>
  </div>
  <nav aria-label="Footer navigation">
    <a href={appUrl}>App</a>
    <a href={docsUrl}>Docs</a>
    <a href="/tools">Tools</a>
    <a href="/blog">Blog</a>
    <a href="/compare/social-media-schedulers">Compare</a>
    <a href="/tips/best-times-to-post">Tips</a>
    <a href={githubUrl}>GitHub</a>
  </nav>
</footer>

<style>
  :global(*) {
    box-sizing: border-box;
  }

  :global(html) {
    scroll-behavior: smooth;
    background: #090909;
  }

  :global(body) {
    margin: 0;
    font-family:
      Avenir Next,
      Aptos,
      ui-sans-serif,
      sans-serif;
    background:
      radial-gradient(
        circle at 18% 8%,
        rgba(101, 188, 255, 0.14),
        transparent 28rem
      ),
      radial-gradient(
        circle at 82% 18%,
        rgba(244, 110, 54, 0.12),
        transparent 28rem
      ),
      #090909;
    color: #f5f1e8;
  }

  .shell {
    width: min(1180px, calc(100% - 32px));
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
    background: rgba(9, 9, 9, 0.8);
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

  .nav nav a,
  .login,
  .text-link {
    text-decoration: none;
  }

  .nav nav a:hover,
  .text-link:hover {
    color: #75d69c;
  }

  .login {
    justify-content: center;
    border: 1px solid rgba(245, 241, 232, 0.18);
    border-radius: 999px;
    padding: 8px 14px;
    color: #f5f1e8;
    font-weight: 800;
  }

  .hero {
    min-height: 720px;
    display: grid;
    grid-template-columns: minmax(0, 0.95fr) minmax(420px, 1.05fr);
    align-items: center;
    gap: clamp(34px, 7vw, 88px);
    padding: clamp(38px, 6vw, 70px) 0 42px;
  }

  .platform-cloud {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    max-width: 520px;
    margin-bottom: 28px;
  }

  .platform-cloud span {
    border: 1px solid rgba(245, 241, 232, 0.16);
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.05);
    padding: 7px 10px;
    color: #c9c1b8;
    font-size: 0.84rem;
    font-weight: 800;
  }

  .eyebrow,
  .section-kicker {
    margin: 0;
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

  h1 {
    max-width: 10ch;
    margin-bottom: 24px;
    font-family:
      Charter,
      Iowan Old Style,
      Georgia,
      serif;
    font-size: clamp(4.2rem, 7.2vw, 7.8rem);
    line-height: 0.82;
    letter-spacing: 0;
  }

  h2 {
    font-family:
      Charter,
      Iowan Old Style,
      Georgia,
      serif;
    font-size: clamp(2.6rem, 6vw, 5.7rem);
    line-height: 0.9;
    letter-spacing: 0;
  }

  .lede {
    max-width: 640px;
    color: #d5cec4;
    font-size: clamp(1.18rem, 2vw, 1.52rem);
    line-height: 1.5;
  }

  .hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 34px;
  }

  .button {
    display: inline-flex;
    min-height: 48px;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    padding: 0 20px;
    text-decoration: none;
    font-weight: 900;
  }

  .primary {
    background: #75d69c;
    color: #07120c;
  }

  .secondary {
    border: 1px solid rgba(245, 241, 232, 0.24);
    background: rgba(255, 255, 255, 0.04);
    color: #f5f1e8;
  }

  .microcopy {
    margin-top: 16px;
    color: #938b83;
    font-size: 0.92rem;
  }

  .demo-stage {
    overflow: hidden;
    border: 1px solid rgba(245, 241, 232, 0.15);
    border-radius: 12px;
    background: #111;
    box-shadow: 22px 22px 0 #3520a6;
  }

  .demo-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    border-bottom: 1px solid rgba(245, 241, 232, 0.1);
    padding: 12px;
    color: #938b83;
  }

  .demo-toolbar span {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: #ef6b3a;
  }

  .demo-toolbar span:nth-child(2) {
    background: #75d69c;
  }

  .demo-toolbar span:nth-child(3) {
    background: #65bcff;
  }

  .demo-toolbar strong {
    margin-left: auto;
    font-size: 0.82rem;
  }

  .demo-screen {
    position: relative;
    background: #060606;
  }

  .demo-screen img {
    width: 100%;
    opacity: 0.82;
  }

  .play-button {
    position: absolute;
    inset: 50% auto auto 50%;
    display: grid;
    width: 74px;
    height: 74px;
    place-items: center;
    transform: translate(-50%, -50%);
    border-radius: 999px;
    background: #75d69c;
    color: #07120c;
    font-size: 1.55rem;
  }

  .demo-caption {
    display: grid;
    gap: 4px;
    padding: 18px;
    color: #d5cec4;
  }

  .demo-caption span {
    color: #938b83;
    line-height: 1.45;
  }

  .proof {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
    margin-top: 0;
  }

  .proof div {
    border: 1px solid rgba(245, 241, 232, 0.12);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.04);
    padding: 22px;
  }

  .proof strong {
    display: block;
    font-family:
      Charter,
      Iowan Old Style,
      Georgia,
      serif;
    font-size: 2.4rem;
  }

  .proof span {
    color: #938b83;
    font-size: 0.9rem;
  }

  section {
    padding: clamp(62px, 9vw, 120px) 0;
  }

  .agentic {
    text-align: center;
  }

  .agentic h2 {
    max-width: 960px;
    margin: 12px auto 34px;
  }

  .agent-grid,
  .feature-grid,
  .plans,
  .steps {
    display: grid;
    gap: 14px;
  }

  .agent-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .agent-card {
    min-height: 260px;
    border: 1px solid rgba(245, 241, 232, 0.12);
    border-radius: 12px;
    padding: 26px;
    text-align: left;
  }

  .agent-card p {
    color: #f5f1e8;
    font-weight: 900;
  }

  .agent-card h3 {
    margin: 74px 0 0;
    font-size: 1.45rem;
    line-height: 1.25;
  }

  .agent-card.blue {
    background: #102d5d;
  }

  .agent-card.green {
    background: #173a24;
  }

  .agent-card.orange {
    background: #4b2013;
  }

  .section-heading {
    display: grid;
    grid-template-columns: minmax(0, 0.85fr) minmax(300px, 0.65fr);
    gap: clamp(24px, 6vw, 70px);
    align-items: end;
    margin-bottom: 32px;
  }

  .section-heading h2,
  .product-shot h2,
  .platforms h2,
  .open-core h2,
  .faq h2 {
    margin-top: 12px;
  }

  .section-heading p,
  .product-shot p,
  .platforms p,
  .open-core p,
  .faq p {
    color: #b9b1a6;
    font-size: 1.05rem;
    line-height: 1.6;
  }

  .steps {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .steps article,
  .feature-grid article,
  .plans article {
    border: 1px solid rgba(245, 241, 232, 0.12);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.045);
    padding: 24px;
  }

  .steps span {
    display: grid;
    width: 40px;
    height: 40px;
    place-items: center;
    border-radius: 999px;
    background: #75d69c;
    color: #07120c;
    font-weight: 900;
  }

  .steps h3,
  .feature-grid h3 {
    margin: 24px 0 8px;
    font-size: 1.3rem;
  }

  .steps p,
  .feature-grid p {
    margin-bottom: 0;
    color: #938b83;
    line-height: 1.55;
  }

  .feature-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .feature-grid article:nth-child(2),
  .feature-grid article:nth-child(5) {
    background: rgba(53, 32, 166, 0.42);
  }

  .product-shot {
    display: grid;
    grid-template-columns: minmax(0, 0.75fr) minmax(420px, 1.1fr);
    gap: clamp(30px, 6vw, 80px);
    align-items: center;
  }

  .text-link {
    color: #75d69c;
    font-weight: 900;
  }

  .shot-stack {
    position: relative;
    min-height: 440px;
  }

  .shot-stack img {
    position: absolute;
    width: 72%;
    border: 1px solid rgba(245, 241, 232, 0.12);
    border-radius: 10px;
    box-shadow: 0 24px 90px rgba(0, 0, 0, 0.38);
  }

  .shot-stack img:first-child {
    top: 0;
    right: 0;
  }

  .shot-stack img:last-child {
    bottom: 0;
    left: 0;
  }

  .platforms {
    text-align: center;
  }

  .platforms h2 {
    max-width: 850px;
    margin-right: auto;
    margin-left: auto;
  }

  .platform-grid {
    display: grid;
    grid-template-columns: repeat(9, minmax(0, 1fr));
    gap: 10px;
    margin: 34px 0;
  }

  .platform-grid div {
    border: 1px solid rgba(245, 241, 232, 0.12);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.05);
    padding: 18px 10px;
    font-weight: 900;
  }

  .platforms p {
    max-width: 760px;
    margin-right: auto;
    margin-left: auto;
  }

  .plans {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .plans article {
    display: flex;
    flex-direction: column;
    min-height: 260px;
  }

  .plans article.featured {
    background: #75d69c;
    color: #07120c;
  }

  .plans span {
    font-weight: 900;
  }

  .plans strong {
    display: block;
    margin: 26px 0 14px;
    font-family:
      Charter,
      Iowan Old Style,
      Georgia,
      serif;
    font-size: 4.2rem;
  }

  .plans p {
    color: inherit;
    line-height: 1.5;
    opacity: 0.78;
  }

  .plans ul {
    display: grid;
    gap: 8px;
    margin: 22px 0;
    padding: 0;
    list-style: none;
    color: inherit;
    opacity: 0.82;
  }

  .plans li::before {
    content: "✓";
    margin-right: 8px;
    font-weight: 900;
  }

  .plan-cta {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    margin-top: auto;
    border: 1px solid currentColor;
    border-radius: 999px;
    padding: 12px 16px;
    text-decoration: none;
    font-weight: 900;
  }

  .plan-cta:hover {
    transform: translateY(-1px);
  }

  .open-core {
    display: grid;
    grid-template-columns: minmax(0, 0.8fr) minmax(320px, 0.8fr);
    gap: clamp(24px, 6vw, 72px);
    align-items: center;
    border-top: 1px solid rgba(245, 241, 232, 0.12);
    border-bottom: 1px solid rgba(245, 241, 232, 0.12);
  }

  .open-core a {
    color: #75d69c;
    font-weight: 900;
  }

  .faq h2 {
    max-width: 720px;
  }

  .faq-list {
    display: grid;
    gap: 10px;
    margin-top: 28px;
  }

  details {
    border: 1px solid rgba(245, 241, 232, 0.12);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.04);
    padding: 18px 20px;
  }

  summary {
    cursor: pointer;
    font-weight: 900;
  }

  details p {
    margin: 14px 0 0;
  }

  .final-cta {
    margin-top: 40px;
    background: #141414;
    text-align: center;
  }

  .final-cta .shell {
    display: grid;
    gap: 24px;
    place-items: center;
  }

  .final-cta h2 {
    max-width: 760px;
    margin-bottom: 0;
  }

  .footer {
    display: flex;
    justify-content: space-between;
    gap: 24px;
    padding: 38px 0 64px;
    color: #938b83;
  }

  .footer img {
    width: 156px;
    margin-bottom: 12px;
  }

  .footer p {
    margin-bottom: 0;
  }

  .footer nav {
    display: flex;
    gap: 16px;
  }

  .footer a {
    color: #d5cec4;
    text-decoration: none;
  }

  @media (max-width: 980px) {
    .nav {
      grid-template-columns: 1fr auto;
    }

    .nav nav {
      display: none;
    }

    .hero,
    .section-heading,
    .product-shot,
    .open-core {
      grid-template-columns: 1fr;
    }

    .hero {
      min-height: auto;
      padding-top: 34px;
    }

    .demo-stage {
      box-shadow: 10px 10px 0 #3520a6;
    }

    .proof,
    .agent-grid,
    .steps,
    .feature-grid,
    .plans {
      grid-template-columns: 1fr 1fr;
    }

    .platform-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }

  @media (max-width: 620px) {
    .shell {
      width: min(100% - 28px, 1180px);
    }

    h1 {
      font-size: clamp(3.7rem, 16vw, 5.2rem);
    }

    .hero-actions,
    .footer,
    .footer nav {
      align-items: stretch;
      flex-direction: column;
    }

    .button {
      width: 100%;
    }

    .proof,
    .agent-grid,
    .steps,
    .feature-grid,
    .plans,
    .platform-grid {
      grid-template-columns: 1fr;
    }

    .shot-stack {
      min-height: auto;
      display: grid;
      gap: 12px;
    }

    .shot-stack img {
      position: static;
      width: 100%;
    }
  }
</style>
