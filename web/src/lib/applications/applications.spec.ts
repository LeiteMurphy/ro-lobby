import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { request } from '$lib/api/request';
import type { Character } from '$lib/characters/api';
import { TEMPLE } from '$lib/lobbies/fixtures';
import type { ApiLobby } from '$lib/lobbies/api';
import ApplyDialog from './components/ApplyDialog.svelte';
import PlayerPanel from './components/PlayerPanel.svelte';
import RejectDialog from './components/RejectDialog.svelte';
import { applyState, composition, eligibility, HOST_KEY, people } from './detail';
import { applicationFieldMessage, ruleMessage } from './messages';

// Modelo, mensagens e componentes da candidatura no detalhe do lobby (spec
// candidatura-lobby, T-05).

const BRASA = {
	applicationId: 'a1',
	userId: 'u-bia',
	discordName: null,
	characterId: 'c-brasa',
	nick: 'Brasa',
	classId: 'guardiao-real',
	level: 200,
	portrait: 'retrato-3',
	link: 'https://ragnaplace.com/brasa',
	role: 'tank',
	message: null,
	createdAt: '2026-10-06T20:10:00Z'
} as const;
const FOGO = {
	...BRASA,
	applicationId: 'a2',
	userId: 'u-caio',
	characterId: 'c-fogo',
	nick: 'Fogo',
	classId: 'arquimago',
	role: 'dps',
	link: null,
	discordName: 'Caio',
	message: 'tenho buff de ASPD'
} as const;

const WITH_MEMBER: ApiLobby = {
	...TEMPLE,
	occupied: { tank: 1, support: 1, dps: 0 },
	members: [BRASA],
	pendingCount: 1
};
const AS_OWNER: ApiLobby = { ...WITH_MEMBER, pending: [FOGO] };

const character = (over: Partial<Character>): Character => ({
	id: 'c',
	nick: 'X',
	classId: 'arcebispo',
	level: 200,
	role: 'support',
	portrait: 'retrato-1',
	link: null,
	isMain: false,
	createdAt: '',
	...over
});

describe('modelo do detalhe', () => {
	it('RN-31: anfitrião, membros e, para o dono, candidatos', () => {
		expect(people(WITH_MEMBER).map((p) => [p.key, p.kind, p.nick])).toEqual([
			[HOST_KEY, 'host', 'Lirien'],
			['a1', 'member', 'Brasa']
		]);
		expect(people(AS_OWNER).map((p) => p.kind)).toEqual(['host', 'member', 'candidate']);
		const host = people(WITH_MEMBER)[0];
		expect(host.discordName).toBe('Grimbold');
		expect(host.role).toBe('support');
	});

	it('D-03 / CA-02.1: o membro aceito ocupa a vaga da função dele; candidato não', () => {
		const rows = composition(AS_OWNER, people(AS_OWNER));
		const tank = rows.find((r) => r.role === 'tank')!;
		expect(tank.slots.map((s) => s?.nick ?? null)).toEqual(['Brasa']);
		expect(tank.filled).toBe(1);
		const dps = rows.find((r) => r.role === 'dps')!;
		expect(dps.slots).toEqual([null, null, null]);
		const support = rows.find((r) => r.role === 'support')!;
		expect(support.slots.map((s) => s?.kind ?? null)).toEqual(['host', null]);
	});

	it('CA-01.7 / CA-01.11 / RN-05 / RN-30: só pode quem tem nível e vaga; os outros dizem por quê', () => {
		const list = eligibility(WITH_MEMBER, [
			character({ id: 'ok', role: 'support', level: 200 }),
			character({ id: 'low', role: 'dps', level: 159 }),
			character({ id: 'full', role: 'tank', level: 250 }),
			character({ id: 'dps', role: 'dps', level: 160 })
		]);
		expect(list.map((e) => [e.character.id, e.ok, e.why])).toEqual([
			['ok', true, '1 vaga de Suporte'],
			['low', false, 'abaixo do nível 160'],
			['full', false, 'Tank sem vaga'],
			['dps', true, '3 vagas de Dano']
		]);
	});

	it('RN-02 / RN-03 / RN-07 / RN-13: o estado de quem olha', () => {
		const mine = (status: string) =>
			({
				...WITH_MEMBER,
				myApplication: {
					id: 'm',
					characterId: null,
					role: 'dps',
					message: null,
					status,
					reason: null,
					createdAt: '',
					decidedAt: null
				}
			}) as ApiLobby;
		expect(applyState(WITH_MEMBER, null)).toBe('login');
		expect(applyState(WITH_MEMBER, TEMPLE.owner.userId)).toBe('owner');
		expect(applyState(WITH_MEMBER, 'u-duda')).toBe('apply');
		expect(applyState(mine('pending'), 'u-duda')).toBe('pending');
		expect(applyState(mine('accepted'), 'u-duda')).toBe('member');
		expect(applyState(mine('rejected'), 'u-duda')).toBe('rejected');
		expect(applyState(mine('withdrawn'), 'u-duda')).toBe('apply');
		expect(applyState({ ...WITH_MEMBER, status: 'started' }, 'u-duda')).toBe('closed');
		expect(applyState({ ...WITH_MEMBER, status: 'cancelled' }, null)).toBe('closed');
	});
});

describe('mensagens', () => {
	it('D-07: cada regra da API tem mensagem em pt-BR', () => {
		for (const code of [
			'not_open',
			'own_lobby',
			'already_active',
			'role_full',
			'rejected_before',
			'below_min_level',
			'schedule_conflict',
			'not_pending',
			'not_owner',
			'not_yours'
		] as const) {
			expect(ruleMessage(code)).toMatch(/\S/);
		}
		expect(ruleMessage('schedule_conflict')).toContain('2 h');
		expect(ruleMessage(undefined)).toBeNull();
	});

	it('CA-01.3 / CA-02.3 / CA-02.4: campos da candidatura e da recusa', () => {
		expect(applicationFieldMessage({ field: 'message', code: 'too_long' })).toBe(
			'Use até 250 caracteres'
		);
		expect(applicationFieldMessage({ field: 'reason', code: 'required' })).toBe(
			'Escreva a justificativa'
		);
		expect(applicationFieldMessage({ field: 'reason', code: 'too_short' })).toBe(
			'Escreva de 10 a 250 caracteres'
		);
	});

	it('D-07: o 409 application_rule chega com o código da regra', async () => {
		const fetchFn = (async () =>
			new Response(JSON.stringify({ error: 'application_rule', code: 'role_full' }), {
				status: 409
			})) as typeof fetch;
		expect(await request({ fetchFn, apiBaseUrl: 'http://api', token: 't' }, 'POST', '/x')).toEqual({
			ok: false,
			kind: 'conflict',
			code: undefined,
			rule: 'role_full'
		});
	});
});

describe('painel do jogador', () => {
	const panel = (lobby: ApiLobby, key: string, canDecide: boolean) =>
		render(PlayerPanel, {
			props: {
				person: people(lobby).find((p) => p.key === key)!,
				classNames: { 'guardiao-real': 'Guardião Real', arquimago: 'Arquimago' },
				canDecide,
				onreject: () => {}
			}
		}).body;

	it('CA-10.1 / CA-10.2: membro para o visitante, com link e sem Discord', () => {
		const html = panel(WITH_MEMBER, 'a1', false);
		expect(html).toContain('Membro');
		expect(html).toContain('Brasa');
		expect(html).toContain('Guardião Real · Nv 200');
		expect(html).toContain('/portraits/retrato-3.svg');
		expect(html).toContain('href="https://ragnaplace.com/brasa"');
		expect(html).toContain('O Discord aparece para quem está no grupo.');
		expect(html).not.toContain('Aceitar');
	});

	it('RN-32: o anfitrião sem Discord (visitante) mostra o aviso; com Discord (grupo), o nome', () => {
		const visitor: ApiLobby = {
			...WITH_MEMBER,
			owner: { ...WITH_MEMBER.owner, discordName: null }
		};
		const hidden = panel(visitor, HOST_KEY, false);
		expect(hidden).toContain('Anfitrião');
		expect(hidden).toContain('O Discord aparece para quem está no grupo.');
		const shown = panel(WITH_MEMBER, HOST_KEY, false);
		expect(shown).toMatch(/Discord<\/span>\s*<b[^>]*>Grimbold<\/b>/);
	});

	it('CA-10.3: candidato para o dono, com Discord, mensagem, Aceitar e Recusar', () => {
		const html = panel(AS_OWNER, 'a2', true);
		expect(html).toContain('Candidato');
		expect(html.replace(/<!--[\s\S]*?-->/g, '')).toContain('Dano · pendente desde 17:10');
		expect(html).toMatch(/Discord<\/span>\s*<b[^>]*>Caio<\/b>/);
		expect(html).toContain('tenho buff de ASPD');
		expect(html).toMatch(/action="\?\/accept"[\s\S]*name="applicationId" value="a2"/);
		expect(html).toContain('Aceitar');
		expect(html).toContain('Recusar');
		expect(html).not.toContain('O Discord aparece');
	});
});

describe('diálogos', () => {
	it('CA-01.1 / RN-30: Candidatar lista os personagens, desabilita os que não podem e diz por quê', () => {
		const html = render(ApplyDialog, {
			props: {
				title: 'Templo do Demônio Rei',
				options: eligibility(WITH_MEMBER, [
					character({ id: 'ok', nick: 'Lirien', role: 'support' }),
					character({ id: 'low', nick: 'Baixo', role: 'dps', level: 100 })
				]),
				classNames: { arcebispo: 'Arcebispo' },
				onclose: () => {}
			}
		}).body;
		expect(html).toContain('action="?/apply"');
		expect(html).toMatch(/value="ok"[^>]*checked/);
		expect(html).toMatch(/value="low"[^>]*disabled/);
		expect(html).toContain('abaixo do nível 160');
		expect(html).toContain('Arcebispo · Nv 200');
		expect(html).toContain('Mensagem para o anfitrião (opcional)');
		expect(html).toContain('maxlength="250"');
	});

	it('CA-02.3: Recusar volta com o erro ligado ao campo da justificativa', () => {
		const html = render(RejectDialog, {
			props: {
				applicationId: 'a2',
				nick: 'Fogo',
				reason: 'curta',
				error: 'Escreva de 10 a 250 caracteres',
				onclose: () => {}
			}
		}).body;
		expect(html).toContain('Recusar Fogo?');
		expect(html).toContain('action="?/reject"');
		expect(html).toContain('value="a2"');
		expect(html).toMatch(/aria-invalid="true"[^>]*aria-describedby="([^"]+-reason-err)"/);
		expect(html).toContain('Escreva de 10 a 250 caracteres');
	});
});
