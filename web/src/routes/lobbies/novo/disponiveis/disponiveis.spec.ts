import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test' } }));

const { GET } = await import('./+server');
const { SESSION_COOKIE } = await import('$lib/auth/oauth');

// Contagem da prévia da criação (spec banco-de-talentos, T-06), com a API falsa.

const QUERY =
	'instanceId=templo-do-demonio-rei&startsAt=2026-10-07T23%3A00%3A00Z&minLevel=160&tank=1&support=2&dps=3&characterId=c1';

const event = (query: string, token: string | undefined, fetchFn: typeof fetch) =>
	({
		url: new URL(`http://web/lobbies/novo/disponiveis?${query}`),
		cookies: { get: (name: string) => (name === SESSION_COOKIE ? token : undefined) },
		fetch: fetchFn
	}) as unknown as Parameters<typeof GET>[0];

describe('/lobbies/novo/disponiveis', () => {
	it('CA-03.1 / RN-10: repassa o formulário com a sessão e devolve a contagem', async () => {
		const fetchFn = vi.fn(
			async () =>
				new Response(JSON.stringify({ count: 3 }), {
					status: 200,
					headers: { 'content-type': 'application/json' }
				})
		) as unknown as typeof fetch;
		const res = await GET(event(QUERY, 'tok', fetchFn));
		expect(await res.json()).toEqual({ count: 3 });
		const [url, init] = (fetchFn as unknown as ReturnType<typeof vi.fn>).mock.calls[0];
		const called = new URL(String(url));
		expect(called.pathname).toBe('/talents/count');
		expect(Object.fromEntries(called.searchParams)).toEqual({
			instanceId: 'templo-do-demonio-rei',
			startsAt: '2026-10-07T23:00:00Z',
			minLevel: '160',
			tank: '1',
			support: '2',
			dps: '3',
			characterId: 'c1'
		});
		expect(new Headers(init?.headers).get('authorization')).toBe('Bearer tok');
	});

	it('RN-10: sem sessão, campo faltando ou API fora do ar, count null', async () => {
		const down = vi.fn(async () => new Response(null, { status: 500 })) as unknown as typeof fetch;
		expect(await (await GET(event(QUERY, undefined, down))).json()).toEqual({ count: null });
		expect(await (await GET(event('instanceId=x', 'tok', down))).json()).toEqual({ count: null });
		expect(await (await GET(event(QUERY, 'tok', down))).json()).toEqual({ count: null });
	});

	it('CA-05.2 / RN-13 da grupo-livre: no grupo livre, manda a formação e o total, sem as funções', async () => {
		const fetchFn = vi.fn(
			async () =>
				new Response(JSON.stringify({ count: 3 }), {
					status: 200,
					headers: { 'content-type': 'application/json' }
				})
		) as unknown as typeof fetch;
		const query =
			'instanceId=templo-do-demonio-rei&startsAt=2026-10-07T23%3A00%3A00Z&minLevel=160&formation=free&freeSlots=12&characterId=c1';
		expect(await (await GET(event(query, 'tok', fetchFn))).json()).toEqual({ count: 3 });
		const [url] = (fetchFn as unknown as ReturnType<typeof vi.fn>).mock.calls[0];
		expect(Object.fromEntries(new URL(String(url)).searchParams)).toEqual({
			instanceId: 'templo-do-demonio-rei',
			startsAt: '2026-10-07T23:00:00Z',
			minLevel: '160',
			characterId: 'c1',
			formation: 'free',
			freeSlots: '12'
		});
	});

	it('CA-04.1 / RN-07 da lobby-sem-instancia: sem instância, manda anyInstance no lugar do instanceId', async () => {
		const fetchFn = vi.fn(
			async () =>
				new Response(JSON.stringify({ count: 1 }), {
					status: 200,
					headers: { 'content-type': 'application/json' }
				})
		) as unknown as typeof fetch;
		const query =
			'anyInstance=true&startsAt=2026-10-07T23%3A00%3A00Z&minLevel=1&tank=1&support=2&dps=3&characterId=c1';
		expect(await (await GET(event(query, 'tok', fetchFn))).json()).toEqual({ count: 1 });
		const [url] = (fetchFn as unknown as ReturnType<typeof vi.fn>).mock.calls[0];
		const params = new URL(String(url)).searchParams;
		expect(params.get('anyInstance')).toBe('true');
		expect(params.has('instanceId')).toBe(false);
	});
});
