<script lang="ts">
	import { ROLE_ICONS } from '$lib/home/catalog';
	import { ROLE_LABELS } from '$lib/home/types';
	import Icon from '$lib/ui/Icon.svelte';
	import type { Character } from '../api';
	import { portraitInfo } from '../portraits';
	import CardMenu from './CardMenu.svelte';

	interface Props {
		character: Character;
		/** Nome da classe para exibir; o personagem guarda só o id (D-03). */
		className: string;
		onedit: () => void;
		ondelete: () => void;
	}

	let { character, className, onedit, ondelete }: Props = $props();
	const portrait = $derived(portraitInfo(character.portrait));
	const headingId = $props.id();
</script>

<!-- RN-16: carta vertical (1b) com retrato, nick, selo "Principal", classe, nível e função. -->
<article
	class="card"
	class:main={character.isMain}
	aria-labelledby={headingId}
	data-testid="character-card"
>
	<img class="portrait" src={portrait.src} alt={portrait.label} width="236" height="220" />
	<div class="body">
		<div class="head">
			<h2 class="nick" id={headingId}>{character.nick}</h2>
			{#if character.isMain}<span class="badge">Principal</span>{/if}
		</div>
		<span class="cls">{className}</span>
		<dl class="stats">
			<div>
				<dt>Nível</dt>
				<dd class="lvl">{character.level}</dd>
			</div>
			<div>
				<dt>Função</dt>
				<dd class="role role--{character.role}">
					<Icon name={ROLE_ICONS[character.role]} size={15} />{ROLE_LABELS[character.role]}
				</dd>
			</div>
		</dl>
		<div class="foot">
			{#if character.link}
				<!-- RN-09 / CA-01.4: abre em outra aba sem passar a página de origem. -->
				<!-- eslint-disable svelte/no-navigation-without-resolve -- link externo informado pelo Usuário -->
				<a
					class="link"
					href={character.link}
					target="_blank"
					rel="noopener noreferrer"
					aria-label="Perfil externo de {character.nick} (abre em outra aba)"
					title="Perfil externo"
				>
					<Icon name="external-link" size={15} />
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
			{:else}
				<span></span>
			{/if}
			<CardMenu
				id={character.id}
				nick={character.nick}
				isMain={character.isMain}
				{onedit}
				{ondelete}
			/>
		</div>
	</div>
</article>

<style>
	.card {
		position: relative;
		display: flex;
		flex-direction: column;
		background: var(--surface-card);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-lg);
	}
	.card.main {
		border-color: var(--gold-line);
		box-shadow: var(--shadow-glow-gold);
	}
	.portrait {
		display: block;
		width: 100%;
		height: 220px;
		object-fit: cover;
		image-rendering: pixelated;
		border-bottom: 1px solid var(--line-2);
		border-radius: var(--radius-lg) var(--radius-lg) 0 0;
	}
	.body {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px;
	}
	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		min-width: 0;
	}
	.nick {
		margin: 0;
		font: 700 20px/1.1 var(--font-display);
		letter-spacing: 0.01em;
		color: var(--fg-1);
		overflow-wrap: anywhere;
	}
	.badge {
		flex: none;
		display: inline-flex;
		align-items: center;
		height: 20px;
		padding: 0 7px;
		border-radius: var(--radius-xs);
		background: var(--gold-soft);
		color: var(--gold-300);
		font: 600 11px/1 var(--font-ui);
	}
	.cls {
		font: 500 13px/1.3 var(--font-ui);
		color: var(--fg-2);
	}
	.stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
		margin: 0;
		padding: 10px 0 0;
		border-top: 1px solid var(--border-subtle);
	}
	dt {
		font: 600 10px/1.2 var(--font-ui);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--fg-3);
	}
	dd {
		margin: 0;
	}
	.lvl {
		font: 700 22px/1.1 var(--font-display);
		font-variant-numeric: tabular-nums;
		color: var(--fg-1);
	}
	.role {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-top: 4px;
		font: 600 14px/1.2 var(--font-ui);
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
	.foot {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding-top: 6px;
	}
	.link {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: var(--radius-sm);
		border: 1px solid var(--line-2);
		background: var(--surface-raised);
		color: var(--fg-2);
	}
	.link:hover {
		color: var(--fg-1);
	}
</style>
