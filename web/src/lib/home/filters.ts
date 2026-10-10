import { NO_INSTANCE } from '$lib/lobbies/form';
import { roomFor } from './lobbies';
import { toMinutes } from './time';
import { ROLES, type Lobby, type Role } from './types';

/** 'any' ou um bloco de 2 h, como '08-10' e '22-24' (RN-12). */
export type TimeRangeKey = string;

export interface TimeRange {
	key: TimeRangeKey;
	label: string;
	/** Início incluído, em minutos. */
	start: number;
	/** Fim excluído, em minutos. */
	end: number;
}

const hh = (h: number) => String(h % 24).padStart(2, '0');

/** RN-12: "Qualquer horário". */
export const ANY_TIME: TimeRange = { key: 'any', label: 'Qualquer horário', start: 0, end: 1440 };

/**
 * RN-12 (revisão de 2026-10-09): o dia em blocos fixos de 2 h, com o início incluído e o
 * fim excluído, de 00h–02h a 22h–00h.
 */
export const TIME_BLOCKS: readonly TimeRange[] = Array.from({ length: 12 }, (_, i) => ({
	key: `${String(i * 2).padStart(2, '0')}-${String(i * 2 + 2).padStart(2, '0')}`,
	label: `${hh(i * 2)}h–${hh(i * 2 + 2)}h`,
	start: i * 120,
	end: i * 120 + 120
}));

/** RN-12: as faixas da noite, que aparecem sempre. */
const ALWAYS = new Set(['18-20', '20-22', '22-24']);

/** As faixas que aparecem sempre: "Qualquer horário" e as três da noite. */
export const TIME_RANGES: readonly TimeRange[] = [
	ANY_TIME,
	...TIME_BLOCKS.filter((b) => ALWAYS.has(b.key))
];

/**
 * RN-12 (revisão de 2026-10-09): "Qualquer horário", as três da noite, os blocos em que
 * algum lobby do dia começa (sem considerar os filtros) e o marcado, na ordem do relógio.
 */
export function visibleTimeRanges(
	dayLobbies: readonly Lobby[],
	selected: TimeRangeKey
): TimeRange[] {
	const used = new Set(dayLobbies.map((l) => Math.floor(toMinutes(l.time) / 120)));
	return [
		ANY_TIME,
		...TIME_BLOCKS.filter((b, i) => ALWAYS.has(b.key) || used.has(i) || b.key === selected)
	];
}

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
	const range = TIME_BLOCKS.find((r) => r.key === key) ?? ANY_TIME;
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
		instanceMatches(lobby, filters.instance) &&
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
		[ANY_TIME, ...TIME_BLOCKS].map((r) => [r.key, passing.filter((l) => inRange(l, r.key)).length])
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
	// RN-05 da lobby-sem-instancia: os títulos livres não viram opção de instância.
	return [...new Set(lobbies.filter((l) => !l.anyInstance).map((l) => l.instance))].sort((a, b) =>
		a.localeCompare(b, 'pt-BR')
	);
}

/**
 * RN-12; RN-05 da lobby-sem-instancia: "Sem instância definida" traz só os lobbies sem
 * instância, e uma instância do catálogo não traz eles.
 */
function instanceMatches(lobby: Lobby, instance: string): boolean {
	if (!instance) return true;
	if (instance === NO_INSTANCE) return lobby.anyInstance === true;
	return !lobby.anyInstance && lobby.instance === instance;
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
