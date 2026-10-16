<script lang="ts">
	import CharacterCard from '$lib/characters/components/CharacterCard.svelte';
	import CharacterDialog, {
		type DialogForm
	} from '$lib/characters/components/CharacterDialog.svelte';
	import ConfirmDialog from '$lib/characters/components/ConfirmDialog.svelte';
	import type { Character } from '$lib/characters/api';
	import { LIMIT_MESSAGE } from '$lib/characters/messages';
	import { UNDER_CONSTRUCTION } from '$lib/discord/messages';
	import { serverName } from '$lib/catalog/servers';
	import { enhance } from '$app/forms';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import AvailabilityDialog, {
		type AvailabilityForm
	} from '$lib/talents/components/AvailabilityDialog.svelte';
	import Button from '$lib/ui/Button.svelte';
	import { resolve } from '$app/paths';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	const MAX = 10;
	const full = $derived(data.characters.length >= MAX);
	const limitHintId = $props.id();

	const classNames = $derived(new Map(data.classes.map((c) => [c.id, c.name])));
	// Classe fora do catálogo (borda de RN-06): mostra o id até o Usuário editar.
	const classNameOf = (id: string) => classNames.get(id) ?? id;

	type OpenInput =
		| { kind: 'create' }
		| { kind: 'update'; character: Character }
		| { kind: 'delete'; character: Character }
		| { kind: 'availability'; character: Character };
	// `staleForm` guarda o resultado de action que já existia quando o diálogo abriu: ele é
	// de uma tentativa anterior e não volta para o diálogo novo (AJ-04, RN-19).
	type Open = OpenInput & { key: number; staleForm: typeof form };
	// $state.raw: sem proxy, para a comparação com `form` ser por identidade.
	let open = $state.raw<Open | null>(null);
	let nextKey = 0;
	const show = (o: OpenInput) => (open = { ...o, key: ++nextKey, staleForm: form });

	// Erros da action só valem para o diálogo que os pediu, na mesma abertura.
	const dialogForm = $derived(
		form &&
			open &&
			form !== open.staleForm &&
			form.mode === open.kind &&
			(open.kind === 'create' || form.id === open.character.id)
			? form
			: null
	);
	// spec discord-bot, RN-02: a data da conexão, em Brasília.
	const connectedOn = $derived(
		data.discord.connectedAt
			? new Intl.DateTimeFormat('pt-BR', { timeZone: 'America/Sao_Paulo' }).format(
					new Date(data.discord.connectedAt)
				)
			: null
	);

	// RN-15 / RN-16 da spec servidor: personagens migrados que o dono ainda não conferiu.
	const toConfirm = $derived(data.characters.filter((c) => c.serverToConfirm));

	// Erros de excluir, de marcar principal e de conferir aparecem no topo da página.
	const pageMessage = $derived(
		form &&
			(form.mode === 'delete' || form.mode === 'main' || form.mode === 'confirm') &&
			'message' in form
			? form.message
			: null
	);
</script>

<svelte:head>
	<title>RO Lobby · meus personagens</title>
</svelte:head>

<div class="page">
	<TopBar user={data.user} loginHref={data.loginHref} />

	<main class="content">
		<div class="head">
			<div class="titles">
				<h1>Meus personagens</h1>
				<span class="count">
					{data.characters.length} de {MAX} personagens · o principal aparece primeiro
				</span>
			</div>
			<div class="add">
				<!-- CA-02.8: com 10 personagens, o botão fica desabilitado e diz por quê. -->
				<Button
					iconLeft="plus"
					disabled={full}
					describedby={full ? limitHintId : undefined}
					onclick={() => show({ kind: 'create' })}>Adicionar personagem</Button
				>
				{#if full}<span class="limit" id={limitHintId}>{LIMIT_MESSAGE}</span>{/if}
			</div>
		</div>

		{#if data.loadError}
			<p class="alert" role="alert">{data.loadError}</p>
		{/if}
		{#if pageMessage}
			<p class="alert" role="alert">{pageMessage}</p>
		{/if}

		{#if data.discordMessage}
			<p
				class="notice"
				class:alert={!data.discordMessage.ok}
				role={data.discordMessage.ok ? 'status' : 'alert'}
				data-testid="discord-message"
			>
				{data.discordMessage.text}
			</p>
		{/if}

		{#if data.discord.enabled}
			<!-- spec discord-bot, US-01 / RN-01 / RN-04: entrar (ou voltar) no servidor oficial. -->
			<section class="discord" aria-labelledby="discord-title" data-testid="discord-connection">
				<div class="confirm-text">
					<h2 id="discord-title">Servidor do RO Lobby no Discord</h2>
					{#if data.discordUnderConstruction}
						<!-- spec discord-em-construcao, RN-02. -->
						<p>{UNDER_CONSTRUCTION.profile}</p>
					{:else if connectedOn}
						<p>Conectado ao servidor do RO Lobby em {connectedOn}.</p>
					{:else}
						<p>Entre no servidor para as salas dos seus grupos aparecerem para você no Discord.</p>
					{/if}
				</div>
				{#if data.discordUnderConstruction}
					<Button variant="secondary" disabled>{UNDER_CONSTRUCTION.button}</Button>
				{:else}
					<Button variant={connectedOn ? 'ghost' : 'primary'} href={resolve('/perfil/discord')}
						>{connectedOn ? 'Reconectar' : 'Entrar no servidor do RO Lobby'}</Button
					>
				{/if}
			</section>
		{/if}

		{#if toConfirm.length > 0}
			<!-- CA-06.1: o aviso some quando não sobra personagem a conferir (CA-06.3). -->
			<section class="confirm" aria-labelledby="confirm-title" data-testid="server-confirm">
				<div class="confirm-text">
					<h2 id="confirm-title">Confira o servidor dos seus personagens</h2>
					<p>
						Agora o RO Lobby sabe em qual servidor cada personagem joga. Os que você já tinha
						ficaram em Freya: confirme ou edite.
					</p>
				</div>
				<ul>
					{#each toConfirm as character (character.id)}
						<li>
							<span class="who"
								><strong>{character.nick}</strong> · {serverName(character.server)}</span
							>
							<form method="POST" action="?/confirm" use:enhance>
								<input type="hidden" name="id" value={character.id} />
								<Button type="submit" size="sm">Está certo</Button>
							</form>
							<Button variant="ghost" size="sm" onclick={() => show({ kind: 'update', character })}
								>Editar</Button
							>
						</li>
					{/each}
				</ul>
			</section>
		{/if}

		{#if data.characters.length === 0 && !data.loadError}
			<!-- CA-01.2 -->
			<section class="empty" aria-labelledby="empty-title">
				<img src="/brand/c1-symbol-dark.svg" alt="" width="48" height="48" />
				<div class="empty-text">
					<h2 id="empty-title">Você ainda não tem personagens</h2>
					<p>Cadastre o primeiro para criar lobbies e se candidatar.</p>
				</div>
				<Button iconLeft="plus" onclick={() => show({ kind: 'create' })}
					>Adicionar personagem</Button
				>
			</section>
		{:else}
			<!-- RN-17: a API já devolve o principal primeiro e os outros por ordem de cadastro. -->
			<ul class="grid" aria-label="Personagens">
				{#each data.characters as character (character.id)}
					<li>
						<CharacterCard
							{character}
							className={classNameOf(character.classId)}
							onedit={() => show({ kind: 'update', character })}
							ondelete={() => show({ kind: 'delete', character })}
							onavailability={() => show({ kind: 'availability', character })}
						/>
					</li>
				{/each}
			</ul>
		{/if}
	</main>
</div>

{#if open}
	{#key open.key}
		{#if open.kind === 'availability'}
			<AvailabilityDialog
				character={open.character}
				instances={data.instances}
				form={dialogForm as AvailabilityForm | null}
				onclose={() => (open = null)}
			/>
		{:else if open.kind === 'delete'}
			<ConfirmDialog
				id={open.character.id}
				nick={open.character.nick}
				onclose={() => (open = null)}
			/>
		{:else}
			<CharacterDialog
				mode={open.kind}
				character={open.kind === 'update' ? open.character : undefined}
				classes={data.classes}
				form={dialogForm as DialogForm | null}
				onclose={() => (open = null)}
			/>
		{/if}
	{/key}
{/if}

<style>
	/* O fundo é o céu do design system, como na Home (vem do body). */
	.page {
		min-height: 100dvh;
	}
	.content {
		display: flex;
		flex-direction: column;
		gap: 22px;
		max-width: 1180px;
		margin: 0 auto;
		padding: 28px 16px 48px;
	}
	@media (min-width: 900px) {
		.content {
			padding: 28px 28px 64px;
		}
	}
	.head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;
	}
	.titles {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	h1 {
		margin: 0;
		font: var(--type-h2);
	}
	.count {
		font: 500 13px/1.3 var(--font-ui);
		color: var(--fg-3);
	}
	.add {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 6px;
	}
	.limit {
		font: var(--type-caption);
		color: var(--fg-3);
	}
	.alert {
		margin: 0;
		padding: 10px 14px;
		border-radius: var(--radius-sm);
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
		font: 500 14px/1.3 var(--font-ui);
	}
	.notice {
		margin: 0;
		padding: 10px 14px;
		border-radius: var(--radius-sm);
		background: rgba(80, 200, 140, 0.12);
		color: var(--support-300);
		font: 500 14px/1.3 var(--font-ui);
	}
	.notice.alert {
		background: rgba(240, 100, 140, 0.12);
		color: var(--status-error);
	}
	.discord {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;
		padding: 16px 18px;
		border: 1px solid var(--border-default);
		border-radius: var(--radius-lg);
		background: var(--panel-bg);
	}
	.discord h2 {
		margin: 0;
		font: var(--type-title);
	}
	.discord p {
		margin: 0;
		color: var(--fg-2);
	}
	.confirm {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 16px 18px;
		border: 1px solid var(--gold-line);
		border-radius: var(--radius-lg);
		background: var(--panel-bg);
	}
	.confirm-text {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.confirm h2 {
		margin: 0;
		font: var(--type-title);
	}
	.confirm p {
		margin: 0;
		color: var(--fg-2);
	}
	.confirm ul {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.confirm li {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}
	.confirm .who {
		flex: 1;
		min-width: 160px;
	}
	.confirm form {
		margin: 0;
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: 20px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	@media (min-width: 1100px) {
		.grid {
			grid-template-columns: repeat(4, 1fr);
		}
	}
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 14px;
		padding: 48px 24px;
		text-align: center;
		background: var(--panel-bg);
		border: 1px dashed var(--line-2);
		border-radius: var(--radius-lg);
	}
	.empty img {
		image-rendering: pixelated;
	}
	.empty-text {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.empty h2 {
		margin: 0;
		font: var(--type-title);
	}
	.empty p {
		margin: 0;
		color: var(--fg-3);
	}
</style>
