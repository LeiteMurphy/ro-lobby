import { authConfig } from '$lib/auth/config';
import { loginHref } from '$lib/auth/display';
import { SESSION_COOKIE } from '$lib/auth/oauth';
import { listClasses } from '$lib/characters/api';
import { UNAVAILABLE_MESSAGE } from '$lib/characters/messages';
import { listInstances } from '$lib/lobbies/api';
import { listTalents, type CatalogFilter, type Role } from '$lib/talents/api';
import { CLOCK_OPTIONS } from '$lib/talents/format';
import type { PageServerLoad } from './$types';

// Catálogo do banco de talentos (spec banco-de-talentos, RN-11, RN-12, D-05). Público:
// os filtros vêm da query string (funciona sem JS), e a sessão, se houver, vai para a API
// mandar o nome no Discord.

/** Limite de uma resposta da API (risco do design). */
const MAX_TALENTS = 100;

const ROLES: readonly Role[] = ['tank', 'support', 'dps'];

interface TalentFilterValues {
	instancia: string;
	funcao: string;
	dia: string;
	hora: string;
}

/** Os filtros da URL; valor fora da lista é ignorado. */
function readFilters(url: URL): { values: TalentFilterValues; filter: CatalogFilter } {
	const q = url.searchParams;
	const values: TalentFilterValues = {
		instancia: q.get('instancia') ?? '',
		funcao: q.get('funcao') ?? '',
		dia: q.get('dia') ?? '',
		hora: q.get('hora') ?? ''
	};
	const filter: CatalogFilter = {};
	if (values.instancia) filter.instanceId = values.instancia;
	if (ROLES.includes(values.funcao as Role)) filter.role = values.funcao as Role;
	else values.funcao = '';
	if (/^[0-6]$/.test(values.dia)) filter.day = Number(values.dia);
	else values.dia = '';
	if (CLOCK_OPTIONS.includes(values.hora)) filter.time = values.hora;
	else values.hora = '';
	return { values, filter };
}

export const load: PageServerLoad = async ({ url, locals, cookies, fetch }) => {
	const { apiBaseUrl } = authConfig();
	const token = locals.user ? (cookies.get(SESSION_COOKIE) ?? '') : '';
	const { values, filter } = readFilters(url);
	const [talents, instances, classes] = await Promise.all([
		listTalents({ fetchFn: fetch, apiBaseUrl, token }, filter),
		listInstances(fetch, apiBaseUrl),
		listClasses(fetch, apiBaseUrl)
	]);
	const list = talents.ok ? talents.data : [];
	return {
		talents: list,
		limited: list.length >= MAX_TALENTS,
		filters: values,
		instances: instances.ok ? instances.data : [],
		classNames: Object.fromEntries(
			(classes.ok ? classes.data : []).map((c) => [c.id, c.name] as const)
		),
		loadError: talents.ok
			? null
			: talents.kind === 'invalid'
				? 'Algum filtro não vale. Limpe os filtros e tente de novo.'
				: UNAVAILABLE_MESSAGE,
		loginHref: loginHref(new URL(`${url.pathname}${url.search}`, 'http://web'))
	};
};
