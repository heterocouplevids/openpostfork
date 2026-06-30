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
	import { client } from '$lib/api/client';
	import { auth } from '$lib/stores/auth';
	import BotIcon from 'lucide-svelte/icons/bot';
	import ShieldCheckIcon from 'lucide-svelte/icons/shield-check';
	import XIcon from 'lucide-svelte/icons/x';

	let authState = $derived($auth);
	let error = $state('');
	let submitting = $state(false);

	let params = $derived({
		response_type: $page.url.searchParams.get('response_type') ?? '',
		client_id: $page.url.searchParams.get('client_id') ?? '',
		redirect_uri: $page.url.searchParams.get('redirect_uri') ?? '',
		scope: $page.url.searchParams.get('scope') ?? 'mcp:full',
		state: $page.url.searchParams.get('state') ?? '',
		code_challenge: $page.url.searchParams.get('code_challenge') ?? '',
		code_challenge_method: $page.url.searchParams.get('code_challenge_method') ?? '',
		resource: $page.url.searchParams.get('resource') ?? ''
	});

	let scopes = $derived(
		(params.scope || 'mcp:full')
			.split(/[,\s]+/)
			.map((scope) => scope.trim())
			.filter(Boolean)
	);

	let clientLabel = $derived(clientDisplayName(params.client_id));
	let redirectHost = $derived(hostname(params.redirect_uri));
	let requestError = $derived(validateRequest());

	function currentPath() {
		return `${$page.url.pathname}${$page.url.search}`;
	}

	function loginRedirect() {
		return `/login?redirect=${encodeURIComponent(currentPath())}`;
	}

	function clientDisplayName(clientID: string) {
		if (!clientID) return 'MCP client';
		const host = hostname(clientID);
		if (host) return host;
		return clientID;
	}

	function hostname(value: string) {
		try {
			return new URL(value).hostname;
		} catch {
			return '';
		}
	}

	function validateRequest() {
		if (params.response_type !== 'code') return 'This OAuth request is missing response_type=code.';
		if (!params.client_id) return 'This OAuth request is missing a client ID.';
		if (!params.redirect_uri) return 'This OAuth request is missing a redirect URI.';
		if (!params.code_challenge) return 'This OAuth request is missing a PKCE challenge.';
		if (params.code_challenge_method !== 'S256') return 'This OAuth request must use PKCE S256.';
		return '';
	}

	async function submit(approved: boolean) {
		if (approved && requestError) {
			error = '';
			return;
		}
		if (!approved && (!params.redirect_uri || !params.client_id)) {
			error = '';
			return;
		}

		submitting = true;
		error = '';

		try {
			const { data, error: apiError } = await (client as any).POST('/mcp/oauth/authorize', {
				body: { ...params, approved }
			});
			if (apiError || !data?.redirect_url) {
				throw new Error(apiError?.detail ?? 'Failed to finish OAuth authorization');
			}
			window.location.href = data.redirect_url;
		} catch (e) {
			error = (e as Error).message;
		} finally {
			submitting = false;
		}
	}

	$effect(() => {
		if (authState.isLoading) return;
		if (!authState.user && !authState.isAuthenticated) {
			goto(loginRedirect());
			return;
		}
	});
</script>

<svelte:head>
	<title>Authorize OpenPost MCP</title>
</svelte:head>

<div class="flex min-h-[80vh] flex-col items-center justify-center px-4 py-10">
	<Card class="w-full max-w-lg">
		<CardHeader class="space-y-3 text-center">
			<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
				<BotIcon class="h-6 w-6 text-primary" />
			</div>
			<CardTitle class="text-lg font-semibold">Authorize OpenPost MCP</CardTitle>
			<CardDescription>
				{clientLabel} wants to connect to your OpenPost assistant tools.
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if error || requestError}
				<div
					class="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
				>
					{error || requestError}
				</div>
			{/if}

			<div class="space-y-5">
				<div class="rounded-md border bg-muted/30 p-4">
					<div class="text-xl font-semibold">{clientLabel}</div>
					{#if redirectHost}
						<p class="mt-1 text-sm text-muted-foreground">Redirects to {redirectHost}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<p class="text-sm font-medium">Requested access</p>
					<div class="flex flex-wrap gap-2">
						{#each scopes as scope (scope)}
							<Badge>{scope}</Badge>
						{:else}
							<Badge>mcp:full</Badge>
						{/each}
					</div>
				</div>

				<div class="flex flex-col gap-2 sm:flex-row">
					<Button
						class="w-full gap-2"
						onclick={() => submit(true)}
						disabled={submitting || !!requestError}
					>
						<ShieldCheckIcon class="h-4 w-4" />
						Authorize
					</Button>
					<Button
						variant="outline"
						class="w-full gap-2"
						onclick={() => submit(false)}
						disabled={submitting || !params.redirect_uri}
					>
						<XIcon class="h-4 w-4" />
						Deny
					</Button>
				</div>
			</div>
		</CardContent>
	</Card>
</div>
