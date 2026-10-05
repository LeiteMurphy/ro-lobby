import { redirect } from '@sveltejs/kit';
import { deleteSession } from '$lib/auth/api';
import { authConfig } from '$lib/auth/config';
import { SESSION_COOKIE } from '$lib/auth/oauth';
import type { RequestHandler } from './$types';

// RN-11 / CA-03.1: encerra a sessão na API, apaga o cookie e volta para a Home. É um POST
// vindo de formulário, então o SvelteKit confere a origem (proteção CSRF).
export const POST: RequestHandler = async ({ cookies, fetch }) => {
	const token = cookies.get(SESSION_COOKIE);
	if (token) await deleteSession(fetch, authConfig().apiBaseUrl, token);
	cookies.delete(SESSION_COOKIE, { path: '/' });
	redirect(303, '/');
};
