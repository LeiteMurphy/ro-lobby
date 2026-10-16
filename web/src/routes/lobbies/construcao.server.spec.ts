import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({
	env: { API_BASE_URL: 'http://api.test', DISCORD_UNDER_CONSTRUCTION: 'true' }
}));

const detalhe = await import('./[id]/+page.server');
const { SESSION_COOKIE } = await import('$lib/auth/oauth');

// Actions da sala no Discord em construção (spec discord-em-construcao, RN-06).

type Action = Parameters<(typeof detalhe.actions)['openRoom']>[0];

function event(calls: string[]) {
	const jar = new Map([[SESSION_COOKIE, 'tok']]);
	return {
		params: { id: 'l1' },
		cookies: { get: (n: string) => jar.get(n), delete: (n: string) => void jar.delete(n) },
		fetch: (async (input: string | URL) => {
			calls.push(String(input));
			return new Response(null, { status: 204 });
		}) as typeof fetch
	} as unknown as Action;
}

describe('sala no Discord em construção', () => {
	for (const name of ['openRoom', 'alert'] as const) {
		it(`CA-01.4: a action ${name} recusa sem chamar a API`, async () => {
			const calls: string[] = [];
			const result = await detalhe.actions[name](event(calls));
			expect(result).toMatchObject({
				status: 409,
				data: { action: 'discord', text: 'A sala no Discord está em construção.', ok: false }
			});
			expect(calls).toHaveLength(0);
		});
	}
});
