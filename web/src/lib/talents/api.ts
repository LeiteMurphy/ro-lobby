// Chamadas do servidor do web à API do banco de talentos (spec banco-de-talentos, design
// seção 3). O token de sessão só existe no servidor do web.
import { request, type Call } from '$lib/api/request';
import type { components } from '$lib/api/schema.gen';

export type Availability = components['schemas']['Availability'];
export type AvailabilityInput = components['schemas']['AvailabilityInput'];
export type Talent = components['schemas']['Talent'];
export type Role = components['schemas']['Role'];

/** RN-01 a RN-05: liga, edita ou desliga o personagem no banco. */
export function setAvailability(call: Call, characterId: string, input: AvailabilityInput) {
	return request<{ availability: Availability | null }>(
		call,
		'PUT',
		`/characters/${encodeURIComponent(characterId)}/availability`,
		input
	);
}

export interface CatalogFilter {
	instanceId?: string;
	role?: Role;
	/** Dia da semana, domingo = 0. */
	day?: number;
	/** HH:MM de Brasília. */
	time?: string;
}

/**
 * RN-11 / RN-12: o catálogo. O token é opcional; sem ele, a API não manda o nome no
 * Discord (D-05).
 */
export function listTalents(call: Call, filter: CatalogFilter) {
	const query = new URLSearchParams();
	if (filter.instanceId) query.set('instanceId', filter.instanceId);
	if (filter.role) query.set('role', filter.role);
	if (filter.day !== undefined) query.set('day', String(filter.day));
	if (filter.time) query.set('time', filter.time);
	const qs = query.size ? `?${query}` : '';
	return request<Talent[]>(call, 'GET', `/talents${qs}`);
}

/** RN-09: os personagens com afinidade com o lobby aberto do dono. */
export function listLobbyTalents(call: Call, lobbyId: string) {
	return request<Talent[]>(call, 'GET', `/lobbies/${encodeURIComponent(lobbyId)}/talents`);
}

/** RN-15 da candidatura-lobby / CA-06.9: o dono desbloqueia a pessoa do personagem. */
export function unblockInLobby(call: Call, lobbyId: string, characterId: string) {
	return request<void>(call, 'POST', `/lobbies/${encodeURIComponent(lobbyId)}/unblock`, {
		characterId
	});
}

export interface CountInput {
	instanceId: string;
	startsAt: string;
	minLevel: number;
	characterId: string;
	/** Por função: as vagas de cada uma. */
	tank?: number;
	support?: number;
	dps?: number;
	/** Grupo livre (RN-13 da grupo-livre): formation=free e o total de vagas. */
	formation?: 'roles' | 'free';
	freeSlots?: number;
}

/** RN-10: a contagem da prévia da criação. */
export function countTalents(call: Call, input: CountInput) {
	const query = new URLSearchParams(
		Object.entries(input)
			.filter(([, v]) => v !== undefined)
			.map(([k, v]) => [k, String(v)] as [string, string])
	);
	return request<{ count: number }>(call, 'GET', `/talents/count?${query}`);
}
