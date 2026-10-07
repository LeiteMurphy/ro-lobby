export type Role = 'tank' | 'support' | 'dps';

/** Ordem fixa das funções: vale para exibição e para desempate (RN-14). */
export const ROLES: readonly Role[] = ['tank', 'support', 'dps'];

export const ROLE_LABELS: Record<Role, string> = {
	tank: 'Tank',
	support: 'Suporte',
	dps: 'Dano'
};

export interface Slots {
	filled: number;
	total: number;
}

export type Composition = Record<Role, Slots>;

export interface Lobby {
	id: string;
	/** Dia do lobby em America/Sao_Paulo, no formato YYYY-MM-DD (RN-09). */
	date: string;
	/** Horário de início em America/Sao_Paulo, no formato HH:MM. */
	time: string;
	instance: string;
	host: string;
	hostClass: string;
	minLevel: number;
	composition: Composition;
	/** Usuário dono do lobby, para o selo de pendentes (RN-34 da candidatura-lobby). */
	ownerId?: string;
	/** Candidaturas pendentes (RN-28 da candidatura-lobby). */
	pendingCount?: number;
}
