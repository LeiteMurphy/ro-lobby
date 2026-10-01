<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon, { type IconName } from './Icon.svelte';

	export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'outline';
	export type ButtonSize = 'sm' | 'md' | 'lg';

	interface Props {
		variant?: ButtonVariant;
		size?: ButtonSize;
		iconLeft?: IconName;
		block?: boolean;
		disabled?: boolean;
		/**
		 * Ação que ainda não existe (RN-18 da home-local): o botão continua focável, fica
		 * com aria-disabled e mostra a dica "Disponível em breve".
		 */
		soon?: boolean;
		onclick?: (event: MouseEvent) => void;
		children: Snippet;
		class?: string;
	}

	let {
		variant = 'primary',
		size = 'md',
		iconLeft,
		block = false,
		disabled = false,
		soon = false,
		onclick,
		children,
		class: className
	}: Props = $props();

	const iconSize = $derived(size === 'sm' ? 14 : size === 'lg' ? 18 : 16);
	const hintId = $props.id();

	function handleClick(event: MouseEvent) {
		if (soon) {
			event.preventDefault();
			return;
		}
		onclick?.(event);
	}
</script>

<span class="wrap" class:block>
	<button
		type="button"
		class="btn btn--{variant} btn--{size} {className ?? ''}"
		class:btn--block={block}
		class:btn--soon={soon}
		{disabled}
		aria-disabled={soon ? 'true' : undefined}
		aria-describedby={soon ? hintId : undefined}
		onclick={handleClick}
	>
		{#if iconLeft}<Icon name={iconLeft} size={iconSize} />{/if}
		{@render children()}
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
	.wrap.block {
		display: flex;
		width: 100%;
	}
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		height: var(--control-md);
		padding: 0 16px;
		font: 600 14px/1 var(--font-ui);
		letter-spacing: 0.005em;
		border-radius: var(--radius-sm);
		border: 1px solid transparent;
		cursor: pointer;
		white-space: nowrap;
		transition:
			background var(--dur-fast) var(--ease-out),
			border-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out),
			transform var(--dur-fast) var(--ease-out);
	}
	.btn:active:not(:disabled, .btn--soon) {
		transform: translateY(1px);
	}
	.btn:disabled,
	.btn--soon {
		opacity: 0.42;
		cursor: not-allowed;
	}
	.btn--sm {
		height: var(--control-sm);
		padding: 0 10px;
		font-size: 13px;
		gap: 6px;
	}
	.btn--lg {
		height: var(--control-lg);
		padding: 0 22px;
		font-size: 15px;
	}
	.btn--block {
		width: 100%;
	}
	.btn--primary {
		background: var(--gradient-accent);
		color: var(--on-accent);
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3);
	}
	.btn--primary:hover:not(:disabled, .btn--soon) {
		background: var(--gradient-accent-hover);
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			var(--glow-gold-strong);
	}
	.btn--primary:active:not(:disabled, .btn--soon) {
		background: var(--gradient-accent-press);
		box-shadow: none;
	}
	.btn--secondary {
		background: var(--surface-raised);
		border-color: var(--border-default);
		color: var(--text-primary);
	}
	.btn--secondary:hover:not(:disabled, .btn--soon) {
		background: var(--surface-hover);
		border-color: var(--border-strong);
	}
	.btn--ghost {
		background: transparent;
		color: var(--text-secondary);
	}
	.btn--ghost:hover:not(:disabled, .btn--soon) {
		background: var(--surface-raised);
		color: var(--text-primary);
	}
	.btn--outline {
		background: transparent;
		border-color: var(--accent-line);
		color: var(--gold-300);
	}
	.btn--outline:hover:not(:disabled, .btn--soon) {
		background: var(--accent-soft);
		border-color: var(--accent);
		box-shadow: var(--glow-gold);
	}
	/* RNF-01: o anel de foco vence o box-shadow das variantes e do hover. */
	.btn.btn:focus-visible {
		box-shadow:
			var(--focus-ring),
			0 0 14px rgba(246, 187, 69, 0.28);
	}
	.hint {
		position: absolute;
		bottom: calc(100% + 6px);
		left: 50%;
		transform: translateX(-50%);
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
