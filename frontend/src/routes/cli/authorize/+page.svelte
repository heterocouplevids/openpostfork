<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { client } from '$lib/api/client';
	import { auth } from '$lib/stores/auth';
	import { m } from '$lib/paraglide/messages';
	import TerminalIcon from 'lucide-svelte/icons/terminal';
	import ShieldCheckIcon from 'lucide-svelte/icons/shield-check';
	import XIcon from 'lucide-svelte/icons/x';

	type CLIAuthSession = {
		client_name: string;
		client_version?: string;
		client_os?: string;
		requested_scopes?: string;
		expires_at?: string;
	};

	let authState = $derived($auth);
	let userCode = $derived($page.url.searchParams.get('user_code') ?? '');
	let session = $state<CLIAuthSession | null>(null);
	let tokenName = $state('OpenPost CLI');
	let error = $state('');
	let loading = $state(false);
	let submitting = $state(false);
	let completed = $state<'approved' | 'denied' | null>(null);
	let loadedUserCode = $state('');

	let scopes = $derived(
		(session?.requested_scopes ?? '')
			.split(/[,\s]+/)
			.map((scope) => scope.trim())
			.filter(Boolean)
	);

	function loginRedirect(code: string) {
		return `/login?redirect=${encodeURIComponent(`/cli/authorize?user_code=${encodeURIComponent(code)}`)}`;
	}

	async function loadSession(code: string) {
		if (!code || loadedUserCode === code) return;

		loading = true;
		error = '';

		try {
			const { data, error: apiError } = await (client as any).GET('/cli/auth/session', {
				params: { query: { user_code: code } }
			});

			if (apiError || !data) {
				throw new Error(apiError?.detail ?? m.cli_authorize_load_failed());
			}

			session = data as CLIAuthSession;
			loadedUserCode = code;
		} catch (e) {
			error = (e as Error).message;
		} finally {
			loading = false;
		}
	}

	async function approve() {
		await submitDecision('approved');
	}

	async function deny() {
		await submitDecision('denied');
	}

	async function submitDecision(decision: 'approved' | 'denied') {
		if (!userCode) return;

		submitting = true;
		error = '';

		try {
			const path = decision === 'approved' ? '/cli/auth/approve' : '/cli/auth/deny';
			const body =
				decision === 'approved'
					? { user_code: userCode, name: tokenName || 'OpenPost CLI' }
					: { user_code: userCode };
			const { error: apiError } = await (client as any).POST(path, { body });

			if (apiError) {
				throw new Error(apiError?.detail ?? m.cli_authorize_decision_failed());
			}

			completed = decision;
		} catch (e) {
			error = (e as Error).message;
		} finally {
			submitting = false;
		}
	}

	$effect(() => {
		if (authState.isLoading) return;

		if (!userCode) {
			error = m.cli_authorize_missing_code();
			return;
		}

		if (!authState.user && !authState.isAuthenticated) {
			goto(loginRedirect(userCode));
			return;
		}

		loadSession(userCode);
	});
</script>

<svelte:head>
	<title>{m.cli_authorize_title()}</title>
</svelte:head>

<div class="flex min-h-[80vh] flex-col items-center justify-center px-4 py-10">
	<Card class="w-full max-w-lg">
		<CardHeader class="space-y-3 text-center">
			<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
				<TerminalIcon class="h-6 w-6 text-primary" />
			</div>
			<CardTitle class="text-lg font-semibold">
				{#if completed === 'approved'}
					{m.cli_authorize_approved()}
				{:else if completed === 'denied'}
					{m.cli_authorize_denied()}
				{:else}
					{m.cli_authorize_heading()}
				{/if}
			</CardTitle>
			<CardDescription>
				{#if completed}
					{m.cli_authorize_close_tab()}
				{:else}
					{m.cli_authorize_description()}
				{/if}
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if error}
				<div
					class="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
				>
					{error}
				</div>
			{/if}

			{#if completed}
				<div class="flex justify-center">
					<div class="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
						{#if completed === 'approved'}
							<ShieldCheckIcon class="h-6 w-6 text-primary" />
						{:else}
							<XIcon class="h-6 w-6 text-muted-foreground" />
						{/if}
					</div>
				</div>
			{:else if loading}
				<p class="text-center text-sm text-muted-foreground">{m.cli_authorize_loading()}</p>
			{:else if session}
				<div class="space-y-5">
					<div class="rounded-md border bg-muted/30 p-4">
						<div class="text-xl font-semibold">{session.client_name}</div>
						<p class="mt-1 text-sm text-muted-foreground">
							{session.client_version || m.cli_authorize_unknown_version()} · {session.client_os ||
								m.cli_authorize_unknown_os()}
						</p>
					</div>

					<div class="space-y-2">
						<Label>{m.cli_authorize_scopes()}</Label>
						<div class="flex flex-wrap gap-2">
							{#each scopes as scope (scope)}
								<Badge>{scope}</Badge>
							{:else}
								<Badge>{m.cli_authorize_default_scope()}</Badge>
							{/each}
						</div>
					</div>

					<div class="space-y-2">
						<Label for="token-name">{m.cli_authorize_token_name()}</Label>
						<Input id="token-name" bind:value={tokenName} autocomplete="off" />
					</div>

					<div class="flex flex-col gap-2 sm:flex-row">
						<Button class="w-full gap-2" onclick={approve} disabled={submitting}>
							<ShieldCheckIcon class="h-4 w-4" />
							{m.cli_authorize_approve()}
						</Button>
						<Button variant="outline" class="w-full gap-2" onclick={deny} disabled={submitting}>
							<XIcon class="h-4 w-4" />
							{m.cli_authorize_deny()}
						</Button>
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>
</div>
