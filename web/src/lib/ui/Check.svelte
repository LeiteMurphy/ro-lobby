<script lang="ts">
	import Icon, { type IconName } from './Icon.svelte';

	// Checkbox e Radio do design system: o mesmo visual, com a caixa quadrada ou redonda.
	interface Props {
		type: 'checkbox' | 'radio';
		label: string;
		checked: boolean;
		name?: string;
		value?: string;
		/** Contagem à direita, em números tabulares. */
		meta?: number;
		icon?: IconName;
		iconColor?: string;
		onchange: (checked: boolean) => void;
	}

	let { type, label, checked, name, value, meta, icon, iconColor, onchange }: Props = $props();
</script>

<label class="check" class:check--radio={type === 'radio'}>
	<input
		{type}
		{name}
		{value}
		{checked}
		onchange={(event) => onchange((event.currentTarget as HTMLInputElement).checked)}
	/>
	<span class="box">
		{#if type === 'radio'}
			<span class="mark dot"></span>
		{:else}
			<span class="mark tick"><Icon name="check" size={13} /></span>
		{/if}
	</span>
	{#if icon}<Icon name={icon} size={16} color={iconColor} />{/if}
	<span>{label}</span>
	{#if meta !== undefined}<span class="meta">{meta}</span>{/if}
</label>

<style>
	.check {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		cursor: pointer;
		font: 500 14px/1.3 var(--font-ui);
		color: var(--text-primary);
		user-select: none;
		min-height: 24px;
	}
	input {
		position: absolute;
		opacity: 0;
		width: 0;
		height: 0;
	}
	.box {
		flex: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		border-radius: var(--radius-xs);
		border: 1px solid var(--border-strong);
		background: var(--surface-input);
		color: var(--on-accent);
		transition:
			background var(--dur-fast) var(--ease-out),
			border-color var(--dur-fast) var(--ease-out);
	}
	.check--radio .box {
		border-radius: 50%;
	}
	.check:hover .box {
		border-color: var(--fg-3);
	}
	input:checked + .box {
		background: var(--accent);
		border-color: var(--accent);
	}
	input:focus-visible + .box {
		box-shadow: var(--focus-ring);
	}
	.mark {
		display: inline-flex;
		opacity: 0;
		transition: opacity var(--dur-fast) var(--ease-out);
	}
	input:checked + .box .mark {
		opacity: 1;
	}
	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--on-accent);
	}
	.meta {
		margin-left: auto;
		font: var(--type-time);
		font-size: 12px;
		color: var(--text-muted);
	}
</style>
