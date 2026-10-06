// Chamada do servidor do web à API, com o resultado já traduzido (spec personagens D-05,
// spec lobbies D-06). O token de sessão só existe no servidor do web.
import type { components } from './schema.gen';

export type FieldError = components['schemas']['FieldError'];
export type ApiErrorCode = components['schemas']['Error']['error'];

const TIMEOUT_MS = 6000;

/**
 * Resultado de uma chamada:
 * - `no_session`: a API respondeu 401 (o cookie deve ser apagado e o Usuário levado ao
 *   login);
 * - `not_found`: não existe ou é de outro Usuário;
 * - `conflict`: 409, com o código do erro (limite atingido, lobby fechado, personagem num
 *   lobby aberto);
 * - `bad_request`: 400, com o código do erro;
 * - `invalid`: erros de campo;
 * - `unavailable`: a API não respondeu ou respondeu algo inesperado.
 */
export type Result<T> =
	| { ok: true; data: T }
	| { ok: false; kind: 'no_session' | 'not_found' | 'unavailable' }
	| { ok: false; kind: 'conflict' | 'bad_request'; code: ApiErrorCode | undefined }
	| { ok: false; kind: 'invalid'; fields: FieldError[] };

export interface Call {
	fetchFn: typeof fetch;
	apiBaseUrl: string;
	/** Vazio nas rotas públicas. */
	token: string;
}

export async function request<T>(
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
			case 400:
			case 409: {
				const parsed = (await res.json().catch(() => ({}))) as { error?: ApiErrorCode };
				return {
					ok: false,
					kind: res.status === 409 ? 'conflict' : 'bad_request',
					code: parsed.error
				};
			}
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
