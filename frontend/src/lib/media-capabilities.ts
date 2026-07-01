import { getPlatformKey } from '$lib/utils';

export type MediaCapabilitySeverity = 'error' | 'warning';

export interface MediaCapabilityItem {
	id: string;
	mimeType: string;
	size?: number;
}

export interface MediaCapabilityIssue {
	provider: string;
	mediaId?: string;
	severity: MediaCapabilitySeverity;
	message: string;
}

const videoTypeMP4 = 'video/mp4';
const blueskyVideoLimitBytes = 100 * 1024 * 1024;

export function mediaCapabilityItemsFromIds(
	mediaIds: string[],
	mimeTypes: ReadonlyMap<string, string>,
	sizes?: ReadonlyMap<string, number>
): MediaCapabilityItem[] {
	return mediaIds
		.map((id) => ({
			id,
			mimeType: mimeTypes.get(id) ?? '',
			size: sizes?.get(id)
		}))
		.filter((item) => item.id);
}

export function providerMediaWarningMessages(
	provider: string,
	media: MediaCapabilityItem[]
): string[] {
	return validateProviderMedia(provider, media).map((issue) => issue.message);
}

export function validateProviderMedia(
	provider: string,
	media: MediaCapabilityItem[]
): MediaCapabilityIssue[] {
	const key = getPlatformKey(provider);

	switch (key) {
		case 'x':
			return validateXMedia(media);
		case 'mastodon':
			return validateMastodonMedia(media);
		case 'bluesky':
			return validateBlueskyMedia(media);
		case 'linkedin':
			return validateLinkedInMedia(media);
		case 'threads':
			return validateThreadsMedia(media);
		case 'facebook':
			return validateFacebookMedia(media);
		case 'instagram':
			return validateInstagramMedia(media);
		case 'tiktok':
			return validateTikTokMedia(media);
		case 'youtube':
			return validateYouTubeMedia(media);
		default:
			return [];
	}
}

export function videoProviderSupportLabel(mimeType: string): string | null {
	return isVideoMime(mimeType) ? 'Provider-limited' : null;
}

export function videoProviderSupportDetail(mimeType: string): string | null {
	if (!isVideoMime(mimeType)) return null;
	const normalized = normalizeMime(mimeType);
	if (normalized === videoTypeMP4) {
		return 'MP4 is the safest video format. YouTube and TikTok require one video; X and Bluesky cannot mix video with images.';
	}
	if (normalized === 'video/quicktime') {
		return 'MOV works for Threads, but most providers are safest with MP4.';
	}
	if (normalized === 'video/webm') {
		return 'WebM is mainly Mastodon-friendly; use MP4 for wider provider support.';
	}
	return 'Video publishing depends on the selected provider. MP4 is the safest format.';
}

function validateXMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length === 0) return [];

	const videos = media.filter((item) => isVideoMime(item.mimeType));
	if (videos.length === 0) {
		if (media.length > 4) {
			return [issue('x', 'error', 'X supports up to 4 images per post.')];
		}
		return [];
	}

	if (videos.length > 1 || media.length > 1) {
		return [issue('x', 'error', 'X supports one video per post and cannot mix video with images.')];
	}
	if (normalizeMime(videos[0].mimeType) !== videoTypeMP4) {
		return [
			issue('x', 'warning', 'X video publishing is most reliable with MP4 video.', videos[0].id)
		];
	}
	return [];
}

function validateMastodonMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length === 0) return [];
	if (media.length > 4) {
		return [issue('mastodon', 'error', 'Mastodon supports up to 4 media attachments per post.')];
	}
	for (const item of media) {
		if (isVideoMime(item.mimeType) && !isMastodonLikelyVideoMime(item.mimeType)) {
			return [
				issue(
					'mastodon',
					'warning',
					'Mastodon video support depends on the instance; MP4 and WebM are the safest formats.',
					item.id
				)
			];
		}
	}
	return [];
}

function validateBlueskyMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length === 0) return [];

	let hasVideo = false;
	for (const item of media) {
		if (!isVideoMime(item.mimeType)) continue;
		if (hasVideo) {
			return [issue('bluesky', 'error', 'Bluesky supports only 1 video per post.', item.id)];
		}
		hasVideo = true;
		if (normalizeMime(item.mimeType) !== videoTypeMP4) {
			return [issue('bluesky', 'error', 'Bluesky supports MP4 video only.', item.id)];
		}
		if (typeof item.size === 'number' && item.size > blueskyVideoLimitBytes) {
			return [issue('bluesky', 'error', 'Bluesky video must be under 100MB.', item.id)];
		}
	}

	if (hasVideo) {
		if (media.length > 1) {
			return [
				issue('bluesky', 'error', 'Bluesky does not support mixing video and images in one post.')
			];
		}
		return [];
	}

	if (media.length > 4) {
		return [issue('bluesky', 'error', 'Bluesky supports up to 4 images per post.')];
	}
	return [];
}

function validateLinkedInMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length === 0) return [];
	if (media.length > 1) {
		return [
			issue(
				'linkedin',
				'warning',
				'OpenPost currently publishes only the first LinkedIn attachment.'
			)
		];
	}
	if (isVideoMime(media[0].mimeType) && normalizeMime(media[0].mimeType) !== videoTypeMP4) {
		return [
			issue(
				'linkedin',
				'warning',
				'LinkedIn video publishing is most reliable with MP4 video.',
				media[0].id
			)
		];
	}
	return [];
}

function validateThreadsMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length === 0) return [];

	const issues: MediaCapabilityIssue[] = [];
	for (const item of media) {
		if (isVideoMime(item.mimeType) && !isThreadsVideoMime(item.mimeType)) {
			issues.push(issue('threads', 'error', 'Threads supports MP4 or MOV video.', item.id));
		}
	}
	if (media.length > 1) {
		issues.push(
			issue('threads', 'warning', 'OpenPost currently publishes only the first Threads attachment.')
		);
	}
	return issues;
}

function validateFacebookMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length <= 1) return [];
	return [
		issue(
			'facebook',
			'error',
			'Facebook publishing currently supports at most one media attachment.'
		)
	];
}

function validateInstagramMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length !== 1) {
		return [
			issue(
				'instagram',
				'error',
				'Instagram publishing currently requires exactly one image or video attachment.'
			)
		];
	}
	if (!isVideoMime(media[0].mimeType) && !isImageMime(media[0].mimeType)) {
		return [
			issue(
				'instagram',
				'error',
				'Instagram publishing supports image or video attachments only.',
				media[0].id
			)
		];
	}
	return [];
}

function validateTikTokMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length !== 1) {
		return [
			issue('tiktok', 'error', 'TikTok publishing currently requires exactly one video attachment.')
		];
	}
	if (!isVideoMime(media[0].mimeType)) {
		return [
			issue(
				'tiktok',
				'error',
				'TikTok publishing currently supports video attachments only.',
				media[0].id
			)
		];
	}
	return [];
}

function validateYouTubeMedia(media: MediaCapabilityItem[]): MediaCapabilityIssue[] {
	if (media.length !== 1) {
		return [
			issue(
				'youtube',
				'error',
				'YouTube publishing currently requires exactly one video attachment.'
			)
		];
	}
	if (!isVideoMime(media[0].mimeType)) {
		return [
			issue('youtube', 'error', 'YouTube publishing supports video attachments only.', media[0].id)
		];
	}
	return [];
}

function issue(
	provider: string,
	severity: MediaCapabilitySeverity,
	message: string,
	mediaId?: string
): MediaCapabilityIssue {
	return { provider, severity, message, ...(mediaId ? { mediaId } : {}) };
}

function isVideoMime(mimeType: string): boolean {
	return normalizeMime(mimeType).startsWith('video/');
}

function isImageMime(mimeType: string): boolean {
	return normalizeMime(mimeType).startsWith('image/');
}

function isThreadsVideoMime(mimeType: string): boolean {
	const normalized = normalizeMime(mimeType);
	return normalized === videoTypeMP4 || normalized === 'video/quicktime';
}

function isMastodonLikelyVideoMime(mimeType: string): boolean {
	const normalized = normalizeMime(mimeType);
	return normalized === videoTypeMP4 || normalized === 'video/webm' || normalized === 'image/gif';
}

function normalizeMime(mimeType: string): string {
	return mimeType.trim().toLowerCase();
}
