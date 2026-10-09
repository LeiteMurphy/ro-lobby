import { roomFor } from './lobbies';
import { toMinutes } from './time';
import { ROLES, type Lobby, type Role } from './types';

export type TimeRangeKey = 'any' | '18-20' | '20-22' | '22-24';

export interface TimeRange {
	key: TimeRangeKey;
	label: string;
	/** Início incluído, em minutos. */
	start: number;
	/** Fim excluído, em minutos. */
	end: number;
}

/** RN-12: faixas de horário, com o início incluído e o fim excluído. */
export const TIME_RANGES: readonly TimeRange[] = [
	{ key: 'any', label: 'Qualquer horário', start: 0, end: 1440 },
	{ key: '18-20', label: '18h–20h', start: 1080, end: 1200 },
	{ key: '20-22', label: '20h–22h', start: 1200, end: 1320 },
	{ key: '22-24', label: '22h–00h', start: 1320, end: 1440 }
];

export const LEVEL_OPTIONS = [
	{ value: '', label: 'Qualquer nível' },
	{ value: '150', label: 'Até Nv 150' },
	{ value: '160', label: 'Até Nv 160' },
	{ value: '170', label: 'Até Nv 170' },
	{ value: '180', label: 'Até Nv 180' }
] as const;

export interface Filters {
	/** Instância escolhida, ou '' para todas. */
	instance: string;
	/** Funções marcadas em "Vaga para". */
	roles: Role[];
	/** Nível mínimo escolhido ("Até Nv N"), ou null para qualquer nível. */
	maxMinLevel: number | null;
	timeRange: TimeRangeKey;
}

export const NO_FILTERS: Filters = { instance: '', roles: [], maxMinLevel: null, timeRange: 'any' };

function inRange(lobby: Lobby, key: TimeRangeKey): boolean {
	const range = TIME_RANGES.find((r) => r.key === key) ?? TIME_RANGES[0];
	const start = toMinutes(lobby.time);
	return start >= range.start && start < range.end;
}

/** RN-12: todos os filtros combinados com E; entre as funções marcadas, OU. */
export function matches(
	lobby: Lobby,
	filters: Filters,
	options: { skipTime?: boolean } = {}
): boolean {
	return (
		(!filters.instance || lobby.instance === filters.instance) &&
		(filters.roles.length === 0 || filters.roles.some((role) => roomFor(lobby, role) > 0)) &&
		(filters.maxMinLevel === null || lobby.minLevel <= filters.maxMinLevel) &&
		(options.skipTime === true || inRange(lobby, filters.timeRange))
	);
}

export function applyFilters(lobbies: readonly Lobby[], filters: Filters): Lobby[] {
	return lobbies.filter((l) => matches(l, filters));
}

/** RN-13: quantos lobbies do dia têm vaga em cada função, sem considerar os filtros. */
export function roleCounts(dayLobbies: readonly Lobby[]): Record<Role, number> {
	return Object.fromEntries(
		ROLES.map((role) => [role, dayLobbies.filter((l) => roomFor(l, role) > 0).length])
	) as Record<Role, number>;
}

/** RN-13: quantos lobbies passam nos outros filtros e começam em cada faixa. */
export function timeRangeCounts(
	dayLobbies: readonly Lobby[],
	filters: Filters
): Record<TimeRangeKey, number> {
	const passing = dayLobbies.filter((l) => matches(l, filters, { skipTime: true }));
	return Object.fromEntries(
		TIME_RANGES.map((r) => [r.key, passing.filter((l) => inRange(l, r.key)).length])
	) as Record<TimeRangeKey, number>;
}

/** RN-19: quantidade de filtros ativos, mostrada em "Filtros (N)". */
export function activeFilterCount(filters: Filters): number {
	return (
		(filters.instance ? 1 : 0) +
		filters.roles.length +
		(filters.maxMinLevel !== null ? 1 : 0) +
		(filters.timeRange !== 'any' ? 1 : 0)
	);
}

/** Instâncias que aparecem no filtro, em ordem alfabética. */
export function instanceOptions(lobbies: readonly Lobby[]): string[] {
	return [...new Set(lobbies.map((l) => l.instance))].sort((a, b) => a.localeCompare(b, 'pt-BR'));
}

/** RN-17: qual estado vazio mostrar, se algum. */
export function emptyState(
	dayLobbies: readonly Lobby[],
	filtered: readonly Lobby[]
): 'day' | 'filters' | null {
	if (dayLobbies.length === 0) return 'day';
	if (filtered.length === 0) return 'filters';
	return null;
}

/** CA-04.8: "5 grupos", "1 grupo" ou "2 de 5 grupos" quando há filtro ativo. */
export function countLabel(total: number, filtered: number, filters: Filters): string {
	const noun = total === 1 ? 'grupo' : 'grupos';
	return activeFilterCount(filters) === 0 ? `${total} ${noun}` : `${filtered} de ${total} ${noun}`;
}
