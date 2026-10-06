import { isRedirect } from '@sveltejs/kit';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({
	env: {
		API_BASE_URL: 'http://api.test',
		DISCORD_CLIENT_ID: '123',
		DISCORD_AUTHORIZE_URL: 'http://discord.test/oauth2/authorize'
	}
}));

const { GET: login } = await import('./discord/login/+server');
const { GET: callback } = await import('./discord/callback/+server');
const { POST: logout } = await import('./logout/+server');
const { handle } = await import('../../hooks.server');
const { SESSION_COOKIE, STATE_COOKIE, encodeStateCookie } = await import('$lib/auth/oauth');

type CookieOpts = Record<string, unknown>;

/** Cookies em memória, com os atributos de cada um. */
function fakeCookies(initial: Record<string, string> = {}) {
	const jar = new Map(Object.entries(initial));
	const options = new Map<string, CookieOpts>();
	return {
		jar,
		options,
		get: (name: string) => jar.get(name),
		set: (name: string, value: string, opts: CookieOpts) => {
			jar.set(name, value);
			options.set(name, opts);
		},
		delete: (name: string) => {
			jar.delete(name);
		}
	};
}

/** fetch falso que registra as chamadas à API e responde conforme a rota. */
function fakeFetch(responses: Record<string, Response | (() => Promise<Response>)>) {
	const calls: { url: string; init?: RequestInit }[] = [];
	const fn = (async (input: string | URL | Request, init?: RequestInit) => {
		const url = String(input);
		calls.push({ url, init });
		const key = `${init?.method ?? 'GET'} ${new URL(url).pathname}`;
		const r = responses[key];
		if (!r) throw new TypeError(`sem resposta para ${key}`);
		return typeof r === 'function' ? r() : r.clone();
	}) as typeof fetch;
	return { fn, calls };
}

const json = (status: number, body: unknown) =>
	new Response(JSON.stringify(body), { status, headers: { 'content-type': 'application/json' } });

const USER = {
	id: '6f1c2b8e-3a4d-4e5f-9a1b-2c3d4e5f6a7b',
	username: 'grimbold',
	globalName: 'Grimbold'
};

/** Executa o handler e devolve o destino do redirect. */
async function location(run: () => unknown): Promise<string> {
	try {
		await run();
	} catch (e) {
		if (isRedirect(e)) return e.location;
		throw e;
	}
	throw new Error('o handler não redirecionou');
}

function event(
	url: string,
	cookies: ReturnType<typeof fakeCookies>,
	fetchFn: typeof fetch = fakeFetch({}).fn
) {
	// Só o que os handlers usam.
	return { url: new URL(url), cookies, fetch: fetchFn } as never;
}

describe('GET /auth/discord/login', () => {
	it('CA-01.3 / RN-02: grava o cookie de state e redireciona com scope=identify e o mesmo state', async () => {
		const cookies = fakeCookies();
		const to = new URL(
			await location(() =>
				login(event('http://localhost:3000/auth/discord/login?next=/status', cookies))
			)
		);
		expect(to.origin + to.pathname).toBe('http://discord.test/oauth2/authorize');
		expect(to.searchParams.get('scope')).toBe('identify');
		expect(to.searchParams.get('redirect_uri')).toBe('http://localhost:3000/auth/discord/callback');

		const saved = JSON.parse(Buffer.from(cookies.jar.get(STATE_COOKIE)!, 'base64url').toString());
		expect(saved).toEqual({ state: to.searchParams.get('state'), next: '/status' });
		expect(cookies.options.get(STATE_COOKIE)).toMatchObject({
			httpOnly: true,
			sameSite: 'lax',
			path: '/',
			maxAge: 600
		});
	});
});

describe('GET /auth/discord/callback', () => {
	let api: ReturnType<typeof fakeFetch>;
	beforeEach(() => {
		api = fakeFetch({
			'POST /auth/discord': json(201, { sessionToken: 'token-novo', user: USER })
		});
	});
	const withState = (next = '/') =>
		fakeCookies({ [STATE_COOKIE]: encodeStateCookie('estado-certo', next) });

	it('CA-01.1 / CA-06.5: código aceito grava o cookie de sessão e volta para o destino', async () => {
		const cookies = withState('/status');
		const to = await location(() =>
			callback(
				event(
					'http://localhost:3000/auth/discord/callback?code=abc&state=estado-certo',
					cookies,
					api.fn
				)
			)
		);
		expect(to).toBe('/status');
		expect(cookies.jar.get(SESSION_COOKIE)).toBe('token-novo');
		expect(cookies.options.get(SESSION_COOKIE)).toEqual({
			httpOnly: true,
			sameSite: 'lax',
			path: '/',
			secure: false,
			maxAge: 30 * 24 * 60 * 60
		});
		expect(JSON.parse(String(api.calls[0].init?.body))).toEqual({
			code: 'abc',
			redirectUri: 'http://localhost:3000/auth/discord/callback'
		});
	});

	it('CA-04.1: cancelou no Discord → Home com erro, sem chamar a API', async () => {
		const cookies = withState();
		const to = await location(() =>
			callback(
				event(
					'http://localhost:3000/auth/discord/callback?error=access_denied&state=estado-certo',
					cookies,
					api.fn
				)
			)
		);
		expect(to).toBe('/?login=erro');
		expect(api.calls).toHaveLength(0);
		expect(cookies.jar.has(SESSION_COOKIE)).toBe(false);
		expect(cookies.jar.has(STATE_COOKIE)).toBe(false);
	});

	it('CA-04.2: state diferente → erro, sem chamar a API', async () => {
		const cookies = withState();
		const to = await location(() =>
			callback(
				event('http://localhost:3000/auth/discord/callback?code=abc&state=outro', cookies, api.fn)
			)
		);
		expect(to).toBe('/?login=erro');
		expect(api.calls).toHaveLength(0);
	});

	it('CA-04.3: sem cookie de state (vencido) → erro, sem chamar a API', async () => {
		const to = await location(() =>
			callback(
				event(
					'http://localhost:3000/auth/discord/callback?code=abc&state=estado-certo',
					fakeCookies(),
					api.fn
				)
			)
		);
		expect(to).toBe('/?login=erro');
		expect(api.calls).toHaveLength(0);
	});

	it('CA-04.6: o mesmo retorno repetido é recusado (state de uso único)', async () => {
		const cookies = withState();
		const url = 'http://localhost:3000/auth/discord/callback?code=abc&state=estado-certo';
		expect(await location(() => callback(event(url, cookies, api.fn)))).toBe('/');
		cookies.delete(SESSION_COOKIE);
		expect(await location(() => callback(event(url, cookies, api.fn)))).toBe('/?login=erro');
		expect(api.calls).toHaveLength(1);
	});

	it('CA-04.4 / CA-04.5: a API recusa ou não responde → erro, sem cookie de sessão', async () => {
		for (const response of [
			json(400, { error: 'invalid_code' }),
			json(502, { error: 'discord_unavailable' })
		]) {
			const cookies = withState();
			const failing = fakeFetch({ 'POST /auth/discord': response });
			const to = await location(() =>
				callback(
					event(
						'http://localhost:3000/auth/discord/callback?code=abc&state=estado-certo',
						cookies,
						failing.fn
					)
				)
			);
			expect(to).toBe('/?login=erro');
			expect(cookies.jar.has(SESSION_COOKIE)).toBe(false);
		}
	});
});

describe('POST /auth/logout', () => {
	it('CA-03.1: encerra a sessão na API, apaga o cookie e volta para a Home', async () => {
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'token-valido' });
		const api = fakeFetch({ 'DELETE /session': new Response(null, { status: 204 }) });
		const to = await location(() =>
			logout(event('http://localhost:3000/auth/logout', cookies, api.fn))
		);
		expect(to).toBe('/');
		expect(cookies.jar.has(SESSION_COOKIE)).toBe(false);
		expect(new Headers(api.calls[0].init?.headers).get('authorization')).toBe(
			'Bearer token-valido'
		);
	});
});

describe('hook de sessão', () => {
	const resolve = vi.fn(async () => new Response('ok'));

	it('CA-06.1: sessão válida põe o usuário em locals', async () => {
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'token-valido' });
		const api = fakeFetch({ 'GET /me': json(200, USER) });
		const ev = {
			url: new URL('http://localhost:3000/'),
			cookies,
			fetch: api.fn,
			locals: {} as App.Locals
		};
		await handle({ event: ev as never, resolve });
		expect(ev.locals.user).toEqual(USER);
		expect(cookies.jar.get(SESSION_COOKIE)).toBe('token-valido');
	});

	it('CA-06.2 / RN-10: sessão recusada (401) vira visitante e apaga o cookie', async () => {
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'token-vencido' });
		const api = fakeFetch({ 'GET /me': json(401, { error: 'no_session' }) });
		const ev = {
			url: new URL('http://localhost:3000/'),
			cookies,
			fetch: api.fn,
			locals: {} as App.Locals
		};
		await handle({ event: ev as never, resolve });
		expect(ev.locals.user).toBeNull();
		expect(cookies.jar.has(SESSION_COOKIE)).toBe(false);
	});

	it('RN-10: API fora do ar vira visitante, mas mantém o cookie', async () => {
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'token-valido' });
		const ev = {
			url: new URL('http://localhost:3000/'),
			cookies,
			fetch: fakeFetch({}).fn,
			locals: {} as App.Locals
		};
		await handle({ event: ev as never, resolve });
		expect(ev.locals.user).toBeNull();
		expect(cookies.jar.get(SESSION_COOKIE)).toBe('token-valido');
	});

	it('RN-10: sem cookie, é visitante e não chama a API', async () => {
		const api = fakeFetch({});
		const ev = {
			url: new URL('http://localhost:3000/'),
			cookies: fakeCookies(),
			fetch: api.fn,
			locals: {} as App.Locals
		};
		await handle({ event: ev as never, resolve });
		expect(ev.locals.user).toBeNull();
		expect(api.calls).toHaveLength(0);
	});
});
