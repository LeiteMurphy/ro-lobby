import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import FilterPanel from './components/FilterPanel.svelte';
import LobbyCard from './components/LobbyCard.svelte';
import {
	activeFilterCount,
	applyFilters,
	byServer,
	emptyState,
	NO_FILTERS,
	roleCounts,
	serverCounts,
	timeRangeCounts
} from './filters';
import { buildDays } from './days';
import { featuredLobby } from './lobbies';
import type { Lobby } from './types';

// Filtro "Servidor" da Home (spec servidor, T-07).

const lobby = (id: string, server: 'freya' | 'nidhogg', tank = 0): Lobby => ({
	id,
	date: '2026-10-07',
	time: '20:00',
	instance: 'Templo do Demônio Rei',
	host: 'Lirien',
	hostClass: 'Arcebispo',
	minLevel: 160,
	server,
	composition: {
		tank: { filled: 0, total: tank },
		support: { filled: 1, total: 2 },
		dps: { filled: 0, total: 3 }
	}
});

// CA-04.1: 3 lobbies hoje em Freya (um com vaga de Tank) e 2 em Nidhogg.
const DAY = [
	lobby('f1', 'freya', 1),
	lobby('f2', 'freya'),
	lobby('f3', 'freya'),
	lobby('n1', 'nidhogg'),
	lobby('n2', 'nidhogg')
];
const NOW = { date: '2026-10-06', minutes: 600 };

describe('filtro de servidor da Home', () => {
	it('CA-04.1 / RN-12 / RN-13: começa em Todos, com Todos 5, Freya 3 e Nidhogg 2', () => {
		expect(NO_FILTERS.server).toBe('');
		expect(serverCounts(DAY)).toEqual({ '': 5, freya: 3, nidhogg: 2 });
		expect(applyFilters(DAY, NO_FILTERS)).toHaveLength(5);
		const html = render(FilterPanel, {
			props: {
				filters: NO_FILTERS,
				instances: [],
				roleCounts: roleCounts(DAY),
				timeCounts: timeRangeCounts(DAY, NO_FILTERS),
				serverCounts: serverCounts(DAY),
				name: 'faixa',
				onchange: () => {}
			}
		}).body;
		const panel = html.split('data-testid="server-filter"')[1].split('</section>')[0];
		expect(panel).toMatch(/Todos[\s\S]*5[\s\S]*Freya[\s\S]*3[\s\S]*Nidhogg[\s\S]*2/);
		expect(panel).toMatch(/value="todos"[^>]*checked/);
	});

	it('CA-04.2 / RN-12: com Nidhogg, só os 2 de Nidhogg', () => {
		const filters = { ...NO_FILTERS, server: 'nidhogg' as const };
		expect(applyFilters(DAY, filters).map((l) => l.id)).toEqual(['n1', 'n2']);
		expect(activeFilterCount(filters)).toBe(1);
	});

	it('CA-04.3 / RN-14: Nidhogg com "Vaga para: Tank" fica vazio, e o destaque não é de Freya', () => {
		const filters = { ...NO_FILTERS, server: 'nidhogg' as const, roles: ['tank' as const] };
		const filtered = applyFilters(DAY, filters);
		expect(filtered).toEqual([]);
		expect(emptyState(DAY, filtered)).toBe('filters');
		const scoped = byServer(DAY, filters.server);
		expect(featuredLobby(scoped, NOW)?.server).toBe('nidhogg');
		expect(roleCounts(scoped).tank).toBe(0);
		expect(featuredLobby(byServer(DAY, ''), NOW)?.id).toBe('f1');
	});

	it('RN-14: o seletor de dias conta só os lobbies do servidor escolhido', () => {
		const week = [...DAY, { ...lobby('n3', 'nidhogg'), date: '2026-10-08' }];
		const count = (server: '' | 'freya' | 'nidhogg') =>
			buildDays('2026-10-07', byServer(week, server))
				.slice(0, 2)
				.map((d) => d.count);
		expect(count('')).toEqual([5, 1]);
		expect(count('freya')).toEqual([3, 0]);
		expect(count('nidhogg')).toEqual([2, 1]);
	});

	it('borda da RN-14: servidor sem lobby num dia com lobby é o estado vazio de filtros', () => {
		const onlyFreya = DAY.filter((l) => l.server === 'freya');
		const filtered = applyFilters(onlyFreya, { ...NO_FILTERS, server: 'nidhogg' });
		expect(emptyState(onlyFreya, filtered)).toBe('filters');
	});

	it('CA-04.4 / CA-04.5: "Limpar filtros" e o recarregar voltam para Todos', () => {
		expect(NO_FILTERS).toMatchObject({ server: '' });
		expect(activeFilterCount(NO_FILTERS)).toBe(0);
	});

	it('CA-02.4 / RN-08: o card mostra o servidor', () => {
		const html = render(LobbyCard, { props: { lobby: DAY[3], relative: '' } }).body;
		expect(html).toMatch(/data-testid="lobby-card-server"[^>]*>Nidhogg</);
	});
});
