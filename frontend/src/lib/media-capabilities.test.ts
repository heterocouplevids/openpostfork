import { describe, expect, it } from 'vitest';
import {
	mediaCapabilityItemsFromIds,
	providerMediaWarningMessages,
	validateProviderMedia,
	videoProviderSupportDetail,
	videoProviderSupportLabel
} from './media-capabilities';

describe('media-capabilities', () => {
	it('builds ordered capability items from selected media IDs', () => {
		const mimeTypes = new Map([
			['video-1', 'video/mp4'],
			['image-1', 'image/png']
		]);
		const sizes = new Map([
			['video-1', 123],
			['image-1', 456]
		]);

		expect(mediaCapabilityItemsFromIds(['video-1', 'image-1'], mimeTypes, sizes)).toEqual([
			{ id: 'video-1', mimeType: 'video/mp4', size: 123 },
			{ id: 'image-1', mimeType: 'image/png', size: 456 }
		]);
	});

	it('warns when video is mixed with images on X and Bluesky', () => {
		const media = [
			{ id: 'video-1', mimeType: 'video/mp4' },
			{ id: 'image-1', mimeType: 'image/png' }
		];

		expect(providerMediaWarningMessages('x', media)).toContain(
			'X supports one video per post and cannot mix video with images.'
		);
		expect(providerMediaWarningMessages('bluesky', media)).toContain(
			'Bluesky does not support mixing video and images in one post.'
		);
	});

	it('catches Bluesky video format and size limits when metadata is available', () => {
		expect(
			validateProviderMedia('bluesky', [
				{ id: 'video-1', mimeType: 'video/webm', size: 10 * 1024 * 1024 }
			])
		).toEqual([
			{
				provider: 'bluesky',
				mediaId: 'video-1',
				severity: 'error',
				message: 'Bluesky supports MP4 video only.'
			}
		]);

		expect(
			providerMediaWarningMessages('bluesky', [
				{ id: 'video-1', mimeType: 'video/mp4', size: 101 * 1024 * 1024 }
			])
		).toContain('Bluesky video must be under 100MB.');
	});

	it('warns when video-only providers receive an image', () => {
		const media = [{ id: 'image-1', mimeType: 'image/png' }];

		expect(providerMediaWarningMessages('tiktok', media)).toContain(
			'TikTok publishing currently supports video attachments only.'
		);
		expect(providerMediaWarningMessages('youtube', media)).toContain(
			'YouTube publishing supports video attachments only.'
		);
	});

	it('labels videos as provider-limited for the media library', () => {
		expect(videoProviderSupportLabel('video/mp4')).toBe('Provider-limited');
		expect(videoProviderSupportLabel('image/png')).toBeNull();
		expect(videoProviderSupportDetail('video/webm')).toContain('Mastodon-friendly');
	});
});
