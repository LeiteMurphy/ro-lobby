import { expect, test } from '@playwright/test';
import { loginAs, newPage } from './pages';
import { apiLogin, createCharacter, createLobby, rand } from './seed';

// Ponta a ponta do servidor do jogo (spec servidor), com o Discord falso e o banco do ponta
// a ponta. Os lobbies ficam no dia 11, longe dos outros testes.

test.describe('servidor do jogo', () => {
	test('CA-02.4 / CA-03.6: o detalhe mostra o servidor e a candidatura desabilita outro servidor', async ({
		browser,
		request
	}) => {
		const s = rand();
		const ownerToken = await apiLogin(request, `dono${s}`);
		const host = await createCharacter(request, ownerToken, {
			nick: `Anf${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support',
			server: 'nidhogg'
		});
		const lobby = await createLobby(request, ownerToken, {
			instanceId: 'templo-do-demonio-rei',
			day: 11,
			time: '20:00',
			characterId: host.id,
			minLevel: 160
		});
		const url = `/lobbies/${lobby.id}`;

		const biaToken = await apiLogin(request, `bia${s}`);
		await createCharacter(request, biaToken, {
			nick: `Fre${s}`,
			classId: 'guardiao-real',
			level: 250,
			role: 'tank',
			server: 'freya'
		});
		await createCharacter(request, biaToken, {
			nick: `Nid${s}`,
			classId: 'arquimago',
			level: 250,
			role: 'dps',
			server: 'nidhogg'
		});

		const bia = await newPage(browser);
		await loginAs(bia, `bia${s}`, url);
		// CA-02.4: o servidor no detalhe.
		await expect(bia.getByTestId('lobby-server')).toHaveText('Servidor Nidhogg');

		// CA-03.6: o de Freya fica desabilitado com o motivo; o de Nidhogg pode ser escolhido.
		await bia.getByRole('button', { name: 'Candidatar' }).click();
		const apply = bia.getByTestId('apply-dialog');
		await expect(apply.getByRole('radio', { name: new RegExp(`Fre${s}`) })).toBeDisabled();
		await expect(apply).toContainText('Outro servidor (Freya)');
		await expect(apply.getByRole('radio', { name: new RegExp(`Nid${s}`) })).toBeChecked();
		await apply.getByRole('button', { name: 'Enviar candidatura' }).click();
		await expect(bia.getByTestId('my-application')).toContainText('Sua candidatura está pendente.');
	});

	test('CA-02.2: a criação mostra o servidor do personagem escolhido', async ({
		browser,
		request
	}) => {
		const s = rand();
		const token = await apiLogin(request, `ana${s}`);
		await createCharacter(request, token, {
			nick: `Fre${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support',
			server: 'freya'
		});
		await createCharacter(request, token, {
			nick: `Nid${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support',
			server: 'nidhogg'
		});
		const ana = await newPage(browser);
		await loginAs(ana, `ana${s}`, '/lobbies/novo');
		await ana.getByRole('radio', { name: new RegExp(`Fre${s}`) }).check({ force: true });
		await expect(ana.getByTestId('lobby-server')).toContainText('O lobby fica em Freya');
		await ana.getByRole('radio', { name: new RegExp(`Nid${s}`) }).check({ force: true });
		await expect(ana.getByTestId('lobby-server')).toContainText('O lobby fica em Nidhogg');
	});

	test('CA-04.1 a CA-04.5 / CA-02.4: filtro "Servidor" na Home, com o servidor no card', async ({
		page,
		request
	}) => {
		test.slow();
		await page.setViewportSize({ width: 1280, height: 900 });
		const s = rand();
		const fre = await apiLogin(request, `fre${s}`);
		const nid = await apiLogin(request, `nid${s}`);
		const make = async (token: string, server: 'freya' | 'nidhogg', nick: string, time: string) => {
			const c = await createCharacter(request, token, {
				nick,
				classId: 'arcebispo',
				level: 250,
				role: 'support',
				server
			});
			await createLobby(request, token, {
				instanceId: 'templo-do-demonio-rei',
				day: 12,
				time,
				characterId: c.id,
				minLevel: 160
			});
		};
		await make(fre, 'freya', `FreA${s}`, '19:00');
		await make(fre, 'freya', `FreB${s}`, '22:00');
		await make(nid, 'nidhogg', `NidA${s}`, '20:00');

		await page.goto('/');
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(12).click();
		const cards = page.getByTestId('lobby-card');
		const sidebar = page.getByRole('complementary', { name: 'Filtros' });
		const filter = sidebar.getByTestId('server-filter');

		// CA-04.1: começa em Todos, com as contagens do dia.
		await expect(filter.getByRole('radio', { name: /Todos/ })).toBeChecked();
		await expect(filter).toContainText(/Todos\s*3/);
		await expect(filter).toContainText(/Freya\s*2/);
		await expect(filter).toContainText(/Nidhogg\s*1/);
		await expect(cards).toHaveCount(3);

		// CA-04.2 / CA-02.4: só o de Nidhogg, com o servidor no card.
		await filter.getByText('Nidhogg', { exact: true }).click();
		await expect(cards).toHaveCount(1);
		await expect(cards.first().getByTestId('lobby-card-server')).toHaveText('Nidhogg');

		// CA-04.3 / RN-14: o destaque respeita o servidor escolhido.
		await expect(page.getByTestId('featured-server')).toHaveText('Nidhogg');

		// CA-04.4: "Limpar" volta para Todos.
		await sidebar.getByRole('button', { name: 'Limpar' }).click();
		await expect(filter.getByRole('radio', { name: /Todos/ })).toBeChecked();
		await expect(cards).toHaveCount(3);

		// CA-04.5: recarregar volta para Todos.
		await filter.getByText('Freya', { exact: true }).click();
		await expect(cards).toHaveCount(2);
		await page.reload();
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(12).click();
		await expect(filter.getByRole('radio', { name: /Todos/ })).toBeChecked();
	});
});
