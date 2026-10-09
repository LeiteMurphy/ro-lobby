import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import CharacterCard from '$lib/characters/components/CharacterCard.svelte';
import type { Character } from '$lib/characters/api';
import AvailabilityDialog from './components/AvailabilityDialog.svelte';
import { CLOCK_OPTIONS, daysLabel, instancesLabel, rangeLabel } from './format';
import { talentFieldMessages } from './messages';

// Textos, mensagens e o diálogo de disponibilidade do banco de talentos (spec
// banco-de-talentos, T-05).

const LIRIEN: Character = {
	id: 'c1',
	nick: 'Lirien',
	classId: 'arcebispo',
	level: 178,
	role: 'support',
	portrait: 'retrato-2',
	link: null,
	isMain: true,
	createdAt: '',
	availability: null
};
const INSTANCES = [
	{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei', level: 160 },
	{ id: 'sonho-sombrio', name: 'Sonho Sombrio', level: 120 }
];

describe('textos do banco de talentos', () => {
	it('RN-02: dias seguidos viram faixa; todos, "todos os dias"; soltos, lista', () => {
		expect(daysLabel([1, 2, 3, 4, 5])).toBe('seg a sex');
		expect(daysLabel([0, 1, 2, 3, 4, 5, 6])).toBe('todos os dias');
		expect(daysLabel([5, 6, 0])).toBe('sex a dom');
		expect(daysLabel([1, 3, 5])).toBe('seg, qua, sex');
		expect(daysLabel([3])).toBe('qua');
	});

	it('RN-03: horas de 30 em 30 minutos e a faixa que vira a meia-noite', () => {
		expect(CLOCK_OPTIONS).toHaveLength(48);
		expect(CLOCK_OPTIONS[0]).toBe('00:00');
		expect(CLOCK_OPTIONS[47]).toBe('23:30');
		expect(rangeLabel('19:00', '23:00')).toBe('19:00–23:00');
		expect(rangeLabel('22:00', '02:00')).toBe('22:00–02:00 (até o dia seguinte)');
	});

	it('RN-04: "Qualquer instância" ou os nomes; sem nenhuma no catálogo, diz isso', () => {
		expect(instancesLabel(true, [])).toBe('Qualquer instância');
		expect(instancesLabel(false, [])).toBe('Nenhuma instância do catálogo');
		expect(instancesLabel(false, ['Templo do Demônio Rei', 'Sonho Sombrio'])).toBe(
			'Templo do Demônio Rei, Sonho Sombrio'
		);
	});

	it('CA-01.3 / D-06: mensagens dos erros por campo', () => {
		expect(
			talentFieldMessages([
				{ field: 'days', code: 'required' },
				{ field: 'end', code: 'same_as_start' },
				{ field: 'instanceIds', code: 'required' }
			])
		).toEqual({
			days: 'Escolha pelo menos um dia',
			end: 'O fim precisa ser diferente do início',
			instanceIds: 'Escolha uma instância ou marque Qualquer instância'
		});
	});
});

describe('disponibilidade no perfil', () => {
	const renderDialog = (character: Character, form: unknown = null) =>
		render(AvailabilityDialog, {
			props: { character, instances: INSTANCES, form, onclose: () => {} } as never
		}).body;

	it('CA-01.4: reabre com os dias, a faixa e as instâncias guardados, mesmo desligado', () => {
		const html = renderDialog({
			...LIRIEN,
			availability: {
				enabled: false,
				days: [1, 2],
				start: '22:00',
				end: '02:00',
				anyInstance: false,
				instanceIds: ['sonho-sombrio']
			}
		});
		expect(html).not.toMatch(/name="enabled"[^>]*checked/);
		expect(html).toMatch(/value="1"[^>]*checked/);
		expect(html).toMatch(/value="2"[^>]*checked/);
		expect(html).not.toMatch(/value="3"[^>]*checked/);
		expect(html).toMatch(/<option value="22:00"[^>]*selected/);
		expect(html).toMatch(/<option value="02:00"[^>]*selected/);
		expect(html).toMatch(/value="sonho-sombrio"[^>]*checked/);
		expect(html).not.toMatch(/value="templo-do-demonio-rei"[^>]*checked/);
		expect(html).toContain('Fora do banco. Os dias, a faixa e as instâncias ficam guardados.');
	});

	it('CA-01.1: sem disponibilidade, começa ligado, 19:00–23:00 e "Qualquer instância"', () => {
		const html = renderDialog(LIRIEN);
		expect(html).toMatch(/name="enabled"[^>]*checked/);
		expect(html).toMatch(/<option value="19:00"[^>]*selected/);
		expect(html).toMatch(/<option value="23:00"[^>]*selected/);
		expect(html).toMatch(/name="anyInstance"[^>]*checked/);
		expect(html).not.toContain('name="instanceIds"');
	});

	it('CA-01.3 / RNF-03: os erros voltam ligados aos campos', () => {
		const html = renderDialog(LIRIEN, {
			values: {
				enabled: true,
				days: [],
				start: '19:00',
				end: '19:00',
				anyInstance: false,
				instanceIds: []
			},
			errors: { days: 'Escolha pelo menos um dia', end: 'O fim precisa ser diferente do início' }
		});
		expect(html).toContain('Escolha pelo menos um dia');
		expect(html).toMatch(/aria-describedby="([^"]+-end-err)"[\s\S]*id="\1"/);
		expect(html).toMatch(/<fieldset[^>]*aria-describedby="[^"]+-days-err"/);
		expect(html).toContain('name="instanceIds"');
	});

	it('RN-01: o card mostra se o personagem está no banco', () => {
		const card = (availability: Character['availability']) =>
			render(CharacterCard, {
				props: {
					character: { ...LIRIEN, availability },
					className: 'Arcebispo',
					onedit: () => {},
					ondelete: () => {},
					onavailability: () => {}
				}
			}).body;
		expect(card(null)).toContain('Fora do banco de talentos');
		const on = {
			enabled: true,
			days: [1],
			start: '19:00',
			end: '23:00',
			anyInstance: true,
			instanceIds: []
		};
		expect(card(on)).toContain('No banco de talentos');
		expect(card({ ...on, enabled: false })).toContain('Fora do banco de talentos');
	});
});
