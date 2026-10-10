// Retratos do RO Lobby: a arte da linha da classe do personagem (spec retrato-por-classe,
// RN-01, RN-02, D-04), em pixel art original servida pelo próprio web em
// static/portraits/classes/. Os 4 retratos genéricos de antes (static/portraits/*.svg)
// ficam para a classe fora do catálogo (RN-04) e para dados antigos.
import type { Portrait } from './api';

export interface PortraitInfo {
	id: Portrait;
	/** Texto alternativo: o nome da linha, ex. "Linha do Arruaceiro" (RN-05). */
	label: string;
	src: string;
}

const GENERIC = {
	'retrato-1': 'Retrato 1, espada',
	'retrato-2': 'Retrato 2, lua',
	'retrato-3': 'Retrato 3, estrela',
	'retrato-4': 'Retrato 4, escudo'
} as const;

const ARTS = {
	superaprendiz: 'Linha do Superaprendiz',
	espadachim: 'Família Espadachim',
	cavaleiro: 'Linha do Cavaleiro',
	templario: 'Linha do Templário',
	mago: 'Família Mago',
	bruxo: 'Linha do Bruxo',
	sabio: 'Linha do Sábio',
	gatuno: 'Família Gatuno',
	mercenario: 'Linha do Mercenário',
	arruaceiro: 'Linha do Arruaceiro',
	mercador: 'Família Mercador',
	ferreiro: 'Linha do Ferreiro',
	alquimista: 'Linha do Alquimista',
	novico: 'Família Noviço',
	sacerdote: 'Linha do Sacerdote',
	monge: 'Linha do Monge',
	arqueiro: 'Família Arqueiro',
	cacador: 'Linha do Caçador',
	bardo: 'Linha do Bardo',
	odalisca: 'Linha da Odalisca',
	taekwon: 'Família Taekwon',
	'mestre-taekwon': 'Linha do Mestre Taekwon',
	espiritualista: 'Linha do Espiritualista',
	ninja: 'Linha do Ninja',
	justiceiro: 'Linha do Justiceiro',
	invocador: 'Linha do Invocador (Doram)',
	druida: 'Linha do Druida'
} as const satisfies Partial<Record<Portrait, string>>;

/** Os ids das 27 artes de classe, cada uma com um PNG em static/portraits/classes/. */
export const CLASS_ARTS = Object.keys(ARTS) as (keyof typeof ARTS)[];

/** Os 4 retratos genéricos de antes, em static/portraits/. */
export const GENERIC_PORTRAITS = Object.keys(GENERIC) as (keyof typeof GENERIC)[];

/** O retrato de quem tem a classe fora do catálogo (RN-04). */
export const DEFAULT_PORTRAIT: Portrait = 'retrato-1';

/** RN-05: arquivo e texto alternativo do retrato; desconhecido vira o padrão. */
export function portraitInfo(id: string | null | undefined): PortraitInfo {
	if (id && id in ARTS) {
		const art = id as keyof typeof ARTS;
		return { id: art, label: ARTS[art], src: `/portraits/classes/${art}.png` };
	}
	const generic = (id && id in GENERIC ? id : DEFAULT_PORTRAIT) as keyof typeof GENERIC;
	return { id: generic, label: GENERIC[generic], src: `/portraits/${generic}.svg` };
}

/** O endereço do retrato, para as listas que só mostram a imagem. */
export function portraitSrc(id: string | null | undefined): string {
	return portraitInfo(id).src;
}

/** RN-05: o texto alternativo do retrato, o nome da linha. */
export function portraitAlt(id: string | null | undefined): string {
	return portraitInfo(id).label;
}

/** RN-01 / CA-01.4: o retrato da classe escolhida, pelo catálogo de GET /classes. */
export function portraitForClass(
	classId: string,
	classes: readonly { id: string; art: string }[]
): PortraitInfo {
	return portraitInfo(classes.find((c) => c.id === classId)?.art);
}
