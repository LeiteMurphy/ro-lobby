<script lang="ts">
	import '$lib/ui/tokens/index.css';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href="/brand/c1-symbol-dark.svg" type="image/svg+xml" />
</svelte:head>

{@render children()}
