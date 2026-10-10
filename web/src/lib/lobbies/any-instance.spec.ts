import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import type { Character } from '$lib/characters/api';
import { applyFilters, instanceOptions, NO_FILTERS } from '$lib/home/filters';
import type { Lobby } from '$lib/home/types';
import LobbyForm from './components/LobbyForm.svelte';
import { TEMPLE } from './fixtures';
import {
	ANY_INSTANCE_NAME,
	NO_INSTANCE,
	readLobbyForm,
	toLobbyInput,
	toLobbyUpdate,
	type LobbyFormValues
} from './form';
import { lobbyFieldMessages } from './messages';
import { inviteText } from './share';
import { toHomeLobby } from './toHome';
import type { ApiLobby } from './api';

// Lobby sem instância no formulário, na Home e no convite (spec lobby-sem-instancia, T-02).

const VALUES: LobbyFormValues = {
	instanceId: 'templo-do-demonio-rei',
	date: '2026-10-07',
	time: '20:00',
	tank: '1',
	support: '2',
	dps: '3',
	minLevel: '160',
	characterId: 'c1',
	note: '',
	formation: 'roles',
	freeSlots: '12',
	title: ''
};
const MVP: ApiLobby = {
	...TEMPLE,
	instance: { id: null, name: 'Caça ao MVP', level: 1, reset: null },
	minLevel: 1
};
const LIRIEN: Character = {
	id: 'c1',
	nick: 'Lirien',
	classId: 'arcebispo',
	level: 178,
	role: 'support',
	portrait: 'retrato-1',
	link: null,
	isMain: true,
	createdAt: '',
	availability: null
};

const formData = (v: LobbyFormValues) => {
	const d = new FormData();
	for (const [k, val] of Object.entries(v)) d.set(k, val);
	return d;
};

describe('formulário sem instância', () => {
	it('CA-01.1 / RN-01 / RN-02: manda anyInstance e o título, sem instanceId', () => {
		const v = readLobbyForm(
			formData({ ...VALUES, instanceId: NO_INSTANCE, title: '  Caça ao MVP  ' })
		);
		const parsed = toLobbyInput(v);
		expect(parsed.ok && parsed.input).toMatchObject({ anyInstance: true, title: 'Caça ao MVP' });
		expect(parsed.ok && 'instanceId' in parsed.input).toBe(false);
		const upd = toLobbyUpdate(v);
		expect(upd.ok && upd.input).toMatchObject({ anyInstance: true, title: 'Caça ao MVP' });
	});

	it('CA-01.5 / RN-01: com instância, manda o instanceId como antes', () => {
		const parsed = toLobbyInput(readLobbyForm(formData(VALUES)));
		expect(parsed.ok && parsed.input).toMatchObject({ instanceId: 'templo-do-demonio-rei' });
		expect(parsed.ok && 'anyInstance' in parsed.input).toBe(false);
	});

	const renderForm = (values: LobbyFormValues, errors = {}) =>
		render(LobbyForm, {
			props: {
				mode: 'create',
				instances: [{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei', level: 160 }],
				characters: [LIRIEN],
				classes: [],
				days: [{ date: '2026-10-07', label: 'qua, 7 out' }],
				values,
				errors,
				cancelHref: '/'
			} as never
		}).body;

	it('CA-01.5 / RNF-03: a instância do catálogo vem marcada e não há campo de título', () => {
		const html = renderForm(VALUES);
		expect(html).toMatch(/<option value="templo-do-demonio-rei"[^>]*selected/);
		expect(html).toContain(`<option value="${NO_INSTANCE}"`);
		expect(html).not.toContain('name="title"');
	});

	it('CA-01.1 / CA-01.2 / CA-01.4 / RNF-03: sem instância, o título com rótulo e a prévia com o nome', () => {
		const html = renderForm({
			...VALUES,
			instanceId: NO_INSTANCE,
			minLevel: '1',
			title: 'Caça ao MVP'
		});
		expect(html).toMatch(/<label[^>]*>Título \(opcional\)<\/label>/);
		expect(html).toMatch(/name="title"[^>]*maxlength="40"|maxlength="40"[^>]*name="title"/);
		expect(html).toContain('Caça ao MVP');
		expect(html).toMatch(/name="minLevel"[^>]*min="1"|min="1"[^>]*name="minLevel"/);
		const empty = renderForm({ ...VALUES, instanceId: NO_INSTANCE, minLevel: '1' });
		expect(empty).toContain(ANY_INSTANCE_NAME);
	});

	it('CA-01.3 / RN-02: título longo volta com a mensagem', () => {
		const errors = lobbyFieldMessages([{ field: 'title', code: 'too_long' }]);
		expect(errors).toEqual({ title: 'Use até 40 caracteres' });
		expect(renderForm({ ...VALUES, instanceId: NO_INSTANCE }, errors)).toContain(
			'Use até 40 caracteres'
		);
	});
});

describe('lobby sem instância na Home e no convite', () => {
	it('CA-01.1 / RN-04: o convite usa o título', () => {
		expect(inviteText(MVP, 'https://rolobby.com.br').split('\n')[0]).toMatch(
			/^Grupo para Caça ao MVP · /
		);
	});

	it('CA-02.1 / RN-05: "Sem instância definida" traz só o sem instância; a instância, só ela', () => {
		const mvp = toHomeLobby(MVP, new Map());
		const temple = toHomeLobby(TEMPLE, new Map());
		expect(mvp).toMatchObject({ anyInstance: true, instance: 'Caça ao MVP' });
		const day: Lobby[] = [mvp, temple];
		const by = (instance: string) =>
			applyFilters(day, { ...NO_FILTERS, instance }).map((l) => l.instance);
		expect(by(NO_INSTANCE)).toEqual(['Caça ao MVP']);
		expect(by('Templo do Demônio Rei')).toEqual(['Templo do Demônio Rei']);
		expect(by('')).toHaveLength(2);
		// O título livre não vira opção de instância, mesmo com o nome igual.
		expect(by('Caça ao MVP')).toEqual([]);
		expect(instanceOptions(day)).toEqual(['Templo do Demônio Rei']);
	});
});
