// Mensagens em pt-BR da sala do grupo no Discord (spec discord-bot). A API não carrega texto
// de tela.
import type { DiscordRuleCode } from './api';

/** Resultado da conexão ao servidor, na query `?discord=` do /perfil (US-01). */
export const CONNECT_MESSAGES = {
	conectado: 'Você entrou no servidor do RO Lobby.',
	conta: 'Use a mesma conta do Discord do seu login.',
	cancelado: 'Conexão cancelada.',
	erro: 'Não foi possível conectar agora. Tente de novo em instantes.',
	desligado: 'A sala no Discord está desligada no momento.'
} as const;

export type ConnectOutcome = keyof typeof CONNECT_MESSAGES;

export function connectMessage(value: string | null): { text: string; ok: boolean } | null {
	if (!value || !(value in CONNECT_MESSAGES)) return null;
	return { text: CONNECT_MESSAGES[value as ConnectOutcome], ok: value === 'conectado' };
}

/** RN-20 e regras da abertura da sala. */
export const OPEN_ROOM_MESSAGES: Record<string, string> = {
	unavailable: 'Não foi possível abrir a sala agora. Tente de novo em instantes.',
	guild_full: 'O servidor do RO Lobby está cheio agora. Tente mais tarde.',
	disabled: 'A sala no Discord está desligada no momento.',
	not_owner: 'Só o anfitrião abre a sala.',
	not_open: 'Esse lobby já começou ou foi cancelado.',
	room_exists: 'A sala desse lobby já foi aberta.'
};

/** RN-26 a RN-28: o aviso no Discord. */
export const ALERT_MESSAGES: Record<string, string> = {
	unavailable: 'Não foi possível avisar agora. Tente de novo em instantes.',
	disabled: 'A sala no Discord está desligada no momento.',
	not_owner: 'Só o anfitrião avisa o grupo.',
	no_room: 'Abra a sala antes de avisar o grupo.'
};

export function ruleMessage(
	table: Record<string, string>,
	kind: string,
	code?: DiscordRuleCode
): string {
	return (code && table[code]) || table[kind] || table.unavailable;
}

/** RN-27: "Aguarde 6 min para avisar de novo", arredondando para cima. */
export function cooldownMessage(seconds: number): string {
	return `Aguarde ${Math.max(1, Math.ceil(seconds / 60))} min para avisar de novo`;
}
