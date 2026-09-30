import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import type { ApiStatus } from '$lib/api/health';
import Page from './+page.svelte';

function renderStatus(apiStatus: ApiStatus): string {
	// O PageProps completo inclui `params`, que esta página não usa.
	const props = { data: { apiStatus }, params: {} } as unknown as Parameters<typeof Page>[1];
	return render(Page, { props }).body;
}

// CA-05.4: o HTML renderizado no servidor já traz o texto, sem depender de JavaScript
// no navegador.
describe('página de status (SSR)', () => {
	it('CA-05.1 / CA-05.4: mostra "API online" no HTML do servidor', () => {
		expect(renderStatus('online')).toContain('API online');
	});

	it('CA-05.2: mostra "API com problema"', () => {
		expect(renderStatus('degraded')).toContain('API com problema');
	});

	it('CA-05.3: mostra "API indisponível" sem erro nem stack trace', () => {
		const html = renderStatus('unavailable');
		expect(html).toContain('API indisponível');
		expect(html).not.toMatch(/error|stack|at .+:\d+/i);
	});
});
