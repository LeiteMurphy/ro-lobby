import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import DiscordSection from './components/DiscordSection.svelte';
import type { LobbyDiscord } from './api';

// Seção "Sala no Discord" renderizada no servidor (spec discord-bot, T-10).

const NOW = '2026-10-14T22:00:00.000Z';
const MEMBERS = [
	{ userId: 'u1', nick: 'Lirien', connected: true },
	{ userId: 'u2', nick: 'Fogo', connected: false }
];
const base: LobbyDiscord = {
	enabled: true,
	room: null,
	members: MEMBERS,
	viewerConnected: true,
	ownerLink: null
};
const OPEN = {
	status: 'open' as const,
	url: 'https://discord.com/channels/900/901',
	alertAvailableAt: null
};

const html = (discord: Partial<LobbyDiscord>, props: Record<string, unknown> = {}) =>
	render(DiscordSection, {
		props: {
			discord: { ...base, ...discord },
			isOwner: false,
			lobbyOpen: true,
			now: NOW,
			...props
		} as never
	}).body;

describe('seção Sala no Discord', () => {
	it('CA-03.3 / RN-06: sem sala, o dono vê "Abrir sala no Discord"; o membro, que não foi aberta', () => {
		const owner = html({}, { isOwner: true });
		expect(owner).toContain('action="?/openRoom"');
		expect(owner).toContain('Abrir sala no Discord');
		const member = html({});
		expect(member).toContain('A sala ainda não foi aberta.');
		expect(member).not.toContain('?/openRoom');
	});

	it('CA-03.1 / RN-18: com a sala aberta, o link e quem está na sala', () => {
		const out = html({ room: OPEN });
		expect(out).toMatch(
			/href="https:\/\/discord\.com\/channels\/900\/901"[^>]*target="_blank"[^>]*rel="noopener noreferrer"/
		);
		expect(out).toMatch(/Lirien<\/span> <span class="state[^"]*on[^"]*"[^>]*>na sala/);
		expect(out).toMatch(/Fogo<\/span> <span class="state[^"]*"[^>]*>não conectado/);
	});

	it('CA-03.2 / RN-18: o membro não conectado vê o atalho para se conectar', () => {
		expect(html({ room: OPEN, viewerConnected: false })).toContain(
			'Entrar no servidor do RO Lobby'
		);
		expect(html({ room: OPEN, viewerConnected: true })).not.toContain(
			'Entrar no servidor do RO Lobby'
		);
	});

	it('CA-03.5: com a sala encerrando ou apagada, diz que foi encerrada, sem link nem botão', () => {
		for (const status of ['closing', 'deleted'] as const) {
			const out = html({ room: { status, url: null, alertAvailableAt: null } }, { isOwner: true });
			expect(out).toContain('A sala foi encerrada.');
			expect(out).not.toContain('room-link');
			expect(out).not.toContain('?/openRoom');
			expect(out).not.toContain('?/alert');
		}
	});

	it('CA-08.5 / RN-26: só o dono, com a sala aberta, vê "Avisar no Discord"', () => {
		expect(html({ room: OPEN }, { isOwner: true })).toContain('action="?/alert"');
		expect(html({ room: OPEN })).not.toContain('?/alert');
		expect(html({}, { isOwner: true })).not.toContain('?/alert');
	});

	it('CA-08.2 / RN-27: no intervalo, o botão fica desabilitado com o tempo que falta', () => {
		const out = html(
			{ room: { ...OPEN, alertAvailableAt: '2026-10-14T22:06:00.000Z' } },
			{ isOwner: true }
		);
		expect(out).toMatch(/<button[^>]*disabled[^>]*>[\s\S]*Avisar no Discord/);
		expect(out).toContain('Aguarde 6 min para avisar de novo');
	});

	it('CA-07.4 / CA-07.7 / RN-24: o Discord do anfitrião aparece, mesmo com a sala desligada', () => {
		const out = html({ enabled: false, ownerLink: 'https://discord.gg/abc123' });
		expect(out).toMatch(/data-testid="owner-link"[^>]*>[\s\S]*Discord do anfitrião/);
		expect(out).toContain('href="https://discord.gg/abc123"');
		expect(out).not.toContain('A sala ainda não foi aberta');
	});

	it('RN-22: com a sala desligada e sem link do anfitrião, a seção some', () => {
		expect(html({ enabled: false })).not.toContain('data-testid="discord-section"');
	});

	it('CA-02.7 / CA-02.8: o erro da última ação aparece na seção', () => {
		const out = html(
			{},
			{
				isOwner: true,
				result: { text: 'O servidor do RO Lobby está cheio agora. Tente mais tarde.', ok: false }
			}
		);
		expect(out).toMatch(/role="alert"[^>]*>O servidor do RO Lobby está cheio agora/);
	});
});

describe('seção Sala no Discord em construção (spec discord-em-construcao)', () => {
	it('CA-01.2 / RN-03 / RN-04: só o aviso e o botão desabilitado; o link do dono continua', () => {
		for (const room of [null, OPEN]) {
			const out = html(
				{ room, viewerConnected: false, ownerLink: 'https://discord.gg/abc' },
				{ isOwner: true, underConstruction: true }
			);
			expect(out).toContain('A sala no Discord está em construção.');
			expect(out).toMatch(/<button[^>]*disabled[^>]*>[\s\S]*Em construção/);
			expect(out).not.toContain('?/openRoom');
			expect(out).not.toContain('?/alert');
			expect(out).not.toContain('data-testid="room-link"');
			expect(out).not.toContain('Quem está na sala');
			expect(out).not.toContain('/perfil/discord');
			expect(out).toContain('data-testid="owner-link"');
			expect(out).toContain('https://discord.gg/abc');
		}
	});

	it('CA-01.5 / RN-07: sem a chave, a seção funciona como antes', () => {
		const out = html({}, { isOwner: true, underConstruction: false });
		expect(out).toContain('action="?/openRoom"');
		expect(out).not.toContain('em construção');
	});
});
