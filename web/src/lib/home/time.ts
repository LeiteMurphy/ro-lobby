// RN-09: "hoje", os horários e o tempo relativo usam sempre America/Sao_Paulo, no
// servidor e no navegador.
export const TIME_ZONE = 'America/Sao_Paulo';

export interface ZonedNow {
	/** YYYY-MM-DD */
	date: string;
	/** Minutos desde a meia-noite. */
	minutes: number;
}

const formatter = new Intl.DateTimeFormat('en-CA', {
	timeZone: TIME_ZONE,
	year: 'numeric',
	month: '2-digit',
	day: '2-digit',
	hour: '2-digit',
	minute: '2-digit',
	hourCycle: 'h23'
});

export function zonedNow(now: Date): ZonedNow {
	const parts = Object.fromEntries(formatter.formatToParts(now).map((p) => [p.type, p.value]));
	return {
		date: `${parts.year}-${parts.month}-${parts.day}`,
		minutes: Number(parts.hour) * 60 + Number(parts.minute)
	};
}

export function toMinutes(time: string): number {
	const [h, m] = time.split(':').map(Number);
	return h * 60 + m;
}

/** Soma dias a uma data YYYY-MM-DD, sem depender do fuso da máquina. */
export function addDays(date: string, days: number): string {
	const d = new Date(`${date}T12:00:00Z`);
	d.setUTCDate(d.getUTCDate() + days);
	return d.toISOString().slice(0, 10);
}

/**
 * RN-15: tempo relativo só para hoje e horário futuro: "em N min", "em H h" ou
 * "em H h M min". Devolve string vazia nos outros casos.
 */
export function relativeLabel(date: string, time: string, now: ZonedNow): string {
	if (date !== now.date) return '';
	const diff = toMinutes(time) - now.minutes;
	if (diff <= 0) return '';
	if (diff < 60) return `em ${diff} min`;
	const h = Math.floor(diff / 60);
	const m = diff % 60;
	return m ? `em ${h} h ${m} min` : `em ${h} h`;
}
