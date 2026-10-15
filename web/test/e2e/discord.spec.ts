import { expect, test } from '@playwright/test';
import { loginAs, newPage } from './pages';
import { acceptApplication, apiLogin, applyTo, createCharacter, createLobby, rand } from './seed';

// Ponta a ponta da sala do grupo no Discord (spec discord-bot), com o Discord falso.

async function authorizeAs(page: import('@playwright/test').Page, username: string) {
	await expect(page.getByRole('heading', { name: 'Discord falso' })).toBeVisible();
	await page.getByLabel('Entrar como (opcional)').fill(username);
	await page.getByRole('button', { name: 'Autorizar' }).click();
	await page.waitForURL((url) => url.pathname === '/perfil');
}

test.describe('sala do grupo no Discord: conexão ao servidor', () => {
	test('CA-01.1 / CA-01.3 / CA-01.6: entrar no servidor e reconectar', async ({ browser }) => {
		const s = rand();
		const page = await newPage(browser);
		await loginAs(page, `ana${s}`, '/perfil');
		const section = page.getByTestId('discord-connection');
		await section.getByRole('link', { name: 'Entrar no servidor do RO Lobby' }).click();
		// CA-01.3: a autorização pede identify e guilds.join.
		await expect(page.locator('code')).toHaveText('identify guilds.join');
		await authorizeAs(page, `ana${s}`);
		await expect(page.getByTestId('discord-message')).toHaveText(
			'Você entrou no servidor do RO Lobby.'
		);
		await expect(section).toContainText('Conectado ao servidor do RO Lobby em');
		await section.getByRole('link', { name: 'Reconectar' }).click();
		await authorizeAs(page, `ana${s}`);
		await expect(page.getByTestId('discord-message')).toHaveText(
			'Você entrou no servidor do RO Lobby.'
		);
	});

	test('CA-01.4: outra conta do Discord é recusada', async ({ browser }) => {
		const s = rand();
		const page = await newPage(browser);
		await loginAs(page, `ana${s}`, '/perfil');
		await page
			.getByTestId('discord-connection')
			.getByRole('link', { name: /Entrar no servidor/ })
			.click();
		await authorizeAs(page, `outra${s}`);
		await expect(page.getByTestId('discord-message')).toHaveText(
			'Use a mesma conta do Discord do seu login.'
		);
		await expect(page.getByTestId('discord-connection')).toContainText(
			'Entrar no servidor do RO Lobby'
		);
	});

	test('CA-01.5: cancelar a autorização volta com "Conexão cancelada."', async ({ browser }) => {
		const s = rand();
		const page = await newPage(browser);
		await loginAs(page, `ana${s}`, '/perfil');
		await page
			.getByTestId('discord-connection')
			.getByRole('link', { name: /Entrar no servidor/ })
			.click();
		await page.getByRole('button', { name: 'Cancelar' }).click();
		await page.waitForURL((url) => url.pathname === '/perfil');
		await expect(page.getByTestId('discord-message')).toHaveText('Conexão cancelada.');
	});
});

test.describe('sala do grupo no Discord: detalhe do lobby', () => {
	test('CA-02.1 / CA-03.1 / CA-03.3 / CA-03.4 / CA-08.1 / CA-08.2: abrir a sala, ver quem está e avisar', async ({
		browser,
		request
	}) => {
		const s = rand();
		const ownerToken = await apiLogin(request, `dono${s}`);
		const host = await createCharacter(request, ownerToken, {
			nick: `Anf${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support'
		});
		const lobby = await createLobby(request, ownerToken, {
			instanceId: 'templo-do-demonio-rei',
			day: 13,
			time: '20:00',
			characterId: host.id,
			minLevel: 160
		});
		const url = `/lobbies/${lobby.id}`;
		const memberToken = await apiLogin(request, `mem${s}`);
		const brasa = await createCharacter(request, memberToken, {
			nick: `Bra${s}`,
			classId: 'guardiao-real',
			level: 250,
			role: 'tank'
		});
		const app = await applyTo(request, memberToken, lobby.id, brasa.id);
		await acceptApplication(request, ownerToken, app.id);

		// CA-03.4: o visitante não vê a seção.
		const visitor = await newPage(browser);
		await visitor.goto(url);
		await expect(visitor.getByRole('heading', { name: 'Templo do Demônio Rei' })).toBeVisible();
		await expect(visitor.getByTestId('discord-section')).toHaveCount(0);

		// CA-03.3: o membro vê que a sala ainda não foi aberta.
		const member = await newPage(browser);
		await loginAs(member, `mem${s}`, url);
		await expect(member.getByTestId('room-not-open')).toHaveText('A sala ainda não foi aberta.');

		// CA-02.1: o dono abre a sala.
		const owner = await newPage(browser);
		await loginAs(owner, `dono${s}`, url);
		const section = owner.getByTestId('discord-section');
		await section.getByRole('button', { name: 'Abrir sala no Discord' }).click();
		await expect(section).toContainText('Sala aberta no Discord.');
		await expect(section.getByTestId('room-link')).toHaveAttribute(
			'href',
			/^https:\/\/discord\.com\/channels\/900000000000000001\//
		);
		await expect(section).toContainText(`Bra${s}`);

		// CA-03.1 / CA-03.2: o membro vê o link e, sem conexão, o atalho.
		await member.reload();
		await expect(member.getByTestId('room-link')).toBeVisible();
		await expect(
			member
				.getByTestId('discord-section')
				.getByRole('link', { name: 'Entrar no servidor do RO Lobby' })
		).toBeVisible();
		await expect(member.getByRole('button', { name: 'Avisar no Discord' })).toHaveCount(0);

		// CA-08.1 / CA-08.2: o dono avisa e o botão entra no intervalo.
		await section.getByRole('button', { name: 'Avisar no Discord' }).click();
		await expect(section).toContainText('Grupo avisado no Discord.');
		await expect(section.getByRole('button', { name: 'Avisar no Discord' })).toBeDisabled();
		await expect(section.getByTestId('alert-wait')).toHaveText(
			/Aguarde (9|10) min para avisar de novo/
		);
	});
});
