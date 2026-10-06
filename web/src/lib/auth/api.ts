// Chamadas do servidor do web à API de autenticação, pelos tipos gerados do contrato
// (RN-13 da fundação, ADR-07).
import type { components } from '$lib/api/schema.gen';

export type SessionUser = components['schemas']['User'];
type SessionCreated = components['schemas']['SessionCreated'];

const TIMEOUT_MS = 6000;

/** CA-04.4 / CA-04.5: troca o código; qualquer falha vira null (nada é criado). */
export async function exchangeCode(
	fetchFn: typeof fetch,
	apiBaseUrl: string,
	code: string,
	redirectUri: string
): Promise<SessionCreated | null> {
	try {
		const res = await fetchFn(new URL('/auth/discord', apiBaseUrl), {
			method: 'POST',
			headers: { 'content-type': 'application/json', accept: 'application/json' },
			body: JSON.stringify({ code, redirectUri }),
			signal: AbortSignal.timeout(TIMEOUT_MS)
		});
		if (res.status !== 201) return null;
		return (await res.json()) as SessionCreated;
	} catch {
		return null;
	}
}

/**
 * CA-06.1 / CA-06.2: usuário da sessão. 'none' quando a API responde 401 (o cookie deve
 * ser apagado); 'unavailable' quando a API não responde (o cookie fica).
 */
export async function fetchMe(
	fetchFn: typeof fetch,
	apiBaseUrl: string,
	token: string
): Promise<SessionUser | 'none' | 'unavailable'> {
	try {
		const res = await fetchFn(new URL('/me', apiBaseUrl), {
			headers: { authorization: `Bearer ${token}`, accept: 'application/json' },
			signal: AbortSignal.timeout(3000)
		});
		if (res.status === 401) return 'none';
		if (res.status !== 200) return 'unavailable';
		return (await res.json()) as SessionUser;
	} catch {
		return 'unavailable';
	}
}

/** RN-11: encerra a sessão na API; falha de rede não impede apagar o cookie. */
export async function deleteSession(
	fetchFn: typeof fetch,
	apiBaseUrl: string,
	token: string
): Promise<void> {
	try {
		await fetchFn(new URL('/session', apiBaseUrl), {
			method: 'DELETE',
			headers: { authorization: `Bearer ${token}` },
			signal: AbortSignal.timeout(3000)
		});
	} catch {
		// O cookie é apagado de qualquer forma; a sessão vence sozinha em 30 dias.
	}
}
