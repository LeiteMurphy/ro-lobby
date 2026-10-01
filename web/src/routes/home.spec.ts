import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { DAY_COUNTS, getHomeLobbies } from '$lib/home/fixtures';
import { lobbiesForDay } from '$lib/home/lobbies';
import Page from './+page.svelte';

// 2026-09-30 19:40 UTC = 16:40 em São Paulo, como no design.
const NOW = '2026-09-30T19:40:00.000Z';
const TODAY = '2026-09-30';

function renderHome(): string {
	const data = { now: NOW, today: TODAY, lobbies: getHomeLobbies(TODAY) };
	// O PageProps completo inclui `params`, que a Home não usa.
	const props = { data, params: {} } as unknown as Parameters<typeof Page>[1];
	return render(Page, { props }).body;
}

describe('Home renderizada no servidor', () => {
	const html = renderHome();
	const today = lobbiesForDay(getHomeLobbies(TODAY), TODAY);

	it('CA-02.6 / RN-20: o HTML do servidor já traz os cards de hoje', () => {
		const cards = html.match(/data-testid="lobby-card"/g) ?? [];
		expect(cards).toHaveLength(DAY_COUNTS[0]);
		for (const l of today) expect(html).toContain(`${l.instance} às ${l.time}`);
	});

	it('CA-02.1: título, subtítulo e dados de cada card', () => {
		expect(html).toContain('Grupos para hoje');
		expect(html).toContain(`qua, 30 set · ${DAY_COUNTS[0]} grupos · por horário`);
		expect(html).toContain('Lyrae');
		expect(html).toContain('Arcebispo');
		expect(html).toContain('Nv 160+');
	});

	it('CA-02.2: o lobby lotado mostra o selo "Lotado"', () => {
		expect(html).toMatch(/data-edge="full"[\s\S]*?Lotado/);
	});

	it('CA-02.4: tempo relativo no card das 18:00', () => {
		expect(html).toContain('em 1 h 20 min');
	});

	it('CA-02.5 / RN-18: ações sem backend ficam desabilitadas com "Disponível em breve"', () => {
		const buttons = [...html.matchAll(/<button(\s[^>]*)?>([\s\S]*?)<\/button>/g)].map(
			([, attrs, inner]) => ({
				disabled: (attrs ?? '').includes('aria-disabled="true"'),
				text: inner.replace(/<[^>]+>/g, '').trim()
			})
		);
		for (const label of ['Criar lobby', 'Entrar com Discord', 'Candidatar', 'Ver grupo']) {
			const matching = buttons.filter((b) => b.text === label);
			expect(matching.length, label).toBeGreaterThan(0);
			expect(
				matching.every((b) => b.disabled),
				label
			).toBe(true);
		}
		expect(html).toContain('Disponível em breve');
	});

	it('CA-05.1: o destaque aparece com o próximo grupo com vaga', () => {
		expect(html).toContain('data-testid="featured-lobby"');
		expect(html).toContain('Próximo grupo com vaga');
	});

	it('CA-03.1: o seletor mostra 14 dias, com hoje selecionado', () => {
		expect(html.match(/role="tab"/g)).toHaveLength(14);
		expect(html).toMatch(/aria-selected="true"[^>]*data-date="2026-09-30"/);
	});

	it('RN-22: sem marca da Gravity e sem nomes oficiais de mapa ou monstro na tela', () => {
		expect(html).not.toMatch(/Gravity|Prontera|Glast Heim|Poring/i);
	});
});
