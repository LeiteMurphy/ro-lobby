// Textos do banco de talentos (spec banco-de-talentos, RN-02 a RN-04, RN-12).

/** Dias da semana, domingo = 0, como a API manda. */
export const WEEKDAYS = ['dom', 'seg', 'ter', 'qua', 'qui', 'sex', 'sáb'] as const;
export const WEEKDAY_NAMES = [
	'domingo',
	'segunda',
	'terça',
	'quarta',
	'quinta',
	'sexta',
	'sábado'
] as const;

/** Ordem de exibição: a semana começa na segunda. */
export const WEEK_ORDER = [1, 2, 3, 4, 5, 6, 0] as const;

/** RN-03: as horas de 30 em 30 minutos, de 00:00 a 23:30. */
export const CLOCK_OPTIONS: readonly string[] = Array.from({ length: 48 }, (_, i) => {
	const h = String(Math.floor(i / 2)).padStart(2, '0');
	return `${h}:${i % 2 ? '30' : '00'}`;
});

/**
 * RN-02: "todos os dias", "seg a sex" para dias seguidos (três ou mais, na ordem de
 * segunda a domingo) ou a lista, como "seg, qua, sex".
 */
export function daysLabel(days: readonly number[]): string {
	const set = new Set(days);
	if (set.size === 7) return 'todos os dias';
	const ordered = WEEK_ORDER.filter((d) => set.has(d));
	const first = WEEK_ORDER.indexOf(ordered[0] as (typeof WEEK_ORDER)[number]);
	const consecutive = ordered.every((d, i) => WEEK_ORDER[first + i] === d);
	if (ordered.length >= 3 && consecutive) {
		return `${WEEKDAYS[ordered[0]]} a ${WEEKDAYS[ordered[ordered.length - 1]]}`;
	}
	return ordered.map((d) => WEEKDAYS[d]).join(', ');
}

/** RN-03: "19:00–23:00"; a faixa que vira a meia-noite ganha "(até o dia seguinte)". */
export function rangeLabel(start: string, end: string): string {
	const range = `${start}–${end}`;
	return end < start ? `${range} (até o dia seguinte)` : range;
}

/** RN-04: "Qualquer instância" ou os nomes. */
export function instancesLabel(anyInstance: boolean, names: readonly string[]): string {
	return anyInstance || names.length === 0 ? 'Qualquer instância' : names.join(', ');
}
