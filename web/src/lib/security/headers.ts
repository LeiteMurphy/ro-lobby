// Cabeçalhos de segurança de toda resposta do web (spec endurecimento, RN-11 e RN-12).
// Funções puras, para serem testadas sem servidor.

/** RN-11: vale para qualquer origem. */
export const SECURITY_HEADERS: Readonly<Record<string, string>> = {
	// Outro site não exibe o RO Lobby dentro de uma moldura (clickjacking).
	'X-Frame-Options': 'DENY',
	// O navegador não "adivinha" o tipo de um arquivo diferente do declarado.
	'X-Content-Type-Options': 'nosniff',
	// Links para fora levam só a origem, nunca o caminho (ex.: /lobbies/<id>).
	'Referrer-Policy': 'strict-origin-when-cross-origin',
	// O site não usa câmera, microfone nem localização.
	'Permissions-Policy': 'camera=(), microphone=(), geolocation=()'
};

/** RN-12: um ano de HTTPS obrigatório, só quando a origem já é https. */
export const HSTS = 'max-age=31536000';

export function securityHeaders(url: URL): Record<string, string> {
	const headers: Record<string, string> = { ...SECURITY_HEADERS };
	if (url.protocol === 'https:') headers['Strict-Transport-Security'] = HSTS;
	return headers;
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
