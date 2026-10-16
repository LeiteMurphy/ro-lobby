import { expect, test } from '@playwright/test';

// Ponta a ponta dos cabeçalhos de segurança (spec endurecimento, US-04). O e2e roda em
// http://localhost, então o HSTS não aparece (CA-04.2); o caso https fica no teste unitário.

const expected = {
	'x-frame-options': 'DENY',
	'x-content-type-options': 'nosniff',
	'referrer-policy': 'strict-origin-when-cross-origin',
	'permissions-policy': 'camera=(), microphone=(), geolocation=()'
};

for (const [name, path] of [
	['Home', '/'],
	['rota inexistente', '/essa-pagina-nao-existe']
] as const) {
	test(`CA-04.1 / CA-04.2: cabeçalhos de segurança na ${name}`, async ({ request }) => {
		const response = await request.get(path);
		const headers = response.headers();
		for (const [header, value] of Object.entries(expected)) {
			expect(headers[header], header).toBe(value);
		}
		expect(headers['strict-transport-security']).toBeUndefined();
	});
}
