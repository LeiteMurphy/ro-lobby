<script lang="ts">
	import { createLobbyHref } from '$lib/lobbies/time';
	import Button from '$lib/ui/Button.svelte';

	interface Props {
		/** RN-17: 'day' = dia sem grupos; 'filters' = nenhum grupo passa nos filtros. */
		kind: 'day' | 'filters';
		onreset: () => void;
		/** Dia escolhido na Home; a criação abre nele (RN-24 da spec lobbies). */
		createDate?: string | null;
	}

	let { kind, onreset, createDate = null }: Props = $props();
</script>

<div class="empty" data-testid="empty-state">
	<img src="/brand/c1-symbol-dark.svg" alt="" width="56" height="56" />
	<div class="text">
		{#if kind === 'filters'}
			<span class="title">Nenhum grupo com esses filtros</span>
			<span class="desc">Ajuste os filtros ou crie um lobby com a composição que você procura.</span
			>
		{:else}
			<span class="title">Nenhum grupo neste dia</span>
			<span class="desc">Crie o primeiro e os jogadores vão se candidatar.</span>
		{/if}
	</div>
	<div class="actions">
		{#if kind === 'filters'}
			<Button variant="secondary" onclick={onreset}>Limpar filtros</Button>
		{/if}
		<!-- RN-23 (lobbies): "Criar lobby" leva à criação, como o do cabeçalho. -->
		<Button iconLeft="plus" href={createLobbyHref(createDate)}>Criar lobby</Button>
	</div>
</div>

<style>
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 14px;
		padding: 56px 24px;
		background: var(--panel-bg);
		backdrop-filter: blur(var(--panel-blur));
		border: 1px dashed var(--line-2);
		border-radius: var(--radius-lg);
	}
	img {
		opacity: 0.9;
		image-rendering: pixelated;
	}
	.text {
		display: flex;
		flex-direction: column;
		gap: 6px;
		max-width: 420px;
	}
	.title {
		font: var(--type-title);
		color: var(--fg-1);
	}
	.desc {
		font: var(--type-body);
		color: var(--fg-3);
		text-wrap: pretty;
	}
	.actions {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		justify-content: center;
	}
</style>
