<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { STATUS_LABELS } from '$lib/applications/messages';
	import { buildDays } from '$lib/home/days';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import { ROLE_LABELS } from '$lib/home/types';
	import { fromUtcIso } from '$lib/lobbies/time';
	import { DELETED_CHARACTER } from '$lib/lobbies/toHome';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const when = (iso: string) => {
		const { date, time } = fromUtcIso(iso);
		return `${buildDays(date, [])[0].label} · ${time}`;
	};
	// Cor do selo de cada estado (Candidatura 1e).
	const TONE: Record<string, string> = {
		pending: 'pend',
		accepted: 'ok',
		rejected: 'no',
		removed: 'no'
	};
</script>

<svelte:head>
	<title>RO Lobby · Minhas candidaturas</title>
</svelte:head>

<div class="page">
	<TopBar user={data.user} loginHref="/" />
	<main class="content">
		<a class="back" href={resolve('/')}
			><Icon name="chevron-left" size={14} />Voltar para os grupos</a
		>
		<h1>Minhas candidaturas</h1>
		{#if form?.message}<p class="form-error" role="alert">{form.message}</p>{/if}
		{#if data.applications.length === 0}
			<p class="empty">
				Você ainda não se candidatou a nenhum grupo. Escolha um na <a href={resolve('/')}>Home</a>.
			</p>
		{:else}
			<ul class="list">
				{#each data.applications as a (a.application.id)}
					{@const c = a.character}
					<li class="row" data-testid="my-application-row">
						{#if c}<img
								class="portrait"
								src="/portraits/{c.portrait}.svg"
								alt=""
								width="52"
								height="52"
							/>{:else}<span class="portrait"></span>{/if}
						<span class="who">
							<a class="ttl" href={resolve('/lobbies/[id]', { id: a.application.lobbyId })}
								>{a.lobby.instanceName}</a
							>
							<span class="sub"
								>{when(a.lobby.startsAt)} · com {c?.nick ?? DELETED_CHARACTER} ({ROLE_LABELS[
									a.application.role
								]}){#if c}
									· {data.classNames[c.classId] ?? c.classId} Nv {c.level}{/if}</span
							>
						</span>
						<span class="pill pill--{TONE[a.application.status] ?? 'off'}"
							>{STATUS_LABELS[a.application.status]}</span
						>
						{#if a.application.status === 'pending' && a.lobby.status === 'open'}
							<form method="POST" action="?/withdraw" use:enhance>
								<input type="hidden" name="applicationId" value={a.application.id} />
								<Button type="submit" variant="secondary" size="sm">Retirar</Button>
							</form>
						{:else}
							<Button
								variant="ghost"
								size="sm"
								href={resolve('/lobbies/[id]', { id: a.application.lobbyId })}>Ver grupo</Button
							>
						{/if}
						{#if a.application.reason}
							<!-- RN-29: a justificativa da recusa só para quem se candidatou. -->
							<p class="why"><b>Justificativa:</b> {a.application.reason}</p>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</main>
</div>

<style>
	.page {
		min-height: 100dvh;
	}
	.content {
		display: flex;
		flex-direction: column;
		gap: 18px;
		max-width: 860px;
		margin: 0 auto;
		padding: 28px 16px 48px;
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
	h1 {
		margin: 0;
		font: var(--type-h2);
	}
	.empty {
		margin: 0;
		padding: 18px 20px;
		border-radius: var(--radius-lg);
		background: var(--panel-bg);
		border: 1px solid var(--panel-border);
		color: var(--fg-2);
	}
	.empty a {
		color: var(--gold-300);
	}
	.list {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.row {
		display: grid;
		grid-template-columns: 52px minmax(0, 1fr) auto auto;
		gap: 10px 14px;
		align-items: center;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		background: var(--panel-bg);
		backdrop-filter: blur(var(--panel-blur));
		border: 1px solid var(--panel-border);
	}
	@media (max-width: 560px) {
		.row {
			grid-template-columns: 52px minmax(0, 1fr);
		}
	}
	.row form {
		margin: 0;
	}
	.portrait {
		width: 52px;
		height: 52px;
		border-radius: var(--radius-sm);
		border: 1px solid var(--line-2);
		background: var(--ink-2);
		image-rendering: pixelated;
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}
	.ttl {
		font: 700 15px/1.2 var(--font-display);
		color: var(--fg-1);
		text-decoration: none;
	}
	.ttl:hover {
		color: var(--gold-200);
	}
	.sub {
		font: 500 12px/1.3 var(--font-ui);
		color: var(--fg-3);
		overflow-wrap: anywhere;
	}
	.pill {
		justify-self: start;
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
	.pill--ok {
		background: rgba(79, 179, 161, 0.16);
		color: var(--support-300);
	}
	.pill--no {
		background: rgba(240, 100, 140, 0.16);
		color: var(--status-error);
	}
	.pill--off {
		background: var(--ink-3);
		color: var(--fg-3);
	}
	.why {
		grid-column: 2 / -1;
		margin: 0;
		padding: 8px 10px;
		border-radius: var(--radius-sm);
		background: rgba(240, 100, 140, 0.08);
		color: var(--fg-2);
		font: 400 13px/1.4 var(--font-ui);
		overflow-wrap: anywhere;
	}
	.form-error {
		margin: 0;
		padding: 10px 12px;
		border-radius: var(--radius-sm);
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
		font: 500 13px/1.3 var(--font-ui);
	}
</style>
