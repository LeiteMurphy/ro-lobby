<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import Button from '$lib/ui/Button.svelte';

	// spec lobbies, CA-03.2: lobby inexistente (e qualquer outra página que não existe) mostra
	// a página de não encontrado, no visual do RO Lobby.
	const notFound = $derived(page.status === 404);
	const title = $derived(notFound ? 'Página não encontrada' : 'Algo deu errado');
</script>

<svelte:head>
	<title>RO Lobby · {title}</title>
</svelte:head>

<div class="page">
	<TopBar user={page.data.user ?? null} loginHref="/auth/discord/login?next=%2F" />
	<main class="content" data-testid="error-page">
		<img src="/brand/c1-symbol-dark.svg" alt="" width="56" height="56" />
		<h1>{title}</h1>
		<p>{page.error?.message ?? 'Tente de novo em instantes.'}</p>
		<Button href={resolve('/')}>Voltar para os grupos</Button>
	</main>
</div>

<style>
	.page {
		min-height: 100dvh;
	}
	.content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		max-width: 560px;
		margin: 0 auto;
		padding: 72px 16px;
		text-align: center;
	}
	img {
		image-rendering: pixelated;
	}
	h1 {
		margin: 0;
		font: var(--type-h2);
	}
	p {
		margin: 0 0 8px;
		color: var(--fg-3);
	}
</style>
