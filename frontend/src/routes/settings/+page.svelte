<script lang="ts">
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
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { client } from '$lib/api/client';
	import { getLocaleTag } from '$lib/i18n';

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

	let securityStatus = $state<SecurityStatus | null>(null);
	let apiTokens = $state<APITokenSummary[]>([]);
	let apiTokensLoading = $state(true);
	let mcpActivity = $state.raw<MCPActivityItem[]>([]);
	let mcpActivityLoading = $state(true);
	let mcpActivityError = $state('');
	let apiTokenBusy = $state(false);
	let apiTokenName = $state('OpenPost MCP');
	let apiTokenScope = $state('mcp:full');
	let createdAPIToken = $state('');
	let billingBusyPlan = $state('');
	let billingPortalBusy = $state(false);
	let billingError = $state('');
	let billingStatusLoading = $state(false);
	let billingStatus = $state<BillingStatus | null>(null);

	interface APITokenSummary {
		id: string;
		name: string;
		token_prefix: string;
		scope: string;
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

	const authState = $derived($auth);
	const passkeyCount = $derived(securityStatus?.passkeys.length ?? 0);
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
	const billingMetricLabels: Record<string, string> = {
		scheduled_posts_monthly: 'Scheduled posts',
		published_posts_monthly: 'Published posts',
		media_bytes_uploaded_monthly: 'Uploaded media',
		provider_write_calls_monthly: 'Provider publish calls'
	};
	const currentBillingPlan = $derived(
		billingPlans.find((plan) => plan.id === billingStatus?.plan_id) ?? null
	);
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

	async function createAPIToken() {
		apiTokenBusy = true;
		securityError = '';
		createdAPIToken = '';
		const fallbackName = apiTokenScope === 'mcp:full' ? 'OpenPost MCP' : 'OpenPost CLI';
		try {
			const { data, error: err } = await (client as any).POST('/api-tokens', {
				body: { name: apiTokenName.trim() || fallbackName, scope: apiTokenScope }
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

	$effect(() => {
		if (workspaceCtx.currentWorkspace) {
			loadBillingStatus();
			loadSchedules();
		}
	});

	$effect(() => {
		if (authState.isAuthenticated) {
			loadSecurityStatus();
			loadAPITokens();
			loadMCPActivity();
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
		<section class="space-y-4">
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

		<section class="rounded-lg border p-6">
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

		<section class="rounded-lg border p-6">
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

		<section class="rounded-lg border p-6">
			<h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
				<TerminalIcon class="h-5 w-5 text-muted-foreground" />
				CLI Devices & API Tokens
			</h2>
			<p class="mb-4 text-sm text-muted-foreground">
				Create dedicated tokens for ChatGPT, Claude, the MCP server, the OpenPost CLI, CI, cron, and
				other automation. Revoke any token here without changing your password.
			</p>

			<div class="mb-4 grid gap-3 lg:grid-cols-[1fr_280px_auto]">
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
				<div class="flex items-end">
					<Button onclick={createAPIToken} disabled={apiTokenBusy}>
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

		<section class="space-y-4">
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

		<section class="space-y-4">
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

		<section class="rounded-lg border p-6">
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

		<section class="space-y-4">
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

		<section class="space-y-4">
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
