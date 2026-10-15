import { expect, test } from '@playwright/test';
import { loginAs, newPage } from './pages';
import { rand } from './seed';

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
