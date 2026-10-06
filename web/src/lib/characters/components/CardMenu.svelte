<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '$lib/ui/Icon.svelte';

	interface Props {
		id: string;
		nick: string;
		isMain: boolean;
		onedit: () => void;
		ondelete: () => void;
	}

	let { id, nick, isMain, onedit, ondelete }: Props = $props();
	const menuId = $props.id();

	let open = $state(false);
	let root: HTMLDivElement | undefined = $state();
	let trigger: HTMLButtonElement | undefined = $state();
	let menu: HTMLDivElement | undefined = $state();

	const items = () => [...(menu?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])];

	// RNF-01: o menu abre pelo teclado com o foco no primeiro item, as setas andam entre os
	// itens, e Escape fecha devolvendo o foco ao botão "…" (mesmo padrão do UserMenu).
	async function toggle() {
		open = !open;
		if (open) {
			await Promise.resolve();
			items()[0]?.focus();
		}
	}

	function close(returnFocus = true) {
		open = false;
		if (returnFocus) trigger?.focus();
	}

	function onkeydown(event: KeyboardEvent) {
		if (!open) return;
		if (!root?.contains(document.activeElement)) return;
		if (event.key === 'Escape') {
			event.preventDefault();
			close();
			return;
		}
		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
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

	function pick(action: () => void) {
		close(false);
		action();
	}
</script>

<svelte:window {onkeydown} {onpointerdown} />

<div class="wrap" bind:this={root}>
	<button
		type="button"
		class="trigger"
		class:on={open}
		bind:this={trigger}
		aria-haspopup="menu"
		aria-expanded={open}
		aria-controls={menuId}
		aria-label="Ações de {nick}"
		title="Ações de {nick}"
		onclick={toggle}
	>
		<Icon name="ellipsis" size={16} />
	</button>
	{#if open}
		<div class="menu" id={menuId} role="menu" aria-label="Ações de {nick}" bind:this={menu}>
			{#if !isMain}
				<form method="POST" action="?/main" use:enhance={() => () => close(false)}>
					<input type="hidden" name="id" value={id} />
					<button type="submit" role="menuitem" class="item">
						<Icon name="star" size={15} color="var(--gold-300)" />Tornar principal
					</button>
				</form>
			{/if}
			<button type="button" role="menuitem" class="item" onclick={() => pick(onedit)}>
				<Icon name="pencil" size={15} color="var(--fg-2)" />Editar
			</button>
			<span class="sep" role="separator"></span>
			<button type="button" role="menuitem" class="item danger" onclick={() => pick(ondelete)}>
				<Icon name="trash-2" size={15} />Excluir
			</button>
		</div>
	{/if}
</div>

<style>
	.wrap {
		position: relative;
	}
	.trigger {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: var(--radius-sm);
		border: 1px solid transparent;
		background: transparent;
		color: var(--fg-2);
		cursor: pointer;
	}
	.trigger:hover,
	.trigger.on {
		background: var(--ink-3);
		border-color: var(--line-2);
		color: var(--fg-1);
	}
	.menu {
		position: absolute;
		right: 0;
		bottom: calc(100% + 6px);
		z-index: var(--z-tooltip);
		width: 190px;
		display: flex;
		flex-direction: column;
		padding: 4px;
		background: var(--ink-2);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-pop);
	}
	form {
		margin: 0;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 9px;
		width: 100%;
		height: 34px;
		padding: 0 10px;
		background: transparent;
		border: 0;
		border-radius: var(--radius-sm);
		color: var(--fg-1);
		font: 500 13px/1 var(--font-ui);
		text-align: left;
		cursor: pointer;
	}
	.item:hover,
	.item:focus-visible {
		background: var(--ink-3);
	}
	.item.danger {
		color: var(--status-error);
	}
	.sep {
		height: 1px;
		margin: 4px 6px;
		background: var(--border-subtle);
	}
</style>
