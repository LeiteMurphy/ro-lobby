// Horário de Brasília na borda, UTC no resto (spec lobbies, RN-05, design D-04). O
// Usuário informa e vê dia e hora em America/Sao_Paulo; a API recebe e devolve UTC.
import { resolve } from '$app/paths';
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

/** Hora padrão de um lobby novo (RN-24). */
export const DEFAULT_TIME = '20:00';

/**
 * RN-24: início padrão da criação. `requested` é o dia escolhido na Home (YYYY-MM-DD) e só
 * vale se estiver em `days`, os dias que aceitam lobby (o primeiro é hoje). Sem dia, é
 * amanhã. Hoje, a hora nunca fica no passado: 20:00 ou a próxima hora cheia; sem hora
 * cheia sobrando, amanhã às 20:00.
 */
export function defaultStart(
	requested: string | null,
	now: { date: string; minutes: number },
	days: readonly string[]
): ZonedDateTime {
	const tomorrow = days[1] ?? now.date;
	const date = requested && days.includes(requested) ? requested : tomorrow;
	if (date !== now.date) return { date, time: DEFAULT_TIME };
	const [h, m] = DEFAULT_TIME.split(':').map(Number);
	if (now.minutes < h * 60 + m) return { date, time: DEFAULT_TIME };
	const next = Math.floor(now.minutes / 60) + 1;
	if (next > 23) return { date: tomorrow, time: DEFAULT_TIME };
	return { date, time: `${String(next).padStart(2, '0')}:00` };
}

/** RN-23 / RN-24: link de "Criar lobby", com o dia escolhido na Home quando há. */
export function createLobbyHref(date: string | null | undefined): string {
	const path = resolve('/lobbies/novo');
	return date ? `${path}?${new URLSearchParams({ dia: date })}` : path;
}
