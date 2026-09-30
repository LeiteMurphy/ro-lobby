import { describe, expect, it } from 'vitest';
import { buildDays, DAYS_AHEAD } from './days';
import { DAY_COUNTS, getHomeLobbies } from './fixtures';
import {
	activeFilterCount,
	applyFilters,
	countLabel,
	emptyState,
	NO_FILTERS,
	roleCounts,
	timeRangeCounts,
	type Filters
} from './filters';
import { edgeRole, featuredLobby, headcount, isFull, lobbiesForDay, openSlots } from './lobbies';
import { addDays, relativeLabel, zonedNow, type ZonedNow } from './time';
import type { Composition, Lobby } from './types';

const TODAY = '2026-09-30';
const at = (hhmm: string, date = TODAY): ZonedNow => {
	const [h, m] = hhmm.split(':').map(Number);
	return { date, minutes: h * 60 + m };
};
const comp = (t: [number, number], s: [number, number], d: [number, number]): Composition => ({
	tank: { filled: t[0], total: t[1] },
	support: { filled: s[0], total: s[1] },
	dps: { filled: d[0], total: d[1] }
});
let seq = 0;
const lobby = (overrides: Partial<Lobby> = {}): Lobby => ({
	id: `l${seq++}`,
	date: TODAY,
	time: '20:00',
	instance: 'Torre sem fim',
	host: 'Kaizen',
	hostClass: 'Paladino',
	minLevel: 160,
	composition: comp([0, 1], [1, 3], [3, 6]),
	...overrides
});
const filters = (f: Partial<Filters>): Filters => ({ ...NO_FILTERS, ...f });

describe('RN-09: fuso America/Sao_Paulo', () => {
	it('CA-02.4: 19:40 UTC é 16:40 em São Paulo', () => {
		expect(zonedNow(new Date('2026-09-30T19:40:00Z'))).toEqual({
			date: '2026-09-30',
			minutes: 16 * 60 + 40
		});
	});

	it('RN-09: 01:30 UTC ainda é o dia anterior em São Paulo', () => {
		expect(zonedNow(new Date('2026-10-01T01:30:00Z')).date).toBe('2026-09-30');
	});

	it('addDays atravessa o fim do mês', () => {
		expect(addDays('2026-09-30', 1)).toBe('2026-10-01');
	});
});

describe('US-02 — grupos do dia', () => {
	it('CA-02.1: dados fictícios de hoje, em ordem de horário', () => {
		const all = getHomeLobbies(TODAY);
		const today = lobbiesForDay(all, TODAY);
		expect(today.length).toBe(DAY_COUNTS[0]);
		expect(today.map((l) => l.time)).toEqual([...today.map((l) => l.time)].sort());
		for (const l of today) {
			expect(l.host && l.hostClass && l.instance && l.minLevel).toBeTruthy();
		}
	});

	it('CA-02.1 / RN-11: ordena por horário mesmo fora de ordem na origem', () => {
		const list = [lobby({ time: '22:45' }), lobby({ time: '18:00' }), lobby({ time: '19:30' })];
		expect(lobbiesForDay(list, TODAY).map((l) => l.time)).toEqual(['18:00', '19:30', '22:45']);
	});

	it('CA-02.2: lobby com todas as vagas ocupadas está lotado e não tem borda de função', () => {
		const full = comp([1, 1], [3, 3], [8, 8]);
		expect(isFull(full)).toBe(true);
		expect(headcount(full)).toEqual({ filled: 12, total: 12 });
		expect(edgeRole(full)).toBeNull();
	});

	it('CA-02.3: borda na cor de Suporte com 0 de Tank, 2 de Suporte e 2 de Dano abertas', () => {
		expect(edgeRole(comp([1, 1], [1, 3], [4, 6]))).toBe('support');
	});

	it('RN-14: no empate vale Tank, Suporte, Dano', () => {
		expect(edgeRole(comp([0, 2], [0, 2], [0, 2]))).toBe('tank');
		expect(edgeRole(comp([1, 1], [0, 2], [0, 2]))).toBe('support');
	});

	it('RN-14: vagas abertas = total − ocupadas', () => {
		expect(openSlots(comp([0, 1], [1, 3], [5, 5]), 'support')).toBe(2);
		expect(openSlots(comp([0, 1], [1, 3], [5, 5]), 'dps')).toBe(0);
	});

	it('CA-02.4: às 16:40, lobby de hoje às 18:00 mostra "em 1 h 20 min"', () => {
		expect(relativeLabel(TODAY, '18:00', at('16:40'))).toBe('em 1 h 20 min');
	});

	it('CA-02.4: lobby de hoje que já passou e lobby de amanhã não mostram tempo relativo', () => {
		expect(relativeLabel(TODAY, '16:00', at('16:40'))).toBe('');
		expect(relativeLabel(addDays(TODAY, 1), '18:00', at('16:40'))).toBe('');
	});

	it('RN-15: formatos "em N min" e "em H h"', () => {
		expect(relativeLabel(TODAY, '17:05', at('16:40'))).toBe('em 25 min');
		expect(relativeLabel(TODAY, '18:40', at('16:40'))).toBe('em 2 h');
	});
});

describe('US-03 — troca de dia', () => {
	it('CA-03.1: 14 dias a partir de hoje, com hoje primeiro e a contagem de cada dia', () => {
		const days = buildDays(TODAY, getHomeLobbies(TODAY));
		expect(days).toHaveLength(DAYS_AHEAD);
		expect(days[0]).toMatchObject({
			date: TODAY,
			today: true,
			weekday: 'Qua',
			day: '30',
			label: 'qua, 30 set'
		});
		expect(days.filter((d) => d.today)).toHaveLength(1);
		expect(days.map((d) => d.count)).toEqual([...DAY_COUNTS]);
		expect(days[1]).toMatchObject({ date: '2026-10-01', day: '01' });
	});

	it('CA-03.3: dia sem grupos mostra o estado "Nenhum grupo neste dia"', () => {
		const empty = lobbiesForDay(getHomeLobbies(TODAY), addDays(TODAY, 2));
		expect(empty).toHaveLength(0);
		expect(emptyState(empty, empty)).toBe('day');
	});
});

describe('US-04 — filtros', () => {
	const day = [
		lobby({
			instance: 'Caverna de gelo',
			time: '18:00',
			minLevel: 150,
			composition: comp([1, 1], [2, 3], [3, 6])
		}),
		lobby({
			instance: 'Torre sem fim',
			time: '19:30',
			minLevel: 160,
			composition: comp([0, 1], [3, 3], [5, 5])
		}),
		lobby({
			instance: 'Torre sem fim',
			time: '20:00',
			minLevel: 185,
			composition: comp([1, 1], [3, 3], [2, 5])
		}),
		lobby({
			instance: 'Templo submerso',
			time: '22:45',
			minLevel: 170,
			composition: comp([1, 1], [1, 2], [8, 8])
		})
	];
	const times = (list: Lobby[]) => list.map((l) => l.time);

	it('CA-04.1: instância', () => {
		expect(times(applyFilters(day, filters({ instance: 'Torre sem fim' })))).toEqual([
			'19:30',
			'20:00'
		]);
	});

	it('CA-04.2: vaga para Tank ou Suporte (OU entre funções)', () => {
		// 18:00 tem Suporte; 19:30 tem Tank; 20:00 só tem Dano; 22:45 tem Suporte.
		expect(times(applyFilters(day, filters({ roles: ['tank', 'support'] })))).toEqual([
			'18:00',
			'19:30',
			'22:45'
		]);
	});

	it('CA-04.3: "Até Nv 160" mostra nível mínimo 150 e 160', () => {
		expect(applyFilters(day, filters({ maxMinLevel: 160 })).map((l) => l.minLevel)).toEqual([
			150, 160
		]);
	});

	it('CA-04.4: "20h–22h" pega 20:00 e deixa 19:30 e 22:45 de fora', () => {
		expect(times(applyFilters(day, filters({ timeRange: '20-22' })))).toEqual(['20:00']);
	});

	it('RN-12: o fim da faixa é excluído', () => {
		expect(
			times(applyFilters([lobby({ time: '22:00' })], filters({ timeRange: '20-22' })))
		).toEqual([]);
		expect(
			times(applyFilters([lobby({ time: '22:00' })], filters({ timeRange: '22-24' })))
		).toEqual(['22:00']);
	});

	it('CA-04.5: filtros combinados com E', () => {
		expect(
			times(applyFilters(day, filters({ instance: 'Torre sem fim', timeRange: '18-20' })))
		).toEqual(['19:30']);
	});

	it('CA-04.6: "Vaga para" ignora os filtros; faixas respeitam os outros filtros', () => {
		expect(roleCounts(day)).toEqual({ tank: 1, support: 2, dps: 2 });
		expect(timeRangeCounts(day, filters({ instance: 'Torre sem fim' }))).toEqual({
			any: 2,
			'18-20': 1,
			'20-22': 1,
			'22-24': 0
		});
	});

	it('CA-04.6: a contagem das faixas não depende da faixa escolhida', () => {
		expect(timeRangeCounts(day, filters({ timeRange: '22-24' }))['18-20']).toBe(2);
	});

	it('CA-04.7: dia com grupos e nenhum resultado mostra "Nenhum grupo com esses filtros"', () => {
		const f = filters({ instance: 'Caverna de gelo', timeRange: '22-24' });
		const filtered = applyFilters(day, f);
		expect(emptyState(day, filtered)).toBe('filters');
		expect(applyFilters(day, NO_FILTERS)).toHaveLength(day.length);
	});

	it('CA-04.8: subtítulo "2 de 5 grupos" com filtro, "5 grupos" sem', () => {
		expect(countLabel(5, 2, filters({ instance: 'x' }))).toBe('2 de 5 grupos');
		expect(countLabel(5, 5, NO_FILTERS)).toBe('5 grupos');
		expect(countLabel(1, 1, NO_FILTERS)).toBe('1 grupo');
	});

	it('CA-06.2 / RN-19: conta os filtros ativos', () => {
		expect(activeFilterCount(NO_FILTERS)).toBe(0);
		expect(
			activeFilterCount(
				filters({ instance: 'x', roles: ['tank', 'dps'], maxMinLevel: 160, timeRange: '18-20' })
			)
		).toBe(5);
	});
});

describe('US-05 — destaque', () => {
	it('CA-05.1: pula o lotado das 18:00 e destaca o das 19:00', () => {
		const list = [
			lobby({ time: '18:00', composition: comp([1, 1], [3, 3], [8, 8]) }),
			lobby({ time: '19:00', composition: comp([0, 1], [1, 3], [5, 5]) })
		];
		expect(featuredLobby(list, at('16:40'))?.time).toBe('19:00');
	});

	it('CA-05.2: às 19:10, ignora o que começou às 19:00 e destaca o das 21:30', () => {
		const list = [lobby({ time: '19:00' }), lobby({ time: '21:30' })];
		expect(featuredLobby(list, at('19:10'))?.time).toBe('21:30');
	});

	it('CA-05.2: em outro dia, o horário atual não importa', () => {
		const tomorrow = addDays(TODAY, 1);
		expect(featuredLobby([lobby({ date: tomorrow, time: '08:00' })], at('19:10'))?.time).toBe(
			'08:00'
		);
	});

	it('CA-05.3: todos lotados, sem destaque', () => {
		const full = comp([1, 1], [3, 3], [8, 8]);
		expect(
			featuredLobby([lobby({ composition: full }), lobby({ composition: full })], at('10:00'))
		).toBeNull();
	});
});
