import { loginHref } from '$lib/auth/display';
import { getHomeLobbies } from '$lib/home/fixtures';
import { zonedNow } from '$lib/home/time';
import type { PageServerLoad } from './$types';

// RN-20: a Home é renderizada no servidor com os grupos de hoje. "Hoje" vem do fuso
// America/Sao_Paulo (RN-09) e só muda ao recarregar a página.
export const load: PageServerLoad = ({ url }) => {
	const now = new Date();
	const today = zonedNow(now).date;
	return {
		now: now.toISOString(),
		today,
		lobbies: getHomeLobbies(today),
		// spec login-discord: link de login que volta para cá (RN-12) e aviso de falha (RN-13).
		loginHref: loginHref(url),
		loginError: url.searchParams.get('login') === 'erro'
	};
};
