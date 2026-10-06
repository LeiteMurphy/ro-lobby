// Chamadas do servidor do web à API de lobbies e ao catálogo de instâncias, pelos tipos
// gerados do contrato (spec lobbies, design D-06).
import { request, type Call } from '$lib/api/request';
import type { components } from '$lib/api/schema.gen';

export type ApiLobby = components['schemas']['Lobby'];
export type Instance = components['schemas']['Instance'];
export type LobbyInput = components['schemas']['LobbyInput'];
export type LobbyUpdate = components['schemas']['LobbyUpdate'];
export type Slots = components['schemas']['Slots'];
export type { Result } from '$lib/api/request';

const publicCall = (fetchFn: typeof fetch, apiBaseUrl: string): Call => ({
	fetchFn,
	apiBaseUrl,
	token: ''
});

/** CA-06.1: catálogo de instâncias, sem sessão, já na ordem da escolha (RN-02). */
export function listInstances(fetchFn: typeof fetch, apiBaseUrl: string) {
	return request<Instance[]>(publicCall(fetchFn, apiBaseUrl), 'GET', '/instances');
}

/** RN-13 / RN-22: lobbies abertos entre os dias `from` e `to` (YYYY-MM-DD, São Paulo). */
export function listLobbies(fetchFn: typeof fetch, apiBaseUrl: string, from: string, to: string) {
	const query = new URLSearchParams({ from, to });
	return request<ApiLobby[]>(publicCall(fetchFn, apiBaseUrl), 'GET', `/lobbies?${query}`);
}

/**
 * RN-15: o lobby em qualquer estado. A sessão é opcional e muda o que vem: membros com
 * Discord, pendentes e a própria candidatura (D-06 da candidatura-lobby).
 */
export function getLobby(fetchFn: typeof fetch, apiBaseUrl: string, id: string, token = '') {
	return request<ApiLobby>(
		{ fetchFn, apiBaseUrl, token },
		'GET',
		`/lobbies/${encodeURIComponent(id)}`
	);
}

export function createLobby(call: Call, input: LobbyInput) {
	return request<ApiLobby>(call, 'POST', '/lobbies', input);
}

export function updateLobby(call: Call, id: string, input: LobbyUpdate) {
	return request<ApiLobby>(call, 'PUT', `/lobbies/${encodeURIComponent(id)}`, input);
}

export function cancelLobby(call: Call, id: string, reason: string) {
	return request<ApiLobby>(call, 'POST', `/lobbies/${encodeURIComponent(id)}/cancel`, { reason });
}
