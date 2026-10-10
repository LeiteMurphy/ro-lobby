import { expect, test, type Page } from '@playwright/test';
import { apiLogin, createCharacter, rand, spDay } from './seed';

// Ponta a ponta do lobby sem instância (spec lobby-sem-instancia, T-02), com o Discord
// falso e o banco do ponta a ponta. O lobby fica no dia 7, longe dos outros testes.

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

test.describe('lobby sem instância', () => {
	test('CA-01.1 / CA-01.4 / CA-01.5 / CA-02.1 / CA-03.1: criar com título, achar no filtro e trocar na edição', async ({
		page,
		request
	}) => {
		const s = rand();
		const owner = `si${s}`;
		const token = await apiLogin(request, owner);
		await createCharacter(request, token, {
			nick: `Si${s}`,
			classId: 'arcebispo',
			level: 178,
			role: 'support'
		});
		await loginAs(page, owner, `/lobbies/novo?dia=${spDay(7)}`);

		// CA-01.5: vem com uma instância do catálogo; CA-01.4: sem instância, o nível vira 1.
		await expect(page.getByLabel('Instância')).not.toHaveValue('__none__');
		await page.getByLabel('Instância').selectOption('__none__');
		await expect(page.getByLabel('Nível mínimo')).toHaveValue('1');
		await page.getByLabel('Título (opcional)').fill(`Caça ao MVP ${s}`);
		await page.getByLabel('Hora (Brasília)').fill('07:00');
		await page.getByRole('main').getByRole('button', { name: 'Criar lobby' }).click();

		// CA-01.1: o detalhe mostra o título.
		await expect(page).toHaveURL(/\/lobbies\/[0-9a-f-]{36}$/);
		const url = page.url();
		await expect(page.getByRole('heading', { level: 1 })).toHaveText(`Caça ao MVP ${s}`);

		// CA-02.1: na Home, o filtro "Sem instância definida" traz o lobby.
		await page.goto('/');
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(7).click();
		const card = page.getByTestId('lobby-card').filter({ hasText: `Caça ao MVP ${s}` });
		await expect(card).toBeVisible();
		const filter = page.locator('aside').getByLabel('Instância');
		await filter.selectOption('__none__');
		await expect(card).toBeVisible();

		// CA-03.1: na edição, volta para "Sonho Sombrio" com o nível 120.
		await page.goto(`${url.replace(/^https?:\/\/[^/]+/, '')}/editar`);
		await expect(page.getByLabel('Instância')).toHaveValue('__none__');
		await expect(page.getByLabel('Título (opcional)')).toHaveValue(`Caça ao MVP ${s}`);
		await page.getByLabel('Instância').selectOption('sonho-sombrio');
		await expect(page.getByLabel('Nível mínimo')).toHaveValue('120');
		await page.getByRole('button', { name: 'Salvar alterações' }).click();
		await expect(page.getByRole('heading', { level: 1 })).toHaveText('Sonho Sombrio');
	});
});
