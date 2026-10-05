import { describe, expect, it } from 'vitest';
import { displayName, initial, loginHref } from './display';

describe('exibição do usuário', () => {
	it('CA-02.1 / RN-15: nome de exibição e inicial', () => {
		expect(displayName({ username: 'grimbold', globalName: 'Grimbold' })).toBe('Grimbold');
		expect(initial('Grimbold')).toBe('G');
	});

	it('CA-02.2 / RN-15: sem nome de exibição, usa o nome de usuário', () => {
		expect(displayName({ username: 'mirai.exe', globalName: null })).toBe('mirai.exe');
		expect(displayName({ username: 'mirai.exe', globalName: '  ' })).toBe('mirai.exe');
		expect(initial('mirai.exe')).toBe('M');
	});

	it('RN-15: inicial com acento e emoji não quebra', () => {
		expect(initial('élise')).toBe('É');
		expect(initial('')).toBe('?');
	});

	it('CA-05.1 / RN-12: o link de login volta para a página atual, sem o aviso de erro', () => {
		expect(loginHref(new URL('http://localhost:3000/status'))).toBe(
			'/auth/discord/login?next=%2Fstatus'
		);
		expect(loginHref(new URL('http://localhost:3000/?login=erro'))).toBe(
			'/auth/discord/login?next=%2F'
		);
	});
});
