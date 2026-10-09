<script lang="ts">
	import { resolve } from '$app/paths';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import { ROLE_LABELS, ROLES } from '$lib/home/types';
	import TalentCard from '$lib/talents/components/TalentCard.svelte';
	import { CLOCK_OPTIONS, WEEK_ORDER, WEEKDAY_NAMES } from '$lib/talents/format';
	import Button from '$lib/ui/Button.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const HIGH_LEVEL = 130;
	const uid = $props.id();
	const groups = $derived([
		{ label: 'Nível 130 ou mais', items: data.instances.filter((i) => i.level >= HIGH_LEVEL) },
		{ label: 'Nível menor', items: data.instances.filter((i) => i.level < HIGH_LEVEL) }
	]);
	const filtered = $derived(Object.values(data.filters).some(Boolean));
</script>

<svelte:head>
	<title>RO Lobby · banco de talentos</title>
</svelte:head>

<div class="page">
	<TopBar user={data.user} loginHref={data.loginHref} />
	<main class="content">
		<div class="titles">
			<h1>Banco de talentos</h1>
			<span class="hint"
				>Jogadores disponíveis para grupos, com os dias, o horário (Brasília) e as instâncias de
				interesse. Para entrar no banco, abra o personagem em Meus personagens.</span
			>
		</div>

		<!-- RN-11: filtros opcionais, combinados com "E", pela query string (funciona sem JS). -->
		<form class="filters" method="GET" aria-label="Filtros do banco de talentos">
			<div class="field">
				<label class="lbl" for="{uid}-instancia">Instância</label>
				<select id="{uid}-instancia" class="input" name="instancia">
					<option value="">Qualquer instância</option>
					{#each groups as g (g.label)}
						<optgroup label={g.label}>
							{#each g.items as i (i.id)}
								<option value={i.id} selected={data.filters.instancia === i.id}>{i.name}</option>
							{/each}
						</optgroup>
					{/each}
				</select>
			</div>
			<div class="field">
				<label class="lbl" for="{uid}-funcao">Função</label>
				<select id="{uid}-funcao" class="input" name="funcao">
					<option value="">Qualquer função</option>
					{#each ROLES as r (r)}
						<option value={r} selected={data.filters.funcao === r}>{ROLE_LABELS[r]}</option>
					{/each}
				</select>
			</div>
			<div class="field">
				<label class="lbl" for="{uid}-dia">Dia</label>
				<select id="{uid}-dia" class="input" name="dia">
					<option value="">Qualquer dia</option>
					{#each WEEK_ORDER as d (d)}
						<option value={String(d)} selected={data.filters.dia === String(d)}
							>{WEEKDAY_NAMES[d]}</option
						>
					{/each}
				</select>
			</div>
			<div class="field">
				<label class="lbl" for="{uid}-hora">Hora (Brasília)</label>
				<select id="{uid}-hora" class="input" name="hora">
					<option value="">Qualquer hora</option>
					{#each CLOCK_OPTIONS as c (c)}
						<option value={c} selected={data.filters.hora === c}>{c}</option>
					{/each}
				</select>
			</div>
			<div class="go">
				<Button type="submit" variant="secondary">Filtrar</Button>
				{#if filtered}<a class="clear" href={resolve('/talentos')}>Limpar</a>{/if}
			</div>
		</form>

		{#if data.loadError}
			<p class="alert" role="alert">{data.loadError}</p>
		{:else if data.talents.length === 0}
			<p class="empty" data-testid="talents-empty">
				{filtered ? 'Ninguém no banco com esses filtros.' : 'Ninguém no banco de talentos ainda.'}
			</p>
		{:else}
			{#if data.limited}
				<p class="hint" role="status">
					Mostrando os 100 primeiros. Use os filtros para achar mais.
				</p>
			{/if}
			<ul class="grid" aria-label="Jogadores no banco de talentos">
				{#each data.talents as t (t.characterId)}
					<li>
						<TalentCard
							talent={t}
							className={data.classNames[t.classId] ?? t.classId}
							loginHref={data.user ? undefined : data.loginHref}
						/>
					</li>
				{/each}
			</ul>
		{/if}
	</main>
</div>

<style>
	/* O fundo é o céu do design system, como na Home (vem do body). */
	.page {
		min-height: 100dvh;
	}
	.content {
		display: flex;
		flex-direction: column;
		gap: 20px;
		max-width: 1180px;
		margin: 0 auto;
		padding: 28px 16px 48px;
	}
	@media (min-width: 900px) {
		.content {
			padding: 28px 28px 64px;
		}
	}
	.titles {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	h1 {
		margin: 0;
		font: var(--type-h1);
	}
	.hint {
		margin: 0;
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.filters {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 12px;
		align-items: end;
		padding: 16px 18px;
		border-radius: var(--radius-lg);
		background: var(--panel-bg);
		backdrop-filter: blur(var(--panel-blur));
		border: 1px solid var(--panel-border);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}
	.lbl {
		font: 600 12px/1.2 var(--font-ui);
		color: var(--fg-2);
	}
	.input {
		height: 36px;
		padding: 0 12px;
		background: var(--surface-input);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-sm);
		color: var(--fg-1);
		font: 400 14px/1 var(--font-ui);
		color-scheme: dark;
	}
	.go {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.clear {
		font: 600 13px/1 var(--font-ui);
		color: var(--fg-2);
	}
	.alert,
	.empty {
		margin: 0;
		padding: 12px 16px;
		border-radius: var(--radius-md);
		font: 500 14px/1.4 var(--font-ui);
	}
	.alert {
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
	}
	.empty {
		background: var(--panel-bg);
		border: 1px solid var(--panel-border);
		color: var(--fg-2);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 12px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	@media (max-width: 360px) {
		.grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
