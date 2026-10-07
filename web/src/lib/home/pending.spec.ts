import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { TEMPLE } from '$lib/lobbies/fixtures';
import { toHomeLobby } from '$lib/lobbies/toHome';
import FeaturedLobby from './components/FeaturedLobby.svelte';
import LobbyCard from './components/LobbyCard.svelte';
import { pendingFor } from './lobbies';

// Selo "N pendentes" no card da Home, só para o dono (spec candidatura-lobby, RN-34,
// CA-10.5).

const OWNER = TEMPLE.owner.userId;
const lobby = (pendingCount: number) =>
	toHomeLobby({ ...TEMPLE, pendingCount }, new Map([['arcebispo', 'Arcebispo']]));

describe('selo de pendentes', () => {
	it('CA-10.5 / RN-34: o texto só existe para o dono e com pendentes', () => {
		expect(pendingFor(lobby(2), OWNER)).toBe('2 pendentes');
		expect(pendingFor(lobby(1), OWNER)).toBe('1 pendente');
		expect(pendingFor(lobby(0), OWNER)).toBeNull();
		expect(pendingFor(lobby(2), 'outro')).toBeNull();
		expect(pendingFor(lobby(2), null)).toBeNull();
	});

	it('D-06: a Home guarda o dono e a quantidade que vêm da API', () => {
		expect(lobby(3)).toMatchObject({ ownerId: OWNER, pendingCount: 3 });
	});

	it('CA-10.5: o card e o destaque mostram o selo ao dono e não aos outros', () => {
		const card = (viewerId: string | null) =>
			render(LobbyCard, { props: { lobby: lobby(2), relative: '', viewerId } }).body;
		expect(card(OWNER)).toMatch(/data-testid="pending-badge"[^>]*>2 pendentes</);
		expect(card('outro')).not.toContain('pending-badge');
		expect(card(null)).not.toContain('pending-badge');
		const featured = (viewerId: string | null) =>
			render(FeaturedLobby, { props: { lobby: lobby(2), relative: '', viewerId } }).body;
		expect(featured(OWNER)).toMatch(/data-testid="pending-badge"[^>]*>2 pendentes</);
		expect(featured('outro')).not.toContain('pending-badge');
	});
});
