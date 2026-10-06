import { describe, expect, it, vi } from 'vitest';
import { TEMPLE } from '$lib/lobbies/fixtures';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test' } }));

const { load } = await import('./+page.server');

// load da Home com a API de lobbies (spec lobbies, RN-22, D-09).

function fakeFetch(responses: Record<string, Response>) {
	const urls: string[] = [];
	const fn = (async (input: string | URL) => {
		const url = new URL(String(input));
		urls.push(url.pathname + url.search);
		const r = responses[url.pathname];
		if (!r) throw new TypeError(`sem resposta para ${url.pathname}`);
		return r.clone();
	}) as typeof fetch;
	return { fn, urls };
}

const json = (status: number, body: unknown) =>
	new Response(JSON.stringify(body), { status, headers: { 'content-type': 'application/json' } });

type Event = Parameters<typeof load>[0];
const event = (f: typeof fetch) => ({ url: new URL('http://web/'), fetch: f }) as unknown as Event;

describe('load da Home', () => {
	it('CA-02.1 / RN-22: busca os lobbies dos 14 dias e converte para a Home', async () => {
		const { fn, urls } = fakeFetch({
			'/lobbies': json(200, [TEMPLE]),
			'/classes': json(200, [{ id: 'arcebispo', name: 'Arcebispo' }])
		});
		const data = (await load(event(fn))) as Record<string, unknown>;
		const today = data.today as string;
		const last = new Date(`${today}T12:00:00Z`);
		last.setUTCDate(last.getUTCDate() + 13);
		expect(urls).toContain(`/lobbies?from=${today}&to=${last.toISOString().slice(0, 10)}`);
		expect(data.lobbies).toEqual([
			expect.objectContaining({
				id: TEMPLE.id,
				host: 'Lirien',
				hostClass: 'Arcebispo',
				time: '20:00'
			})
		]);
		expect(data.loadError).toBeNull();
	});

	it('RN-21 (personagens): API fora do ar deixa a lista vazia com o aviso', async () => {
		const { fn } = fakeFetch({ '/lobbies': json(500, {}), '/classes': json(500, {}) });
		const data = (await load(event(fn))) as Record<string, unknown>;
		expect(data.lobbies).toEqual([]);
		expect(data.loadError).toBe('Não foi possível falar com o servidor. Tente de novo.');
	});
});
