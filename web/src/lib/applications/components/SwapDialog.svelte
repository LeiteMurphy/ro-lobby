<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { ROLE_LABELS } from '$lib/home/types';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { Eligibility } from '../detail';

	interface Props {
		/**
		 * `request`: o membro pede a troca, com motivo, e o anfitrião decide (RN-20, Candidatura
		 * 2i). `owner`: o anfitrião troca o próprio personagem, sem aprovação (RN-19, 2h).
		 */
		mode: 'request' | 'owner';
		/** Linha abaixo do título. */
		hint: string;
		options: Eligibility[];
		classNames: Record<string, string>;
		/** A candidatura do membro, no modo `request`. */
		applicationId?: string;
		/** Valores e erros de uma tentativa anterior desta abertura. */
		characterId?: string;
		reason?: string;
		errors?: Record<string, string>;
		formError?: string | null;
		onclose: () => void;
	}

	let {
		mode,
		hint,
		options,
		classNames,
		applicationId = '',
		characterId = '',
		reason = '',
		errors = {},
		formError = null,
		onclose
	}: Props = $props();
	const uid = $props.id();
	let dialog: HTMLDialogElement | undefined = $state();
	let submitting = $state(false);
	let chosen = $state(
		untrack(() => characterId || (options.find((o) => o.ok)?.character.id ?? ''))
	);
	const anyOk = $derived(options.some((o) => o.ok));

	onMount(() => dialog?.showModal());
</script>

<dialog
	bind:this={dialog}
	class="dlg"
	aria-labelledby="{uid}-title"
	aria-describedby="{uid}-desc"
	{onclose}
	data-testid="swap-dialog"
>
	<form
		method="POST"
		action={mode === 'request' ? '?/requestSwap' : '?/ownerSwap'}
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
		{#if mode === 'request'}<input type="hidden" name="applicationId" value={applicationId} />{/if}
		<div class="dlg-h">
			<h2 id="{uid}-title">
				{mode === 'request' ? 'Pedir troca de personagem' : 'Trocar seu personagem'}
			</h2>
			<p id="{uid}-desc">{hint}</p>
		</div>
		<div class="dlg-b">
			{#if formError}<p class="form-error" role="alert">{formError}</p>{/if}
			<fieldset aria-describedby={errors.characterId ? `${uid}-char-err` : undefined}>
				<legend class="lbl">Para qual personagem?</legend>
				{#each options as o (o.character.id)}
					<label class="pick" class:on={chosen === o.character.id} class:off={!o.ok}>
						<input
							type="radio"
							name="characterId"
							value={o.character.id}
							bind:group={chosen}
							disabled={!o.ok}
						/>
						<img src="/portraits/{o.character.portrait}.svg" alt="" width="36" height="36" />
						<span class="who">
							<span class="nm">{o.character.nick}</span>
							<span class="sub"
								>{classNames[o.character.classId] ?? o.character.classId} · Nv {o.character.level} ·
								{ROLE_LABELS[o.character.role]}</span
							>
						</span>
						<span class="why">{o.why}</span>
					</label>
				{:else}
					<span class="hint">Você não tem outro personagem para trocar.</span>
				{/each}
				{#if errors.characterId}<span class="hint-err" id="{uid}-char-err"
						><Icon name="circle-x" size={14} />{errors.characterId}</span
					>{/if}
			</fieldset>
			{#if mode === 'request'}
				<label class="lbl" for="{uid}-reason">Motivo</label>
				<textarea
					id="{uid}-reason"
					class="input"
					class:err={errors.reason}
					name="reason"
					rows="3"
					maxlength="250"
					value={reason}
					aria-invalid={errors.reason ? 'true' : undefined}
					aria-describedby={errors.reason ? `${uid}-reason-err` : `${uid}-hint`}></textarea>
				{#if errors.reason}<span class="hint-err" id="{uid}-reason-err"
						><Icon name="circle-x" size={14} />{errors.reason}</span
					>{/if}
				<span class="hint" id="{uid}-hint"
					>De 10 a 250 caracteres. O pedido pode esperar a vaga abrir.</span
				>
			{:else}
				<span class="hint"
					>A vaga que você deixa conta como livre. O personagem não pode estar em outro grupo a
					menos de 2 h.</span
				>
			{/if}
		</div>
		<footer class="dlg-f">
			<Button variant="ghost" onclick={() => dialog?.close()}>Voltar</Button>
			<Button type="submit" disabled={submitting || !anyOk || !chosen}
				>{mode === 'request' ? 'Enviar pedido' : 'Trocar'}</Button
			>
		</footer>
	</form>
</dialog>

<style>
	.dlg {
		width: min(520px, calc(100vw - 32px));
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
	.dlg-h p {
		margin: 0;
		color: var(--fg-2);
	}
	.dlg-b {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 18px 22px;
	}
	fieldset {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0 0 8px;
		padding: 0;
		border: 0;
	}
	legend {
		margin-bottom: 8px;
		padding: 0;
	}
	.lbl {
		font: 600 12px/1.2 var(--font-ui);
		color: var(--fg-2);
	}
	.pick {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 12px 8px 8px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-default);
		background: var(--ink-2);
		cursor: pointer;
	}
	.pick:focus-within {
		outline: 2px solid var(--gold-400);
		outline-offset: 2px;
	}
	.pick.on {
		border-color: var(--gold-line);
		background: var(--gold-soft);
	}
	.pick.off {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.pick input {
		position: absolute;
		opacity: 0;
		width: 1px;
		height: 1px;
	}
	.pick img {
		flex: none;
		border-radius: var(--radius-sm);
		border: 1px solid var(--line-2);
		image-rendering: pixelated;
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 3px;
		min-width: 0;
	}
	.nm {
		font: 700 14px/1.15 var(--font-display);
	}
	.sub,
	.hint {
		font: 500 12px/1.3 var(--font-ui);
		color: var(--fg-3);
	}
	.why {
		margin-left: auto;
		text-align: right;
		font: 500 11px/1.2 var(--font-ui);
		color: var(--fg-3);
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
</style>
