<script lang="ts">
	import { enhance } from '$app/forms';
	import { ROLE_LABELS } from '$lib/home/types';
	import type { ApiLobby } from '$lib/lobbies/api';
	import { DELETED_CHARACTER } from '$lib/lobbies/toHome';
	import Button from '$lib/ui/Button.svelte';
	import type { LobbySwapRequest } from '../api';

	interface Props {
		swap: LobbySwapRequest;
		lobby: ApiLobby;
		onreject: (swap: LobbySwapRequest) => void;
	}

	let { swap, lobby, onreject }: Props = $props();

	// D-11: a vaga que sai conta como livre; para outra função, precisa de 1 vaga livre nela.
	const sameRole = $derived(swap.from.role === swap.to.role);
	const free = $derived(lobby.slots[swap.to.role] - lobby.occupied[swap.to.role]);
	const toLabel = $derived(ROLE_LABELS[swap.to.role]);
	let accepting = $state(false);
</script>

<!-- RN-22 / RN-38: o pedido escolhido, com o atual, o novo, o motivo e a decisão (Candidatura 2d). -->
<section class="panel" aria-labelledby="swap-title" aria-live="polite" data-testid="swap-panel">
	<span class="over" id="swap-title">Pedido de troca</span>
	<div class="swap">
		{#each [swap.from, swap.to] as c, i (i)}
			{#if i === 1}<span class="arrow" aria-label="para">→</span>{/if}
			<span class="char">
				{#if c.portrait}<img
						class="portrait"
						src="/portraits/{c.portrait}.svg"
						alt=""
						width="48"
						height="48"
					/>{/if}
				<span class="who">
					<span class="nm">{c.nick ?? DELETED_CHARACTER}</span>
					<span class="sub"
						>{[ROLE_LABELS[c.role], c.level ? `Nv ${c.level}` : null]
							.filter(Boolean)
							.join(' · ')}</span
					>
				</span>
			</span>
		{/each}
	</div>
	{#if swap.discordName}<div class="kv"><span>Membro</span><b>{swap.discordName}</b></div>{/if}
	<div class="kv">
		<span>Vaga de {toLabel}</span>
		<b>{sameRole ? 'a mesma do membro' : free === 1 ? '1 livre' : `${Math.max(free, 0)} livres`}</b>
	</div>
	<span class="over">Motivo</span>
	<p class="msg">{swap.reason}</p>
	<div class="acts">
		<form
			method="POST"
			action="?/acceptSwap"
			use:enhance={() => {
				accepting = true;
				return async ({ update }) => {
					accepting = false;
					await update();
				};
			}}
		>
			<input type="hidden" name="swapId" value={swap.id} />
			<Button type="submit" iconLeft="check" block disabled={accepting}>Aceitar</Button>
		</form>
		<Button variant="outline" class="danger" block onclick={() => onreject(swap)}>Recusar</Button>
	</div>
	{#if !sameRole && free < 1}
		<span class="hint"
			>Se a vaga de {toLabel} não abrir, o aceite falha e o pedido continua pendente.</span
		>
	{/if}
</section>

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
	.over {
		font: 600 10px/1.2 var(--font-ui);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--fg-3);
	}
	.swap {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
	}
	.char {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}
	.arrow {
		color: var(--gold-300);
		font: 700 18px/1 var(--font-ui);
	}
	.portrait {
		flex: none;
		border-radius: var(--radius-sm);
		border: 1px solid var(--line-2);
		image-rendering: pixelated;
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 3px;
		min-width: 0;
	}
	.nm {
		font: 700 15px/1.15 var(--font-display);
		color: var(--fg-1);
		overflow-wrap: anywhere;
	}
	.sub,
	.hint {
		font: 500 12px/1.3 var(--font-ui);
		color: var(--fg-3);
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
		overflow-wrap: anywhere;
	}
	.msg {
		margin: 0;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		background: var(--ink-2);
		border-left: 3px solid var(--gold-400);
		color: var(--fg-2);
		font: 400 13px/1.45 var(--font-ui);
		overflow-wrap: anywhere;
	}
	.acts {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.acts form {
		margin: 0;
		display: flex;
	}
	.acts :global(.danger) {
		color: var(--status-error);
		border-color: rgba(240, 100, 140, 0.4);
	}
</style>
