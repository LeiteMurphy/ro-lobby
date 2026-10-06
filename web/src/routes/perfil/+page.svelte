<script lang="ts">
	import CharacterCard from '$lib/characters/components/CharacterCard.svelte';
	import CharacterDialog from '$lib/characters/components/CharacterDialog.svelte';
	import ConfirmDialog from '$lib/characters/components/ConfirmDialog.svelte';
	import type { Character } from '$lib/characters/api';
	import { LIMIT_MESSAGE } from '$lib/characters/messages';
	import TopBar from '$lib/home/components/TopBar.svelte';
	import Button from '$lib/ui/Button.svelte';
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
		| { kind: 'delete'; character: Character };
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
	// Erros de excluir e de marcar principal aparecem no topo da página.
	const pageMessage = $derived(
		form && (form.mode === 'delete' || form.mode === 'main') && 'message' in form
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
						/>
					</li>
				{/each}
			</ul>
		{/if}
	</main>
</div>

{#if open}
	{#key open.key}
		{#if open.kind === 'delete'}
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
				form={dialogForm}
				onclose={() => (open = null)}
			/>
		{/if}
	{/key}
{/if}

<style>
	.page {
		min-height: 100dvh;
		background: var(--ink-0);
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
