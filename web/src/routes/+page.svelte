<script lang="ts">
	import { onMount } from 'svelte';
	import { buildDays } from '$lib/home/days';
	import {
		activeFilterCount,
		applyFilters,
		countLabel,
		emptyState,
		instanceOptions,
		NO_FILTERS,
		roleCounts,
		timeRangeCounts,
		type Filters
	} from '$lib/home/filters';
	import { featuredLobby, lobbiesForDay } from '$lib/home/lobbies';
	import { relativeLabel, zonedNow } from '$lib/home/time';
	import DaySelector from '$lib/home/components/DaySelector.svelte';
	import EmptyState from '$lib/home/components/EmptyState.svelte';
	import FeaturedLobby from '$lib/home/components/FeaturedLobby.svelte';
	import FilterPanel from '$lib/home/components/FilterPanel.svelte';
	import LobbyCard from '$lib/home/components/LobbyCard.svelte';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import Button from '$lib/ui/Button.svelte';
	import Drawer from '$lib/ui/Drawer.svelte';
	import { LOGIN_ERROR_MESSAGE } from '$lib/auth/oauth';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// RN-15: o "agora" começa no do servidor (a hidratação bate com o HTML) e anda a cada
	// minuto no navegador.
	let nowIso = $derived(data.now);
	onMount(() => {
		nowIso = new Date().toISOString();
		const timer = setInterval(() => (nowIso = new Date().toISOString()), 60_000);
		return () => clearInterval(timer);
	});
	const now = $derived(zonedNow(new Date(nowIso)));

	let selectedDate = $derived(data.today);
	let filters = $state<Filters>(NO_FILTERS);

	const days = $derived(buildDays(data.today, data.lobbies));
	const day = $derived(days.find((d) => d.date === selectedDate) ?? days[0]);
	const dayLobbies = $derived(lobbiesForDay(data.lobbies, day.date));
	const filtered = $derived(applyFilters(dayLobbies, filters));
	const featured = $derived(featuredLobby(dayLobbies, now));
	const empty = $derived(emptyState(dayLobbies, filtered));
	const instances = $derived(instanceOptions(data.lobbies));
	const activeFilters = $derived(activeFilterCount(filters));
	const noFilters = $derived(activeFilters === 0);
	const dayRoleCounts = $derived(roleCounts(dayLobbies));
	const dayTimeCounts = $derived(timeRangeCounts(dayLobbies, filters));

	// RN-19: abaixo de 900 px os filtros ficam numa gaveta.
	let drawerOpen = $state(false);

	const heading = $derived(day.today ? 'Grupos para hoje' : `Grupos para ${day.label}`);
	const subheading = $derived(
		dayLobbies.length === 0
			? `${day.label} · nenhum grupo`
			: `${day.today ? `${day.label} · ` : ''}${countLabel(dayLobbies.length, filtered.length, filters)} · por horário`
	);

	const resetFilters = () => (filters = NO_FILTERS);
</script>

<svelte:head>
	<title>RO Lobby · grupos para hoje</title>
	<meta
		name="description"
		content="Grupos de Ragnarok Online para instâncias difíceis, com horário e composição por função."
	/>
</svelte:head>

<div class="page">
	<TopBar user={data.user} loginHref={data.loginHref} />

	{#if data.loginError}
		<div class="login-error" role="alert">{LOGIN_ERROR_MESSAGE}</div>
	{/if}
	{#if data.loadError}
		<div class="login-error" role="alert">{data.loadError}</div>
	{/if}

	<div class="days-bar">
		<div class="container">
			<DaySelector {days} value={day.date} onchange={(date) => (selectedDate = date)} />
		</div>
	</div>

	<main>
		<aside class="sidebar" aria-label="Filtros">
			<div class="sidebar-head">
				<span class="sidebar-title">Filtros</span>
				<Button variant="ghost" size="sm" disabled={noFilters} onclick={resetFilters}>Limpar</Button
				>
			</div>
			<FilterPanel
				{filters}
				{instances}
				roleCounts={dayRoleCounts}
				timeCounts={dayTimeCounts}
				name="faixa"
				onchange={(f) => (filters = f)}
			/>
		</aside>

		<section class="content">
			{#if featured}
				<FeaturedLobby
					lobby={featured}
					relative={relativeLabel(featured.date, featured.time, now)}
				/>
			{/if}

			<div class="list-head">
				<div class="titles">
					<h1>{heading}</h1>
					<span class="sub" data-testid="list-subheading">{subheading}</span>
				</div>
				<span class="filters-button">
					<Button
						variant="secondary"
						size="sm"
						iconLeft="sliders-horizontal"
						onclick={() => (drawerOpen = true)}
					>
						{activeFilters ? `Filtros (${activeFilters})` : 'Filtros'}
					</Button>
				</span>
				{#if filtered.length > 0}
					<div class="legend">
						<span>Borda: função com mais vagas</span>
						<span class="key"><span class="swatch tank"></span>Tank</span>
						<span class="key"><span class="swatch support"></span>Suporte</span>
						<span class="key"><span class="swatch dps"></span>Dano</span>
					</div>
				{/if}
			</div>

			{#if empty}
				<EmptyState kind={empty} onreset={resetFilters} />
			{:else}
				<div class="grid">
					{#each filtered as lobby (lobby.id)}
						<LobbyCard {lobby} relative={relativeLabel(lobby.date, lobby.time, now)} />
					{/each}
				</div>
			{/if}
		</section>
	</main>

	<Drawer open={drawerOpen} title="Filtros" onclose={() => (drawerOpen = false)}>
		<div class="drawer-body">
			<FilterPanel
				{filters}
				{instances}
				roleCounts={dayRoleCounts}
				timeCounts={dayTimeCounts}
				name="faixa-gaveta"
				size="md"
				onchange={(f) => (filters = f)}
			/>
			<div class="drawer-actions">
				<Button variant="ghost" size="lg" block onclick={resetFilters}>Limpar</Button>
				<Button variant="secondary" size="lg" block onclick={() => (drawerOpen = false)}>
					Ver {filtered.length}
					{filtered.length === 1 ? 'grupo' : 'grupos'}
				</Button>
			</div>
		</div>
	</Drawer>
</div>

<style>
	.page {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
	}
	/* RN-13 (login-discord): aviso de falha no login, com as cores de erro do design system. */
	.login-error {
		margin: 14px 28px 0;
		padding: 10px 14px;
		border-radius: var(--radius-md);
		background: var(--status-error-soft);
		border: 1px solid rgba(240, 100, 140, 0.4);
		color: var(--status-error);
		font: 600 14px/1.4 var(--font-ui);
	}
	.days-bar {
		background: rgba(17, 24, 34, 0.55);
		backdrop-filter: blur(var(--panel-blur));
		border-bottom: 1px solid var(--panel-border);
		padding: 14px 28px;
	}
	.container {
		max-width: 1640px;
		margin: 0 auto;
	}
	main {
		flex: 1;
		width: 100%;
		max-width: 1640px;
		margin: 0 auto;
		padding: 24px 28px 48px;
		display: flex;
		gap: 24px;
		align-items: flex-start;
	}
	.sidebar {
		width: var(--sidebar-w);
		flex: none;
		position: sticky;
		top: calc(var(--topbar-h) + 20px);
		background: var(--panel-bg);
		backdrop-filter: blur(var(--panel-blur));
		border: 1px solid var(--panel-border);
		border-radius: var(--radius-lg);
		padding: 18px;
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.sidebar-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.sidebar-title {
		font: 600 16px/1 var(--font-display);
	}
	.content {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	.list-head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
	}
	.titles {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	h1 {
		margin: 0;
		font: var(--type-h2);
		color: var(--fg-1);
	}
	.sub {
		font: var(--type-small);
		color: var(--fg-3);
	}
	.legend {
		display: flex;
		align-items: center;
		gap: 12px;
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.key {
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.swatch {
		width: 3px;
		height: 12px;
		border-radius: 1px;
	}
	.swatch.tank {
		background: var(--tank-400);
	}
	.swatch.support {
		background: var(--support-400);
	}
	.swatch.dps {
		background: var(--dps-400);
	}
	.filters-button {
		display: none;
	}
	.drawer-body {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.drawer-actions {
		display: grid;
		grid-template-columns: 1fr 2fr;
		gap: 8px;
	}
	/* RN-19: abaixo de 900 px, a barra lateral some e os filtros vão para a gaveta. */
	@media (max-width: 899px) {
		.days-bar {
			padding: 14px 16px;
		}
		main {
			padding: 24px 16px 48px;
		}
		.sidebar,
		.legend {
			display: none;
		}
		.filters-button {
			display: inline-flex;
		}
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 330px), 1fr));
		gap: 14px;
	}
</style>
