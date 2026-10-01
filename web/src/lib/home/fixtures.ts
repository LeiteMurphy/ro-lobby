// RN-08 / D-06: dados fictícios da Home, iguais aos do design "Home v2". É o único
// lugar que conhece esses dados; o resto do web usa getHomeLobbies(), que a feature de
// lobby vai trocar pela API.
import { CLASSES } from '$lib/catalog/classes';
import { addDays } from './time';
import type { Composition, Lobby } from './types';

type Template = Omit<Lobby, 'id' | 'date'> & { key: string };

const comp = (
	tank: [number, number],
	support: [number, number],
	dps: [number, number]
): Composition => ({
	tank: { filled: tank[0], total: tank[1] },
	support: { filled: support[0], total: support[1] },
	dps: { filled: dps[0], total: dps[1] }
});

const TEMPLATES: readonly Template[] = [
	{
		key: 'a',
		time: '18:00',
		instance: 'Caverna de gelo',
		host: 'Lyrae',
		hostClass: 'Arcebispo',
		minLevel: 160,
		composition: comp([1, 1], [2, 3], [3, 6])
	},
	{
		key: 'b',
		time: '19:00',
		instance: 'Torre sem fim',
		host: 'Kaizen',
		hostClass: 'Paladino',
		minLevel: 175,
		composition: comp([0, 1], [1, 3], [5, 5])
	},
	{
		key: 'c',
		time: '19:30',
		instance: 'Templo submerso',
		host: 'mirai.exe',
		hostClass: 'Feiticeiro',
		minLevel: 150,
		composition: comp([1, 1], [3, 3], [8, 8])
	},
	{
		key: 'd',
		time: '20:30',
		instance: 'Fortaleza do deserto',
		host: 'Nhoque',
		hostClass: 'Sicário',
		minLevel: 170,
		composition: comp([0, 1], [2, 2], [5, 5])
	},
	{
		key: 'e',
		time: '21:30',
		instance: 'Ruínas ao norte',
		host: 'Valkyrja',
		hostClass: 'Guardião Real',
		minLevel: 185,
		composition: comp([1, 1], [1, 4], [5, 7])
	},
	{
		key: 'f',
		time: '22:45',
		instance: 'Torre sem fim',
		host: 'Zé Morcego',
		hostClass: 'Musa',
		minLevel: 160,
		composition: comp([0, 1], [0, 2], [2, 9])
	}
];

/** Classes que fazem instância difícil: 3ª, 4ª e expandidas, na ordem do catálogo. */
const HOST_CLASSES = CLASSES.filter((c) =>
	['terceira', 'quarta', 'expandida'].includes(c.tier)
).map((c) => c.name);

/** Quantidade de lobbies em cada um dos 14 dias, a partir de hoje (como no design). */
export const DAY_COUNTS: readonly number[] = [6, 4, 0, 5, 6, 6, 3, 2, 4, 0, 5, 6, 6, 2];

/** Lobbies fictícios dos 14 dias a partir de `today` (YYYY-MM-DD). */
export function getHomeLobbies(today: string): Lobby[] {
	return DAY_COUNTS.flatMap((count, day) => {
		const date = addDays(today, day);
		return TEMPLATES.filter((_, i) => (i + day * 2) % TEMPLATES.length < count).map(
			({ key, hostClass, ...rest }, i) => ({
				...rest,
				// Hoje fica igual ao design; os outros dias passam pelas classes do catálogo.
				hostClass:
					day === 0 ? hostClass : HOST_CLASSES[(day * TEMPLATES.length + i) % HOST_CLASSES.length],
				id: `${date}-${key}`,
				date
			})
		);
	});
}
