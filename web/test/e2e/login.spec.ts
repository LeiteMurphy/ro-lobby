import { expect, test, type Page } from '@playwright/test';

// Ponta a ponta do login com Discord (spec login-discord), contra o Discord falso que o
// playwright.config.ts sobe. O usuário do falso é "grimbold", com o nome "Grimbold".

const loginLink = (page: Page) =>
	page.getByRole('banner').getByRole('link', { name: 'Entrar com Discord' });
const userMenu = (page: Page) => page.getByRole('button', { name: /^Menu de / });

async function loginAs(page: Page, from = '/') {
	await page.goto(from);
	await loginLink(page).click();
	await expect(page.getByRole('heading', { name: 'Discord falso' })).toBeVisible();
	await page.getByRole('button', { name: 'Autorizar' }).click();
}

test.describe('login com Discord', () => {
	test.use({ viewport: { width: 1280, height: 900 } });

	test('CA-01.1 / CA-01.3 / CA-02.1: entrar mostra a inicial e o nome na barra', async ({
		page
	}) => {
		await page.goto('/');
		await loginLink(page).click();
		await page.waitForURL(/127\.0\.0\.1:8090\/oauth2\/authorize/);

		// CA-01.3: a autorização foi pedida com scope=identify, response_type=code e state.
		const authorize = new URL(page.url());
		expect(authorize.searchParams.get('scope')).toBe('identify');
		expect(authorize.searchParams.get('response_type')).toBe('code');
		expect(authorize.searchParams.get('state')?.length).toBeGreaterThan(20);

		await page.getByRole('button', { name: 'Autorizar' }).click();
		await expect(page).toHaveURL('/');
		await expect(userMenu(page)).toContainText('Grimbold');
		await expect(userMenu(page).locator('.tile')).toHaveText('G');
		await expect(loginLink(page)).toHaveCount(0);
	});

	test('CA-06.5: cookie de sessão HttpOnly, SameSite=Lax e Path=/', async ({ page, context }) => {
		await loginAs(page);
		await expect(userMenu(page)).toBeVisible();
		const session = (await context.cookies()).find((c) => c.name === 'rol_session');
		expect(session).toMatchObject({ httpOnly: true, sameSite: 'Lax', path: '/', secure: false });
		expect((await context.cookies()).some((c) => c.name === 'rol_oauth_state')).toBe(false);
	});

	test('CA-03.1: sair encerra a sessão e volta para "Entrar com Discord"', async ({
		page,
		context
	}) => {
		await loginAs(page);
		await userMenu(page).click();
		await page.getByRole('menuitem', { name: 'Sair' }).click();
		await expect(page).toHaveURL('/');
		await expect(loginLink(page)).toBeVisible();
		expect((await context.cookies()).some((c) => c.name === 'rol_session')).toBe(false);
		await page.reload();
		await expect(loginLink(page)).toBeVisible();
	});

	test('CA-04.1: cancelar no Discord volta para a Home com a mensagem de erro', async ({
		page,
		context
	}) => {
		await page.goto('/');
		await loginLink(page).click();
		await page.getByRole('button', { name: 'Cancelar' }).click();
		await expect(page).toHaveURL('/?login=erro');
		await expect(page.getByRole('alert')).toHaveText(
			'Não foi possível entrar com o Discord. Tente de novo.'
		);
		expect((await context.cookies()).some((c) => c.name === 'rol_session')).toBe(false);
	});

	test('CA-04.2: retorno com state diferente é recusado', async ({ page, context }) => {
		await page.goto('/');
		await loginLink(page).click(); // grava o cookie de state certo
		await expect(page.getByRole('heading', { name: 'Discord falso' })).toBeVisible();
		await page.goto('/auth/discord/callback?code=qualquer&state=state-de-outro-lugar');
		await expect(page).toHaveURL('/?login=erro');
		await expect(page.getByRole('alert')).toBeVisible();
		expect((await context.cookies()).some((c) => c.name === 'rol_session')).toBe(false);
	});

	test('CA-05.1: depois de entrar, volta logado para a página de origem', async ({ page }) => {
		// AJ-01: /perfil só abre com sessão, então chegar nela prova que a pessoa voltou logada.
		await page.goto('/auth/discord/login?next=%2Fperfil');
		await page.getByRole('button', { name: 'Autorizar' }).click();
		await expect(page).toHaveURL('/perfil');
		await expect(page.getByRole('heading', { name: 'Meus personagens' })).toBeVisible();
		await expect(userMenu(page)).toContainText('Grimbold');
	});

	test('RNF-02: o menu do usuário funciona pelo teclado', async ({ page }) => {
		await loginAs(page);
		await userMenu(page).focus();
		await page.keyboard.press('Enter');
		// O foco vai para o primeiro item, "Meu perfil" (RN-18 da spec personagens), e a
		// seta para baixo chega em "Sair".
		await expect(page.getByRole('menuitem', { name: 'Meu perfil' })).toBeFocused();
		await page.keyboard.press('ArrowDown');
		const sair = page.getByRole('menuitem', { name: 'Sair' });
		await expect(sair).toBeFocused();
		await page.keyboard.press('Escape');
		await expect(sair).toHaveCount(0);
		await expect(userMenu(page)).toBeFocused();
	});

	test('CA-02.4 / RNF-05: logado, a Home não carrega nada de fora do servidor', async ({
		page,
		baseURL
	}) => {
		await loginAs(page);
		await expect(userMenu(page)).toBeVisible();
		const origins = new Set<string>();
		page.on('request', (req) => origins.add(new URL(req.url()).origin));
		await page.reload();
		await page.waitForLoadState('networkidle');
		expect([...origins].filter((o) => o !== new URL(baseURL!).origin)).toEqual([]);
	});
});
