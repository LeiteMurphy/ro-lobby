<script lang="ts">
	import Icon, { type IconName } from './Icon.svelte';

	interface Props {
		icon: IconName;
		/** Rótulo acessível e dica (RNF-01). */
		label: string;
		variant?: 'ghost' | 'secondary' | 'primary';
		size?: 'sm' | 'md' | 'lg';
		/** Ação que ainda não existe (RN-18 da home-local). */
		soon?: boolean;
		onclick?: (event: MouseEvent) => void;
	}

	let { icon, label, variant = 'ghost', size = 'md', soon = false, onclick }: Props = $props();
	const iconSize = $derived(size === 'sm' ? 14 : size === 'lg' ? 20 : 18);
	const hintId = $props.id();
</script>

<span class="wrap">
	<button
		type="button"
		class="ib ib--{variant} ib--{size}"
		class:ib--soon={soon}
		aria-label={label}
		aria-disabled={soon ? 'true' : undefined}
		aria-describedby={soon ? hintId : undefined}
		title={soon ? undefined : label}
		onclick={(event) => (soon ? event.preventDefault() : onclick?.(event))}
	>
		<Icon name={icon} size={iconSize} />
	</button>
	{#if soon}
		<span id={hintId} role="tooltip" class="hint">Disponível em breve</span>
	{/if}
</span>

<style>
	.wrap {
		position: relative;
		display: inline-flex;
	}
	.ib {
		position: relative;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: var(--control-md);
		height: var(--control-md);
		border-radius: var(--radius-sm);
		border: 1px solid transparent;
		cursor: pointer;
		color: var(--text-secondary);
		background: transparent;
		transition:
			background var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out),
			border-color var(--dur-fast) var(--ease-out);
	}
	.ib:hover:not(.ib--soon) {
		background: var(--surface-raised);
		color: var(--text-primary);
	}
	.ib--soon {
		opacity: 0.42;
		cursor: not-allowed;
	}
	.ib--sm {
		width: var(--control-sm);
		height: var(--control-sm);
	}
	.ib--lg {
		width: var(--control-lg);
		height: var(--control-lg);
	}
	.ib--secondary {
		background: var(--surface-raised);
		border-color: var(--border-default);
		color: var(--text-primary);
	}
	.ib--secondary:hover:not(.ib--soon) {
		background: var(--surface-hover);
		border-color: var(--border-strong);
	}
	.ib--primary {
		background: var(--gradient-accent);
		color: var(--on-accent);
	}
	.ib--primary:hover:not(.ib--soon) {
		background: var(--gradient-accent-hover);
		color: var(--on-accent);
		box-shadow: var(--glow-gold-strong);
	}
	.hint {
		position: absolute;
		top: calc(100% + 6px);
		right: 0;
		z-index: var(--z-tooltip);
		pointer-events: none;
		white-space: nowrap;
		padding: 6px 9px;
		border-radius: var(--radius-sm);
		background: var(--ink-5);
		color: var(--fg-1);
		font: 500 12px/1.3 var(--font-ui);
		box-shadow: var(--shadow-pop);
		opacity: 0;
		transition: opacity var(--dur-fast) var(--ease-out);
	}
	.wrap:hover .hint,
	.wrap:focus-within .hint {
		opacity: 1;
	}
</style>
