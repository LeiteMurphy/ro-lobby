import { addDays } from './time';
import type { Lobby } from './types';

const WEEKDAYS = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb'];
const MONTHS = ['jan', 'fev', 'mar', 'abr', 'mai', 'jun', 'jul', 'ago', 'set', 'out', 'nov', 'dez'];

export const DAYS_AHEAD = 14;

export interface Day {
	/** YYYY-MM-DD */
	date: string;
	weekday: string;
	day: string;
	/** "ter, 29 set" */
	label: string;
	count: number;
	today: boolean;
}

/** RN-10: 14 dias a partir de hoje, cada um com a quantidade de grupos. */
export function buildDays(today: string, lobbies: readonly Lobby[]): Day[] {
	return Array.from({ length: DAYS_AHEAD }, (_, i) => {
		const date = addDays(today, i);
		const d = new Date(`${date}T12:00:00Z`);
		const weekday = WEEKDAYS[d.getUTCDay()];
		return {
			date,
			weekday,
			day: String(d.getUTCDate()).padStart(2, '0'),
			label: `${weekday.toLowerCase()}, ${d.getUTCDate()} ${MONTHS[d.getUTCMonth()]}`,
			count: lobbies.filter((l) => l.date === date).length,
			today: i === 0
		};
	});
}
