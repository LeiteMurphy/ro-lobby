<script lang="ts">
	import { resolve } from '$app/paths';
	import { buildDays } from '$lib/home/days';
	import { instanceArt, ROLE_ICONS } from '$lib/home/catalog';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import { relativeLabel, zonedNow } from '$lib/home/time';
	import { ROLE_LABELS, ROLES, type Role } from '$lib/home/types';
	import CancelDialog from '$lib/lobbies/components/CancelDialog.svelte';
	import { fromUtcIso } from '$lib/lobbies/time';
	import { DELETED_CHARACTER } from '$lib/lobbies/toHome';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const lobby = $derived(data.lobby);
	const when = $derived(fromUtcIso(lobby.startsAt));
	const dayLabel = $derived(buildDays(when.date, [])[0].label);
	const relative = $derived(relativeLabel(when.date, when.time, zonedNow(new Date(data.now))));
	const occupied = $derived(lobby.occupied.tank + lobby.occupied.support + lobby.occupied.dps);
	const total = $derived(lobby.slots.tank + lobby.slots.support + lobby.slots.dps);

	const RESET_LABELS = {
		daily: 'Retorno diário',
		three_days: 'Retorno em 3 dias',
		hours: 'Retorno em horas',
		weekly: 'Retorno semanal'
	} as const;
	const STATUS_LABELS = { open: 'Aberto', started: 'Iniciado', cancelled: 'Cancelado' } as const;

	// RN-15: as vagas de cada função, com o personagem do dono na vaga da função dele.
	type Slot = { owner: boolean };
	const rows = $derived(
		ROLES.map((role: Role) => ({
			role,
			total: lobby.slots[role],
			filled: lobby.occupied[role],
			slots: Array.from({ length: lobby.slots[role] }, (_, i): Slot => ({
				owner: role === lobby.owner.role && i === 0
			}))
		}))
	);
	const ownerName = $derived(lobby.owner.nick ?? DELETED_CHARACTER);
	const art = $derived(instanceArt(lobby.instance.name));
	const ownerPortrait = $derived(
		lobby.owner.portrait ? `/portraits/${lobby.owner.portrait}.svg` : null
	);
	const ownerClassLevel = $derived(
		[data.ownerClass, lobby.owner.level ? `Nv ${lobby.owner.level}` : null]
			.filter(Boolean)
			.join(' · ')
	);

	let cancelOpen = $state(false);
	let cancelKey = $state(0);
	// O erro de uma tentativa só volta para o diálogo da mesma abertura (AJ-04).
	let staleForm = $state.raw<typeof form>(null);
	const openCancel = () => {
		staleForm = form;
		cancelKey++;
		cancelOpen = true;
	};
	const cancelForm = $derived(form && form !== staleForm && 'errors' in form ? form : null);
</script>

<svelte:head>
	<title>RO Lobby · {lobby.instance.name}</title>
</svelte:head>

<div class="page">
	<TopBar user={data.user} loginHref={data.loginHref} />
	<main class="content">
		<a class="back" href={resolve('/')}
			><Icon name="chevron-left" size={14} />Voltar para os grupos</a
		>

		<section class="hero" aria-labelledby="lobby-title">
			<div class="cover" style="background:{art.cover}">
				<img src={art.icon} alt="" width="88" height="88" />
			</div>
			<div class="info">
				<span class="status status--{lobby.status}" data-testid="lobby-status">
					{STATUS_LABELS[lobby.status]}{#if lobby.status === 'open' && relative}
						· começa {relative}{/if}
				</span>
				<h1 id="lobby-title">{lobby.instance.name}</h1>
				<div class="meta">
					<span><b>{dayLabel} · {when.time}</b> (Brasília)</span>
					<span>Nível mínimo <b>{lobby.minLevel}</b></span>
					{#if lobby.instance.reset}<span>{RESET_LABELS[lobby.instance.reset]}</span>{/if}
					<span><b>{occupied} de {total}</b> vagas ocupadas</span>
				</div>
			</div>
			{#if lobby.status === 'open'}
				<div class="acts">
					{#if data.isOwner}
						<Button
							variant="secondary"
							iconLeft="pencil"
							href={resolve('/lobbies/[id]/editar', { id: lobby.id })}>Editar</Button
						>
						<Button variant="outline" class="danger" onclick={openCancel}>Cancelar lobby</Button>
					{:else}
						<!-- RN-16: a candidatura entra na feature seguinte. -->
						<Button iconLeft="user-plus" soon>Candidatar</Button>
					{/if}
				</div>
			{/if}
		</section>

		{#if lobby.status === 'cancelled'}
			<div class="banner" role="status">
				<span class="over">Motivo do cancelamento</span>
				<span>{lobby.cancelReason}</span>
			</div>
		{/if}

		<div class="cols">
			<section class="panel" aria-labelledby="comp-title">
				<h2 id="comp-title">Composição</h2>
				{#each rows as r (r.role)}
					<div class="rolerow">
						<div class="rolehead">
							<span class="role role--{r.role}"
								><Icon name={ROLE_ICONS[r.role]} size={15} />{ROLE_LABELS[r.role]}</span
							>
							<span class="hint">{r.total === 0 ? 'Sem vagas' : `${r.filled} de ${r.total}`}</span>
						</div>
						{#if r.total > 0}
							<ul class="slots">
								{#each r.slots as s, i (i)}
									{#if s.owner}
										<li class="slot filled slot--{r.role}" data-testid="owner-slot">
											{#if ownerPortrait}<img
													class="portrait"
													src={ownerPortrait}
													alt=""
													width="52"
													height="52"
												/>{/if}
											<span class="who">
												<span class="nm">{ownerName}</span>
												{#if ownerClassLevel}<span class="sub">{ownerClassLevel}</span>{/if}
												<span class="tag tag--{r.role}"
													><Icon name="star" size={11} />Anfitrião</span
												>
											</span>
										</li>
									{:else}
										<li class="slot slot--{r.role}">
											<span class="empty-icon"><Icon name={ROLE_ICONS[r.role]} size={18} /></span>
											<span class="who">
												<span class="open">Vaga aberta</span>
												<span class="sub">{ROLE_LABELS[r.role]}</span>
											</span>
										</li>
									{/if}
								{/each}
							</ul>
						{/if}
					</div>
				{/each}
			</section>

			<div class="side">
				<section class="panel" aria-labelledby="host-title">
					<h2 id="host-title">Anfitrião</h2>
					<div class="host">
						{#if ownerPortrait}<img
								class="portrait"
								src={ownerPortrait}
								alt=""
								width="56"
								height="56"
							/>{/if}
						<span class="who"
							><span class="nm big">{ownerName}</span><span class="sub"
								>{ROLE_LABELS[lobby.owner.role]}</span
							></span
						>
					</div>
					<div class="kv"><span>Personagem</span><b>{ownerName}</b></div>
					{#if ownerClassLevel}
						<div class="kv"><span>Classe</span><b>{ownerClassLevel}</b></div>
					{/if}
					<div class="kv"><span>Discord</span><b>{lobby.owner.discordName}</b></div>
				</section>
				{#if lobby.note}
					<section class="panel" aria-labelledby="note-title">
						<h2 id="note-title">Observação</h2>
						<p class="note">{lobby.note}</p>
					</section>
				{/if}
			</div>
		</div>
	</main>
</div>

{#if cancelOpen}
	{#key cancelKey}
		<CancelDialog
			title="{lobby.instance.name}, {dayLabel} às {when.time}"
			reason={cancelForm?.reason ?? ''}
			error={(cancelForm?.errors as { reason?: string } | undefined)?.reason ?? null}
			message={cancelForm?.message ?? null}
			onclose={() => (cancelOpen = false)}
		/>
	{/key}
{/if}

<style>
	/* O fundo é o céu do design system, como na Home (vem do body). */
	.page {
		min-height: 100dvh;
	}
	.content {
		display: flex;
		flex-direction: column;
		gap: 22px;
		max-width: 1180px;
		margin: 0 auto;
		padding: 28px 16px 48px;
	}
	@media (min-width: 900px) {
		.content {
			padding: 28px 28px 64px;
		}
	}
	.back {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		align-self: flex-start;
		color: var(--fg-3);
		font: 500 13px/1 var(--font-ui);
		text-decoration: none;
	}
	.back:hover {
		color: var(--fg-1);
	}
	.hero {
		position: relative;
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		gap: 22px;
		align-items: center;
		padding: 22px;
		border-radius: var(--radius-lg);
		border: 1px solid var(--panel-border);
		background: var(--gradient-hero);
		box-shadow: var(--shadow-hover);
	}
	.cover {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 140px;
		height: 140px;
		border-radius: var(--radius-md);
		border: 1px solid var(--line-2);
	}
	.cover img {
		image-rendering: pixelated;
	}
	.info {
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
	}
	@media (max-width: 640px) {
		.hero {
			grid-template-columns: minmax(0, 1fr);
		}
		.cover {
			width: 100%;
			height: 110px;
		}
	}
	h1 {
		margin: 0;
		font: var(--type-h1);
		overflow-wrap: anywhere;
	}
	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 16px;
		font: 500 14px/1.3 var(--font-ui);
		color: var(--fg-2);
	}
	.meta b {
		color: var(--fg-1);
	}
	.status {
		align-self: flex-start;
		display: inline-flex;
		align-items: center;
		height: 24px;
		padding: 0 9px;
		border-radius: 999px;
		font: 600 12px/1 var(--font-ui);
	}
	.status--open {
		background: rgba(79, 179, 161, 0.16);
		color: var(--support-300);
	}
	.status--started {
		background: var(--ink-3);
		color: var(--fg-2);
	}
	.status--cancelled {
		background: rgba(240, 100, 140, 0.16);
		color: var(--status-error);
	}
	.acts {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.acts :global(.danger) {
		color: var(--status-error);
		border-color: rgba(240, 100, 140, 0.4);
	}
	@media (min-width: 900px) {
		.acts {
			position: absolute;
			right: 22px;
			top: 22px;
		}
	}
	.banner {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 12px 16px;
		border-radius: var(--radius-md);
		background: rgba(240, 100, 140, 0.1);
		border: 1px solid rgba(240, 100, 140, 0.3);
	}
	.over {
		font: 600 10px/1.2 var(--font-ui);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--status-error);
	}
	.cols {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 22px;
		align-items: start;
	}
	@media (min-width: 900px) {
		.cols {
			grid-template-columns: minmax(0, 1fr) 320px;
		}
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.panel {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 20px 22px;
		border-radius: var(--radius-lg);
		background: var(--panel-bg);
		backdrop-filter: blur(var(--panel-blur));
		border: 1px solid var(--panel-border);
	}
	h2 {
		margin: 0;
		font: var(--type-title);
	}
	.rolerow {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.rolehead {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.role {
		display: flex;
		align-items: center;
		gap: 6px;
		font: 600 13px/1 var(--font-ui);
	}
	.role--tank {
		color: var(--tank-300);
	}
	.role--support {
		color: var(--support-300);
	}
	.role--dps {
		color: var(--dps-300);
	}
	.hint {
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.slots {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
		gap: 10px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.slot {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 76px;
		padding: 10px 14px;
		border-radius: var(--radius-md);
		border: 1px dashed var(--line-2);
		background: rgba(11, 16, 23, 0.35);
	}
	.slot.filled {
		border-style: solid;
		background: var(--surface-raised);
	}
	.slot--tank.filled {
		border-color: var(--tank-line);
		background: var(--gradient-tank);
	}
	.slot--support.filled {
		border-color: var(--support-line);
		background: var(--gradient-support);
	}
	.slot--dps.filled {
		border-color: var(--dps-line);
		background: var(--gradient-dps);
	}
	.portrait {
		flex: none;
		border-radius: var(--radius-sm);
		border: 1px solid var(--line-2);
		image-rendering: pixelated;
	}
	.empty-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 52px;
		height: 52px;
		flex: none;
		border-radius: var(--radius-sm);
		border: 1px dashed var(--line-2);
		color: var(--fg-4);
	}
	.slot--tank .empty-icon {
		color: var(--tank-400);
	}
	.slot--support .empty-icon {
		color: var(--support-400);
	}
	.slot--dps .empty-icon {
		color: var(--dps-400);
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}
	.nm {
		font: 700 15px/1.15 var(--font-display);
		color: var(--fg-1);
		overflow-wrap: anywhere;
	}
	.nm.big {
		font-size: 18px;
	}
	.open {
		font: 600 13px/1.2 var(--font-ui);
		color: var(--fg-2);
	}
	.sub {
		font: 500 12px/1.3 var(--font-ui);
		color: var(--fg-3);
	}
	.tag {
		align-self: flex-start;
		display: inline-flex;
		align-items: center;
		gap: 4px;
		height: 20px;
		padding: 0 7px;
		border-radius: var(--radius-xs);
		background: var(--gold-soft);
		color: var(--gold-300);
		font: 600 11px/1 var(--font-ui);
	}
	.host {
		display: flex;
		align-items: center;
		gap: 12px;
		padding-bottom: 4px;
		border-bottom: 1px solid var(--border-subtle);
	}
	.kv {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		font: 500 13px/1.3 var(--font-ui);
		color: var(--fg-3);
	}
	.kv b {
		color: var(--fg-1);
		font-weight: 600;
		text-align: right;
	}
	.note {
		margin: 0;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		background: var(--ink-2);
		color: var(--fg-2);
		font: 400 14px/1.45 var(--font-ui);
		overflow-wrap: anywhere;
	}
</style>
