<script lang="ts">
	import { page } from '$app/state';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Select from '$lib/components/ui/select';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import PageContainer from '$lib/components/page-container.svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { createPasskeyCredential } from '$lib/auth/webauthn';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import SettingsIcon from 'lucide-svelte/icons/settings';
	import SaveIcon from 'lucide-svelte/icons/save';
	import XIcon from 'lucide-svelte/icons/x';
	import ClockIcon from 'lucide-svelte/icons/clock';
	import ImageIcon from 'lucide-svelte/icons/image';
	import CalendarIcon from 'lucide-svelte/icons/calendar';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import TrashIcon from 'lucide-svelte/icons/trash';
	import SparklesIcon from 'lucide-svelte/icons/sparkles';
	import ShieldCheckIcon from 'lucide-svelte/icons/shield-check';
	import SmartphoneIcon from 'lucide-svelte/icons/smartphone';
	import KeyRoundIcon from 'lucide-svelte/icons/key-round';
	import TerminalIcon from 'lucide-svelte/icons/terminal';
	import CreditCardIcon from 'lucide-svelte/icons/credit-card';
	import ExternalLinkIcon from 'lucide-svelte/icons/external-link';
	import ActivityIcon from 'lucide-svelte/icons/activity';
	import UsersIcon from 'lucide-svelte/icons/users';
	import UserPlusIcon from 'lucide-svelte/icons/user-plus';
	import CopyIcon from 'lucide-svelte/icons/copy';
	import MonitorIcon from 'lucide-svelte/icons/monitor';
	import LogOutIcon from 'lucide-svelte/icons/log-out';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { client } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { getLocaleTag } from '$lib/i18n';
	import { hostedPlanFromSearchParams } from '$lib/billing';

	const timezones = [
		{ group: 'Americas', value: 'America/New_York', label: 'New York (ET)' },
		{ group: 'Americas', value: 'America/Chicago', label: 'Chicago (CT)' },
		{ group: 'Americas', value: 'America/Denver', label: 'Denver (MT)' },
		{ group: 'Americas', value: 'America/Los_Angeles', label: 'Los Angeles (PT)' },
		{ group: 'Americas', value: 'America/Phoenix', label: 'Phoenix (AZ)' },
		{ group: 'Americas', value: 'America/Anchorage', label: 'Anchorage (AK)' },
		{ group: 'Americas', value: 'Pacific/Honolulu', label: 'Honolulu (HI)' },
		{ group: 'Americas', value: 'America/Toronto', label: 'Toronto (ET)' },
		{ group: 'Americas', value: 'America/Vancouver', label: 'Vancouver (PT)' },
		{ group: 'Americas', value: 'America/Mexico_City', label: 'Mexico City (CT)' },
		{ group: 'Americas', value: 'America/Bogota', label: 'Bogota' },
		{ group: 'Americas', value: 'America/Lima', label: 'Lima' },
		{ group: 'Americas', value: 'America/Santiago', label: 'Santiago' },
		{ group: 'Americas', value: 'America/Sao_Paulo', label: 'Sao Paulo' },
		{ group: 'Americas', value: 'America/Buenos_Aires', label: 'Buenos Aires' },
		{ group: 'Europe', value: 'UTC', label: 'UTC' },
		{ group: 'Europe', value: 'Europe/London', label: 'London (GMT/BST)' },
		{ group: 'Europe', value: 'Europe/Dublin', label: 'Dublin (GMT/IST)' },
		{ group: 'Europe', value: 'Europe/Lisbon', label: 'Lisbon (WET/WEST)' },
		{ group: 'Europe', value: 'Europe/Madrid', label: 'Madrid (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Paris', label: 'Paris (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Amsterdam', label: 'Amsterdam (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Brussels', label: 'Brussels (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Berlin', label: 'Berlin (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Vienna', label: 'Vienna (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Zurich', label: 'Zurich (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Rome', label: 'Rome (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Stockholm', label: 'Stockholm (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Oslo', label: 'Oslo (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Copenhagen', label: 'Copenhagen (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Helsinki', label: 'Helsinki (EET/EEST)' },
		{ group: 'Europe', value: 'Europe/Warsaw', label: 'Warsaw (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Prague', label: 'Prague (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Budapest', label: 'Budapest (CET/CEST)' },
		{ group: 'Europe', value: 'Europe/Athens', label: 'Athens (EET/EEST)' },
		{ group: 'Europe', value: 'Europe/Bucharest', label: 'Bucharest (EET/EEST)' },
		{ group: 'Europe', value: 'Europe/Kiev', label: 'Kiev (EET/EEST)' },
		{ group: 'Europe', value: 'Europe/Moscow', label: 'Moscow (MSK)' },
		{ group: 'Europe', value: 'Europe/Istanbul', label: 'Istanbul (TRT)' },
		{ group: 'Asia', value: 'Asia/Dubai', label: 'Dubai (GST)' },
		{ group: 'Asia', value: 'Asia/Riyadh', label: 'Riyadh (AST)' },
		{ group: 'Asia', value: 'Asia/Tehran', label: 'Tehran (IRST/IRDT)' },
		{ group: 'Asia', value: 'Asia/Kolkata', label: 'Mumbai/Delhi (IST)' },
		{ group: 'Asia', value: 'Asia/Bangkok', label: 'Bangkok (ICT)' },
		{ group: 'Asia', value: 'Asia/Jakarta', label: 'Jakarta (WIB)' },
		{ group: 'Asia', value: 'Asia/Singapore', label: 'Singapore (SGT)' },
		{ group: 'Asia', value: 'Asia/Hong_Kong', label: 'Hong Kong (HKT)' },
		{ group: 'Asia', value: 'Asia/Shanghai', label: 'Shanghai (CST)' },
		{ group: 'Asia', value: 'Asia/Tokyo', label: 'Tokyo (JST)' },
		{ group: 'Asia', value: 'Asia/Seoul', label: 'Seoul (KST)' },
		{ group: 'Asia', value: 'Asia/Manila', label: 'Manila (PHT)' },
		{ group: 'Asia', value: 'Asia/Kuala_Lumpur', label: 'Kuala Lumpur (MYT)' },
		{ group: 'Pacific', value: 'Australia/Perth', label: 'Perth (AWST)' },
		{ group: 'Pacific', value: 'Australia/Eucla', label: 'Eucla (AWST+)' },
		{ group: 'Pacific', value: 'Australia/Adelaide', label: 'Adelaide (ACST)' },
		{ group: 'Pacific', value: 'Australia/Brisbane', label: 'Brisbane (AEST)' },
		{ group: 'Pacific', value: 'Australia/Sydney', label: 'Sydney (AEST/AEDT)' },
		{ group: 'Pacific', value: 'Pacific/Auckland', label: 'Auckland (NZST/NZDT)' },
		{ group: 'Pacific', value: 'Pacific/Fiji', label: 'Fiji (FJT/FJST)' },
		{ group: 'Africa', value: 'Africa/Cairo', label: 'Cairo (EET)' },
		{ group: 'Africa', value: 'Africa/Johannesburg', label: 'Johannesburg (SAST)' },
		{ group: 'Africa', value: 'Africa/Lagos', label: 'Lagos (WAT)' },
		{ group: 'Africa', value: 'Africa/Nairobi', label: 'Nairobi (EAT)' }
	];

	const groupedTimezones = $derived.by(() => {
		const groups: Record<string, typeof timezones> = {};
		for (const tz of timezones) {
			if (!groups[tz.group]) groups[tz.group] = [];
			groups[tz.group].push(tz);
		}
		return groups;
	});

	function getTimezoneLabel(value: string): string {
		const tz = timezones.find((t) => t.value === value);
		return tz?.label ?? value;
	}

	const cleanupDaysOptions = [
		{ value: 0, label: 'Disabled' },
		{ value: 7, label: '7 days' },
		{ value: 14, label: '14 days' },
		{ value: 30, label: '30 days' },
		{ value: 60, label: '60 days' },
		{ value: 90, label: '90 days' },
		{ value: 180, label: '180 days' },
		{ value: 365, label: '1 year' }
	];

	let saving = $state(false);
	let toastMessage = $state('');
	let loadingSecurity = $state(true);
	let securityBusy = $state(false);
	let securityError = $state('');
	let authSessions = $state.raw<AuthSessionSummary[]>([]);
	let authSessionsLoading = $state(true);
	let authSessionsError = $state('');
	let authSessionBusyID = $state('');
	let currentPassword = $state('');
	let totpSetupChallengeId = $state('');
	let totpManualEntryKey = $state('');
	let totpQRCodeDataURL = $state('');
	let totpCode = $state('');
	let newPasskeyName = $state('');

	interface PasskeySummary {
		id: string;
		name: string;
		created_at: string;
		last_used_at: string;
	}

	interface SecurityStatus {
		user: {
			id: string;
			email: string;
			created_at: string;
		};
		totp_enabled: boolean;
		passkeys: PasskeySummary[];
		methods: string[];
	}

	interface AuthSessionSummary {
		id: string;
		user_agent: string;
		ip_address: string;
		current: boolean;
		expires_at: string;
		last_used_at: string;
		created_at: string;
	}

	let securityStatus = $state<SecurityStatus | null>(null);
	let apiTokens = $state<APITokenSummary[]>([]);
	let apiTokensLoading = $state(true);
	let mcpActivity = $state.raw<MCPActivityItem[]>([]);
	let mcpActivityLoading = $state(true);
	let mcpActivityError = $state('');
	let apiTokenBusy = $state(false);
	let apiTokenName = $state('OpenPost MCP');
	let apiTokenScope = $state('mcp:full');
	let apiTokenWorkspaceScope = $state('current');
	let createdAPIToken = $state('');
	let billingBusyPlan = $state('');
	let billingPortalBusy = $state(false);
	let billingError = $state('');
	let billingStatusLoading = $state(false);
	let billingStatus = $state<BillingStatus | null>(null);
	let handledCheckoutPlan = '';
	let teamLoading = $state(false);
	let teamBusy = $state(false);
	let teamError = $state('');
	let workspaceTeam = $state<WorkspaceTeam | null>(null);
	let inviteEmail = $state('');
	let inviteRole = $state('editor');
	let createdInviteURL = $state('');
	let providerApps = $state.raw<ProviderApp[]>([]);
	let providerAppsLoading = $state(false);
	let providerAppsBusy = $state(false);
	let providerAppsError = $state('');
	let providerAppsRestartRequired = $state(false);
	let editingProviderAppID = $state('');
	let providerAppForm = $state<ProviderAppForm>(emptyProviderAppForm());
	let providerAppsLoadedForAdmin = false;

	interface APITokenSummary {
		id: string;
		name: string;
		token_prefix: string;
		scope: string;
		workspace_id?: string;
		expires_at?: string | null;
		last_used_at?: string | null;
		revoked_at?: string | null;
		created_at: string;
	}

	interface MCPActivityItem {
		id: string;
		workspace_id?: string;
		client_id?: string;
		client_name?: string;
		client_scope?: string;
		client_token_prefix?: string;
		tool_name: string;
		status: string;
		error_message?: string;
		duration_ms: number;
		created_at: string;
	}

	interface BillingStatus {
		workspace_id: string;
		provider?: string;
		status: string;
		plan_id?: string;
		current_period_end?: string;
		cancel_at_period_end: boolean;
		limits: Record<string, number>;
		usage: Record<string, number>;
		period_start: string;
	}

	interface TeamMember {
		user_id: string;
		email: string;
		role: string;
	}

	interface WorkspaceInvitation {
		id: string;
		workspace_id: string;
		email: string;
		role: string;
		invited_by_user_id: string;
		accepted_by_user_id?: string;
		token?: string;
		accept_url?: string;
		expires_at: string;
		accepted_at?: string;
		revoked_at?: string;
		created_at: string;
	}

	interface WorkspaceTeam {
		members: TeamMember[];
		invitations: WorkspaceInvitation[];
		current_seats: number;
	}

	type ProviderApp = components['schemas']['ProviderAppResponse'];
	type SaveProviderAppInput = components['schemas']['SaveProviderAppInputBody'];

	interface ProviderAppForm {
		provider: string;
		name: string;
		client_id: string;
		client_secret: string;
		redirect_uri: string;
		instance_url: string;
		is_active: boolean;
	}

	function emptyProviderAppForm(): ProviderAppForm {
		return {
			provider: 'x',
			name: '',
			client_id: '',
			client_secret: '',
			redirect_uri: '',
			instance_url: '',
			is_active: true
		};
	}

	const providerAppOptions = [
		{
			value: 'x',
			label: 'X / Twitter',
			description: 'OAuth app for X publishing and account connection.'
		},
		{
			value: 'mastodon',
			label: 'Mastodon',
			description: 'OAuth app for one federated Mastodon instance.'
		},
		{
			value: 'linkedin',
			label: 'LinkedIn',
			description: 'LinkedIn app for company and member posting.'
		},
		{
			value: 'threads',
			label: 'Threads',
			description: 'Meta app credentials for Threads publishing.'
		},
		{
			value: 'facebook',
			label: 'Facebook Pages',
			description: 'Meta app credentials for Facebook Page publishing.'
		},
		{
			value: 'instagram',
			label: 'Instagram Business',
			description: 'Meta app credentials for Instagram media publishing.'
		},
		{
			value: 'youtube',
			label: 'YouTube',
			description: 'Google OAuth app for YouTube channel uploads.'
		},
		{
			value: 'tiktok',
			label: 'TikTok',
			description: 'TikTok developer app for video publishing.'
		}
	];

	const authState = $derived($auth);
	const userIsInstanceAdmin = $derived(Boolean(authState.user?.is_admin));
	const passkeyCount = $derived(securityStatus?.passkeys.length ?? 0);
	const teamMembers = $derived(workspaceTeam?.members ?? []);
	const pendingInvitations = $derived(workspaceTeam?.invitations ?? []);
	const currentTeamSeats = $derived(workspaceTeam?.current_seats ?? 0);
	const inviteRoleOptions = [
		{ value: 'editor', label: 'Editor', description: 'Can create and manage workspace content.' },
		{
			value: 'viewer',
			label: 'Viewer',
			description: 'Can inspect workspace content and settings.'
		},
		{
			value: 'admin',
			label: 'Admin',
			description: 'Can manage billing, team access, and settings.'
		}
	];
	const selectedInviteRole = $derived(
		inviteRoleOptions.find((option) => option.value === inviteRole) ?? inviteRoleOptions[0]
	);
	const apiTokenScopeOptions = [
		{
			value: 'mcp:full',
			label: 'MCP / ChatGPT App',
			description: 'For ChatGPT, Claude, and other MCP clients.'
		},
		{
			value: 'cli:full',
			label: 'CLI / automation',
			description: 'For OpenPost CLI, CI, cron, and scripts.'
		}
	];
	const selectedAPITokenScope = $derived(
		apiTokenScopeOptions.find((option) => option.value === apiTokenScope) ?? apiTokenScopeOptions[0]
	);
	const apiTokenWorkspaceOptions = $derived([
		{
			value: 'current',
			label: 'Current workspace',
			description: workspaceCtx.currentWorkspace?.name ?? 'Selected workspace only.'
		},
		{
			value: 'all',
			label: 'All workspaces',
			description: 'Every workspace you can access.'
		}
	]);
	const selectedAPITokenWorkspaceScope = $derived(
		apiTokenWorkspaceOptions.find((option) => option.value === apiTokenWorkspaceScope) ??
			apiTokenWorkspaceOptions[0]
	);
	const billingPlans = [
		{
			id: 'starter',
			name: 'Starter',
			price: '$6',
			description: 'Small projects that need managed posting without extra workspace overhead.',
			limits: ['1 workspace', '3 social accounts', '100 scheduled posts/month', '1 GB media']
		},
		{
			id: 'creator',
			name: 'Creator',
			price: '$12',
			description: 'Mainstream platform scheduling for active creators and operator-led brands.',
			limits: ['3 workspaces', '6 social accounts', '500 scheduled posts/month', '5 GB media'],
			featured: true
		},
		{
			id: 'pro',
			name: 'Pro',
			price: '$24',
			description: 'Higher limits for teams, heavier media use, and larger publishing operations.',
			limits: ['10 workspaces', '15 social accounts', '2,500 scheduled posts/month', '25 GB media']
		}
	];
	const selectedProviderAppOption = $derived(
		providerAppOptions.find((option) => option.value === providerAppForm.provider) ??
			providerAppOptions[0]
	);
	const editingProviderApp = $derived(
		providerApps.find((app) => app.id === editingProviderAppID) ?? null
	);
	const providerAppNeedsInstanceURL = $derived(providerAppForm.provider === 'mastodon');
	const providerAppNeedsSecret = $derived(
		providerAppForm.provider === 'mastodon' && !editingProviderApp?.secret_configured
	);
	const settingsSections = $derived([
		{ id: 'workspace', label: 'Workspace' },
		{ id: 'team', label: 'Team' },
		{ id: 'billing', label: 'Billing' },
		...(userIsInstanceAdmin ? [{ id: 'provider-apps', label: 'Provider Apps' }] : []),
		{ id: 'security', label: 'Security' },
		{ id: 'tokens', label: 'Tokens & MCP' },
		{ id: 'date-time', label: 'Date & Time' },
		{ id: 'media-cleanup', label: 'Media Cleanup' },
		{ id: 'posting-schedule', label: 'Posting Schedule' },
		{ id: 'natural-posting', label: 'Natural Posting' },
		{ id: 'slot-defaults', label: 'Slot Defaults' }
	]);
	const billingMetricLabels: Record<string, string> = {
		scheduled_posts_monthly: 'Scheduled posts',
		published_posts_monthly: 'Published posts',
		media_bytes_uploaded_monthly: 'Uploaded media',
		provider_write_calls_monthly: 'Provider publish calls'
	};
	const currentBillingPlan = $derived(
		billingPlans.find((plan) => plan.id === billingStatus?.plan_id) ?? null
	);
	const requestedBillingPlan = $derived.by(() => {
		const planID = hostedPlanFromSearchParams(page.url.searchParams);
		return billingPlans.some((plan) => plan.id === planID) ? planID : '';
	});
	const monthlyBillingUsageRows = $derived.by(() => {
		if (!billingStatus) return [];
		return Object.entries(billingStatus.limits)
			.filter(([metric]) => metric.endsWith('_monthly'))
			.map(([metric, limit]) => ({
				metric,
				label: billingMetricLabels[metric] ?? metric.replaceAll('_', ' '),
				current: billingStatus?.usage[metric] ?? 0,
				limit
			}));
	});

	async function loadSecurityStatus() {
		loadingSecurity = true;
		securityError = '';
		try {
			const { data, error: err } = await (client as any).GET('/auth/security');
			if (err || !data) throw new Error(err?.detail || 'Failed to load account security');
			securityStatus = data;
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			loadingSecurity = false;
		}
	}

	async function loadAuthSessions() {
		authSessionsLoading = true;
		authSessionsError = '';
		try {
			const { data, error: err } = await (client as any).GET('/auth/sessions');
			if (err || !data) throw new Error(err?.detail || 'Failed to load active sessions');
			authSessions = data as AuthSessionSummary[];
		} catch (e) {
			authSessions = [];
			authSessionsError = (e as Error).message;
		} finally {
			authSessionsLoading = false;
		}
	}

	async function loadAPITokens() {
		apiTokensLoading = true;
		securityError = '';
		try {
			const { data, error: err } = await (client as any).GET('/api-tokens');
			if (err || !data) throw new Error(err?.detail || 'Failed to load API tokens');
			apiTokens = data as APITokenSummary[];
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			apiTokensLoading = false;
		}
	}

	async function loadMCPActivity() {
		mcpActivityLoading = true;
		mcpActivityError = '';
		try {
			const { data, error: err } = await (client as any).GET('/mcp/activity', {
				params: { query: { limit: 8 } }
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to load MCP activity');
			mcpActivity = data as MCPActivityItem[];
		} catch (e) {
			mcpActivityError = (e as Error).message;
			mcpActivity = [];
		} finally {
			mcpActivityLoading = false;
		}
	}

	async function loadWorkspaceTeam() {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		teamLoading = true;
		teamError = '';
		try {
			const { data, error: err } = await (client as any).GET('/workspaces/{id}/team', {
				params: { path: { id: workspaceID } }
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to load workspace team');
			workspaceTeam = data as WorkspaceTeam;
		} catch (e) {
			workspaceTeam = null;
			teamError = (e as Error).message;
		} finally {
			teamLoading = false;
		}
	}

	async function createWorkspaceInvitation(event: SubmitEvent) {
		event.preventDefault();
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		teamBusy = true;
		teamError = '';
		createdInviteURL = '';
		try {
			const { data, error: err } = await (client as any).POST('/workspaces/{id}/invitations', {
				params: { path: { id: workspaceID } },
				body: {
					email: inviteEmail.trim(),
					role: inviteRole
				}
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to create workspace invitation');
			const invitation = data as WorkspaceInvitation;
			createdInviteURL =
				invitation.accept_url ||
				(invitation.token ? `${window.location.origin}/invite?token=${invitation.token}` : '');
			inviteEmail = '';
			inviteRole = 'editor';
			await loadWorkspaceTeam();
			toastMessage = 'Invitation link created';
		} catch (e) {
			teamError = (e as Error).message;
		} finally {
			teamBusy = false;
		}
	}

	async function revokeWorkspaceInvitation(invitationID: string) {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		if (!confirm('Revoke this workspace invitation? The invite link will stop working.')) return;
		teamBusy = true;
		teamError = '';
		try {
			const { error: err } = await (client as any).DELETE(
				'/workspaces/{id}/invitations/{invitation_id}',
				{
					params: { path: { id: workspaceID, invitation_id: invitationID } }
				}
			);
			if (err) throw new Error(err.detail || 'Failed to revoke workspace invitation');
			await loadWorkspaceTeam();
			toastMessage = 'Invitation revoked';
		} catch (e) {
			teamError = (e as Error).message;
		} finally {
			teamBusy = false;
		}
	}

	async function copyCreatedInviteURL() {
		if (!createdInviteURL) return;
		await navigator.clipboard.writeText(createdInviteURL);
		toastMessage = 'Invite link copied';
	}

	async function createAPIToken() {
		apiTokenBusy = true;
		securityError = '';
		createdAPIToken = '';
		const fallbackName = apiTokenScope === 'mcp:full' ? 'OpenPost MCP' : 'OpenPost CLI';
		const workspaceID =
			apiTokenWorkspaceScope === 'current' ? (workspaceCtx.currentWorkspace?.id ?? '') : '';
		try {
			const { data, error: err } = await (client as any).POST('/api-tokens', {
				body: {
					name: apiTokenName.trim() || fallbackName,
					scope: apiTokenScope,
					...(workspaceID ? { workspace_id: workspaceID } : {})
				}
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to create API token');
			createdAPIToken = data.token;
			apiTokenName = fallbackName;
			await loadAPITokens();
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			apiTokenBusy = false;
		}
	}

	async function revokeAuthSession(session: AuthSessionSummary) {
		const prompt = session.current
			? 'Sign out of this browser? You will need to log in again.'
			: 'Revoke this browser session?';
		if (!confirm(prompt)) return;
		authSessionBusyID = session.id;
		authSessionsError = '';
		try {
			const { data, error: err } = await (client as any).DELETE('/auth/sessions/{session_id}', {
				params: { path: { session_id: session.id } }
			});
			if (err) throw new Error(err.detail || 'Failed to revoke session');
			if (data?.revoked_current || session.current) {
				auth.logout();
				await goto('/login');
				return;
			}
			await loadAuthSessions();
			toastMessage = 'Session revoked';
		} catch (e) {
			authSessionsError = (e as Error).message;
		} finally {
			authSessionBusyID = '';
		}
	}

	async function revokeAPIToken(tokenID: string) {
		if (!confirm('Revoke this API token? Any CLI or automation using it will stop working.'))
			return;
		apiTokenBusy = true;
		securityError = '';
		try {
			const { error: err } = await (client as any).DELETE('/api-tokens/{id}', {
				params: { path: { id: tokenID } }
			});
			if (err) throw new Error(err.detail || 'Failed to revoke API token');
			await loadAPITokens();
			toastMessage = 'API token revoked';
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			apiTokenBusy = false;
		}
	}

	async function loadBillingStatus() {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		billingStatusLoading = true;
		billingError = '';
		try {
			const { data, error: err } = await client.GET('/billing/status', {
				params: { query: { workspace_id: workspaceID } }
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to load billing status');
			billingStatus = data as BillingStatus;
		} catch (e) {
			billingStatus = null;
			billingError = (e as Error).message;
		} finally {
			billingStatusLoading = false;
		}
	}

	async function startCheckout(planID: string) {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		billingBusyPlan = planID;
		billingError = '';
		try {
			const { data, error: err } = await client.POST('/billing/checkout', {
				body: { workspace_id: workspaceID, plan_id: planID }
			});
			if (err || !data?.url) throw new Error(err?.detail || 'Failed to create checkout');
			window.location.assign(data.url);
		} catch (e) {
			billingError = (e as Error).message;
		} finally {
			billingBusyPlan = '';
		}
	}

	async function openBillingPortal() {
		const workspaceID = workspaceCtx.currentWorkspace?.id;
		if (!workspaceID) return;
		billingPortalBusy = true;
		billingError = '';
		try {
			const { data, error: err } = await client.POST('/billing/portal', {
				body: { workspace_id: workspaceID }
			});
			if (err || !data?.url) throw new Error(err?.detail || 'Failed to open billing portal');
			window.location.assign(data.url);
		} catch (e) {
			billingError = (e as Error).message;
		} finally {
			billingPortalBusy = false;
		}
	}

	function providerAppLabel(provider: string): string {
		return providerAppOptions.find((option) => option.value === provider)?.label ?? provider;
	}

	function resetProviderAppForm() {
		providerAppForm = emptyProviderAppForm();
		editingProviderAppID = '';
		providerAppsError = '';
	}

	function editProviderApp(app: ProviderApp) {
		providerAppForm = {
			provider: app.provider,
			name: app.name ?? '',
			client_id: app.client_id,
			client_secret: '',
			redirect_uri: app.redirect_uri ?? '',
			instance_url: app.instance_url ?? '',
			is_active: app.is_active
		};
		editingProviderAppID = app.id;
		providerAppsError = '';
	}

	async function loadProviderApps() {
		if (!userIsInstanceAdmin) return;
		providerAppsLoading = true;
		providerAppsError = '';
		try {
			const { data, error: err } = await client.GET('/admin/provider-apps');
			if (err) throw new Error(err.detail || 'Failed to load provider apps');
			providerApps = data ?? [];
		} catch (e) {
			providerApps = [];
			providerAppsError = (e as Error).message;
		} finally {
			providerAppsLoading = false;
		}
	}

	async function saveProviderApp(event: SubmitEvent) {
		event.preventDefault();
		providerAppsError = '';
		const clientID = providerAppForm.client_id.trim();
		const instanceURL = providerAppForm.instance_url.trim();
		if (!clientID) {
			providerAppsError = 'Client ID is required.';
			return;
		}
		if (providerAppNeedsInstanceURL && !instanceURL) {
			providerAppsError = 'Mastodon provider apps need an instance URL.';
			return;
		}
		if (providerAppNeedsSecret && !providerAppForm.client_secret.trim()) {
			providerAppsError = 'Mastodon provider apps need a client secret.';
			return;
		}

		const body: SaveProviderAppInput = {
			provider: providerAppForm.provider,
			client_id: clientID,
			is_active: providerAppForm.is_active
		};
		const name = providerAppForm.name.trim();
		const redirectURI = providerAppForm.redirect_uri.trim();
		const clientSecret = providerAppForm.client_secret.trim();
		if (name) body.name = name;
		if (redirectURI) body.redirect_uri = redirectURI;
		if (instanceURL) body.instance_url = instanceURL;
		if (clientSecret) body.client_secret = clientSecret;

		providerAppsBusy = true;
		try {
			const { data, error: err } = await client.POST('/admin/provider-apps', { body });
			if (err || !data) throw new Error(err?.detail || 'Failed to save provider app');
			providerAppsRestartRequired = providerAppsRestartRequired || data.requires_restart;
			toastMessage = data.existed ? 'Provider app updated' : 'Provider app created';
			resetProviderAppForm();
			await loadProviderApps();
		} catch (e) {
			providerAppsError = (e as Error).message;
		} finally {
			providerAppsBusy = false;
		}
	}

	async function deleteProviderApp(app: ProviderApp) {
		if (
			!confirm(
				`Delete ${providerAppLabel(app.provider)} provider app? Connected accounts keep their stored tokens, but new OAuth connections for this provider may stop working after restart.`
			)
		) {
			return;
		}
		providerAppsBusy = true;
		providerAppsError = '';
		try {
			const { data, error: err } = await client.DELETE('/admin/provider-apps/{id}', {
				params: { path: { id: app.id } }
			});
			if (err) throw new Error(err.detail || 'Failed to delete provider app');
			providerAppsRestartRequired = providerAppsRestartRequired || Boolean(data?.requires_restart);
			if (editingProviderAppID === app.id) resetProviderAppForm();
			await loadProviderApps();
			toastMessage = 'Provider app deleted';
		} catch (e) {
			providerAppsError = (e as Error).message;
		} finally {
			providerAppsBusy = false;
		}
	}

	async function startTOTPSetup() {
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await (client as any).POST('/auth/security/totp/setup', {
				body: { current_password: currentPassword }
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to start authenticator setup');
			totpSetupChallengeId = data.challenge_id;
			totpManualEntryKey = data.manual_entry_key;
			totpQRCodeDataURL = data.qr_code_data_url;
			totpCode = '';
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function confirmTOTPSetup() {
		if (!totpSetupChallengeId) return;
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await (client as any).POST('/auth/security/totp/confirm', {
				body: {
					challenge_id: totpSetupChallengeId,
					code: totpCode
				}
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to confirm authenticator app');
			securityStatus = data;
			totpSetupChallengeId = '';
			totpManualEntryKey = '';
			totpQRCodeDataURL = '';
			totpCode = '';
			currentPassword = '';
			toastMessage = 'Authenticator app enabled';
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function disableTOTP() {
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await (client as any).POST('/auth/security/totp/disable', {
				body: { current_password: currentPassword }
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to disable authenticator app');
			securityStatus = data;
			currentPassword = '';
			toastMessage = 'Authenticator app disabled';
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function addPasskey() {
		securityBusy = true;
		securityError = '';
		try {
			const { data: beginData, error: beginError } = await (client as any).POST(
				'/auth/security/passkeys/begin',
				{
					body: {
						current_password: currentPassword,
						name: newPasskeyName
					}
				}
			);
			if (beginError || !beginData) {
				throw new Error(beginError?.detail || 'Failed to start passkey registration');
			}

			const credential = await createPasskeyCredential(beginData.options);
			const { data, error: err } = await (client as any).POST('/auth/security/passkeys/finish', {
				body: {
					challenge_id: beginData.challenge_id,
					name: newPasskeyName,
					credential
				}
			});
			if (err || !data) throw new Error(err?.detail || 'Failed to save passkey');
			securityStatus = data;
			currentPassword = '';
			newPasskeyName = '';
			toastMessage = 'Passkey added';
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function removePasskey(passkeyId: string) {
		securityBusy = true;
		securityError = '';
		try {
			const { data, error: err } = await (client as any).POST(
				'/auth/security/passkeys/{passkey_id}/remove',
				{
					params: { path: { passkey_id: passkeyId } },
					body: { current_password: currentPassword }
				}
			);
			if (err || !data) throw new Error(err?.detail || 'Failed to remove passkey');
			securityStatus = data;
			currentPassword = '';
			toastMessage = 'Passkey removed';
		} catch (e) {
			securityError = (e as Error).message;
		} finally {
			securityBusy = false;
		}
	}

	async function saveSettings() {
		saving = true;
		try {
			await workspaceCtx.saveSettings({
				timezone: workspaceCtx.settings.timezone,
				week_start: workspaceCtx.settings.week_start,
				media_cleanup_days: workspaceCtx.settings.media_cleanup_days,
				random_delay_minutes: workspaceCtx.settings.random_delay_minutes,
				draft_gap_minutes: workspaceCtx.settings.draft_gap_minutes,
				slot_start_hour: workspaceCtx.settings.slot_start_hour,
				slot_end_hour: workspaceCtx.settings.slot_end_hour,
				slot_interval_minutes: workspaceCtx.settings.slot_interval_minutes
			});
			toastMessage = 'Settings saved successfully';
		} catch (e) {
			toastMessage = (e as Error).message;
		} finally {
			saving = false;
		}
	}

	function parseDurationInput(input: string, allowZero: boolean = false): number | null {
		input = input.trim().toLowerCase();
		const direct = parseInt(input, 10);
		if (!isNaN(direct) && String(direct) === input && (direct > 0 || (allowZero && direct === 0))) {
			return direct;
		}
		const hourMatch = input.match(/(\d+)\s*h/);
		const minMatch = input.match(/(\d+)\s*m/);
		let total = 0;
		if (hourMatch) total += parseInt(hourMatch[1], 10) * 60;
		if (minMatch) total += parseInt(minMatch[1], 10);
		if (total > 0) return total;
		return null;
	}

	let intervalInput = $state(String(workspaceCtx.settings.slot_interval_minutes));
	let intervalError = $state('');
	let draftGapInput = $state(String(workspaceCtx.settings.draft_gap_minutes));
	let draftGapError = $state('');

	function handleIntervalChange(value: string) {
		intervalInput = value;
		const parsed = parseDurationInput(value);
		if (parsed !== null && parsed >= 1 && parsed <= 180) {
			intervalError = '';
			workspaceCtx.settings.slot_interval_minutes = parsed;
		} else if (value.trim() !== '') {
			intervalError = 'Enter a duration between 1 minute and 3 hours (e.g. 15m, 1h, 30)';
		}
	}

	function handleDraftGapChange(value: string) {
		draftGapInput = value;
		const parsed = parseDurationInput(value, true);
		if (parsed !== null && parsed >= 0 && parsed <= 24 * 60) {
			draftGapError = '';
			workspaceCtx.settings.draft_gap_minutes = parsed;
		} else if (value.trim() !== '') {
			draftGapError = 'Enter a duration between 0 minutes and 24 hours (e.g. 45m, 2h, 0)';
		}
	}

	interface PostingSchedule {
		id: string;
		workspace_id: string;
		set_id: string;
		utc_hour: number;
		utc_minute: number;
		day_of_week: number;
		local_hour: number;
		local_minute: number;
		local_day_of_week: number;
		label: string;
		is_active: boolean;
		created_at: string;
	}

	interface ScheduleRow {
		key: string;
		local_hour: number;
		local_minute: number;
		label: string;
		days: Record<number, PostingSchedule | undefined>;
	}

	let schedules = $state<PostingSchedule[]>([]);
	let loadingSchedules = $state(false);
	let showSuggestSchedule = $state(false);
	let suggestedPostsPerDay = $state(3);
	let generatingSchedule = $state(false);
	let newTimeInput = $state('09:00');
	let newTimeError = $state('');
	let newTimeDays = $state<number[]>([1, 2, 3, 4, 5]);

	const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
	const dayShortNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

	const dayOrder = $derived.by(() => {
		const start = workspaceCtx.settings.week_start === 0 ? 0 : 1;
		return Array.from({ length: 7 }, (_, index) => (start + index) % 7);
	});

	const scheduleRows = $derived.by(() => {
		const rows = new Map<string, ScheduleRow>();
		for (const schedule of schedules) {
			const key = `${schedule.local_hour}:${schedule.local_minute}`;
			if (!rows.has(key)) {
				rows.set(key, {
					key,
					local_hour: schedule.local_hour,
					local_minute: schedule.local_minute,
					label: schedule.label,
					days: {}
				});
			}
			const row = rows.get(key)!;
			row.days[schedule.local_day_of_week] = schedule;
			if (!row.label && schedule.label) {
				row.label = schedule.label;
			}
		}
		return Array.from(rows.values()).sort(
			(a, b) => a.local_hour * 60 + a.local_minute - (b.local_hour * 60 + b.local_minute)
		);
	});

	async function loadSchedules() {
		if (!workspaceCtx.currentWorkspace) return;
		loadingSchedules = true;
		try {
			const { data, error: err } = await (client as any).GET('/posting-schedules', {
				params: { query: { workspace_id: workspaceCtx.currentWorkspace.id } }
			});
			if (!err && data) {
				schedules = data;
			}
		} catch (e) {
			console.error('Failed to load schedules:', e);
		} finally {
			loadingSchedules = false;
		}
	}

	function parseClockInput(value: string): { hour: number; minute: number } | null {
		const match = value.trim().match(/^(\d{1,2}):(\d{2})$/);
		if (!match) return null;
		const hour = Number(match[1]);
		const minute = Number(match[2]);
		if (hour < 0 || hour > 23 || minute < 0 || minute > 59) return null;
		return { hour, minute };
	}

	async function createSchedule(dayOfWeek: number, localHour: number, localMinute: number) {
		if (!workspaceCtx.currentWorkspace) return;
		const { error: err } = await (client as any).POST('/posting-schedules', {
			body: {
				workspace_id: workspaceCtx.currentWorkspace.id,
				local_day_of_week: dayOfWeek,
				local_hour: localHour,
				local_minute: localMinute,
				day_of_week: 0,
				utc_hour: 0,
				utc_minute: 0,
				label: ''
			}
		});
		if (err) throw err;
	}

	async function addTimeRow() {
		const parsed = parseClockInput(newTimeInput);
		if (!parsed) {
			newTimeError = 'Use HH:MM in 24-hour format.';
			return;
		}
		if (newTimeDays.length === 0) {
			newTimeError = 'Select at least one day.';
			return;
		}
		newTimeError = '';
		try {
			for (const day of newTimeDays) {
				const exists = schedules.some(
					(schedule) =>
						schedule.local_day_of_week === day &&
						schedule.local_hour === parsed.hour &&
						schedule.local_minute === parsed.minute
				);
				if (!exists) {
					await createSchedule(day, parsed.hour, parsed.minute);
				}
			}
			await loadSchedules();
			toastMessage = 'Time row added successfully';
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to add schedule row';
		}
	}

	async function deleteSchedule(id: string) {
		try {
			const { error: err } = await (client as any).DELETE('/posting-schedules/{id}', {
				params: { path: { id } }
			});
			if (err) throw err;
			await loadSchedules();
			toastMessage = 'Schedule deleted successfully';
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to delete schedule';
		}
	}

	async function toggleScheduleCell(row: ScheduleRow, dayOfWeek: number) {
		try {
			const existing = row.days[dayOfWeek];
			if (existing) {
				await deleteSchedule(existing.id);
				return;
			}
			await createSchedule(dayOfWeek, row.local_hour, row.local_minute);
			await loadSchedules();
			toastMessage = 'Schedule updated successfully';
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to update schedule';
		}
	}

	async function removeTimeRow(row: ScheduleRow) {
		try {
			for (const schedule of Object.values(row.days)) {
				if (schedule) {
					const { error: err } = await (client as any).DELETE('/posting-schedules/{id}', {
						params: { path: { id: schedule.id } }
					});
					if (err) throw err;
				}
			}
			await loadSchedules();
			toastMessage = 'Time row removed successfully';
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to remove schedule row';
		}
	}

	function toggleNewDay(dayOfWeek: number) {
		if (newTimeDays.includes(dayOfWeek)) {
			newTimeDays = newTimeDays.filter((value) => value !== dayOfWeek);
			return;
		}
		newTimeDays = [...newTimeDays, dayOfWeek].sort((a, b) => a - b);
	}

	async function generateSuggestedSchedule() {
		if (!workspaceCtx.currentWorkspace) return;
		generatingSchedule = true;
		try {
			const { error: err } = await (client as any).POST('/posting-schedules/suggest', {
				body: {
					workspace_id: workspaceCtx.currentWorkspace.id,
					posts_per_day: suggestedPostsPerDay
				}
			});
			if (err) throw err;
			showSuggestSchedule = false;
			await loadSchedules();
			toastMessage = `Generated suggested schedule with ${suggestedPostsPerDay} posts per day`;
		} catch (e) {
			toastMessage = (e as Error).message || 'Failed to generate schedule';
		} finally {
			generatingSchedule = false;
		}
	}

	function formatTime(hour: number, minute: number): string {
		return new Date(Date.UTC(2024, 0, 1, hour, minute)).toLocaleTimeString(getLocaleTag(), {
			hour: 'numeric',
			minute: '2-digit',
			timeZone: 'UTC'
		});
	}

	function formatBillingValue(metric: string, value: number): string {
		if (metric.includes('bytes')) {
			return formatBytes(value);
		}
		return new Intl.NumberFormat(getLocaleTag()).format(value);
	}

	function formatBytes(value: number): string {
		if (value >= 1_000_000_000) {
			return `${(value / 1_000_000_000).toFixed(value % 1_000_000_000 === 0 ? 0 : 1)} GB`;
		}
		if (value >= 1_000_000) {
			return `${(value / 1_000_000).toFixed(value % 1_000_000 === 0 ? 0 : 1)} MB`;
		}
		return `${new Intl.NumberFormat(getLocaleTag()).format(value)} B`;
	}

	function formatSessionUserAgent(value: string): string {
		const trimmed = value.trim();
		if (!trimmed) return 'Unknown browser';
		return trimmed.length > 120 ? `${trimmed.slice(0, 117)}...` : trimmed;
	}

	function formatSessionTime(value: string): string {
		if (!value || value.startsWith('0001-01-01')) return 'Never';
		return new Date(value).toLocaleString();
	}

	$effect(() => {
		if (workspaceCtx.currentWorkspace) {
			loadBillingStatus();
			loadWorkspaceTeam();
			loadSchedules();
		}
	});

	$effect(() => {
		if (
			requestedBillingPlan &&
			workspaceCtx.currentWorkspace &&
			handledCheckoutPlan !== requestedBillingPlan &&
			!billingBusyPlan
		) {
			handledCheckoutPlan = requestedBillingPlan;
			startCheckout(requestedBillingPlan);
		}
	});

	$effect(() => {
		if (authState.isAuthenticated) {
			loadSecurityStatus();
			loadAuthSessions();
			loadAPITokens();
			loadMCPActivity();
		}
	});

	$effect(() => {
		if (userIsInstanceAdmin) {
			if (!providerAppsLoadedForAdmin) {
				providerAppsLoadedForAdmin = true;
				loadProviderApps();
			}
		} else {
			providerAppsLoadedForAdmin = false;
		}
	});

	$effect(() => {
		intervalInput = String(workspaceCtx.settings.slot_interval_minutes);
		draftGapInput = String(workspaceCtx.settings.draft_gap_minutes);
	});

	function handleTimezoneChange(value: string) {
		workspaceCtx.settings.timezone = value;
	}

	function handleWeekStartChange(value: number) {
		workspaceCtx.settings.week_start = value;
	}

	function handleCleanupDaysChange(value: number) {
		workspaceCtx.settings.media_cleanup_days = value;
	}
</script>

<svelte:head>
	<title>Settings - OpenPost</title>
</svelte:head>

{#if toastMessage}
	<div
		class="pointer-events-auto fixed right-4 bottom-4 z-50 mb-4 flex items-center gap-2 rounded-lg border bg-background px-4 py-3 shadow-lg"
	>
		<span class="text-sm">{toastMessage}</span>
		<button onclick={() => (toastMessage = '')}>
			<XIcon class="size-4" />
		</button>
	</div>
{/if}

<PageContainer
	title="Settings"
	description="Manage your workspace preferences"
	icon={SettingsIcon}
	loading={!workspaceCtx.currentWorkspace}
	loadingMessage="Loading workspace..."
>
	<div class="space-y-8">
		<nav
			aria-label="Settings sections"
			data-testid="settings-section-nav"
			class="sticky top-0 z-10 -mx-4 border-y bg-background/95 px-4 py-3 backdrop-blur supports-[backdrop-filter]:bg-background/80 md:top-2 md:mx-0 md:rounded-lg md:border"
		>
			<div
				class="flex [scrollbar-width:none] gap-2 overflow-x-auto [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
			>
				{#each settingsSections as section (section.id)}
					<a
						href={`#${section.id}`}
						class="shrink-0 rounded-md border bg-background px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
					>
						{section.label}
					</a>
				{/each}
			</div>
		</nav>

		<section id="workspace" class="scroll-mt-24 space-y-4">
			<h2 class="mb-4 text-lg font-semibold">Workspace</h2>
			<div class="flex items-center gap-4">
				<span class="text-sm font-medium">Current Workspace</span>
				<span class="text-sm text-muted-foreground">{workspaceCtx.currentWorkspace?.name}</span>
			</div>
			<div class="rounded-lg border bg-muted/20 p-4">
				<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<p class="text-sm font-medium">Connected social accounts live separately</p>
						<p class="text-sm text-muted-foreground">
							Account security is personal. Social connections and sets are still managed per
							workspace on the accounts page.
						</p>
					</div>
					<Button variant="outline" onclick={() => goto('/accounts')}>Open Accounts</Button>
				</div>
			</div>
		</section>

		<section id="team" class="scroll-mt-24 rounded-lg border p-6">
			<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
				<div>
					<h2 class="flex items-center gap-2 text-lg font-semibold">
						<UsersIcon class="h-5 w-5 text-muted-foreground" />
						Team
					</h2>
					<p class="mt-2 text-sm text-muted-foreground">
						Invite collaborators into this workspace and keep pending seats visible.
					</p>
				</div>
				<div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
					<span class="font-medium">{currentTeamSeats}</span>
					<span class="text-muted-foreground"> seats reserved</span>
				</div>
			</div>

			{#if teamError}
				<div
					data-testid="team-error"
					class="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
				>
					{teamError}
				</div>
			{/if}

			<form
				onsubmit={createWorkspaceInvitation}
				class="mb-4 grid gap-3 lg:grid-cols-[minmax(0,1fr)_220px_auto]"
			>
				<div class="space-y-2">
					<Label for="team-invite-email">Invite email</Label>
					<Input
						id="team-invite-email"
						data-testid="team-invite-email"
						type="email"
						bind:value={inviteEmail}
						placeholder="teammate@example.com"
						autocomplete="email"
						required
					/>
				</div>
				<div class="space-y-2">
					<Label for="team-invite-role">Role</Label>
					<Select.Root
						type="single"
						value={inviteRole}
						onValueChange={(value) => value && (inviteRole = value)}
					>
						<Select.Trigger id="team-invite-role" data-testid="team-invite-role" class="w-full">
							{selectedInviteRole.label}
						</Select.Trigger>
						<Select.Content>
							{#each inviteRoleOptions as option (option.value)}
								<Select.Item value={option.value}>
									<div class="flex flex-col gap-0.5 text-left">
										<span>{option.label}</span>
										<span class="text-xs text-muted-foreground">{option.description}</span>
									</div>
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
				<div class="flex items-end">
					<Button type="submit" disabled={teamBusy || !inviteEmail.trim()}>
						{#if teamBusy}
							<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
						{:else}
							<UserPlusIcon class="mr-2 h-4 w-4" />
						{/if}
						Send Invite
					</Button>
				</div>
			</form>

			{#if createdInviteURL}
				<div
					data-testid="team-invite-link"
					class="mb-4 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4"
				>
					<p class="text-sm font-medium text-emerald-900">Invite link created</p>
					<div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
						<p
							class="min-w-0 flex-1 rounded-md bg-background px-3 py-2 font-mono text-xs break-all"
						>
							{createdInviteURL}
						</p>
						<Button type="button" variant="outline" size="sm" onclick={copyCreatedInviteURL}>
							<CopyIcon class="mr-2 h-4 w-4" />
							Copy
						</Button>
					</div>
				</div>
			{/if}

			{#if teamLoading}
				<div class="grid gap-3 lg:grid-cols-2">
					<Skeleton class="h-28 rounded-lg" />
					<Skeleton class="h-28 rounded-lg" />
				</div>
			{:else}
				<div class="grid gap-4 lg:grid-cols-2">
					<div>
						<h3 class="mb-2 text-sm font-semibold">Members</h3>
						<div data-testid="team-members-list" class="space-y-2">
							{#each teamMembers as member (member.user_id)}
								<div
									class="flex flex-col gap-2 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
								>
									<div class="min-w-0">
										<p class="truncate text-sm font-medium">{member.email}</p>
										<p class="text-xs text-muted-foreground">User {member.user_id}</p>
									</div>
									<span
										class="inline-flex w-fit items-center rounded-full border px-2 py-0.5 text-xs font-medium capitalize"
									>
										{member.role}
									</span>
								</div>
							{:else}
								<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
									No members found for this workspace.
								</p>
							{/each}
						</div>
					</div>

					<div>
						<h3 class="mb-2 text-sm font-semibold">Pending Invitations</h3>
						<div data-testid="team-invitations-list" class="space-y-2">
							{#each pendingInvitations as invitation (invitation.id)}
								<div
									class="flex flex-col gap-2 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
								>
									<div class="min-w-0">
										<p class="truncate text-sm font-medium">{invitation.email}</p>
										<p class="text-xs text-muted-foreground">
											{invitation.role} · expires
											{new Date(invitation.expires_at).toLocaleDateString()}
										</p>
									</div>
									<Button
										type="button"
										variant="ghost"
										size="sm"
										class="text-destructive hover:text-destructive"
										onclick={() => revokeWorkspaceInvitation(invitation.id)}
										disabled={teamBusy}
									>
										Revoke
									</Button>
								</div>
							{:else}
								<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
									No pending invitations.
								</p>
							{/each}
						</div>
					</div>
				</div>
			{/if}
		</section>

		<section id="billing" class="scroll-mt-24 rounded-lg border p-6">
			<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
				<div>
					<h2 class="flex items-center gap-2 text-lg font-semibold">
						<CreditCardIcon class="h-5 w-5 text-muted-foreground" />
						Billing
					</h2>
					<p class="mt-2 text-sm text-muted-foreground">
						Manage the OpenPost Cloud plan for this workspace.
					</p>
				</div>
				<Button variant="outline" onclick={openBillingPortal} disabled={billingPortalBusy}>
					{#if billingPortalBusy}
						<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
					{:else}
						<ExternalLinkIcon class="mr-2 h-4 w-4" />
					{/if}
					Customer Portal
				</Button>
			</div>

			{#if billingStatusLoading}
				<div class="mb-4 grid gap-3 lg:grid-cols-2">
					<Skeleton class="h-24 rounded-lg" />
					<Skeleton class="h-24 rounded-lg" />
				</div>
			{:else if billingStatus}
				<div class="mb-4 grid gap-3 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]">
					<div class="rounded-lg border bg-muted/20 p-4">
						<p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
							Current plan
						</p>
						<div class="mt-2 flex flex-wrap items-end gap-x-3 gap-y-1">
							<p class="text-2xl font-semibold">
								{currentBillingPlan?.name ?? (billingStatus.plan_id || 'No active plan')}
							</p>
							<p class="pb-1 text-sm text-muted-foreground capitalize">{billingStatus.status}</p>
						</div>
						{#if billingStatus.current_period_end}
							<p class="mt-2 text-sm text-muted-foreground">
								Period ends {new Date(billingStatus.current_period_end).toLocaleDateString()}
								{#if billingStatus.cancel_at_period_end}
									· cancels after this period
								{/if}
							</p>
						{:else}
							<p class="mt-2 text-sm text-muted-foreground">
								Start checkout to activate hosted billing for this workspace.
							</p>
						{/if}
					</div>

					<div class="rounded-lg border bg-muted/20 p-4">
						<p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
							Usage this month
						</p>
						{#if monthlyBillingUsageRows.length}
							<div class="mt-3 grid gap-3 sm:grid-cols-2">
								{#each monthlyBillingUsageRows as row (row.metric)}
									<div>
										<div class="mb-1 flex items-center justify-between gap-2 text-sm">
											<span>{row.label}</span>
											<span class="text-muted-foreground">
												{formatBillingValue(row.metric, row.current)} / {formatBillingValue(
													row.metric,
													row.limit
												)}
											</span>
										</div>
										<div class="h-2 overflow-hidden rounded-full bg-muted">
											<div
												class="h-full rounded-full bg-primary"
												style:width={`${Math.min(100, Math.round((row.current / Math.max(row.limit, 1)) * 100))}%`}
											></div>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<p class="mt-2 text-sm text-muted-foreground">
								Usage counters appear here after an active subscription snapshot is received.
							</p>
						{/if}
					</div>
				</div>
			{/if}

			<div class="grid gap-3 lg:grid-cols-3">
				{#each billingPlans as plan (plan.id)}
					<article
						class={`rounded-lg border p-4 ${plan.featured ? 'border-primary bg-primary/5 shadow-sm' : 'bg-background'}`}
					>
						<div class="mb-3 flex items-start justify-between gap-3">
							<div>
								<h3 class="font-semibold">{plan.name}</h3>
								<p class="text-sm text-muted-foreground">{plan.description}</p>
							</div>
							<div class="text-right">
								<div class="text-xl font-semibold">{plan.price}</div>
								<div class="text-xs text-muted-foreground">/mo</div>
							</div>
						</div>
						<ul class="mb-4 space-y-1 text-sm text-muted-foreground">
							{#each plan.limits as limit (limit)}
								<li>{limit}</li>
							{/each}
						</ul>
						<Button
							class="w-full"
							variant={plan.featured ? 'default' : 'outline'}
							onclick={() => startCheckout(plan.id)}
							disabled={Boolean(billingBusyPlan)}
						>
							{#if billingBusyPlan === plan.id}
								<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
							{/if}
							Start Checkout
						</Button>
					</article>
				{/each}
			</div>

			{#if billingError}
				<div
					class="mt-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
				>
					{billingError}
				</div>
			{/if}
		</section>

		{#if userIsInstanceAdmin}
			<section
				id="provider-apps"
				data-testid="provider-apps-settings"
				class="scroll-mt-24 rounded-lg border p-6"
			>
				<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
					<div>
						<h2 class="flex items-center gap-2 text-lg font-semibold">
							<KeyRoundIcon class="h-5 w-5 text-muted-foreground" />
							Provider Apps
						</h2>
						<p class="mt-2 text-sm text-muted-foreground">
							Manage encrypted OAuth app credentials used for new social account connections.
						</p>
					</div>
					<Button variant="outline" onclick={loadProviderApps} disabled={providerAppsLoading}>
						{#if providerAppsLoading}
							<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
						{/if}
						Refresh
					</Button>
				</div>

				{#if providerAppsRestartRequired}
					<div
						data-testid="provider-app-restart-required"
						class="mb-4 rounded-md border border-amber-300/60 bg-amber-50 p-3 text-sm text-amber-950"
					>
						Provider app changes are saved. Restart the OpenPost server before adapter registry
						changes apply.
					</div>
				{/if}

				{#if providerAppsError}
					<div
						data-testid="provider-app-error"
						class="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
					>
						{providerAppsError}
					</div>
				{/if}

				<form onsubmit={saveProviderApp} class="mb-6 rounded-lg border bg-muted/20 p-4">
					<div class="mb-4 flex flex-col gap-1">
						<h3 class="text-sm font-semibold">
							{editingProviderAppID ? 'Edit provider app' : 'Add provider app'}
						</h3>
						<p class="text-sm text-muted-foreground">
							Secrets are write-only. Leave the secret blank while editing to keep the stored value.
						</p>
					</div>

					<div class="grid gap-4 lg:grid-cols-2">
						<div class="space-y-2">
							<Label for="provider-app-provider">Provider</Label>
							{#if editingProviderAppID}
								<Input
									id="provider-app-provider"
									value={selectedProviderAppOption.label}
									disabled
								/>
							{:else}
								<Select.Root
									type="single"
									value={providerAppForm.provider}
									onValueChange={(value) => {
										if (!value) return;
										providerAppForm.provider = value;
										if (value !== 'mastodon') providerAppForm.instance_url = '';
									}}
								>
									<Select.Trigger
										id="provider-app-provider"
										data-testid="provider-app-provider"
										class="w-full"
									>
										{selectedProviderAppOption.label}
									</Select.Trigger>
									<Select.Content>
										{#each providerAppOptions as option (option.value)}
											<Select.Item value={option.value}>
												<div class="flex flex-col gap-0.5 text-left">
													<span>{option.label}</span>
													<span class="text-xs text-muted-foreground">{option.description}</span>
												</div>
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							{/if}
						</div>

						<div class="space-y-2">
							<Label for="provider-app-name">Display name</Label>
							<Input
								id="provider-app-name"
								bind:value={providerAppForm.name}
								placeholder="Production app, EU Mastodon, Meta main"
							/>
						</div>

						{#if providerAppNeedsInstanceURL}
							<div class="space-y-2">
								<Label for="provider-app-instance">Instance URL</Label>
								<Input
									id="provider-app-instance"
									bind:value={providerAppForm.instance_url}
									placeholder="https://mastodon.social"
									disabled={Boolean(editingProviderAppID)}
									required
								/>
							</div>
						{/if}

						<div class="space-y-2">
							<Label for="provider-app-client-id">Client ID</Label>
							<Input
								id="provider-app-client-id"
								data-testid="provider-app-client-id"
								bind:value={providerAppForm.client_id}
								placeholder="OAuth client ID"
								autocomplete="off"
								required
							/>
						</div>

						<div class="space-y-2">
							<Label for="provider-app-client-secret">
								Client secret{providerAppNeedsSecret ? '' : ' (optional on edit)'}
							</Label>
							<Input
								id="provider-app-client-secret"
								data-testid="provider-app-client-secret"
								type="password"
								bind:value={providerAppForm.client_secret}
								placeholder={editingProviderApp?.secret_configured
									? 'Leave blank to keep stored secret'
									: 'OAuth client secret'}
								autocomplete="off"
								required={providerAppNeedsSecret}
							/>
						</div>

						<div class="space-y-2 lg:col-span-2">
							<Label for="provider-app-redirect">Redirect URI</Label>
							<Input
								id="provider-app-redirect"
								bind:value={providerAppForm.redirect_uri}
								placeholder="Optional. Defaults to the provider callback under OPENPOST_APP_URL."
							/>
						</div>
					</div>

					<div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<label class="flex items-center gap-2 text-sm">
							<Checkbox
								checked={providerAppForm.is_active}
								onCheckedChange={(checked) => (providerAppForm.is_active = checked === true)}
							/>
							<span>Active after the next server start</span>
						</label>
						<div class="flex gap-2">
							{#if editingProviderAppID}
								<Button type="button" variant="outline" onclick={resetProviderAppForm}>
									Cancel
								</Button>
							{/if}
							<Button
								type="submit"
								data-testid="provider-app-save"
								disabled={providerAppsBusy || !providerAppForm.client_id.trim()}
							>
								{#if providerAppsBusy}
									<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
								{/if}
								{editingProviderAppID ? 'Update App' : 'Save App'}
							</Button>
						</div>
					</div>
				</form>

				{#if providerAppsLoading}
					<div class="grid gap-3 lg:grid-cols-2">
						<Skeleton class="h-28 rounded-lg" />
						<Skeleton class="h-28 rounded-lg" />
					</div>
				{:else if providerApps.length === 0}
					<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
						No database-backed provider apps are configured yet. Legacy env and JSON bootstrap apps
						still load at server startup.
					</p>
				{:else}
					<div data-testid="provider-app-list" class="grid gap-3 lg:grid-cols-2">
						{#each providerApps as app (app.id)}
							<div data-testid="provider-app-row" class="flex flex-col gap-3 rounded-lg border p-4">
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<div class="flex flex-wrap items-center gap-2">
											<h3 class="font-semibold">{providerAppLabel(app.provider)}</h3>
											<span
												class={[
													'rounded-full border px-2 py-0.5 text-xs font-medium',
													app.is_active
														? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700'
														: 'bg-muted text-muted-foreground'
												]}
											>
												{app.is_active ? 'Active' : 'Inactive'}
											</span>
										</div>
										{#if app.name}
											<p class="mt-1 text-sm text-muted-foreground">{app.name}</p>
										{/if}
										{#if app.instance_url}
											<p class="mt-1 font-mono text-xs break-all text-muted-foreground">
												{app.instance_url}
											</p>
										{/if}
									</div>
									<div class="flex shrink-0 gap-1">
										<Button
											type="button"
											variant="ghost"
											size="sm"
											onclick={() => editProviderApp(app)}
											disabled={providerAppsBusy}
										>
											Edit
										</Button>
										<Button
											type="button"
											variant="ghost"
											size="sm"
											class="text-destructive hover:text-destructive"
											onclick={() => deleteProviderApp(app)}
											disabled={providerAppsBusy}
										>
											Delete
										</Button>
									</div>
								</div>
								<div class="rounded-md bg-muted/30 p-3 text-xs text-muted-foreground">
									<p>
										Client ID
										<span class="font-mono break-all text-foreground">{app.client_id}</span>
									</p>
									<p class="mt-1">
										Client secret:
										<span class="font-medium text-foreground">
											{app.secret_configured ? 'stored' : 'not stored'}
										</span>
									</p>
									{#if app.redirect_uri}
										<p class="mt-1 break-all">Redirect URI {app.redirect_uri}</p>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</section>
		{/if}

		<section id="security" class="scroll-mt-24 rounded-lg border p-6">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
				<ShieldCheckIcon class="h-5 w-5 text-muted-foreground" />
				Account Security
			</h2>
			<p class="mb-4 text-sm text-muted-foreground">
				Turn on two-factor authentication for your user account with an authenticator app and
				optional passkeys. These protections follow your login, not your workspace.
			</p>

			{#if loadingSecurity}
				<div class="space-y-3">
					<Skeleton class="h-24 rounded-lg" />
					<Skeleton class="h-40 rounded-lg" />
				</div>
			{:else}
				<div class="space-y-4">
					<div class="rounded-lg border bg-muted/20 p-4">
						<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
							<div>
								<p class="text-sm font-medium">{securityStatus?.user.email}</p>
								<p class="text-sm text-muted-foreground">
									Active methods:
									{securityStatus?.methods.length
										? securityStatus.methods.join(', ')
										: 'none configured'}
								</p>
							</div>
							<p class="text-sm text-muted-foreground">
								Passkeys: {passkeyCount}
							</p>
						</div>
					</div>

					<div class="rounded-lg border p-4">
						<div class="mb-4 flex items-center justify-between gap-3">
							<div>
								<h3 class="flex items-center gap-2 font-medium">
									<MonitorIcon class="h-4 w-4 text-muted-foreground" />
									Active Sessions
								</h3>
								<p class="mt-1 text-sm text-muted-foreground">
									Review signed-in browsers and revoke access without changing your password.
								</p>
							</div>
							<Button
								variant="outline"
								size="sm"
								onclick={loadAuthSessions}
								disabled={authSessionsLoading}
							>
								{#if authSessionsLoading}
									<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
								{/if}
								Refresh
							</Button>
						</div>

						{#if authSessionsError}
							<div
								class="mb-3 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
							>
								{authSessionsError}
							</div>
						{/if}

						{#if authSessionsLoading}
							<div class="space-y-2">
								<Skeleton class="h-16 rounded-md" />
								<Skeleton class="h-16 rounded-md" />
							</div>
						{:else if authSessions.length === 0}
							<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
								No active web sessions found.
							</p>
						{:else}
							<div class="space-y-2" data-testid="auth-session-list">
								{#each authSessions as session (session.id)}
									<div
										class="flex flex-col gap-3 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
										data-testid="auth-session-row"
									>
										<div class="min-w-0">
											<div class="flex flex-wrap items-center gap-2">
												<p class="text-sm font-medium break-words">
													{formatSessionUserAgent(session.user_agent)}
												</p>
												{#if session.current}
													<span class="rounded-full bg-primary/10 px-2 py-0.5 text-xs text-primary">
														Current
													</span>
												{/if}
											</div>
											<p class="mt-1 text-xs text-muted-foreground">
												{session.ip_address || 'Unknown IP'} · Last used
												{formatSessionTime(session.last_used_at)} · Expires
												{formatSessionTime(session.expires_at)}
											</p>
										</div>
										<Button
											variant="ghost"
											size="sm"
											class="self-start text-destructive hover:text-destructive sm:self-center"
											onclick={() => revokeAuthSession(session)}
											disabled={Boolean(authSessionBusyID)}
										>
											{#if authSessionBusyID === session.id}
												<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
											{:else}
												<LogOutIcon class="mr-2 h-4 w-4" />
											{/if}
											{session.current ? 'Sign out' : 'Revoke'}
										</Button>
									</div>
								{/each}
							</div>
						{/if}
					</div>

					<div class="grid gap-4 lg:grid-cols-2">
						<div class="rounded-lg border p-4">
							<div class="mb-3 flex items-center gap-2">
								<SmartphoneIcon class="h-4 w-4 text-muted-foreground" />
								<h3 class="font-medium">Authenticator App</h3>
							</div>
							<p class="mb-4 text-sm text-muted-foreground">
								Scan a QR code in Authy, 1Password, Google Authenticator, or any standard TOTP app.
							</p>

							{#if securityStatus?.totp_enabled}
								<div class="space-y-3">
									<div class="rounded-md bg-emerald-500/10 px-3 py-2 text-sm text-emerald-700">
										Authenticator app is enabled.
									</div>
									<div class="space-y-2">
										<Label for="disable-password">Current password</Label>
										<Input
											id="disable-password"
											type="password"
											bind:value={currentPassword}
											placeholder="Required to disable"
										/>
									</div>
									<Button
										variant="outline"
										onclick={disableTOTP}
										disabled={securityBusy || !currentPassword.trim()}
									>
										Disable Authenticator App
									</Button>
								</div>
							{:else}
								<div class="space-y-3">
									<div class="space-y-2">
										<Label for="totp-password">Current password</Label>
										<Input
											id="totp-password"
											type="password"
											bind:value={currentPassword}
											placeholder="Required to start setup"
										/>
									</div>
									<Button
										onclick={startTOTPSetup}
										disabled={securityBusy || !currentPassword.trim()}
									>
										Start Authenticator Setup
									</Button>

									{#if totpSetupChallengeId}
										<div class="space-y-3 rounded-lg border bg-muted/20 p-4">
											<img
												src={totpQRCodeDataURL}
												alt="TOTP QR code"
												class="mx-auto h-48 w-48 rounded-lg border bg-white p-2"
											/>
											<div class="space-y-1">
												<p class="text-sm font-medium">Manual entry key</p>
												<p class="font-mono text-xs break-all text-muted-foreground">
													{totpManualEntryKey}
												</p>
											</div>
											<div class="space-y-2">
												<Label for="totp-code">Enter the 6-digit code from your app</Label>
												<Input
													id="totp-code"
													bind:value={totpCode}
													inputmode="numeric"
													autocomplete="one-time-code"
													maxlength={6}
													placeholder="123456"
												/>
											</div>
											<Button
												onclick={confirmTOTPSetup}
												disabled={securityBusy || totpCode.trim().length !== 6}
											>
												Confirm Authenticator App
											</Button>
										</div>
									{/if}
								</div>
							{/if}
						</div>

						<div class="rounded-lg border p-4">
							<div class="mb-3 flex items-center gap-2">
								<KeyRoundIcon class="h-4 w-4 text-muted-foreground" />
								<h3 class="font-medium">Passkeys</h3>
							</div>
							<p class="mb-4 text-sm text-muted-foreground">
								Add device-backed passkeys as a second factor for faster sign-ins.
							</p>

							<div class="space-y-3">
								<div class="space-y-2">
									<Label for="passkey-password">Current password</Label>
									<Input
										id="passkey-password"
										type="password"
										bind:value={currentPassword}
										placeholder="Required to add or remove passkeys"
									/>
								</div>
								<div class="space-y-2">
									<Label for="passkey-name">Passkey name</Label>
									<Input
										id="passkey-name"
										bind:value={newPasskeyName}
										placeholder="MacBook, iPhone, YubiKey"
									/>
								</div>
								<Button onclick={addPasskey} disabled={securityBusy || !currentPassword.trim()}>
									Add Passkey
								</Button>
							</div>

							<div class="mt-4 space-y-2">
								{#if securityStatus?.passkeys.length}
									{#each securityStatus.passkeys as passkey (passkey.id)}
										<div class="flex items-center justify-between rounded-md border px-3 py-2">
											<div>
												<p class="text-sm font-medium">{passkey.name}</p>
												<p class="text-xs text-muted-foreground">
													{#if passkey.last_used_at && passkey.last_used_at !== '0001-01-01T00:00:00Z'}
														Last used {new Date(passkey.last_used_at).toLocaleString()}
													{:else}
														Added {new Date(passkey.created_at).toLocaleString()}
													{/if}
												</p>
											</div>
											<Button
												variant="ghost"
												size="sm"
												class="text-destructive hover:text-destructive"
												onclick={() => removePasskey(passkey.id)}
												disabled={securityBusy || !currentPassword.trim()}
											>
												Remove
											</Button>
										</div>
									{/each}
								{:else}
									<p class="text-sm text-muted-foreground">No passkeys added yet.</p>
								{/if}
							</div>
						</div>
					</div>

					{#if securityError}
						<div
							class="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
						>
							{securityError}
						</div>
					{/if}
				</div>
			{/if}
		</section>

		<section id="tokens" class="scroll-mt-24 rounded-lg border p-6">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
				<TerminalIcon class="h-5 w-5 text-muted-foreground" />
				CLI Devices & API Tokens
			</h2>
			<p class="mb-4 text-sm text-muted-foreground">
				Create dedicated tokens for ChatGPT, Claude, the MCP server, the OpenPost CLI, CI, cron, and
				other automation. Revoke any token here without changing your password.
			</p>

			<div class="mb-4 grid gap-3 lg:grid-cols-[1fr_240px_240px_auto]">
				<div class="space-y-2">
					<Label for="api-token-name">New token name</Label>
					<Input
						id="api-token-name"
						bind:value={apiTokenName}
						placeholder="ChatGPT App, MacBook CLI, GitHub CI"
					/>
				</div>
				<div class="space-y-2">
					<Label for="api-token-scope">Token scope</Label>
					<Select.Root
						type="single"
						value={apiTokenScope}
						onValueChange={(value) => value && (apiTokenScope = value)}
					>
						<Select.Trigger id="api-token-scope" data-testid="api-token-scope" class="w-full">
							{selectedAPITokenScope.label}
						</Select.Trigger>
						<Select.Content>
							{#each apiTokenScopeOptions as option (option.value)}
								<Select.Item value={option.value}>
									<div class="flex flex-col gap-0.5 text-left">
										<span>{option.label}</span>
										<span class="text-xs text-muted-foreground">{option.description}</span>
									</div>
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
				<div class="space-y-2">
					<Label for="api-token-workspace">Access boundary</Label>
					<Select.Root
						type="single"
						value={apiTokenWorkspaceScope}
						onValueChange={(value) => value && (apiTokenWorkspaceScope = value)}
					>
						<Select.Trigger id="api-token-workspace" class="w-full">
							{selectedAPITokenWorkspaceScope.label}
						</Select.Trigger>
						<Select.Content>
							{#each apiTokenWorkspaceOptions as option (option.value)}
								<Select.Item value={option.value}>
									<div class="flex flex-col gap-0.5 text-left">
										<span>{option.label}</span>
										<span class="text-xs text-muted-foreground">{option.description}</span>
									</div>
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
				<div class="flex items-end">
					<Button
						onclick={createAPIToken}
						disabled={apiTokenBusy ||
							(apiTokenWorkspaceScope === 'current' && !workspaceCtx.currentWorkspace)}
					>
						{#if apiTokenBusy}
							<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
						{/if}
						Create Token
					</Button>
				</div>
			</div>

			{#if createdAPIToken}
				<div
					class="mb-4 rounded-lg border border-amber-300/50 bg-amber-50 p-4 text-sm text-amber-950"
				>
					<p class="font-medium">Copy this token now. It will not be shown again.</p>
					<p class="mt-2 font-mono text-xs break-all">{createdAPIToken}</p>
				</div>
			{/if}

			{#if apiTokensLoading}
				<div class="space-y-2">
					<Skeleton class="h-14 rounded-md" />
					<Skeleton class="h-14 rounded-md" />
				</div>
			{:else if apiTokens.length === 0}
				<p class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground">
					No API tokens or CLI devices are currently authorized.
				</p>
			{:else}
				<div class="space-y-2">
					{#each apiTokens as token (token.id)}
						<div
							class="flex flex-col gap-3 rounded-md border px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
						>
							<div>
								<p class="text-sm font-medium">{token.name}</p>
								<p class="text-xs text-muted-foreground">
									Prefix <span class="font-mono">{token.token_prefix}</span> · {token.scope} · Created
									{new Date(token.created_at).toLocaleString()}
									{#if token.workspace_id}
										· Workspace <span class="font-mono">{token.workspace_id}</span>
									{:else}
										· All workspaces
									{/if}
									{#if token.last_used_at}
										· Last used {new Date(token.last_used_at).toLocaleString()}
									{/if}
								</p>
							</div>
							<Button
								variant="ghost"
								size="sm"
								class="text-destructive hover:text-destructive"
								onclick={() => revokeAPIToken(token.id)}
								disabled={apiTokenBusy}
							>
								Revoke
							</Button>
						</div>
					{/each}
				</div>
			{/if}

			<div class="mt-6 border-t pt-6">
				<div class="mb-4 flex items-center justify-between gap-3">
					<div>
						<h3 class="flex items-center gap-2 text-sm font-semibold">
							<ActivityIcon class="h-4 w-4 text-muted-foreground" />
							Recent MCP Activity
						</h3>
						<p class="mt-1 text-sm text-muted-foreground">
							Recent tool calls from ChatGPT, Claude, the CLI proxy, and other MCP clients.
						</p>
					</div>
					<Button
						variant="outline"
						size="sm"
						onclick={loadMCPActivity}
						disabled={mcpActivityLoading}
					>
						{#if mcpActivityLoading}
							<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
						{/if}
						Refresh
					</Button>
				</div>

				{#if mcpActivityError}
					<div
						data-testid="mcp-activity-error"
						class="mb-3 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
					>
						{mcpActivityError}
					</div>
				{/if}

				{#if mcpActivityLoading}
					<div class="space-y-2">
						<Skeleton class="h-16 rounded-md" />
						<Skeleton class="h-16 rounded-md" />
					</div>
				{:else if mcpActivity.length === 0}
					<p
						data-testid="mcp-activity-empty"
						class="rounded-md border bg-muted/20 p-4 text-sm text-muted-foreground"
					>
						No MCP tool calls have been recorded yet.
					</p>
				{:else}
					<div data-testid="mcp-activity-list" class="space-y-2">
						{#each mcpActivity as call (call.id)}
							<div class="rounded-md border px-3 py-3">
								<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
									<div class="min-w-0">
										<p class="truncate text-sm font-medium">{call.tool_name}</p>
										<p class="mt-1 text-xs text-muted-foreground">
											{new Date(call.created_at).toLocaleString()} · {call.duration_ms} ms
											{#if call.workspace_id}
												· Workspace <span class="font-mono">{call.workspace_id}</span>
											{/if}
										</p>
										{#if call.client_name || call.client_scope}
											<p class="mt-1 truncate text-xs text-muted-foreground">
												Client {call.client_name || call.client_scope}
												{#if call.client_token_prefix}
													· <span class="font-mono">{call.client_token_prefix}</span>
												{/if}
											</p>
										{/if}
									</div>
									<span
										class={[
											'inline-flex w-fit items-center rounded-full border px-2 py-0.5 text-xs font-medium',
											call.status === 'success'
												? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700'
												: 'border-destructive/30 bg-destructive/10 text-destructive'
										]}
									>
										{call.status}
									</span>
								</div>
								{#if call.error_message}
									<p class="mt-2 text-xs text-destructive">{call.error_message}</p>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</section>

		<section id="date-time" class="scroll-mt-24 space-y-4">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
				<ClockIcon class="h-5 w-5 text-muted-foreground" />
				Date & Time
			</h2>
			<div class="grid gap-4 sm:grid-cols-2">
				<div class="space-y-2">
					<label class="text-sm font-medium" for="timezone-select">Timezone</label>
					<Select.Root
						type="single"
						value={workspaceCtx.settings.timezone}
						onValueChange={handleTimezoneChange}
					>
						<Select.Trigger id="timezone-select" class="w-full">
							{getTimezoneLabel(workspaceCtx.settings.timezone)}
						</Select.Trigger>
						<Select.Content class="max-h-80 overflow-y-auto">
							{#each Object.entries(groupedTimezones) as [group, tzs] (group)}
								<Select.Group>
									<Select.GroupHeading class="text-xs">{group}</Select.GroupHeading>
									{#each tzs as tz (tz.value)}
										<Select.Item value={tz.value}>{tz.label}</Select.Item>
									{/each}
								</Select.Group>
							{/each}
						</Select.Content>
					</Select.Root>
					<p class="text-sm text-muted-foreground">
						Detected from your browser the first time a workspace loads, then saved here.
					</p>
				</div>

				<div class="space-y-2">
					<label class="text-sm font-medium" for="week-start-select">Week Starts On</label>
					<Select.Root
						type="single"
						value={String(workspaceCtx.settings.week_start)}
						onValueChange={(v) => handleWeekStartChange(Number(v))}
					>
						<Select.Trigger id="week-start-select" class="w-full">
							{workspaceCtx.settings.week_start === 0 ? 'Sunday' : 'Monday'}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="0">Sunday</Select.Item>
							<Select.Item value="1">Monday</Select.Item>
						</Select.Content>
					</Select.Root>
					<p class="text-sm text-muted-foreground">
						Defaulted from your locale on first load and used for calendar layout.
					</p>
				</div>
			</div>
		</section>

		<section id="media-cleanup" class="scroll-mt-24 space-y-4">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
				<ImageIcon class="h-5 w-5 text-muted-foreground" />
				Media Cleanup
			</h2>
			<div class="space-y-2">
				<label class="text-sm font-medium" for="cleanup-select">Auto-delete unused media</label>
				<Select.Root
					type="single"
					value={String(workspaceCtx.settings.media_cleanup_days)}
					onValueChange={(v) => handleCleanupDaysChange(Number(v))}
				>
					<Select.Trigger id="cleanup-select" class="w-full">
						{cleanupDaysOptions.find((o) => o.value === workspaceCtx.settings.media_cleanup_days)
							?.label || 'Disabled'}
					</Select.Trigger>
					<Select.Content>
						{#each cleanupDaysOptions as option (option.value)}
							<Select.Item value={String(option.value)}>{option.label}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
				<p class="text-sm text-muted-foreground">
					Automatically delete unused, non-favorited media after this period. Favorited media is
					always kept.
				</p>
			</div>
		</section>

		<section id="posting-schedule" class="scroll-mt-24 rounded-lg border p-6">
			<div class="mb-4 flex items-center justify-between">
				<h2 class="flex items-center gap-2 text-lg font-semibold">
					<CalendarIcon class="h-5 w-5 text-muted-foreground" />
					Posting Schedule
				</h2>
				<Button
					onclick={() => (showSuggestSchedule = !showSuggestSchedule)}
					variant="outline"
					size="sm"
				>
					<SparklesIcon class="mr-2 h-4 w-4" />
					Suggest Weekly Pattern
				</Button>
			</div>
			<p class="mb-4 text-sm text-muted-foreground">
				Define reusable posting times in your workspace timezone. Toggle each weekday checkbox to
				decide when that time is active. The "Suggest Time" action will use these slots first.
			</p>

			<div class="mb-4 rounded-xl border bg-muted/20 p-4">
				<div class="grid gap-4 lg:grid-cols-[180px_1fr_auto]">
					<div class="space-y-2">
						<label class="text-sm font-medium" for="new-time">Add time row</label>
						<Input id="new-time" bind:value={newTimeInput} type="time" step="900" />
					</div>
					<div class="space-y-2">
						<span class="text-sm font-medium">Active days</span>
						<div class="flex flex-wrap gap-3">
							{#each dayOrder as dayIndex (dayIndex)}
								<label
									class="flex items-center gap-2 rounded-md border bg-background px-3 py-2 text-sm"
								>
									<Checkbox
										checked={newTimeDays.includes(dayIndex)}
										onCheckedChange={() => toggleNewDay(dayIndex)}
									/>
									<span>{dayShortNames[dayIndex]}</span>
								</label>
							{/each}
						</div>
					</div>
					<div class="flex items-end">
						<Button onclick={addTimeRow} class="w-full lg:w-auto">
							<PlusIcon class="mr-2 h-4 w-4" />
							Add Time
						</Button>
					</div>
				</div>
				{#if newTimeError}
					<p class="mt-3 text-xs text-destructive">{newTimeError}</p>
				{:else}
					<p class="mt-3 text-xs text-muted-foreground">
						New rows are created in {getTimezoneLabel(workspaceCtx.settings.timezone)}.
					</p>
				{/if}
			</div>

			{#if showSuggestSchedule}
				<div class="mb-4 rounded-xl border bg-background p-4">
					<div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
						<div class="space-y-2">
							<label class="text-sm font-medium" for="posts-per-day">Suggested posts per day</label>
							<Select.Root
								type="single"
								value={String(suggestedPostsPerDay)}
								onValueChange={(v) => (suggestedPostsPerDay = Number(v))}
							>
								<Select.Trigger id="posts-per-day" class="w-28">
									{suggestedPostsPerDay}
								</Select.Trigger>
								<Select.Content class="max-h-60 overflow-y-auto">
									{#each Array.from({ length: 10 }, (_, i) => i + 1) as n (n)}
										<Select.Item value={String(n)}>{n}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<div class="flex gap-2">
							<Button onclick={() => (showSuggestSchedule = false)} variant="outline" size="sm"
								>Cancel</Button
							>
							<Button onclick={generateSuggestedSchedule} size="sm" disabled={generatingSchedule}>
								{#if generatingSchedule}
									<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
								{/if}
								Generate
							</Button>
						</div>
					</div>
				</div>
			{/if}

			{#if loadingSchedules}
				<div class="space-y-2">
					<Skeleton class="h-14 rounded-md" />
					<Skeleton class="h-14 rounded-md" />
					<Skeleton class="h-14 rounded-md" />
				</div>
			{:else}
				<div class="overflow-hidden rounded-xl border">
					<div class="grid grid-cols-[120px_repeat(7,minmax(56px,1fr))_52px] border-b bg-muted/30">
						<div
							class="px-4 py-3 text-xs font-semibold tracking-wide text-muted-foreground uppercase"
						>
							Time
						</div>
						{#each dayOrder as dayIndex (dayIndex)}
							<div
								class="px-2 py-3 text-center text-xs font-semibold tracking-wide text-muted-foreground uppercase"
							>
								{dayShortNames[dayIndex]}
							</div>
						{/each}
						<div class="px-2 py-3"></div>
					</div>

					{#if scheduleRows.length === 0}
						<div class="px-4 py-10 text-center text-sm text-muted-foreground">
							No posting times yet. Add a row above or generate a suggested weekly pattern.
						</div>
					{:else}
						{#each scheduleRows as row (row.key)}
							<div
								class="grid grid-cols-[120px_repeat(7,minmax(56px,1fr))_52px] border-b last:border-b-0"
							>
								<div class="px-4 py-3">
									<div class="font-medium">{formatTime(row.local_hour, row.local_minute)}</div>
									{#if row.label}
										<div class="text-xs text-muted-foreground">{row.label}</div>
									{/if}
								</div>
								{#each dayOrder as dayIndex (`${row.key}-${dayIndex}`)}
									<div class="flex items-center justify-center px-2 py-3">
										<Checkbox
											checked={Boolean(row.days[dayIndex])}
											onCheckedChange={() => toggleScheduleCell(row, dayIndex)}
											aria-label={`Toggle ${dayNames[dayIndex]} ${formatTime(row.local_hour, row.local_minute)}`}
										/>
									</div>
								{/each}
								<div class="flex items-center justify-center px-2 py-3">
									<Button
										variant="ghost"
										size="icon"
										class="h-8 w-8"
										onclick={() => removeTimeRow(row)}
										aria-label={`Remove ${formatTime(row.local_hour, row.local_minute)} row`}
									>
										<TrashIcon class="h-4 w-4" />
									</Button>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			{/if}
		</section>

		<section id="natural-posting" class="scroll-mt-24 space-y-4">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
				<ClockIcon class="h-5 w-5 text-muted-foreground" />
				Natural Posting
			</h2>
			<div class="space-y-4">
				<p class="text-sm text-muted-foreground">
					Add a small random delay to scheduled posts so they don't all go out at exactly the same
					minute. This makes your posting pattern look more natural.
				</p>
				<div class="space-y-2">
					<label class="text-sm font-medium" for="random-delay">Random delay range</label>
					<Select.Root
						type="single"
						value={String(workspaceCtx.settings.random_delay_minutes)}
						onValueChange={(v) => (workspaceCtx.settings.random_delay_minutes = Number(v))}
					>
						<Select.Trigger id="random-delay" class="w-full sm:w-64">
							{#if workspaceCtx.settings.random_delay_minutes === 0}
								No delay (exact time)
							{:else}
								±{workspaceCtx.settings.random_delay_minutes} minutes
							{/if}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="0">No delay (exact time)</Select.Item>
							<Select.Item value="5">±5 minutes</Select.Item>
							<Select.Item value="10">±10 minutes</Select.Item>
							<Select.Item value="15">±15 minutes</Select.Item>
							<Select.Item value="30">±30 minutes</Select.Item>
							<Select.Item value="45">±45 minutes</Select.Item>
							<Select.Item value="60">±1 hour</Select.Item>
						</Select.Content>
					</Select.Root>
				</div>
				<div class="space-y-2">
					<label class="text-sm font-medium" for="draft-gap">Draft spillover gap</label>
					<Input
						id="draft-gap"
						type="text"
						value={draftGapInput}
						oninput={(e) => handleDraftGapChange((e.target as HTMLInputElement).value)}
						placeholder="e.g. 45m, 2h, 0"
						class={draftGapError ? 'border-destructive' : ''}
					/>
					{#if draftGapError}
						<p class="text-xs text-destructive">{draftGapError}</p>
					{:else}
						<p class="text-xs text-muted-foreground">
							When a day has no unused schedule slots left, "Suggest Time" will place the next post
							at least {workspaceCtx.settings.draft_gap_minutes} minutes after the latest scheduled post
							that day. Use `0` to disable the spillover rule.
						</p>
					{/if}
				</div>
			</div>
		</section>

		<section id="slot-defaults" class="scroll-mt-24 space-y-4">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
				<ClockIcon class="h-5 w-5 text-muted-foreground" />
				Time Slot Defaults
			</h2>
			<div class="space-y-4">
				<p class="text-sm text-muted-foreground">
					Configure the default time range and interval shown in the compose page scheduler.
				</p>
				<div class="grid gap-4 sm:grid-cols-3">
					<div class="space-y-2">
						<label class="text-sm font-medium" for="start-time">Start time</label>
						<Select.Root
							type="single"
							value={String(workspaceCtx.settings.slot_start_hour)}
							onValueChange={(v) => (workspaceCtx.settings.slot_start_hour = Number(v))}
						>
							<Select.Trigger id="start-time" class="w-full">
								{workspaceCtx.settings.slot_start_hour.toString().padStart(2, '0')}:00
							</Select.Trigger>
							<Select.Content class="max-h-60 overflow-y-auto">
								{#each Array.from({ length: 24 }, (_, i) => i) as hour (hour)}
									<Select.Item value={String(hour)}
										>{hour.toString().padStart(2, '0')}:00</Select.Item
									>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="space-y-2">
						<label class="text-sm font-medium" for="end-time">End time</label>
						<Select.Root
							type="single"
							value={String(workspaceCtx.settings.slot_end_hour)}
							onValueChange={(v) => (workspaceCtx.settings.slot_end_hour = Number(v))}
						>
							<Select.Trigger id="end-time" class="w-full">
								{workspaceCtx.settings.slot_end_hour.toString().padStart(2, '0')}:00
							</Select.Trigger>
							<Select.Content class="max-h-60 overflow-y-auto">
								{#each Array.from({ length: 24 }, (_, i) => i) as hour (hour)}
									<Select.Item value={String(hour)}
										>{hour.toString().padStart(2, '0')}:00</Select.Item
									>
								{/each}
							</Select.Content>
						</Select.Root>
					</div>
					<div class="space-y-2">
						<label class="text-sm font-medium" for="interval">Interval</label>
						<input
							id="interval"
							type="text"
							value={intervalInput}
							oninput={(e) => handleIntervalChange((e.target as HTMLInputElement).value)}
							placeholder="e.g. 15m, 30 min, 1h"
							class="h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm {intervalError
								? 'border-destructive'
								: ''}"
						/>
						{#if intervalError}
							<p class="text-xs text-destructive">{intervalError}</p>
						{:else}
							<p class="text-xs text-muted-foreground">
								Current: {workspaceCtx.settings.slot_interval_minutes} minutes
							</p>
						{/if}
					</div>
				</div>
			</div>
		</section>

		<div class="flex justify-end">
			<Button onclick={saveSettings} disabled={saving}>
				{#if saving}
					<LoaderIcon class="mr-2 h-4 w-4 animate-spin" />
				{:else}
					<SaveIcon class="mr-2 h-4 w-4" />
				{/if}
				Save Changes
			</Button>
		</div>
	</div>
</PageContainer>
