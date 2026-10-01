<script lang="ts">
	import IconButton from '$lib/ui/IconButton.svelte';
	import { dayAriaLabel, type Day } from '../days';

	interface Props {
		days: readonly Day[];
		value: string;
		onchange: (date: string) => void;
	}

	let { days, value, onchange }: Props = $props();
	let track: HTMLDivElement | undefined = $state();

	const scroll = (direction: number) => track?.scrollBy({ left: direction * 300 });
</script>

<div class="ds">
	<span class="arrow"
		><IconButton
			icon="chevron-left"
			label="Dias anteriores"
			variant="secondary"
			onclick={() => scroll(-1)}
		/></span
	>
	<div class="track" bind:this={track} role="tablist" aria-label="Dias">
		{#each days as day (day.date)}
			{@const selected = day.date === value}
			<button
				type="button"
				role="tab"
				aria-selected={selected}
				aria-label={dayAriaLabel(day)}
				data-date={day.date}
				class="day"
				class:sel={selected}
				class:empty={!day.count}
				onclick={() => onchange(day.date)}
			>
				{#if day.today}<span class="today-dot"></span>{/if}
				<span class="wd">{day.today ? 'Hoje' : day.weekday}</span>
				<span class="num">{day.day}</span>
				<span class="count">{day.count ? day.count : '—'}</span>
			</button>
		{/each}
	</div>
	<span class="arrow"
		><IconButton
			icon="chevron-right"
			label="Próximos dias"
			variant="secondary"
			onclick={() => scroll(1)}
		/></span
	>
</div>

<style>
	.ds {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}
	.track {
		display: flex;
		gap: 8px;
		overflow-x: auto;
		scroll-behavior: smooth;
		scrollbar-width: none;
		flex: 1;
		min-width: 0;
		padding: 2px;
	}
	.track::-webkit-scrollbar {
		display: none;
	}
	.day {
		flex: none;
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 5px;
		width: 68px;
		height: 80px;
		padding: 0;
		background: var(--surface-card);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-md);
		color: var(--text-primary);
		cursor: pointer;
		font: inherit;
		transition:
			background var(--dur-fast) var(--ease-out),
			border-color var(--dur-fast) var(--ease-out);
	}
	.day:hover {
		background: var(--surface-raised);
		border-color: var(--border-strong);
		box-shadow: var(--shadow-soft);
	}
	.day.sel,
	.day.sel:hover {
		background: var(--gradient-featured);
		border-color: var(--accent-line);
		box-shadow: var(--glow-gold);
	}
	/* RNF-01: o anel de foco aparece também no dia selecionado e durante o hover. */
	.day.day:focus-visible {
		box-shadow:
			var(--focus-ring),
			0 0 14px rgba(246, 187, 69, 0.28);
	}
	.day.sel.sel:focus-visible {
		box-shadow: var(--focus-ring), var(--glow-gold);
	}
	.wd {
		font: 600 10px/1 var(--font-ui);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--text-muted);
	}
	.sel .wd {
		color: var(--gold-300);
	}
	.num {
		font: 700 22px/1 var(--font-display);
	}
	.empty .num {
		color: var(--text-muted);
	}
	.sel .num {
		color: var(--gold-200);
	}
	.count {
		min-width: 22px;
		height: 16px;
		padding: 0 5px;
		border-radius: var(--radius-pill);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font: 600 10px/1 var(--font-mono);
		background: var(--ink-4);
		color: var(--fg-2);
	}
	.sel .count {
		background: var(--gradient-accent);
		color: var(--on-accent);
	}
	.empty .count {
		background: transparent;
		color: var(--fg-4);
	}
	.today-dot {
		position: absolute;
		top: 6px;
		right: 6px;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--accent);
	}
	@media (max-width: 899px) {
		.arrow {
			display: none;
		}
	}
</style>
