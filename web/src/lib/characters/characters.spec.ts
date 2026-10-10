import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { describe, expect, it, vi } from 'vitest';
import {
	createCharacter,
	deleteCharacter,
	listCharacters,
	listClasses,
	setMainCharacter,
	updateCharacter,
	type CharacterInput
} from './api';
import { fieldMessage, fieldMessages, LIMIT_MESSAGE } from './messages';
import {
	CLASS_ARTS,
	DEFAULT_PORTRAIT,
	GENERIC_PORTRAITS,
	portraitForClass,
	portraitInfo
} from './portraits';

const base = 'http://api.test';
const input: CharacterInput = { nick: 'Brasa', classId: 'guardiao-real', level: 172, role: 'tank' };

function fakeFetch(status: number, body?: unknown) {
	return vi.fn<typeof fetch>(async () =>
		body === undefined
			? new Response(null, { status })
			: new Response(JSON.stringify(body), {
					status,
					headers: { 'content-type': 'application/json' }
				})
	);
}

describe('cliente da API de personagens', () => {
	it('RN-01 / D-05: manda o token como Bearer e o corpo em JSON', async () => {
		const f = fakeFetch(201, { id: '1', nick: 'Brasa' });
		const result = await createCharacter({ fetchFn: f, apiBaseUrl: base, token: 'tok' }, input);
		expect(result).toEqual({ ok: true, data: { id: '1', nick: 'Brasa' } });
		const [url, init] = f.mock.calls[0];
		expect(String(url)).toBe('http://api.test/characters');
		expect(init?.method).toBe('POST');
		expect(new Headers(init?.headers).get('authorization')).toBe('Bearer tok');
		expect(JSON.parse(String(init?.body))).toEqual(input);
	});

	it('CA-07.1: o catálogo é pedido sem token', async () => {
		const f = fakeFetch(200, [{ id: 'aprendiz' }]);
		const result = await listClasses(f, base);
		expect(result).toEqual({ ok: true, data: [{ id: 'aprendiz' }] });
		expect(new Headers(f.mock.calls[0][1]?.headers).has('authorization')).toBe(false);
	});

	it('D-01 / RN-02: monta as rotas de editar, excluir e principal com o id escapado', async () => {
		const f = fakeFetch(204);
		const call = { fetchFn: f, apiBaseUrl: base, token: 'tok' };
		await deleteCharacter(call, 'abc');
		await setMainCharacter(call, 'abc');
		expect(f.mock.calls.map(([u, i]) => `${i?.method} ${String(u)}`)).toEqual([
			'DELETE http://api.test/characters/abc',
			'PUT http://api.test/characters/abc/main'
		]);
		const g = fakeFetch(200, {});
		await updateCharacter({ ...call, fetchFn: g }, 'a/b', input);
		expect(String(g.mock.calls[0][0])).toBe('http://api.test/characters/a%2Fb');
	});

	it.each([
		[401, undefined, { ok: false, kind: 'no_session' }],
		[404, { error: 'not_found' }, { ok: false, kind: 'not_found' }],
		[409, { error: 'character_limit' }, { ok: false, kind: 'conflict', code: 'character_limit' }],
		[500, undefined, { ok: false, kind: 'unavailable' }]
	])('RN-01 / RN-02 / CA-02.8 / RN-21: status %i vira %o', async (status, body, want) => {
		const result = await listCharacters({
			fetchFn: fakeFetch(status, body),
			apiBaseUrl: base,
			token: 't'
		});
		expect(result).toEqual(want);
	});

	it('RN-19: 422 repassa os erros de campo', async () => {
		const fields = [{ field: 'nick', code: 'taken' }];
		const result = await createCharacter(
			{ fetchFn: fakeFetch(422, { error: 'validation', fields }), apiBaseUrl: base, token: 't' },
			input
		);
		expect(result).toEqual({ ok: false, kind: 'invalid', fields });
	});

	it('RN-21: API fora do ar vira unavailable, sem lançar', async () => {
		const f = vi.fn<typeof fetch>(async () => {
			throw new TypeError('fetch failed');
		});
		expect(await listCharacters({ fetchFn: f, apiBaseUrl: base, token: 't' })).toEqual({
			ok: false,
			kind: 'unavailable'
		});
	});
});

describe('mensagens em pt-BR (D-07)', () => {
	it.each([
		['nick', 'taken', 'Esse nick já está em uso'],
		['nick', 'required', 'Informe o nick'],
		['nick', 'too_long', 'Use até 24 caracteres'],
		['classId', 'invalid', 'Escolha uma classe da lista'],
		['classId', 'required', 'Escolha uma classe da lista'],
		['level', 'invalid', 'O nível vai de 1 a 275'],
		['role', 'invalid', 'Escolha a função'],
		['portrait', 'invalid', 'Escolha um retrato da lista'],
		['link', 'invalid', 'Use um link https:// válido'],
		['link', 'too_long', 'Use um link de até 300 caracteres']
	] as const)('CA-02.9: %s/%s → "%s"', (field, code, message) => {
		expect(fieldMessage({ field, code })).toBe(message);
	});

	it('CA-02.8: mensagem do limite', () => {
		expect(LIMIT_MESSAGE).toBe('Você já tem 10 personagens');
	});

	it('RN-19: agrupa por campo, com a primeira mensagem de cada um', () => {
		expect(
			fieldMessages([
				{ field: 'nick', code: 'taken' },
				{ field: 'level', code: 'invalid' },
				{ field: 'nick', code: 'too_long' }
			])
		).toEqual({ nick: 'Esse nick já está em uso', level: 'O nível vai de 1 a 275' });
	});
});

describe('retratos pela classe (spec retrato-por-classe)', () => {
	const dir = new URL('../../../static/portraits/', import.meta.url);
	const classDir = new URL('classes/', dir);

	it('CA-01.1 / RN-05: a arte da linha, com o nome da linha como texto alternativo', () => {
		expect(portraitInfo('arruaceiro')).toEqual({
			id: 'arruaceiro',
			label: 'Linha do Arruaceiro',
			src: '/portraits/classes/arruaceiro.png'
		});
		expect(portraitInfo('superaprendiz').label).toBe('Linha do Superaprendiz');
		expect(portraitInfo('gatuno').label).toBe('Família Gatuno');
	});

	it('CA-01.6 / RN-04: retrato desconhecido ou vazio vira o Retrato 1', () => {
		expect(portraitInfo('desconhecido')).toMatchObject({
			id: DEFAULT_PORTRAIT,
			src: '/portraits/retrato-1.svg'
		});
		expect(portraitInfo(null).id).toBe('retrato-1');
		expect(portraitInfo('retrato-3').src).toBe('/portraits/retrato-3.svg');
	});

	it('CA-01.4 / RN-01: a arte da classe escolhida vem do catálogo', () => {
		const classes = [{ id: 'renegado', art: 'arruaceiro' }];
		expect(portraitForClass('renegado', classes).label).toBe('Linha do Arruaceiro');
		expect(portraitForClass('', classes).id).toBe('retrato-1');
	});

	it('RNF-01 / RNF-02 / D-04: cada arte tem seu PNG de até 4 KB, e só elas', () => {
		expect(CLASS_ARTS).toHaveLength(27);
		expect(readdirSync(classDir).sort()).toEqual(CLASS_ARTS.map((a) => `${a}.png`).sort());
		for (const a of CLASS_ARTS) {
			const png = readFileSync(new URL(`${a}.png`, classDir));
			expect(png.subarray(1, 4).toString()).toBe('PNG');
			expect(png.length).toBeLessThanOrEqual(4096);
		}
	});

	it('RN-04: os 4 retratos genéricos continuam, sem nada de fora', () => {
		for (const id of GENERIC_PORTRAITS) {
			expect(existsSync(new URL(`${id}.svg`, dir))).toBe(true);
			expect(readFileSync(new URL(`${id}.svg`, dir), 'utf8')).not.toMatch(
				/(href|src)=|url\(|@import/
			);
		}
	});
});
