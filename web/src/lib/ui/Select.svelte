<script lang="ts">
	import Icon from './Icon.svelte';

	export interface SelectOption {
		value: string;
		label: string;
	}

	interface Props {
		/** Rótulo acessível (RNF-01). */
		label: string;
		options: ReadonlyArray<SelectOption>;
		value: string;
		placeholder?: string;
		size?: 'sm' | 'md';
		onchange: (value: string) => void;
	}

	let { label, options, value, placeholder, size = 'md', onchange }: Props = $props();
</script>

<div class="input input--{size}">
	<select
		aria-label={label}
		{value}
		onchange={(event) => onchange((event.currentTarget as HTMLSelectElement).value)}
	>
		{#if placeholder}<option value="">{placeholder}</option>{/if}
		{#each options as option (option.value)}
			<option value={option.value}>{option.label}</option>
		{/each}
	</select>
	<span class="chevron"><Icon name="chevron-down" size={16} /></span>
</div>

<style>
	.input {
		position: relative;
		display: flex;
		align-items: center;
		gap: 8px;
		height: var(--control-md);
		padding: 0 12px;
		background: var(--surface-input);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-sm);
		color: var(--text-muted);
		transition:
			border-color var(--dur-fast) var(--ease-out),
			box-shadow var(--dur-fast) var(--ease-out);
	}
	.input:hover {
		border-color: var(--border-strong);
	}
	.input:focus-within {
		border-color: var(--accent);
		box-shadow: 0 0 0 3px var(--accent-soft);
	}
	.input--sm {
		height: var(--control-sm);
		padding: 0 10px;
	}
	select {
		flex: 1;
		min-width: 0;
		height: 100%;
		background: transparent;
		border: 0;
		outline: 0;
		color: var(--text-primary);
		font: 400 14px/1.4 var(--font-ui);
		padding: 0 22px 0 0;
		appearance: none;
		cursor: pointer;
	}
	select:focus-visible {
		box-shadow: none;
	}
	option {
		background: var(--ink-2);
		color: var(--fg-1);
	}
	.chevron {
		position: absolute;
		right: 10px;
		display: flex;
		pointer-events: none;
	}
</style>
