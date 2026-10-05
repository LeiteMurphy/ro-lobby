// Regras do login com Discord no web (spec login-discord, ADR-07). Funções puras, para
// serem testadas sem servidor.

export const SESSION_COOKIE = 'rol_session';
export const STATE_COOKIE = 'rol_oauth_state';
/** RN-09: a Sessão vale 30 dias a partir do último uso. */
export const SESSION_MAX_AGE = 30 * 24 * 60 * 60;
/** RN-02: o state vale 10 minutos. */
export const STATE_MAX_AGE = 10 * 60;
export const CALLBACK_PATH = '/auth/discord/callback';
export const DEFAULT_AUTHORIZE_URL = 'https://discord.com/oauth2/authorize';
/** RN-13: o que a pessoa vê quando o login falha. */
export const LOGIN_ERROR_MESSAGE = 'Não foi possível entrar com o Discord. Tente de novo.';
export const LOGIN_ERROR_PATH = '/?login=erro';

/** RN-02: state aleatório de 256 bits (a regra pede pelo menos 128), em base64url. */
export function newState(): string {
	const bytes = crypto.getRandomValues(new Uint8Array(32));
	return Buffer.from(bytes).toString('base64url');
}

/** CA-01.3: URL de autorização com scope=identify, response_type=code e o state (RN-01). */
export function authorizeUrl(params: {
	authorizeUrl: string;
	clientId: string;
	redirectUri: string;
	state: string;
}): string {
	const url = new URL(params.authorizeUrl);
	url.searchParams.set('client_id', params.clientId);
	url.searchParams.set('response_type', 'code');
	url.searchParams.set('scope', 'identify');
	url.searchParams.set('redirect_uri', params.redirectUri);
	url.searchParams.set('state', params.state);
	return url.toString();
}

/** D-08: o redirect_uri é montado a partir da origem do próprio web. */
export function redirectUri(origin: string): string {
	return `${origin}${CALLBACK_PATH}`;
}

/**
 * RN-12: só caminhos relativos do próprio site ("/..." sem "//" nem "/\"). Qualquer outro
 * valor volta para a Home.
 */
export function safeNext(next: string | null | undefined): string {
	if (!next || !next.startsWith('/') || next.startsWith('//') || next.startsWith('/\\')) return '/';
	// Caracteres de controle (quebra de linha, tab) também não valem.
	if ([...next].some((char) => char.charCodeAt(0) < 0x20)) return '/';
	return next;
}

/** O cookie de state guarda o state e o destino. */
export function encodeStateCookie(state: string, next: string): string {
	return Buffer.from(JSON.stringify({ state, next })).toString('base64url');
}

export function decodeStateCookie(
	value: string | undefined
): { state: string; next: string } | null {
	if (!value) return null;
	try {
		const parsed = JSON.parse(Buffer.from(value, 'base64url').toString('utf8'));
		if (typeof parsed?.state !== 'string' || typeof parsed?.next !== 'string') return null;
		return { state: parsed.state, next: safeNext(parsed.next) };
	} catch {
		return null;
	}
}

/** Compara o state recebido com o do cookie sem vazar tempo pela comparação. */
export function sameState(a: string, b: string): boolean {
	if (a.length !== b.length || a.length === 0) return false;
	let diff = 0;
	for (let i = 0; i < a.length; i++) diff |= a.charCodeAt(i) ^ b.charCodeAt(i);
	return diff === 0;
}

/** RN-08: HttpOnly, SameSite=Lax, Path=/ e Secure fora de localhost. */
export function cookieOptions(url: URL, maxAge: number) {
	const local = url.hostname === 'localhost' || url.hostname === '127.0.0.1';
	return { httpOnly: true, sameSite: 'lax' as const, path: '/', secure: !local, maxAge };
}
