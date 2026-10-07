import { error, fail } from '@sveltejs/kit';
import {
	acceptApplication,
	applyToLobby,
	leaveLobby,
	rejectApplication,
	requestSwap,
	withdrawApplication,
	withdrawSwap
} from '$lib/applications/api';
import { applicationFieldMessages, ruleMessage, swapRuleMessage } from '$lib/applications/messages';
import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import { SESSION_COOKIE } from '$lib/auth/oauth';
import { sessionCall, toLogin } from '$lib/auth/session';
import { listCharacters, listClasses, type Character } from '$lib/characters/api';
import { UNAVAILABLE_MESSAGE } from '$lib/characters/messages';
import { cancelLobby, getLobby } from '$lib/lobbies/api';
import { lobbyConflictMessage, lobbyFieldMessages } from '$lib/lobbies/messages';
import type { Result } from '$lib/api/request';
import type { Actions, PageServerLoad } from './$types';

// Detalhe do lobby conforme quem olha, cancelamento pelo dono e candidatura: candidatar,
// aceitar, recusar e retirar; sair do grupo e pedir ou retirar troca (specs lobbies RN-15,
// RN-16, RN-19 e candidatura-lobby RN-01 a RN-14, RN-20, RN-27, RN-28, RN-31, RN-32).

export const load: PageServerLoad = async ({ params, locals, cookies, fetch }) => {
	const { apiBaseUrl } = authConfig();
	const token = locals.user ? (cookies.get(SESSION_COOKIE) ?? '') : '';
	const [lobby, classes] = await Promise.all([
		getLobby(fetch, apiBaseUrl, params.id, token),
		listClasses(fetch, apiBaseUrl)
	]);
	if (!lobby.ok && lobby.kind === 'not_found') error(404, 'Lobby não encontrado');
	if (!lobby.ok) error(503, UNAVAILABLE_MESSAGE);
	const isOwner = locals.user?.id === lobby.data.owner.userId;

	// O diálogo de candidatura lista os personagens de quem pode se candidatar.
	let characters: Character[] = [];
	if (token && !isOwner && lobby.data.status === 'open') {
		const list = await listCharacters({ fetchFn: fetch, apiBaseUrl, token });
		if (list.ok) characters = list.data;
	}
	const classNames = Object.fromEntries(
		(classes.ok ? classes.data : []).map((c) => [c.id, c.name] as const)
	);
	const classId = lobby.data.owner.classId;
	return {
		lobby: lobby.data,
		classNames,
		characters,
		ownerClass: classId ? (classNames[classId] ?? classId) : null,
		// RN-16: as ações do dono aparecem só para ele.
		isOwner,
		now: new Date().toISOString(),
		loginHref: loginHref(new URL(`/lobbies/${params.id}`, 'http://web'))
	};
};

type Failure = Exclude<Result<unknown>, { ok: true }>;

/**
 * 409 de candidatura, 404 e indisponível, comuns às ações de candidatura; `swap` fala do
 * pedido de troca em vez da candidatura.
 */
function applicationFailure<A extends string>(
	action: A,
	result: Failure,
	extra: object = {},
	subject: 'application' | 'swap' = 'application'
) {
	const rules = subject === 'swap' ? swapRuleMessage : ruleMessage;
	if (result.kind === 'conflict') {
		return fail(409, {
			action,
			...extra,
			errors: {} as Record<string, string>,
			message: rules(result.rule) ?? lobbyConflictMessage(result.code) ?? UNAVAILABLE_MESSAGE
		});
	}
	if (result.kind === 'not_found') {
		return fail(404, {
			action,
			...extra,
			errors: {} as Record<string, string>,
			message:
				subject === 'swap'
					? 'Esse pedido de troca não existe mais.'
					: 'Essa candidatura ou esse lobby não existe mais.'
		});
	}
	return fail(503, {
		action,
		...extra,
		errors: {} as Record<string, string>,
		message: UNAVAILABLE_MESSAGE
	});
}

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
				return fail(422, {
					action: 'cancel' as const,
					reason,
					errors: lobbyFieldMessages(result.fields),
					message: null
				});
			case 'not_found':
				error(404, 'Lobby não encontrado');
				break;
			case 'conflict':
				return fail(409, {
					action: 'cancel' as const,
					reason,
					errors: {},
					message: lobbyConflictMessage(result.code) ?? UNAVAILABLE_MESSAGE
				});
		}
		return fail(503, {
			action: 'cancel' as const,
			reason,
			errors: {},
			message: UNAVAILABLE_MESSAGE
		});
	},

	// RN-01 a RN-07, RN-30: candidatura com personagem e mensagem opcional.
	apply: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const form = await request.formData();
		const characterId = String(form.get('characterId') ?? '');
		const message = String(form.get('message') ?? '').trim();
		const result = await applyToLobby(call, params.id, characterId, message);
		if (result.ok) return { done: 'apply' as const };
		if (result.kind === 'no_session') return toLogin(cookies, path);
		if (result.kind === 'invalid') {
			return fail(422, {
				action: 'apply' as const,
				characterId,
				message: null,
				text: message,
				errors: applicationFieldMessages(result.fields)
			});
		}
		return applicationFailure('apply' as const, result, { characterId, text: message });
	},

	// RN-08, RN-10 a RN-12: o dono aceita.
	accept: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const applicationId = String((await request.formData()).get('applicationId') ?? '');
		const result = await acceptApplication(call, applicationId);
		if (result.ok) return { done: 'accept' as const };
		if (result.kind === 'no_session') return toLogin(cookies, path);
		return applicationFailure('accept' as const, result, { applicationId });
	},

	// RN-08, RN-09: o dono recusa com justificativa.
	reject: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const form = await request.formData();
		const applicationId = String(form.get('applicationId') ?? '');
		const reason = String(form.get('reason') ?? '');
		const result = await rejectApplication(call, applicationId, reason);
		if (result.ok) return { done: 'reject' as const };
		if (result.kind === 'no_session') return toLogin(cookies, path);
		if (result.kind === 'invalid') {
			return fail(422, {
				action: 'reject' as const,
				applicationId,
				reason,
				errors: applicationFieldMessages(result.fields),
				message: null
			});
		}
		return applicationFailure('reject' as const, result, { applicationId, reason });
	},

	// RN-13: o candidato retira a própria pendente.
	withdraw: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const applicationId = String((await request.formData()).get('applicationId') ?? '');
		const result = await withdrawApplication(call, applicationId);
		if (result.ok) return { done: 'withdraw' as const };
		if (result.kind === 'no_session') return toLogin(cookies, path);
		return applicationFailure('withdraw' as const, result, { applicationId });
	},

	// RN-14: o membro sai do grupo.
	leave: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const applicationId = String((await request.formData()).get('applicationId') ?? '');
		const result = await leaveLobby(call, applicationId);
		if (result.ok) return { done: 'leave' as const };
		if (result.kind === 'no_session') return toLogin(cookies, path);
		return applicationFailure('leave' as const, result, { applicationId });
	},

	// RN-20, RN-36: o membro pede a troca do personagem, com motivo.
	requestSwap: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const form = await request.formData();
		const applicationId = String(form.get('applicationId') ?? '');
		const characterId = String(form.get('characterId') ?? '');
		const reason = String(form.get('reason') ?? '');
		const result = await requestSwap(call, applicationId, characterId, reason);
		if (result.ok) return { done: 'requestSwap' as const };
		if (result.kind === 'no_session') return toLogin(cookies, path);
		if (result.kind === 'invalid') {
			return fail(422, {
				action: 'requestSwap' as const,
				characterId,
				reason,
				errors: applicationFieldMessages(result.fields),
				message: null
			});
		}
		return applicationFailure('requestSwap' as const, result, { characterId, reason });
	},

	// RN-27: o membro retira o próprio pedido de troca.
	withdrawSwap: async ({ params, request, cookies, fetch }) => {
		const path = `/lobbies/${params.id}`;
		const call = sessionCall(cookies, fetch, path);
		const swapId = String((await request.formData()).get('swapId') ?? '');
		const result = await withdrawSwap(call, swapId);
		if (result.ok) return { done: 'withdrawSwap' as const };
		if (result.kind === 'no_session') return toLogin(cookies, path);
		return applicationFailure('withdrawSwap' as const, result, { swapId }, 'swap');
	}
};
