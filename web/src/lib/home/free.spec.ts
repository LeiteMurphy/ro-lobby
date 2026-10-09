import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { TEMPLE } from '$lib/lobbies/fixtures';
import { toHomeLobby } from '$lib/lobbies/toHome';
import FeaturedLobby from './components/FeaturedLobby.svelte';
import LobbyCard from './components/LobbyCard.svelte';
import { applyFilters, NO_FILTERS, roleCounts } from './filters';
import { featuredLobby, lobbyIsFull, roomFor, seats } from './lobbies';
import type { Composition, Lobby } from './types';

// Grupo livre na Home (spec grupo-livre, T-05).

const NO_ROLES: Composition = {
	tank: { filled: 0, total: 0 },
	support: { filled: 0, total: 0 },
	dps: { filled: 0, total: 0 }
};
const base: Lobby = {
	id: 'x',
	date: '2026-10-07',
	time: '20:00',
	instance: 'Templo do Demônio Rei',
	host: 'Lirien',
	hostClass: 'Arcebispo',
	minLevel: 160,
	composition: NO_ROLES
};
const free = (filled: number, total: number): Lobby => ({
	...base,
	id: `f${filled}`,
	free: { filled, total }
});
// Por função, sem vaga de Tank.
const noTank: Lobby = {
	...base,
	id: 'r',
	composition: {
		tank: { filled: 1, total: 1 },
		support: { filled: 1, total: 2 },
		dps: { filled: 0, total: 3 }
	}
};

describe('grupo livre na Home', () => {
	it('CA-03.1 / RN-08: a API vira "free" com os ocupantes de qualquer função', () => {
		const home = toHomeLobby(
			{
				...TEMPLE,
				formation: 'free',
				freeSlots: 12,
				slots: { tank: 0, support: 0, dps: 0 },
				occupied: { tank: 1, support: 2, dps: 1 }
			},
			new Map()
		);
		expect(home.free).toEqual({ filled: 4, total: 12 });
		expect(toHomeLobby(TEMPLE, new Map()).free).toBeUndefined();
	});

	it('RN-04 / RN-08: vaga no total, lotado e ocupação', () => {
		expect(roomFor(free(4, 12), 'tank')).toBe(8);
		expect(roomFor(free(12, 12), 'dps')).toBe(0);
		expect(lobbyIsFull(free(12, 12))).toBe(true);
		expect(lobbyIsFull(free(11, 12))).toBe(false);
		expect(seats(free(4, 12))).toEqual({ filled: 4, total: 12 });
		expect(roomFor(noTank, 'tank')).toBe(0);
	});

	it('CA-03.2 / RN-09: "Vaga para: Tank" traz o grupo livre e conta ele na lateral', () => {
		const day = [free(4, 12), noTank, free(12, 12)];
		expect(applyFilters(day, { ...NO_FILTERS, roles: ['tank'] }).map((l) => l.id)).toEqual(['f4']);
		expect(roleCounts(day)).toEqual({ tank: 1, support: 2, dps: 2 });
	});

	it('RN-16: o destaque pula o grupo livre lotado', () => {
		const now = { date: '2026-10-06', minutes: 0 };
		expect(featuredLobby([free(12, 12), free(3, 6)], now)?.id).toBe('f3');
	});

	it('CA-03.1 / RN-08: card e destaque com o selo, "X de N" e a borda neutra, sem chips', () => {
		const card = render(LobbyCard, { props: { lobby: free(4, 12), relative: '' } }).body;
		expect(card).toContain('Grupo livre');
		expect(card).toContain('4 de 12');
		expect(card).toContain('data-edge="free"');
		expect(card).not.toContain('data-role=');
		expect(card).toMatch(/4\/12/);
		const hero = render(FeaturedLobby, { props: { lobby: free(4, 12), relative: '' } }).body;
		expect(hero).toContain('data-testid="free-composition"');
		expect(hero).not.toContain('data-role=');
	});
});
