import { json } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { SESSION_COOKIE } from '$lib/auth/oauth';
import { countTalents } from '$lib/talents/api';
import type { RequestHandler } from './$types';

// Contagem da prévia da criação (spec banco-de-talentos, RN-10, design seção 3). O
// navegador não fala com a API; este endpoint repassa a sessão do cookie. Sem sessão ou
// sem resposta, devolve count null e a prévia esconde a linha.

const FIELDS = ['instanceId', 'startsAt', 'minLevel', 'tank', 'support', 'dps', 'characterId'];

export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const token = cookies.get(SESSION_COOKIE);
	const q = url.searchParams;
	if (!token || FIELDS.some((f) => !q.get(f))) return json({ count: null });
	const result = await countTalents(
		{ fetchFn: fetch, apiBaseUrl: authConfig().apiBaseUrl, token },
		{
			instanceId: q.get('instanceId') ?? '',
			startsAt: q.get('startsAt') ?? '',
			minLevel: Number(q.get('minLevel')),
			tank: Number(q.get('tank')),
			support: Number(q.get('support')),
			dps: Number(q.get('dps')),
			characterId: q.get('characterId') ?? ''
		}
	);
	return json({ count: result.ok ? result.data.count : null });
};
