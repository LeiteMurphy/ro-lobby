import { isRedirect } from '@sveltejs/kit';
import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({
	env: {
		API_BASE_URL: 'http://api.test',
		DISCORD_CLIENT_ID: '123',
		DISCORD_AUTHORIZE_URL: 'http://discord.test/oauth2/authorize',
		DISCORD_UNDER_CONSTRUCTION: 'true'
	}
}));

const { GET: start } = await import('./+server');
const { GET: back } = await import('./retorno/+server');
const { CONNECT_STATE_COOKIE, SESSION_COOKIE } = await import('$lib/auth/oauth');
const { connectMessage } = await import('$lib/discord/messages');

// Conexão ao servidor fechada em construção (spec discord-em-construcao, RN-05).

function fakeCookies(initial: Record<string, string> = {}) {
	const jar = new Map(Object.entries(initial));
	return {
		jar,
		get: (name: string) => jar.get(name),
		set: (name: string, value: string) => void jar.set(name, value),
		delete: (name: string) => void jar.delete(name)
	};
}

async function location(run: () => unknown): Promise<string> {
	try {
		await run();
	} catch (e) {
		if (isRedirect(e)) return e.location;
		throw e;
	}
	throw new Error('não redirecionou');
}

describe('conexão ao servidor em construção', () => {
	it('CA-01.3: /perfil/discord volta ao perfil sem ir ao Discord', async () => {
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'tok' });
		const to = await location(() =>
			start({
				url: new URL('https://rolobby.com.br/perfil/discord'),
				cookies,
				locals: { user: { id: 'u1' } }
			} as never)
		);
		expect(to).toBe('/perfil?discord=construcao');
		expect(cookies.jar.has(CONNECT_STATE_COOKIE)).toBe(false);
	});

	it('CA-01.3: o retorno volta ao perfil sem chamar a API', async () => {
		const calls: string[] = [];
		const fetchFn = (async (input: string | URL) => {
			calls.push(String(input));
			return new Response(null, { status: 200 });
		}) as typeof fetch;
		const cookies = fakeCookies({ [SESSION_COOKIE]: 'tok', [CONNECT_STATE_COOKIE]: 'estado' });
		const to = await location(() =>
			back({
				url: new URL('https://rolobby.com.br/perfil/discord/retorno?code=c1&state=estado'),
				cookies,
				fetch: fetchFn
			} as never)
		);
		expect(to).toBe('/perfil?discord=construcao');
		expect(calls).toHaveLength(0);
		expect(cookies.jar.has(CONNECT_STATE_COOKIE)).toBe(false);
	});

	it('RN-05: o perfil mostra a mensagem de construção', () => {
		expect(connectMessage('construcao')?.text).toBe(
			'O servidor do RO Lobby no Discord está em construção.'
		);
	});
});
