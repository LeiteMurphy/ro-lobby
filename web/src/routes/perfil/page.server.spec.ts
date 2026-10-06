import { isRedirect } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test' } }));

const { load, actions } = await import('./+page.server');
const { SESSION_COOKIE } = await import('$lib/auth/oauth');

// Testes do servidor da página de perfil (spec personagens, T-06), com a API falsa.

const USER = { id: 'u1', username: 'grimbold', globalName: 'Grimbold' };
const LOGIN = '/auth/discord/login?next=%2Fperfil';

function fakeCookies(token?: string) {
	const jar = new Map(token ? [[SESSION_COOKIE, token]] : []);
	return {
		jar,
		get: (name: string) => jar.get(name),
		delete: (name: string) => void jar.delete(name)
	};
}

const json = (status: number, body?: unknown) =>
	body === undefined
		? new Response(null, { status })
		: new Response(JSON.stringify(body), {
				status,
				headers: { 'content-type': 'application/json' }
			});

function fakeFetch(responses: Record<string, Response>) {
	const calls: { key: string; body?: unknown; auth?: string | null }[] = [];
	const fn = (async (input: string | URL, init?: RequestInit) => {
		const key = `${init?.method ?? 'GET'} ${new URL(String(input)).pathname}`;
		calls.push({
			key,
			body: init?.body ? JSON.parse(String(init.body)) : undefined,
			auth: new Headers(init?.headers).get('authorization')
		});
		const r = responses[key];
		if (!r) throw new TypeError(`sem resposta para ${key}`);
		return r.clone();
	}) as typeof fetch;
	return { fn, calls };
}

async function redirectOf(run: () => unknown): Promise<string> {
	try {
		await run();
	} catch (e) {
		if (isRedirect(e)) return e.location;
		throw e;
	}
	throw new Error('não redirecionou');
}

function form(fields: Record<string, string>) {
	const data = new FormData();
	for (const [k, v] of Object.entries(fields)) data.set(k, v);
	return new Request('http://web/perfil', { method: 'POST', body: data });
}

const BRASA_FORM = {
	nick: 'Brasa',
	classId: 'guardiao-real',
	level: '172',
	role: 'tank',
	portrait: '',
	link: ''
};

type LoadEvent = Parameters<typeof load>[0];
type ActionEvent = Parameters<(typeof actions)['create']>[0];

const loadEvent = (
	user: typeof USER | null,
	cookies: ReturnType<typeof fakeCookies>,
	f = fakeFetch({}).fn
) => ({ locals: { user }, cookies, fetch: f }) as unknown as LoadEvent;
const actionEvent = (request: Request, cookies: ReturnType<typeof fakeCookies>, f: typeof fetch) =>
	({ request, cookies, fetch: f }) as unknown as ActionEvent;

describe('load de /perfil', () => {
	it('CA-06.1 / RN-03: visitante vai ao login e volta para /perfil', async () => {
		expect(await redirectOf(() => load(loadEvent(null, fakeCookies())))).toBe(LOGIN);
	});

	it('borda de RN-01 / RN-03: sessão recusada pela API apaga o cookie e manda ao login', async () => {
		const cookies = fakeCookies('tok');
		const { fn } = fakeFetch({ 'GET /characters': json(401), 'GET /classes': json(200, []) });
		expect(await redirectOf(() => load(loadEvent(USER, cookies, fn)))).toBe(LOGIN);
		expect(cookies.jar.has(SESSION_COOKIE)).toBe(false);
	});

	it('CA-01.1 / RN-01: busca os personagens com o token da sessão e o catálogo sem token', async () => {
		const { fn, calls } = fakeFetch({
			'GET /characters': json(200, [{ id: 'c1', nick: 'Brasa' }]),
			'GET /classes': json(200, [{ id: 'paladino', name: 'Paladino' }])
		});
		const data = (await load(loadEvent(USER, fakeCookies('tok'), fn))) as Record<string, unknown>;
		expect(data.characters).toEqual([{ id: 'c1', nick: 'Brasa' }]);
		expect(data.classes).toEqual([{ id: 'paladino', name: 'Paladino' }]);
		expect(data.loadError).toBeNull();
		expect(calls.find((c) => c.key === 'GET /characters')?.auth).toBe('Bearer tok');
		expect(calls.find((c) => c.key === 'GET /classes')?.auth).toBeNull();
	});

	it('RN-21: API fora do ar mostra o aviso, sem quebrar a página', async () => {
		const { fn } = fakeFetch({ 'GET /characters': json(500), 'GET /classes': json(500) });
		const data = (await load(loadEvent(USER, fakeCookies('tok'), fn))) as Record<string, unknown>;
		expect(data.characters).toEqual([]);
		expect(data.loadError).toBe('Não foi possível falar com o servidor. Tente de novo.');
	});
});

describe('actions de /perfil', () => {
	it('CA-02.1: criar manda os campos para a API', async () => {
		const { fn, calls } = fakeFetch({ 'POST /characters': json(201, { id: 'c1' }) });
		const result = await actions.create(
			actionEvent(form({ ...BRASA_FORM, link: 'https://exemplo.com' }), fakeCookies('tok'), fn)
		);
		expect(result).toEqual({ done: 'create' });
		expect(calls[0].body).toEqual({
			nick: 'Brasa',
			classId: 'guardiao-real',
			level: 172,
			role: 'tank',
			link: 'https://exemplo.com'
		});
		expect(calls[0].auth).toBe('Bearer tok');
	});

	it('CA-02.2: o retrato escolhido vai junto; sem escolha, fica para a API', async () => {
		const { fn, calls } = fakeFetch({ 'POST /characters': json(201, { id: 'c1' }) });
		await actions.create(
			actionEvent(form({ ...BRASA_FORM, portrait: 'retrato-3' }), fakeCookies('t'), fn)
		);
		await actions.create(actionEvent(form(BRASA_FORM), fakeCookies('t'), fn));
		expect((calls[0].body as Record<string, unknown>).portrait).toBe('retrato-3');
		expect(calls[1].body).not.toHaveProperty('portrait');
	});

	it('RNF-04: nível vazio ou não inteiro vai como 0, e a API devolve o erro', async () => {
		const { fn, calls } = fakeFetch({ 'POST /characters': json(201, { id: 'c1' }) });
		for (const level of ['', 'abc', '1.5']) {
			await actions.create(actionEvent(form({ ...BRASA_FORM, level }), fakeCookies('t'), fn));
		}
		expect(calls.map((c) => (c.body as Record<string, unknown>).level)).toEqual([0, 0, 0]);
	});

	it('CA-02.9 / RN-19: erro de campo volta com a mensagem e os valores digitados', async () => {
		const { fn } = fakeFetch({
			'POST /characters': json(422, {
				error: 'validation',
				fields: [{ field: 'nick', code: 'taken' }]
			})
		});
		const values = { ...BRASA_FORM, link: 'https://exemplo.com' };
		const result = await actions.create(actionEvent(form(values), fakeCookies('t'), fn));
		expect(result).toMatchObject({
			status: 422,
			data: { mode: 'create', values, errors: { nick: 'Esse nick já está em uso' } }
		});
	});

	it('CA-02.8: o limite volta com "Você já tem 10 personagens"', async () => {
		const { fn } = fakeFetch({ 'POST /characters': json(409, { error: 'character_limit' }) });
		const result = await actions.create(actionEvent(form(BRASA_FORM), fakeCookies('t'), fn));
		expect(result).toMatchObject({ status: 409, data: { message: 'Você já tem 10 personagens' } });
	});

	it('CA-03.1 / CA-03.3: editar usa o id do formulário; personagem de outro é 404', async () => {
		const { fn, calls } = fakeFetch({ 'PUT /characters/c9': json(404, { error: 'not_found' }) });
		const result = await actions.update(
			actionEvent(form({ ...BRASA_FORM, id: 'c9' }), fakeCookies('t'), fn)
		);
		expect(calls[0].key).toBe('PUT /characters/c9');
		expect(result).toMatchObject({ status: 404, data: { mode: 'update', id: 'c9' } });
	});

	it('CA-04.1 / CA-05.1: excluir e marcar principal chamam as rotas certas', async () => {
		const { fn, calls } = fakeFetch({
			'DELETE /characters/c1': json(204),
			'PUT /characters/c1/main': json(204)
		});
		expect(await actions.delete(actionEvent(form({ id: 'c1' }), fakeCookies('t'), fn))).toEqual({
			done: 'delete'
		});
		expect(await actions.main(actionEvent(form({ id: 'c1' }), fakeCookies('t'), fn))).toEqual({
			done: 'main'
		});
		expect(calls.map((c) => c.key)).toEqual(['DELETE /characters/c1', 'PUT /characters/c1/main']);
	});

	it('borda de RN-01 / RN-03: sessão vencida no meio da edição leva ao login e apaga o cookie', async () => {
		const cookies = fakeCookies('t');
		const { fn } = fakeFetch({ 'PUT /characters/c1': json(401) });
		const to = await redirectOf(() =>
			actions.update(actionEvent(form({ ...BRASA_FORM, id: 'c1' }), cookies, fn))
		);
		expect(to).toBe(LOGIN);
		expect(cookies.jar.has(SESSION_COOKIE)).toBe(false);
	});

	it('RN-01: sem cookie, nenhuma action chama a API', async () => {
		const { fn, calls } = fakeFetch({});
		for (const name of ['create', 'update', 'delete', 'main'] as const) {
			expect(
				await redirectOf(() => actions[name](actionEvent(form(BRASA_FORM), fakeCookies(), fn)))
			).toBe(LOGIN);
		}
		expect(calls).toEqual([]);
	});
});
