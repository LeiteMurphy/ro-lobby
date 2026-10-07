import { expect, test, type Browser, type Page } from '@playwright/test';
import { apiLogin, createCharacter, createLobby, rand, spDay } from './seed';

// Ponta a ponta dos lobbies (spec lobbies, T-10), com o Discord falso e o banco do ponta a
// ponta. Cada teste usa contas e nicks novos; os lobbies ficam em dias 3 a 5, longe dos da
// Home (dias 1 e 2).

async function loginAs(page: Page, username: string, from: string) {
	await page.goto(from);
	if (!page.url().includes('/oauth2/authorize')) {
		await page.getByRole('banner').getByRole('link', { name: 'Entrar com Discord' }).click();
	}
	await expect(page.getByRole('heading', { name: 'Discord falso' })).toBeVisible();
	await page.getByLabel('Entrar como (opcional)').fill(username);
	await page.getByRole('button', { name: 'Autorizar' }).click();
	await page.waitForURL(
		(url) => !url.href.includes('/oauth2/') && !url.pathname.startsWith('/auth/')
	);
}

async function newPage(browser: Browser) {
	return (await browser.newContext({ viewport: { width: 1280, height: 900 } })).newPage();
}

/** Conta com um personagem, criada pela API, logada no navegador. */
async function ownerWithCharacter(
	browser: Browser,
	request: Parameters<typeof apiLogin>[0],
	s: string,
	from = '/'
) {
	const username = `own${s}`;
	const token = await apiLogin(request, username);
	const character = await createCharacter(request, token, {
		nick: `Lir${s}`,
		classId: 'arcebispo',
		level: 178,
		role: 'support'
	});
	const page = await newPage(browser);
	await loginAs(page, username, from);
	return { page, token, character };
}

const dayLabel = (day: number) => {
	const d = new Date(`${spDay(day)}T12:00:00Z`);
	const weekdays = ['dom', 'seg', 'ter', 'qua', 'qui', 'sex', 'sáb'];
	const months = [
		'jan',
		'fev',
		'mar',
		'abr',
		'mai',
		'jun',
		'jul',
		'ago',
		'set',
		'out',
		'nov',
		'dez'
	];
	return `${weekdays[d.getUTCDay()]}, ${d.getUTCDate()} ${months[d.getUTCMonth()]}`;
};

async function fillLobby(
	page: Page,
	opts: { instance: string; day: number; time: string; note?: string }
) {
	await page.getByLabel('Instância').selectOption(opts.instance);
	await page.getByLabel('Dia').selectOption(spDay(opts.day));
	await page.getByLabel('Hora (Brasília)').fill(opts.time);
	if (opts.note) await page.getByLabel('Observação (opcional)').fill(opts.note);
}

test.describe('lobbies', () => {
	test('CA-02.3 / CA-01.10: visitante vai ao login e volta para a criação; sem personagem, vê o aviso', async ({
		page
	}) => {
		await page.goto('/');
		await page.getByRole('banner').getByRole('link', { name: 'Criar lobby' }).click();
		await expect(page).toHaveURL(/\/oauth2\/authorize/);
		await page.getByLabel('Entrar como (opcional)').fill(`semchar${rand()}`);
		await page.getByRole('button', { name: 'Autorizar' }).click();
		await expect(page).toHaveURL(`/lobbies/novo?dia=${spDay(0)}`);
		await expect(page.getByText('Cadastre um personagem para criar lobbies')).toBeVisible();
		await page.getByRole('link', { name: 'Ir para o perfil' }).click();
		await expect(page).toHaveURL('/perfil');
	});

	test('CA-02.5 / RN-24: "Criar lobby" depois de escolher um dia na Home abre a criação nele', async ({
		browser,
		request
	}) => {
		const { page } = await ownerWithCharacter(browser, request, rand());
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(9).click();
		await page.getByRole('banner').getByRole('link', { name: 'Criar lobby' }).click();
		await expect(page).toHaveURL(`/lobbies/novo?dia=${spDay(9)}`);
		await expect(page.getByLabel('Dia')).toHaveValue(spDay(9));
		await expect(page.getByLabel('Hora (Brasília)')).toHaveValue('20:00');
	});

	test('CA-01.1 / CA-02.1 / CA-02.4 / CA-03.1: criar, ver na Home e abrir pelo "Ver grupo"', async ({
		browser,
		request
	}) => {
		const s = rand();
		const { page, character } = await ownerWithCharacter(browser, request, s);
		await page.getByRole('banner').getByRole('link', { name: 'Criar lobby' }).click();
		await expect(page).toHaveURL(`/lobbies/novo?dia=${spDay(0)}`);
		// CA-01.1: padrões 1/2/3 e o personagem principal marcado.
		await expect(page.getByLabel('Tank', { exact: true })).toHaveValue('1');
		await expect(page.getByLabel('Suporte', { exact: true })).toHaveValue('2');
		await expect(page.getByLabel('Dano', { exact: true })).toHaveValue('3');
		await expect(page.getByRole('radio', { name: new RegExp(character.nick) })).toBeChecked();
		await fillLobby(page, {
			instance: 'templo-do-demonio-rei',
			day: 3,
			time: '20:00',
			note: 'Chamar no Discord antes'
		});
		await expect(page.getByLabel('Nível mínimo')).toHaveValue('160');
		// 1b: a prévia acompanha o formulário.
		const preview = page.getByRole('complementary', { name: 'Prévia do card na Home' });
		await expect(preview).toContainText('Templo do Demônio Rei');
		await expect(preview).toContainText('20:00');
		await page.getByRole('main').getByRole('button', { name: 'Criar lobby' }).click();

		// CA-03.1: o detalhe.
		await expect(page).toHaveURL(/\/lobbies\/[0-9a-f-]{36}$/);
		const detailUrl = page.url();
		await expect(page.getByRole('heading', { level: 1 })).toHaveText('Templo do Demônio Rei');
		await expect(page.getByTestId('lobby-status')).toContainText('Aberto');
		await expect(page.getByRole('main')).toContainText(`${dayLabel(3)} · 20:00`);
		await expect(page.getByRole('main')).toContainText('Nível mínimo 160');
		await expect(page.getByTestId('owner-slot')).toContainText(character.nick);
		await expect(page.getByTestId('owner-slot')).toContainText('Arcebispo · Nv 178');
		await expect(page.getByTestId('owner-slot')).toContainText('Anfitrião');
		await expect(page.getByText('Vaga aberta')).toHaveCount(5);
		await expect(page.getByText('Chamar no Discord antes')).toBeVisible();

		// CA-02.1 / CA-02.4: na Home, no dia, e "Ver grupo" volta ao detalhe.
		await page.goto('/');
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(3).click();
		const card = page.getByTestId('lobby-card').filter({ hasText: character.nick });
		await expect(card).toContainText('Templo do Demônio Rei');
		await expect(card).toContainText('Arcebispo');
		await expect(card).toContainText('Nv 160+');
		await expect(
			card.getByRole('link', { name: /^Ver grupo: Templo do Demônio Rei às 20:00/ })
		).toBeVisible();
		// RN-23: clicar em qualquer ponto do card (aqui, perto do canto da capa) abre o lobby.
		await card.click({ position: { x: 24, y: 24 } });
		await expect(page).toHaveURL(detailUrl);
	});

	test('CA-01.8: conflito de horário recusado junto do campo, sem perder o formulário', async ({
		browser,
		request
	}) => {
		const s = rand();
		const { page, token, character } = await ownerWithCharacter(
			browser,
			request,
			s,
			'/lobbies/novo'
		);
		await createLobby(request, token, {
			instanceId: 'ilha-bios',
			day: 4,
			time: '20:00',
			characterId: character.id,
			minLevel: 160
		});
		await page.reload();
		await fillLobby(page, {
			instance: 'caverna-de-mors',
			day: 4,
			time: '21:30',
			note: 'Segundo grupo'
		});
		await page.getByRole('main').getByRole('button', { name: 'Criar lobby' }).click();
		await expect(page.getByText('Esse personagem já está num grupo nesse horário')).toBeVisible();
		await expect(page.getByLabel('Hora (Brasília)')).toHaveAttribute('aria-invalid', 'true');
		await expect(page.getByLabel('Instância')).toHaveValue('caverna-de-mors');
		await expect(page.getByLabel('Observação (opcional)')).toHaveValue('Segundo grupo');
		// 2 h depois já pode.
		await page.getByLabel('Hora (Brasília)').fill('22:00');
		await page.getByRole('main').getByRole('button', { name: 'Criar lobby' }).click();
		await expect(page).toHaveURL(/\/lobbies\/[0-9a-f-]{36}$/);
	});

	test('CA-03.3 / CA-04.1 / CA-05.1 / CA-06.4: o dono edita e cancela; o outro só vê; o personagem fica travado', async ({
		browser,
		request
	}) => {
		const s = rand();
		const { page, token, character } = await ownerWithCharacter(browser, request, s);
		const lobby = await createLobby(request, token, {
			instanceId: 'templo-do-demonio-rei',
			day: 5,
			time: '20:00',
			characterId: character.id,
			minLevel: 160
		});
		const url = `/lobbies/${lobby.id}`;

		// CA-03.3: outra conta vê "Candidatar" (candidatura-lobby) e nenhuma ação do dono.
		const other = await newPage(browser);
		await loginAs(other, `oth${s}`, url);
		await expect(other.getByRole('button', { name: 'Candidatar' })).toBeEnabled();
		await expect(other.getByRole('link', { name: 'Editar' })).toHaveCount(0);
		await expect(other.getByRole('button', { name: 'Cancelar lobby' })).toHaveCount(0);
		await other.goto(`${url}/editar`);
		await expect(other.getByTestId('error-page')).toContainText('Página não encontrada');

		// CA-06.4: no perfil, o personagem dono não é excluído.
		await page.goto('/perfil');
		await page.getByRole('button', { name: `Ações de ${character.nick}` }).click();
		await page.getByRole('menuitem', { name: 'Excluir' }).click();
		await page.getByTestId('confirm-dialog').getByRole('button', { name: 'Excluir' }).click();
		await expect(page.getByRole('alert')).toHaveText(
			'Esse personagem está num lobby aberto. Saia ou cancele antes de mudar nível ou função.'
		);
		await expect(page.getByTestId('character-card')).toHaveCount(1);

		// CA-04.1: o dono edita horário e vagas.
		await page.goto(url);
		await page.getByRole('link', { name: 'Editar' }).click();
		await expect(page).toHaveURL(`${url}/editar`);
		await page.getByLabel('Hora (Brasília)').fill('21:00');
		await page.getByRole('button', { name: 'Mais uma vaga de Dano' }).click();
		await page.getByRole('button', { name: 'Mais uma vaga de Dano' }).click();
		await page.getByRole('button', { name: 'Salvar alterações' }).click();
		await expect(page).toHaveURL(url);
		await expect(page.getByRole('main')).toContainText(`${dayLabel(5)} · 21:00`);
		await expect(page.getByText('0 de 5')).toBeVisible();
		// A Home também mostra o horário e as vagas novas.
		await page.goto('/');
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(5).click();
		const edited = page.getByTestId('lobby-card').filter({ hasText: character.nick });
		await expect(edited).toContainText('21:00');
		await expect(edited).toContainText('1/8');
		await page.goto(url);

		// CA-05.1: cancelar com motivo curto, depois válido.
		await page.getByRole('button', { name: 'Cancelar lobby' }).click();
		const dialog = page.getByTestId('cancel-dialog');
		await dialog.getByLabel('Motivo').fill('não dá');
		await dialog.getByRole('button', { name: 'Cancelar lobby' }).click();
		await expect(dialog.getByText('Escreva de 10 a 250 caracteres')).toBeVisible();
		await dialog.getByLabel('Motivo').fill('Metade do grupo não pode hoje');
		await dialog.getByRole('button', { name: 'Cancelar lobby' }).click();
		await expect(dialog).toHaveCount(0);
		await expect(page.getByTestId('lobby-status')).toHaveText('Cancelado');
		await expect(page.getByText('Metade do grupo não pode hoje')).toBeVisible();

		// CA-02.2: cancelado sai da Home.
		await page.goto('/');
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(5).click();
		await expect(page.getByTestId('lobby-card').filter({ hasText: character.nick })).toHaveCount(0);
	});

	test('RNF-01: formulário e diálogo de cancelamento pelo teclado', async ({
		browser,
		request
	}) => {
		const s = rand();
		const { page, token, character } = await ownerWithCharacter(
			browser,
			request,
			s,
			'/lobbies/novo'
		);
		await page.getByLabel('Instância').focus();
		await page.keyboard.press('Tab');
		await expect(page.getByLabel('Dia')).toBeFocused();
		await page.getByRole('button', { name: 'Mais uma vaga de Tank' }).focus();
		await page.keyboard.press('Enter');
		await expect(page.getByLabel('Tank', { exact: true })).toHaveValue('2');

		const lobby = await createLobby(request, token, {
			instanceId: 'sala-final',
			day: 3,
			time: '23:00',
			characterId: character.id,
			minLevel: 150
		});
		await page.goto(`/lobbies/${lobby.id}`);
		await page.getByRole('button', { name: 'Cancelar lobby' }).focus();
		await page.keyboard.press('Enter');
		await expect(page.getByTestId('cancel-dialog').getByLabel('Motivo')).toBeFocused();
		await page.keyboard.press('Escape');
		await expect(page.getByTestId('cancel-dialog')).toHaveCount(0);
	});

	test('RNF-02: criação e detalhe não carregam nada de fora do servidor', async ({
		browser,
		request,
		baseURL
	}) => {
		const s = rand();
		const { page, token, character } = await ownerWithCharacter(
			browser,
			request,
			s,
			'/lobbies/novo'
		);
		const lobby = await createLobby(request, token, {
			instanceId: 'sala-final',
			day: 4,
			time: '23:30',
			characterId: character.id,
			minLevel: 150
		});
		const origins = new Set<string>();
		page.on('request', (req) => origins.add(new URL(req.url()).origin));
		for (const path of ['/lobbies/novo', `/lobbies/${lobby.id}`]) {
			await page.goto(path);
			await page.waitForLoadState('networkidle');
		}
		expect([...origins].filter((o) => o !== new URL(baseURL!).origin)).toEqual([]);
	});
});
