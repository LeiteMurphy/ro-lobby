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
});
