<script lang="ts">
	import { onMount } from 'svelte';
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/Button.svelte';

	interface Props {
		id: string;
		nick: string;
		onclose: () => void;
	}

	let { id, nick, onclose }: Props = $props();
	const uid = $props.id();
	let dialog: HTMLDialogElement | undefined = $state();
	let submitting = $state(false);

	// RN-15: excluir é definitivo, então pede confirmação. O foco começa em "Cancelar".
	onMount(() => {
		dialog?.showModal();
	});
</script>

<dialog
	bind:this={dialog}
	class="dlg"
	aria-labelledby="{uid}-title"
	aria-describedby="{uid}-desc"
	{onclose}
	data-testid="confirm-dialog"
>
	<form
		method="POST"
		action="?/delete"
		use:enhance={() => {
			submitting = true;
			return async ({ update }) => {
				submitting = false;
				await update();
				dialog?.close();
			};
		}}
	>
		<input type="hidden" name="id" value={id} />
		<div class="dlg-h">
			<h2 id="{uid}-title">Excluir {nick}?</h2>
			<p id="{uid}-desc">O personagem sai do seu perfil, e o nick fica livre para outra pessoa.</p>
		</div>
		<footer class="dlg-f">
			<!-- O showModal leva o foco ao primeiro botão: "Cancelar", que não apaga nada. -->
			<Button variant="ghost" onclick={() => dialog?.close()}>Cancelar</Button>
			<Button type="submit" variant="outline" class="danger" disabled={submitting}>Excluir</Button>
		</footer>
	</form>
</dialog>

<style>
	.dlg {
		width: min(420px, calc(100vw - 32px));
		padding: 0;
		border: 0;
		border-radius: var(--radius-lg);
		background: var(--surface-raised);
		color: var(--fg-1);
		box-shadow: var(--shadow-pop);
	}
	.dlg::backdrop {
		background: rgba(5, 8, 12, 0.72);
	}
	form {
		margin: 0;
	}
	.dlg-h {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 22px 22px 18px;
	}
	h2 {
		margin: 0;
		font: var(--type-h3);
		overflow-wrap: anywhere;
	}
	p {
		margin: 0;
		color: var(--fg-2);
	}
	.dlg-f {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		padding: 14px 22px;
		border-top: 1px solid var(--border-default);
		background: var(--ink-2);
		border-radius: 0 0 var(--radius-lg) var(--radius-lg);
	}
	.dlg-f :global(.danger) {
		color: var(--status-error);
		border-color: rgba(240, 100, 140, 0.4);
	}
</style>
