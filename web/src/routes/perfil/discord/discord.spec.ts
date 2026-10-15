import { isRedirect } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({
	env: {
		API_BASE_URL: 'http://api.test',
		DISCORD_CLIENT_ID: '123',
		DISCORD_AUTHORIZE_URL: 'http://discord.test/oauth2/authorize'
	}
}));

const { GET: start } = await import('./+server');
const { GET: back } = await import('./retorno/+server');
const { authorizeUrl, CONNECT_STATE_COOKIE, SESSION_COOKIE } = await import('$lib/auth/oauth');

// Conexão ao servidor do RO Lobby no web (spec discord-bot, T-09).

function fakeCookies(initial: Record<string, string> = {}) {
	const jar = new Map(Object.entries(initial));
	return {
		jar,
		get: (name: string) => jar.get(name),
		set: (name: string, value: string) => void jar.set(name, value),
		delete: (name: string) => void jar.delete(name)
	};
}

function fakeFetch(status: number, body: unknown) {
	const calls: { url: string; init?: RequestInit }[] = [];
	const fn = (async (input: string | URL | Request, init?: RequestInit) => {
		calls.push({ url: String(input), init });
		return new Response(body === null ? null : JSON.stringify(body), {
			status,
			headers: { 'content-type': 'application/json' }
		});
	}) as typeof fetch;
	return { fn, calls };
}

async function location(run: () => unknown): Promise<string> {
	try {
		await run();
	} catch (e) {
		if (isRedirect(e)) return e.location;
		throw e;
	}
	throw new Error('o handler não redirecionou');
}

const USER = { id: 'u1', username: 'ana', globalName: null };

describe('URL de autorização', () => {
	it('CA-01.2: o login continua só com identify', () => {
		const url = new URL(
			authorizeUrl({ authorizeUrl: 'http://d.test/a', clientId: '1', redirectUri: 'x', state: 's' })
		);
		expect(url.searchParams.get('scope')).toBe('identify');
	});
});

describe('/perfil/discord', () => {
	it('CA-01.3 / RN-01: pede identify e guilds.join com um state novo, guardado no cookie', async () => {
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'tok' });
		const to = new URL(
			await location(() =>
				start({
					url: new URL('http://web.test/perfil/discord'),
					cookies,
					locals: { user: USER }
				} as never)
			)
		);
		expect(to.origin + to.pathname).toBe('http://discord.test/oauth2/authorize');
		expect(to.searchParams.get('scope')).toBe('identify guilds.join');
		expect(to.searchParams.get('redirect_uri')).toBe('http://web.test/perfil/discord/retorno');
		expect(to.searchParams.get('state')).toBe(cookies.jar.get(CONNECT_STATE_COOKIE));
		expect(to.searchParams.get('state')?.length).toBeGreaterThanOrEqual(22);
	});

	it('sem sessão, vai ao login e volta para o perfil', async () => {
		const to = await location(() =>
			start({
				url: new URL('http://web.test/perfil/discord'),
				cookies: fakeCookies(),
				locals: {}
			} as never)
		);
		expect(to).toContain('/auth/discord/login?next=%2Fperfil');
	});
});

describe('/perfil/discord/retorno', () => {
	const run = (query: string, fetchFn: typeof fetch, state = 'estado-123') => {
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'tok', [CONNECT_STATE_COOKIE]: state });
		return {
			cookies,
			to: location(() =>
				back({
					url: new URL(`http://web.test/perfil/discord/retorno?${query}`),
					cookies,
					fetch: fetchFn
				} as never)
			)
		};
	};

	it('CA-01.1: com o state certo, chama a API e volta com "conectado"', async () => {
		const api = fakeFetch(200, { enabled: true, connectedAt: '2026-10-14T12:00:00Z' });
		const { cookies, to } = run('code=c1&state=estado-123', api.fn);
		expect(await to).toBe('/perfil?discord=conectado');
		expect(cookies.jar.has(CONNECT_STATE_COOKIE)).toBe(false);
		expect(api.calls[0].url).toBe('http://api.test/me/discord-connection');
		expect(JSON.parse(String(api.calls[0].init?.body))).toEqual({
			code: 'c1',
			redirectUri: 'http://web.test/perfil/discord/retorno'
		});
		expect((api.calls[0].init?.headers as Record<string, string>).authorization).toBe('Bearer tok');
	});

	it('CA-01.4 / RN-03: outra conta volta com "conta"', async () => {
		const api = fakeFetch(409, { error: 'discord_rule', code: 'wrong_account' });
		expect(await run('code=c1&state=estado-123', api.fn).to).toBe('/perfil?discord=conta');
	});

	it('CA-01.5: autorização cancelada volta com "cancelado", sem chamar a API', async () => {
		const api = fakeFetch(200, {});
		expect(await run('error=access_denied&state=estado-123', api.fn).to).toBe(
			'/perfil?discord=cancelado'
		);
		expect(api.calls).toHaveLength(0);
	});

	it('state diferente volta com erro, sem chamar a API', async () => {
		const api = fakeFetch(200, {});
		expect(await run('code=c1&state=outro', api.fn).to).toBe('/perfil?discord=erro');
		expect(api.calls).toHaveLength(0);
	});

	it('CA-01.7 / RN-22: com a sala desligada, volta com "desligado"', async () => {
		const api = fakeFetch(404, { error: 'discord_disabled' });
		expect(await run('code=c1&state=estado-123', api.fn).to).toBe('/perfil?discord=desligado');
	});

	it('Discord fora do ar volta com "erro"', async () => {
		const api = fakeFetch(503, { error: 'discord_unavailable' });
		expect(await run('code=c1&state=estado-123', api.fn).to).toBe('/perfil?discord=erro');
	});
});
