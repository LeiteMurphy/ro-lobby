import { expect, test } from '@playwright/test';

// CA-07.1 (fundação) / CA-02.7 (home-local): com banco, backend e web no ar, /status mostra "API online".
test('CA-07.1 / CA-02.7: fluxo completo navegador → web → API → banco', async ({ page }) => {
	await page.goto('/status');
	await expect(page.getByTestId('api-status')).toHaveText('API online');
	await expect(page.locator('body')).not.toContainText(/error|stack/i);
});

// CA-05.4 ponta a ponta: o HTML que o servidor devolve já traz o texto, sem JavaScript.
test('CA-05.4: o HTML do servidor já contém "API online"', async ({ request }) => {
	const response = await request.get('/status');
	expect(response.status()).toBe(200);
	expect(await response.text()).toContain('API online');
});
