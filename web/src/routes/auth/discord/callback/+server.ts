import { redirect } from '@sveltejs/kit';
import { exchangeCode } from '$lib/auth/api';
import { authConfig } from '$lib/auth/config';
import {
	cookieOptions,
	decodeStateCookie,
	LOGIN_ERROR_PATH,
	redirectUri,
	sameState,
	SESSION_COOKIE,
	SESSION_MAX_AGE,
	STATE_COOKIE
} from '$lib/auth/oauth';
import type { RequestHandler } from './$types';

// Retorno do Discord (ADR-07). O state é de uso único: o cookie é apagado aqui, aceito ou
// não (RN-02). Qualquer falha volta para a Home com a mensagem de erro, sem chamar a API
// quando o state não confere (RN-13).
export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const saved = decodeStateCookie(cookies.get(STATE_COOKIE));
	cookies.delete(STATE_COOKIE, { path: '/' });

	const code = url.searchParams.get('code');
	const state = url.searchParams.get('state') ?? '';
	if (!saved || !code || url.searchParams.has('error') || !sameState(state, saved.state)) {
		redirect(303, LOGIN_ERROR_PATH);
	}

	const { apiBaseUrl } = authConfig();
	const session = await exchangeCode(fetch, apiBaseUrl, code, redirectUri(url.origin));
	if (!session) redirect(303, LOGIN_ERROR_PATH);

	cookies.set(SESSION_COOKIE, session.sessionToken, cookieOptions(url, SESSION_MAX_AGE));
	redirect(303, saved.next);
};
