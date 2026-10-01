// Catálogo de classes do bRO, conforme https://browiki.org/wiki/Classes (consultado em
// 2026-10-01). `plural` é o título da página da classe no bROWiki, exatamente como lá;
// `name` é o singular usado na tela (RN-22 da spec home-local).
//
// `role` é só uma sugestão de função para colorir o ícone da classe: no jogo, a função
// depende da build do personagem, e o cadastro de personagem vai perguntar ao jogador.
import type { IconName } from '$lib/ui/Icon.svelte';
import type { Role } from '$lib/home/types';

export type ClassTier =
	'aprendiz' | 'primeira' | 'segunda' | 'transcendental' | 'terceira' | 'quarta' | 'expandida';

export interface RoClass {
	name: string;
	plural: string;
	tier: ClassTier;
	/** Classe de 1ª (ou a classe base, nas expandidas) de onde a linha sai. */
	family: string;
	icon: IconName;
	role: Role;
}

type Entry = [name: string, plural: string, tier: ClassTier];

function line(family: string, icon: IconName, role: Role, entries: Entry[]): RoClass[] {
	return entries.map(([name, plural, tier]) => ({ name, plural, tier, family, icon, role }));
}

export const CLASSES: readonly RoClass[] = [
	...line('Aprendiz', 'user', 'dps', [['Aprendiz', 'Aprendizes', 'aprendiz']]),

	// Espadachim
	...line('Espadachim', 'sword', 'dps', [
		['Espadachim', 'Espadachins', 'primeira'],
		['Cavaleiro', 'Cavaleiros', 'segunda'],
		['Lorde', 'Lordes', 'transcendental'],
		['Cavaleiro Rúnico', 'Cavaleiros Rúnicos', 'terceira'],
		['Cavaleiro Draconiano', 'Cavaleiros Draconianos', 'quarta']
	]),
	...line('Espadachim', 'shield-half', 'tank', [
		['Templário', 'Templários', 'segunda'],
		['Paladino', 'Paladinos', 'transcendental']
	]),
	...line('Espadachim', 'shield-plus', 'tank', [
		['Guardião Real', 'Guardiões Reais', 'terceira'],
		['Guardião Imperial', 'Guardiões Imperiais', 'quarta']
	]),

	// Mago
	...line('Mago', 'zap', 'dps', [
		['Mago', 'Magos', 'primeira'],
		['Bruxo', 'Bruxos', 'segunda'],
		['Arquimago', 'Arquimagos', 'transcendental'],
		['Arcano', 'Arcanos', 'terceira'],
		['Magus', 'Magus', 'quarta']
	]),
	...line('Mago', 'wand-sparkles', 'dps', [
		['Sábio', 'Sábios', 'segunda'],
		['Professor', 'Professores', 'transcendental'],
		['Feiticeiro', 'Feiticeiros', 'terceira'],
		['Elementalista', 'Elementalistas', 'quarta']
	]),

	// Gatuno
	...line('Gatuno', 'venetian-mask', 'dps', [
		['Gatuno', 'Gatunos', 'primeira'],
		['Mercenário', 'Mercenários', 'segunda'],
		['Algoz', 'Algozes', 'transcendental'],
		['Sicário', 'Sicários', 'terceira'],
		['Executor', 'Executores', 'quarta']
	]),
	...line('Gatuno', 'drama', 'dps', [
		['Arruaceiro', 'Arruaceiros', 'segunda'],
		['Desordeiro', 'Desordeiros', 'transcendental'],
		['Renegado', 'Renegados', 'terceira'],
		['Mandraque', 'Mandraques', 'quarta']
	]),

	// Mercador
	...line('Mercador', 'hammer', 'dps', [
		['Mercador', 'Mercadores', 'primeira'],
		['Ferreiro', 'Ferreiros', 'segunda'],
		['Mestre-Ferreiro', 'Mestres-Ferreiros', 'transcendental'],
		['Mecânico', 'Mecânicos', 'terceira'],
		['Engenheiro', 'Engenheiros', 'quarta']
	]),
	...line('Mercador', 'flask-conical', 'dps', [
		['Alquimista', 'Alquimistas', 'segunda'],
		['Criador', 'Criadores', 'transcendental'],
		['Bioquímico', 'Bioquímicos', 'terceira'],
		['Cientista', 'Cientistas', 'quarta']
	]),

	// Noviço
	...line('Noviço', 'cross', 'support', [
		['Noviço', 'Noviços', 'primeira'],
		['Sacerdote', 'Sacerdotes', 'segunda'],
		['Sumo Sacerdote', 'Sumo Sacerdotes', 'transcendental'],
		['Arcebispo', 'Arcebispos', 'terceira'],
		['Cardeal', 'Cardeais', 'quarta']
	]),
	...line('Noviço', 'hand-fist', 'dps', [
		['Monge', 'Monges', 'segunda'],
		['Mestre', 'Mestres', 'transcendental'],
		['Shura', 'Shuras', 'terceira'],
		['Inquisidor', 'Inquisidores', 'quarta']
	]),

	// Arqueiro
	...line('Arqueiro', 'crosshair', 'dps', [
		['Arqueiro', 'Arqueiros', 'primeira'],
		['Caçador', 'Caçadores', 'segunda'],
		['Atirador de Elite', 'Atiradores de Elite', 'transcendental'],
		['Sentinela', 'Sentinelas', 'terceira'],
		['Falcão do Vento', 'Falcões do Vento', 'quarta']
	]),
	...line('Arqueiro', 'music', 'support', [
		['Bardo', 'Bardos', 'segunda'],
		['Odalisca', 'Odaliscas', 'segunda'],
		['Menestrel', 'Menestréis', 'transcendental'],
		['Cigana', 'Ciganas', 'transcendental'],
		['Trovador', 'Trovadores', 'terceira'],
		['Musa', 'Musas', 'terceira'],
		['Maestro', 'Maestros', 'quarta'],
		['Diva', 'Divas', 'quarta']
	]),

	// Classes expandidas
	...line('Taekwon', 'star', 'dps', [
		['Taekwon', 'Taekwons', 'expandida'],
		['Mestre Taekwon', 'Mestres Taekwons', 'expandida'],
		['Mestre Estelar', 'Mestres Estelares', 'expandida'],
		['Mestre Celestial', 'Mestres Celestiais', 'expandida']
	]),
	...line('Taekwon', 'sparkles', 'support', [
		['Espiritualista', 'Espiritualistas', 'expandida'],
		['Ceifador de Almas', 'Ceifadores de Almas', 'expandida'],
		['Asceta das Almas', 'Ascetas das Almas', 'expandida']
	]),
	...line('Aprendiz', 'user', 'dps', [
		['Superaprendiz', 'Superaprendizes', 'expandida'],
		['Superaprendiz EX', 'Superaprendizes EX', 'expandida'],
		['Hiperaprendiz', 'Hiperaprendizes', 'expandida']
	]),
	...line('Justiceiro', 'target', 'dps', [
		['Justiceiro', 'Justiceiros', 'expandida'],
		['Insurgente', 'Insurgentes', 'expandida'],
		['Guerrilheiro', 'Guerrilheiros', 'expandida']
	]),
	...line('Ninja', 'moon', 'dps', [
		['Ninja', 'Ninjas', 'expandida'],
		['Kagerou', 'Kagerou', 'expandida'],
		['Oboro', 'Oboro', 'expandida'],
		['Shinkiro', 'Shinkiro', 'expandida'],
		['Shiranui', 'Shiranui', 'expandida']
	]),
	...line('Druida', 'leaf', 'support', [
		['Druida', 'Druidas', 'expandida'],
		['Karnos', 'Karnos', 'expandida'],
		['Alitea', 'Alitea', 'expandida']
	]),
	...line('Invocador', 'paw-print', 'support', [
		['Invocador', 'Invocadores', 'expandida'],
		['Animista', 'Animistas', 'expandida']
	])
];

const BY_NAME = new Map(CLASSES.map((c) => [c.name, c]));

export function findClass(name: string): RoClass | undefined {
	return BY_NAME.get(name);
}
