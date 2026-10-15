import { describe, expect, it } from 'vitest';
import { CLIENT_IP_HEADER, clientIp, withClientIp } from './client-ip';

const API = 'http://api:8080';

describe('IP do visitante para a API (spec endurecimento, RN-16)', () => {
	it('CA-05.5: pelo túnel, usa o CF-Connecting-IP', () => {
		const request = new Request('https://rolobby.com.br/', {
			headers: { 'CF-Connecting-IP': '203.0.113.7' }
		});
		expect(clientIp(request, () => '172.18.0.9')).toBe('203.0.113.7');
	});

	it('RN-16: fora do túnel, usa o endereço da conexão', () => {
		const request = new Request('http://localhost:3000/');
		expect(clientIp(request, () => '127.0.0.1')).toBe('127.0.0.1');
	});

	it('RN-16: sem endereço disponível, não quebra', () => {
		const request = new Request('http://localhost:3000/');
		expect(
			clientIp(request, () => {
				throw new Error('sem endereço');
			})
		).toBe('');
	});

	it('CA-05.5: a chamada à API leva o IP', () => {
		const forwarded = withClientIp(new Request(`${API}/lobbies`), API, '203.0.113.7');
		expect(forwarded.headers.get(CLIENT_IP_HEADER)).toBe('203.0.113.7');
	});

	it('RN-05: chamada para fora da API não leva o IP', () => {
		const request = new Request('https://discord.com/api/users/@me');
		expect(withClientIp(request, API, '203.0.113.7').headers.get(CLIENT_IP_HEADER)).toBeNull();
	});

	it('CA-05.5: um X-Client-IP forjado pelo navegador é substituído', () => {
		const request = new Request(`${API}/lobbies`, { headers: { [CLIENT_IP_HEADER]: '6.6.6.6' } });
		expect(withClientIp(request, API, '203.0.113.7').headers.get(CLIENT_IP_HEADER)).toBe(
			'203.0.113.7'
		);
	});
});
