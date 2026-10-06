<script lang="ts">
	import { resolve } from '$app/paths';
	import Badge from '$lib/ui/Badge.svelte';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import { instanceArt } from '../catalog';
	import { edgeRole, headcount, isFull } from '../lobbies';
	import type { Lobby } from '../types';
	import HostLine from './HostLine.svelte';
	import RoleComposition from './RoleComposition.svelte';

	interface Props {
		lobby: Lobby;
		relative: string;
	}

	let { lobby, relative }: Props = $props();

	const art = $derived(instanceArt(lobby.instance));
	const full = $derived(isFull(lobby.composition));
	const edge = $derived(edgeRole(lobby.composition));
	const count = $derived(headcount(lobby.composition));
</script>

<article
	class="card"
	class:full
	data-testid="lobby-card"
	data-edge={full ? 'full' : edge}
	aria-label="{lobby.instance} às {lobby.time}"
>
	<span class="edge" style="background:{full ? 'var(--fg-4)' : `var(--${edge}-400)`}"></span>
	<div class="cover" style="background:{art.cover}">
		<div class="when">
			<span class="time">{lobby.time}</span>
			{#if relative}<span class="relative">{relative}</span>{/if}
		</div>
		<img src={art.icon} alt="" width="68" height="68" />
	</div>
	<div class="body">
		<div class="title-row">
			<!-- spec lobbies, RN-23: o nome da instância leva ao detalhe ("Ver grupo"). -->
			<h3>
				<a
					class="detail"
					href={resolve('/lobbies/[id]', { id: lobby.id })}
					aria-label="Ver grupo: {lobby.instance} às {lobby.time}">{lobby.instance}</a
				>
			</h3>
			<span class="headcount" title="Vagas ocupadas">
				<Icon name="users" size={14} color="var(--fg-3)" />{count.filled}/{count.total}
			</span>
		</div>
		<HostLine host={lobby.host} hostClass={lobby.hostClass} minLevel={lobby.minLevel} />
		<div class="footer">
			<RoleComposition composition={lobby.composition} />
			<div class="action">
				{#if full}
					<Badge icon="lock">Lotado</Badge>
				{:else}
					<Button size="sm" variant="outline" iconLeft="user-plus" soon>Candidatar</Button>
				{/if}
			</div>
		</div>
	</div>
</article>

<style>
	.detail {
		color: inherit;
		text-decoration: none;
	}
	.detail:hover {
		color: var(--gold-200);
		text-decoration: underline;
	}
	.card {
		position: relative;
		display: flex;
		flex-direction: column;
		background: var(--surface-card);
		backdrop-filter: blur(var(--panel-blur));
		border: 1px solid var(--border-default);
		border-radius: var(--radius-md);
		overflow: hidden;
		transition:
			background var(--dur-fast) var(--ease-out),
			border-color var(--dur-fast) var(--ease-out),
			box-shadow var(--dur-fast) var(--ease-out);
	}
	.card:hover {
		background: var(--surface-raised);
		border-color: var(--border-strong);
		box-shadow: var(--shadow-hover);
	}
	.card.full {
		opacity: 0.55;
	}
	.edge {
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		width: var(--card-accent-w);
		z-index: 1;
	}
	.cover {
		position: relative;
		height: 92px;
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		padding: 12px 14px 12px 17px;
		overflow: hidden;
	}
	.full .cover {
		filter: grayscale(0.85) brightness(0.8);
	}
	.when {
		display: flex;
		flex-direction: column;
		gap: 4px;
		position: relative;
		z-index: 1;
	}
	.time {
		font: 700 22px/1 var(--font-ui);
		color: var(--fg-1);
		text-shadow: 0 1px 8px rgba(0, 0, 0, 0.45);
	}
	.relative {
		font: 600 12px/1 var(--font-ui);
		color: rgba(233, 238, 245, 0.78);
	}
	.cover img {
		position: absolute;
		right: 10px;
		top: 50%;
		transform: translateY(-50%);
		filter: drop-shadow(0 4px 10px rgba(0, 0, 0, 0.4));
	}
	.body {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 14px 16px 14px 17px;
	}
	.title-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	h3 {
		margin: 0;
		flex: 1;
		min-width: 0;
		font: 600 16px/1.3 var(--font-ui);
		color: var(--fg-1);
		text-wrap: pretty;
	}
	.headcount {
		display: flex;
		align-items: center;
		gap: 5px;
		font: 600 12px/1 var(--font-ui);
		color: var(--fg-2);
	}
	.footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
		margin-top: auto;
		padding-top: 10px;
		border-top: 1px solid var(--border-subtle);
	}
	.action {
		margin-left: auto;
		display: flex;
	}
</style>
