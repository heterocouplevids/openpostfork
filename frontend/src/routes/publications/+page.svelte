<script lang="ts">
	import { goto } from '$app/navigation';
	import { client, type Publication } from '$lib/api/client';
	import MediaUpload from '$lib/components/media-upload.svelte';
	import PageContainer from '$lib/components/page-container.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { getStatusColor } from '$lib/utils';
	import CheckIcon from 'lucide-svelte/icons/check';
	import ClipboardIcon from 'lucide-svelte/icons/clipboard';
	import FileTextIcon from 'lucide-svelte/icons/file-text';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import RefreshCwIcon from 'lucide-svelte/icons/refresh-cw';
	import SendIcon from 'lucide-svelte/icons/send';

	const statusFilters = ['all', 'draft', 'ready', 'scheduled', 'published', 'failed'];

	let publications = $state<Publication[]>([]);
	let loading = $state(false);
	let creating = $state(false);
	let savingStatus = $state('');
	let selectedStatus = $state('all');
	let selectedPublicationId = $state<string | null>(null);
	let errorMessage = $state('');
	let toastMessage = $state('');
	let lastLoadKey = '';

	let newTitle = $state('');
	let newSourceContent = $state('');
	let newSourceURL = $state('');
	let newGoal = $state('');
	let newAudience = $state('');
	let newMediaIds = $state<string[]>([]);

	const selectedPublication = $derived(
		publications.find((item) => item.id === selectedPublicationId) ?? publications[0] ?? null
	);
	const workspaceID = $derived(workspaceCtx.currentWorkspace?.id ?? '');
	const canCreate = $derived(
		workspaceID !== '' &&
			newTitle.trim().length > 0 &&
			newSourceContent.trim().length > 0 &&
			!creating
	);

	$effect(() => {
		const key = `${workspaceID}:${selectedStatus}`;
		if (!workspaceID || key === lastLoadKey) return;
		lastLoadKey = key;
		loadPublications();
	});

	function problemMessage(error: unknown, fallback: string) {
		if (!error) return fallback;
		const candidate = error as { detail?: string; message?: string; title?: string };
		return candidate.detail || candidate.message || candidate.title || fallback;
	}

	async function loadPublications() {
		if (!workspaceID) return;
		loading = true;
		errorMessage = '';
		try {
			const query: { workspace_id: string; status?: string; limit: number } = {
				workspace_id: workspaceID,
				limit: 100
			};
			if (selectedStatus !== 'all') query.status = selectedStatus;
			const { data, error } = await client.GET('/publications', {
				params: { query }
			});
			if (error) throw error;
			publications = data ?? [];
			if (!publications.some((item) => item.id === selectedPublicationId)) {
				selectedPublicationId = publications[0]?.id ?? null;
			}
		} catch (error) {
			publications = [];
			selectedPublicationId = null;
			errorMessage = problemMessage(error, 'Failed to load publications');
		} finally {
			loading = false;
		}
	}

	async function createPublication() {
		if (!canCreate) return;
		creating = true;
		errorMessage = '';
		try {
			const { data, error } = await client.POST('/publications', {
				body: {
					workspace_id: workspaceID,
					title: newTitle.trim(),
					source_content: newSourceContent.trim(),
					source_url: newSourceURL.trim() || undefined,
					goal: newGoal.trim() || undefined,
					audience: newAudience.trim() || undefined,
					media_ids: newMediaIds.length > 0 ? newMediaIds : undefined
				}
			});
			if (error) throw error;
			if (data) {
				publications = [data, ...publications.filter((item) => item.id !== data.id)];
				selectedPublicationId = data.id;
			}
			resetCreateForm();
			showToast('Publication created');
		} catch (error) {
			errorMessage = problemMessage(error, 'Failed to create publication');
		} finally {
			creating = false;
		}
	}

	async function setPublicationStatus(publication: Publication, status: string) {
		if (publication.status === status) return;
		savingStatus = status;
		errorMessage = '';
		try {
			const { data, error } = await client.PATCH('/publications/{id}', {
				params: { path: { id: publication.id } },
				body: { status }
			});
			if (error) throw error;
			if (data) {
				publications = publications.map((item) => (item.id === data.id ? data : item));
				selectedPublicationId = data.id;
			}
			showToast(`Marked ${status}`);
		} catch (error) {
			errorMessage = problemMessage(error, 'Failed to update publication');
		} finally {
			savingStatus = '';
		}
	}

	function resetCreateForm() {
		newTitle = '';
		newSourceContent = '';
		newSourceURL = '';
		newGoal = '';
		newAudience = '';
		newMediaIds = [];
	}

	async function copyPublicationID(publication: Publication) {
		try {
			await navigator.clipboard.writeText(publication.id);
			showToast('Publication ID copied');
		} catch {
			errorMessage = 'Clipboard access is unavailable in this browser';
		}
	}

	function composeFromPublication(publication: Publication) {
		goto(`/?publication=${encodeURIComponent(publication.id)}`);
	}

	function showToast(message: string) {
		toastMessage = message;
		window.setTimeout(() => {
			if (toastMessage === message) toastMessage = '';
		}, 2600);
	}

	function formatDate(value: string) {
		if (!value) return '-';
		return new Intl.DateTimeFormat(undefined, {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		}).format(new Date(value));
	}

	function preview(text: string, limit = 170) {
		const compact = text.replace(/\s+/g, ' ').trim();
		if (compact.length <= limit) return compact;
		return compact.slice(0, limit - 3).trim() + '...';
	}

	function mediaCount(publication: Publication) {
		return publication.media_ids?.length ?? 0;
	}

	function statusLabel(status: string) {
		return status.charAt(0).toUpperCase() + status.slice(1);
	}
</script>

<svelte:head>
	<title>Publications - OpenPost</title>
</svelte:head>

{#if toastMessage}
	<div
		class="fixed right-4 bottom-4 z-50 rounded-md border bg-background px-3 py-2 text-sm shadow-lg"
	>
		{toastMessage}
	</div>
{/if}

{#snippet actions()}
	<Button
		variant="outline"
		class="gap-2"
		onclick={loadPublications}
		disabled={loading || !workspaceID}
	>
		<RefreshCwIcon class={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
		Refresh
	</Button>
{/snippet}

<PageContainer
	title="Publications"
	description="Source ideas, briefs, and campaign context before they become platform posts."
	icon={FileTextIcon}
	{actions}
	loading={!workspaceCtx.currentWorkspace}
	loadingMessage="Loading workspace..."
>
	<div class="space-y-5">
		{#if errorMessage}
			<div
				class="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive"
			>
				{errorMessage}
			</div>
		{/if}

		<section class="rounded-lg border bg-card/55 p-4">
			<div class="grid gap-4 lg:grid-cols-[minmax(0,1.15fr)_minmax(280px,0.85fr)]">
				<div class="space-y-3">
					<div>
						<h2 class="text-sm font-semibold">Create source</h2>
						<p class="mt-1 max-w-2xl text-sm text-muted-foreground">
							Capture the canonical idea once, then reference it from posts, threads, CLI, or MCP.
						</p>
					</div>
					<div class="grid gap-3 sm:grid-cols-2">
						<div class="space-y-1.5">
							<label for="publication-title" class="text-xs font-medium text-muted-foreground">
								Title
							</label>
							<Input
								id="publication-title"
								bind:value={newTitle}
								placeholder="June launch notes"
								disabled={creating}
							/>
						</div>
						<div class="space-y-1.5">
							<label for="publication-url" class="text-xs font-medium text-muted-foreground">
								Source URL
							</label>
							<Input
								id="publication-url"
								bind:value={newSourceURL}
								placeholder="https://..."
								disabled={creating}
							/>
						</div>
					</div>
					<div class="grid gap-3 sm:grid-cols-2">
						<div class="space-y-1.5">
							<label for="publication-goal" class="text-xs font-medium text-muted-foreground">
								Goal
							</label>
							<Input
								id="publication-goal"
								bind:value={newGoal}
								placeholder="announce, explain, launch"
								disabled={creating}
							/>
						</div>
						<div class="space-y-1.5">
							<label for="publication-audience" class="text-xs font-medium text-muted-foreground">
								Audience
							</label>
							<Input
								id="publication-audience"
								bind:value={newAudience}
								placeholder="builders, customers, founders"
								disabled={creating}
							/>
						</div>
					</div>
					<div class="space-y-1.5">
						<label for="publication-source" class="text-xs font-medium text-muted-foreground">
							Source content
						</label>
						<Textarea
							id="publication-source"
							bind:value={newSourceContent}
							rows={7}
							placeholder="Paste the brief, source idea, article notes, or launch context."
							disabled={creating}
						/>
					</div>
				</div>
				<div class="flex flex-col gap-3 rounded-md border bg-background/55 p-3">
					<div>
						<h3 class="text-sm font-medium">Source media</h3>
						<p class="mt-1 text-xs text-muted-foreground">
							Attach optional assets that assistants and drafts can reuse.
						</p>
					</div>
					{#if workspaceID}
						<MediaUpload
							workspaceId={workspaceID}
							bind:mediaIds={newMediaIds}
							disabled={creating}
						/>
					{/if}
					<div class="mt-auto flex items-center justify-between gap-3 pt-2">
						<span class="text-xs text-muted-foreground">{newMediaIds.length} attached</span>
						<Button class="gap-2" onclick={createPublication} disabled={!canCreate}>
							{#if creating}
								<LoaderIcon class="h-4 w-4 animate-spin" />
							{:else}
								<PlusIcon class="h-4 w-4" />
							{/if}
							Create
						</Button>
					</div>
				</div>
			</div>
		</section>

		<div class="flex flex-wrap items-center gap-2">
			{#each statusFilters as status (status)}
				<Button
					variant={selectedStatus === status ? 'default' : 'outline'}
					size="sm"
					onclick={() => (selectedStatus = status)}
				>
					{status === 'all' ? 'All' : statusLabel(status)}
				</Button>
			{/each}
		</div>

		<div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_390px]">
			<section class="overflow-hidden rounded-lg border">
				<div class="border-b bg-muted/30 px-4 py-3">
					<h2 class="text-sm font-semibold">Sources</h2>
				</div>
				{#if loading}
					<div class="divide-y">
						{#each Array(5) as _, index (index)}
							<div class="space-y-2 px-4 py-4">
								<div class="h-4 w-1/3 animate-pulse rounded bg-muted"></div>
								<div class="h-3 w-2/3 animate-pulse rounded bg-muted"></div>
							</div>
						{/each}
					</div>
				{:else if publications.length === 0}
					<div class="px-4 py-12 text-center">
						<FileTextIcon class="mx-auto h-8 w-8 text-muted-foreground" />
						<h3 class="mt-3 text-sm font-semibold">No publications yet</h3>
						<p class="mx-auto mt-1 max-w-sm text-sm text-muted-foreground">
							Create a source before drafting when you want posts, threads, and assistants to share
							the same context.
						</p>
					</div>
				{:else}
					<div class="divide-y">
						{#each publications as publication (publication.id)}
							<button
								type="button"
								class="grid w-full gap-2 px-4 py-3 text-left transition-colors hover:bg-muted/45 sm:grid-cols-[1fr_auto] {selectedPublication?.id ===
								publication.id
									? 'bg-muted/55'
									: ''}"
								onclick={() => (selectedPublicationId = publication.id)}
							>
								<div class="min-w-0">
									<div class="flex flex-wrap items-center gap-2">
										<h3 class="truncate text-sm font-medium">{publication.title}</h3>
										<Badge class={`${getStatusColor(publication.status)} capitalize`}>
											{publication.status}
										</Badge>
									</div>
									<p class="mt-1 line-clamp-2 text-sm text-muted-foreground">
										{preview(publication.source_content)}
									</p>
								</div>
								<div class="flex items-center gap-3 text-xs text-muted-foreground sm:justify-end">
									<span>{mediaCount(publication)} media</span>
									<span>{formatDate(publication.updated_at)}</span>
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</section>

			<aside class="rounded-lg border bg-card/55">
				{#if selectedPublication}
					<div class="border-b px-4 py-3">
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0">
								<h2 class="truncate text-sm font-semibold">{selectedPublication.title}</h2>
								<p class="mt-1 text-xs text-muted-foreground">
									Updated {formatDate(selectedPublication.updated_at)}
								</p>
							</div>
							<Badge class={`${getStatusColor(selectedPublication.status)} capitalize`}>
								{selectedPublication.status}
							</Badge>
						</div>
					</div>
					<div class="space-y-4 p-4">
						<div class="grid grid-cols-2 gap-3 text-xs">
							<div>
								<div class="font-medium text-muted-foreground">Goal</div>
								<div class="mt-1">{selectedPublication.goal || '-'}</div>
							</div>
							<div>
								<div class="font-medium text-muted-foreground">Audience</div>
								<div class="mt-1">{selectedPublication.audience || '-'}</div>
							</div>
							<div class="col-span-2">
								<div class="font-medium text-muted-foreground">Publication ID</div>
								<code class="mt-1 block truncate rounded bg-muted px-2 py-1 text-[11px]">
									{selectedPublication.id}
								</code>
							</div>
						</div>

						{#if selectedPublication.source_url}
							<a
								class="block truncate text-sm text-primary underline-offset-4 hover:underline"
								href={selectedPublication.source_url}
								target="_blank"
								rel="noreferrer"
							>
								{selectedPublication.source_url}
							</a>
						{/if}

						<div>
							<div class="mb-1 text-xs font-medium text-muted-foreground">Source content</div>
							<p
								class="max-h-64 overflow-auto rounded-md border bg-background/70 p-3 text-sm leading-relaxed whitespace-pre-wrap"
							>
								{selectedPublication.source_content}
							</p>
						</div>

						{#if mediaCount(selectedPublication) > 0}
							<div>
								<div class="mb-1 text-xs font-medium text-muted-foreground">Media IDs</div>
								<div class="flex flex-wrap gap-1.5">
									{#each selectedPublication.media_ids ?? [] as mediaID (mediaID)}
										<code class="rounded bg-muted px-2 py-1 text-[11px]">{mediaID}</code>
									{/each}
								</div>
							</div>
						{/if}

						<div class="grid gap-2">
							<Button
								class="gap-2"
								onclick={() => setPublicationStatus(selectedPublication, 'ready')}
								disabled={savingStatus !== '' || selectedPublication.status === 'ready'}
							>
								{#if savingStatus === 'ready'}
									<LoaderIcon class="h-4 w-4 animate-spin" />
								{:else}
									<CheckIcon class="h-4 w-4" />
								{/if}
								Mark ready
							</Button>
							<Button
								variant="outline"
								class="gap-2"
								onclick={() => setPublicationStatus(selectedPublication, 'draft')}
								disabled={savingStatus !== '' || selectedPublication.status === 'draft'}
							>
								{#if savingStatus === 'draft'}
									<LoaderIcon class="h-4 w-4 animate-spin" />
								{:else}
									<FileTextIcon class="h-4 w-4" />
								{/if}
								Back to draft
							</Button>
							<div class="grid grid-cols-2 gap-2">
								<Button
									variant="outline"
									class="gap-2"
									onclick={() => copyPublicationID(selectedPublication)}
								>
									<ClipboardIcon class="h-4 w-4" />
									Copy ID
								</Button>
								<Button
									variant="outline"
									class="gap-2"
									onclick={() => composeFromPublication(selectedPublication)}
								>
									<SendIcon class="h-4 w-4" />
									Compose
								</Button>
							</div>
						</div>
					</div>
				{:else}
					<div class="px-4 py-10 text-center">
						<FileTextIcon class="mx-auto h-8 w-8 text-muted-foreground" />
						<p class="mt-3 text-sm text-muted-foreground">Select a source to inspect it.</p>
					</div>
				{/if}
			</aside>
		</div>
	</div>
</PageContainer>
