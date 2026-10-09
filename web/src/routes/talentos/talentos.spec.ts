import { render } from 'svelte/server';
import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test' } }));

const { load } = await import('./+page.server');
const Page = (await import('./+page.svelte')).default;
const { SESSION_COOKIE } = await import('$lib/auth/oauth');

// Catálogo /talentos (spec banco-de-talentos, T-07): o load com a API falsa e a página
// renderizada no servidor.

const json = (status: number, body?: unknown) =>
	new Response(body === undefined ? null : JSON.stringify(body), {
		status,
		headers: { 'content-type': 'application/json' }
	});

function fakeFetch(talents: Response) {
	const calls: { url: URL; auth: string | null }[] = [];
	const fn = (async (input: string | URL, init?: RequestInit) => {
		const url = new URL(String(input));
		calls.push({ url, auth: new Headers(init?.headers).get('authorization') });
		if (url.pathname === '/talents') return talents.clone();
		return json(200, []);
	}) as typeof fetch;
	return { fn, calls };
}

type LoadEvent = Parameters<typeof load>[0];
const loadEvent = (query: string, user: unknown, f: typeof fetch) =>
	({
		url: new URL(`http://web/talentos${query}`),
		locals: { user },
		cookies: { get: (n: string) => (n === SESSION_COOKIE && user ? 'tok' : undefined) },
		fetch: f
	}) as unknown as LoadEvent;

const BRASA = {
	characterId: 'b1',
	nick: 'Brasa',
	classId: 'guardiao-real',
	level: 172,
	role: 'tank',
	portrait: 'retrato-3',
	link: null,
	days: [3],
	start: '19:00',
	end: '23:00',
	anyInstance: false,
	instances: [{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei' }]
};
const USER = { id: 'u1', username: 'ana', globalName: null };

describe('load de /talentos', () => {
	it('CA-04.1 / CA-01.2 / RN-11: os filtros da URL vão para a API', async () => {
		const { fn, calls } = fakeFetch(json(200, [BRASA]));
		const data = (await load(
			loadEvent('?instancia=templo-do-demonio-rei&funcao=tank&dia=6&hora=01:00', null, fn)
		)) as Record<string, unknown>;
		const api = calls.find((c) => c.url.pathname === '/talents')!;
		expect(Object.fromEntries(api.url.searchParams)).toEqual({
			instanceId: 'templo-do-demonio-rei',
			role: 'tank',
			day: '6',
			time: '01:00'
		});
		expect(data.talents).toEqual([BRASA]);
		expect(data.filters).toEqual({
			instancia: 'templo-do-demonio-rei',
			funcao: 'tank',
			dia: '6',
			hora: '01:00'
		});
	});

	it('RN-11: valores fora da lista são ignorados', async () => {
		const { fn, calls } = fakeFetch(json(200, []));
		const data = (await load(loadEvent('?funcao=healer&dia=9&hora=01:15', null, fn))) as {
			filters: Record<string, string>;
		};
		expect([...calls.find((c) => c.url.pathname === '/talents')!.url.searchParams]).toEqual([]);
		expect(data.filters).toEqual({ instancia: '', funcao: '', dia: '', hora: '' });
	});

	it('CA-04.2 / D-05: a sessão só vai para a API com o Usuário logado', async () => {
		const anon = fakeFetch(json(200, []));
		await load(loadEvent('', null, anon.fn));
		expect(anon.calls.find((c) => c.url.pathname === '/talents')!.auth).toBeNull();
		const logged = fakeFetch(json(200, []));
		await load(loadEvent('', USER, logged.fn));
		expect(logged.calls.find((c) => c.url.pathname === '/talents')!.auth).toBe('Bearer tok');
	});

	it('risco do design: 100 resultados ligam o aviso; filtro recusado mostra o erro', async () => {
		const many = Array.from({ length: 100 }, (_, i) => ({ ...BRASA, characterId: `c${i}` }));
		const full = (await load(loadEvent('', null, fakeFetch(json(200, many)).fn))) as {
			limited: boolean;
		};
		expect(full.limited).toBe(true);
		const bad = (await load(
			loadEvent('', null, fakeFetch(json(422, { error: 'validation', fields: [] })).fn)
		)) as { loadError: string; talents: unknown[] };
		expect(bad.loadError).toBe('Algum filtro não vale. Limpe os filtros e tente de novo.');
		expect(bad.talents).toEqual([]);
	});
});

describe('/talentos renderizada no servidor', () => {
	const renderPage = (over: Record<string, unknown>) =>
		render(Page, {
			props: {
				data: {
					user: null,
					talents: [BRASA],
					limited: false,
					filters: { instancia: '', funcao: '', dia: '', hora: '' },
					instances: [{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei', level: 160 }],
					classNames: { 'guardiao-real': 'Guardião Real' },
					loadError: null,
					loginHref: '/auth/discord/login?next=%2Ftalentos',
					...over
				},
				params: {}
			} as never
		}).body;

	it('CA-01.1 / RN-12: o card mostra classe, nível, função, dias, faixa e instâncias', () => {
		const html = renderPage({});
		expect(html).toContain('Brasa');
		expect(html).toContain('Guardião Real · Nv 172');
		expect(html).toContain('Tank');
		expect(html).toContain('qua · 19:00–23:00 (Brasília)');
		expect(html).toContain('Templo do Demônio Rei');
	});

	it('CA-04.2 / RNF-05: visitante vê "Entre para ver o Discord", com o login que volta', () => {
		const html = renderPage({});
		expect(html).toMatch(
			/<a class="login[^"]*" href="\/auth\/discord\/login\?next=%2Ftalentos"[^>]*>Entre para ver o Discord/
		);
		expect(html).not.toContain('@');
		const logged = renderPage({
			user: USER,
			talents: [{ ...BRASA, discordUsername: 'brasa_ro' }]
		});
		expect(logged).toContain('@brasa_ro');
	});

	it('CA-04.3 / RN-11: sem ninguém, o aviso muda com e sem filtro; os filtros voltam marcados', () => {
		expect(renderPage({ talents: [] })).toContain('Ninguém no banco de talentos ainda.');
		const html = renderPage({
			talents: [],
			filters: { instancia: 'templo-do-demonio-rei', funcao: 'tank', dia: '6', hora: '01:00' }
		});
		expect(html).toContain('Ninguém no banco com esses filtros.');
		expect(html).toMatch(/<option value="tank"[^>]*selected/);
		expect(html).toMatch(/<option value="6"[^>]*selected/);
		expect(html).toMatch(/<option value="01:00"[^>]*selected/);
		expect(html).toContain('Limpar');
	});

	it('risco do design: com 100, mostra o aviso dos 100 primeiros', () => {
		expect(renderPage({ limited: true })).toContain(
			'Mostrando os 100 primeiros. Use os filtros para achar mais.'
		);
	});
});
