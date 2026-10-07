import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import type { Character } from '$lib/characters/api';
import { TEMPLE } from '$lib/lobbies/fixtures';
import type { ApiLobby } from '$lib/lobbies/api';
import type { LobbySwapRequest } from './api';
import PlayerPanel from './components/PlayerPanel.svelte';
import RejectDialog from './components/RejectDialog.svelte';
import RemoveDialog from './components/RemoveDialog.svelte';
import SwapDialog from './components/SwapDialog.svelte';
import SwapPanel from './components/SwapPanel.svelte';
import { people, swapEligibility } from './detail';

// Componentes da Parte 2 do lado do dono: remover, trocar o próprio personagem e decidir
// pedidos de troca (spec candidatura-lobby, T-15).

const noText = (html: string) => html.replace(/<!--[\s\S]*?-->/g, '');

const FAISCA = {
	applicationId: 'a1',
	userId: 'u-bia',
	discordName: 'faisca.ro',
	characterId: 'c-faisca',
	nick: 'Faísca',
	classId: 'elementalista',
	level: 255,
	portrait: 'retrato-1',
	link: null,
	role: 'dps',
	message: null,
	createdAt: '2026-10-06T17:30:00Z'
} as const;
const LOBBY: ApiLobby = { ...TEMPLE, occupied: { tank: 1, support: 1, dps: 1 }, members: [FAISCA] };
const [HOST, MEMBER] = people(LOBBY);

const SWAP: LobbySwapRequest = {
	id: 's1',
	applicationId: 'a1',
	userId: 'u-bia',
	discordName: 'faisca.ro',
	from: {
		characterId: 'c-faisca',
		nick: 'Faísca',
		classId: 'elementalista',
		level: 255,
		portrait: 'retrato-1',
		role: 'dps'
	},
	to: {
		characterId: 'c-brisa',
		nick: 'Brisa',
		classId: 'guardiao-real',
		level: 246,
		portrait: 'retrato-3',
		role: 'tank'
	},
	reason: 'Brasa avisou que talvez não consiga ir; posso cobrir de tank.',
	createdAt: '2026-10-06T18:00:00Z'
};

const panel = (person: typeof HOST, canDecide: boolean) =>
	noText(
		render(PlayerPanel, {
			props: {
				person,
				classNames: {},
				canDecide,
				onreject: () => {},
				onremove: () => {},
				onswap: () => {}
			}
		}).body
	);

describe('painel do dono', () => {
	it('CA-06.8 / RN-15 / RN-38: com um membro escolhido, o dono tem Remover do grupo', () => {
		expect(panel(MEMBER, true)).toContain('Remover do grupo');
		expect(panel(MEMBER, false)).not.toContain('Remover do grupo');
	});

	it('CA-07.1 / RN-19 / RN-38: com o próprio card escolhido, o dono tem Trocar personagem', () => {
		const html = panel(HOST, true);
		expect(html).toMatch(/Anfitrião\s\(você\)/);
		expect(html).toContain('Trocar personagem');
		expect(html).not.toContain('Remover do grupo');
		// Quem não é dono, ou com o lobby fechado, não troca.
		expect(panel(HOST, false)).not.toContain('Trocar personagem');
	});

	it('CA-08.14 / RN-22 / D-11: o pedido escolhido mostra atual → novo, o motivo, a vaga e as decisões', () => {
		const html = noText(
			render(SwapPanel, { props: { swap: SWAP, lobby: LOBBY, onreject: () => {} } }).body
		);
		expect(html).toContain('Pedido de troca');
		expect(html).toMatch(/Faísca[\s\S]*?Dano · Nv 255[\s\S]*?→[\s\S]*?Brisa[\s\S]*?Tank · Nv 246/);
		expect(html).toContain('faisca.ro');
		expect(html).toMatch(/Vaga de Tank<\/span>\s*<b[^>]*>0 livres<\/b>/);
		expect(html).toContain(SWAP.reason);
		expect(html).toMatch(/action="\?\/acceptSwap"[\s\S]*?name="swapId" value="s1"[\s\S]*?Aceitar/);
		expect(html).toContain('Recusar');
		expect(html).toContain('o aceite falha e o pedido continua pendente');

		// Com vaga, o aviso de falha some.
		const free = noText(
			render(SwapPanel, {
				props: {
					swap: SWAP,
					lobby: { ...LOBBY, occupied: { tank: 0, support: 1, dps: 1 } },
					onreject: () => {}
				}
			}).body
		);
		expect(free).toMatch(/Vaga de Tank<\/span>\s*<b[^>]*>1 livre<\/b>/);
		expect(free).not.toContain('o aceite falha');
	});
});

describe('diálogos do dono', () => {
	it('CA-06.1 / CA-06.6 / RN-15: Remover pede justificativa e oferece o bloqueio só neste lobby', () => {
		const html = render(RemoveDialog, {
			props: { applicationId: 'a1', nick: 'Faísca', onclose: () => {} }
		}).body;
		expect(html).toContain('Remover Faísca?');
		expect(html).toContain('action="?/remove"');
		expect(html).toContain('value="a1"');
		expect(html).toMatch(/<textarea[^>]*name="reason"[^>]*maxlength="250"/);
		expect(html).toMatch(/<input type="checkbox" name="block"/);
		expect(html).toContain('Bloquear neste lobby');
		expect(html).toMatch(/Os seus\s+outros lobbies não mudam\./);
	});

	it('CA-06.2: a justificativa vazia volta ligada ao campo, com o bloqueio marcado mantido', () => {
		const html = render(RemoveDialog, {
			props: {
				applicationId: 'a1',
				nick: 'Faísca',
				block: true,
				error: 'Escreva a justificativa',
				onclose: () => {}
			}
		}).body;
		expect(html).toMatch(/aria-invalid="true"[^>]*aria-describedby="([^"]+-reason-err)"/);
		expect(html).toMatch(/name="block"[^>]*checked/);
	});

	it('CA-08.8 / CA-08.9 / RN-22: Recusar a troca manda o pedido para a ação de troca', () => {
		const html = render(RejectDialog, {
			props: { swapId: 's1', nick: 'Faísca', onclose: () => {} }
		}).body;
		expect(html).toContain('Recusar a troca de Faísca?');
		expect(html).toContain('action="?/rejectSwap"');
		expect(html).toMatch(/name="swapId" value="s1"/);
		expect(html).not.toContain('name="applicationId"');
		expect(html).toContain('O membro continua no grupo com o personagem atual.');
	});

	it('CA-07.1 / CA-07.3 / CA-07.6: Trocar seu personagem lista os do dono, sem motivo, com quem não pode desabilitado', () => {
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
		const html = render(SwapDialog, {
			props: {
				mode: 'owner',
				hint: 'Templo do Demônio Rei · nível mínimo 160 · sem aprovação',
				options: swapEligibility(
					LOBBY,
					[
						character({ id: TEMPLE.owner.characterId!, nick: 'Lirien' }),
						character({ id: 'garoa', nick: 'Garoa' }),
						character({ id: 'brisa', nick: 'Brisa', role: 'tank' })
					],
					{ characterId: TEMPLE.owner.characterId, role: 'support' },
					'owner'
				),
				classNames: {},
				onclose: () => {}
			}
		}).body;
		expect(html).toContain('Trocar seu personagem');
		expect(html).toContain('action="?/ownerSwap"');
		expect(html).not.toContain(`value="${TEMPLE.owner.characterId}"`);
		expect(html).toMatch(/value="garoa"[^>]*checked/);
		expect(html).toMatch(/value="brisa"[^>]*disabled/);
		expect(html).toContain('Tank sem vaga');
		expect(html).not.toContain('name="reason"');
		expect(html).not.toContain('name="applicationId"');
		expect(html).toContain('>Trocar<');
	});
});
