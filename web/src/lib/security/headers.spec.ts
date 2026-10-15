import { describe, expect, it } from 'vitest';
import { HSTS, SECURITY_HEADERS, securityHeaders, withSecurityHeaders } from './headers';

describe('cabeçalhos de segurança (spec endurecimento)', () => {
	it('CA-04.1 / RN-11: os quatro cabeçalhos com os valores da regra', () => {
		const headers = securityHeaders(new URL('https://rolobby.com.br/'));
		expect(headers['X-Frame-Options']).toBe('DENY');
		expect(headers['X-Content-Type-Options']).toBe('nosniff');
		expect(headers['Referrer-Policy']).toBe('strict-origin-when-cross-origin');
		expect(headers['Permissions-Policy']).toBe('camera=(), microphone=(), geolocation=()');
	});

	it('CA-04.2 / RN-12: com https, leva HSTS de um ano', () => {
		const headers = securityHeaders(new URL('https://rolobby.com.br/lobbies/1'));
		expect(headers['Strict-Transport-Security']).toBe('max-age=31536000');
		expect(HSTS).toBe('max-age=31536000');
	});

	it('CA-04.2 / RN-12: em http://localhost, sem HSTS', () => {
		const headers = securityHeaders(new URL('http://localhost:3000/'));
		expect(headers['Strict-Transport-Security']).toBeUndefined();
		expect(Object.keys(headers)).toEqual(Object.keys(SECURITY_HEADERS));
	});

	it('CA-04.1: aplica na resposta e mantém status e corpo', async () => {
		const response = withSecurityHeaders(
			new Response('oi', { status: 404 }),
			new URL('https://rolobby.com.br/nao-existe')
		);
		expect(response.status).toBe(404);
		expect(await response.text()).toBe('oi');
		expect(response.headers.get('X-Frame-Options')).toBe('DENY');
		expect(response.headers.get('Strict-Transport-Security')).toBe('max-age=31536000');
	});

	it('CA-04.1: resposta com cabeçalhos imutáveis é copiada com os cabeçalhos', async () => {
		const frozen = Response.redirect('https://rolobby.com.br/', 303);
		expect(() => frozen.headers.set('X-Test', '1')).toThrow();
		const response = withSecurityHeaders(frozen, new URL('https://rolobby.com.br/'));
		expect(response.status).toBe(303);
		expect(response.headers.get('Location')).toBe('https://rolobby.com.br/');
		expect(response.headers.get('X-Content-Type-Options')).toBe('nosniff');
	});
});
