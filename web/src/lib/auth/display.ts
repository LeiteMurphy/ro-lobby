import type { SessionUser } from './api';

/** RN-15: o nome de exibição, ou o nome de usuário quando não há nome de exibição. */
export function displayName(user: Pick<SessionUser, 'username' | 'globalName'>): string {
	return user.globalName?.trim() || user.username;
}

/** RN-15: a inicial do nome, em maiúscula, no lugar do avatar. */
export function initial(name: string): string {
	const first = [...name.trim()][0] ?? '?';
	return first.toLocaleUpperCase('pt-BR');
}

/** RN-12: caminho atual para voltar depois do login, sem o aviso de erro de login. */
export function loginHref(url: URL): string {
	const params = new URLSearchParams(url.search);
	params.delete('login');
	const query = params.toString();
	const next = url.pathname + (query ? `?${query}` : '');
	return `/auth/discord/login?next=${encodeURIComponent(next)}`;
}
