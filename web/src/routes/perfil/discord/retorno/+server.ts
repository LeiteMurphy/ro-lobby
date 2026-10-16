import { redirect } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { CONNECT_PATH, CONNECT_STATE_COOKIE, sameState, SESSION_COOKIE } from '$lib/auth/oauth';
import { connect } from '$lib/discord/api';
import type { ConnectOutcome } from '$lib/discord/messages';
import type { RequestHandler } from './$types';

// Retorno da conexão ao servidor do RO Lobby (spec discord-bot, RN-01 a RN-04). O state é de
// uso único; qualquer resultado volta ao perfil com a mensagem em `?discord=`.
export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	const saved = cookies.get(CONNECT_STATE_COOKIE) ?? '';
	cookies.delete(CONNECT_STATE_COOKIE, { path: '/' });
	const back: (outcome: ConnectOutcome) => never = (outcome) =>
		redirect(303, `/perfil?discord=${outcome}`);

	if (url.searchParams.has('error')) back('cancelado'); // CA-01.5
	const code = url.searchParams.get('code');
	const state = url.searchParams.get('state') ?? '';
	const token = cookies.get(SESSION_COOKIE);
	if (!code || !token || !sameState(state, saved)) back('erro');

	const result = await connect(
		{ fetchFn: fetch, apiBaseUrl: authConfig().apiBaseUrl, token: token ?? '' },
		code ?? '',
		`${url.origin}${CONNECT_PATH}`
	);
	if (result.ok) back('conectado');
	if (!result.ok && result.kind === 'rule' && result.code === 'wrong_account') back('conta'); // CA-01.4
	if (!result.ok && result.kind === 'disabled') back('desligado');
	if (!result.ok && result.kind === 'rate_limited') back('muitas'); // RN-15 da endurecimento
	back('erro');
};
