<script lang="ts">
	import type { Snippet } from 'svelte';
	import IconButton from './IconButton.svelte';

	interface Props {
		open: boolean;
		title: string;
		onclose: () => void;
		children: Snippet;
		width?: number;
	}

	let { open, title, onclose, children, width = 320 }: Props = $props();

	function onkeydown(event: KeyboardEvent) {
		if (open && event.key === 'Escape') onclose();
	}
</script>

<svelte:window {onkeydown} />

<div class="root" class:open aria-hidden={!open} inert={!open}>
	<div class="scrim" onclick={onclose} role="presentation"></div>
	<div
		class="panel"
		role="dialog"
		aria-modal="true"
		aria-label={title}
		style="width:min({width}px, 88vw)"
	>
		<header>
			<span class="title">{title}</span>
			<IconButton icon="x" label="Fechar" onclick={onclose} />
		</header>
		<div class="body">{@render children()}</div>
	</div>
</div>

<style>
	.root {
		position: fixed;
		inset: 0;
		z-index: var(--z-drawer);
		pointer-events: none;
	}
	.root.open {
		pointer-events: auto;
	}
	.scrim {
		position: absolute;
		inset: 0;
		background: var(--scrim);
		opacity: 0;
		transition: opacity var(--dur-slow) var(--ease-out);
	}
	.open .scrim {
		opacity: 1;
	}
	.panel {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		display: flex;
		flex-direction: column;
		background: var(--surface-sidebar);
		backdrop-filter: blur(var(--panel-blur));
		border-right: 1px solid var(--border-default);
		transform: translateX(-100%);
		transition: transform var(--dur-slow) var(--ease-out);
	}
	.open .panel {
		transform: none;
	}
	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 12px 12px 12px 20px;
		border-bottom: 1px solid var(--border-subtle);
	}
	.title {
		font: var(--type-h3);
		font-size: 18px;
	}
	.body {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
	}
</style>
