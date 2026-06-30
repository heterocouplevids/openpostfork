<script lang="ts">
	import x from '../../../../assets/logos/x.svg?raw';
	import mastodon from '../../../../assets/logos/mastodon.svg?raw';
	import threads from '../../../../assets/logos/threads.svg?raw';
	import bluesky from '../../../../assets/logos/bluesky.svg?raw';
	import linkedin from '../../../../assets/logos/linkedin.svg?raw';
	import { getPlatformKey } from '$lib/utils';
	import FacebookIcon from 'lucide-svelte/icons/facebook';
	import InstagramIcon from 'lucide-svelte/icons/instagram';
	import Music2Icon from 'lucide-svelte/icons/music-2';
	import YoutubeIcon from 'lucide-svelte/icons/youtube';

	interface Props {
		platform: string;
		class?: string;
	}

	let { platform, class: className = '' }: Props = $props();

	const svgs: Record<string, string> = { x, mastodon, threads, bluesky, linkedin };
	const platformKey = $derived(getPlatformKey(platform));

	const svg = $derived(
		svgs[platformKey] ? svgs[platformKey].replace('<svg ', `<svg class="${className}" `) : ''
	);
</script>

{#if svg}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html svg}
{:else if platformKey === 'facebook'}
	<FacebookIcon class={className} />
{:else if platformKey === 'instagram'}
	<InstagramIcon class={className} />
{:else if platformKey === 'tiktok'}
	<Music2Icon class={className} />
{:else if platformKey === 'youtube'}
	<YoutubeIcon class={className} />
{/if}
