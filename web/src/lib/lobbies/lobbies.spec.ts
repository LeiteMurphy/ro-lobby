import { describe, expect, it, vi } from 'vitest';
import { createLobby, cancelLobby, getLobby, listInstances, listLobbies, updateLobby } from './api';
import { createLobbyHref, defaultStart, fromUtcIso, isTime, toUtcIso } from './time';
import { TEMPLE } from './fixtures';
import { DELETED_CHARACTER, toHomeLobby } from './toHome';

const base = 'http://api.test';

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

describe('horário de Brasília (D-04)', () => {
	it.each([
		['2026-10-07', '20:00', '2026-10-07T23:00:00Z'],
		['2026-10-06', '23:30', '2026-10-07T02:30:00Z'],
		['2026-10-06', '00:00', '2026-10-06T03:00:00Z'],
		['2026-12-31', '22:15', '2027-01-01T01:15:00Z']
	])('RN-05: %s %s em Brasília é %s em UTC', (date, time, utc) => {
		expect(toUtcIso({ date, time })).toBe(utc);
		expect(fromUtcIso(utc)).toEqual({ date, time });
	});

	it('borda de RN-05: 02:30 UTC de amanhã ainda é hoje, 23:30, em São Paulo', () => {
		expect(fromUtcIso('2026-10-07T02:30:00Z')).toEqual({ date: '2026-10-06', time: '23:30' });
	});

	it('RNF-04: só aceita HH:MM válido', () => {
		for (const ok of ['00:00', '09:05', '23:59']) expect(isTime(ok)).toBe(true);
		for (const bad of ['24:00', '9:05', '12:60', '', 'agora']) expect(isTime(bad)).toBe(false);
	});
});

describe('lobby da API para a Home (RN-22, D-09)', () => {
	const classNames = new Map([['arcebispo', 'Arcebispo']]);

	it('CA-02.1: dia e hora em Brasília, anfitrião, classe, nível e composição', () => {
		expect(toHomeLobby(TEMPLE, classNames)).toEqual({
			id: TEMPLE.id,
			date: '2026-10-07',
			time: '20:00',
			instance: 'Templo do Demônio Rei',
			host: 'Lirien',
			hostClass: 'Arcebispo',
			minLevel: 160,
			composition: {
				tank: { filled: 0, total: 1 },
				support: { filled: 1, total: 2 },
				dps: { filled: 0, total: 3 }
			}
		});
	});

	it('D-01: personagem excluído aparece como "Personagem excluído"; classe fora do catálogo pelo id', () => {
		const gone = toHomeLobby(
			{
				...TEMPLE,
				owner: { ...TEMPLE.owner, characterId: null, nick: null, classId: null, level: null }
			},
			classNames
		);
		expect(gone.host).toBe(DELETED_CHARACTER);
		expect(gone.hostClass).toBe('');
		const removed = toHomeLobby(
			{ ...TEMPLE, owner: { ...TEMPLE.owner, classId: 'classe-removida' } },
			classNames
		);
		expect(removed.hostClass).toBe('classe-removida');
	});
});

describe('cliente da API de lobbies (D-06)', () => {
	it('RN-22: lista e detalhe são pedidos sem token', async () => {
		const f = fakeFetch(200, [TEMPLE]);
		expect(await listLobbies(f, base, '2026-10-06', '2026-10-19')).toEqual({
			ok: true,
			data: [TEMPLE]
		});
		expect(String(f.mock.calls[0][0])).toBe(
			'http://api.test/lobbies?from=2026-10-06&to=2026-10-19'
		);
		expect(new Headers(f.mock.calls[0][1]?.headers).has('authorization')).toBe(false);

		const g = fakeFetch(200, TEMPLE);
		await getLobby(g, base, 'a/b');
		expect(String(g.mock.calls[0][0])).toBe('http://api.test/lobbies/a%2Fb');

		const h = fakeFetch(200, []);
		await listInstances(h, base);
		expect(String(h.mock.calls[0][0])).toBe('http://api.test/instances');
	});

	it('RN-04: criar, editar e cancelar mandam o token e o corpo', async () => {
		const f = fakeFetch(200, TEMPLE);
		const call = { fetchFn: f, apiBaseUrl: base, token: 'tok' };
		const input = {
			instanceId: 'templo-do-demonio-rei',
			startsAt: '2026-10-07T23:00:00Z',
			slots: { tank: 1, support: 2, dps: 3 },
			minLevel: 160,
			characterId: 'c1'
		};
		await createLobby(call, input);
		await updateLobby(call, 'l1', { startsAt: input.startsAt, slots: input.slots, minLevel: 170 });
		await cancelLobby(call, 'l1', 'Metade do grupo não pode');
		expect(f.mock.calls.map(([u, i]) => `${i?.method} ${String(u)}`)).toEqual([
			'POST http://api.test/lobbies',
			'PUT http://api.test/lobbies/l1',
			'POST http://api.test/lobbies/l1/cancel'
		]);
		for (const [, init] of f.mock.calls)
			expect(new Headers(init?.headers).get('authorization')).toBe('Bearer tok');
		expect(JSON.parse(String(f.mock.calls[2][1]?.body))).toEqual({
			reason: 'Metade do grupo não pode'
		});
	});

	it.each([
		[409, { error: 'lobby_limit' }, { ok: false, kind: 'conflict', code: 'lobby_limit' }],
		[409, { error: 'lobby_not_open' }, { ok: false, kind: 'conflict', code: 'lobby_not_open' }],
		[400, { error: 'invalid_range' }, { ok: false, kind: 'bad_request', code: 'invalid_range' }],
		[404, { error: 'not_found' }, { ok: false, kind: 'not_found' }]
	])('RN-11 / D-08: status %i vira %o', async (status, body, want) => {
		expect(await listLobbies(fakeFetch(status, body), base, 'a', 'b')).toEqual(want);
	});
});

describe('início padrão da criação', () => {
	const DAYS = ['2026-10-06', '2026-10-07', '2026-10-08', '2026-10-09'];
	const at = (h: number, m: number) => ({ date: '2026-10-06', minutes: h * 60 + m });

	it('CA-02.5 / RN-24: o dia escolhido vem com 20:00; sem dia ou fora dos 14, amanhã', () => {
		expect(defaultStart('2026-10-09', at(16, 40), DAYS)).toEqual({
			date: '2026-10-09',
			time: '20:00'
		});
		expect(defaultStart(null, at(16, 40), DAYS)).toEqual({ date: '2026-10-07', time: '20:00' });
		expect(defaultStart('2026-12-25', at(16, 40), DAYS)).toEqual({
			date: '2026-10-07',
			time: '20:00'
		});
	});

	it('CA-02.6 / RN-24: hoje, a hora nunca fica no passado', () => {
		expect(defaultStart('2026-10-06', at(16, 40), DAYS)).toEqual({
			date: '2026-10-06',
			time: '20:00'
		});
		expect(defaultStart('2026-10-06', at(19, 59), DAYS)).toEqual({
			date: '2026-10-06',
			time: '20:00'
		});
		expect(defaultStart('2026-10-06', at(20, 0), DAYS)).toEqual({
			date: '2026-10-06',
			time: '21:00'
		});
		expect(defaultStart('2026-10-06', at(20, 40), DAYS)).toEqual({
			date: '2026-10-06',
			time: '21:00'
		});
		expect(defaultStart('2026-10-06', at(22, 59), DAYS)).toEqual({
			date: '2026-10-06',
			time: '23:00'
		});
		expect(defaultStart('2026-10-06', at(23, 10), DAYS)).toEqual({
			date: '2026-10-07',
			time: '20:00'
		});
	});

	it('RN-23 / RN-24: o link de "Criar lobby" leva o dia quando há', () => {
		expect(createLobbyHref('2026-10-09')).toBe('/lobbies/novo?dia=2026-10-09');
		expect(createLobbyHref(null)).toBe('/lobbies/novo');
	});
});
