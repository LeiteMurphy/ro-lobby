// Lobby de exemplo para os testes (spec lobbies). Não é usado pela aplicação.
import type { ApiLobby } from './api';

export const TEMPLE: ApiLobby = {
	id: '4b1c2d3e-5f60-4a7b-8c9d-0e1f2a3b4c5d',
	instance: {
		id: 'templo-do-demonio-rei',
		name: 'Templo do Demônio Rei',
		level: 160,
		reset: 'daily'
	},
	startsAt: '2026-10-07T23:00:00Z',
	status: 'open',
	slots: { tank: 1, support: 2, dps: 3 },
	occupied: { tank: 0, support: 1, dps: 0 },
	minLevel: 160,
	note: null,
	owner: {
		userId: '6f1c2b8e-3a4d-4e5f-9a1b-2c3d4e5f6a7b',
		discordName: 'Grimbold',
		characterId: '0b7e7f0e-8d4c-4f34-9a55-3f9f4c1a2b3c',
		nick: 'Lirien',
		classId: 'arcebispo',
		level: 178,
		portrait: 'retrato-2',
		role: 'support'
	},
	cancelReason: null,
	createdAt: '2026-10-06T19:40:00Z'
};
