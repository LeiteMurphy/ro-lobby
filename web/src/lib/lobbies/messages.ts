// Mensagens em pt-BR para os erros de lobby que a API devolve por código (spec lobbies,
// design D-07).
import type { ApiErrorCode, FieldError } from '$lib/api/request';

type Field = FieldError['field'];
type Code = FieldError['code'];

const MESSAGES: Partial<Record<Field, Partial<Record<Code, string>> & { fallback: string }>> = {
	instanceId: { fallback: 'Escolha uma instância da lista' },
	startsAt: {
		conflict: 'Esse personagem já está num grupo nesse horário',
		fallback: 'Escolha um horário no futuro, em até 14 dias'
	},
	slots: {
		below_occupied: 'Essa função já tem ocupante',
		fallback: 'De 1 a 12 vagas, com pelo menos 1 na função do seu personagem'
	},
	minLevel: {
		above_owner: 'Seu personagem precisa ter o nível mínimo',
		fallback: 'Entre o nível da instância e 275'
	},
	characterId: {
		level_too_low: 'Seu personagem precisa ter o nível mínimo',
		fallback: 'Escolha um dos seus personagens'
	},
	note: { fallback: 'Use até 250 caracteres' },
	reason: { fallback: 'Escreva de 10 a 250 caracteres' }
};

export const LOBBY_LIMIT_MESSAGE = 'Você já tem 5 lobbies abertos';
export const LOBBY_NOT_OPEN_MESSAGE = 'Esse lobby já começou ou foi cancelado.';
export const NO_CHARACTER_MESSAGE = 'Cadastre um personagem para criar lobbies';

export function lobbyFieldMessage({ field, code }: FieldError): string {
	const messages = MESSAGES[field];
	return messages?.[code] ?? messages?.fallback ?? 'Valor inválido';
}

/** Erros da API como `{campo: mensagem}`, com a primeira mensagem de cada campo. */
export function lobbyFieldMessages(fields: FieldError[]): Partial<Record<Field, string>> {
	const out: Partial<Record<Field, string>> = {};
	for (const f of fields) out[f.field] ??= lobbyFieldMessage(f);
	return out;
}

/** Mensagem de um 409 de lobby. */
export function lobbyConflictMessage(code: ApiErrorCode | undefined): string | null {
	if (code === 'lobby_limit') return LOBBY_LIMIT_MESSAGE;
	if (code === 'lobby_not_open') return LOBBY_NOT_OPEN_MESSAGE;
	return null;
}
