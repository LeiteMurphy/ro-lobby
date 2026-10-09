<script lang="ts">
	import { portraitInfo } from '$lib/characters/portraits';
	import { ROLE_ICONS } from '$lib/home/catalog';
	import { ROLE_LABELS } from '$lib/home/types';
	import Icon from '$lib/ui/Icon.svelte';
	import type { Talent } from '../api';
	import { daysLabel, instancesLabel, rangeLabel } from '../format';

	interface Props {
		talent: Talent;
		/** Nome da classe para exibir; a API manda só o id. */
		className: string;
		/** Link de login para quem não está logado ver o Discord (RN-12). */
		loginHref?: string;
	}

	let { talent, className, loginHref }: Props = $props();
	const portrait = $derived(portraitInfo(talent.portrait));
	const headingId = $props.id();
</script>

<!-- RN-12: retrato, nick, classe, nível, função, dias e faixa, instâncias, link e Discord. -->
<article class="talent" aria-labelledby={headingId} data-testid="talent-card">
	<img class="portrait" src={portrait.src} alt="" width="52" height="52" />
	<div class="who">
		<div class="head">
			<h3 class="nick" id={headingId}>{talent.nick}</h3>
			<span class="role role--{talent.role}"
				><Icon name={ROLE_ICONS[talent.role]} size={14} />{ROLE_LABELS[talent.role]}</span
			>
		</div>
		<span class="sub">{className} · Nv {talent.level}</span>
		<span class="sub"
			>{daysLabel(talent.days)} · {rangeLabel(talent.start, talent.end)} (Brasília)</span
		>
		<span class="sub"
			>{instancesLabel(
				talent.anyInstance,
				talent.instances.map((i) => i.name)
			)}</span
		>
		<div class="contact">
			{#if talent.discordUsername}
				<span class="discord" data-testid="talent-discord">@{talent.discordUsername}</span>
			{:else if loginHref}
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- login com volta -->
				<a class="login" href={loginHref}>Entre para ver o Discord</a>
			{:else}
				<span class="muted">Entre para ver o Discord</span>
			{/if}
			{#if talent.link}
				<!-- eslint-disable svelte/no-navigation-without-resolve -- link externo informado pelo jogador -->
				<a
					class="ext"
					href={talent.link}
					target="_blank"
					rel="noopener noreferrer"
					aria-label="Perfil externo de {talent.nick} (abre em outra aba)"
					><Icon name="external-link" size={14} /></a
				>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
			{/if}
		</div>
	</div>
</article>

<style>
	.talent {
		display: flex;
		gap: 12px;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-default);
		background: var(--surface-raised);
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
		flex: 1;
	}
	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}
	.nick {
		margin: 0;
		font: 700 15px/1.15 var(--font-display);
		color: var(--fg-1);
		overflow-wrap: anywhere;
	}
	.role {
		flex: none;
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font: 600 12px/1 var(--font-ui);
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
	.sub {
		font: 500 12px/1.35 var(--font-ui);
		color: var(--fg-3);
		overflow-wrap: anywhere;
	}
	.contact {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		margin-top: 2px;
	}
	.discord {
		font: 600 13px/1.2 var(--font-ui);
		color: var(--fg-1);
		overflow-wrap: anywhere;
	}
	.login {
		font: 600 12px/1.2 var(--font-ui);
		color: var(--gold-300);
	}
	.muted {
		font: 500 12px/1.2 var(--font-ui);
		color: var(--fg-3);
	}
	.ext {
		display: inline-flex;
		color: var(--fg-2);
	}
	.ext:hover {
		color: var(--fg-1);
	}
</style>
