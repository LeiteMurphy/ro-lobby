import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test' } }));

const { actions } = await import('./+page.server');
const { SESSION_COOKIE } = await import('$lib/auth/oauth');

// Actions da sala no Discord no detalhe do lobby (spec discord-bot, T-10).

const LOBBY = '4b1c2d3e-5f60-4a7b-8c9d-0e1f2a3b4c5d';

function fakeCookies() {
	const jar = new Map([[SESSION_COOKIE, 'tok']]);
	return { jar, get: (n: string) => jar.get(n), delete: (n: string) => void jar.delete(n) };
}

function reply(status: number, body?: unknown) {
	const calls: string[] = [];
	const fn = (async (input: string | URL, init?: RequestInit) => {
		calls.push(`${init?.method} ${new URL(String(input)).pathname}`);
		return new Response(body === undefined ? null : JSON.stringify(body), {
			status,
			headers: { 'content-type': 'application/json' }
		});
	}) as typeof fetch;
	return { fn, calls };
}

const event = (fn: typeof fetch) =>
	({
		params: { id: LOBBY },
		cookies: fakeCookies(),
		fetch: fn,
		request: new Request('http://web')
	}) as never;

describe('openRoom', () => {
	it('CA-02.1: abre a sala pela API', async () => {
		const api = reply(201, {
			status: 'open',
			url: 'https://discord.com/channels/1/2',
			alertAvailableAt: null
		});
		expect(await actions.openRoom(event(api.fn))).toEqual({
			action: 'discord',
			text: 'Sala aberta no Discord.',
			ok: true
		});
		expect(api.calls).toEqual([`POST /lobbies/${LOBBY}/discord-room`]);
	});

	it.each([
		[
			503,
			{ error: 'discord_unavailable' },
			'Não foi possível abrir a sala agora. Tente de novo em instantes.'
		],
		[503, { error: 'guild_full' }, 'O servidor do RO Lobby está cheio agora. Tente mais tarde.'],
		[409, { error: 'discord_rule', code: 'room_exists' }, 'A sala desse lobby já foi aberta.']
	])('CA-02.7 / CA-02.8 / RN-20: %i %o vira a mensagem', async (status, body, text) => {
		const result = await actions.openRoom(event(reply(status, body).fn));
		expect(result).toMatchObject({ data: { action: 'discord', text, ok: false } });
	});
});

describe('alert', () => {
	it('CA-08.1: avisa pela API', async () => {
		const api = reply(204);
		expect(await actions.alert(event(api.fn))).toMatchObject({
			text: 'Grupo avisado no Discord.',
			ok: true
		});
		expect(api.calls).toEqual([`POST /lobbies/${LOBBY}/discord-alert`]);
	});

	it('CA-08.2 / RN-27: no intervalo, "Aguarde 6 min para avisar de novo"', async () => {
		const result = await actions.alert(
			event(reply(429, { error: 'cooldown', retryAfterSeconds: 360 }).fn)
		);
		expect(result).toMatchObject({
			status: 429,
			data: { text: 'Aguarde 6 min para avisar de novo', ok: false }
		});
	});

	it('CA-08.6 / RN-28: Discord fora do ar', async () => {
		const result = await actions.alert(event(reply(503, { error: 'discord_unavailable' }).fn));
		expect(result).toMatchObject({
			data: { text: 'Não foi possível avisar agora. Tente de novo em instantes.', ok: false }
		});
	});
});
