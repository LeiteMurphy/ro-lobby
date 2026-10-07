// Modelo do detalhe do lobby com candidaturas: pessoas do painel, composição e quem pode se
// candidatar (spec candidatura-lobby, RN-05, RN-28, RN-30, RN-31, design D-06).
import type { Character } from '$lib/characters/api';
import { ROLE_LABELS, ROLES, type Role } from '$lib/home/types';
import type { ApiLobby } from '$lib/lobbies/api';
import { DELETED_CHARACTER } from '$lib/lobbies/toHome';

/** Uma pessoa que o painel da direita mostra (RN-31). */
export interface Person {
	/** Chave da seleção: "host" ou o id da candidatura. */
	key: string;
	kind: 'host' | 'member' | 'candidate';
	role: Role;
	nick: string;
	classId: string | null;
	level: number | null;
	portrait: string | null;
	link: string | null;
	/** Nulo quando quem olha não pode ver (RN-32). */
	discordName: string | null;
	/** Só para o dono (RN-28). */
	message: string | null;
	applicationId: string | null;
	createdAt: string | null;
}

export const HOST_KEY = 'host';

/** Anfitrião, membros aceitos e, para o dono, os candidatos pendentes. */
export function people(lobby: ApiLobby): Person[] {
	const host: Person = {
		key: HOST_KEY,
		kind: 'host',
		role: lobby.owner.role,
		nick: lobby.owner.nick ?? DELETED_CHARACTER,
		classId: lobby.owner.classId,
		level: lobby.owner.level,
		portrait: lobby.owner.portrait,
		link: lobby.owner.link,
		discordName: lobby.owner.discordName,
		message: null,
		applicationId: null,
		createdAt: null
	};
	const toPerson =
		(kind: 'member' | 'candidate') => (p: NonNullable<ApiLobby['members']>[number]) =>
			({
				key: p.applicationId,
				kind,
				role: p.role,
				nick: p.nick ?? DELETED_CHARACTER,
				classId: p.classId,
				level: p.level,
				portrait: p.portrait,
				link: p.link,
				discordName: p.discordName,
				message: p.message,
				applicationId: p.applicationId,
				createdAt: p.createdAt
			}) satisfies Person;
	return [
		host,
		...(lobby.members ?? []).map(toPerson('member')),
		...(lobby.pending ?? []).map(toPerson('candidate'))
	];
}

export interface CompositionRow {
	role: Role;
	total: number;
	filled: number;
	/** Uma posição por vaga: a pessoa que ocupa ou nulo (vaga aberta). */
	slots: (Person | null)[];
}

/** RN-05 / D-03: as vagas de cada função com o anfitrião e os membros aceitos. */
export function composition(lobby: ApiLobby, list: Person[]): CompositionRow[] {
	return ROLES.map((role) => {
		const occupants = list.filter((p) => p.kind !== 'candidate' && p.role === role);
		const total = lobby.slots[role];
		const slots: (Person | null)[] = Array.from(
			{ length: Math.max(total, occupants.length) },
			(_, i) => (i < occupants.length ? occupants[i] : null)
		);
		return { role, total, filled: lobby.occupied[role], slots };
	});
}

export interface Eligibility {
	character: Character;
	ok: boolean;
	/** Motivo curto ao lado do personagem no diálogo (Candidatura 1c). */
	why: string;
}

/**
 * RN-19 / RN-20 / RN-36 / D-11: para qual personagem dá para trocar. A vaga que sai conta
 * como livre; o atual não entra. O pedido do membro pode esperar a vaga abrir (Candidatura
 * 2i), a troca do dono não (2h).
 */
export function swapEligibility(
	lobby: ApiLobby,
	characters: Character[],
	current: { characterId: string | null; role: Role },
	mode: 'owner' | 'request'
): Eligibility[] {
	return characters
		.filter((c) => c.id !== current.characterId)
		.map((character) => {
			const label = ROLE_LABELS[character.role];
			if (character.level < lobby.minLevel) {
				return { character, ok: false, why: `abaixo do nível ${lobby.minLevel}` };
			}
			if (character.role === current.role) {
				return { character, ok: true, why: `mesma vaga de ${label}` };
			}
			const free = lobby.slots[character.role] - lobby.occupied[character.role];
			if (free < 1) {
				return mode === 'owner'
					? { character, ok: false, why: `${label} sem vaga` }
					: { character, ok: true, why: `${label} sem vaga agora` };
			}
			return {
				character,
				ok: true,
				why: free === 1 ? `1 vaga de ${label}` : `${free} vagas de ${label}`
			};
		});
}

/** RN-05 / RN-30: quais personagens podem se candidatar e por quê. */
export function eligibility(lobby: ApiLobby, characters: Character[]): Eligibility[] {
	return characters.map((character) => {
		const free = lobby.slots[character.role] - lobby.occupied[character.role];
		const label = ROLE_LABELS[character.role];
		if (character.level < lobby.minLevel) {
			return { character, ok: false, why: `abaixo do nível ${lobby.minLevel}` };
		}
		if (free < 1) return { character, ok: false, why: `${label} sem vaga` };
		return {
			character,
			ok: true,
			why: free === 1 ? `1 vaga de ${label}` : `${free} vagas de ${label}`
		};
	});
}

/**
 * O que quem olha pode fazer quanto à candidatura (RN-02, RN-03, RN-07, RN-13, RN-15,
 * RN-37):
 * - `login`: sem sessão;
 * - `owner`: é o anfitrião;
 * - `closed`: o lobby não está aberto;
 * - `apply`: pode se candidatar (também depois de sair ou de ser removido sem bloqueio);
 * - `pending`: tem pendente (pode retirar);
 * - `member`: foi aceito;
 * - `rejected`: foi recusado e não volta;
 * - `blocked`: foi removido com bloqueio e não volta.
 */
export type ApplyState =
	'login' | 'owner' | 'closed' | 'apply' | 'pending' | 'member' | 'rejected' | 'blocked';

export function applyState(lobby: ApiLobby, userId: string | null): ApplyState {
	if (!userId) return lobby.status === 'open' ? 'login' : 'closed';
	if (userId === lobby.owner.userId) return 'owner';
	const mine = lobby.myApplication?.status;
	if (mine === 'accepted') return 'member';
	if (lobby.status !== 'open') return 'closed';
	if (mine === 'pending') return 'pending';
	if (mine === 'rejected') return 'rejected';
	if (mine === 'removed' && lobby.myApplication?.blocked) return 'blocked';
	return 'apply';
}
