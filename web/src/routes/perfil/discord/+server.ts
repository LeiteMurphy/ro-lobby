import { redirect } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import {
	authorizeUrl,
	CONNECT_PATH,
	CONNECT_SCOPE,
	CONNECT_STATE_COOKIE,
	cookieOptions,
	newState,
	SESSION_COOKIE,
	STATE_MAX_AGE
} from '$lib/auth/oauth';
import type { RequestHandler } from './$types';

// Início da conexão ao servidor do RO Lobby (spec discord-bot, RN-01, CA-01.3): pede
// identify e guilds.join, com um state novo num cookie de 10 minutos. Sem sessão, vai ao
// login e volta para o perfil.
export const GET: RequestHandler = ({ url, cookies, locals }) => {
	if (!locals.user || !cookies.get(SESSION_COOKIE)) {
		redirect(303, loginHref(new URL('http://web/perfil')));
	}
	const { clientId, authorizeUrl: base } = authConfig();
	const state = newState();
	cookies.set(CONNECT_STATE_COOKIE, state, cookieOptions(url, STATE_MAX_AGE));
	redirect(
		302,
		authorizeUrl({
			authorizeUrl: base,
			clientId,
			redirectUri: `${url.origin}${CONNECT_PATH}`,
			state,
			scope: CONNECT_SCOPE
		})
	);
};
