import { describe, expect, it } from 'vitest';
import { getHomeLobbies } from '$lib/home/fixtures';
import { classArt } from '$lib/home/catalog';
import { ICONS } from '$lib/ui/Icon.svelte';
import { CLASSES, findClass } from './classes';

// Títulos das páginas de classe em https://browiki.org/wiki/Classes, como estavam em
// 2026-10-01, na ordem da página.
const BROWIKI_PAGES = [
	'Aprendizes',
	'Espadachins',
	'Cavaleiros',
	'Lordes',
	'Cavaleiros Rúnicos',
	'Cavaleiros Draconianos',
	'Templários',
	'Paladinos',
	'Guardiões Reais',
	'Guardiões Imperiais',
	'Magos',
	'Bruxos',
	'Arquimagos',
	'Arcanos',
	'Magus',
	'Sábios',
	'Professores',
	'Feiticeiros',
	'Elementalistas',
	'Gatunos',
	'Mercenários',
	'Algozes',
	'Sicários',
	'Executores',
	'Arruaceiros',
	'Desordeiros',
	'Renegados',
	'Mandraques',
	'Mercadores',
	'Ferreiros',
	'Mestres-Ferreiros',
	'Mecânicos',
	'Engenheiros',
	'Alquimistas',
	'Criadores',
	'Bioquímicos',
	'Cientistas',
	'Noviços',
	'Sacerdotes',
	'Sumo Sacerdotes',
	'Arcebispos',
	'Cardeais',
	'Monges',
	'Mestres',
	'Shuras',
	'Inquisidores',
	'Arqueiros',
	'Caçadores',
	'Atiradores de Elite',
	'Sentinelas',
	'Falcões do Vento',
	'Bardos',
	'Odaliscas',
	'Menestréis',
	'Ciganas',
	'Trovadores',
	'Musas',
	'Maestros',
	'Divas',
	'Taekwons',
	'Mestres Taekwons',
	'Mestres Estelares',
	'Mestres Celestiais',
	'Espiritualistas',
	'Ceifadores de Almas',
	'Ascetas das Almas',
	'Superaprendizes',
	'Superaprendizes EX',
	'Hiperaprendizes',
	'Justiceiros',
	'Insurgentes',
	'Guerrilheiros',
	'Ninjas',
	'Kagerou',
	'Oboro',
	'Shinkiro',
	'Shiranui',
	'Druidas',
	'Karnos',
	'Alitea',
	'Invocadores',
	'Animistas'
];

describe('catálogo de classes do bRO', () => {
	it('RN-22: tem exatamente as classes da página do bROWiki', () => {
		expect([...CLASSES.map((c) => c.plural)].sort()).toEqual([...BROWIKI_PAGES].sort());
		expect(CLASSES).toHaveLength(82);
	});

	it('RN-22: nomes no singular sem repetição', () => {
		expect(new Set(CLASSES.map((c) => c.name)).size).toBe(CLASSES.length);
	});

	it('RN-22: cada uma das 6 classes de 1ª tem as linhas completas até a 4ª classe', () => {
		const families = ['Espadachim', 'Mago', 'Gatuno', 'Mercador', 'Noviço', 'Arqueiro'];
		for (const family of families) {
			const tiers = new Set(CLASSES.filter((c) => c.family === family).map((c) => c.tier));
			expect([...tiers].sort(), family).toEqual(
				['primeira', 'quarta', 'segunda', 'terceira', 'transcendental'].sort()
			);
			// Duas linhas por família: duas classes de 4ª.
			expect(CLASSES.filter((c) => c.family === family && c.tier === 'quarta').length, family).toBe(
				family === 'Arqueiro' ? 3 : 2
			);
		}
	});

	it('RNF-02: todo ícone do catálogo está empacotado no build', () => {
		for (const c of CLASSES) expect(Object.keys(ICONS), c.name).toContain(c.icon);
	});

	it('RN-22: as classes do design mantêm o ícone e a cor de função do Home v2', () => {
		expect(classArt('Arcebispo')).toEqual({ icon: 'cross', role: 'support' });
		expect(classArt('Paladino')).toEqual({ icon: 'shield-half', role: 'tank' });
		expect(classArt('Feiticeiro')).toEqual({ icon: 'wand-sparkles', role: 'dps' });
		expect(classArt('Sicário')).toEqual({ icon: 'venetian-mask', role: 'dps' });
		expect(classArt('Guardião Real')).toEqual({ icon: 'shield-plus', role: 'tank' });
		expect(classArt('Musa')).toEqual({ icon: 'music', role: 'support' });
	});

	it('RN-22: os dados fictícios só usam classes do catálogo', () => {
		const used = new Set(getHomeLobbies('2026-09-30').map((l) => l.hostClass));
		for (const name of used) expect(findClass(name), name).toBeDefined();
		// Os 14 dias passam por várias classes, não só pelas 6 de hoje.
		expect(used.size).toBeGreaterThan(20);
	});
});
