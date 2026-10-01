<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon, { type IconName } from './Icon.svelte';

	type Tone = 'neutral' | 'gold' | 'info';

	interface Props {
		tone?: Tone;
		icon?: IconName;
		dot?: boolean;
		children: Snippet;
	}

	let { tone = 'neutral', icon, dot = false, children }: Props = $props();
</script>

<span class="badge badge--{tone}">
	{#if dot}<span class="dot"></span>{/if}
	{#if icon}<Icon name={icon} size={13} />{/if}
	{@render children()}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 24px;
		padding: 0 8px;
		border-radius: var(--radius-xs);
		font: 600 12px/1 var(--font-ui);
		white-space: nowrap;
		border: 1px solid transparent;
	}
	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
	}
	.badge--neutral {
		background: var(--ink-4);
		color: var(--fg-2);
	}
	.badge--neutral .dot {
		background: var(--fg-2);
	}
	.badge--gold {
		background: var(--gold-soft);
		color: var(--gold-300);
	}
	.badge--gold .dot {
		background: var(--gold-400);
	}
	.badge--info {
		background: var(--status-info-soft);
		color: var(--status-info);
	}
	.badge--info .dot {
		background: var(--status-info);
	}
</style>
