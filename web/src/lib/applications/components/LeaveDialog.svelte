<script lang="ts">
	import { onMount } from 'svelte';
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/Button.svelte';

	interface Props {
		applicationId: string;
		/** Instância, dia e hora, para o cabeçalho. */
		title: string;
		/** Nome da função da vaga que fica livre; nulo no grupo livre (spec grupo-livre, RN-04). */
		roleLabel: string | null;
		message?: string | null;
		onclose: () => void;
	}

	let { applicationId, title, roleLabel, message = null, onclose }: Props = $props();
	const uid = $props.id();
	let dialog: HTMLDialogElement | undefined = $state();
	let submitting = $state(false);

	// RN-14 / RN-38: sair pede confirmação (Candidatura 2f).
	onMount(() => dialog?.showModal());
</script>

<dialog
	bind:this={dialog}
	class="dlg"
	aria-labelledby="{uid}-title"
	aria-describedby="{uid}-desc"
	{onclose}
	data-testid="leave-dialog"
>
	<form
		method="POST"
		action="?/leave"
		use:enhance={() => {
			submitting = true;
			return async ({ result, update }) => {
				submitting = false;
				await update({ reset: false });
				if (result.type === 'success') dialog?.close();
			};
		}}
	>
		<input type="hidden" name="applicationId" value={applicationId} />
		<div class="dlg-h">
			<h2 id="{uid}-title">Sair do grupo?</h2>
			<p class="hint">{title}</p>
		</div>
		<div class="dlg-b">
			{#if message}<p class="form-error" role="alert">{message}</p>{/if}
			<p class="desc" id="{uid}-desc">
				{roleLabel ? `Sua vaga de ${roleLabel}` : 'Sua vaga'} fica livre para outro jogador. Você pode
				se candidatar de novo enquanto houver vaga.
			</p>
		</div>
		<footer class="dlg-f">
			<Button variant="ghost" onclick={() => dialog?.close()}>Voltar</Button>
			<Button type="submit" variant="outline" class="danger" disabled={submitting}
				>Sair do grupo</Button
			>
		</footer>
	</form>
</dialog>

<style>
	.dlg {
		width: min(440px, calc(100vw - 32px));
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
		padding: 22px 22px 0;
	}
	h2 {
		margin: 0;
		font: var(--type-h3);
	}
	.hint {
		margin: 0;
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.dlg-b {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 18px 22px;
	}
	.desc {
		margin: 0;
		color: var(--fg-2);
		font: 400 14px/1.45 var(--font-ui);
	}
	.form-error {
		margin: 0;
		padding: 10px 12px;
		border-radius: var(--radius-sm);
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
		font: 500 13px/1.3 var(--font-ui);
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
