<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { LobbyDiscord } from '../api';
	import { cooldownMessage } from '../messages';

	// Seção "Sala no Discord" do detalhe do lobby, só para os membros (spec discord-bot,
	// RN-18, RN-19, RN-24, RN-26, RN-27).

	interface Props {
		discord: LobbyDiscord;
		isOwner: boolean;
		/** Lobby aberto: só ele deixa abrir a sala (RN-06). */
		lobbyOpen: boolean;
		/** Resultado da última action desta seção. */
		result?: { text: string; ok: boolean } | null;
		/** Agora, em ISO, para o tempo que falta até poder avisar de novo (RN-27). */
		now: string;
	}

	let { discord, isOwner, lobbyOpen, result = null, now }: Props = $props();

	const room = $derived(discord.room);
	const waitSeconds = $derived(
		room?.alertAvailableAt
			? Math.max(0, (Date.parse(room.alertAvailableAt) - Date.parse(now)) / 1000)
			: 0
	);
	let submitting = $state(false);
	const show = $derived(discord.enabled || !!discord.ownerLink);
</script>

{#if show}
	<section class="panel" aria-labelledby="discord-title" data-testid="discord-section">
		<h2 id="discord-title">Sala no Discord</h2>

		{#if result}
			<p class="msg" class:err={!result.ok} role={result.ok ? 'status' : 'alert'}>{result.text}</p>
		{/if}

		{#if discord.enabled}
			{#if !room}
				{#if isOwner && lobbyOpen}
					<!-- RN-06: só o dono abre, com o lobby aberto. -->
					<p class="hint">
						Abra uma sala privada no servidor do RO Lobby, só com os membros do grupo, com texto e
						voz.
					</p>
					<form
						method="POST"
						action="?/openRoom"
						use:enhance={() => {
							submitting = true;
							return async ({ update }) => {
								submitting = false;
								await update();
							};
						}}
					>
						<Button type="submit" iconLeft="message-circle" disabled={submitting}
							>Abrir sala no Discord</Button
						>
					</form>
				{:else}
					<p class="hint" data-testid="room-not-open">A sala ainda não foi aberta.</p>
				{/if}
			{:else if room.status === 'open' && room.url}
				<!-- eslint-disable svelte/no-navigation-without-resolve -- link do Discord -->
				<a
					class="room-link"
					href={room.url}
					target="_blank"
					rel="noopener noreferrer"
					data-testid="room-link"
				>
					<Icon name="external-link" size={15} />Abrir a sala no Discord
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
				<ul class="members" aria-label="Quem está na sala">
					{#each discord.members as m (m.userId)}
						<li>
							<span>{m.nick}</span>
							<span class="state" class:on={m.connected}
								>{m.connected ? 'na sala' : 'não conectado'}</span
							>
						</li>
					{/each}
				</ul>
				{#if !discord.viewerConnected}
					<!-- RN-18: o membro não conectado vê o atalho para se conectar. -->
					<Button variant="secondary" href={resolve('/perfil/discord')}
						>Entrar no servidor do RO Lobby</Button
					>
				{/if}
				{#if isOwner}
					<!-- RN-26 / RN-27: só o dono avisa, com intervalo de 10 minutos. -->
					<form method="POST" action="?/alert" use:enhance>
						<Button
							type="submit"
							variant="outline"
							iconLeft="megaphone"
							disabled={waitSeconds > 0}
							describedby={waitSeconds > 0 ? 'alert-wait' : undefined}>Avisar no Discord</Button
						>
						{#if waitSeconds > 0}
							<span class="hint" id="alert-wait" data-testid="alert-wait"
								>{cooldownMessage(waitSeconds)}</span
							>
						{/if}
					</form>
				{/if}
			{:else}
				<p class="hint" data-testid="room-closed">A sala foi encerrada.</p>
			{/if}
		{/if}

		{#if discord.ownerLink}
			<!-- RN-23 / RN-24: o Discord do anfitrião, só para os membros. -->
			<!-- eslint-disable svelte/no-navigation-without-resolve -- link informado pelo dono -->
			<a
				class="room-link"
				href={discord.ownerLink}
				target="_blank"
				rel="noopener noreferrer"
				data-testid="owner-link"
			>
				<Icon name="external-link" size={15} />Discord do anfitrião
			</a>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		{/if}
	</section>
{/if}

<style>
	.panel {
		display: flex;
		flex-direction: column;
		gap: 14px;
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
	.hint {
		margin: 0;
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.msg {
		margin: 0;
		padding: 8px 12px;
		border-radius: var(--radius-sm);
		background: rgba(80, 200, 140, 0.12);
		color: var(--support-300);
		font: 500 13px/1.3 var(--font-ui);
	}
	.msg.err {
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 6px;
		margin: 0;
	}
	.room-link {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		color: var(--gold-300);
		font: 600 14px/1.2 var(--font-ui);
	}
	.members {
		display: flex;
		flex-direction: column;
		gap: 6px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.members li {
		display: flex;
		justify-content: space-between;
		gap: 8px;
		font: 500 14px/1.3 var(--font-ui);
	}
	.state {
		color: var(--fg-3);
		font: var(--type-caption);
	}
	.state.on {
		color: var(--support-300);
	}
</style>
