import type { operations } from './schema.gen';

// RN-13: o web só fala com a API pelos tipos gerados do openapi.yaml.
type GetHealthz = operations['getHealthz'];
export type HealthBody = GetHealthz['responses'][200]['content']['application/json'];
export type DegradedBody = GetHealthz['responses'][503]['content']['application/json'];

/** Estado da API como a página de status mostra (RN-15). */
export type ApiStatus = 'online' | 'problema' | 'indisponivel';

export const statusMessages: Record<ApiStatus, string> = {
	online: 'API online',
	problema: 'API com problema',
	indisponivel: 'API indisponível'
};

/** Prazo para a API responder antes de a página mostrar "API indisponível". */
export const REQUEST_TIMEOUT_MS = 3000;

/**
 * Consulta o GET /healthz e traduz a resposta para a página de status. Nunca lança:
 * qualquer falha de rede ou de prazo vira "indisponivel" (CA-05.3).
 */
export async function checkApi(
	fetchFn: typeof fetch,
	baseUrl: string,
	timeoutMs = REQUEST_TIMEOUT_MS
): Promise<ApiStatus> {
	let response: Response;
	try {
		response = await fetchFn(new URL('/healthz', baseUrl), {
			signal: AbortSignal.timeout(timeoutMs),
			headers: { accept: 'application/json' }
		});
	} catch {
		return 'indisponivel';
	}

	if (response.status === 200) {
		const body: HealthBody = await response.json().catch(() => null);
		return body?.status === 'ok' ? 'online' : 'problema';
	}
	// 503 é o caso descrito no contrato. Qualquer outra resposta também significa que a
	// API respondeu, mas não está saudável.
	return 'problema';
}
