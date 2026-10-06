<script lang="ts">
	import { resolve } from '$app/paths';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import LobbyForm from '$lib/lobbies/components/LobbyForm.svelte';
	import { NO_CHARACTER_MESSAGE } from '$lib/lobbies/messages';
	import Button from '$lib/ui/Button.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>RO Lobby · criar lobby</title>
</svelte:head>

<div class="page">
	<TopBar user={data.user} loginHref={data.loginHref} />
	<main class="content">
		<div class="titles">
			<h1>Criar lobby</h1>
			<span class="hint">Horários no horário de Brasília. O grupo sai da lista quando começa.</span>
		</div>

		{#if data.loadError}
			<p class="alert" role="alert">{data.loadError}</p>
		{:else if data.characters.length === 0}
			<!-- RN-12 / CA-01.10 -->
			<section class="empty">
				<h2>{NO_CHARACTER_MESSAGE}</h2>
				<p>O lobby usa um dos seus personagens, que ocupa a vaga da função dele.</p>
				<Button iconLeft="plus" href={resolve('/perfil')}>Ir para o perfil</Button>
			</section>
		{:else}
			{#key form}
				<LobbyForm
					mode="create"
					instances={data.instances}
					characters={data.characters}
					classes={data.classes}
					days={data.days}
					values={form?.values ?? data.values}
					errors={form?.errors}
					message={form?.message}
					cancelHref={resolve('/')}
				/>
			{/key}
		{/if}
	</main>
</div>

<style>
	.page {
		min-height: 100dvh;
		background: var(--ink-0);
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
	.titles {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	h1 {
		margin: 0;
		font: var(--type-h2);
	}
	.hint {
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.alert {
		margin: 0;
		padding: 10px 14px;
		border-radius: var(--radius-sm);
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
	}
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 48px 24px;
		text-align: center;
		background: var(--panel-bg);
		border: 1px dashed var(--line-2);
		border-radius: var(--radius-lg);
	}
	.empty h2 {
		margin: 0;
		font: var(--type-title);
	}
	.empty p {
		margin: 0;
		color: var(--fg-3);
	}
</style>
