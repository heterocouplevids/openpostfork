<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { client, type Workspace, type SocialAccount } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageContainer from '$lib/components/page-container.svelte';
	import EmptyState from '$lib/components/empty-state.svelte';
	import ChevronDownIcon from 'lucide-svelte/icons/chevron-down';
	import LayersIcon from 'lucide-svelte/icons/layers';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import TrashIcon from 'lucide-svelte/icons/trash-2';
	import { getPlatformName, getPlatformColor } from '$lib/utils';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import SettingsIcon from 'lucide-svelte/icons/settings';
	import UsersIcon from 'lucide-svelte/icons/users';
	import { Skeleton } from '$lib/components/ui/skeleton';

	interface SetAccount {
		social_account_id: string;
		platform: string;
		account_username: string;
		is_main: boolean;
	}

	interface SocialMediaSet {
		id: string;
		workspace_id: string;
		name: string;
		is_default: boolean;
		created_at: string;
		accounts: SetAccount[];
	}

	interface ProviderInfo {
		platform: string;
		display_name: string;
		auth_mode: string;
		configured: boolean;
		status?: string;
		description?: string;
		capabilities?: string[];
		name?: string;
		instance_url?: string;
	}

	let workspaces = $state<Workspace[] | null>(null);
	let selectedWorkspaceId = $state('');
	let loading = $state(true);
	let error = $state('');

	let accounts = $state<SocialAccount[]>([]);
	let accountsLoading = $state(false);

	let sets = $state<SocialMediaSet[]>([]);
	let setsLoading = $state(false);

	let providerEntries = $state.raw<ProviderInfo[]>([]);
	let providersLoading = $state(false);
	let customMastodonInstance = $state('');
	let customMastodonLoading = $state(false);
	let selectedWorkspaceName = $derived(
		workspaces?.find((workspace) => workspace.id === selectedWorkspaceId)?.name ||
			'Select workspace'
	);

	const fallbackProviderEntries: ProviderInfo[] = [
		{
			platform: 'bluesky',
			display_name: 'Bluesky',
			auth_mode: 'app_password',
			configured: true,
			status: 'available',
			description: 'Handle and app-password connection.'
		},
		{
			platform: 'x',
			display_name: 'X (Twitter)',
			auth_mode: 'oauth',
			configured: false,
			status: 'needs_configuration',
			description: 'Requires an X provider app.'
		},
		{
			platform: 'mastodon',
			display_name: 'Mastodon',
			auth_mode: 'oauth_oob',
			configured: false,
			status: 'needs_configuration',
			description: 'Configure Mastodon instances first.'
		},
		{
			platform: 'threads',
			display_name: 'Threads',
			auth_mode: 'oauth',
			configured: false,
			status: 'needs_configuration',
			description: 'Requires a Meta provider app.'
		},
		{
			platform: 'linkedin',
			display_name: 'LinkedIn',
			auth_mode: 'oauth',
			configured: false,
			status: 'needs_configuration',
			description: 'Requires a LinkedIn provider app.'
		},
		{
			platform: 'instagram',
			display_name: 'Instagram',
			auth_mode: 'oauth',
			configured: false,
			status: 'needs_configuration',
			description: 'Requires a Meta provider app.'
		},
		{
			platform: 'facebook',
			display_name: 'Facebook',
			auth_mode: 'oauth',
			configured: false,
			status: 'needs_configuration',
			description: 'Requires a Meta provider app.'
		},
		{
			platform: 'youtube',
			display_name: 'YouTube',
			auth_mode: 'oauth',
			configured: false,
			status: 'needs_configuration',
			description: 'Requires a Google OAuth provider app.'
		},
		{
			platform: 'tiktok',
			display_name: 'TikTok',
			auth_mode: 'oauth',
			configured: false,
			status: 'needs_configuration',
			description: 'Requires a TikTok provider app.'
		}
	];

	let blueskyModalOpen = $state(false);
	let blueskyHandle = $state('');
	let blueskyAppPassword = $state('');
	let blueskyLoading = $state(false);
	let blueskyError = $state('');

	let createSetDialogOpen = $state(false);
	let newSetName = $state('');
	let newSetDefault = $state(false);
	let newSetAccountIds = $state<string[]>([]);
	let createSetLoading = $state(false);

	let editSetDialogOpen = $state(false);
	let editingSet = $state<SocialMediaSet | null>(null);
	let editSetName = $state('');
	let editSetDefault = $state(false);
	let editSetAccountIds = $state<string[]>([]);
	let editSetLoading = $state(false);
	let editAccountDialogOpen = $state(false);
	let editingAccount = $state<SocialAccount | null>(null);
	let editAccountSlug = $state('');
	let editAccountLoading = $state(false);
	let editAccountError = $state('');
	const accountSlugPattern = '[a-z0-9][a-z0-9-]{0,62}';

	function resetCreateSetForm() {
		newSetName = '';
		newSetDefault = false;
		newSetAccountIds = [];
	}

	function closeEditSetDialog() {
		editSetDialogOpen = false;
		editingSet = null;
		editSetName = '';
		editSetDefault = false;
		editSetAccountIds = [];
	}

	function syncSetSelectionsWithAccounts(nextAccounts: SocialAccount[]) {
		const validIds = new Set(nextAccounts.map((account) => account.id));
		newSetAccountIds = newSetAccountIds.filter((id) => validIds.has(id));
		editSetAccountIds = editSetAccountIds.filter((id) => validIds.has(id));
	}

	async function loadAccounts() {
		if (!selectedWorkspaceId) return;
		accountsLoading = true;
		try {
			const { data, error: err } = await client.GET('/accounts', {
				params: { query: { workspace_id: selectedWorkspaceId } }
			});
			accounts = data ?? [];
			syncSetSelectionsWithAccounts(accounts);
		} catch (e) {
			console.error('Failed to load accounts:', e);
			accounts = [];
			syncSetSelectionsWithAccounts([]);
		} finally {
			accountsLoading = false;
		}
	}

	async function loadSets() {
		if (!selectedWorkspaceId) return;
		setsLoading = true;
		try {
			const { data, error: err } = await client.GET('/sets', {
				params: { query: { workspace_id: selectedWorkspaceId } }
			});
			sets = (data ?? []) as unknown as SocialMediaSet[];
			const currentEditingSet = editingSet;
			if (currentEditingSet) {
				const refreshedSet = sets.find((set) => set.id === currentEditingSet.id) ?? null;
				if (!refreshedSet) {
					closeEditSetDialog();
				} else {
					editingSet = refreshedSet;
					editSetAccountIds = refreshedSet.accounts.map((account) => account.social_account_id);
					editSetDefault = refreshedSet.is_default;
					editSetName = refreshedSet.name;
				}
			}
		} catch (e) {
			console.error('Failed to load sets:', e);
			sets = [];
		} finally {
			setsLoading = false;
		}
	}

	async function loadProviders() {
		providersLoading = true;
		try {
			const { data } = await (client as any).GET('/accounts/providers');
			providerEntries = (data ?? []) as ProviderInfo[];
		} catch (e) {
			console.error('Failed to load account providers:', e);
			providerEntries = fallbackProviderEntries;
		} finally {
			providersLoading = false;
		}
	}

	async function disconnectAccount(accountId: string) {
		try {
			await (client as any).DELETE('/accounts/{account_id}', {
				params: { path: { account_id: accountId } }
			});
			await loadAccounts();
			await loadSets();
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function createSet() {
		if (!newSetName.trim() || !selectedWorkspaceId) return;
		createSetLoading = true;
		try {
			await (client as any).POST('/sets', {
				body: {
					workspace_id: selectedWorkspaceId,
					name: newSetName.trim(),
					is_default: newSetDefault,
					account_ids: newSetAccountIds
				}
			});
			createSetDialogOpen = false;
			resetCreateSetForm();
			await loadSets();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			createSetLoading = false;
		}
	}

	async function updateSet() {
		if (!editingSet || !editSetName.trim()) return;
		editSetLoading = true;
		try {
			await (client as any).PATCH('/sets/{id}', {
				params: { path: { id: editingSet.id } },
				body: {
					name: editSetName.trim(),
					is_default: editSetDefault
				}
			});

			const currentAccIds = editingSet.accounts.map((a) => a.social_account_id);
			const toAdd = editSetAccountIds.filter((id) => !currentAccIds.includes(id));
			const toRemove = currentAccIds.filter((id) => !editSetAccountIds.includes(id));

			for (const accId of toAdd) {
				await (client as any).POST('/sets/{id}/accounts', {
					params: { path: { id: editingSet.id } },
					body: { account_ids: [accId] }
				});
			}

			for (const accId of toRemove) {
				await (client as any).DELETE('/sets/{id}/accounts/{account_id}', {
					params: { path: { id: editingSet.id, account_id: accId } }
				});
			}

			closeEditSetDialog();
			await loadSets();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			editSetLoading = false;
		}
	}

	async function deleteSet(setId: string) {
		if (!confirm('Delete this set? Posts using this set will keep their account selections.'))
			return;
		try {
			await (client as any).DELETE('/sets/{id}', {
				params: { path: { id: setId } }
			});
			await loadSets();
		} catch (e) {
			error = (e as Error).message;
		}
	}

	function openEditSet(set: SocialMediaSet) {
		editingSet = set;
		editSetName = set.name;
		editSetDefault = set.is_default;
		editSetAccountIds = set.accounts.map((a) => a.social_account_id);
		editSetDialogOpen = true;
	}

	function accountDisplayName(account: SocialAccount): string {
		if (account.account_username) return `@${account.account_username}`;
		if (account.instance_url) return account.instance_url.replace('https://', '');
		return account.account_id || account.platform;
	}

	function openEditAccount(account: SocialAccount) {
		editingAccount = account;
		editAccountSlug = (account as SocialAccount & { slug?: string }).slug ?? '';
		editAccountError = '';
		editAccountDialogOpen = true;
	}

	async function updateAccountSlug() {
		if (!editingAccount) return;
		editAccountLoading = true;
		editAccountError = '';
		try {
			const { error: err } = await (client as any).PATCH('/accounts/{account_id}', {
				params: { path: { account_id: editingAccount.id } },
				body: { slug: editAccountSlug.trim() }
			});
			if (err) throw new Error(err.detail || 'Failed to update account slug');
			editAccountDialogOpen = false;
			editingAccount = null;
			await loadAccounts();
		} catch (e) {
			editAccountError = (e as Error).message;
		} finally {
			editAccountLoading = false;
		}
	}

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		const urlError = params.get('error');
		if (urlError) {
			error = urlError;
			window.history.replaceState({}, document.title, window.location.pathname);
		}

		const unsubscribe = auth.subscribe(async (state) => {
			if (!state.isLoading && !state.isAuthenticated) {
				goto(resolve('/login'));
			} else if (!state.isLoading && state.isAuthenticated) {
				try {
					const { data, error: err } = await client.GET('/workspaces');
					workspaces = data ?? [];
					if (workspaces && workspaces.length > 0) {
						selectedWorkspaceId = workspaces[0].id;
						await loadAccounts();
						await loadSets();
					}
					await loadProviders();
				} catch (e) {
					console.error('Failed to load workspaces:', e);
				} finally {
					loading = false;
				}
			}
		});
		return unsubscribe;
	});

	$effect(() => {
		if (selectedWorkspaceId) {
			loadAccounts();
			loadSets();
		} else {
			accounts = [];
			sets = [];
			resetCreateSetForm();
			closeEditSetDialog();
		}
	});

	async function connectTwitter() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}
		try {
			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: { path: { platform: 'x' }, query: { workspace_id: selectedWorkspaceId } }
			});
			if (err) throw new Error((err as any).detail || 'Failed to get X auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function connectMastodon(options: { serverName?: string; instanceURL?: string }) {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);
			if (options.instanceURL) {
				localStorage.setItem('oauth_mastodon_instance_url', options.instanceURL);
				localStorage.removeItem('oauth_mastodon_server');
			} else if (options.serverName) {
				localStorage.setItem('oauth_mastodon_server', options.serverName);
				localStorage.removeItem('oauth_mastodon_instance_url');
			}

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'mastodon' },
					query: {
						workspace_id: selectedWorkspaceId,
						server_name: options.serverName,
						instance_url: options.instanceURL
					}
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get Mastodon auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function connectCustomMastodon() {
		const instanceURL = customMastodonInstance.trim();
		if (!instanceURL) {
			error = 'Enter a Mastodon instance URL';
			return;
		}
		customMastodonLoading = true;
		try {
			await connectMastodon({ instanceURL });
		} finally {
			customMastodonLoading = false;
		}
	}

	async function connectBluesky() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}
		blueskyHandle = '';
		blueskyAppPassword = '';
		blueskyError = '';
		blueskyModalOpen = true;
	}

	async function submitBlueskyLogin() {
		if (!blueskyHandle.trim() || !blueskyAppPassword.trim()) {
			blueskyError = 'Please enter both handle and app password';
			return;
		}

		blueskyLoading = true;
		blueskyError = '';

		try {
			const { error: err } = await (client as any).POST('/accounts/bluesky/login', {
				body: {
					workspace_id: selectedWorkspaceId,
					handle: blueskyHandle.trim(),
					app_password: blueskyAppPassword.trim()
				}
			});
			if (err) throw new Error(err.detail || 'Login failed');
			blueskyModalOpen = false;
			await loadAccounts();
			await loadSets();
		} catch (e) {
			blueskyError = (e as Error).message;
		} finally {
			blueskyLoading = false;
		}
	}

	async function connectLinkedIn() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'linkedin' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function connectThreads() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'threads' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function connectTikTok() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'tiktok' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function connectFacebook() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'facebook' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function connectInstagram() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'instagram' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	async function connectYouTube() {
		if (!selectedWorkspaceId) {
			alert('Please create a workspace first');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'youtube' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			error = (e as Error).message;
		}
	}

	function providerKey(provider: ProviderInfo): string {
		if (provider.platform === 'mastodon') {
			return provider.instance_url || provider.name || provider.platform;
		}
		return provider.platform;
	}

	function providerTitle(provider: ProviderInfo): string {
		if (provider.platform === 'mastodon' && provider.name) return provider.name;
		return provider.display_name || getPlatformName(provider.platform);
	}

	function providerDescription(provider: ProviderInfo): string {
		if (provider.description) return provider.description;
		if (!provider.configured) return 'Not configured';
		if (isCustomMastodonProvider(provider)) {
			return 'Connect any public Mastodon instance';
		}
		if (provider.platform === 'mastodon' && provider.instance_url) {
			return provider.instance_url.replace('https://', '');
		}
		switch (provider.platform) {
			case 'x':
				return 'Post tweets';
			case 'threads':
				return 'Post to Threads';
			case 'bluesky':
				return 'Post to Bluesky';
			case 'linkedin':
				return 'Post to LinkedIn';
			case 'instagram':
				return 'Post to Instagram Business';
			case 'facebook':
				return 'Post to Facebook Pages';
			case 'youtube':
				return 'Upload YouTube videos and Shorts';
			case 'tiktok':
				return 'Post short-form video to TikTok';
			default:
				return 'Connect account';
		}
	}

	function providerStatus(provider: ProviderInfo): string {
		if (provider.status) return provider.status;
		return provider.configured ? 'available' : 'needs_configuration';
	}

	function providerStatusLabel(provider: ProviderInfo): string {
		switch (providerStatus(provider)) {
			case 'available':
				return 'Available';
			case 'planned':
				return 'Planned';
			case 'needs_configuration':
				return 'Needs app config';
			default:
				return provider.configured ? 'Available' : 'Unavailable';
		}
	}

	function providerStatusClass(provider: ProviderInfo): string {
		switch (providerStatus(provider)) {
			case 'available':
				return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300';
			case 'planned':
				return 'border-blue-500/20 bg-blue-500/10 text-blue-700 dark:text-blue-300';
			case 'needs_configuration':
				return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300';
			default:
				return 'border-muted bg-muted text-muted-foreground';
		}
	}

	function providerCanConnect(provider: ProviderInfo): boolean {
		return provider.configured && providerStatus(provider) !== 'planned';
	}

	function providerActionLabel(provider: ProviderInfo): string {
		if (providerStatus(provider) === 'planned') return 'Planned';
		return provider.configured ? 'Connect' : 'Unavailable';
	}

	function visibleProviderCapabilities(provider: ProviderInfo): string[] {
		return (provider.capabilities ?? []).slice(0, 4);
	}

	function isCustomMastodonProvider(provider: ProviderInfo): boolean {
		return provider.platform === 'mastodon' && provider.configured && !provider.instance_url;
	}

	function rememberMastodonProvider(provider: ProviderInfo) {
		if (!selectedWorkspaceId) return;
		localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);
		if (isCustomMastodonProvider(provider)) {
			const instanceURL = customMastodonInstance.trim();
			if (instanceURL) {
				localStorage.setItem('oauth_mastodon_instance_url', instanceURL);
				localStorage.removeItem('oauth_mastodon_server');
			}
			return;
		}
		const serverName = provider.name || provider.instance_url || '';
		if (serverName) {
			localStorage.setItem('oauth_mastodon_server', serverName);
			localStorage.removeItem('oauth_mastodon_instance_url');
		}
	}

	function connectProvider(provider: ProviderInfo) {
		if (!providerCanConnect(provider)) return;
		switch (provider.platform) {
			case 'x':
				connectTwitter();
				break;
			case 'mastodon':
				if (isCustomMastodonProvider(provider)) {
					connectCustomMastodon();
				} else {
					connectMastodon({ serverName: provider.name || provider.instance_url || '' });
				}
				break;
			case 'threads':
				connectThreads();
				break;
			case 'bluesky':
				connectBluesky();
				break;
			case 'linkedin':
				connectLinkedIn();
				break;
			case 'instagram':
				connectInstagram();
				break;
			case 'facebook':
				connectFacebook();
				break;
			case 'youtube':
				connectYouTube();
				break;
			case 'tiktok':
				connectTikTok();
				break;
		}
	}

	function toggleNewSetAccount(accId: string) {
		if (newSetAccountIds.includes(accId)) {
			newSetAccountIds = newSetAccountIds.filter((id) => id !== accId);
		} else {
			newSetAccountIds = [...newSetAccountIds, accId];
		}
	}

	function toggleEditSetAccount(accId: string) {
		if (editSetAccountIds.includes(accId)) {
			editSetAccountIds = editSetAccountIds.filter((id) => id !== accId);
		} else {
			editSetAccountIds = [...editSetAccountIds, accId];
		}
	}

	const accountsByPlatform = $derived.by(() => {
		const grouped = new Map<string, SocialAccount[]>();
		for (const acc of accounts) {
			const key = acc.platform;
			if (!grouped.has(key)) grouped.set(key, []);
			grouped.get(key)!.push(acc);
		}
		return grouped;
	});
</script>

<svelte:head>
	<title>Connected Accounts - OpenPost</title>
</svelte:head>

{#if loading}
	<div class="mx-auto w-full max-w-6xl px-4 py-6 lg:px-8">
		<div class="mb-6 flex items-center gap-2">
			<Skeleton class="h-8 w-8 rounded-md" />
			<Skeleton class="h-7 w-48" />
		</div>
		<div class="mb-6"><Skeleton class="h-9 w-40 rounded-md" /></div>
		<div class="mb-8 space-y-3">
			<Skeleton class="h-6 w-32" />
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<Skeleton class="h-28 rounded-lg" />
				<Skeleton class="h-28 rounded-lg" />
			</div>
		</div>
		<div class="space-y-3">
			<Skeleton class="h-6 w-40" />
			<Skeleton class="h-40 rounded-lg" />
		</div>
	</div>
{:else if !workspaces || workspaces.length === 0}
	<div class="mx-auto max-w-md px-4 py-16 text-center">
		<h2 class="mb-2 text-xl font-semibold">No Workspaces Found</h2>
		<p class="mb-4 text-sm text-muted-foreground">
			Create a workspace first before connecting social accounts.
		</p>
		<Button href="/">Create Workspace</Button>
	</div>
{:else}
	<PageContainer
		title="Accounts"
		description="Connect and manage your social accounts and sets."
		icon={UsersIcon}
	>
		{#if error}
			<div
				class="mb-4 flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
			>
				{error}
				<Button variant="ghost" size="sm" onclick={() => (error = '')}>Dismiss</Button>
			</div>
		{/if}

		<!-- Workspace Selector -->
		<div class="mb-6">
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button {...props} variant="outline" class="gap-2">
							<span
								class="flex h-6 w-6 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary"
							>
								{selectedWorkspaceName.slice(0, 2).toUpperCase()}
							</span>
							<span class="truncate">{selectedWorkspaceName}</span>
							<ChevronDownIcon class="size-3.5 opacity-50" />
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content class="w-56 rounded-lg" align="start" side="bottom" sideOffset={6}>
					<DropdownMenu.Label class="text-xs text-muted-foreground">Workspaces</DropdownMenu.Label>
					{#each workspaces as workspace (workspace.id)}
						<DropdownMenu.Item
							onSelect={() => {
								selectedWorkspaceId = workspace.id;
								resetCreateSetForm();
								closeEditSetDialog();
							}}
							class="gap-2 p-2"
						>
							<span
								class="flex size-6 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary"
							>
								{workspace.name.slice(0, 2).toUpperCase()}
							</span>
							<span class="truncate">{workspace.name}</span>
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>

		<!-- Social Media Sets -->
		<div class="mb-8">
			<div class="mb-4 flex items-center justify-between">
				<h2 class="text-lg font-semibold">Social Media Sets</h2>
				<Button onclick={() => (createSetDialogOpen = true)} size="sm" class="gap-1.5">
					<PlusIcon class="h-3.5 w-3.5" />
					New Set
				</Button>
			</div>

			{#if setsLoading}
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
					<Skeleton class="h-28 rounded-lg" />
					<Skeleton class="h-28 rounded-lg" />
				</div>
			{:else if sets.length === 0}
				<EmptyState
					icon={LayersIcon}
					title="No sets yet"
					description="Group accounts together for quick posting"
					actionLabel="Create your first set"
					onAction={() => (createSetDialogOpen = true)}
					variant="muted"
					size="md"
				/>
			{:else}
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
					{#each sets as set (set.id)}
						<div class="group rounded-lg border bg-card p-4 transition-all hover:shadow-sm">
							<div class="mb-3 flex items-start justify-between">
								<div class="flex items-center gap-3">
									<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10">
										<LayersIcon class="h-5 w-5 text-primary" />
									</div>
									<div>
										<h3 class="text-sm font-medium">{set.name}</h3>
										<p class="text-sm text-muted-foreground">
											{set.accounts.length} account{set.accounts.length !== 1 ? 's' : ''}
										</p>
									</div>
								</div>
								<div
									class="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100"
								>
									{#if set.is_default}
										<span
											class="mr-1 rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary"
										>
											Default
										</span>
									{/if}
									<Button
										variant="ghost"
										size="icon"
										class="h-7 w-7"
										onclick={() => openEditSet(set)}
									>
										<SettingsIcon class="h-3.5 w-3.5" />
									</Button>
									<Button
										variant="ghost"
										size="icon"
										class="h-7 w-7 text-destructive hover:text-destructive"
										onclick={() => deleteSet(set.id)}
									>
										<TrashIcon class="h-3.5 w-3.5" />
									</Button>
								</div>
							</div>
							{#if set.accounts.length > 0}
								<div class="flex flex-wrap gap-1.5">
									{#each set.accounts as acc (acc.social_account_id)}
										<span
											class="inline-flex items-center gap-1.5 rounded-full bg-muted px-2.5 py-1 text-sm font-medium"
										>
											<PlatformIcon platform={acc.platform} class="h-3 w-3" />
											{acc.account_username || acc.platform}
										</span>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-muted-foreground">No accounts in this set</p>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Connected Accounts -->
		<div class="mb-8">
			<h2 class="mb-4 text-lg font-semibold">Connected Accounts</h2>

			{#if accountsLoading}
				<div class="space-y-3">
					<Skeleton class="h-12 rounded-lg" />
					<Skeleton class="h-12 rounded-lg" />
					<Skeleton class="h-12 rounded-lg" />
				</div>
			{:else if !accounts || accounts.length === 0}
				<EmptyState
					icon={UsersIcon}
					title="No accounts connected"
					description="Connect a platform below to get started"
					variant="muted"
					size="md"
				/>
			{:else}
				<div class="space-y-3">
					{#each [...accountsByPlatform.entries()] as [platform, platformAccounts] (platform)}
						<div class="rounded-lg border bg-card">
							<div class="flex items-center gap-3 border-b px-4 py-3">
								<div
									class="flex h-9 w-9 items-center justify-center rounded-full {getPlatformColor(
										platform
									)}"
								>
									<PlatformIcon {platform} class="h-4 w-4 text-white" />
								</div>
								<div class="flex-1">
									<h3 class="text-sm font-medium">{getPlatformName(platform)}</h3>
									<p class="text-sm text-muted-foreground">
										{platformAccounts.length} account{platformAccounts.length !== 1 ? 's' : ''}
									</p>
								</div>
							</div>
							<div class="divide-y">
								{#each platformAccounts as account (account.id)}
									<div class="flex items-center justify-between px-4 py-3">
										<div class="flex items-center gap-3">
											<div class="flex h-7 w-7 items-center justify-center rounded-full bg-muted">
												<PlatformIcon platform={account.platform} class="h-3.5 w-3.5" />
											</div>
											<div>
												<p class="text-sm font-medium">
													{accountDisplayName(account)}
												</p>
												<p class="text-sm text-muted-foreground">
													Slug:
													<span class="font-mono"
														>{(account as SocialAccount & { slug?: string }).slug ||
															'not set'}</span
													>
													· {account.is_active ? 'Connected' : 'Disconnected'}
												</p>
											</div>
										</div>
										<div class="flex items-center gap-2">
											<Button
												variant="outline"
												size="sm"
												onclick={() => openEditAccount(account)}
												class="text-xs"
											>
												Edit Slug
											</Button>
											<Button
												variant="ghost"
												size="sm"
												onclick={() => disconnectAccount(account.id)}
												class="text-xs text-muted-foreground hover:text-destructive"
											>
												Disconnect
											</Button>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Connect a Platform -->
		<div>
			<h2 class="mb-4 text-lg font-semibold">Connect a Platform</h2>

			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				{#if providersLoading}
					<Skeleton class="h-20 rounded-lg" />
					<Skeleton class="h-20 rounded-lg" />
					<Skeleton class="h-20 rounded-lg" />
					<Skeleton class="h-20 rounded-lg" />
				{:else}
					{#each providerEntries as provider (providerKey(provider))}
						<div
							data-testid={`provider-card-${provider.platform}`}
							class="group rounded-lg border bg-card p-4 transition-all hover:shadow-sm {providerCanConnect(
								provider
							)
								? ''
								: 'bg-muted/20'}"
						>
							<div class="flex items-center gap-3">
								<div
									class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full {getPlatformColor(
										provider.platform
									)}"
								>
									<PlatformIcon platform={provider.platform} class="h-4 w-4 text-white" />
								</div>
								<div class="min-w-0 flex-1">
									<div class="flex flex-wrap items-center gap-2">
										<h3 class="text-sm font-medium">{providerTitle(provider)}</h3>
										<span
											class="inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium {providerStatusClass(
												provider
											)}"
										>
											{providerStatusLabel(provider)}
										</span>
									</div>
									<p class="truncate text-sm text-muted-foreground">
										{providerDescription(provider)}
									</p>
								</div>
								{#if provider.platform === 'mastodon' && provider.configured && !isCustomMastodonProvider(provider)}
									<div class="flex gap-1.5">
										<Button
											href="/accounts/mastodon/callback"
											variant="outline"
											size="sm"
											class="text-xs"
											onclick={() => rememberMastodonProvider(provider)}>Code</Button
										>
										<Button onclick={() => connectProvider(provider)} size="sm">Connect</Button>
									</div>
								{:else if !isCustomMastodonProvider(provider)}
									<Button
										onclick={() => connectProvider(provider)}
										size="sm"
										disabled={!providerCanConnect(provider)}
									>
										{providerActionLabel(provider)}
									</Button>
								{/if}
							</div>
							{#if visibleProviderCapabilities(provider).length > 0}
								<div class="mt-3 flex flex-wrap gap-1.5">
									{#each visibleProviderCapabilities(provider) as capability (capability)}
										<span
											class="rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground"
										>
											{capability}
										</span>
									{/each}
								</div>
							{/if}
							{#if isCustomMastodonProvider(provider)}
								<form
									class="mt-4 grid gap-2 sm:grid-cols-[1fr_auto_auto]"
									onsubmit={(e: SubmitEvent) => {
										e.preventDefault();
										connectCustomMastodon();
									}}
								>
									<div class="space-y-1.5">
										<Label for="custom-mastodon-instance" class="text-xs">Instance URL</Label>
										<Input
											id="custom-mastodon-instance"
											bind:value={customMastodonInstance}
											placeholder="mastodon.social"
											autocomplete="url"
										/>
									</div>
									<Button
										href="/accounts/mastodon/callback"
										variant="outline"
										size="sm"
										class="self-end"
										onclick={() => rememberMastodonProvider(provider)}>Code</Button
									>
									<Button type="submit" size="sm" class="self-end" disabled={customMastodonLoading}>
										{customMastodonLoading ? 'Connecting...' : 'Connect'}
									</Button>
								</form>
							{/if}
						</div>
					{/each}
				{/if}
			</div>
		</div>
	</PageContainer>
{/if}

<Dialog.Root bind:open={blueskyModalOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Connect Bluesky</Dialog.Title>
			<Dialog.Description>
				Enter your Bluesky handle and an app password. You can create an app password in Bluesky
				Settings &gt; App Passwords.
			</Dialog.Description>
		</Dialog.Header>
		<form
			class="space-y-4"
			onsubmit={(e) => {
				e.preventDefault();
				submitBlueskyLogin();
			}}
		>
			<div class="space-y-2">
				<Label for="bluesky-handle">Handle</Label>
				<Input
					type="text"
					id="bluesky-handle"
					bind:value={blueskyHandle}
					placeholder="user.bsky.social"
					required
				/>
			</div>
			<div class="space-y-2">
				<Label for="bluesky-password">App Password</Label>
				<Input
					type="password"
					id="bluesky-password"
					bind:value={blueskyAppPassword}
					placeholder="xxxx-xxxx-xxxx-xxxx"
					required
				/>
			</div>
			{#if blueskyError}
				<div
					class="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
				>
					{blueskyError}
				</div>
			{/if}
			<div class="flex justify-end gap-2">
				<Dialog.Close>
					<Button variant="outline" type="button">Cancel</Button>
				</Dialog.Close>
				<Button type="submit" disabled={blueskyLoading}>
					{blueskyLoading ? 'Connecting...' : 'Connect'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={editAccountDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Edit Account Slug</Dialog.Title>
			<Dialog.Description>
				Slugs are stable shortcuts for the CLI, for example
				<code class="rounded bg-muted px-1 py-0.5"
					>openpost post create --accounts {editAccountSlug || 'main-x'}</code
				>.
			</Dialog.Description>
		</Dialog.Header>
		{#if editingAccount}
			<form
				class="space-y-4"
				onsubmit={(e) => {
					e.preventDefault();
					updateAccountSlug();
				}}
			>
				<div class="rounded-md border bg-muted/30 p-3 text-sm">
					<div class="font-medium">{accountDisplayName(editingAccount)}</div>
					<div class="text-muted-foreground">{getPlatformName(editingAccount.platform)}</div>
				</div>
				<div class="space-y-2">
					<Label for="account-slug">Slug</Label>
					<Input
						id="account-slug"
						bind:value={editAccountSlug}
						placeholder="main-x"
						pattern={accountSlugPattern}
						required
					/>
					<p class="text-xs text-muted-foreground">
						Use lowercase letters, numbers, and hyphens. Slugs must be unique within this workspace.
					</p>
				</div>
				{#if editAccountError}
					<div
						class="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
					>
						{editAccountError}
					</div>
				{/if}
				<div class="flex justify-end gap-2">
					<Dialog.Close>
						<Button variant="outline" type="button">Cancel</Button>
					</Dialog.Close>
					<Button type="submit" disabled={editAccountLoading || !editAccountSlug.trim()}>
						{editAccountLoading ? 'Saving...' : 'Save Slug'}
					</Button>
				</div>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={createSetDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Create Social Media Set</Dialog.Title>
			<Dialog.Description>
				Group accounts together for quick posting. Sets appear in the compose post dropdown.
			</Dialog.Description>
		</Dialog.Header>
		<form
			class="space-y-4"
			onsubmit={(e) => {
				e.preventDefault();
				createSet();
			}}
		>
			<div class="space-y-2">
				<Label for="set-name">Set Name</Label>
				<Input
					id="set-name"
					bind:value={newSetName}
					placeholder="e.g. Tech Twitter, Professional"
					required
				/>
			</div>
			<div class="flex items-center gap-2">
				<Checkbox id="set-default" bind:checked={newSetDefault} />
				<Label for="set-default" class="text-sm font-normal">
					Set as the default set for this workspace
				</Label>
			</div>
			{#if accounts.length > 0}
				<div class="space-y-2">
					<Label>Accounts to Include</Label>
					<div class="max-h-48 space-y-2 overflow-y-auto rounded-md border p-2">
						{#each accounts as account (account.id)}
							<label class="flex items-center gap-2 rounded p-2 hover:bg-muted/50">
								<Checkbox
									checked={newSetAccountIds.includes(account.id)}
									onCheckedChange={() => toggleNewSetAccount(account.id)}
								/>
								<PlatformIcon platform={account.platform} class="h-4 w-4" />
								<span class="text-sm">
									{#if account.account_username}
										@{account.account_username}
									{:else if account.instance_url}
										{account.instance_url.replace('https://', '')}
									{:else}
										{account.platform}
									{/if}
								</span>
							</label>
						{/each}
					</div>
				</div>
			{/if}
			<div class="flex justify-end gap-2">
				<Dialog.Close>
					<Button variant="outline" type="button" onclick={resetCreateSetForm}>Cancel</Button>
				</Dialog.Close>
				<Button type="submit" disabled={createSetLoading || !newSetName.trim()}>
					{createSetLoading ? 'Creating...' : 'Create Set'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={editSetDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Edit Social Media Set</Dialog.Title>
			<Dialog.Description>Update the set name, default status, and accounts.</Dialog.Description>
		</Dialog.Header>
		{#if editingSet}
			<form
				class="space-y-4"
				onsubmit={(e) => {
					e.preventDefault();
					updateSet();
				}}
			>
				<div class="space-y-2">
					<Label for="edit-set-name">Set Name</Label>
					<Input
						id="edit-set-name"
						bind:value={editSetName}
						placeholder="e.g. Tech Twitter, Professional"
						required
					/>
				</div>
				<div class="flex items-center gap-2">
					<Checkbox id="edit-set-default" bind:checked={editSetDefault} />
					<Label for="edit-set-default" class="text-sm font-normal">
						Set as the default set for this workspace
					</Label>
				</div>
				{#if accounts.length > 0}
					<div class="space-y-2">
						<Label>Accounts in Set</Label>
						<div class="max-h-48 space-y-2 overflow-y-auto rounded-md border p-2">
							{#each accounts as account (account.id)}
								<label class="flex items-center gap-2 rounded p-2 hover:bg-muted/50">
									<Checkbox
										checked={editSetAccountIds.includes(account.id)}
										onCheckedChange={() => toggleEditSetAccount(account.id)}
									/>
									<PlatformIcon platform={account.platform} class="h-4 w-4" />
									<span class="text-sm">
										{#if account.account_username}
											@{account.account_username}
										{:else if account.instance_url}
											{account.instance_url.replace('https://', '')}
										{:else}
											{account.platform}
										{/if}
									</span>
								</label>
							{/each}
						</div>
					</div>
				{/if}
				<div class="flex justify-end gap-2">
					<Dialog.Close>
						<Button variant="outline" type="button" onclick={closeEditSetDialog}>Cancel</Button>
					</Dialog.Close>
					<Button type="submit" disabled={editSetLoading || !editSetName.trim()}>
						{editSetLoading ? 'Saving...' : 'Save Changes'}
					</Button>
				</div>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>
