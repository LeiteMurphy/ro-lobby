import { expect, test, type Page } from '@playwright/test';
import { apiLogin, createCharacter, createLobby, rand, setAvailability, spDay } from './seed';

// Ponta a ponta do banco de talentos (spec banco-de-talentos, T-05 a T-07), com o Discord
// falso e o banco do ponta a ponta. Cada teste usa contas e nicks novos.

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

const card = (page: Page, nick: string) =>
	page
		.getByTestId('character-card')
		.filter({ has: page.getByRole('heading', { name: nick, exact: true }) });
const dialog = (page: Page) => page.getByTestId('availability-dialog');

test.describe('banco de talentos', () => {
	test('CA-01.1 / CA-01.3 / CA-01.4 / RNF-03: ligar pelo teclado, errar, salvar, desligar e reabrir', async ({
		page,
		request
	}) => {
		const s = rand();
		const username = `tal${s}`;
		const token = await apiLogin(request, username);
		const nick = `Lir${s}`;
		await createCharacter(request, token, {
			nick,
			classId: 'arcebispo',
			level: 178,
			role: 'support'
		});
		await loginAs(page, username, '/perfil');

		const toggle = card(page, nick).getByRole('button', { name: /Banco de talentos de/ });
		await expect(toggle).toHaveText(/Fora do banco de talentos/);
		await toggle.focus();
		await page.keyboard.press('Enter');
		const d = dialog(page);
		await expect(d).toBeVisible();

		// CA-01.3: sem dia e sem instância, os erros voltam nos campos.
		await d.getByLabel('Qualquer instância').uncheck();
		await d.getByRole('button', { name: 'Salvar' }).click();
		await expect(d.getByText('Escolha pelo menos um dia')).toBeVisible();
		await expect(d.getByText('Escolha uma instância ou marque Qualquer instância')).toBeVisible();

		// CA-01.1: segunda a sexta, 19:00 às 23:00, Templo e Sonho Sombrio.
		for (const day of ['segunda', 'terça', 'quarta', 'quinta', 'sexta']) {
			await d.getByLabel(day, { exact: true }).check({ force: true });
		}
		await d.getByLabel('Das (Brasília)').selectOption('19:00');
		await d.getByLabel('Até').selectOption('23:00');
		await d.getByLabel(/Templo do Demônio Rei/).check();
		await d.getByLabel(/Sonho Sombrio/).check();
		await d.getByRole('button', { name: 'Salvar' }).click();
		await expect(d).toHaveCount(0);
		await expect(toggle).toHaveText(/No banco de talentos/);

		// CA-01.4: desligar e reabrir com os dados guardados.
		await toggle.click();
		await dialog(page).getByLabel('Disponível no banco de talentos').uncheck();
		await dialog(page).getByRole('button', { name: 'Salvar' }).click();
		await expect(dialog(page)).toHaveCount(0);
		await expect(toggle).toHaveText(/Fora do banco de talentos/);
		await toggle.click();
		const again = dialog(page);
		await expect(again.getByLabel('segunda', { exact: true })).toBeChecked();
		await expect(again.getByLabel('sábado', { exact: true })).not.toBeChecked();
		await expect(again.getByLabel('Das (Brasília)')).toHaveValue('19:00');
		await expect(again.getByLabel(/Sonho Sombrio/)).toBeChecked();
		await page.keyboard.press('Escape');
		await expect(again).toHaveCount(0);
	});

	test('CA-02.1 / CA-02.4 / CA-03.1: o dono vê quem combina; a criação conta os disponíveis', async ({
		browser,
		request
	}) => {
		const s = rand();
		// Brasa (Tank, 172) disponível todo dia das 05:00 às 07:00, no Templo.
		const biaToken = await apiLogin(request, `bia${s}`);
		const brasa = await createCharacter(request, biaToken, {
			nick: `Bra${s}`,
			classId: 'guardiao-real',
			level: 172,
			role: 'tank'
		});
		await setAvailability(request, biaToken, brasa.id, {
			days: [0, 1, 2, 3, 4, 5, 6],
			start: '05:00',
			end: '07:00',
			instanceIds: ['templo-do-demonio-rei']
		});
		const owner = `ana${s}`;
		const anaToken = await apiLogin(request, owner);
		const lirien = await createCharacter(request, anaToken, {
			nick: `Lir${s}`,
			classId: 'arcebispo',
			level: 178,
			role: 'support'
		});
		const lobby = await createLobby(request, anaToken, {
			instanceId: 'templo-do-demonio-rei',
			day: 4,
			time: '06:00',
			characterId: lirien.id,
			minLevel: 160
		});

		// CA-02.1: o dono vê Brasa com o Discord.
		const page = await (await browser.newContext()).newPage();
		await loginAs(page, owner, `/lobbies/${lobby.id}`);
		const panel = page.getByTestId('talents-panel');
		await expect(panel.getByTestId('talent-card').filter({ hasText: brasa.nick })).toContainText(
			`@bia${s}`
		);

		// CA-02.4: o visitante não vê o painel.
		const visitor = await (await browser.newContext()).newPage();
		await visitor.goto(`/lobbies/${lobby.id}`);
		await expect(visitor.getByRole('heading', { level: 1 })).toBeVisible();
		await expect(visitor.getByTestId('talents-panel')).toHaveCount(0);

		// CA-03.1: na criação, 06:00 tem 1 disponível; 09:00, nenhum.
		await page.goto(`/lobbies/novo?dia=${spDay(5)}`);
		await page.getByLabel('Instância').selectOption('templo-do-demonio-rei');
		await page.getByLabel('Hora (Brasília)').fill('06:00');
		await expect(page.getByTestId('available-count')).toHaveText(
			/1 jogador disponível no banco de talentos/
		);
		await page.getByLabel('Hora (Brasília)').fill('09:00');
		await expect(page.getByTestId('available-count')).toHaveText(
			/0 jogadores disponíveis no banco de talentos/
		);
	});

	test('CA-04.1 / CA-04.2 / CA-04.3 / CA-01.2: catálogo com filtros e Discord só para logado', async ({
		browser,
		request
	}) => {
		const s = rand();
		const tok = await apiLogin(request, `cat${s}`);
		const tank = await createCharacter(request, tok, {
			nick: `Esc${s}`,
			classId: 'guardiao-real',
			level: 172,
			role: 'tank'
		});
		const night = await createCharacter(request, tok, {
			nick: `Nox${s}`,
			classId: 'arquimago',
			level: 200,
			role: 'dps'
		});
		const off = await createCharacter(request, tok, {
			nick: `Off${s}`,
			classId: 'arquimago',
			level: 190,
			role: 'dps'
		});
		await setAvailability(request, tok, tank.id, {
			days: [3],
			start: '19:00',
			end: '23:00',
			instanceIds: ['templo-do-demonio-rei']
		});
		await setAvailability(request, tok, night.id, {
			days: [5],
			start: '22:00',
			end: '02:00',
			anyInstance: true
		});
		await setAvailability(request, tok, off.id, {
			days: [3],
			start: '19:00',
			end: '23:00',
			anyInstance: true
		});
		await request.put(
			`http://localhost:${process.env.API_PORT || '8080'}/characters/${off.id}/availability`,
			{
				data: { enabled: false },
				headers: { authorization: `Bearer ${tok}` }
			}
		);

		const visitor = await (await browser.newContext()).newPage();
		const cardOf = (nick: string) => visitor.getByTestId('talent-card').filter({ hasText: nick });

		// CA-04.2: o visitante vê o catálogo sem o Discord; o nome não vai no HTML.
		const res = await visitor.goto('/talentos');
		expect(await res!.text()).not.toContain(`@cat${s}`);
		await expect(cardOf(tank.nick)).toContainText('Entre para ver o Discord');
		// CA-04.3: o desligado não aparece.
		await expect(cardOf(off.nick)).toHaveCount(0);

		// CA-04.1: por função Tank, só o tank; por Sonho Sombrio, só quem marcou Qualquer.
		await visitor.getByLabel('Função').selectOption('tank');
		await visitor.getByRole('button', { name: 'Filtrar' }).click();
		await expect(visitor).toHaveURL(/funcao=tank/);
		await expect(cardOf(tank.nick)).toBeVisible();
		await expect(cardOf(night.nick)).toHaveCount(0);
		await visitor.goto('/talentos?instancia=sonho-sombrio');
		await expect(cardOf(night.nick)).toBeVisible();
		await expect(cardOf(tank.nick)).toHaveCount(0);

		// CA-01.2: sexta 22:00–02:00 aparece no sábado à 01:00, não no sábado às 22:00.
		await visitor.goto('/talentos?dia=6&hora=01:00');
		await expect(cardOf(night.nick)).toBeVisible();
		await visitor.goto('/talentos?dia=6&hora=22:00');
		await expect(visitor.getByRole('heading', { name: 'Banco de talentos' })).toBeVisible();
		await expect(cardOf(night.nick)).toHaveCount(0);

		// CA-04.2: logado, vê o Discord; o link da TopBar leva ao catálogo.
		const page = await (await browser.newContext()).newPage();
		await loginAs(page, `ver${s}`, '/');
		await page.getByRole('banner').getByRole('link', { name: 'Banco de talentos' }).click();
		await expect(page).toHaveURL('/talentos');
		await expect(page.getByTestId('talent-card').filter({ hasText: tank.nick })).toContainText(
			`@cat${s}`
		);
	});
});
