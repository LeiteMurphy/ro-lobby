import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import LobbyForm from './components/LobbyForm.svelte';
import { readLobbyForm, toLobbyInput, toLobbyUpdate, type LobbyFormValues } from './form';
import { lobbyFieldMessages } from './messages';

// Link do Discord do dono no formulário do lobby (spec discord-bot, T-11).

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
	title: '',
	discordLink: ''
};

const renderForm = (values: Partial<LobbyFormValues>, errors: Record<string, string> = {}) =>
	render(LobbyForm, {
		props: {
			mode: 'create',
			instances: [
				{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei', level: 160, reset: 'daily' }
			],
			characters: [],
			classes: [],
			days: [{ date: '2026-10-07', label: 'qua, 7 out' }],
			values: { ...VALUES, ...values },
			errors,
			cancelHref: '/'
		} as never
	}).body;

describe('link do Discord do dono no formulário', () => {
	it('CA-07.1 / RN-23: o campo é opcional, com a ajuda "Só os membros aceitos veem"', () => {
		const html = renderForm({});
		expect(html).toContain('Link do seu Discord (opcional)');
		expect(html).toMatch(/name="discordLink"/);
		expect(html).toContain('Só os membros aceitos veem.');
	});

	it('CA-07.1: o link vai para a API ao criar; vazio não vai', () => {
		const form = new FormData();
		for (const [k, v] of Object.entries({ ...VALUES, discordLink: ' https://discord.gg/abc123 ' }))
			form.set(k, v);
		const parsed = toLobbyInput(readLobbyForm(form));
		expect(parsed.ok && parsed.input.discordLink).toBe('https://discord.gg/abc123');
		const empty = toLobbyInput(VALUES);
		expect(empty.ok && 'discordLink' in empty.input).toBe(false);
	});

	it('CA-07.3: na edição, apagar o link manda sem o campo, e a API apaga', () => {
		const parsed = toLobbyUpdate({ ...VALUES, discordLink: '' });
		expect(parsed.ok && 'discordLink' in parsed.input).toBe(false);
	});

	it('CA-07.2: o erro "Use um link de convite do Discord" aparece junto do campo', () => {
		const errors = lobbyFieldMessages([{ field: 'discordLink', code: 'invalid' }]);
		expect(errors.discordLink).toBe('Use um link de convite do Discord');
		const html = renderForm(
			{ discordLink: 'https://exemplo.com' },
			errors as Record<string, string>
		);
		expect(html).toMatch(/aria-describedby="[^"]+-discordLink-err"/);
		expect(html).toContain('Use um link de convite do Discord');
		expect(html).toContain('value="https://exemplo.com"');
	});
});
