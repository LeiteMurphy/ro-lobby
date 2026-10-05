<script lang="ts">
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/Button.svelte';
	import IconButton from '$lib/ui/IconButton.svelte';
	import type { SessionUser } from '$lib/auth/api';
	import UserMenu from './UserMenu.svelte';

	interface Props {
		/** Usuário da sessão, ou null para visitante (spec login-discord). */
		user: SessionUser | null;
		/** Link de login que volta para a página atual (RN-12). */
		loginHref: string;
	}

	let { user, loginHref }: Props = $props();
</script>

<header class="topbar">
	<a href={resolve('/')} class="brand" aria-label="RO Lobby, página inicial">
		<img src="/brand/c1-symbol-dark.svg" alt="" width="34" height="34" />
		<span>RO Lobby</span>
	</a>
	<div class="actions">
		<span class="desktop"><Button iconLeft="plus" soon>Criar lobby</Button></span>
		<span class="mobile"><IconButton icon="plus" label="Criar lobby" variant="primary" soon /></span
		>
		{#if user}
			<UserMenu {user} />
		{:else}
			<!-- RN-14 (login-discord): o login já funciona; os outros controles seguem "em breve". -->
			<span class="desktop"
				><Button variant="secondary" iconLeft="log-in" href={loginHref}>Entrar com Discord</Button
				></span
			>
			<span class="mobile"
				><IconButton
					icon="log-in"
					label="Entrar com Discord"
					variant="secondary"
					href={loginHref}
				/></span
			>
		{/if}
	</div>
</header>

<style>
	.topbar {
		position: sticky;
		top: 0;
		z-index: 20;
		height: var(--topbar-h);
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 0 28px;
		background: rgba(11, 16, 23, 0.78);
		backdrop-filter: blur(var(--panel-blur));
		border-bottom: 1px solid var(--panel-border);
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
		color: var(--fg-1);
		text-decoration: none;
	}
	.brand img {
		image-rendering: pixelated;
	}
	.brand span {
		font: 700 21px/1 var(--font-display);
		letter-spacing: 0.01em;
		white-space: nowrap;
	}
	.actions {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.mobile {
		display: none;
	}
	@media (max-width: 899px) {
		.topbar {
			padding: 0 16px;
		}
		.desktop {
			display: none;
		}
		.mobile {
			display: inline-flex;
		}
	}
</style>
