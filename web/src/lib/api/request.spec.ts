import { describe, expect, it } from 'vitest';
import { request } from './request';
import {
	RATE_LIMITED_MESSAGE,
	UNAVAILABLE_MESSAGE,
	unavailableMessage
} from '$lib/characters/messages';

const call = (res: Response) => ({
	fetchFn: (async () => res) as typeof fetch,
	apiBaseUrl: 'http://api.test',
	token: 't'
});

describe('excesso de requisições (spec endurecimento, RN-15)', () => {
	it('CA-05.2: um 429 da API vira rate_limited', async () => {
		const res = new Response('{"error":"rate_limited"}', {
			status: 429,
			headers: { 'content-type': 'application/json', 'retry-after': '40' }
		});
		expect(await request(call(res), 'POST', '/lobbies', {})).toEqual({
			ok: false,
			kind: 'rate_limited'
		});
	});

	it('CA-05.2: a mensagem de rate_limited é a da regra; o resto continua igual', () => {
		expect(unavailableMessage({ kind: 'rate_limited' })).toBe(
			'Muitas tentativas. Aguarde um pouco e tente de novo.'
		);
		expect(RATE_LIMITED_MESSAGE).toBe('Muitas tentativas. Aguarde um pouco e tente de novo.');
		expect(unavailableMessage({ kind: 'unavailable' })).toBe(UNAVAILABLE_MESSAGE);
	});
});

describe('erro de carregamento de página (spec endurecimento, RN-15)', () => {
	it('CA-05.2: página barrada pelo limite mostra "Muitas tentativas"', async () => {
		const { loadErrorMessage } = await import('$lib/characters/messages');
		expect(loadErrorMessage({ ok: true }, { ok: false, kind: 'rate_limited' })).toBe(
			RATE_LIMITED_MESSAGE
		);
		expect(loadErrorMessage({ ok: false, kind: 'unavailable' })).toBe(UNAVAILABLE_MESSAGE);
		expect(loadErrorMessage({ ok: true }, { ok: true })).toBeNull();
	});
});
