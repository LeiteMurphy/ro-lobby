// Retratos de exemplo do RO Lobby (spec personagens, RN-10, design D-04 e D-10): arte
// original em pixel art, servida pelo próprio web em static/portraits/. A lista definitiva
// vem depois; os ids seguem o enum Portrait do contrato.
import type { Portrait } from './api';

export interface PortraitInfo {
	id: Portrait;
	label: string;
	src: string;
}

const LABELS = {
	'retrato-1': 'Retrato 1, espada',
	'retrato-2': 'Retrato 2, lua',
	'retrato-3': 'Retrato 3, estrela',
	'retrato-4': 'Retrato 4, escudo'
} satisfies Record<Portrait, string>;

export const PORTRAITS: readonly PortraitInfo[] = (Object.keys(LABELS) as Portrait[]).map((id) => ({
	id,
	label: LABELS[id],
	src: `/portraits/${id}.svg`
}));

/** O primeiro da lista é o padrão de quem não escolhe (RN-10). */
export const DEFAULT_PORTRAIT: Portrait = 'retrato-1';

export function portraitInfo(id: string): PortraitInfo {
	return PORTRAITS.find((p) => p.id === id) ?? PORTRAITS[0];
}
