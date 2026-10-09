// Convite e preview do lobby (spec compartilhar-lobby). Funções puras: a página só junta o
// lobby com a origem do web.
import type { ApiLobby } from './api';
import { fromUtcIso } from './time';
import { ROLE_LABELS } from '$lib/home/types';
import { isFree, openSeats } from './seats';
import { DELETED_CHARACTER } from './toHome';

const WEEKDAYS = ['domingo', 'segunda', 'terça', 'quarta', 'quinta', 'sexta', 'sábado'];

/** RN-02: "sábado, 10/10 às 20:00", em Brasília e com a data absoluta. */
export function whenLabel(startsAt: string): string {
	const { date, time } = fromUtcIso(startsAt);
	const [y, m, d] = date.split('-').map(Number);
	const weekday = WEEKDAYS[new Date(Date.UTC(y, m - 1, d)).getUTCDay()];
	return `${weekday}, ${String(d).padStart(2, '0')}/${String(m).padStart(2, '0')} às ${time}`;
}

/**
 * RN-03: "Vagas: 1 Tank, 2 Dano", só as funções com vaga, ou "Grupo lotado". No grupo
 * livre (RN-12 da grupo-livre): "Vagas: 8 livres" ou "Vagas: 1 livre".
 */
export function openSlotsLabel(
	lobby: Pick<ApiLobby, 'slots' | 'occupied'> & Partial<Pick<ApiLobby, 'formation' | 'freeSlots'>>
): string {
	if (lobby.formation && isFree({ formation: lobby.formation })) {
		const n = openSeats({ ...lobby, formation: lobby.formation, freeSlots: lobby.freeSlots ?? 0 });
		if (n < 1) return 'Grupo lotado';
		return `Vagas: ${n} ${n === 1 ? 'livre' : 'livres'}`;
	}
	const open = (['tank', 'support', 'dps'] as const)
		.map((role) => ({ role, n: lobby.slots[role] - lobby.occupied[role] }))
		.filter(({ n }) => n > 0)
		.map(({ role, n }) => `${n} ${ROLE_LABELS[role]}`);
	return open.length ? `Vagas: ${open.join(', ')}` : 'Grupo lotado';
}

/** Link público do lobby, a partir da origem do web (ORIGIN no adapter-node). */
export function lobbyUrl(origin: string, id: string): string {
	return `${origin}/lobbies/${encodeURIComponent(id)}`;
}

/** RN-02 / RN-03: o convite de três linhas que "Compartilhar" copia. */
export function inviteText(lobby: ApiLobby, origin: string): string {
	return [
		`Grupo para ${lobby.instance.name} · ${whenLabel(lobby.startsAt)} (horário de Brasília)`,
		`${openSlotsLabel(lobby)} · Nível mínimo ${lobby.minLevel}`,
		`Candidate-se: ${lobbyUrl(origin, lobby.id)}`
	].join('\n');
}

/** RN-05: theme-color do preview, a cor de destaque da marca (--gold-400). */
export const THEME_COLOR = '#f6bb45';

export interface ShareMeta {
	title: string;
	description: string;
	url: string;
}

/** RN-05: título, descrição e link das meta tags Open Graph. */
export function shareMeta(lobby: ApiLobby, origin: string): ShareMeta {
	const description =
		lobby.status === 'cancelled'
			? 'Esse grupo foi cancelado.'
			: lobby.status === 'started'
				? 'Esse grupo já começou.'
				: `${openSlotsLabel(lobby)} · Nível mínimo ${lobby.minLevel} · Anfitrião ${lobby.owner.nick ?? DELETED_CHARACTER}`;
	return {
		title: `${lobby.instance.name} · ${whenLabel(lobby.startsAt)}`,
		description,
		url: lobbyUrl(origin, lobby.id)
	};
}
