<script lang="ts">
	import { resolve } from '$app/paths';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import LobbyForm from '$lib/lobbies/components/LobbyForm.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>RO Lobby · editar lobby</title>
</svelte:head>

<div class="page">
	<TopBar user={data.user} loginHref={data.loginHref} />
	<main class="content">
		<div class="titles">
			<h1>Editar lobby</h1>
			<span class="hint">O seu personagem fica como está.</span>
		</div>
		{#key form}
			<LobbyForm
				mode="update"
				instances={data.instances}
				characters={data.characters}
				classes={data.classes}
				days={data.days}
				values={form?.values ?? data.values}
				errors={form?.errors}
				message={form?.message}
				cancelHref={resolve('/lobbies/[id]', { id: data.lobby.id })}
			/>
		{/key}
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
</style>
