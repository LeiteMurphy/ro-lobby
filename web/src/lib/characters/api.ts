// Chamadas do servidor do web à API de personagens e ao catálogo de classes, pelos tipos
// gerados do contrato (spec personagens, design D-05 e D-07). O token de sessão só existe
// no servidor do web; o navegador nunca fala com a API direto.
import type { components } from '$lib/api/schema.gen';

export type Character = components['schemas']['Character'];
export type CharacterInput = components['schemas']['CharacterInput'];
export type RoClassEntry = components['schemas']['Class'];
export type FieldError = components['schemas']['FieldError'];
export type Portrait = components['schemas']['Portrait'];
export type Role = components['schemas']['Role'];

const TIMEOUT_MS = 6000;

/**
 * Resultado de uma chamada:
 * - `no_session`: a API respondeu 401 (o cookie deve ser apagado e o Usuário levado ao
 *   login);
 * - `not_found`: o personagem não existe ou é de outro Usuário (RN-02);
 * - `limit`: o Usuário já tem 10 personagens (RN-11);
 * - `invalid`: erros de campo (RN-19);
 * - `unavailable`: a API não respondeu ou respondeu algo inesperado.
 */
export type Result<T> =
	| { ok: true; data: T }
	| { ok: false; kind: 'no_session' | 'not_found' | 'limit' | 'unavailable' }
	| { ok: false; kind: 'invalid'; fields: FieldError[] };

interface Call {
	fetchFn: typeof fetch;
	apiBaseUrl: string;
	token: string;
}

async function request<T>(
	{ fetchFn, apiBaseUrl, token }: Call,
	method: string,
	path: string,
	body?: unknown
): Promise<Result<T>> {
	let res: Response;
	try {
		const headers: Record<string, string> = { accept: 'application/json' };
		if (token) headers.authorization = `Bearer ${token}`;
		if (body !== undefined) headers['content-type'] = 'application/json';
		res = await fetchFn(new URL(path, apiBaseUrl), {
			method,
			headers,
			body: body === undefined ? undefined : JSON.stringify(body),
			signal: AbortSignal.timeout(TIMEOUT_MS)
		});
	} catch {
		return { ok: false, kind: 'unavailable' };
	}
	try {
		switch (res.status) {
			case 200:
			case 201:
				return { ok: true, data: (await res.json()) as T };
			case 204:
				return { ok: true, data: undefined as T };
			case 401:
				return { ok: false, kind: 'no_session' };
			case 404:
				return { ok: false, kind: 'not_found' };
			case 409:
				return { ok: false, kind: 'limit' };
			case 422: {
				const parsed = (await res.json()) as components['schemas']['ValidationError'];
				return { ok: false, kind: 'invalid', fields: parsed.fields };
			}
			default:
				return { ok: false, kind: 'unavailable' };
		}
	} catch {
		return { ok: false, kind: 'unavailable' };
	}
}

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
