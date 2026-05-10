<script lang="ts">
	import PlatformIcon from './platform-icon.svelte';
	import PlatformPreview from './platform-preview.svelte';

	export interface PreviewPost {
		key: string;
		content: string;
		mediaIds: string[];
	}

	export interface PreviewGroup {
		key: string;
		platformKey: string;
		platformName: string;
		username: string;
		displayName: string;
		isUnsynced: boolean;
		posts: PreviewPost[];
	}

	interface Props {
		groups: PreviewGroup[];
	}

	let { groups }: Props = $props();
</script>

<div class="space-y-5">
	{#each groups as group (group.key)}
		<div>
			<div class="mb-1.5 flex items-center gap-1.5 text-xs text-muted-foreground">
				<PlatformIcon platform={group.platformKey} class="h-3 w-3" />
				<span>{group.platformName}</span>
				{#if group.username !== 'username'}
					<span class="text-muted-foreground/60">@{group.username}</span>
				{/if}
			</div>

			<div class={group.posts.length > 1 ? 'space-y-3' : ''}>
				{#each group.posts as post (post.key)}
					{#key `${group.key}:${post.key}:${group.platformKey}`}
						<PlatformPreview
							platform={group.platformKey}
							content={post.content}
							mediaIds={post.mediaIds}
							username={group.username}
							displayName={group.displayName}
							isUnsynced={group.isUnsynced}
						/>
					{/key}
				{/each}
			</div>
		</div>
	{/each}
</div>
