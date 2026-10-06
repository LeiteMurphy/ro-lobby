import { describe, expect, it } from 'vitest';
import {
	authorizeUrl,
	cookieOptions,
	decodeStateCookie,
	encodeStateCookie,
	newState,
	redirectUri,
	safeNext,
	sameState
} from './oauth';

describe('regras do OAuth no web', () => {
	it('CA-01.3 / RN-01: URL de autorização com scope=identify, response_type=code e state', () => {
		const url = new URL(
			authorizeUrl({
				authorizeUrl: 'https://discord.com/oauth2/authorize',
				clientId: '123',
				redirectUri: 'http://localhost:3000/auth/discord/callback',
				state: 'abc'
			})
		);
		expect(url.origin + url.pathname).toBe('https://discord.com/oauth2/authorize');
		expect(url.searchParams.get('scope')).toBe('identify');
		expect(url.searchParams.get('response_type')).toBe('code');
		expect(url.searchParams.get('state')).toBe('abc');
		expect(url.searchParams.get('client_id')).toBe('123');
		expect(url.searchParams.get('redirect_uri')).toBe(
			'http://localhost:3000/auth/discord/callback'
		);
	});

	it('RN-02: state aleatório de pelo menos 128 bits', () => {
		const a = newState();
		const b = newState();
		expect(a).not.toBe(b);
		expect(Buffer.from(a, 'base64url').length * 8).toBeGreaterThanOrEqual(128);
	});

	it('D-08: redirect_uri a partir da origem do web', () => {
		expect(redirectUri('http://localhost:5173')).toBe(
			'http://localhost:5173/auth/discord/callback'
		);
	});

	it('CA-05.1 / RN-12: caminho relativo do próprio site é mantido', () => {
		expect(safeNext('/status')).toBe('/status');
		expect(safeNext('/?dia=2026-10-06')).toBe('/?dia=2026-10-06');
	});

	it('CA-05.2 / RN-12: endereço externo ou estranho vira Home', () => {
		for (const bad of [
			'https://exemplo.com',
			'//exemplo.com',
			'/\\exemplo.com',
			'status',
			'',
			null,
			'/\nx'
		]) {
			expect(safeNext(bad), String(bad)).toBe('/');
		}
	});

	it('RN-02: o cookie de state guarda o state e o destino, e recusa conteúdo adulterado', () => {
		expect(decodeStateCookie(encodeStateCookie('abc', '/status'))).toEqual({
			state: 'abc',
			next: '/status'
		});
		expect(decodeStateCookie(encodeStateCookie('abc', '//mal.com'))).toEqual({
			state: 'abc',
			next: '/'
		});
		expect(decodeStateCookie('lixo')).toBeNull();
		expect(decodeStateCookie(undefined)).toBeNull();
	});

	it('RN-02: comparação de state', () => {
		expect(sameState('abc', 'abc')).toBe(true);
		expect(sameState('abc', 'abd')).toBe(false);
		expect(sameState('', '')).toBe(false);
	});

	it('CA-06.5 / RN-08: cookie HttpOnly, SameSite=Lax, Path=/ e Secure fora de localhost', () => {
		expect(cookieOptions(new URL('http://localhost:3000/'), 60)).toEqual({
			httpOnly: true,
			sameSite: 'lax',
			path: '/',
			secure: false,
			maxAge: 60
		});
		expect(cookieOptions(new URL('http://127.0.0.1:5173/'), 60).secure).toBe(false);
		expect(cookieOptions(new URL('https://rolobby.example/'), 60).secure).toBe(true);
	});
});
