import { fail, redirect, type Cookies } from '@sveltejs/kit';
import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import { SESSION_COOKIE } from '$lib/auth/oauth';
import {
	createCharacter,
	deleteCharacter,
	listCharacters,
	listClasses,
	setMainCharacter,
	updateCharacter,
	type CharacterInput,
	type Portrait,
	type Result,
	type Role
} from '$lib/characters/api';
import {
	fieldMessages,
	LIMIT_MESSAGE,
	NOT_FOUND_MESSAGE,
	UNAVAILABLE_MESSAGE
} from '$lib/characters/messages';
import type { Actions, PageServerLoad } from './$types';

// Página de perfil (spec personagens, D-05). O servidor do web lê o token do cookie
// HttpOnly e fala com a API; o navegador só vê o resultado.

const PROFILE = new URL('http://web/perfil');

/** RN-03: visitante (ou sessão recusada pela API) vai ao login e volta para /perfil. */
function toLogin(cookies: Cookies): never {
	cookies.delete(SESSION_COOKIE, { path: '/' });
	redirect(303, loginHref(PROFILE));
}

export const load: PageServerLoad = async ({ locals, cookies, fetch }) => {
	const token = cookies.get(SESSION_COOKIE);
	if (!locals.user || !token) toLogin(cookies);

	const { apiBaseUrl } = authConfig();
	const [characters, classes] = await Promise.all([
		listCharacters({ fetchFn: fetch, apiBaseUrl, token }),
		listClasses(fetch, apiBaseUrl)
	]);
	if (!characters.ok && characters.kind === 'no_session') toLogin(cookies);

	return {
		characters: characters.ok ? characters.data : [],
		classes: classes.ok ? classes.data : [],
		loadError: characters.ok && classes.ok ? null : UNAVAILABLE_MESSAGE,
		loginHref: loginHref(PROFILE)
	};
};

/** Valores do formulário como o Usuário digitou, para voltar ao diálogo num erro (RN-19). */
export interface FormValues {
	nick: string;
	classId: string;
	level: string;
	role: string;
	portrait: string;
	link: string;
}

function readValues(data: FormData): FormValues {
	const text = (name: string) => String(data.get(name) ?? '');
	return {
		nick: text('nick'),
		classId: text('classId'),
		level: text('level'),
		role: text('role'),
		portrait: text('portrait'),
		link: text('link')
	};
}

function toInput(v: FormValues): CharacterInput {
	const level = Number(v.level);
	return {
		nick: v.nick,
		classId: v.classId,
		// Nível vazio ou não inteiro vai como 0, e a API responde o erro do campo (RNF-04).
		level: v.level.trim() !== '' && Number.isInteger(level) ? level : 0,
		role: v.role as Role,
		...(v.portrait ? { portrait: v.portrait as Portrait } : {}),
		link: v.link
	};
}

type Mode = 'create' | 'update';

/** Traduz a resposta da API para o resultado da action. */
function outcome(
	result: Result<unknown>,
	cookies: Cookies,
	form: { mode: Mode | 'delete' | 'main'; id?: string; values?: FormValues }
) {
	if (result.ok) return { done: form.mode };
	switch (result.kind) {
		case 'no_session':
			return toLogin(cookies);
		case 'invalid':
			return fail(422, { ...form, errors: fieldMessages(result.fields), message: null });
		case 'limit':
			return fail(409, { ...form, errors: {}, message: LIMIT_MESSAGE });
		case 'not_found':
			return fail(404, { ...form, errors: {}, message: NOT_FOUND_MESSAGE });
		default:
			return fail(503, { ...form, errors: {}, message: UNAVAILABLE_MESSAGE });
	}
}

function session(cookies: Cookies, fetchFn: typeof fetch) {
	const token = cookies.get(SESSION_COOKIE);
	if (!token) toLogin(cookies);
	return { fetchFn, apiBaseUrl: authConfig().apiBaseUrl, token };
}

export const actions: Actions = {
	create: async ({ request, cookies, fetch }) => {
		const call = session(cookies, fetch);
		const values = readValues(await request.formData());
		return outcome(await createCharacter(call, toInput(values)), cookies, {
			mode: 'create',
			values
		});
	},
	update: async ({ request, cookies, fetch }) => {
		const call = session(cookies, fetch);
		const data = await request.formData();
		const id = String(data.get('id') ?? '');
		const values = readValues(data);
		return outcome(await updateCharacter(call, id, toInput(values)), cookies, {
			mode: 'update',
			id,
			values
		});
	},
	delete: async ({ request, cookies, fetch }) => {
		const call = session(cookies, fetch);
		const id = String((await request.formData()).get('id') ?? '');
		return outcome(await deleteCharacter(call, id), cookies, { mode: 'delete', id });
	},
	main: async ({ request, cookies, fetch }) => {
		const call = session(cookies, fetch);
		const id = String((await request.formData()).get('id') ?? '');
		return outcome(await setMainCharacter(call, id), cookies, { mode: 'main', id });
	}
};
