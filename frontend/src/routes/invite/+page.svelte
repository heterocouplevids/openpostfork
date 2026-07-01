<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader } from '$lib/components/ui/card';
	import { client } from '$lib/api/client';
	import { auth } from '$lib/stores/auth';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import ShieldCheckIcon from 'lucide-svelte/icons/shield-check';
	import UsersIcon from 'lucide-svelte/icons/users';

	let authState = $derived($auth);
	let token = $derived($page.url.searchParams.get('token') ?? '');
	let attemptedToken = $state('');
	let loading = $state(false);
	let accepted = $state(false);
	let error = $state('');
	let workspaceID = $state('');
	let role = $state('');

	function loginRedirect(inviteToken: string) {
		return `/login?redirect=${encodeURIComponent(`/invite?token=${encodeURIComponent(inviteToken)}`)}`;
	}

	async function acceptInvitation(inviteToken: string) {
		if (!inviteToken || attemptedToken === inviteToken) return;
		attemptedToken = inviteToken;
		loading = true;
		error = '';
		try {
			const { data, error: apiError } = await (client as any).POST(
				'/workspace-invitations/accept',
				{
					body: { token: inviteToken }
				}
			);
			if (apiError || !data) {
				throw new Error(apiError?.detail || 'Failed to accept workspace invitation');
			}
			workspaceID = data.workspace_id;
			role = data.role;
			accepted = true;
			await workspaceCtx.initialize(data.workspace_id);
		} catch (e) {
			error = (e as Error).message;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (authState.isLoading) return;

		if (!token) {
			error = 'Invite token is missing.';
			return;
		}

		if (!authState.isAuthenticated) {
			goto(loginRedirect(token));
			return;
		}

		acceptInvitation(token);
	});
</script>

<svelte:head>
	<title>Accept Invitation - OpenPost</title>
</svelte:head>

<div class="flex min-h-[80vh] flex-col items-center justify-center px-4 py-10">
	<Card class="w-full max-w-lg">
		<CardHeader class="space-y-3 text-center">
			<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
				{#if accepted}
					<ShieldCheckIcon class="h-6 w-6 text-primary" />
				{:else}
					<UsersIcon class="h-6 w-6 text-primary" />
				{/if}
			</div>
			<h1 class="text-lg font-semibold">
				{accepted ? 'Invitation accepted' : 'Workspace invitation'}
			</h1>
			<CardDescription>
				{#if accepted}
					You now have {role} access to this workspace.
				{:else}
					Sign in with the invited email address to join the workspace.
				{/if}
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if error}
				<div
					data-testid="invite-error"
					class="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
				>
					{error}
				</div>
			{/if}

			{#if loading}
				<div class="flex items-center justify-center gap-2 text-sm text-muted-foreground">
					<LoaderIcon class="h-4 w-4 animate-spin" />
					Accepting invitation...
				</div>
			{:else if accepted}
				<div class="space-y-4">
					<div class="rounded-md border bg-muted/30 p-4 text-sm">
						<p class="font-medium">Workspace joined</p>
						<p class="mt-1 font-mono text-xs text-muted-foreground">{workspaceID}</p>
					</div>
					<Button class="w-full" onclick={() => goto('/settings?tab=organization')}
						>Open Settings</Button
					>
				</div>
			{:else if !error}
				<p class="text-center text-sm text-muted-foreground">Checking invitation...</p>
			{/if}
		</CardContent>
	</Card>
</div>
