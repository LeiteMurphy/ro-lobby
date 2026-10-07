<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import ApplyDialog from '$lib/applications/components/ApplyDialog.svelte';
	import LeaveDialog from '$lib/applications/components/LeaveDialog.svelte';
	import PlayerPanel from '$lib/applications/components/PlayerPanel.svelte';
	import RejectDialog from '$lib/applications/components/RejectDialog.svelte';
	import SwapDialog from '$lib/applications/components/SwapDialog.svelte';
	import {
		applyState,
		composition,
		eligibility,
		HOST_KEY,
		people,
		swapEligibility,
		type Person
	} from '$lib/applications/detail';
	import { buildDays } from '$lib/home/days';
	import { instanceArt, ROLE_ICONS } from '$lib/home/catalog';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import { relativeLabel, zonedNow } from '$lib/home/time';
	import { ROLE_LABELS } from '$lib/home/types';
	import CancelDialog from '$lib/lobbies/components/CancelDialog.svelte';
	import { fromUtcIso } from '$lib/lobbies/time';
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

	// RN-15 / D-03: as vagas de cada função com o anfitrião e os membros aceitos.
	const list = $derived(people(lobby));
	const rows = $derived(composition(lobby, list));
	const candidates = $derived(list.filter((p) => p.kind === 'candidate'));
	const art = $derived(instanceArt(lobby.instance.name));
	const viewerState = $derived(applyState(lobby, data.user?.id ?? null));
	const canDecide = $derived(data.isOwner && lobby.status === 'open');

	// RN-31: a pessoa do painel da direita; começa no anfitrião e volta para ele se a
	// escolhida sair da lista (aceita, recusada ou retirada).
	let selectedKey = $state(HOST_KEY);
	const selected = $derived(list.find((p) => p.key === selectedKey) ?? list[0]);
	const classLevel = (p: Person) =>
		[p.classId ? (data.classNames[p.classId] ?? p.classId) : null, p.level ? `Nv ${p.level}` : null]
			.filter(Boolean)
			.join(' · ');

	// O erro de uma tentativa só volta para o diálogo da mesma abertura (AJ-04).
	let staleForm = $state.raw<typeof form>(null);
	const freshForm = $derived(form && form !== staleForm && 'action' in form ? form : null);
	const formOf = <A extends string>(action: A) =>
		freshForm?.action === action ? (freshForm as Extract<typeof freshForm, { action: A }>) : null;

	let cancelOpen = $state(false);
	let applyOpen = $state(false);
	let leaveOpen = $state(false);
	let swapOpen = $state(false);
	let rejecting = $state<Person | null>(null);
	let dialogKey = $state(0);
	const open = (set: () => void) => {
		staleForm = form;
		dialogKey++;
		set();
	};
	const cancelForm = $derived(formOf('cancel'));
	const applyForm = $derived(formOf('apply'));
	const rejectForm = $derived(formOf('reject'));
	const leaveForm = $derived(formOf('leave'));
	const swapForm = $derived(formOf('requestSwap'));
	// Erros de aceitar e retirar aparecem num aviso na página.
	const pageError = $derived(
		formOf('accept')?.message ??
			formOf('withdraw')?.message ??
			formOf('withdrawSwap')?.message ??
			null
	);
	const options = $derived(eligibility(lobby, data.characters));
	const mine = $derived(lobby.myApplication);

	// RN-21 / RN-38: o membro, o personagem dele no grupo e o pedido de troca pendente.
	const myPlace = $derived(list.find((p) => p.kind === 'member' && p.key === mine?.id) ?? null);
	const pendingSwap = $derived(
		mine?.status === 'accepted' && mine.swapRequest?.status === 'pending' ? mine.swapRequest : null
	);
	const swapTarget = $derived(
		pendingSwap ? data.characters.find((c) => c.id === pendingSwap.toCharacterId) : undefined
	);
	const swapOptions = $derived(
		mine
			? swapEligibility(
					lobby,
					data.characters,
					{ characterId: mine.characterId, role: mine.role },
					'request'
				)
			: []
	);
	const lobbyTitle = $derived(`${lobby.instance.name} · ${dayLabel} às ${when.time}`);
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
					{#if data.isOwner && lobby.pendingCount > 0}
						<!-- RN-34: o selo de pendentes é do dono. -->
						<span class="pill pill--pend" data-testid="pending-badge"
							>{lobby.pendingCount} {lobby.pendingCount === 1 ? 'pendente' : 'pendentes'}</span
						>
					{/if}
				</div>
			</div>
			{#if lobby.status === 'open'}
				<div class="acts">
					{#if viewerState === 'owner'}
						<Button
							variant="secondary"
							iconLeft="pencil"
							href={resolve('/lobbies/[id]/editar', { id: lobby.id })}>Editar</Button
						>
						<Button variant="outline" class="danger" onclick={() => open(() => (cancelOpen = true))}
							>Cancelar lobby</Button
						>
					{:else if viewerState === 'login'}
						<Button iconLeft="log-in" href={data.loginHref}>Entrar para se candidatar</Button>
					{:else if viewerState === 'apply'}
						<Button iconLeft="user-plus" onclick={() => open(() => (applyOpen = true))}
							>Candidatar</Button
						>
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

		{#if pageError}<p class="page-error" role="alert">{pageError}</p>{/if}
		{#if form && 'done' in form && form.done === 'apply'}
			<p class="page-ok" role="status">Candidatura enviada. O anfitrião vai aceitar ou recusar.</p>
		{/if}
		{#if form && 'done' in form && form.done === 'requestSwap'}
			<p class="page-ok" role="status">
				Pedido de troca enviado. O anfitrião vai aceitar ou recusar.
			</p>
		{/if}

		{#if mine && viewerState !== 'owner'}
			<!-- RN-29 / RN-13: a própria candidatura, com a justificativa e o botão de retirar. -->
			<div class="mine mine--{mine.status}" data-testid="my-application" role="status">
				{#if mine.status === 'pending'}
					<span><b>Sua candidatura está pendente.</b> O anfitrião ainda não decidiu.</span>
					{#if lobby.status === 'open'}
						<form method="POST" action="?/withdraw" use:enhance>
							<input type="hidden" name="applicationId" value={mine.id} />
							<Button type="submit" variant="secondary" size="sm">Retirar candidatura</Button>
						</form>
					{/if}
				{:else if mine.status === 'accepted' && pendingSwap && lobby.status === 'open'}
					<!-- RN-21 / RN-27: o pedido pendente, com o personagem pedido (Candidatura 2b). -->
					<span
						><b>Pedido de troca pendente</b> para {swapTarget?.nick ?? 'outro personagem'} ({ROLE_LABELS[
							pendingSwap.toRole
						]}{#if swapTarget}, Nv {swapTarget.level}{/if}). Você continua com {myPlace?.nick ??
							'o personagem atual'} até o anfitrião decidir.</span
					>
					<form method="POST" action="?/withdrawSwap" use:enhance>
						<input type="hidden" name="swapId" value={pendingSwap.id} />
						<Button type="submit" variant="secondary" size="sm">Retirar pedido</Button>
					</form>
				{:else if mine.status === 'accepted'}
					<span
						><b>Você está no grupo</b>{myPlace
							? ` com ${myPlace.nick} (${ROLE_LABELS[myPlace.role]})`
							: ''}. Combine os detalhes no Discord.</span
					>
					{#if lobby.status === 'open'}
						<!-- RN-14 / RN-20 / RN-38: pedir troca e sair (Candidatura 2a). -->
						<span class="mine-acts">
							<Button variant="secondary" size="sm" onclick={() => open(() => (swapOpen = true))}
								>Pedir troca</Button
							>
							<Button
								variant="outline"
								size="sm"
								class="danger"
								onclick={() => open(() => (leaveOpen = true))}>Sair do grupo</Button
							>
						</span>
					{/if}
				{:else if mine.status === 'removed'}
					<!-- RN-15 / RN-38: a justificativa e, com bloqueio, o aviso (Candidatura 2j). -->
					<span class="col"
						><b>Você foi removido deste lobby.</b>{#if mine.reason}<span
								>Justificativa: {mine.reason}</span
							>{/if}</span
					>
					{#if mine.blocked}<span class="sub">Você não pode se candidatar a este lobby.</span>{/if}
				{:else if mine.status === 'left'}
					<span>Você saiu do grupo. Pode se candidatar de novo enquanto houver vaga.</span>
				{:else if mine.status === 'rejected'}
					<span class="col"
						><b>Sua candidatura foi recusada.</b>{#if mine.reason}<span
								>Justificativa: {mine.reason}</span
							>{/if}</span
					>
				{:else if mine.status === 'withdrawn'}
					<span>Você retirou sua candidatura. Pode se candidatar de novo enquanto houver vaga.</span
					>
				{:else if mine.status === 'expired'}
					<span>Sua candidatura expirou sem decisão.</span>
				{/if}
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
						{#if r.slots.length > 0}
							<ul class="slots">
								{#each r.slots as p, i (p?.key ?? `open-${i}`)}
									{#if p}
										<li>
											<button
												type="button"
												class="slot filled slot--{r.role}"
												class:sel={selected.key === p.key}
												aria-pressed={selected.key === p.key}
												data-testid={p.kind === 'host' ? 'owner-slot' : 'member-slot'}
												onclick={() => (selectedKey = p.key)}
											>
												{#if p.portrait}<img
														class="portrait"
														src="/portraits/{p.portrait}.svg"
														alt=""
														width="52"
														height="52"
													/>{/if}
												<span class="who">
													<span class="nm">{p.nick}</span>
													{#if classLevel(p)}<span class="sub">{classLevel(p)}</span>{/if}
													{#if p.kind === 'host'}<span class="tag tag--{r.role}"
															><Icon name="star" size={11} />Anfitrião</span
														>{/if}
												</span>
											</button>
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

				{#if data.isOwner && lobby.status === 'open'}
					<div class="pending" data-testid="pending-list">
						<div class="rolehead">
							<span class="role">Candidaturas pendentes</span>
							<span class="hint">só você vê</span>
						</div>
						{#if candidates.length === 0}
							<span class="hint">Nenhuma candidatura pendente.</span>
						{:else}
							<ul class="slots">
								{#each candidates as p (p.key)}
									<li>
										<button
											type="button"
											class="slot cand"
											class:sel={selected.key === p.key}
											aria-pressed={selected.key === p.key}
											data-testid="candidate"
											onclick={() => (selectedKey = p.key)}
										>
											{#if p.portrait}<img
													class="portrait"
													src="/portraits/{p.portrait}.svg"
													alt=""
													width="52"
													height="52"
												/>{/if}
											<span class="who">
												<span class="nm">{p.nick}</span>
												<span class="sub"
													>{[classLevel(p), ROLE_LABELS[p.role]].filter(Boolean).join(' · ')}</span
												>
											</span>
										</button>
									</li>
								{/each}
							</ul>
						{/if}
					</div>
				{:else if lobby.pendingCount > 0}
					<!-- RN-28: os outros veem só a quantidade. -->
					<span class="hint" data-testid="pending-count"
						>{lobby.pendingCount}
						{lobby.pendingCount === 1 ? 'candidatura pendente' : 'candidaturas pendentes'}</span
					>
				{/if}
			</section>

			<div class="side">
				<PlayerPanel
					person={selected}
					classNames={data.classNames}
					{canDecide}
					onreject={(p) => open(() => (rejecting = p))}
				/>
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

{#key dialogKey}
	{#if cancelOpen}
		<CancelDialog
			title="{lobby.instance.name}, {dayLabel} às {when.time}"
			reason={cancelForm?.reason ?? ''}
			error={(cancelForm?.errors as { reason?: string } | undefined)?.reason ?? null}
			message={cancelForm?.message ?? null}
			onclose={() => (cancelOpen = false)}
		/>
	{/if}
	{#if applyOpen}
		<ApplyDialog
			title="{lobby.instance.name} · {dayLabel} às {when.time} · nível mínimo {lobby.minLevel}"
			{options}
			classNames={data.classNames}
			characterId={applyForm?.characterId ?? ''}
			message={applyForm?.text ?? ''}
			errors={applyForm?.errors ?? {}}
			formError={applyForm?.message ?? null}
			onclose={() => (applyOpen = false)}
		/>
	{/if}
	{#if leaveOpen && mine}
		<LeaveDialog
			applicationId={mine.id}
			title={lobbyTitle}
			roleLabel={ROLE_LABELS[mine.role]}
			message={leaveForm?.message ?? null}
			onclose={() => (leaveOpen = false)}
		/>
	{/if}
	{#if swapOpen && mine}
		<SwapDialog
			mode="request"
			hint="O anfitrião decide. Você continua no grupo com {myPlace?.nick ??
				'o personagem atual'} até lá."
			options={swapOptions}
			classNames={data.classNames}
			applicationId={mine.id}
			characterId={swapForm?.characterId ?? ''}
			reason={swapForm?.reason ?? ''}
			errors={swapForm?.errors ?? {}}
			formError={swapForm?.message ?? null}
			onclose={() => (swapOpen = false)}
		/>
	{/if}
	{#if rejecting && rejecting.applicationId}
		<RejectDialog
			applicationId={rejecting.applicationId}
			nick={rejecting.nick}
			reason={rejectForm?.reason ?? ''}
			error={rejectForm?.errors?.reason ?? null}
			message={rejectForm?.message ?? null}
			onclose={() => (rejecting = null)}
		/>
	{/if}
{/key}

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
	.note {
		margin: 0;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		background: var(--ink-2);
		color: var(--fg-2);
		font: 400 14px/1.45 var(--font-ui);
		overflow-wrap: anywhere;
	}
	button.slot {
		width: 100%;
		color: inherit;
		font: inherit;
		text-align: left;
		cursor: pointer;
		transition: box-shadow var(--dur-fast, 120ms) ease;
	}
	button.slot:hover {
		box-shadow: 0 0 0 1px var(--gold-line);
	}
	button.slot:focus-visible {
		outline: 2px solid var(--gold-400);
		outline-offset: 2px;
	}
	/* RN-31: o card escolhido fica com o anel âmbar (Candidatura 1a). */
	.slot.sel,
	button.slot.sel:hover {
		box-shadow:
			0 0 0 2px var(--gold-400),
			var(--shadow-glow-gold);
	}
	.slots > li {
		display: flex;
	}
	.slot.cand {
		border: 1px solid var(--gold-line);
		background: var(--gold-soft);
	}
	.pending {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding-top: 10px;
		border-top: 1px solid var(--border-subtle);
	}
	.pill {
		display: inline-flex;
		align-items: center;
		height: 22px;
		padding: 0 8px;
		border-radius: 999px;
		font: 600 11px/1 var(--font-ui);
	}
	.pill--pend {
		background: var(--gold-soft);
		color: var(--gold-300);
	}
	.mine {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 10px 16px;
		padding: 12px 16px;
		border-radius: var(--radius-md);
		border: 1px solid var(--panel-border);
		background: var(--panel-bg);
		color: var(--fg-2);
		font: 500 14px/1.4 var(--font-ui);
	}
	.mine b {
		color: var(--fg-1);
	}
	.mine form {
		margin: 0;
	}
	.mine-acts {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.mine :global(.danger) {
		color: var(--status-error);
		border-color: rgba(240, 100, 140, 0.4);
	}
	.mine--removed {
		border-color: rgba(240, 100, 140, 0.4);
	}
	.mine--pending {
		border-color: var(--gold-line);
	}
	.mine--accepted {
		border-color: var(--support-line);
	}
	.mine--rejected {
		border-color: rgba(240, 100, 140, 0.4);
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: 4px;
		overflow-wrap: anywhere;
	}
	.page-error,
	.page-ok {
		margin: 0;
		padding: 10px 14px;
		border-radius: var(--radius-sm);
		font: 500 13px/1.3 var(--font-ui);
	}
	.page-error {
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
	}
	.page-ok {
		background: rgba(79, 179, 161, 0.14);
		color: var(--support-300);
	}
</style>
