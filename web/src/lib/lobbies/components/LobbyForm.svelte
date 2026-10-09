<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Character, RoClassEntry } from '$lib/characters/api';
	import { ROLE_ICONS } from '$lib/home/catalog';
	import LobbyCard from '$lib/home/components/LobbyCard.svelte';
	import { ROLE_LABELS, ROLES, type Lobby as HomeLobby, type Role } from '$lib/home/types';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { FieldError } from '$lib/api/request';
	import type { Instance } from '../api';
	import type { LobbyFormValues } from '../form';

	export interface DayOption {
		date: string;
		label: string;
	}

	interface Props {
		mode: 'create' | 'update';
		/** O catálogo; na edição, também a instância atual se ela saiu dele (RN-17). */
		instances: readonly Pick<Instance, 'id' | 'name' | 'level'>[];
		/** Na criação, os personagens do Usuário; na edição, só o do dono (fixo). */
		characters: readonly Character[];
		classes: readonly RoClassEntry[];
		days: readonly DayOption[];
		values: LobbyFormValues;
		errors?: Partial<Record<FieldError['field'], string>>;
		message?: string | null;
		cancelHref: string;
	}

	let {
		mode,
		instances,
		characters,
		classes,
		days,
		values,
		errors = {},
		message = null,
		cancelHref
	}: Props = $props();

	const uid = $props.id();
	const HIGH_LEVEL = 130;
	const MAX_SLOTS = 12;

	// O formulário guarda o que o Usuário digita; os valores de partida vêm da página.
	// svelte-ignore state_referenced_locally
	const start = values;
	let instanceId = $state(start.instanceId);
	let date = $state(start.date);
	let time = $state(start.time);
	let slots = $state<Record<Role, number>>({
		tank: Number(start.tank),
		support: Number(start.support),
		dps: Number(start.dps)
	});
	let minLevel = $state(start.minLevel);
	let characterId = $state(start.characterId);
	let note = $state(start.note);
	let submitting = $state(false);

	// RN-02: nível 130+ primeiro, depois um separador e as de nível menor (a API já ordena).
	const groups = $derived([
		{ label: 'Nível 130 ou mais', items: instances.filter((i) => i.level >= HIGH_LEVEL) },
		{ label: 'Nível menor', items: instances.filter((i) => i.level < HIGH_LEVEL) }
	]);
	const instance = $derived(instances.find((i) => i.id === instanceId) ?? { name: '', level: 1 });
	const character = $derived(characters.find((c) => c.id === characterId));
	const className = (id: string) => classes.find((c) => c.id === id)?.name ?? id;
	const total = $derived(slots.tank + slots.support + slots.dps);

	// RN-07: o nível mínimo padrão é o da instância; ao trocar de instância, ele volta para o
	// nível de entrada da nova.
	function onInstanceChange() {
		const level = instances.find((i) => i.id === instanceId)?.level;
		if (level) minLevel = String(level);
	}

	function step(role: Role, delta: number) {
		const next = Math.min(MAX_SLOTS, Math.max(0, slots[role] + delta));
		if (delta > 0 && total >= MAX_SLOTS) return;
		slots[role] = next;
	}

	// 1b: prévia do card como vai aparecer na Home.
	const preview: HomeLobby = $derived({
		id: 'previa',
		date,
		time: time || '--:--',
		instance: instance.name || 'Escolha a instância',
		host: character?.nick ?? '—',
		hostClass: character ? className(character.classId) : '',
		minLevel: Number(minLevel) || instance.level,
		composition: {
			tank: { filled: character?.role === 'tank' ? 1 : 0, total: slots.tank },
			support: { filled: character?.role === 'support' ? 1 : 0, total: slots.support },
			dps: { filled: character?.role === 'dps' ? 1 : 0, total: slots.dps }
		}
	});
	const dayLabel = $derived(days.find((d) => d.date === date)?.label ?? '');

	const errId = (field: string) => `${uid}-${field}-err`;
	const describe = (field: FieldError['field']) => (errors[field] ? errId(field) : undefined);
</script>

<div class="cols">
	<form
		class="panel"
		method="POST"
		action={mode === 'create' ? '?/create' : '?/update'}
		novalidate
		use:enhance={() => {
			submitting = true;
			return async ({ update }) => {
				submitting = false;
				await update({ reset: false });
			};
		}}
	>
		{#if message}<p class="form-error" role="alert">{message}</p>{/if}

		<div class="field">
			<label class="lbl" for="{uid}-instance">Instância</label>
			<select
				id="{uid}-instance"
				class="input"
				class:err={errors.instanceId}
				name="instanceId"
				bind:value={instanceId}
				onchange={onInstanceChange}
				aria-invalid={errors.instanceId ? 'true' : undefined}
				aria-describedby={describe('instanceId')}
			>
				<option value="">Escolha a instância</option>
				{#each groups as g (g.label)}
					<optgroup label={g.label}>
						{#each g.items as i (i.id)}
							<option value={i.id}>{i.name} · Nv {i.level}</option>
						{/each}
					</optgroup>
				{/each}
			</select>
			{#if errors.instanceId}<span class="hint-err" id={errId('instanceId')}
					><Icon name="circle-x" size={14} />{errors.instanceId}</span
				>{/if}
		</div>

		<div class="row2">
			<div class="field">
				<label class="lbl" for="{uid}-date">Dia</label>
				<select
					id="{uid}-date"
					class="input"
					class:err={errors.startsAt}
					name="date"
					bind:value={date}
					aria-invalid={errors.startsAt ? 'true' : undefined}
					aria-describedby={describe('startsAt')}
				>
					{#each days as d (d.date)}<option value={d.date}>{d.label}</option>{/each}
				</select>
			</div>
			<div class="field">
				<label class="lbl" for="{uid}-time">Hora (Brasília)</label>
				<input
					id="{uid}-time"
					class="input"
					class:err={errors.startsAt}
					type="time"
					name="time"
					bind:value={time}
					aria-invalid={errors.startsAt ? 'true' : undefined}
					aria-describedby={describe('startsAt')}
				/>
			</div>
		</div>
		{#if errors.startsAt}<span class="hint-err" id={errId('startsAt')}
				><Icon name="circle-x" size={14} />{errors.startsAt}</span
			>{/if}

		<fieldset class="field" aria-describedby={describe('slots')}>
			<legend class="lbl total"
				><span>Vagas por função</span><span class="muted">{total} de {MAX_SLOTS}</span></legend
			>
			<div class="row3">
				{#each ROLES as role (role)}
					<div class="stepper" class:err={errors.slots}>
						<label class="role role--{role}" for="{uid}-{role}"
							><Icon name={ROLE_ICONS[role]} size={15} />{ROLE_LABELS[role]}</label
						>
						<div class="ctl">
							<button
								type="button"
								class="sbtn"
								aria-label="Menos uma vaga de {ROLE_LABELS[role]}"
								onclick={() => step(role, -1)}>−</button
							>
							<input
								id="{uid}-{role}"
								class="n"
								type="number"
								inputmode="numeric"
								min="0"
								max={MAX_SLOTS}
								name={role}
								bind:value={slots[role]}
							/>
							<button
								type="button"
								class="sbtn"
								aria-label="Mais uma vaga de {ROLE_LABELS[role]}"
								onclick={() => step(role, 1)}>+</button
							>
						</div>
					</div>
				{/each}
			</div>
			{#if errors.slots}<span class="hint-err" id={errId('slots')}
					><Icon name="circle-x" size={14} />{errors.slots}</span
				>{/if}
		</fieldset>

		<div class="row2">
			<div class="field">
				<label class="lbl" for="{uid}-min">Nível mínimo</label>
				<input
					id="{uid}-min"
					class="input"
					class:err={errors.minLevel}
					type="number"
					inputmode="numeric"
					min={instance.level}
					max="275"
					name="minLevel"
					bind:value={minLevel}
					aria-invalid={errors.minLevel ? 'true' : undefined}
					aria-describedby={describe('minLevel')}
				/>
				{#if errors.minLevel}<span class="hint-err" id={errId('minLevel')}
						><Icon name="circle-x" size={14} />{errors.minLevel}</span
					>{:else}<span class="hint">Mínimo da instância: {instance.level}</span>{/if}
			</div>
			<div class="field">
				<label class="lbl" for="{uid}-note">Observação (opcional)</label>
				<input
					id="{uid}-note"
					class="input"
					class:err={errors.note}
					name="note"
					maxlength="250"
					placeholder="Ex.: chamar no Discord antes"
					bind:value={note}
					aria-invalid={errors.note ? 'true' : undefined}
					aria-describedby={describe('note')}
				/>
				{#if errors.note}<span class="hint-err" id={errId('note')}
						><Icon name="circle-x" size={14} />{errors.note}</span
					>{/if}
			</div>
		</div>

		<fieldset class="field" aria-describedby={describe('characterId')}>
			<legend class="lbl">Seu personagem</legend>
			<div class="chars">
				{#each characters as c (c.id)}
					<label class="char" class:on={characterId === c.id} class:fixed={mode === 'update'}>
						<input
							type="radio"
							name="characterId"
							value={c.id}
							bind:group={characterId}
							disabled={mode === 'update'}
						/>
						<img src="/portraits/{c.portrait}.svg" alt="" width="34" height="34" />
						<span class="who">
							<span class="nm"
								>{c.nick}{#if c.isMain}<span class="badge">Principal</span>{/if}</span
							>
							<span class="sub">{className(c.classId)} · Nv {c.level} · {ROLE_LABELS[c.role]}</span>
						</span>
					</label>
				{/each}
			</div>
			{#if errors.characterId}<span class="hint-err" id={errId('characterId')}
					><Icon name="circle-x" size={14} />{errors.characterId}</span
				>{/if}
		</fieldset>

		<div class="actions">
			<Button variant="ghost" href={cancelHref}>Cancelar</Button>
			<Button type="submit" disabled={submitting}
				>{mode === 'create' ? 'Criar lobby' : 'Salvar alterações'}</Button
			>
		</div>
	</form>

	<aside class="preview" aria-label="Prévia do card na Home">
		<span class="over">Como vai aparecer na Home{dayLabel ? ` · ${dayLabel}` : ''}</span>
		<LobbyCard lobby={preview} relative="" preview />
	</aside>
</div>

<style>
	.cols {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 22px;
		align-items: start;
	}
	@media (min-width: 960px) {
		.cols {
			grid-template-columns: minmax(0, 1fr) 320px;
		}
		.preview {
			position: sticky;
			top: 80px;
		}
	}
	.panel {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 18px 20px;
		border-radius: var(--radius-lg);
		background: var(--surface-card);
		border: 1px solid var(--border-default);
	}
	.preview {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.over {
		font: 600 10px/1.2 var(--font-ui);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--fg-3);
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
		font: 600 12px/1.2 var(--font-ui);
		color: var(--fg-2);
	}
	legend.lbl {
		margin-bottom: 6px;
	}
	.total {
		display: flex;
		justify-content: space-between;
		width: 100%;
	}
	.muted,
	.hint {
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.row2 {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}
	.row3 {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 10px;
	}
	@media (max-width: 560px) {
		.row2,
		.row3 {
			grid-template-columns: 1fr;
		}
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
		min-width: 0;
	}
	.input.err,
	.stepper.err {
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
		margin: 0;
		padding: 10px 12px;
		border-radius: var(--radius-sm);
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
		font: 500 13px/1.3 var(--font-ui);
	}
	.stepper {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-default);
		background: var(--ink-2);
	}
	.role {
		display: flex;
		align-items: center;
		gap: 6px;
		font: 600 13px/1 var(--font-ui);
	}
	.role--tank {
		color: var(--tank-300);
	}
	.role--support {
		color: var(--support-300);
	}
	.role--dps {
		color: var(--dps-300);
	}
	.ctl {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 6px;
	}
	.n {
		width: 44px;
		text-align: center;
		background: transparent;
		border: 0;
		color: var(--fg-1);
		font: 700 22px/1 var(--font-display);
		font-variant-numeric: tabular-nums;
		-moz-appearance: textfield;
		appearance: textfield;
	}
	.n::-webkit-inner-spin-button,
	.n::-webkit-outer-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}
	.sbtn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: var(--radius-sm);
		border: 1px solid var(--line-2);
		background: var(--surface-raised);
		color: var(--fg-1);
		font: 600 16px/1 var(--font-ui);
		cursor: pointer;
	}
	.chars {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.char {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 6px 12px 6px 6px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-default);
		background: var(--ink-2);
		cursor: pointer;
	}
	.char.fixed {
		cursor: default;
	}
	.char.on {
		border-color: var(--gold-line);
		background: var(--gold-soft);
	}
	.char:has(input:focus-visible) {
		box-shadow: var(--focus-ring);
	}
	.char input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}
	.char img {
		border-radius: var(--radius-sm);
		image-rendering: pixelated;
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}
	.nm {
		display: flex;
		align-items: center;
		gap: 6px;
		font: 700 14px/1.1 var(--font-display);
	}
	.sub {
		font: 500 11px/1.2 var(--font-ui);
		color: var(--fg-3);
	}
	.badge {
		display: inline-flex;
		align-items: center;
		height: 18px;
		padding: 0 6px;
		border-radius: var(--radius-xs);
		background: var(--gold-soft);
		color: var(--gold-300);
		font: 600 10px/1 var(--font-ui);
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}
</style>
