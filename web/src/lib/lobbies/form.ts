// Leitura do formulário de lobby nas actions (spec lobbies, RN-05 a RN-09, RN-17, RNF-04).
// O web confere só o formato de dia e hora, para montar o horário em UTC; as regras valem
// na API.
import type { FieldError } from '$lib/api/request';
import type { LobbyInput, LobbyUpdate } from './api';
import { isTime, toUtcIso } from './time';

/** Valores como o Usuário digitou, para voltar ao formulário num erro (RN-19). */
export interface LobbyFormValues {
	instanceId: string;
	date: string;
	time: string;
	tank: string;
	support: string;
	dps: string;
	minLevel: string;
	characterId: string;
	note: string;
	/** spec grupo-livre, RN-01 e RN-02: formação e, no grupo livre, o total de vagas. */
	formation: 'roles' | 'free';
	freeSlots: string;
}

export function readLobbyForm(data: FormData): LobbyFormValues {
	const text = (name: string) => String(data.get(name) ?? '').trim();
	return {
		instanceId: text('instanceId'),
		date: text('date'),
		time: text('time'),
		tank: text('tank'),
		support: text('support'),
		dps: text('dps'),
		minLevel: text('minLevel'),
		characterId: text('characterId'),
		note: String(data.get('note') ?? ''),
		formation: text('formation') === 'free' ? 'free' : 'roles',
		freeSlots: text('freeSlots')
	};
}

/** Número inteiro, ou -1 quando vazio ou inválido (a API devolve o erro do campo). */
function int(value: string): number {
	const n = Number(value);
	return value !== '' && Number.isInteger(n) ? n : -1;
}

type Parsed<T> = { ok: true; input: T } | { ok: false; fields: FieldError[] };

/** Início em UTC a partir do dia e da hora de Brasília, ou null se faltar algum. */
export function startsAt(v: Pick<LobbyFormValues, 'date' | 'time'>): string | null {
	return /^\d{4}-\d{2}-\d{2}$/.test(v.date) && isTime(v.time)
		? toUtcIso({ date: v.date, time: v.time })
		: null;
}

function base(v: LobbyFormValues) {
	// spec grupo-livre, RN-02: no grupo livre, as vagas por função vão zeradas.
	const free = v.formation === 'free';
	return {
		formation: v.formation,
		slots: free
			? { tank: 0, support: 0, dps: 0 }
			: { tank: int(v.tank), support: int(v.support), dps: int(v.dps) },
		...(free ? { freeSlots: int(v.freeSlots) } : {}),
		minLevel: int(v.minLevel),
		...(v.note.trim() ? { note: v.note } : {})
	};
}

export function toLobbyInput(v: LobbyFormValues): Parsed<LobbyInput> {
	const start = startsAt(v);
	if (!start) return { ok: false, fields: [{ field: 'startsAt', code: 'invalid' }] };
	return {
		ok: true,
		input: { instanceId: v.instanceId, characterId: v.characterId, startsAt: start, ...base(v) }
	};
}

export function toLobbyUpdate(v: LobbyFormValues): Parsed<LobbyUpdate> {
	const start = startsAt(v);
	if (!start) return { ok: false, fields: [{ field: 'startsAt', code: 'invalid' }] };
	return { ok: true, input: { instanceId: v.instanceId, startsAt: start, ...base(v) } };
}
