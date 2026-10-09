<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import Button from '$lib/ui/Button.svelte';

	interface Props {
		/** O convite de três linhas (RN-02). */
		invite: string;
	}

	let { invite }: Props = $props();
	const uid = $props.id();
	let copied = $state(false);
	let fallbackOpen = $state(false);
	let dialog: HTMLDialogElement | undefined = $state();
	let textarea: HTMLTextAreaElement | undefined = $state();
	let timer: ReturnType<typeof setTimeout> | undefined;

	// RN-04: copia e mostra "Convite copiado" por alguns segundos; se o navegador recusar,
	// abre o diálogo com o convite selecionado para copiar na mão.
	async function share() {
		try {
			await navigator.clipboard.writeText(invite);
			copied = true;
			clearTimeout(timer);
			timer = setTimeout(() => (copied = false), 3000);
		} catch {
			fallbackOpen = true;
			await tick();
			dialog?.showModal();
			textarea?.select();
		}
	}

	onDestroy(() => clearTimeout(timer));
</script>

<Button variant="secondary" iconLeft="share-2" onclick={share}
	>{copied ? 'Convite copiado' : 'Compartilhar'}</Button
>
<span class="sr" role="status">{copied ? 'Convite copiado' : ''}</span>

{#if fallbackOpen}
	<dialog
		bind:this={dialog}
		class="dlg"
		aria-labelledby="{uid}-title"
		aria-describedby="{uid}-desc"
		onclose={() => (fallbackOpen = false)}
		data-testid="share-dialog"
	>
		<div class="dlg-h">
			<h2 id="{uid}-title">Copie o convite</h2>
			<p id="{uid}-desc">O navegador não deixou copiar. Use Ctrl+C ou segure para copiar.</p>
		</div>
		<div class="dlg-b">
			<textarea
				bind:this={textarea}
				class="input"
				rows="4"
				readonly
				aria-labelledby="{uid}-title"
				value={invite}></textarea>
		</div>
		<footer class="dlg-f">
			<Button variant="ghost" onclick={() => dialog?.close()}>Fechar</Button>
		</footer>
	</dialog>
{/if}

<style>
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
	.dlg {
		width: min(460px, calc(100vw - 32px));
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
	.dlg-h p {
		margin: 0;
		color: var(--fg-2);
	}
	.dlg-b {
		display: flex;
		flex-direction: column;
		padding: 18px 22px;
	}
	.input {
		padding: 10px 12px;
		background: var(--surface-input);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-sm);
		color: var(--fg-1);
		font: 400 14px/1.4 var(--font-ui);
		resize: none;
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
</style>
