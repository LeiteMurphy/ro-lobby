import { redirect } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import {
	authorizeUrl,
	cookieOptions,
	encodeStateCookie,
	newState,
	redirectUri,
	safeNext,
	STATE_COOKIE,
	STATE_MAX_AGE
} from '$lib/auth/oauth';
import type { RequestHandler } from './$types';

// RN-02 / CA-01.3: grava o state (e o destino) num cookie de 10 minutos e manda a pessoa
// para a autorização do Discord.
export const GET: RequestHandler = ({ url, cookies }) => {
	const { clientId, authorizeUrl: base } = authConfig();
	const state = newState();
	const next = safeNext(url.searchParams.get('next'));
	cookies.set(STATE_COOKIE, encodeStateCookie(state, next), cookieOptions(url, STATE_MAX_AGE));
	redirect(
		302,
		authorizeUrl({ authorizeUrl: base, clientId, redirectUri: redirectUri(url.origin), state })
	);
};
