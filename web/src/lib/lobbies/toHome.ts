// Converte o lobby da API para o lobby que os componentes da Home já usam (spec lobbies,
// RN-22, design D-09).
import type { Lobby } from '$lib/home/types';
import type { ApiLobby } from './api';
import { fromUtcIso } from './time';

/** Nome mostrado quando o personagem do dono foi excluído depois do início (D-01). */
export const DELETED_CHARACTER = 'Personagem excluído';

/**
 * @param classNames id da classe → nome, do catálogo de GET /classes. Classe fora do
 *   catálogo aparece pelo id (borda de RN-06 da spec personagens).
 */
export function toHomeLobby(lobby: ApiLobby, classNames: ReadonlyMap<string, string>): Lobby {
	const { date, time } = fromUtcIso(lobby.startsAt);
	const classId = lobby.owner.classId;
	return {
		id: lobby.id,
		date,
		time,
		instance: lobby.instance.name,
		host: lobby.owner.nick ?? DELETED_CHARACTER,
		hostClass: classId ? (classNames.get(classId) ?? classId) : '',
		minLevel: lobby.minLevel,
		ownerId: lobby.owner.userId,
		...(lobby.instance.id === null ? { anyInstance: true } : {}),
		pendingCount: lobby.pendingCount,
		composition: {
			tank: { filled: lobby.occupied.tank, total: lobby.slots.tank },
			support: { filled: lobby.occupied.support, total: lobby.slots.support },
			dps: { filled: lobby.occupied.dps, total: lobby.slots.dps }
		},
		// RN-08 da grupo-livre: ocupantes de qualquer função contra o total livre.
		...(lobby.formation === 'free'
			? {
					free: {
						filled: lobby.occupied.tank + lobby.occupied.support + lobby.occupied.dps,
						total: lobby.freeSlots ?? 0
					}
				}
			: {})
	};
}
