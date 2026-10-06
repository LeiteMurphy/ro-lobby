import { error, fail, redirect } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import { sessionCall, toLogin } from '$lib/auth/session';
import { listCharacters, listClasses } from '$lib/characters/api';
import { UNAVAILABLE_MESSAGE } from '$lib/characters/messages';
import { buildDays } from '$lib/home/days';
import { zonedNow } from '$lib/home/time';
import { getLobby, updateLobby } from '$lib/lobbies/api';
import { readLobbyForm, toLobbyUpdate, type LobbyFormValues } from '$lib/lobbies/form';
import { lobbyConflictMessage, lobbyFieldMessages } from '$lib/lobbies/messages';
import { fromUtcIso } from '$lib/lobbies/time';
import type { Actions, PageServerLoad } from './$types';

// Edição do lobby: só o dono, só aberto; instância e personagem fixos (spec lobbies,
// RN-17, RN-18, RN-20).

export const load: PageServerLoad = async ({ params, locals, cookies, fetch }) => {
	const path = `/lobbies/${params.id}/editar`;
	if (!locals.user) toLogin(cookies, path);
	const call = sessionCall(cookies, fetch, path);
	const { apiBaseUrl } = authConfig();
	const [lobby, characters, classes] = await Promise.all([
		getLobby(fetch, apiBaseUrl, params.id),
		listCharacters(call),
		listClasses(fetch, apiBaseUrl)
	]);
	if (!lobby.ok && lobby.kind === 'not_found') error(404, 'Lobby não encontrado');
	if (!lobby.ok) error(503, UNAVAILABLE_MESSAGE);
	// RN-20: só o dono edita; para os outros, o lobby "não existe" aqui.
	if (lobby.data.owner.userId !== locals.user.id) error(404, 'Lobby não encontrado');
	// RN-17: lobby iniciado ou cancelado não se edita; volta para o detalhe.
	if (lobby.data.status !== 'open') redirect(303, `/lobbies/${params.id}`);

	const l = lobby.data;
	const { date, time } = fromUtcIso(l.startsAt);
	const today = zonedNow(new Date()).date;
	const days = buildDays(today, []).map((d) => ({
		date: d.date,
		label: d.today ? `Hoje, ${d.label}` : d.label
	}));
	const values: LobbyFormValues = {
		instanceId: l.instance.id,
		date,
		time,
		tank: String(l.slots.tank),
		support: String(l.slots.support),
		dps: String(l.slots.dps),
		minLevel: String(l.minLevel),
		characterId: l.owner.characterId ?? '',
		note: l.note ?? ''
	};
	const owner = characters.ok ? characters.data.filter((c) => c.id === l.owner.characterId) : [];
	return {
		lobby: l,
		characters: owner,
		classes: classes.ok ? classes.data : [],
		days,
		values,
		loginHref: loginHref(new URL(path, 'http://web'))
	};
};

export const actions: Actions = {
	update: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}/editar`;
		const call = sessionCall(cookies, fetch, path);
		const values = readLobbyForm(await request.formData());
		const parsed = toLobbyUpdate(values);
		if (!parsed.ok)
			return fail(422, { values, errors: lobbyFieldMessages(parsed.fields), message: null });

		const result = await updateLobby(call, params.id, parsed.input);
		if (result.ok) redirect(303, `/lobbies/${params.id}`);
		switch (result.kind) {
			case 'no_session':
				return toLogin(cookies, path);
			case 'invalid':
				return fail(422, { values, errors: lobbyFieldMessages(result.fields), message: null });
			case 'not_found':
				error(404, 'Lobby não encontrado');
				break;
			case 'conflict':
				return fail(409, {
					values,
					errors: {},
					message: lobbyConflictMessage(result.code) ?? UNAVAILABLE_MESSAGE
				});
		}
		return fail(503, { values, errors: {}, message: UNAVAILABLE_MESSAGE });
	}
};
