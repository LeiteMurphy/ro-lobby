<script lang="ts">
	import { portraitAlt, portraitSrc } from '$lib/characters/portraits';
	import { onMount, untrack } from 'svelte';
	import { enhance } from '$app/forms';
	import { ROLE_LABELS } from '$lib/home/types';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { Eligibility } from '../detail';

	interface Props {
		/** Instância, dia, hora e nível mínimo, para o cabeçalho. */
		title: string;
		options: Eligibility[];
		classNames: Record<string, string>;
		/** Valores e erros de uma tentativa anterior desta abertura. */
		characterId?: string;
		message?: string;
		errors?: Record<string, string>;
		formError?: string | null;
		onclose: () => void;
	}

	let {
		title,
		options,
		classNames,
		characterId = '',
		message = '',
		errors = {},
		formError = null,
		onclose
	}: Props = $props();
	const uid = $props.id();
	let dialog: HTMLDialogElement | undefined = $state();
	let submitting = $state(false);
	// Começa no personagem já escolhido ou no primeiro que pode (Candidatura 1c).
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
	data-testid="apply-dialog"
>
	<form
		method="POST"
		action="?/apply"
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
		<div class="dlg-h">
			<h2 id="{uid}-title">Candidatar-se</h2>
			<p id="{uid}-desc">{title}</p>
		</div>
		<div class="dlg-b">
			{#if formError}<p class="form-error" role="alert">{formError}</p>{/if}
			<fieldset aria-describedby={errors.characterId ? `${uid}-char-err` : undefined}>
				<legend class="lbl">Com qual personagem?</legend>
				{#each options as o (o.character.id)}
					<label class="pick" class:on={chosen === o.character.id} class:off={!o.ok}>
						<input
							type="radio"
							name="characterId"
							value={o.character.id}
							bind:group={chosen}
							disabled={!o.ok}
						/>
						<img
							src={portraitSrc(o.character.portrait)}
							alt={portraitAlt(o.character.portrait)}
							width="36"
							height="36"
						/>
						<span class="who">
							<span class="nm">{o.character.nick}</span>
							<span class="sub"
								>{classNames[o.character.classId] ?? o.character.classId} · Nv {o.character.level} ·
								{ROLE_LABELS[o.character.role]}</span
							>
						</span>
						<span class="why">{o.why}</span>
					</label>
				{/each}
				{#if errors.characterId}<span class="hint-err" id="{uid}-char-err"
						><Icon name="circle-x" size={14} />{errors.characterId}</span
					>{/if}
			</fieldset>
			<label class="lbl" for="{uid}-message">Mensagem para o anfitrião (opcional)</label>
			<textarea
				id="{uid}-message"
				class="input"
				class:err={errors.message}
				name="message"
				rows="3"
				maxlength="250"
				value={message}
				aria-invalid={errors.message ? 'true' : undefined}
				aria-describedby={errors.message ? `${uid}-message-err` : `${uid}-hint`}></textarea>
			{#if errors.message}<span class="hint-err" id="{uid}-message-err"
					><Icon name="circle-x" size={14} />{errors.message}</span
				>{/if}
			<span class="hint" id="{uid}-hint"
				>Enquanto a candidatura estiver pendente ou aceita, o personagem não muda nível nem função.</span
			>
		</div>
		<footer class="dlg-f">
			<Button variant="ghost" onclick={() => dialog?.close()}>Voltar</Button>
			<Button type="submit" disabled={submitting || !anyOk || !chosen}>Enviar candidatura</Button>
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
