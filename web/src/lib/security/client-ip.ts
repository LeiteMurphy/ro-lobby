// IP de quem acessa, repassado à API para o limite de requisições (spec endurecimento, RN-16).
// Pelo túnel, o Cloudflare informa o IP real em CF-Connecting-IP e sobrescreve qualquer
// valor que o visitante tente mandar. O web só é alcançado pelo túnel ou pelo próprio PC
// (RN-06), então o cabeçalho é confiável aqui.

/** Cabeçalho que a API lê (ClientIPHeader no backend). */
export const CLIENT_IP_HEADER = 'X-Client-IP';

export function clientIp(request: Request, connectionAddress: () => string): string {
	const fromCloudflare = request.headers.get('cf-connecting-ip')?.trim();
	if (fromCloudflare) return fromCloudflare;
	try {
		return connectionAddress();
	} catch {
		return '';
	}
}

/** Só as chamadas à API levam o IP; nada vai para o Discord nem para outro destino. */
export function withClientIp(request: Request, apiBaseUrl: string, ip: string): Request {
	if (!ip || !isApiRequest(request.url, apiBaseUrl)) return request;
	const forwarded = new Request(request);
	forwarded.headers.set(CLIENT_IP_HEADER, ip);
	return forwarded;
}

function isApiRequest(url: string, apiBaseUrl: string): boolean {
	try {
		return new URL(url).origin === new URL(apiBaseUrl).origin;
	} catch {
		return false;
	}
}
