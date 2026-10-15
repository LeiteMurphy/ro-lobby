// Chamadas do servidor do web à API da sala do grupo no Discord (spec discord-bot). As
// respostas de erro dessas rotas têm formato próprio (DiscordError, DiscordRuleError e
// DiscordCooldown), então o resultado também é próprio.
import type { Call } from '$lib/api/request';
import type { components } from '$lib/api/schema.gen';

export type DiscordConnection = components['schemas']['DiscordConnection'];
export type DiscordRoom = components['schemas']['DiscordRoom'];
export type LobbyDiscord = components['schemas']['LobbyDiscord'];
export type DiscordRuleCode = components['schemas']['DiscordRuleError']['code'];

const TIMEOUT_MS = 12000; // a abertura da sala pode levar até 10 s (RNF-02)

export type DiscordResult<T> =
	| { ok: true; data: T }
	| {
			ok: false;
			kind:
				| 'no_session'
				| 'not_found'
				| 'disabled'
				| 'invalid_code'
				| 'guild_full'
				| 'rate_limited'
				| 'unavailable';
	  }
	| { ok: false; kind: 'rule'; code: DiscordRuleCode }
	| { ok: false; kind: 'cooldown'; retryAfterSeconds: number };

async function discordRequest<T>(
	{ fetchFn, apiBaseUrl, token }: Call,
	method: string,
	path: string,
	body?: unknown
): Promise<DiscordResult<T>> {
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
		if (res.status === 204) return { ok: true, data: undefined as T };
		if (res.status === 200 || res.status === 201)
			return { ok: true, data: (await res.json()) as T };
		if (res.status === 401) return { ok: false, kind: 'no_session' };
		const parsed = (await res.json().catch(() => ({}))) as {
			error?: string;
			code?: DiscordRuleCode;
			retryAfterSeconds?: number;
		};
		switch (res.status) {
			case 404:
				return { ok: false, kind: parsed.error === 'discord_disabled' ? 'disabled' : 'not_found' };
			case 409:
				return parsed.code
					? { ok: false, kind: 'rule', code: parsed.code }
					: { ok: false, kind: 'unavailable' };
			case 422:
				return { ok: false, kind: 'invalid_code' };
			case 429:
				// RN-15 da endurecimento: o limite geral de requisições também responde 429.
				if (parsed.error === 'rate_limited') return { ok: false, kind: 'rate_limited' };
				return { ok: false, kind: 'cooldown', retryAfterSeconds: parsed.retryAfterSeconds ?? 0 };
			case 503:
				return { ok: false, kind: parsed.error === 'guild_full' ? 'guild_full' : 'unavailable' };
			default:
				return { ok: false, kind: 'unavailable' };
		}
	} catch {
		return { ok: false, kind: 'unavailable' };
	}
}

/** RN-01 / RN-22: se a sala está ligada e quando o Usuário se conectou. */
export function getConnection(call: Call) {
	return discordRequest<DiscordConnection>(call, 'GET', '/me/discord-connection');
}

/** RN-01 a RN-04: troca o código da conexão e inclui o Usuário no servidor. */
export function connect(call: Call, code: string, redirectUri: string) {
	return discordRequest<DiscordConnection>(call, 'POST', '/me/discord-connection', {
		code,
		redirectUri
	});
}

/** RN-05 a RN-09: abre a sala do lobby. */
export function openRoom(call: Call, lobbyId: string) {
	return discordRequest<DiscordRoom>(
		call,
		'POST',
		`/lobbies/${encodeURIComponent(lobbyId)}/discord-room`
	);
}

/** RN-26 a RN-28: avisa o grupo na sala. */
export function alertRoom(call: Call, lobbyId: string) {
	return discordRequest<void>(
		call,
		'POST',
		`/lobbies/${encodeURIComponent(lobbyId)}/discord-alert`
	);
}
