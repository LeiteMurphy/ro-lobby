import { json } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { SESSION_COOKIE } from '$lib/auth/oauth';
import { countTalents } from '$lib/talents/api';
import type { RequestHandler } from './$types';

// Contagem da prévia da criação (spec banco-de-talentos, RN-10, design seção 3). O
// navegador não fala com a API; este endpoint repassa a sessão do cookie. Sem sessão ou
// sem resposta, devolve count null e a prévia esconde a linha.

const FIELDS = ['instanceId', 'startsAt', 'minLevel', 'characterId'];
// Por função, as três vagas; no grupo livre (RN-13 da grupo-livre), o total.
const ROLE_FIELDS = ['tank', 'support', 'dps'];

export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const token = cookies.get(SESSION_COOKIE);
	const q = url.searchParams;
	const free = q.get('formation') === 'free';
	const needed = [...FIELDS, ...(free ? ['freeSlots'] : ROLE_FIELDS)];
	if (!token || needed.some((f) => !q.get(f))) return json({ count: null });
	const result = await countTalents(
		{ fetchFn: fetch, apiBaseUrl: authConfig().apiBaseUrl, token },
		{
			instanceId: q.get('instanceId') ?? '',
			startsAt: q.get('startsAt') ?? '',
			minLevel: Number(q.get('minLevel')),
			characterId: q.get('characterId') ?? '',
			...(free
				? { formation: 'free' as const, freeSlots: Number(q.get('freeSlots')) }
				: {
						tank: Number(q.get('tank')),
						support: Number(q.get('support')),
						dps: Number(q.get('dps'))
					})
		}
	);
	return json({ count: result.ok ? result.data.count : null });
};
