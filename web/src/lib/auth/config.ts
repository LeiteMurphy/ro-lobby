import { env } from '$env/dynamic/private';
import { DEFAULT_AUTHORIZE_URL } from './oauth';

/** Configuração do login no servidor do web (D-09). O Client Secret não existe aqui (RN-03). */
export function authConfig() {
	return {
		apiBaseUrl: env.API_BASE_URL || 'http://localhost:8080',
		clientId: env.DISCORD_CLIENT_ID ?? '',
		authorizeUrl: env.DISCORD_AUTHORIZE_URL || DEFAULT_AUTHORIZE_URL
	};
}
