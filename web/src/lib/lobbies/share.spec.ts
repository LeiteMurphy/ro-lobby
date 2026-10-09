import { describe, expect, it } from 'vitest';
import { TEMPLE } from './fixtures';
import { inviteText, openSlotsLabel, shareMeta, whenLabel } from './share';
import type { ApiLobby } from './api';

const origin = 'https://rolobby.com.br';

// O lobby do CA-01.3: Glast Heim no sábado 10/10 às 20:00 (23:00 UTC), nível mínimo 160,
// 1 vaga de Tank e 2 de Dano abertas e Suporte cheio, anfitrião "Brasa".
const GLAST: ApiLobby = {
	...TEMPLE,
	id: 'abc-123',
	instance: { ...TEMPLE.instance, name: 'Glast Heim' },
	startsAt: '2026-10-10T23:00:00Z',
	minLevel: 160,
	slots: { tank: 1, support: 2, dps: 3 },
	occupied: { tank: 0, support: 2, dps: 1 },
	owner: { ...TEMPLE.owner, nick: 'Brasa' }
};

describe('convite do lobby', () => {
	it('CA-01.3: copia o convite de três linhas', () => {
		expect(inviteText(GLAST, origin)).toBe(
			[
				'Grupo para Glast Heim · sábado, 10/10 às 20:00 (horário de Brasília)',
				'Vagas: 1 Tank, 2 Dano · Nível mínimo 160',
				'Candidate-se: https://rolobby.com.br/lobbies/abc-123'
			].join('\n')
		);
	});

	it('CA-01.3 / RN-02: a data é absoluta e em Brasília, mesmo virando o dia em UTC', () => {
		// 02:30 UTC de domingo ainda é sábado, 23:30, em São Paulo.
		expect(whenLabel('2026-10-11T02:30:00Z')).toBe('sábado, 10/10 às 23:30');
		expect(whenLabel('2026-01-05T12:05:00Z')).toBe('segunda, 05/01 às 09:05');
	});

	it('CA-01.3 / RN-03: vagas abertas na ordem Tank, Suporte, Dano', () => {
		expect(
			openSlotsLabel({
				slots: { tank: 1, support: 4, dps: 2 },
				occupied: { tank: 0, support: 1, dps: 2 }
			})
		).toBe('Vagas: 1 Tank, 3 Suporte');
	});

	it('CA-01.4: sem vaga aberta, a segunda linha começa com "Grupo lotado"', () => {
		const full = { ...GLAST, occupied: { ...GLAST.slots } };
		expect(inviteText(full, origin).split('\n')[1]).toBe('Grupo lotado · Nível mínimo 160');
	});
});

describe('preview do lobby (Open Graph)', () => {
	it('CA-02.1: título, descrição e link do lobby aberto', () => {
		expect(shareMeta(GLAST, origin)).toEqual({
			title: 'Glast Heim · sábado, 10/10 às 20:00',
			description: 'Vagas: 1 Tank, 2 Dano · Nível mínimo 160 · Anfitrião Brasa',
			url: 'https://rolobby.com.br/lobbies/abc-123'
		});
	});

	it('CA-02.2: lobby cancelado avisa o cancelamento', () => {
		expect(shareMeta({ ...GLAST, status: 'cancelled' }, origin).description).toBe(
			'Esse grupo foi cancelado.'
		);
	});

	it('CA-02.2 / RN-05: lobby iniciado avisa que já começou', () => {
		expect(shareMeta({ ...GLAST, status: 'started' }, origin).description).toBe(
			'Esse grupo já começou.'
		);
	});
});

describe('convite do grupo livre (spec grupo-livre, RN-12)', () => {
	const FREE: ApiLobby = {
		...GLAST,
		formation: 'free',
		freeSlots: 12,
		slots: { tank: 0, support: 0, dps: 0 },
		occupied: { tank: 1, support: 2, dps: 1 }
	};

	it('CA-05.1 / RN-12: "Vagas: 8 livres" no convite e no preview', () => {
		expect(inviteText(FREE, origin).split('\n')[1]).toBe('Vagas: 8 livres · Nível mínimo 160');
		expect(shareMeta(FREE, origin).description).toBe(
			'Vagas: 8 livres · Nível mínimo 160 · Anfitrião Brasa'
		);
	});

	it('RN-12: uma vaga é "1 livre"; sem vaga, "Grupo lotado"', () => {
		expect(openSlotsLabel({ ...FREE, freeSlots: 5 })).toBe('Vagas: 1 livre');
		expect(openSlotsLabel({ ...FREE, freeSlots: 4 })).toBe('Grupo lotado');
	});
});
