import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { eligibility, freeComposition, people, swapEligibility } from '$lib/applications/detail';
import { ruleMessage } from '$lib/applications/messages';
import type { Character } from '$lib/characters/api';
import LeaveDialog from '$lib/applications/components/LeaveDialog.svelte';
import PlayerPanel from '$lib/applications/components/PlayerPanel.svelte';
import SwapDialog from '$lib/applications/components/SwapDialog.svelte';
import SwapPanel from '$lib/applications/components/SwapPanel.svelte';
import LobbyForm from './components/LobbyForm.svelte';
import { TEMPLE } from './fixtures';
import { readLobbyForm, toLobbyInput, type LobbyFormValues } from './form';
import { lobbyFieldMessages } from './messages';
import { isFree, openSeats, roomFor, totalSeats } from './seats';
import type { ApiLobby } from './api';

// Grupo livre no formulário, no detalhe e nos diálogos (spec grupo-livre, T-04).

const FREE: ApiLobby = {
	...TEMPLE,
	formation: 'free',
	freeSlots: 3,
	slots: { tank: 0, support: 0, dps: 0 },
	occupied: { tank: 0, support: 2, dps: 0 }
};
const char = (id: string, role: Character['role'], level: number): Character => ({
	id,
	nick: id,
	classId: 'arcebispo',
	level,
	role,
	portrait: 'retrato-1',
	link: null,
	isMain: false,
	createdAt: '',
	availability: null
});
const VALUES: LobbyFormValues = {
	instanceId: 'templo-do-demonio-rei',
	date: '2026-10-07',
	time: '20:00',
	tank: '1',
	support: '2',
	dps: '3',
	minLevel: '160',
	characterId: 'c1',
	note: '',
	formation: 'roles',
	freeSlots: '12',
	title: ''
};

describe('vagas do grupo livre', () => {
	it('CA-02.1 / CA-02.2 / RN-03 / RN-04: vaga no total para qualquer função', () => {
		expect(isFree(FREE)).toBe(true);
		expect(totalSeats(FREE)).toBe(3);
		expect(openSeats(FREE)).toBe(1);
		expect(roomFor(FREE, 'tank')).toBe(1);
		expect(roomFor(TEMPLE, 'tank')).toBe(1);
		expect(roomFor(TEMPLE, 'support')).toBe(1);
	});
});

describe('formulário com a formação', () => {
	it('CA-01.1 / RN-02: o grupo livre manda a formação, o total e as funções zeradas', () => {
		const data = new FormData();
		for (const [k, v] of Object.entries({ ...VALUES, formation: 'free', freeSlots: '12' }))
			data.set(k, v);
		const parsed = toLobbyInput(readLobbyForm(data));
		expect(parsed.ok && parsed.input).toMatchObject({
			formation: 'free',
			freeSlots: 12,
			slots: { tank: 0, support: 0, dps: 0 }
		});
	});

	it('CA-01.3 / RN-01: sem escolha, a formação é por função', () => {
		const data = new FormData();
		data.set('date', '2026-10-07');
		expect(readLobbyForm(data).formation).toBe('roles');
	});

	const renderForm = (values: LobbyFormValues, errors = {}) =>
		render(LobbyForm, {
			props: {
				mode: 'create',
				instances: [{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei', level: 160 }],
				characters: [char('c1', 'support', 178)],
				classes: [],
				days: [{ date: '2026-10-07', label: 'qua, 7 out' }],
				values,
				errors,
				cancelHref: '/'
			} as never
		}).body;

	it('CA-01.3 / RNF-03: a formação vem marcada "Por função", com as vagas por função', () => {
		const html = renderForm(VALUES);
		expect(html).toMatch(/<input type="radio" name="formation" value="roles"[^>]*checked/);
		expect(html).toContain('name="tank"');
		expect(html).not.toContain('name="freeSlots"');
	});

	it('CA-01.1 / RN-02: no grupo livre, um campo de vagas com 12 e a prévia livre', () => {
		const html = renderForm({ ...VALUES, formation: 'free' });
		expect(html).toMatch(/<input type="radio" name="formation" value="free"[^>]*checked/);
		expect(html).toMatch(/name="freeSlots"[^>]*value="12"|value="12"[^>]*name="freeSlots"/);
		expect(html).not.toContain('name="tank"');
		expect(html).toContain('1 de 12');
	});

	it('CA-04.3 / RN-07: a formação travada e as vagas do grupo livre voltam com mensagem', () => {
		const errors = lobbyFieldMessages([
			{ field: 'formation', code: 'locked' },
			{ field: 'freeSlots', code: 'below_occupied' }
		]);
		expect(errors).toEqual({
			formation: 'Só dá para trocar a formação com o grupo vazio',
			freeSlots: 'O grupo já tem mais gente que isso'
		});
		expect(renderForm({ ...VALUES, formation: 'free' }, errors)).toContain(
			'Só dá para trocar a formação com o grupo vazio'
		);
		expect(ruleMessage('group_full')).toBe('Esse grupo não tem mais vaga.');
	});
});

describe('detalhe e diálogos do grupo livre', () => {
	it('CA-02.5 / RN-11: qualquer função com vaga; fica de fora só pelo nível', () => {
		const got = eligibility({ ...FREE, minLevel: 160 }, [
			char('t', 'tank', 170),
			char('d', 'dps', 150)
		]);
		expect(got.map((e) => [e.character.id, e.ok, e.why])).toEqual([
			['t', true, '1 vaga livre'],
			['d', false, 'abaixo do nível 160']
		]);
		const full = { ...FREE, occupied: { tank: 1, support: 2, dps: 0 } };
		expect(eligibility(full, [char('t', 'tank', 170)])[0]).toMatchObject({
			ok: false,
			why: 'grupo cheio'
		});
	});

	it('CA-02.4 / RN-05: a troca no grupo livre fica com a mesma vaga', () => {
		const full = { ...FREE, occupied: { tank: 0, support: 3, dps: 0 } };
		const got = swapEligibility(
			full,
			[char('t', 'tank', 200)],
			{ characterId: 'x', role: 'support' },
			'owner'
		);
		expect(got[0]).toMatchObject({ ok: true, why: 'mesma vaga' });
	});

	it('CA-01.1 / RN-10: os lugares do grupo livre são o anfitrião e as vagas abertas', () => {
		const places = freeComposition({ ...FREE, members: [] }, people({ ...FREE, members: [] }));
		expect(places).toHaveLength(3);
		expect(places[0]?.kind).toBe('host');
		expect(places.slice(1)).toEqual([null, null]);
	});
});

describe('painéis do grupo livre', () => {
	const SWAP = {
		id: 's1',
		applicationId: 'a1',
		discordName: 'caio',
		reason: 'o grupo precisa de tank',
		createdAt: '',
		from: {
			characterId: 'f',
			nick: 'Fogo',
			classId: 'arquimago',
			level: 200,
			portrait: 'retrato-1',
			role: 'dps'
		},
		to: {
			characterId: 'e',
			nick: 'Escudo',
			classId: 'guardiao-real',
			level: 200,
			portrait: 'retrato-2',
			role: 'tank'
		}
	};

	it('CA-02.4 / RN-05: no grupo livre cheio, o painel da troca não fala de vaga de função', () => {
		const full = { ...FREE, occupied: { tank: 0, support: 3, dps: 0 } };
		const html = render(SwapPanel, {
			props: { swap: SWAP, lobby: full, onreject: () => {} } as never
		}).body;
		expect(html).toContain('a mesma do membro');
		expect(html).not.toContain('Vaga de Tank');
		expect(html).not.toContain('o aceite falha');
		const roles = render(SwapPanel, {
			props: {
				swap: SWAP,
				lobby: { ...TEMPLE, occupied: { tank: 1, support: 1, dps: 0 } },
				onreject: () => {}
			} as never
		}).body;
		expect(roles).toContain('Vaga de Tank');
	});

	it('CA-02.1 / RN-04: sair do grupo livre libera "Sua vaga", sem função', () => {
		const html = render(LeaveDialog, {
			props: { applicationId: 'a1', title: 'Templo', roleLabel: null, onclose: () => {} }
		}).body;
		expect(html).toContain('Sua vaga fica livre');
		expect(html).not.toContain('Sua vaga de');
	});
});

describe('textos de troca do grupo livre', () => {
	it('CA-02.4 / RN-05: a troca do dono no grupo livre não fala de vaga na função', () => {
		const host = people(FREE)[0];
		const panel = (free: boolean) =>
			render(PlayerPanel, {
				props: {
					person: host,
					classNames: {},
					canDecide: true,
					free,
					onreject: () => {},
					onremove: () => {},
					onswap: () => {}
				} as never
			}).body;
		expect(panel(true)).toContain('a vaga continua sua');
		expect(panel(true)).not.toContain('vaga na função');
		expect(panel(false)).toContain('vaga na função');
	});

	it('CA-02.4 / RN-05: o pedido de troca no grupo livre não espera vaga', () => {
		const dialog = (free: boolean) =>
			render(SwapDialog, {
				props: {
					mode: 'request',
					hint: '',
					options: [],
					classNames: {},
					applicationId: 'a1',
					free,
					onclose: () => {}
				} as never
			}).body;
		expect(dialog(true)).toContain('O personagem novo fica com a mesma vaga.');
		expect(dialog(false)).toContain('O pedido pode esperar a vaga abrir.');
	});
});
