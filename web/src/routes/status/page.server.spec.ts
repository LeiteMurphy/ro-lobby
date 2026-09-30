import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({ env: { API_BASE_URL: 'http://api.test:9999' } }));

const { load } = await import('./+page.server');

type LoadEvent = Parameters<typeof load>[0];

function eventWith(fetchFn: typeof fetch): LoadEvent {
	return { fetch: fetchFn } as unknown as LoadEvent;
}

describe('load da página de status', () => {
	it('CA-05.1: consulta o /healthz na API_BASE_URL e devolve online', async () => {
		let called = '';
		const fetchFn: typeof fetch = async (input) => {
			called = String(input);
			return new Response('{"status":"ok","database":"ok"}', { status: 200 });
		};
		await expect(load(eventWith(fetchFn))).resolves.toEqual({ apiStatus: 'online' });
		expect(called).toBe('http://api.test:9999/healthz');
	});

	it('CA-05.3: backend parado não derruba a página', async () => {
		const fetchFn: typeof fetch = async () => {
			throw new TypeError('fetch failed');
		};
		await expect(load(eventWith(fetchFn))).resolves.toEqual({ apiStatus: 'unavailable' });
	});
});
