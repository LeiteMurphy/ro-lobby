import type { APIRequestContext } from '@playwright/test';

// Dados do ponta a ponta criados direto na API (spec lobbies, T-10): o login passa pelo
// Discord falso por HTTP, e personagens e lobbies vão pelas rotas da API com o token. O
// banco do ponta a ponta começa vazio a cada execução (playwright.config.ts).

const API = `http://localhost:${process.env.API_PORT || '8080'}`;
const FAKE_DISCORD = 'http://127.0.0.1:8090';
const REDIRECT = 'http://localhost:4173/auth/discord/callback';

export const rand = () => Math.random().toString(36).slice(2, 7);

/** Entra como `username` no Discord falso e devolve o token de sessão da API. */
export async function apiLogin(request: APIRequestContext, username: string): Promise<string> {
	const authorize = await request.post(`${FAKE_DISCORD}/oauth2/authorize`, {
		form: { redirect_uri: REDIRECT, state: 'seed', decision: 'allow', username },
		maxRedirects: 0
	});
	const code = new URL(authorize.headers()['location']).searchParams.get('code');
	const session = await request.post(`${API}/auth/discord`, {
		data: { code, redirectUri: REDIRECT }
	});
	if (session.status() !== 201)
		throw new Error(`login: ${session.status()} ${await session.text()}`);
	return (await session.json()).sessionToken;
}

async function post<T>(
	request: APIRequestContext,
	token: string,
	path: string,
	data: unknown
): Promise<T> {
	const res = await request.post(`${API}${path}`, {
		data,
		headers: { authorization: `Bearer ${token}` }
	});
	if (!res.ok()) throw new Error(`POST ${path}: ${res.status()} ${await res.text()}`);
	return res.json();
}

export interface SeedCharacter {
	id: string;
	nick: string;
}

export function createCharacter(
	request: APIRequestContext,
	token: string,
	c: { nick: string; classId: string; level: number; role: 'tank' | 'support' | 'dps' }
) {
	return post<SeedCharacter>(request, token, '/characters', c);
}

export function createLobby(
	request: APIRequestContext,
	token: string,
	l: {
		instanceId: string;
		day: number;
		time: string;
		characterId: string;
		minLevel: number;
		slots?: { tank: number; support: number; dps: number };
		note?: string;
	}
) {
	return post<{ id: string }>(request, token, '/lobbies', {
		instanceId: l.instanceId,
		startsAt: startsAt(l.day, l.time),
		slots: l.slots ?? { tank: 1, support: 2, dps: 3 },
		minLevel: l.minLevel,
		characterId: l.characterId,
		...(l.note ? { note: l.note } : {})
	});
}

const spDate = new Intl.DateTimeFormat('en-CA', {
	timeZone: 'America/Sao_Paulo',
	year: 'numeric',
	month: '2-digit',
	day: '2-digit'
});

/** Dia (YYYY-MM-DD) em São Paulo, daqui a `day` dias. */
export function spDay(day: number): string {
	const today = spDate.format(new Date());
	const d = new Date(`${today}T12:00:00Z`);
	d.setUTCDate(d.getUTCDate() + day);
	return d.toISOString().slice(0, 10);
}

/** Início em UTC de `time` (HH:MM, Brasília) daqui a `day` dias. São Paulo está em UTC-3. */
export function startsAt(day: number, time: string): string {
	return new Date(`${spDay(day)}T${time}:00-03:00`).toISOString();
}
