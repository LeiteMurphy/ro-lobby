<script lang="ts">
	import { enhance } from '$app/forms';
	import { ROLE_LABELS } from '$lib/home/types';
	import { fromUtcIso } from '$lib/lobbies/time';
	import Button from '$lib/ui/Button.svelte';
	import Icon from '$lib/ui/Icon.svelte';
	import type { Person } from '../detail';

	interface Props {
		person: Person;
		/** Nome de cada classe pelo id do catálogo. */
		classNames: Record<string, string>;
		/**
		 * As ações do dono, com o lobby aberto: Aceitar e Recusar o candidato (RN-08), Remover o
		 * membro (RN-15) e Trocar o próprio personagem (RN-19).
		 */
		canDecide: boolean;
		/** Grupo livre: a troca do dono não depende de vaga na função (spec grupo-livre, RN-05). */
		free?: boolean;
		onreject: (person: Person) => void;
		onremove?: (person: Person) => void;
		onswap?: () => void;
	}

	let { person, classNames, canDecide, free = false, onreject, onremove, onswap }: Props = $props();

	const OVER = { host: 'Anfitrião', member: 'Membro', candidate: 'Candidato' } as const;
	const className = $derived(
		person.classId ? (classNames[person.classId] ?? person.classId) : null
	);
	const classLevel = $derived(
		[className, person.level ? `Nv ${person.level}` : null].filter(Boolean).join(' · ')
	);
	const since = $derived(person.createdAt ? fromUtcIso(person.createdAt).time : null);
	let accepting = $state(false);
</script>

<!-- RN-31: detalhes da pessoa escolhida na composição ou nos pendentes. -->
<section class="panel" aria-labelledby="player-title" aria-live="polite" data-testid="player-panel">
	<span class="over"
		>{OVER[person.kind]}{#if canDecide && person.kind === 'host'}&nbsp;(você){/if}</span
	>
	<div class="head">
		{#if person.portrait}<img
				class="portrait"
				src="/portraits/{person.portrait}.svg"
				alt=""
				width="72"
				height="72"
			/>{/if}
		<span class="who">
			<span class="nm" id="player-title">{person.nick}</span>
			<span class="sub"
				>{ROLE_LABELS[person.role]}{person.kind === 'candidate' && since
					? ` · pendente desde ${since}`
					: ''}</span
			>
		</span>
	</div>
	{#if classLevel}<div class="kv"><span>Classe</span><b>{classLevel}</b></div>{/if}
	{#if person.discordName}
		<div class="kv"><span>Discord</span><b>{person.discordName}</b></div>
	{/if}
	{#if person.link}
		<div class="kv">
			<span>Perfil externo</span>
			<a href={person.link} target="_blank" rel="noopener noreferrer external"
				>abrir <Icon name="external-link" size={12} /></a
			>
		</div>
	{/if}
	{#if !person.discordName}
		<!-- RN-32: o Discord dos membros fica com o grupo. -->
		<p class="locked"><Icon name="lock" size={13} />O Discord aparece para quem está no grupo.</p>
	{/if}
	{#if person.message}
		<span class="over">Mensagem</span>
		<p class="msg">{person.message}</p>
	{/if}
	{#if canDecide && person.kind === 'candidate' && person.applicationId}
		<div class="acts">
			<form
				method="POST"
				action="?/accept"
				use:enhance={() => {
					accepting = true;
					return async ({ update }) => {
						accepting = false;
						await update();
					};
				}}
			>
				<input type="hidden" name="applicationId" value={person.applicationId} />
				<Button type="submit" iconLeft="check" block disabled={accepting}>Aceitar</Button>
			</form>
			<Button variant="outline" class="danger" block onclick={() => onreject(person)}
				>Recusar</Button
			>
		</div>
	{/if}
	{#if canDecide && person.kind === 'member' && onremove}
		<!-- RN-15 / RN-38: o dono remove o membro escolhido (Candidatura 2c). -->
		<Button variant="outline" class="danger" block onclick={() => onremove(person)}
			>Remover do grupo</Button
		>
	{/if}
	{#if canDecide && person.kind === 'host' && onswap}
		<!-- RN-19 / RN-38: o dono troca o próprio personagem, sem aprovação (Candidatura 2e). -->
		<Button variant="secondary" block onclick={onswap}>Trocar personagem</Button>
		<span class="sub"
			>{free
				? 'Troque sem aprovação, para um personagem seu com o nível mínimo; a vaga continua sua.'
				: 'Troque sem aprovação, para um personagem seu com o nível mínimo e vaga na função.'}</span
		>
	{/if}
</section>

<style>
	.panel {
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 20px 22px;
		border-radius: var(--radius-lg);
		background: var(--panel-bg);
		backdrop-filter: blur(var(--panel-blur));
		border: 1px solid var(--panel-border);
	}
	.over {
		font: 600 10px/1.2 var(--font-ui);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--fg-3);
	}
	.head {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.portrait {
		flex: none;
		border-radius: var(--radius-sm);
		border: 1px solid var(--line-2);
		image-rendering: pixelated;
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}
	.nm {
		font: 700 20px/1.15 var(--font-display);
		color: var(--fg-1);
		overflow-wrap: anywhere;
	}
	.sub {
		font: 500 12px/1.3 var(--font-ui);
		color: var(--fg-3);
	}
	.kv {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		font: 500 13px/1.3 var(--font-ui);
		color: var(--fg-3);
	}
	.kv b {
		color: var(--fg-1);
		font-weight: 600;
		text-align: right;
		overflow-wrap: anywhere;
	}
	.kv a {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		color: var(--gold-300);
		font-weight: 600;
	}
	.locked {
		display: flex;
		align-items: center;
		gap: 6px;
		margin: 0;
		padding: 8px 10px;
		border-radius: var(--radius-sm);
		background: var(--ink-2);
		color: var(--fg-3);
		font: 500 12px/1.3 var(--font-ui);
	}
	.msg {
		margin: 0;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		background: var(--ink-2);
		border-left: 3px solid var(--gold-400);
		color: var(--fg-2);
		font: 400 13px/1.45 var(--font-ui);
		overflow-wrap: anywhere;
	}
	.acts {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.acts form {
		margin: 0;
		display: flex;
	}
	.panel :global(.danger) {
		color: var(--status-error);
		border-color: rgba(240, 100, 140, 0.4);
	}
</style>
