// Chamadas do servidor do web à API de personagens e ao catálogo de classes, pelos tipos
// gerados do contrato (spec personagens, design D-05 e D-07). O token de sessão só existe
// no servidor do web; o navegador nunca fala com a API direto.
import { request, type Call } from '$lib/api/request';
import type { components } from '$lib/api/schema.gen';

export type Character = components['schemas']['Character'];
export type CharacterInput = components['schemas']['CharacterInput'];
export type RoClassEntry = components['schemas']['Class'];
export type FieldError = components['schemas']['FieldError'];
export type Portrait = components['schemas']['Portrait'];
export type Role = components['schemas']['Role'];

export type { Result } from '$lib/api/request';

/** CA-07.1: catálogo de classes, sem sessão. */
export function listClasses(fetchFn: typeof fetch, apiBaseUrl: string) {
	return request<RoClassEntry[]>({ fetchFn, apiBaseUrl, token: '' }, 'GET', '/classes');
}

/** RN-01 / RN-17: personagens do Usuário, o principal primeiro. */
export function listCharacters(call: Call) {
	return request<Character[]>(call, 'GET', '/characters');
}

export function createCharacter(call: Call, input: CharacterInput) {
	return request<Character>(call, 'POST', '/characters', input);
}

export function updateCharacter(call: Call, id: string, input: CharacterInput) {
	return request<Character>(call, 'PUT', `/characters/${encodeURIComponent(id)}`, input);
}

export function deleteCharacter(call: Call, id: string) {
	return request<void>(call, 'DELETE', `/characters/${encodeURIComponent(id)}`);
}

export function setMainCharacter(call: Call, id: string) {
	return request<void>(call, 'PUT', `/characters/${encodeURIComponent(id)}/main`);
}
