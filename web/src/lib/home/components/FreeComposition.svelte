<script lang="ts">
	import Icon from '$lib/ui/Icon.svelte';
	import type { Slots } from '../types';

	interface Props {
		/** Ocupantes e total do grupo livre (spec grupo-livre, RN-08). */
		seats: Slots;
		showBar?: boolean;
	}

	let { seats, showBar = false }: Props = $props();
	const full = $derived(seats.filled >= seats.total);
</script>

<!-- RN-08 da grupo-livre: selo "Grupo livre", "X de N" e a barra, sem chips por função. -->
<div class="comp" data-testid="free-composition">
	<div class="chips">
		<span class="chip" class:chip--full={full} title="Qualquer função ocupa qualquer vaga">
			<Icon name="users" size={14} color="var(--fg-2)" />
			<span class="label">Grupo livre</span>
			<span class="count">{seats.filled} de {seats.total}</span>
		</span>
	</div>
	{#if showBar}
		<div class="bar" aria-hidden="true">
			{#each Array.from({ length: seats.total }, (_, i) => i) as i (i)}
				<span style="background:{i < seats.filled ? 'var(--fg-2)' : 'var(--ink-4)'}"></span>
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
		gap: 6px;
	}
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 26px;
		padding: 0 9px;
		border-radius: 999px;
		border: 1px solid var(--line-2);
		background: var(--ink-2);
		font: 600 12px/1 var(--font-ui);
		color: var(--fg-1);
	}
	.chip--full {
		opacity: 0.6;
	}
	.count {
		color: var(--fg-2);
	}
	.bar {
		display: flex;
		gap: 3px;
	}
	.bar span {
		flex: 1;
		height: 4px;
		border-radius: 2px;
	}
</style>
