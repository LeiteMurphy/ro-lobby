// Cabeçalhos de segurança das páginas e rotas do SvelteKit (spec endurecimento, RN-11 e
// RN-12). Os valores ficam em security-headers.js, na raiz do web, que o server.js também usa
// para os arquivos estáticos.
import { HSTS, headersFor, SECURITY_HEADERS } from '../../../security-headers.js';

export { HSTS, SECURITY_HEADERS };

export function securityHeaders(url: URL): Record<string, string> {
	return headersFor(url.protocol === 'https:');
}

/**
 * Aplica os cabeçalhos na resposta. Uma resposta repassada de um fetch tem cabeçalhos
 * imutáveis; nesse caso, ela é copiada antes.
 */
export function withSecurityHeaders(response: Response, url: URL): Response {
	const headers = securityHeaders(url);
	try {
		for (const [name, value] of Object.entries(headers)) response.headers.set(name, value);
		return response;
	} catch {
		const copy = new Response(response.body, response);
		for (const [name, value] of Object.entries(headers)) copy.headers.set(name, value);
		return copy;
	}
}
