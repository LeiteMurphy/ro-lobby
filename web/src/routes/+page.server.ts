import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import { listClasses } from '$lib/characters/api';
import { UNAVAILABLE_MESSAGE } from '$lib/characters/messages';
import { DAYS_AHEAD } from '$lib/home/days';
import { addDays, zonedNow } from '$lib/home/time';
import { listLobbies } from '$lib/lobbies/api';
import { toHomeLobby } from '$lib/lobbies/toHome';
import type { PageServerLoad } from './$types';

// RN-20 (home-local): a Home é renderizada no servidor com os grupos de hoje. "Hoje" vem do
// fuso America/Sao_Paulo (RN-09) e só muda ao recarregar a página. Os lobbies vêm da API
// (spec lobbies, RN-22, D-09); sem API, a lista fica vazia com o aviso de falha.
export const load: PageServerLoad = async ({ url, fetch }) => {
	const now = new Date();
	const today = zonedNow(now).date;
	const { apiBaseUrl } = authConfig();
	const [lobbies, classes] = await Promise.all([
		listLobbies(fetch, apiBaseUrl, today, addDays(today, DAYS_AHEAD - 1)),
		listClasses(fetch, apiBaseUrl)
	]);
	const classNames = new Map(classes.ok ? classes.data.map((c) => [c.id, c.name]) : []);
	return {
		now: now.toISOString(),
		today,
		lobbies: lobbies.ok ? lobbies.data.map((l) => toHomeLobby(l, classNames)) : [],
		loadError: lobbies.ok ? null : UNAVAILABLE_MESSAGE,
		// spec login-discord: link de login que volta para cá (RN-12) e aviso de falha (RN-13).
		loginHref: loginHref(url),
		loginError: url.searchParams.get('login') === 'erro'
	};
};
