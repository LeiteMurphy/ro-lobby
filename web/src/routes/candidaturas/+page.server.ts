import { error, fail } from '@sveltejs/kit';
import { leaveLobby, listMyApplications, withdrawApplication } from '$lib/applications/api';
import { ruleMessage } from '$lib/applications/messages';
import { SESSION_COOKIE } from '$lib/auth/oauth';
import { sessionCall, toLogin } from '$lib/auth/session';
import { listClasses } from '$lib/characters/api';
import { UNAVAILABLE_MESSAGE } from '$lib/characters/messages';
import type { Actions, PageServerLoad } from './$types';

// "Minhas candidaturas" (spec candidatura-lobby, RN-33, RN-38, CA-03.1, CA-10.4): a lista
// com o estado e a justificativa, retirar a pendente e sair do grupo.

const PATH = '/candidaturas';

export const load: PageServerLoad = async ({ locals, cookies, fetch }) => {
	if (!locals.user || !cookies.get(SESSION_COOKIE)) toLogin(cookies, PATH);
	const call = sessionCall(cookies, fetch, PATH);
	const [mine, classes] = await Promise.all([
		listMyApplications(call),
		listClasses(fetch, call.apiBaseUrl)
	]);
	if (!mine.ok && mine.kind === 'no_session') toLogin(cookies, PATH);
	if (!mine.ok) error(503, UNAVAILABLE_MESSAGE);
	return {
		applications: mine.data,
		classNames: Object.fromEntries(
			(classes.ok ? classes.data : []).map((c) => [c.id, c.name] as const)
		)
	};
};

type Failure = Exclude<Awaited<ReturnType<typeof withdrawApplication>>, { ok: true }>;

function failure(result: Failure) {
	if (result.kind === 'conflict') {
		return fail(409, { message: ruleMessage(result.rule) ?? UNAVAILABLE_MESSAGE });
	}
	if (result.kind === 'not_found') {
		return fail(404, { message: 'Essa candidatura não existe mais.' });
	}
	return fail(503, { message: UNAVAILABLE_MESSAGE });
}

export const actions: Actions = {
	// RN-13: retirar a própria pendente.
	withdraw: async ({ request, cookies, fetch }) => {
		const call = sessionCall(cookies, fetch, PATH);
		const applicationId = String((await request.formData()).get('applicationId') ?? '');
		const result = await withdrawApplication(call, applicationId);
		if (result.ok) return { done: 'withdraw' as const };
		if (result.kind === 'no_session') return toLogin(cookies, PATH);
		return failure(result);
	},

	// RN-14 / RN-38: sair do grupo, também daqui.
	leave: async ({ request, cookies, fetch }) => {
		const call = sessionCall(cookies, fetch, PATH);
		const applicationId = String((await request.formData()).get('applicationId') ?? '');
		const result = await leaveLobby(call, applicationId);
		if (result.ok) return { done: 'leave' as const };
		if (result.kind === 'no_session') return toLogin(cookies, PATH);
		return failure(result);
	}
};
