import { isRedirect } from '@sveltejs/kit';
import { render } from 'svelte/server';
import { describe, expect, it, vi } from 'vitest';
import type { MyApplication } from '$lib/applications/api';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test' } }));

const server = await import('./+page.server');
const Page = (await import('./+page.svelte')).default;
const { SESSION_COOKIE } = await import('$lib/auth/oauth');

// "Minhas candidaturas" (spec candidatura-lobby, T-06: RN-33, CA-03.1, CA-03.2, CA-10.4).

type Load = Parameters<typeof server.load>[0];
type Withdraw = Parameters<(typeof server.actions)['withdraw']>[0];

const BIA = { id: 'u-bia', username: 'bia', globalName: 'Bia' };
const app = (
	id: string,
	status: MyApplication['application']['status'],
	over: Partial<MyApplication> = {}
): MyApplication => ({
	application: {
		id,
		lobbyId: `l-${id}`,
		characterId: 'c1',
		role: 'tank',
		message: null,
		status,
		reason: null,
		blocked: false,
		createdAt: '2026-10-06T20:00:00Z',
		decidedAt: null
	},
	lobby: { instanceName: 'Torre da Constelação', startsAt: '2026-10-07T23:00:00Z', status: 'open' },
	character: { nick: 'Brasa', classId: 'guardiao-real', level: 250, portrait: 'retrato-3' },
	...over
});

function fakeCookies(token?: string) {
	const jar = new Map(token ? [[SESSION_COOKIE, token]] : []);
	return { jar, get: (n: string) => jar.get(n), delete: (n: string) => void jar.delete(n) };
}

const json = (status: number, body: unknown) =>
	new Response(JSON.stringify(body), { status, headers: { 'content-type': 'application/json' } });

function fakeFetch(responses: Record<string, Response>) {
	const calls: string[] = [];
	const fn = (async (input: string | URL, init?: RequestInit) => {
		const key = `${init?.method ?? 'GET'} ${new URL(String(input)).pathname}`;
		calls.push(key);
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
	return undefined;
}

describe('/candidaturas no servidor', () => {
	it('CA-03.1 / RN-33: lista as candidaturas da sessão com os nomes das classes', async () => {
		const { fn } = fakeFetch({
			'GET /me/applications': json(200, [app('a1', 'pending')]),
			'GET /classes': json(200, [{ id: 'guardiao-real', name: 'Guardião Real' }])
		});
		const data = (await server.load({
			locals: { user: BIA },
			cookies: fakeCookies('t'),
			fetch: fn
		} as unknown as Load)) as Record<string, unknown>;
		expect(data.applications).toEqual([app('a1', 'pending')]);
		expect(data.classNames).toEqual({ 'guardiao-real': 'Guardião Real' });
	});

	it('RN-04: sem sessão vai ao login e volta para /candidaturas', async () => {
		const { fn, calls } = fakeFetch({});
		const e = await thrown(() =>
			server.load({ locals: { user: null }, cookies: fakeCookies(), fetch: fn } as unknown as Load)
		);
		expect(isRedirect(e) && e.location).toBe('/auth/discord/login?next=%2Fcandidaturas');
		expect(calls).toEqual([]);
	});

	const withdraw = (f: typeof fetch) =>
		server.actions.withdraw({
			request: (() => {
				const d = new FormData();
				d.set('applicationId', 'a1');
				return new Request('http://web/candidaturas', { method: 'POST', body: d });
			})(),
			cookies: fakeCookies('t'),
			fetch: f
		} as unknown as Withdraw);

	it('CA-03.2: retirar chama a API; não pendente volta com o aviso', async () => {
		const ok = fakeFetch({ 'POST /applications/a1/withdraw': json(200, {}) });
		expect(await withdraw(ok.fn)).toEqual({ done: 'withdraw' });
		const late = fakeFetch({
			'POST /applications/a1/withdraw': json(409, {
				error: 'application_rule',
				code: 'not_pending'
			})
		});
		expect(await withdraw(late.fn)).toMatchObject({
			status: 409,
			data: { message: 'Essa candidatura não está mais pendente.' }
		});
	});
});

describe('/candidaturas renderizada', () => {
	const page = (applications: MyApplication[]) =>
		render(Page, {
			props: {
				data: { user: BIA, applications, classNames: { 'guardiao-real': 'Guardião Real' } },
				form: null,
				params: {}
			} as never
		}).body;

	it('CA-10.4: cada candidatura com lobby, personagem e estado; a recusada com justificativa', () => {
		const html = page([
			app('a1', 'pending'),
			app('a2', 'accepted'),
			app('a3', 'rejected', {
				application: { ...app('a3', 'rejected').application, reason: 'Já temos tank' }
			}),
			app('a4', 'expired', {
				lobby: { instanceName: 'Ilha Bios', startsAt: '2026-10-04T23:00:00Z', status: 'started' }
			})
		]);
		expect(html.match(/data-testid="my-application-row"/g)).toHaveLength(4);
		expect(html).toContain('Torre da Constelação');
		expect(html).toMatch(/qua, 7 out · 20:00 · com Brasa \(Tank\)/);
		expect(html).toContain('Guardião Real Nv 250');
		for (const label of ['Pendente', 'Aceita', 'Recusada', 'Expirada'])
			expect(html).toContain(label);
		expect(html).toContain('Justificativa:</b> Já temos tank');
		expect(html).toContain('href="/lobbies/l-a2"');
	});

	it('CA-03.2: só a pendente de lobby aberto tem "Retirar"', () => {
		const html = page([app('a1', 'pending'), app('a2', 'accepted')]);
		expect(html.match(/action="\?\/withdraw"/g)).toHaveLength(1);
		expect(html).toMatch(/value="a1"[\s\S]*?Retirar/);
	});

	it('RN-26: personagem excluído aparece como "Personagem excluído"', () => {
		const html = page([app('a1', 'expired', { character: null })]);
		expect(html).toContain('com Personagem excluído (Tank)');
	});

	it('vazio convida a escolher um grupo na Home', () => {
		expect(page([])).toContain('Você ainda não se candidatou a nenhum grupo.');
	});
});
