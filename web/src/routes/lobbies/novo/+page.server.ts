import { fail, redirect } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import { sessionCall, toLogin } from '$lib/auth/session';
import { listCharacters, listClasses } from '$lib/characters/api';
import { UNAVAILABLE_MESSAGE } from '$lib/characters/messages';
import { buildDays } from '$lib/home/days';
import { zonedNow } from '$lib/home/time';
import { createLobby, listInstances } from '$lib/lobbies/api';
import { readLobbyForm, toLobbyInput, type LobbyFormValues } from '$lib/lobbies/form';
import { defaultStart } from '$lib/lobbies/time';
import { lobbyConflictMessage, lobbyFieldMessages } from '$lib/lobbies/messages';
import type { Actions, PageServerLoad } from './$types';

// Criação de lobby (spec lobbies, RN-04 a RN-12, tela 1b).

const PATH = '/lobbies/novo';

export const load: PageServerLoad = async ({ locals, cookies, fetch, url }) => {
	// RN-24: o dia escolhido na Home vem em ?dia= e sobrevive ao login.
	const requested = url.searchParams.get('dia');
	const back = requested ? `${PATH}?${new URLSearchParams({ dia: requested })}` : PATH;
	if (!locals.user) toLogin(cookies, back);
	const call = sessionCall(cookies, fetch, back);
	const { apiBaseUrl } = authConfig();
	const [characters, instances, classes] = await Promise.all([
		listCharacters(call),
		listInstances(fetch, apiBaseUrl),
		listClasses(fetch, apiBaseUrl)
	]);
	if (!characters.ok && characters.kind === 'no_session') toLogin(cookies, back);

	const now = zonedNow(new Date());
	const today = now.date;
	const days = buildDays(today, []).map((d) => ({
		date: d.date,
		label: d.today ? `Hoje, ${d.label}` : d.label
	}));
	const mine = characters.ok ? characters.data : [];
	const main = mine.find((c) => c.isMain) ?? mine[0];
	// A instância inicial é a mais alta que o personagem principal alcança (RN-02, RN-08).
	const catalog = instances.ok ? instances.data : [];
	const firstInstance = catalog.find((i) => main && i.level <= main.level) ?? catalog[0];
	const values: LobbyFormValues = {
		instanceId: firstInstance?.id ?? '',
		...defaultStart(
			requested,
			now,
			days.map((d) => d.date)
		),
		tank: '1',
		support: '2',
		dps: '3',
		minLevel: String(firstInstance?.level ?? 1),
		characterId: main?.id ?? '',
		note: '',
		// spec grupo-livre, RN-01 e RN-02: por função; o grupo livre começa com 12 vagas.
		formation: 'roles',
		freeSlots: '12',
		title: ''
	};
	return {
		characters: mine,
		instances: instances.ok ? instances.data : [],
		classes: classes.ok ? classes.data : [],
		days,
		values,
		loadError: characters.ok && instances.ok ? null : UNAVAILABLE_MESSAGE,
		loginHref: loginHref(new URL(back, 'http://web'))
	};
};

export const actions: Actions = {
	create: async ({ request, cookies, fetch }) => {
		const call = sessionCall(cookies, fetch, PATH);
		const values = readLobbyForm(await request.formData());
		const parsed = toLobbyInput(values);
		if (!parsed.ok)
			return fail(422, { values, errors: lobbyFieldMessages(parsed.fields), message: null });

		const result = await createLobby(call, parsed.input);
		if (result.ok) redirect(303, `/lobbies/${result.data.id}`);
		switch (result.kind) {
			case 'no_session':
				return toLogin(cookies, PATH);
			case 'invalid':
				return fail(422, { values, errors: lobbyFieldMessages(result.fields), message: null });
			case 'conflict':
				return fail(409, {
					values,
					errors: {},
					message: lobbyConflictMessage(result.code) ?? UNAVAILABLE_MESSAGE
				});
			default:
				return fail(503, { values, errors: {}, message: UNAVAILABLE_MESSAGE });
		}
	}
};
