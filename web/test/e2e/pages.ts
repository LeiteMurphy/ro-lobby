import { expect, type Browser, type Page } from '@playwright/test';

// Páginas com sessão para os testes com várias contas (spec candidatura-lobby, T-07 e T-16).

/** Entra pelo Discord falso com o nome dado, a partir de `from`, e volta para lá. */
export async function loginAs(page: Page, username: string, from: string) {
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

/** Uma página num contexto novo, sem os cookies das outras contas. */
export async function newPage(browser: Browser) {
	return (await browser.newContext({ viewport: { width: 1280, height: 900 } })).newPage();
}

export const panel = (page: Page) => page.getByTestId('player-panel');
