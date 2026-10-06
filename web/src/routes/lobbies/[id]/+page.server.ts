import { error, fail } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import { sessionCall, toLogin } from '$lib/auth/session';
import { listClasses } from '$lib/characters/api';
import { UNAVAILABLE_MESSAGE } from '$lib/characters/messages';
import { cancelLobby, getLobby } from '$lib/lobbies/api';
import { lobbyConflictMessage, lobbyFieldMessages } from '$lib/lobbies/messages';
import type { Actions, PageServerLoad } from './$types';

// Detalhe público do lobby e cancelamento pelo dono (spec lobbies, RN-15, RN-16, RN-19).

export const load: PageServerLoad = async ({ params, locals, fetch }) => {
	const { apiBaseUrl } = authConfig();
	const [lobby, classes] = await Promise.all([
		getLobby(fetch, apiBaseUrl, params.id),
		listClasses(fetch, apiBaseUrl)
	]);
	if (!lobby.ok && lobby.kind === 'not_found') error(404, 'Lobby não encontrado');
	if (!lobby.ok) error(503, UNAVAILABLE_MESSAGE);
	const classId = lobby.data.owner.classId;
	return {
		lobby: lobby.data,
		ownerClass: classId
			? ((classes.ok ? classes.data.find((c) => c.id === classId)?.name : undefined) ?? classId)
			: null,
		// RN-16: as ações do dono aparecem só para ele.
		isOwner: locals.user?.id === lobby.data.owner.userId,
		now: new Date().toISOString(),
		loginHref: loginHref(new URL(`/lobbies/${params.id}`, 'http://web'))
	};
};

export const actions: Actions = {
	cancel: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const reason = String((await request.formData()).get('reason') ?? '');
		const result = await cancelLobby(call, params.id, reason);
		if (result.ok) return { done: 'cancel' as const };
		switch (result.kind) {
			case 'no_session':
				return toLogin(cookies, path);
			case 'invalid':
				return fail(422, { reason, errors: lobbyFieldMessages(result.fields), message: null });
			case 'not_found':
				error(404, 'Lobby não encontrado');
				break;
			case 'conflict':
				return fail(409, {
					reason,
					errors: {},
					message: lobbyConflictMessage(result.code) ?? UNAVAILABLE_MESSAGE
				});
		}
		return fail(503, { reason, errors: {}, message: UNAVAILABLE_MESSAGE });
	}
};
