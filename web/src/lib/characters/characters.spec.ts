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
import { DEFAULT_PORTRAIT, PORTRAITS, portraitInfo } from './portraits';

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

	it('monta as rotas de editar, excluir e principal com o id escapado', async () => {
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
		[409, { error: 'character_limit' }, { ok: false, kind: 'limit' }],
		[500, undefined, { ok: false, kind: 'unavailable' }]
	])('status %i vira %o', async (status, body, want) => {
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

	it('API fora do ar vira unavailable, sem lançar', async () => {
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

describe('retratos (RN-10, D-04, RNF-02)', () => {
	const dir = new URL('../../../static/portraits/', import.meta.url);

	it('CA-02.2: 4 retratos, com o primeiro como padrão', () => {
		expect(PORTRAITS.map((p) => p.id)).toEqual([
			'retrato-1',
			'retrato-2',
			'retrato-3',
			'retrato-4'
		]);
		expect(DEFAULT_PORTRAIT).toBe(PORTRAITS[0].id);
		expect(portraitInfo('desconhecido').id).toBe('retrato-1');
	});

	it('D-04: cada retrato do contrato tem um arquivo em static/portraits/, e só eles', () => {
		for (const p of PORTRAITS) {
			expect(p.src).toBe(`/portraits/${p.id}.svg`);
			expect(existsSync(new URL(`${p.id}.svg`, dir))).toBe(true);
		}
		expect(readdirSync(dir).sort()).toEqual(PORTRAITS.map((p) => `${p.id}.svg`));
	});

	it('RNF-02: os SVGs não carregam nada de fora', () => {
		for (const p of PORTRAITS) {
			const svg = readFileSync(new URL(`${p.id}.svg`, dir), 'utf8');
			expect(svg).not.toMatch(/(href|src)=|url\(|@import/);
		}
	});
});
