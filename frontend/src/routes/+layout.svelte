<script lang="ts">
	import '../app.css';
	import './layout.css';
	import { ModeWatcher } from 'mode-watcher';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import SidebarLeft from '$lib/components/sidebar-left.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import LanguageSwitcher from '$lib/components/language-switcher.svelte';
	import { IS_CAPACITOR } from '$lib/env';
	import { instanceStore, isInstanceConfigured } from '$lib/stores/instance.svelte';
	import { client } from '$lib/api/client';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { m } from '$lib/paraglide/messages';
	import { onboardingPathForPlan } from '$lib/billing';
	import { safeSameOriginRedirect } from '$lib/redirects';

	let { children } = $props();

	const instance = instanceStore();

	let authState = $derived($auth);
	let currentPath = $derived($page.url.pathname);
	const publicRoutes = [
		'/login',
		'/register',
		'/connect',
		'/demo',
		'/demo/paraglide',
		'/invite',
		'/cli/authorize',
		'/oauth/authorize',
		'/accounts/mastodon/callback',
		'/accounts/callback'
	];

	const standaloneRoutes = [
		'/onboarding',
		'/connect',
		'/invite',
		'/cli/authorize',
		'/oauth/authorize',
		'/accounts/mastodon/callback',
		'/accounts/callback'
	];

	let needsOnboarding = $state(false);
	let onboardingChecked = $state(false);
	let lastOnboardingCheckedPath = $state('');
	let onboardingCheckInFlightForPath = $state('');
	let onboardingFreshForPath = $derived(
		onboardingChecked && lastOnboardingCheckedPath === currentPath
	);

	function authenticatedPublicTarget() {
		if (currentPath !== '/login') return '/';
		const target = safeSameOriginRedirect($page.url);
		if (target === '/login' || target.startsWith('/login?') || target.startsWith('/register')) {
			return '/';
		}
		return target;
	}

	onMount(() => {
		instance.initialize();
		auth.initialize();
	});

	$effect(() => {
		if (instance.isLoading) return;

		if (IS_CAPACITOR && !isInstanceConfigured() && currentPath !== '/connect') {
			goto('/connect');
			return;
		}

		if (authState.isLoading) return;

		const isPublicRoute = publicRoutes.some((route) => currentPath.startsWith(route));
		const isOnboardingPage = currentPath === '/onboarding';

		if (!authState.isAuthenticated && !isPublicRoute && !isOnboardingPage) {
			goto('/login');
		}

		if (authState.isAuthenticated) {
			if (!onboardingFreshForPath) return;

			if (needsOnboarding) {
				if (!isOnboardingPage && currentPath !== '/invite') {
					goto(onboardingPathForPlan($page.url.searchParams.get('plan')));
				}
			} else if (currentPath === '/login' || currentPath === '/register') {
				goto(authenticatedPublicTarget());
			}
		}
	});

	async function checkOnboarding(path: string) {
		if (!authState.isAuthenticated || authState.isLoading) return;
		onboardingCheckInFlightForPath = path;
		let nextNeedsOnboarding = false;
		try {
			const { data, error } = await client.GET('/workspaces');
			if (!error && data && data.length === 0) {
				nextNeedsOnboarding = true;
			} else {
				nextNeedsOnboarding = !!error;
				// Initialize workspace context after successful workspace load
				if (!error && data) {
					await workspaceCtx.initialize();
				}
			}
		} catch {
			// Fail safe: if we cannot verify workspace state, keep user in onboarding flow.
			nextNeedsOnboarding = true;
		} finally {
			if (onboardingCheckInFlightForPath === path) {
				onboardingCheckInFlightForPath = '';
			}
		}
		if (path !== currentPath) return;
		needsOnboarding = nextNeedsOnboarding;
		onboardingChecked = true;
	}

	$effect(() => {
		if (authState.isLoading || !authState.isAuthenticated) {
			onboardingChecked = false;
			lastOnboardingCheckedPath = '';
			onboardingCheckInFlightForPath = '';
			return;
		}

		if (currentPath !== lastOnboardingCheckedPath) {
			onboardingChecked = false;
			lastOnboardingCheckedPath = currentPath;
		}

		if (!onboardingChecked && onboardingCheckInFlightForPath !== currentPath) {
			checkOnboarding(currentPath);
		}
	});
</script>

<svelte:head>
	<title>OpenPost</title>
</svelte:head>

<ModeWatcher />
{#if instance.isLoading || authState.isLoading || (authState.isAuthenticated && !onboardingFreshForPath)}
	<div class="flex min-h-screen flex-col items-center justify-center gap-3">
		<Skeleton class="h-12 w-12 rounded-lg" />
		<Skeleton class="h-3 w-32 rounded" />
		<Skeleton class="h-3 w-24 rounded" />
	</div>
{:else if !authState.isAuthenticated}
	<div class="fixed top-4 right-4 z-20">
		<LanguageSwitcher compact />
	</div>
	{#if currentPath === '/'}
		<div class="flex min-h-[80vh] items-center justify-center">
			<div class="mx-auto max-w-md px-4 py-12 text-center">
				<div class="mb-6 flex justify-center">
					<Logo width={100} height={29} />
				</div>
				<p class="mb-6 text-muted-foreground">{m.landing_tagline()}</p>
				<div class="flex justify-center gap-4">
					<a
						href="/login"
						class="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
						>{m.landing_sign_in()}</a
					>
					<a
						href="/register"
						class="inline-flex items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
						>{m.landing_create_account()}</a
					>
				</div>
			</div>
		</div>
	{:else}
		{@render children()}
	{/if}
{:else if standaloneRoutes.includes(currentPath)}
	<div class="fixed top-4 right-4 z-20">
		<LanguageSwitcher compact />
	</div>
	{@render children()}
{:else}
	<Sidebar.Provider>
		<SidebarLeft />
		<Sidebar.Inset>
			<!-- Mobile header -->
			<div class="flex items-center gap-2 border-b px-3 py-2 md:hidden">
				<Sidebar.Trigger />
				<span class="text-sm font-medium">OpenPost</span>
			</div>
			<div class="flex flex-1 flex-col overflow-auto">
				{@render children()}
			</div>
		</Sidebar.Inset>
	</Sidebar.Provider>
{/if}
