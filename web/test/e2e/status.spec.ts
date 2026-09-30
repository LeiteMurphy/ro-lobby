import { expect, test } from '@playwright/test';

// CA-07.1: com banco, backend e web no ar, a página de status mostra "API online".
test('CA-07.1: fluxo completo navegador → web → API → banco', async ({ page }) => {
	await page.goto('/');
	await expect(page.getByTestId('api-status')).toHaveText('API online');
	await expect(page.locator('body')).not.toContainText(/error|stack/i);
});

// CA-05.4 ponta a ponta: o HTML que o servidor devolve já traz o texto, sem JavaScript.
test('CA-05.4: o HTML do servidor já contém "API online"', async ({ request }) => {
	const response = await request.get('/');
	expect(response.status()).toBe(200);
	expect(await response.text()).toContain('API online');
});
