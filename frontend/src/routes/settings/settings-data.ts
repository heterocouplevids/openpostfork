import type { components } from '$lib/api/types';

export const timezones = [
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

export const cleanupDaysOptions = [
	{ value: 0, label: 'Disabled' },
	{ value: 7, label: '7 days' },
	{ value: 14, label: '14 days' },
	{ value: 30, label: '30 days' },
	{ value: 60, label: '60 days' },
	{ value: 90, label: '90 days' },
	{ value: 180, label: '180 days' },
	{ value: 365, label: '1 year' }
];

export const providerAppOptions = [
	{
		value: 'mastodon',
		label: 'Mastodon',
		description: 'Primary use case: add app credentials for any Mastodon instance you support.',
		guideHref: 'https://docs.openpost.social/providers/mastodon'
	},
	{
		value: 'x',
		label: 'X / Twitter',
		description: 'Optional override for X account connections and publishing.',
		guideHref: 'https://docs.openpost.social/providers/x'
	},
	{
		value: 'linkedin',
		label: 'LinkedIn',
		description: 'Optional override for LinkedIn company and member posting.',
		guideHref: 'https://docs.openpost.social/providers/linkedin'
	},
	{
		value: 'threads',
		label: 'Threads',
		description: 'Optional Meta app override for Threads publishing.',
		guideHref: 'https://docs.openpost.social/providers/threads'
	},
	{
		value: 'facebook',
		label: 'Facebook Pages',
		description: 'Optional Meta app override for Facebook Page publishing.',
		guideHref: 'https://docs.openpost.social/providers/facebook'
	},
	{
		value: 'instagram',
		label: 'Instagram Business',
		description: 'Optional Meta app override for Instagram media publishing.',
		guideHref: 'https://docs.openpost.social/providers/instagram'
	},
	{
		value: 'youtube',
		label: 'YouTube',
		description: 'Optional Google OAuth override for YouTube channel uploads.',
		guideHref: 'https://docs.openpost.social/providers/youtube'
	},
	{
		value: 'tiktok',
		label: 'TikTok',
		description: 'Optional TikTok developer app override for video publishing.',
		guideHref: 'https://docs.openpost.social/providers/tiktok'
	}
];

export const inviteRoleOptions = [
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

export const apiTokenScopeOptions = [
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

export const billingPlans = [
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

export const billingMetricLabels: Record<string, string> = {
	scheduled_posts_monthly: 'Scheduled posts',
	published_posts_monthly: 'Published posts',
	media_bytes_uploaded_monthly: 'Uploaded media',
	provider_write_calls_monthly: 'Provider publish calls'
};

export const dayNames = [
	'Sunday',
	'Monday',
	'Tuesday',
	'Wednesday',
	'Thursday',
	'Friday',
	'Saturday'
];
export const dayShortNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

export interface PasskeySummary {
	id: string;
	name: string;
	created_at: string;
	last_used_at: string;
}

export interface SecurityStatus {
	user: {
		id: string;
		email: string;
		created_at: string;
	};
	totp_enabled: boolean;
	passkeys: PasskeySummary[];
	methods: string[];
}

export interface AuthSessionSummary {
	id: string;
	user_agent: string;
	ip_address: string;
	current: boolean;
	expires_at: string;
	last_used_at: string;
	created_at: string;
}

export interface APITokenSummary {
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

export interface MCPActivityItem {
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

export interface BillingStatus {
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

export interface TeamMember {
	user_id: string;
	email: string;
	role: string;
}

export interface WorkspaceInvitation {
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

export interface WorkspaceTeam {
	members: TeamMember[];
	invitations: WorkspaceInvitation[];
	current_seats: number;
}

export type ProviderApp = components['schemas']['ProviderAppResponse'];
export type SaveProviderAppInput = components['schemas']['SaveProviderAppInputBody'];

export interface ProviderAppForm {
	provider: string;
	name: string;
	client_id: string;
	client_secret: string;
	redirect_uri: string;
	instance_url: string;
	is_active: boolean;
}

export interface PostingSchedule {
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

export interface ScheduleRow {
	key: string;
	local_hour: number;
	local_minute: number;
	label: string;
	days: Record<number, PostingSchedule | undefined>;
}

export function emptyProviderAppForm(): ProviderAppForm {
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

export function getTimezoneLabel(value: string): string {
	return timezones.find((timezone) => timezone.value === value)?.label ?? value;
}
