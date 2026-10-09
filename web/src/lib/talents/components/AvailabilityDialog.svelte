<script lang="ts" module>
	/** Valores do formulário, como o Usuário marcou, para voltar ao diálogo num erro. */
	export interface AvailabilityValues {
		enabled: boolean;
		days: number[];
		start: string;
		end: string;
		anyInstance: boolean;
		instanceIds: string[];
	}

	/** O que a action devolve para o diálogo reabrir com os valores e os erros. */
	export interface AvailabilityForm {
		values?: AvailabilityValues;
		errors?: Partial<Record<string, string>>;
		message?: string | null;
	}
</script>

<script lang="ts">
	import { onMount } from 'svelte';
	import { enhance } from '$app/forms';
	import type { Character } from '$lib/characters/api';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import { CLOCK_OPTIONS, WEEK_ORDER, WEEKDAYS, WEEKDAY_NAMES } from '../format';

	interface Props {
		character: Character;
		instances: readonly { id: string; name: string; level: number }[];
		/** Resultado da última action deste diálogo, quando houve erro. */
		form?: AvailabilityForm | null;
		onclose: () => void;
	}

	let { character, instances, form, onclose }: Props = $props();

	const uid = $props.id();
	const HIGH_LEVEL = 130;
	let dialog: HTMLDialogElement | undefined = $state();
	let submitting = $state(false);

	// CA-01.4: o que voltou de um erro, ou o que estava guardado, ou o padrão.
	const start = (() => {
		const saved = character.availability;
		return (
			form?.values ?? {
				enabled: saved?.enabled ?? true,
				days: saved?.days ?? [],
				start: saved?.start ?? '19:00',
				end: saved?.end ?? '23:00',
				anyInstance: saved?.anyInstance ?? true,
				instanceIds: saved?.instanceIds ?? []
			}
		);
	})();
	let enabled = $state(start.enabled);
	let anyInstance = $state(start.anyInstance);

	const errors = $derived(form?.errors ?? {});
	const groups = $derived([
		{ label: 'Nível 130 ou mais', items: instances.filter((i) => i.level >= HIGH_LEVEL) },
		{ label: 'Nível menor', items: instances.filter((i) => i.level < HIGH_LEVEL) }
	]);

	onMount(() => dialog?.showModal());

	const errId = (field: string) => `${uid}-${field}-err`;
	const describe = (field: string) => (errors[field] ? errId(field) : undefined);
</script>

<dialog
	bind:this={dialog}
	class="dlg"
	aria-labelledby="{uid}-title"
	{onclose}
	data-testid="availability-dialog"
>
	<form
		method="POST"
		action="?/availability"
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
				<h2 id="{uid}-title">Banco de talentos · {character.nick}</h2>
				<p>Quem monta grupos vê seu personagem quando o horário e a instância combinam.</p>
			</div>
			<button type="button" class="x" aria-label="Fechar" onclick={() => dialog?.close()}>
				<Icon name="x" size={18} />
			</button>
		</header>

		<div class="dlg-b">
			{#if form?.message}<p class="form-error" role="alert">{form.message}</p>{/if}
			<input type="hidden" name="id" value={character.id} />

			<!-- RN-01: desligado, sai do banco e os dados ficam guardados. -->
			<label class="switch">
				<input type="checkbox" name="enabled" bind:checked={enabled} />
				<span>Disponível no banco de talentos</span>
			</label>
			{#if !enabled}
				<p class="hint">Fora do banco. Os dias, a faixa e as instâncias ficam guardados.</p>
			{/if}

			<fieldset class="field" aria-describedby={describe('days')}>
				<legend class="lbl">Dias</legend>
				<div class="chips">
					{#each WEEK_ORDER as d (d)}
						<label class="chip">
							<input
								type="checkbox"
								name="days"
								value={d}
								checked={start.days.includes(d)}
								aria-label={WEEKDAY_NAMES[d]}
							/>
							<span aria-hidden="true">{WEEKDAYS[d]}</span>
						</label>
					{/each}
				</div>
				{#if errors.days}<span class="hint-err" id={errId('days')}
						><Icon name="circle-x" size={14} />{errors.days}</span
					>{/if}
			</fieldset>

			<!-- RN-03: horário de Brasília; o fim antes do início vira a meia-noite. -->
			<div class="row">
				<div class="field">
					<label class="lbl" for="{uid}-start">Das (Brasília)</label>
					<select
						id="{uid}-start"
						class="input"
						class:err={errors.start}
						name="start"
						aria-invalid={errors.start ? 'true' : undefined}
						aria-describedby={describe('start')}
					>
						{#each CLOCK_OPTIONS as c (c)}<option value={c} selected={c === start.start}>{c}</option
							>{/each}
					</select>
					{#if errors.start}<span class="hint-err" id={errId('start')}
							><Icon name="circle-x" size={14} />{errors.start}</span
						>{/if}
				</div>
				<div class="field">
					<label class="lbl" for="{uid}-end">Até</label>
					<select
						id="{uid}-end"
						class="input"
						class:err={errors.end}
						name="end"
						aria-invalid={errors.end ? 'true' : undefined}
						aria-describedby={describe('end')}
					>
						{#each CLOCK_OPTIONS as c (c)}<option value={c} selected={c === start.end}>{c}</option
							>{/each}
					</select>
					{#if errors.end}<span class="hint-err" id={errId('end')}
							><Icon name="circle-x" size={14} />{errors.end}</span
						>{/if}
				</div>
			</div>
			<p class="hint">Se o fim for antes do início, a faixa vai até o dia seguinte.</p>

			<!-- RN-04: "Qualquer instância" ou uma ou mais do catálogo. -->
			<fieldset class="field" aria-describedby={describe('instanceIds')}>
				<legend class="lbl">Instâncias de interesse</legend>
				<label class="switch">
					<input type="checkbox" name="anyInstance" bind:checked={anyInstance} />
					<span>Qualquer instância</span>
				</label>
				{#if !anyInstance}
					<div class="instances">
						{#each groups as g (g.label)}
							{#if g.items.length}
								<span class="group">{g.label}</span>
								{#each g.items as i (i.id)}
									<label class="check">
										<input
											type="checkbox"
											name="instanceIds"
											value={i.id}
											checked={start.instanceIds.includes(i.id)}
										/>
										<span>{i.name} <span class="muted">Nv {i.level}</span></span>
									</label>
								{/each}
							{/if}
						{/each}
					</div>
				{/if}
				{#if errors.instanceIds}<span class="hint-err" id={errId('instanceIds')}
						><Icon name="circle-x" size={14} />{errors.instanceIds}</span
					>{/if}
			</fieldset>
		</div>

		<footer class="dlg-f">
			<Button variant="ghost" onclick={() => dialog?.close()}>Cancelar</Button>
			<Button type="submit" disabled={submitting}>Salvar</Button>
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
		overflow-wrap: anywhere;
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
		grid-template-columns: 1fr 1fr;
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
	.hint {
		margin: -8px 0 0;
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.hint-err {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-top: 2px;
		font: var(--type-caption);
		color: var(--status-error);
	}
	.switch,
	.check {
		display: flex;
		align-items: center;
		gap: 8px;
		font: 500 14px/1.3 var(--font-ui);
		color: var(--fg-1);
		cursor: pointer;
	}
	.switch input,
	.check input {
		accent-color: var(--accent);
		width: 16px;
		height: 16px;
		margin: 0;
		flex: none;
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.chip {
		position: relative;
		display: inline-flex;
	}
	.chip input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}
	.chip span {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 46px;
		height: 32px;
		padding: 0 10px;
		border-radius: 999px;
		border: 1px solid var(--border-default);
		background: var(--ink-2);
		color: var(--fg-2);
		font: 600 13px/1 var(--font-ui);
		cursor: pointer;
	}
	.chip input:checked + span {
		border-color: var(--gold-line);
		background: var(--gold-soft);
		color: var(--gold-300);
	}
	.chip input:focus-visible + span {
		box-shadow: var(--focus-ring);
	}
	.instances {
		display: flex;
		flex-direction: column;
		gap: 8px;
		max-height: 220px;
		overflow-y: auto;
		padding: 10px 12px;
		border: 1px solid var(--border-default);
		border-radius: var(--radius-sm);
		background: var(--ink-2);
	}
	.group {
		font: 600 11px/1.2 var(--font-ui);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--fg-3);
	}
	.muted {
		color: var(--fg-3);
	}
	@media (max-width: 520px) {
		.row {
			grid-template-columns: 1fr;
		}
	}
</style>
