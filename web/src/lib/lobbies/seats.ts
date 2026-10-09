// Vagas do lobby conforme a formação (spec grupo-livre, RN-03, RN-04, design seção 3). O
// card, o detalhe, os diálogos e o convite perguntam aqui se há vaga, para a regra do
// grupo livre ficar num ponto só no web.
import type { Role } from '$lib/home/types';
import type { ApiLobby } from './api';

type Seats = Pick<ApiLobby, 'formation' | 'freeSlots' | 'slots' | 'occupied'>;

/** RN-01: o lobby é um grupo livre. */
export function isFree(lobby: Pick<ApiLobby, 'formation'>): boolean {
	return lobby.formation === 'free';
}

/** Ocupantes do lobby inteiro: o dono e os membros aceitos, de qualquer função. */
export function occupiedTotal(lobby: Seats): number {
	return lobby.occupied.tank + lobby.occupied.support + lobby.occupied.dps;
}

/** Total de vagas: o do grupo livre ou a soma das funções. */
export function totalSeats(lobby: Seats): number {
	return isFree(lobby)
		? (lobby.freeSlots ?? 0)
		: lobby.slots.tank + lobby.slots.support + lobby.slots.dps;
}

/** Vagas livres no lobby inteiro. */
export function openSeats(lobby: Seats): number {
	return Math.max(0, totalSeats(lobby) - occupiedTotal(lobby));
}

/**
 * Vagas para um personagem da função: no grupo livre, as livres do total (RN-03, RN-04);
 * por função, as da função (RN-05 da candidatura-lobby).
 */
export function roomFor(lobby: Seats, role: Role): number {
	if (isFree(lobby)) return openSeats(lobby);
	return Math.max(0, lobby.slots[role] - lobby.occupied[role]);
}
