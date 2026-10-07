import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import type { Character } from '$lib/characters/api';
import { TEMPLE } from '$lib/lobbies/fixtures';
import type { ApiLobby } from '$lib/lobbies/api';
import LeaveDialog from './components/LeaveDialog.svelte';
import SwapDialog from './components/SwapDialog.svelte';
import { applyState, swapEligibility } from './detail';
import { applicationFieldMessage, ruleMessage, swapRuleMessage } from './messages';

// Modelo, mensagens e diálogos da Parte 2 do lado do membro: sair, pedir troca e bloqueio
// (spec candidatura-lobby, T-14).

// Tank lotado, 1 Suporte e 3 Danos livres; nível mínimo 160.
const LOBBY: ApiLobby = { ...TEMPLE, occupied: { tank: 1, support: 1, dps: 0 } };

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
const FOGO = character({ id: 'fogo', nick: 'Fogo', role: 'dps' });
const BRISA = character({ id: 'brisa', nick: 'Brisa', role: 'tank', level: 246 });
const CURA = character({ id: 'cura', nick: 'Cura', role: 'support' });
const FARO = character({ id: 'faro', nick: 'Faro', role: 'dps', level: 150 });
const LAMINA = character({ id: 'lamina', nick: 'Lâmina', role: 'dps' });

const viewer = (status: string, over: Record<string, unknown> = {}) =>
	({
		...LOBBY,
		myApplication: {
			id: 'm1',
			characterId: 'fogo',
			role: 'dps',
			message: null,
			status,
			reason: null,
			blocked: false,
			createdAt: '',
			decidedAt: null,
			swapRequest: null,
			...over
		}
	}) as ApiLobby;

describe('troca de personagem: quem dá para escolher', () => {
	const current = { characterId: 'fogo', role: 'dps' as const };

	it('CA-08.1 / CA-08.13 / RN-20 / RN-36 / D-11: o pedido tira o atual, barra o nível e pode esperar a vaga', () => {
		const list = swapEligibility(LOBBY, [FOGO, BRISA, CURA, FARO, LAMINA], current, 'request');
		expect(list.map((o) => [o.character.id, o.ok, o.why])).toEqual([
			['brisa', true, 'Tank sem vaga agora'],
			['cura', true, '1 vaga de Suporte'],
			['faro', false, 'abaixo do nível 160'],
			['lamina', true, 'mesma vaga de Dano']
		]);
	});

	it('CA-07.2 / CA-07.3 / CA-07.6 / RN-19: na troca do dono, função sem vaga fica de fora', () => {
		const owner = { characterId: TEMPLE.owner.characterId, role: 'support' as const };
		const full: ApiLobby = { ...LOBBY, occupied: { tank: 1, support: 2, dps: 0 } };
		const list = swapEligibility(full, [BRISA, CURA, FARO], owner, 'owner');
		expect(list.map((o) => [o.character.id, o.ok, o.why])).toEqual([
			['brisa', false, 'Tank sem vaga'],
			['cura', true, 'mesma vaga de Suporte'],
			['faro', false, 'abaixo do nível 160']
		]);
	});
});

describe('estado de quem olha (Parte 2)', () => {
	it('CA-05.4 / CA-06.5 / CA-06.6 / RN-15 / RN-37: quem saiu ou foi removido sem bloqueio volta; com bloqueio, não', () => {
		expect(applyState(viewer('left'), 'u-bia')).toBe('apply');
		expect(applyState(viewer('removed'), 'u-bia')).toBe('apply');
		expect(applyState(viewer('removed', { blocked: true }), 'u-bia')).toBe('blocked');
		expect(
			applyState({ ...viewer('removed', { blocked: true }), status: 'started' }, 'u-bia')
		).toBe('closed');
		expect(applyState(viewer('accepted'), 'u-bia')).toBe('member');
	});
});

describe('mensagens (Parte 2)', () => {
	it('D-13: os códigos novos têm mensagem; o bloqueio usa o texto da RN-38', () => {
		expect(ruleMessage('blocked')).toBe('Você não pode se candidatar a este lobby.');
		expect(ruleMessage('not_member')).toMatch(/\S/);
		expect(ruleMessage('swap_pending')).toContain('pedido de troca pendente');
		expect(ruleMessage('not_owner')).toBe('Só o anfitrião pode decidir.');
	});

	it('CA-08.12: nas ações do pedido, os códigos falam do pedido', () => {
		expect(swapRuleMessage('not_yours')).toBe('Esse pedido de troca não é seu.');
		expect(swapRuleMessage('not_pending')).toBe('Esse pedido de troca não está mais pendente.');
		expect(swapRuleMessage('role_full')).toBe(ruleMessage('role_full'));
		expect(swapRuleMessage(undefined)).toBeNull();
	});

	it('CA-08.2: o motivo do pedido usa a mensagem da justificativa', () => {
		expect(applicationFieldMessage({ field: 'reason', code: 'required' })).toBe(
			'Escreva a justificativa'
		);
		expect(applicationFieldMessage({ field: 'characterId', code: 'invalid' })).toBe(
			'Escolha um dos seus personagens'
		);
	});
});

describe('diálogos do membro', () => {
	it('CA-05.5 / RN-38: Sair do grupo confirma, diz que a vaga fica livre e manda a candidatura', () => {
		const html = render(LeaveDialog, {
			props: {
				applicationId: 'm1',
				title: 'Templo do Demônio Rei · qua, 7 out às 20:00',
				roleLabel: 'Dano',
				onclose: () => {}
			}
		}).body;
		expect(html).toContain('Sair do grupo?');
		expect(html).toContain('action="?/leave"');
		expect(html).toContain('value="m1"');
		expect(html).toContain('Sua vaga de Dano fica livre para outro jogador.');
	});

	it('CA-08.14 / RN-38: Pedir troca lista os personagens com o motivo de cada um e pede o motivo', () => {
		const html = render(SwapDialog, {
			props: {
				mode: 'request',
				hint: 'O anfitrião decide. Você continua no grupo com Fogo até lá.',
				options: swapEligibility(
					LOBBY,
					[FOGO, BRISA, FARO],
					{ characterId: 'fogo', role: 'dps' },
					'request'
				),
				classNames: { arcebispo: 'Arcebispo' },
				applicationId: 'm1',
				onclose: () => {}
			}
		}).body;
		expect(html).toContain('Pedir troca de personagem');
		expect(html).toContain('action="?/requestSwap"');
		expect(html).toMatch(/name="applicationId" value="m1"/);
		expect(html).not.toContain('value="fogo"');
		expect(html).toMatch(/value="brisa"[^>]*checked/);
		expect(html).toMatch(/value="faro"[^>]*disabled/);
		expect(html).toContain('Tank sem vaga agora');
		expect(html).toMatch(/<textarea[^>]*name="reason"/);
		expect(html).toContain('Enviar pedido');
	});

	it('CA-08.2: o erro do motivo volta ligado ao campo', () => {
		const html = render(SwapDialog, {
			props: {
				mode: 'request',
				hint: '',
				options: swapEligibility(LOBBY, [BRISA], { characterId: 'fogo', role: 'dps' }, 'request'),
				classNames: {},
				applicationId: 'm1',
				characterId: 'brisa',
				reason: 'curto',
				errors: { reason: 'Escreva de 10 a 250 caracteres' },
				formError: 'Você já tem um pedido de troca pendente neste lobby.',
				onclose: () => {}
			}
		}).body;
		expect(html).toMatch(/aria-invalid="true"[^>]*aria-describedby="([^"]+-reason-err)"/);
		expect(html).toContain('Escreva de 10 a 250 caracteres');
		expect(html).toContain('role="alert"');
	});
});
