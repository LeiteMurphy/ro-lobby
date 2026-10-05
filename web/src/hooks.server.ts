import type { Handle } from '@sveltejs/kit';
import { fetchMe } from '$lib/auth/api';
import { authConfig } from '$lib/auth/config';
import { cookieOptions, SESSION_COOKIE, SESSION_MAX_AGE } from '$lib/auth/oauth';

// Identifica o usuário da sessão em cada requisição (D-03). Sessão recusada pela API vira
// visitante e o cookie é apagado (RN-10); API fora do ar também vira visitante, mas o
// cookie fica, porque a sessão pode continuar válida.
export const handle: Handle = async ({ event, resolve }) => {
	event.locals.user = null;
	const token = event.cookies.get(SESSION_COOKIE);
	if (token) {
		const me = await fetchMe(event.fetch, authConfig().apiBaseUrl, token);
		if (me === 'none') {
			event.cookies.delete(SESSION_COOKIE, { path: '/' });
		} else if (me !== 'unavailable') {
			event.locals.user = me;
			// RN-09: o cookie acompanha a renovação da sessão pelo uso.
			event.cookies.set(SESSION_COOKIE, token, cookieOptions(event.url, SESSION_MAX_AGE));
		}
	}
	return resolve(event);
};
