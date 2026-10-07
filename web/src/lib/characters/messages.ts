// Mensagens em pt-BR para os erros que a API devolve por código (spec personagens, RN-19,
// design D-07). A API não carrega texto de tela.
import type { FieldError } from './api';

type Field = FieldError['field'];
type Code = FieldError['code'];

// Só os campos do personagem; os do lobby têm as mensagens em `$lib/lobbies/messages`.
type CharacterField = 'nick' | 'classId' | 'level' | 'role' | 'portrait' | 'link';

const MESSAGES: Record<CharacterField, Partial<Record<Code, string>> & { fallback: string }> = {
	nick: {
		taken: 'Esse nick já está em uso',
		required: 'Informe o nick',
		too_long: 'Use até 24 caracteres',
		fallback: 'Use só letras, números e símbolos visíveis'
	},
	classId: { fallback: 'Escolha uma classe da lista' },
	level: { fallback: 'O nível vai de 1 a 275' },
	role: { fallback: 'Escolha a função' },
	portrait: { fallback: 'Escolha um retrato da lista' },
	link: {
		too_long: 'Use um link de até 300 caracteres',
		fallback: 'Use um link https:// válido'
	}
};

export const LIMIT_MESSAGE = 'Você já tem 10 personagens';
/** RN-21 da spec lobbies e RN-25/RN-26 da candidatura-lobby: personagem num lobby aberto. */
export const IN_OPEN_LOBBY_MESSAGE =
	'Esse personagem está num lobby aberto. Saia ou cancele antes de mudar nível ou função.';
export const UNAVAILABLE_MESSAGE = 'Não foi possível falar com o servidor. Tente de novo.';
export const NOT_FOUND_MESSAGE = 'Esse personagem não existe mais.';

/** Mensagem do erro de um campo. */
export function fieldMessage({ field, code }: FieldError): string {
	const messages = MESSAGES[field as CharacterField];
	return messages?.[code] ?? messages?.fallback ?? 'Valor inválido';
}

/** Erros da API como `{campo: mensagem}`, com a primeira mensagem de cada campo. */
export function fieldMessages(fields: FieldError[]): Partial<Record<Field, string>> {
	const out: Partial<Record<Field, string>> = {};
	for (const f of fields) out[f.field] ??= fieldMessage(f);
	return out;
}
