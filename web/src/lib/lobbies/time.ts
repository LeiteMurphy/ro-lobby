// Horário de Brasília na borda, UTC no resto (spec lobbies, RN-05, design D-04). O
// Usuário informa e vê dia e hora em America/Sao_Paulo; a API recebe e devolve UTC.
import { TIME_ZONE, zonedNow } from '$lib/home/time';

export interface ZonedDateTime {
	/** YYYY-MM-DD em São Paulo. */
	date: string;
	/** HH:MM em São Paulo. */
	time: string;
}

const offsetFormatter = new Intl.DateTimeFormat('en-US', {
	timeZone: TIME_ZONE,
	timeZoneName: 'longOffset'
});

/** Diferença, em minutos, entre São Paulo e UTC no instante `at` (ex.: -180). */
function offsetMinutes(at: Date): number {
	const name = offsetFormatter.formatToParts(at).find((p) => p.type === 'timeZoneName')?.value;
	const match = /GMT([+-])(\d{2}):?(\d{2})?/.exec(name ?? '');
	if (!match) return 0; // "GMT" puro
	const sign = match[1] === '-' ? -1 : 1;
	return sign * (Number(match[2]) * 60 + Number(match[3] ?? 0));
}

/** Converte dia e hora de São Paulo para o instante em UTC (ISO 8601). */
export function toUtcIso({ date, time }: ZonedDateTime): string {
	const [y, mo, d] = date.split('-').map(Number);
	const [h, mi] = time.split(':').map(Number);
	const asUtc = Date.UTC(y, mo - 1, d, h, mi);
	// Primeiro palpite com o deslocamento daquele instante; o segundo corrige se a hora cair
	// numa troca de deslocamento.
	let instant = asUtc - offsetMinutes(new Date(asUtc)) * 60_000;
	instant = asUtc - offsetMinutes(new Date(instant)) * 60_000;
	return new Date(instant).toISOString().replace('.000Z', 'Z');
}

/** Converte um instante (ISO 8601) para dia e hora de São Paulo. */
export function fromUtcIso(iso: string): ZonedDateTime {
	const at = new Date(iso);
	const { date, minutes } = zonedNow(at);
	const hh = String(Math.floor(minutes / 60)).padStart(2, '0');
	const mm = String(minutes % 60).padStart(2, '0');
	return { date, time: `${hh}:${mm}` };
}

/** Valida "HH:MM" de 00:00 a 23:59. */
export function isTime(value: string): boolean {
	return /^([01]\d|2[0-3]):[0-5]\d$/.test(value);
}
