<script lang="ts" module>
	import type { FieldError } from '../api';

	/** O que a action devolve para o diálogo reabrir com os valores e os erros (RN-19). */
	export interface DialogForm {
		values?: Partial<Record<'nick' | 'classId' | 'level' | 'role' | 'portrait' | 'link', string>>;
		errors?: Partial<Record<FieldError['field'], string>>;
		message?: string | null;
	}
</script>

<script lang="ts">
	import { onMount } from 'svelte';
	import { enhance } from '$app/forms';
	import { ROLE_ICONS } from '$lib/home/catalog';
	import { ROLE_LABELS, ROLES } from '$lib/home/types';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { Character, RoClassEntry } from '../api';
	import { DEFAULT_PORTRAIT, PORTRAITS } from '../portraits';

	interface Props {
		mode: 'create' | 'update';
		character?: Character;
		classes: readonly RoClassEntry[];
		/** Resultado da última action deste diálogo, quando houve erro. */
		form?: DialogForm | null;
		onclose: () => void;
	}

	let { mode, character, classes, form, onclose }: Props = $props();

	const uid = $props.id();
	let dialog: HTMLDialogElement | undefined = $state();
	let submitting = $state(false);

	// Valores de partida: o que voltou de um erro, ou o personagem em edição, ou o padrão.
	// Depois disso, os campos guardam o que o Usuário digita.
	const start = (() => {
		const v = form?.values;
		return {
			nick: v?.nick ?? character?.nick ?? '',
			classId: v?.classId ?? character?.classId ?? '',
			level: v?.level ?? (character ? String(character.level) : ''),
			role: v?.role ?? character?.role ?? '',
			portrait: v?.portrait || character?.portrait || DEFAULT_PORTRAIT,
			link: v?.link ?? character?.link ?? ''
		};
	})();
	let portrait = $state(start.portrait);

	const errors = $derived(form?.errors ?? {});
	const title = $derived(
		mode === 'create' ? 'Adicionar personagem' : `Editar ${character?.nick ?? ''}`
	);

	// As classes em grupos pela linha, na ordem do bROWiki (RN-06). Uma linha que aparece
	// duas vezes na página (ex.: Aprendiz e as expandidas do Aprendiz) fica num grupo só.
	const groups = $derived.by(() => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- Map local, refeito a cada cálculo
		const byFamily = new Map<string, RoClassEntry[]>();
		for (const c of classes) {
			const items = byFamily.get(c.family);
			if (items) items.push(c);
			else byFamily.set(c.family, [c]);
		}
		return [...byFamily].map(([family, items]) => ({ family, items }));
	});

	onMount(() => {
		dialog?.showModal();
	});

	const errId = (field: string) => `${uid}-${field}-err`;
</script>

<dialog
	bind:this={dialog}
	class="dlg"
	aria-labelledby="{uid}-title"
	{onclose}
	data-testid="character-dialog"
>
	<form
		method="POST"
		action={mode === 'create' ? '?/create' : '?/update'}
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
		<header class="dlg-h">
			<div class="titles">
				<h2 id="{uid}-title">{title}</h2>
				<p>A classe vem da lista do bROWiki. O nick é único no RO Lobby.</p>
			</div>
			<button type="button" class="x" aria-label="Fechar" onclick={() => dialog?.close()}>
				<Icon name="x" size={18} />
			</button>
		</header>

		<div class="dlg-b">
			{#if form?.message}
				<p class="form-error" role="alert">{form.message}</p>
			{/if}
			{#if character}<input type="hidden" name="id" value={character.id} />{/if}

			<fieldset class="field">
				<legend class="lbl">Retrato</legend>
				<div class="picks">
					{#each PORTRAITS as p (p.id)}
						<label class="pick" class:on={portrait === p.id}>
							<input type="radio" name="portrait" value={p.id} bind:group={portrait} />
							<img src={p.src} alt={p.label} width="96" height="96" />
						</label>
					{/each}
				</div>
			</fieldset>

			<div class="field">
				<label class="lbl" for="{uid}-nick">Nick</label>
				<input
					id="{uid}-nick"
					class="input"
					class:err={errors.nick}
					name="nick"
					value={start.nick}
					maxlength="24"
					autocomplete="off"
					required
					aria-invalid={errors.nick ? 'true' : undefined}
					aria-describedby={errors.nick ? errId('nick') : undefined}
				/>
				{#if errors.nick}<span class="hint-err" id={errId('nick')}
						><Icon name="circle-x" size={14} />{errors.nick}</span
					>{/if}
			</div>

			<div class="row">
				<div class="field">
					<label class="lbl" for="{uid}-class">Classe</label>
					<select
						id="{uid}-class"
						class="input"
						class:err={errors.classId}
						name="classId"
						required
						aria-invalid={errors.classId ? 'true' : undefined}
						aria-describedby={errors.classId ? errId('classId') : undefined}
					>
						<option value="" selected={start.classId === ''}>Escolha a classe</option>
						{#each groups as g (g.family)}
							<optgroup label={g.family}>
								{#each g.items as c (c.id)}
									<option value={c.id} selected={c.id === start.classId}>{c.name}</option>
								{/each}
							</optgroup>
						{/each}
					</select>
					{#if errors.classId}<span class="hint-err" id={errId('classId')}
							><Icon name="circle-x" size={14} />{errors.classId}</span
						>{/if}
				</div>
				<div class="field">
					<label class="lbl" for="{uid}-level">Nível</label>
					<input
						id="{uid}-level"
						class="input"
						class:err={errors.level}
						name="level"
						type="number"
						inputmode="numeric"
						min="1"
						max="275"
						value={start.level}
						required
						aria-invalid={errors.level ? 'true' : undefined}
						aria-describedby={errors.level ? errId('level') : undefined}
					/>
					{#if errors.level}<span class="hint-err" id={errId('level')}
							><Icon name="circle-x" size={14} />{errors.level}</span
						>{/if}
				</div>
			</div>

			<fieldset class="field" aria-describedby={errors.role ? errId('role') : undefined}>
				<legend class="lbl">Função</legend>
				<div class="roles">
					{#each ROLES as role (role)}
						<label class="radio">
							<input type="radio" name="role" value={role} checked={start.role === role} required />
							<Icon name={ROLE_ICONS[role]} size={16} color="var(--{role}-400)" />
							{ROLE_LABELS[role]}
						</label>
					{/each}
				</div>
				{#if errors.role}<span class="hint-err" id={errId('role')}
						><Icon name="circle-x" size={14} />{errors.role}</span
					>{/if}
			</fieldset>

			<div class="field">
				<label class="lbl" for="{uid}-link">Link externo (opcional)</label>
				<input
					id="{uid}-link"
					class="input"
					class:err={errors.link}
					name="link"
					type="url"
					placeholder="https://"
					value={start.link}
					autocomplete="off"
					aria-invalid={errors.link ? 'true' : undefined}
					aria-describedby={errors.link ? errId('link') : undefined}
				/>
				{#if errors.link}<span class="hint-err" id={errId('link')}
						><Icon name="circle-x" size={14} />{errors.link}</span
					>{/if}
			</div>
		</div>

		<footer class="dlg-f">
			<Button variant="ghost" onclick={() => dialog?.close()}>Cancelar</Button>
			<Button type="submit" disabled={submitting}>Salvar personagem</Button>
		</footer>
	</form>
</dialog>

<style>
	.dlg {
		width: min(560px, calc(100vw - 32px));
		max-height: calc(100dvh - 32px);
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
		align-items: flex-start;
		gap: 12px;
		padding: 18px 18px 0 22px;
	}
	.titles {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding-top: 4px;
	}
	h2 {
		margin: 0;
		font: var(--type-h3);
	}
	.titles p {
		margin: 0;
		color: var(--fg-2);
	}
	.x {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border: 0;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--fg-2);
		cursor: pointer;
	}
	.x:hover {
		background: var(--ink-3);
	}
	.dlg-b {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 18px 22px;
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
	.form-error {
		margin: 0;
		padding: 10px 12px;
		border-radius: var(--radius-sm);
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
		font: 500 13px/1.3 var(--font-ui);
	}
	fieldset {
		margin: 0;
		padding: 0;
		border: 0;
		min-width: 0;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}
	.lbl {
		padding: 0;
		margin-bottom: 6px;
		font: 600 12px/1.2 var(--font-ui);
		color: var(--fg-2);
	}
	label.lbl {
		margin-bottom: 0;
	}
	.row {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 12px;
	}
	.input {
		height: 36px;
		padding: 0 12px;
		background: var(--surface-input);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-sm);
		color: var(--fg-1);
		font: 400 14px/1 var(--font-ui);
		color-scheme: dark;
	}
	.input.err {
		border-color: var(--status-error);
	}
	.hint-err {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-top: 2px;
		font: var(--type-caption);
		color: var(--status-error);
	}
	.picks {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 10px;
	}
	.pick {
		position: relative;
		display: flex;
		padding: 6px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-default);
		background: var(--ink-2);
		cursor: pointer;
	}
	.pick.on {
		border-color: var(--gold-line);
		background: var(--gold-soft);
		box-shadow: var(--shadow-glow-gold);
	}
	.pick:has(input:focus-visible) {
		box-shadow: var(--focus-ring);
	}
	.pick input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}
	.pick img {
		width: 100%;
		height: auto;
		aspect-ratio: 1;
		border-radius: var(--radius-sm);
		image-rendering: pixelated;
	}
	.roles {
		display: flex;
		flex-wrap: wrap;
		gap: 18px;
	}
	.radio {
		display: flex;
		align-items: center;
		gap: 8px;
		font: 500 14px/1 var(--font-ui);
		color: var(--fg-1);
		cursor: pointer;
	}
	.radio input {
		accent-color: var(--accent);
		width: 16px;
		height: 16px;
		margin: 0;
	}
	@media (max-width: 520px) {
		.row {
			grid-template-columns: 1fr;
		}
		.picks {
			grid-template-columns: repeat(2, 1fr);
		}
	}
</style>
