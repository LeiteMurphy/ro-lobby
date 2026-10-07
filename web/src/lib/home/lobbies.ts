import { toMinutes, type ZonedNow } from './time';
import { ROLES, type Composition, type Lobby, type Role } from './types';

/** RN-14: vagas abertas da função = total − ocupadas. */
export function openSlots(composition: Composition, role: Role): number {
	const { filled, total } = composition[role];
	return Math.max(0, total - filled);
}

export function headcount(composition: Composition): { filled: number; total: number } {
	return ROLES.reduce(
		(acc, role) => ({
			filled: acc.filled + composition[role].filled,
			total: acc.total + composition[role].total
		}),
		{ filled: 0, total: 0 }
	);
}

/** RN-14: lotado quando as vagas ocupadas somam o total. */
export function isFull(composition: Composition): boolean {
	const { filled, total } = headcount(composition);
	return filled >= total;
}

/**
 * RN-14: função com mais vagas abertas, que dá a cor da borda do card. No empate vale a
 * ordem Tank, Suporte, Dano. Devolve null quando o lobby está lotado.
 */
export function edgeRole(composition: Composition): Role | null {
	let best: Role | null = null;
	for (const role of ROLES) {
		const open = openSlots(composition, role);
		if (open > 0 && (best === null || open > openSlots(composition, best))) best = role;
	}
	return best;
}

/** RN-11: lobbies do dia em ordem crescente de horário. */
export function lobbiesForDay(lobbies: readonly Lobby[], date: string): Lobby[] {
	return lobbies
		.filter((l) => l.date === date)
		.sort((a, b) => toMinutes(a.time) - toMinutes(b.time));
}

/**
 * RN-16: primeiro lobby do dia, por horário, que não está lotado e, se o dia for hoje,
 * ainda não começou. Ignora os filtros.
 */
export function featuredLobby(dayLobbies: readonly Lobby[], now: ZonedNow): Lobby | null {
	return (
		dayLobbies.find(
			(l) => !isFull(l.composition) && (l.date !== now.date || toMinutes(l.time) > now.minutes)
		) ?? null
	);
}

/**
 * RN-34 da candidatura-lobby: o selo "N pendentes" aparece só para o dono do lobby, e só
 * quando há pendentes. Devolve o texto do selo ou nulo.
 */
export function pendingFor(lobby: Lobby, viewerId: string | null | undefined): string | null {
	const n = lobby.pendingCount ?? 0;
	if (!viewerId || viewerId !== lobby.ownerId || n < 1) return null;
	return n === 1 ? '1 pendente' : `${n} pendentes`;
}
