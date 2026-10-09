import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import FilterPanel from './components/FilterPanel.svelte';
import { applyFilters, NO_FILTERS, timeRangeCounts, visibleTimeRanges } from './filters';
import type { Lobby } from './types';

// Faixas de horário dinâmicas na Home (spec home-local, revisão de 2026-10-09).

const lobby = (time: string): Lobby => ({
	id: time,
	date: '2026-10-07',
	time,
	instance: 'Templo do Demônio Rei',
	host: 'Lirien',
	hostClass: 'Arcebispo',
	minLevel: 160,
	composition: {
		tank: { filled: 0, total: 1 },
		support: { filled: 1, total: 2 },
		dps: { filled: 0, total: 3 }
	}
});
const labels = (day: Lobby[], selected = 'any') =>
	visibleTimeRanges(day, selected).map((r) => r.label);

describe('faixas de horário dinâmicas', () => {
	it('CA-04.8 / RN-12: um lobby às 08:30 faz aparecer "08h–10h", na ordem do relógio', () => {
		const day = [lobby('08:30'), lobby('20:00')];
		expect(labels(day)).toEqual(['Qualquer horário', '08h–10h', '18h–20h', '20h–22h', '22h–00h']);
		expect(timeRangeCounts(day, NO_FILTERS)['08-10']).toBe(1);
		expect(applyFilters(day, { ...NO_FILTERS, timeRange: '08-10' }).map((l) => l.time)).toEqual([
			'08:30'
		]);
	});

	it('CA-04.8 / RN-12: o início entra e o fim não; a madrugada vira 00h–02h', () => {
		expect(labels([lobby('10:00'), lobby('01:59')])).toEqual([
			'Qualquer horário',
			'00h–02h',
			'10h–12h',
			'18h–20h',
			'20h–22h',
			'22h–00h'
		]);
	});

	it('CA-04.9 / RN-12: sem lobby antes das 18h, só as fixas; a marcada continua visível', () => {
		expect(labels([lobby('19:00')])).toEqual(['Qualquer horário', '18h–20h', '20h–22h', '22h–00h']);
		expect(labels([], '08-10')).toContain('08h–10h');
	});

	it('CA-04.8 / RN-13: o painel mostra as faixas do dia com a contagem', () => {
		const day = [lobby('08:30')];
		const html = render(FilterPanel, {
			props: {
				filters: NO_FILTERS,
				instances: [],
				roleCounts: { tank: 0, support: 0, dps: 0 },
				timeCounts: timeRangeCounts(day, NO_FILTERS),
				ranges: visibleTimeRanges(day, 'any'),
				name: 'faixa',
				onchange: () => {}
			}
		}).body;
		expect(html).toMatch(/08h–10h[\s\S]*18h–20h/);
		expect(html).not.toContain('10h–12h');
	});
});
