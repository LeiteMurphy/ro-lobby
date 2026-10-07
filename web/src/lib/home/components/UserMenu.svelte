<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '$lib/ui/Icon.svelte';
	import { displayName, initial } from '$lib/auth/display';
	import type { SessionUser } from '$lib/auth/api';

	interface Props {
		user: SessionUser;
	}

	let { user }: Props = $props();
	const name = $derived(displayName(user));
	const menuId = $props.id();

	let open = $state(false);
	let root: HTMLDivElement | undefined = $state();
	let trigger: HTMLButtonElement | undefined = $state();
	let menu: HTMLDivElement | undefined = $state();

	const items = () => [...(menu?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])];

	// RNF-02: o menu abre pelo teclado com o foco no primeiro item ("Meu perfil"), as setas
	// andam entre os itens e Escape fecha.
	async function toggle() {
		open = !open;
		if (open) {
			await Promise.resolve();
			items()[0]?.focus();
		}
	}

	function onkeydown(event: KeyboardEvent) {
		if (!open) return;
		if (event.key === 'Escape') {
			open = false;
			trigger?.focus();
		} else if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();
			const list = items();
			const at = list.indexOf(document.activeElement as HTMLElement);
			const step = event.key === 'ArrowDown' ? 1 : -1;
			list[(at + step + list.length) % list.length]?.focus();
		}
	}

	function onpointerdown(event: PointerEvent) {
		if (open && root && !root.contains(event.target as Node)) open = false;
	}
</script>

<svelte:window {onkeydown} {onpointerdown} />

<div class="user" bind:this={root} data-testid="user-menu">
	<button
		type="button"
		class="trigger"
		bind:this={trigger}
		aria-haspopup="menu"
		aria-expanded={open}
		aria-controls={menuId}
		aria-label="Menu de {name}"
		onclick={toggle}
	>
		<span class="tile" aria-hidden="true">{initial(name)}</span>
		<span class="name">{name}</span>
		<Icon name="chevron-down" size={14} color="var(--fg-3)" />
	</button>
	<!-- O menu fica no HTML e só aparece quando aberto (hidden some também para leitores de tela). -->
	<div class="menu" id={menuId} role="menu" aria-label="Conta" hidden={!open} bind:this={menu}>
		<!-- spec personagens, RN-18: "Meu perfil" acima de "Sair". -->
		<a href={resolve('/perfil')} role="menuitem" class="item" onclick={() => (open = false)}>
			<Icon name="user" size={15} />Meu perfil
		</a>
		<!-- RN-33 da candidatura-lobby: "Minhas candidaturas" logo abaixo do perfil. -->
		<a href={resolve('/candidaturas')} role="menuitem" class="item" onclick={() => (open = false)}>
			<Icon name="users" size={15} />Minhas candidaturas
		</a>
		<form method="POST" action="/auth/logout">
			<button type="submit" role="menuitem" class="item item--logout">
				<Icon name="log-in" size={15} />Sair
			</button>
		</form>
	</div>
</div>

<style>
	.user {
		position: relative;
	}
	.trigger {
		display: flex;
		align-items: center;
		gap: 8px;
		height: 38px;
		padding: 0 8px 0 4px;
		background: transparent;
		border: 1px solid transparent;
		border-radius: var(--radius-sm);
		color: var(--fg-2);
		cursor: pointer;
		font: 600 13px/1 var(--font-ui);
	}
	.trigger:hover {
		background: var(--ink-3);
		border-color: var(--line-2);
	}
	.tile {
		width: 30px;
		height: 30px;
		border-radius: var(--radius-sm);
		background: var(--gradient-featured);
		border: 1px solid var(--accent-line);
		color: var(--gold-200);
		display: flex;
		align-items: center;
		justify-content: center;
		font: 700 15px/1 var(--font-display);
	}
	.name {
		color: var(--fg-1);
		max-width: 180px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.menu {
		position: absolute;
		top: calc(100% + 6px);
		right: 0;
		z-index: var(--z-tooltip);
		min-width: 160px;
		padding: 6px;
		background: var(--surface-raised);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-pop);
	}
	form {
		margin: 0;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 8px 10px;
		background: transparent;
		border: 0;
		border-radius: var(--radius-sm);
		color: var(--fg-1);
		font: 500 14px/1 var(--font-ui);
		text-align: left;
		text-decoration: none;
		cursor: pointer;
	}
	.menu[hidden] {
		display: none;
	}
	.item:hover {
		background: var(--surface-hover);
		text-decoration: none;
	}
	.item--logout :global(svg) {
		transform: scaleX(-1);
	}
	@media (max-width: 899px) {
		.name {
			display: none;
		}
	}
</style>
