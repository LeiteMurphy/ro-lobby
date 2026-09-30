<script lang="ts">
	import Icon from '$lib/ui/Icon.svelte';
	import { ROLE_ICONS } from '../catalog';
	import { headcount } from '../lobbies';
	import { ROLE_LABELS, ROLES, type Composition } from '../types';

	interface Props {
		composition: Composition;
		showTotal?: boolean;
		showBar?: boolean;
	}

	let { composition, showTotal = false, showBar = false }: Props = $props();

	const rows = $derived(
		ROLES.map((role) => ({ role, ...composition[role] })).filter((r) => r.total > 0)
	);
	const total = $derived(headcount(composition));
</script>

<div class="comp">
	<div class="chips">
		{#each rows as row (row.role)}
			{@const full = row.filled >= row.total}
			<span
				class="chip chip--{row.role}"
				class:chip--full={full}
				data-role={row.role}
				title={full
					? `${ROLE_LABELS[row.role]}: completo`
					: `${ROLE_LABELS[row.role]}: ${row.total - row.filled} vaga(s)`}
			>
				<Icon name={ROLE_ICONS[row.role]} size={14} color="var(--{row.role}-400)" />
				<span class="label">{ROLE_LABELS[row.role]}</span>
				<span class="count">{row.filled}/{row.total}</span>
				{#if full}<Icon name="check" size={12} color="var(--fg-3)" />{/if}
			</span>
		{/each}
		{#if showTotal}
			<span class="total">{total.filled}<span class="of">/{total.total}</span></span>
		{/if}
	</div>
	{#if showBar}
		<div class="bar" aria-hidden="true">
			{#each rows as row (row.role)}
				{#each Array.from({ length: row.total }, (_, i) => i) as i (i)}
					<span style="background:{i < row.filled ? `var(--${row.role}-400)` : 'var(--ink-4)'}"
					></span>
				{/each}
			{/each}
		</div>
	{/if}
</div>

<style>
	.comp {
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 0;
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
	}
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 26px;
		padding: 0 8px;
		border-radius: var(--radius-xs);
		border: 1px solid transparent;
	}
	.chip--tank {
		background: var(--gradient-tank);
		border-color: var(--tank-line);
	}
	.chip--support {
		background: var(--gradient-support);
		border-color: var(--support-line);
	}
	.chip--dps {
		background: var(--gradient-dps);
		border-color: var(--dps-line);
	}
	.chip--full {
		background: transparent;
		border-color: var(--line-2);
	}
	.label {
		font: 600 12px/1 var(--font-ui);
	}
	.chip--tank .label {
		color: var(--tank-300);
	}
	.chip--support .label {
		color: var(--support-300);
	}
	.chip--dps .label {
		color: var(--dps-300);
	}
	.chip--full .label {
		color: var(--fg-3);
	}
	.count {
		font: 600 12px/1 var(--font-mono);
		color: var(--fg-1);
	}
	.chip--full .count {
		color: var(--fg-3);
	}
	.total {
		margin-left: auto;
		font: 600 13px/1 var(--font-mono);
		color: var(--fg-2);
	}
	.of {
		color: var(--fg-4);
	}
	.bar {
		display: flex;
		gap: 2px;
		height: 6px;
	}
	.bar span {
		flex: 1;
		border-radius: 1px;
	}
</style>
