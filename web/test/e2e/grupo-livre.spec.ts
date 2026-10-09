import { expect, test, type Page } from '@playwright/test';
import { acceptApplication, apiLogin, createCharacter, rand, spDay } from './seed';

// Ponta a ponta do grupo livre (spec grupo-livre, T-04 a T-06), com o Discord falso e o
// banco do ponta a ponta. O lobby fica no dia 6, longe dos outros testes.

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

test.describe('grupo livre', () => {
	test('CA-01.1 / CA-01.3 / CA-02.5 / CA-03.1 / RNF-03: criar pelo teclado, candidatar com qualquer função e ver na Home', async ({
		browser,
		request
	}) => {
		const s = rand();
		const owner = `gl${s}`;
		const ownerToken = await apiLogin(request, owner);
		await createCharacter(request, ownerToken, {
			nick: `Gl${s}`,
			classId: 'arcebispo',
			level: 178,
			role: 'support'
		});
		const page = await (await browser.newContext()).newPage();
		await loginAs(page, owner, `/lobbies/novo?dia=${spDay(6)}`);

		// CA-01.3: a formação vem "Por função"; pelo teclado, vira "Grupo livre".
		const byRole = page.getByRole('radio', { name: /Por função/ });
		await expect(byRole).toBeChecked();
		await byRole.focus();
		await page.keyboard.press('ArrowRight');
		await expect(page.getByRole('radio', { name: /Grupo livre/ })).toBeChecked();
		await expect(page.getByLabel('Vagas', { exact: true })).toHaveValue('12');
		await page.getByLabel('Instância').selectOption('templo-do-demonio-rei');
		await page.getByLabel('Hora (Brasília)').fill('18:00');
		await page.getByRole('main').getByRole('button', { name: 'Criar lobby' }).click();

		// CA-01.1: 12 lugares, 1 com o anfitrião e 11 abertos.
		await expect(page).toHaveURL(/\/lobbies\/[0-9a-f-]{36}$/);
		const lobbyUrl = page.url();
		const places = page.getByTestId('free-places');
		await expect(places).toContainText('1 de 12');
		await expect(places.getByText('Vaga aberta')).toHaveCount(11);

		// CA-02.5: outro jogador com um Dano se candidata, sem aviso de função.
		const player = `pl${s}`;
		const playerToken = await apiLogin(request, player);
		await createCharacter(request, playerToken, {
			nick: `Pl${s}`,
			classId: 'arquimago',
			level: 200,
			role: 'dps'
		});
		const other = await (await browser.newContext()).newPage();
		await loginAs(other, player, lobbyUrl.replace(/^https?:\/\/[^/]+/, ''));
		await other.getByRole('button', { name: 'Candidatar' }).click();
		const dialog = other.getByRole('dialog');
		await expect(dialog).toContainText('11 vagas livres');
		await dialog
			.getByRole('button', { name: /Enviar candidatura|Candidatar/ })
			.last()
			.click();
		await expect(other.getByText('Candidatura enviada')).toBeVisible();

		// O dono aceita pela API; o grupo fica 2 de 12.
		const id = lobbyUrl.split('/').pop()!;
		const apps = await request.get(
			`http://localhost:${process.env.API_PORT || '8080'}/lobbies/${id}`,
			{ headers: { authorization: `Bearer ${ownerToken}` } }
		);
		const pending = (await apps.json()).pending[0].applicationId;
		await acceptApplication(request, ownerToken, pending);
		await page.reload();
		await expect(page.getByTestId('free-places')).toContainText('2 de 12');
		await expect(page.getByTestId('free-places')).toContainText('Dano');

		// CA-03.1: na Home, o card mostra "Grupo livre" e "2 de 12".
		await page.goto('/');
		await page.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(6).click();
		const card = page.getByTestId('lobby-card').filter({ hasText: `Gl${s}` });
		await expect(card).toContainText('Grupo livre');
		await expect(card).toContainText('2 de 12');
	});
});
