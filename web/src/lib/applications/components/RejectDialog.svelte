<script lang="ts">
	import { onMount } from 'svelte';
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';

	interface Props {
		applicationId: string;
		nick: string;
		/** Valor e erros de uma tentativa anterior desta abertura (RN-09). */
		reason?: string;
		error?: string | null;
		message?: string | null;
		onclose: () => void;
	}

	let { applicationId, nick, reason = '', error = null, message = null, onclose }: Props = $props();
	const uid = $props.id();
	let dialog: HTMLDialogElement | undefined = $state();
	let submitting = $state(false);

	// RN-09: recusar pede justificativa de 10 a 250 caracteres. O foco começa no campo.
	onMount(() => dialog?.showModal());
</script>

<dialog
	bind:this={dialog}
	class="dlg"
	aria-labelledby="{uid}-title"
	aria-describedby="{uid}-desc"
	{onclose}
	data-testid="reject-dialog"
>
	<form
		method="POST"
		action="?/reject"
		novalidate
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
			<h2 id="{uid}-title">Recusar {nick}?</h2>
			<p id="{uid}-desc">A justificativa aparece só para você e para quem se candidatou.</p>
		</div>
		<div class="dlg-b">
			{#if message}<p class="form-error" role="alert">{message}</p>{/if}
			<label class="lbl" for="{uid}-reason">Justificativa</label>
			<textarea
				id="{uid}-reason"
				class="input"
				class:err={error}
				name="reason"
				rows="3"
				maxlength="250"
				value={reason}
				aria-invalid={error ? 'true' : undefined}
				aria-describedby={error ? `${uid}-reason-err` : undefined}></textarea>
			{#if error}<span class="hint-err" id="{uid}-reason-err"
					><Icon name="circle-x" size={14} />{error}</span
				>{/if}
		</div>
		<footer class="dlg-f">
			<Button variant="ghost" onclick={() => dialog?.close()}>Voltar</Button>
			<Button type="submit" variant="outline" class="danger" disabled={submitting}>Recusar</Button>
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
		overflow-wrap: anywhere;
	}
	.dlg-h p {
		margin: 0;
		color: var(--fg-2);
	}
	.dlg-b {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 18px 22px;
	}
	.lbl {
		font: 600 12px/1.2 var(--font-ui);
		color: var(--fg-2);
	}
	.input {
		padding: 10px 12px;
		background: var(--surface-input);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-sm);
		color: var(--fg-1);
		font: 400 14px/1.4 var(--font-ui);
		resize: vertical;
	}
	.input.err {
		border-color: var(--status-error);
	}
	.hint-err {
		display: flex;
		align-items: center;
		gap: 6px;
		font: var(--type-caption);
		color: var(--status-error);
	}
	.form-error {
		margin: 0 0 6px;
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
