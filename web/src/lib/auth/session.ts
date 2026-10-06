// Sessão nas páginas que exigem login (spec personagens RN-03, spec lobbies RN-04).
import { redirect, type Cookies } from '@sveltejs/kit';
import type { Call } from '$lib/api/request';
import { authConfig } from './config';
import { loginHref } from './display';
import { SESSION_COOKIE } from './oauth';

/** Apaga o cookie e leva ao login, voltando para `path` depois. */
export function toLogin(cookies: Cookies, path: string): never {
	cookies.delete(SESSION_COOKIE, { path: '/' });
	redirect(303, loginHref(new URL(path, 'http://web')));
}

/** Chamada à API com o token da sessão; sem cookie, leva ao login. */
export function sessionCall(cookies: Cookies, fetchFn: typeof fetch, path: string): Call {
	const token = cookies.get(SESSION_COOKIE);
	if (!token) toLogin(cookies, path);
	return { fetchFn, apiBaseUrl: authConfig().apiBaseUrl, token };
}
