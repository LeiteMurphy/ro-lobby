<script lang="ts">
	import { statusMessages } from '$lib/api/health';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>RO Lobby · status</title>
</svelte:head>

<main>
	<h1>RO Lobby</h1>
	<p class="status status-{data.apiStatus}" data-testid="api-status">
		{statusMessages[data.apiStatus]}
	</p>
</main>

<style>
	main {
		font-family: system-ui, sans-serif;
		max-width: 32rem;
		margin: 4rem auto;
		padding: 0 1rem;
	}

	.status {
		display: inline-block;
		padding: 0.5rem 1rem;
		border-radius: 0.5rem;
		font-weight: 600;
	}

	.status-online {
		background: #dcfce7;
		color: #166534;
	}

	.status-degraded {
		background: #fef9c3;
		color: #854d0e;
	}

	.status-unavailable {
		background: #fee2e2;
		color: #991b1b;
	}
</style>
