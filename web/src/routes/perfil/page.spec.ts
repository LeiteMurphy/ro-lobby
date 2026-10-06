import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import type { Character, RoClassEntry } from '$lib/characters/api';
import CharacterDialog from '$lib/characters/components/CharacterDialog.svelte';
import { CLASSES } from '$lib/catalog/classes';
import Page from './+page.svelte';

// Página de perfil renderizada no servidor (spec personagens, T-06).

const USER = { id: 'u1', username: 'grimbold', globalName: 'Grimbold' };

const CLASS_LIST: RoClassEntry[] = CLASSES.map((c) => ({
	id: c.name
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/(^-|-$)/g, ''),
	name: c.name,
	plural: c.plural,
	tier: c.tier,
	family: c.family
}));

function character(overrides: Partial<Character>): Character {
	return {
		id: 'c1',
		nick: 'Brasa',
		classId: 'guardiao-real',
		level: 172,
		role: 'tank',
		portrait: 'retrato-1',
		link: null,
		isMain: false,
		createdAt: '2026-10-06T12:00:00Z',
		...overrides
	};
}

function renderPage(characters: Character[], loadError: string | null = null): string {
	const data = {
		user: USER,
		characters,
		classes: CLASS_LIST,
		loadError,
		loginHref: '/auth/discord/login?next=%2Fperfil'
	};
	const props = { data, form: null, params: {} } as unknown as Parameters<typeof Page>[1];
	return render(Page, { props }).body;
}

const cards = (html: string) => html.split('data-testid="character-card"').slice(1);

describe('/perfil renderizada no servidor', () => {
	it('CA-01.1 / RN-16 / RN-17: uma carta por personagem, na ordem da API, com o selo no principal', () => {
		const html = renderPage([
			character({
				id: 'c2',
				nick: 'Lirien',
				classId: 'arcebispo',
				level: 178,
				role: 'support',
				portrait: 'retrato-2',
				isMain: true
			}),
			character({
				id: 'c3',
				nick: 'Faísca',
				classId: 'feiticeiro',
				level: 165,
				role: 'dps',
				portrait: 'retrato-3'
			})
		]);
		const [lirien, faisca] = cards(html);
		expect(cards(html)).toHaveLength(2);
		expect(lirien).toContain('Lirien');
		expect(lirien).toContain('Principal');
		expect(lirien).toContain('Arcebispo');
		expect(lirien).toContain('178');
		expect(lirien).toContain('Suporte');
		expect(lirien).toContain('/portraits/retrato-2.svg');
		expect(faisca).toContain('Faísca');
		expect(faisca).not.toContain('Principal');
		expect(faisca).toContain('Feiticeiro');
		expect(faisca).toContain('Dano');
		expect(html).toContain('2 de 10 personagens');
	});

	it('CA-01.2: sem personagens, mostra o perfil vazio com "Adicionar personagem"', () => {
		const html = renderPage([]);
		expect(html).toContain('Você ainda não tem personagens');
		expect(html.match(/Adicionar personagem/g)).toHaveLength(2);
		expect(cards(html)).toHaveLength(0);
	});

	it('CA-01.4 / RN-09: o link externo abre em outra aba com rel="noopener noreferrer"', () => {
		const html = renderPage([character({ link: 'https://exemplo.com/char/123' })]);
		expect(html).toMatch(
			/<a[^>]*href="https:\/\/exemplo\.com\/char\/123"[^>]*target="_blank"[^>]*rel="noopener noreferrer"/
		);
	});

	it('RN-16: sem link, a carta não tem link externo', () => {
		expect(renderPage([character({})])).not.toContain('target="_blank"');
	});

	it('CA-02.8: com 10 personagens, "Adicionar personagem" fica desabilitado com a dica', () => {
		const ten = Array.from({ length: 10 }, (_, i) => character({ id: `c${i}`, nick: `Char${i}` }));
		const html = renderPage(ten);
		// O botão desabilitado (disabled de verdade, não o aria-disabled de "Criar lobby").
		const button = html.match(
			/<button[^>]*\sdisabled=""[^>]*aria-describedby="([^"]+)"[^>]*>(?:(?!<\/button>)[\s\S])*Adicionar personagem/
		);
		expect(button).not.toBeNull();
		expect(html).toContain(`id="${button?.[1]}">Você já tem 10 personagens`);
	});

	it('RNF-01: o menu de cada carta tem rótulo com o nick', () => {
		const html = renderPage([character({ nick: 'Brasa' })]);
		expect(html).toContain('aria-label="Ações de Brasa"');
		expect(html).toContain('aria-haspopup="menu"');
	});

	it('borda de RN-06: classe fora do catálogo aparece pelo id', () => {
		expect(renderPage([character({ classId: 'classe-removida' })])).toContain('classe-removida');
	});

	it('RN-21: API fora do ar mostra o aviso', () => {
		const html = renderPage([], 'Não foi possível falar com o servidor. Tente de novo.');
		expect(html).toContain('role="alert"');
		expect(html).not.toContain('Você ainda não tem personagens');
	});
});

describe('diálogo de personagem renderizado no servidor', () => {
	const renderDialog = (props: Record<string, unknown>) =>
		render(CharacterDialog, {
			props: { mode: 'create', classes: CLASS_LIST, onclose: () => {}, ...props } as never
		}).body;

	it('RN-06 / RN-20: o select tem as 82 classes, agrupadas pela linha, sem repetir grupo', () => {
		const html = renderDialog({});
		expect(html.match(/<option value="[a-z0-9-]+"/g)).toHaveLength(82);
		const groups = [...html.matchAll(/<optgroup label="([^"]+)"/g)].map((m) => m[1]);
		expect(new Set(groups).size).toBe(groups.length);
	});

	it('CA-02.2 / RN-10: os 4 retratos, com o primeiro marcado por padrão', () => {
		const html = renderDialog({});
		expect(html.match(/name="portrait"/g)).toHaveLength(4);
		expect(html).toMatch(/value="retrato-1"[^>]*checked/);
	});

	it('CA-02.9 / RN-19: o erro aparece junto do campo e os valores voltam preenchidos', () => {
		const html = renderDialog({
			form: {
				values: {
					nick: 'Brasa',
					classId: 'paladino',
					level: '99',
					role: 'tank',
					portrait: 'retrato-4',
					link: 'https://x.com'
				},
				errors: { nick: 'Esse nick já está em uso' }
			}
		});
		expect(html).toMatch(/name="nick"[^>]*value="Brasa"|value="Brasa"[^>]*name="nick"/);
		expect(html).toMatch(/aria-invalid="true"[^>]*aria-describedby="([^"]+-nick-err)"/);
		expect(html).toContain('Esse nick já está em uso');
		expect(html).toMatch(/<option value="paladino" selected/);
		expect(html).toMatch(/value="retrato-4"[^>]*checked/);
		expect(html).toMatch(/value="tank"[^>]*checked/);
		expect(html).toContain('value="https://x.com"');
	});

	it('CA-03.1: na edição, os campos começam com o personagem', () => {
		const html = renderDialog({
			mode: 'update',
			character: character({ nick: 'Faísca', classId: 'feiticeiro', level: 165, role: 'dps' })
		});
		expect(html).toContain('Editar Faísca');
		expect(html).toContain('value="Faísca"');
		expect(html).toMatch(/<option value="feiticeiro" selected/);
		expect(html).toContain('value="165"');
		expect(html).toContain('name="id"');
	});
});
