// Mensagens em pt-BR para os erros que a API devolve por código (spec personagens, RN-19,
// design D-07). A API não carrega texto de tela.
import type { FieldError } from './api';

type Field = FieldError['field'];
type Code = FieldError['code'];

// Só os campos do personagem; os do lobby têm as mensagens em `$lib/lobbies/messages`.
type CharacterField = 'nick' | 'classId' | 'level' | 'role' | 'portrait' | 'link' | 'server';

const MESSAGES: Record<CharacterField, Partial<Record<Code, string>> & { fallback: string }> = {
	nick: {
		// RN-03 da spec servidor: o nick é único em cada servidor.
		taken: 'Esse nick já está em uso nesse servidor',
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
	},
	// RN-02 da spec servidor.
	server: { fallback: 'Escolha o servidor' }
};

export const LIMIT_MESSAGE = 'Você já tem 10 personagens';
/**
 * RN-21 da spec lobbies, RN-25/RN-26 da candidatura-lobby e RN-04 da servidor: personagem
 * num lobby aberto.
 */
export const IN_OPEN_LOBBY_MESSAGE =
	'Esse personagem está num lobby aberto. Saia ou cancele antes de mudar servidor, nível ou função.';
export const UNAVAILABLE_MESSAGE = 'Não foi possível falar com o servidor. Tente de novo.';
/** RN-15 da spec endurecimento: a API recusou por excesso de requisições (429). */
export const RATE_LIMITED_MESSAGE = 'Muitas tentativas. Aguarde um pouco e tente de novo.';

/** A mensagem de uma falha sem tratamento próprio: excesso de requisições ou API fora. */
export function unavailableMessage(result: { kind?: string }): string {
	return result.kind === 'rate_limited' ? RATE_LIMITED_MESSAGE : UNAVAILABLE_MESSAGE;
}

/** Erro de carregamento de uma página: null se todas as chamadas deram certo. */
export function loadErrorMessage(...results: { ok: boolean; kind?: string }[]): string | null {
	const failed = results.find((r) => !r.ok);
	return failed ? unavailableMessage(failed) : null;
}
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
