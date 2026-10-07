import { isHttpError, isRedirect } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';
import { TEMPLE } from '$lib/lobbies/fixtures';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test' } }));

const novo = await import('./novo/+page.server');
const editar = await import('./[id]/editar/+page.server');
const { SESSION_COOKIE } = await import('$lib/auth/oauth');

// Servidor das páginas de criar e editar lobby (spec lobbies, T-07), com a API falsa.

const ANA = { id: TEMPLE.owner.userId, username: 'ana', globalName: 'Ana' };
const BIA = { id: '11111111-2222-4333-8444-555555555555', username: 'bia', globalName: null };
const LIRIEN = {
	id: TEMPLE.owner.characterId,
	nick: 'Lirien',
	classId: 'arcebispo',
	level: 178,
	role: 'support',
	portrait: 'retrato-1',
	link: null,
	isMain: true,
	createdAt: '2026-10-01T00:00:00Z'
};
const BRASA = { ...LIRIEN, id: 'c2', nick: 'Brasa', role: 'tank', isMain: false };
const INSTANCES = [
	{ id: 'torre-da-constelacao', name: 'Torre da Constelação', level: 240, reset: 'three_days' },
	{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei', level: 160, reset: 'daily' }
];

function fakeCookies(token?: string) {
	const jar = new Map(token ? [[SESSION_COOKIE, token]] : []);
	return { jar, get: (n: string) => jar.get(n), delete: (n: string) => void jar.delete(n) };
}

const json = (status: number, body?: unknown) =>
	body === undefined
		? new Response(null, { status })
		: new Response(JSON.stringify(body), {
				status,
				headers: { 'content-type': 'application/json' }
			});

function fakeFetch(responses: Record<string, Response>) {
	const calls: { key: string; body?: unknown }[] = [];
	const fn = (async (input: string | URL, init?: RequestInit) => {
		const url = new URL(String(input));
		const key = `${init?.method ?? 'GET'} ${url.pathname}`;
		calls.push({ key, body: init?.body ? JSON.parse(String(init.body)) : undefined });
		const r = responses[key];
		if (!r) throw new TypeError(`sem resposta para ${key}`);
		return r.clone();
	}) as typeof fetch;
	return { fn, calls };
}

async function thrown(run: () => unknown): Promise<unknown> {
	try {
		await run();
	} catch (e) {
		return e;
	}
	throw new Error('não lançou');
}

function form(fields: Record<string, string>) {
	const data = new FormData();
	for (const [k, v] of Object.entries(fields)) data.set(k, v);
	return new Request('http://web/lobbies/novo', { method: 'POST', body: data });
}

const FORM = {
	instanceId: 'templo-do-demonio-rei',
	date: '2026-10-07',
	time: '20:00',
	tank: '1',
	support: '2',
	dps: '3',
	minLevel: '160',
	characterId: LIRIEN.id as string,
	note: ''
};

type NovoLoad = Parameters<typeof novo.load>[0];
type NovoAction = Parameters<(typeof novo.actions)['create']>[0];
type EditLoad = Parameters<typeof editar.load>[0];
type EditAction = Parameters<(typeof editar.actions)['update']>[0];

const novoLoad = (
	user: unknown,
	cookies: ReturnType<typeof fakeCookies>,
	f: typeof fetch,
	search = ''
) =>
	({
		locals: { user },
		cookies,
		fetch: f,
		url: new URL(`http://web/lobbies/novo${search}`)
	}) as unknown as NovoLoad;
const editLoad = (user: unknown, cookies: ReturnType<typeof fakeCookies>, f: typeof fetch) =>
	({ params: { id: TEMPLE.id }, locals: { user }, cookies, fetch: f }) as unknown as EditLoad;

describe('/lobbies/novo', () => {
	it('CA-02.3 / RN-04: visitante vai ao login e volta para a criação', async () => {
		const e = await thrown(() => novo.load(novoLoad(null, fakeCookies(), fakeFetch({}).fn)));
		expect(isRedirect(e) && e.location).toBe('/auth/discord/login?next=%2Flobbies%2Fnovo');
	});

	it('CA-01.1 / RN-06 / RN-07 / RN-08: padrões 1/2/3, a instância mais alta que o principal alcança e o principal', async () => {
		const { fn } = fakeFetch({
			'GET /characters': json(200, [BRASA, LIRIEN]),
			'GET /instances': json(200, INSTANCES),
			'GET /classes': json(200, [])
		});
		const data = (await novo.load(novoLoad(ANA, fakeCookies('t'), fn))) as Record<string, unknown>;
		expect(data.values).toMatchObject({
			instanceId: 'templo-do-demonio-rei',
			tank: '1',
			support: '2',
			dps: '3',
			minLevel: '160',
			characterId: LIRIEN.id
		});
		expect((data.days as unknown[]).length).toBe(14);
	});

	it('CA-02.5 / RN-24: o dia escolhido na Home chega preenchido; dia fora dos 14 cai no padrão', async () => {
		const responses = () =>
			fakeFetch({
				'GET /characters': json(200, [LIRIEN]),
				'GET /instances': json(200, INSTANCES),
				'GET /classes': json(200, [])
			}).fn;
		const first = (await novo.load(novoLoad(ANA, fakeCookies('t'), responses()))) as {
			days: { date: string }[];
			values: { date: string; time: string };
		};
		const chosen = first.days[5].date;
		const data = (await novo.load(
			novoLoad(ANA, fakeCookies('t'), responses(), `?dia=${chosen}`)
		)) as typeof first;
		expect(data.values).toMatchObject({ date: chosen, time: '20:00' });
		const outside = (await novo.load(
			novoLoad(ANA, fakeCookies('t'), responses(), '?dia=2000-01-01')
		)) as typeof first;
		expect(outside.values.date).toBe(first.values.date);
	});

	it('CA-02.3 / RN-24: sessão recusada pela API também volta do login com o dia', async () => {
		const { fn } = fakeFetch({
			'GET /characters': json(401, { error: 'no_session' }),
			'GET /instances': json(200, INSTANCES),
			'GET /classes': json(200, [])
		});
		const e = await thrown(() =>
			novo.load(novoLoad(ANA, fakeCookies('vencido'), fn, '?dia=2026-10-09'))
		);
		expect(isRedirect(e) && e.location).toBe(
			`/auth/discord/login?next=${encodeURIComponent('/lobbies/novo?dia=2026-10-09')}`
		);
	});

	it('CA-02.3 / RN-24: o visitante volta do login com o dia', async () => {
		const e = await thrown(() =>
			novo.load(novoLoad(null, fakeCookies(), fakeFetch({}).fn, '?dia=2026-10-09'))
		);
		expect(isRedirect(e) && e.location).toBe(
			`/auth/discord/login?next=${encodeURIComponent('/lobbies/novo?dia=2026-10-09')}`
		);
	});

	it('CA-01.1: criar manda o horário de Brasília em UTC e vai para o detalhe', async () => {
		const { fn, calls } = fakeFetch({ 'POST /lobbies': json(201, TEMPLE) });
		const e = await thrown(() =>
			novo.actions.create({
				request: form({ ...FORM, note: 'Chamar no Discord' }),
				cookies: fakeCookies('t'),
				fetch: fn
			} as unknown as NovoAction)
		);
		expect(isRedirect(e) && e.location).toBe(`/lobbies/${TEMPLE.id}`);
		expect(calls[0].body).toEqual({
			instanceId: 'templo-do-demonio-rei',
			characterId: LIRIEN.id,
			startsAt: '2026-10-07T23:00:00Z',
			slots: { tank: 1, support: 2, dps: 3 },
			minLevel: 160,
			note: 'Chamar no Discord'
		});
	});

	it('CA-01.8 / D-07: erro de campo volta com a mensagem e os valores', async () => {
		const { fn } = fakeFetch({
			'POST /lobbies': json(422, {
				error: 'validation',
				fields: [{ field: 'startsAt', code: 'conflict' }]
			})
		});
		const result = await novo.actions.create({
			request: form(FORM),
			cookies: fakeCookies('t'),
			fetch: fn
		} as unknown as NovoAction);
		expect(result).toMatchObject({
			status: 422,
			data: {
				values: FORM,
				errors: { startsAt: 'Esse personagem já está num grupo nesse horário' }
			}
		});
	});

	it('CA-01.9: o limite volta com "Você já tem 5 lobbies abertos"', async () => {
		const { fn } = fakeFetch({ 'POST /lobbies': json(409, { error: 'lobby_limit' }) });
		const result = await novo.actions.create({
			request: form(FORM),
			cookies: fakeCookies('t'),
			fetch: fn
		} as unknown as NovoAction);
		expect(result).toMatchObject({
			status: 409,
			data: { message: 'Você já tem 5 lobbies abertos' }
		});
	});

	it('RNF-04: hora malformada nem chega na API', async () => {
		const { fn, calls } = fakeFetch({});
		const result = await novo.actions.create({
			request: form({ ...FORM, time: '25:00' }),
			cookies: fakeCookies('t'),
			fetch: fn
		} as unknown as NovoAction);
		expect(result).toMatchObject({
			status: 422,
			data: { errors: { startsAt: 'Escolha um horário no futuro, em até 14 dias' } }
		});
		expect(calls).toEqual([]);
	});
});

describe('/lobbies/[id]/editar', () => {
	const lobbyFetch = (lobby: unknown) =>
		fakeFetch({
			'GET /lobbies/4b1c2d3e-5f60-4a7b-8c9d-0e1f2a3b4c5d': json(200, lobby),
			'GET /characters': json(200, [LIRIEN, BRASA]),
			'GET /classes': json(200, [])
		}).fn;

	it('CA-04.1 / RN-17: o dono vê o formulário com os valores do lobby e só o personagem do dono', async () => {
		const data = (await editar.load(editLoad(ANA, fakeCookies('t'), lobbyFetch(TEMPLE)))) as Record<
			string,
			unknown
		>;
		expect(data.values).toMatchObject({
			date: '2026-10-07',
			time: '20:00',
			tank: '1',
			support: '2',
			dps: '3',
			minLevel: '160'
		});
		expect((data.characters as { id: string }[]).map((c) => c.id)).toEqual([LIRIEN.id]);
	});

	it('RN-20: outro Usuário vê não encontrado', async () => {
		const e = await thrown(() => editar.load(editLoad(BIA, fakeCookies('t'), lobbyFetch(TEMPLE))));
		expect(isHttpError(e) && e.status).toBe(404);
	});

	it('RN-17: lobby iniciado ou cancelado volta para o detalhe', async () => {
		const e = await thrown(() =>
			editar.load(editLoad(ANA, fakeCookies('t'), lobbyFetch({ ...TEMPLE, status: 'started' })))
		);
		expect(isRedirect(e) && e.location).toBe(`/lobbies/${TEMPLE.id}`);
	});

	it('CA-04.2 / D-07: vagas abaixo dos ocupantes voltam com a mensagem', async () => {
		const { fn, calls } = fakeFetch({
			[`PUT /lobbies/${TEMPLE.id}`]: json(422, {
				error: 'validation',
				fields: [{ field: 'slots', code: 'below_occupied' }]
			})
		});
		const result = await editar.actions.update({
			params: { id: TEMPLE.id },
			request: form({ ...FORM, support: '0' }),
			cookies: fakeCookies('t'),
			fetch: fn
		} as unknown as EditAction);
		expect(result).toMatchObject({
			status: 422,
			data: { errors: { slots: 'Essa função já tem ocupante' } }
		});
		expect(calls[0].body).toEqual({
			startsAt: '2026-10-07T23:00:00Z',
			slots: { tank: 1, support: 0, dps: 3 },
			minLevel: 160
		});
	});

	it('CA-04.4 / D-08: lobby que começou durante a edição mostra o aviso', async () => {
		const { fn } = fakeFetch({
			[`PUT /lobbies/${TEMPLE.id}`]: json(409, { error: 'lobby_not_open' })
		});
		const result = await editar.actions.update({
			params: { id: TEMPLE.id },
			request: form(FORM),
			cookies: fakeCookies('t'),
			fetch: fn
		} as unknown as EditAction);
		expect(result).toMatchObject({
			status: 409,
			data: { message: 'Esse lobby já começou ou foi cancelado.' }
		});
	});
});

const detalhe = await import('./[id]/+page.server');
type DetailLoad = Parameters<typeof detalhe.load>[0];
type CancelAction = Parameters<(typeof detalhe.actions)['cancel']>[0];

describe('/lobbies/[id]', () => {
	const detailFetch = (lobby: Response) =>
		fakeFetch({
			[`GET /lobbies/${TEMPLE.id}`]: lobby,
			'GET /classes': json(200, [{ id: 'arcebispo', name: 'Arcebispo' }])
		}).fn;
	const detailLoad = (user: unknown, f: typeof fetch) =>
		({ params: { id: TEMPLE.id }, locals: { user }, fetch: f }) as unknown as DetailLoad;

	it('CA-03.1 / RN-15: público, com a classe do dono pelo nome', async () => {
		const data = (await detalhe.load(detailLoad(null, detailFetch(json(200, TEMPLE))))) as Record<
			string,
			unknown
		>;
		expect(data.lobby).toEqual(TEMPLE);
		expect(data.ownerClass).toBe('Arcebispo');
		expect(data.isOwner).toBe(false);
	});

	it('CA-03.3 / RN-16: o dono é reconhecido pela sessão', async () => {
		const data = (await detalhe.load(detailLoad(ANA, detailFetch(json(200, TEMPLE))))) as Record<
			string,
			unknown
		>;
		expect(data.isOwner).toBe(true);
		const other = (await detalhe.load(detailLoad(BIA, detailFetch(json(200, TEMPLE))))) as Record<
			string,
			unknown
		>;
		expect(other.isOwner).toBe(false);
	});

	it('CA-03.2: lobby inexistente é 404', async () => {
		const e = await thrown(() =>
			detalhe.load(detailLoad(null, detailFetch(json(404, { error: 'not_found' }))))
		);
		expect(isHttpError(e) && e.status).toBe(404);
	});

	const cancel = (reason: string, f: typeof fetch, cookies = fakeCookies('t')) =>
		detalhe.actions.cancel({
			params: { id: TEMPLE.id },
			request: (() => {
				const d = new FormData();
				d.set('reason', reason);
				return new Request(`http://web/lobbies/${TEMPLE.id}`, { method: 'POST', body: d });
			})(),
			cookies,
			fetch: f
		} as unknown as CancelAction);

	it('CA-05.1 / RN-19: cancelar manda o motivo', async () => {
		const { fn, calls } = fakeFetch({
			[`POST /lobbies/${TEMPLE.id}/cancel`]: json(200, { ...TEMPLE, status: 'cancelled' })
		});
		expect(await cancel('Metade do grupo não pode', fn)).toEqual({ done: 'cancel' });
		expect(calls[0].body).toEqual({ reason: 'Metade do grupo não pode' });
	});

	it('CA-05.2: motivo curto volta com a mensagem e o texto digitado', async () => {
		const { fn } = fakeFetch({
			[`POST /lobbies/${TEMPLE.id}/cancel`]: json(422, {
				error: 'validation',
				fields: [{ field: 'reason', code: 'too_short' }]
			})
		});
		expect(await cancel('não dá', fn)).toMatchObject({
			status: 422,
			data: { reason: 'não dá', errors: { reason: 'Escreva de 10 a 250 caracteres' } }
		});
	});

	it('RN-04: sem sessão, cancelar leva ao login com volta para o lobby', async () => {
		const { fn, calls } = fakeFetch({});
		const e = await thrown(() => cancel('Metade do grupo não pode', fn, fakeCookies()));
		expect(isRedirect(e) && e.location).toBe(`/auth/discord/login?next=%2Flobbies%2F${TEMPLE.id}`);
		expect(calls).toEqual([]);
	});
});
