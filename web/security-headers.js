// Cabeçalhos de segurança de toda resposta do web (spec endurecimento, RN-11 e RN-12).
// É JavaScript puro porque é usado em dois lugares: no hook do SvelteKit (páginas e rotas)
// e no server.js (arquivos estáticos, que o adapter-node entrega antes do hook).

/** RN-11: vale para qualquer origem. */
export const SECURITY_HEADERS = Object.freeze({
	// Outro site não exibe o RO Lobby dentro de uma moldura (clickjacking).
	'X-Frame-Options': 'DENY',
	// O navegador não "adivinha" o tipo de um arquivo diferente do declarado.
	'X-Content-Type-Options': 'nosniff',
	// Links para fora levam só a origem, nunca o caminho (ex.: /lobbies/<id>).
	'Referrer-Policy': 'strict-origin-when-cross-origin',
	// O site não usa câmera, microfone nem localização.
	'Permissions-Policy': 'camera=(), microphone=(), geolocation=()'
});

/** RN-12: um ano de HTTPS obrigatório, só quando a origem já é https. */
export const HSTS = 'max-age=31536000';

/**
 * @param {boolean} https se a origem do site é https
 * @returns {Record<string, string>}
 */
export function headersFor(https) {
	/** @type {Record<string, string>} */
	const headers = { ...SECURITY_HEADERS };
	if (https) headers['Strict-Transport-Security'] = HSTS;
	return headers;
}

/**
 * Envolve um handler do Node (o do adapter-node) para toda resposta sair com os cabeçalhos,
 * inclusive os arquivos estáticos.
 * @param {(req: import('node:http').IncomingMessage, res: import('node:http').ServerResponse, next?: () => void) => void} next
 * @param {Record<string, string>} headers
 */
export function withHeaders(next, headers) {
	/**
	 * @param {import('node:http').IncomingMessage} req
	 * @param {import('node:http').ServerResponse} res
	 */
	return (req, res) => {
		for (const [name, value] of Object.entries(headers)) res.setHeader(name, value);
		next(req, res);
	};
}
