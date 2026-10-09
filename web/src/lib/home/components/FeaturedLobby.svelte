<script lang="ts">
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/Button.svelte';
	import { instanceArt } from '../catalog';
	import { pendingFor } from '../lobbies';
	import type { Lobby } from '../types';
	import HostLine from './HostLine.svelte';
	import FreeComposition from './FreeComposition.svelte';
	import RoleComposition from './RoleComposition.svelte';

	interface Props {
		lobby: Lobby;
		relative: string;
		/** Usuário da sessão; o dono vê o selo de pendentes (RN-34 da candidatura-lobby). */
		viewerId?: string | null;
	}

	let { lobby, relative, viewerId = null }: Props = $props();
	const pending = $derived(pendingFor(lobby, viewerId));
	const art = $derived(instanceArt(lobby.instance));
</script>

<section class="hero" data-testid="featured-lobby" aria-label="Próximo grupo com vaga">
	<div class="glass">
		<div class="cover" style="background:{art.cover}">
			<img src={art.icon} alt="" width="96" height="96" />
		</div>
		<div class="info">
			<div class="eyebrow-row">
				<span class="eyebrow">Próximo grupo com vaga</span>
				<span class="rule"></span>
				<span class="time">{lobby.time}</span>
				{#if relative}<span class="relative">{relative}</span>{/if}
				{#if pending}<span class="pending" data-testid="pending-badge">{pending}</span>{/if}
			</div>
			<h2>{lobby.instance}</h2>
			<HostLine host={lobby.host} hostClass={lobby.hostClass} minLevel={lobby.minLevel} size="lg" />
			{#if lobby.free}
				<FreeComposition seats={lobby.free} showBar />
			{:else}
				<RoleComposition composition={lobby.composition} showTotal showBar />
			{/if}
			<div class="actions">
				<!-- RN-35 da candidatura-lobby: a candidatura acontece no detalhe do lobby. -->
				<Button size="lg" iconLeft="user-plus" href={resolve('/lobbies/[id]', { id: lobby.id })}
					>Candidatar</Button
				>
				<!-- spec lobbies, RN-23: "Ver grupo" abre o detalhe. -->
				<Button size="lg" variant="secondary" href={resolve('/lobbies/[id]', { id: lobby.id })}
					>Ver grupo</Button
				>
			</div>
		</div>
	</div>
</section>

<style>
	.pending {
		display: inline-flex;
		align-items: center;
		height: 22px;
		padding: 0 8px;
		border-radius: 999px;
		background: var(--gold-400);
		color: var(--on-accent);
		font: 700 11px/1 var(--font-ui);
	}
	.hero {
		position: relative;
		border-radius: var(--radius-xl);
		overflow: hidden;
		border: 1px solid rgba(255, 210, 122, 0.35);
		padding: 28px;
		background:
			linear-gradient(
				180deg,
				rgba(11, 16, 23, 0) 0%,
				rgba(11, 16, 23, 0) 58%,
				rgba(11, 16, 23, 0.88) 58%
			),
			radial-gradient(70% 120% at 0% 0%, #ffe3a6 0%, rgba(255, 227, 166, 0) 60%),
			radial-gradient(60% 140% at 100% 20%, #3b78eb 0%, rgba(59, 120, 235, 0) 70%),
			linear-gradient(120deg, #ffd27a 0%, #f6bb45 40%, #e6a224 62%, #5b6fc4 100%);
	}
	.glass {
		display: flex;
		flex-wrap: wrap;
		gap: 24px;
		align-items: stretch;
		background: rgba(11, 16, 23, 0.64);
		backdrop-filter: blur(16px);
		border: 1px solid rgba(233, 238, 245, 0.14);
		border-radius: var(--radius-lg);
		padding: 22px;
		box-shadow: 0 18px 40px rgba(0, 0, 0, 0.35);
	}
	.cover {
		flex: none;
		width: 220px;
		min-height: 168px;
		border-radius: var(--radius-md);
		border: 1px solid rgba(233, 238, 245, 0.12);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}
	.cover img {
		filter: drop-shadow(0 6px 14px rgba(0, 0, 0, 0.45));
	}
	.info {
		flex: 1;
		min-width: 260px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.eyebrow-row {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}
	.eyebrow {
		font: var(--type-overline);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--gold-300);
	}
	.rule {
		height: 1px;
		flex: 1;
		min-width: 24px;
		background: rgba(233, 238, 245, 0.14);
	}
	.time {
		font: 700 22px/1 var(--font-ui);
		color: var(--gold-200);
	}
	.relative {
		font: var(--type-caption);
		color: var(--fg-2);
	}
	h2 {
		margin: 0;
		font: var(--type-h2);
		color: var(--fg-1);
		text-wrap: balance;
	}
	.actions {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		margin-top: auto;
	}
	@media (max-width: 899px) {
		.hero {
			padding: 14px;
		}
		.cover {
			width: 100%;
		}
	}
</style>
