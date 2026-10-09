// Mensagens em pt-BR dos erros da disponibilidade e dos filtros (spec banco-de-talentos,
// design D-06). A API não carrega texto de tela.
import type { FieldError } from '$lib/api/request';

type Code = FieldError['code'];
type TalentField = 'days' | 'start' | 'end' | 'instanceIds' | 'day' | 'time';

const MESSAGES: Record<TalentField, Partial<Record<Code, string>> & { fallback: string }> = {
	days: { required: 'Escolha pelo menos um dia', fallback: 'Escolha dias da semana' },
	start: { fallback: 'Escolha um horário de 30 em 30 minutos' },
	end: {
		same_as_start: 'O fim precisa ser diferente do início',
		fallback: 'Escolha um horário de 30 em 30 minutos'
	},
	instanceIds: {
		required: 'Escolha uma instância ou marque Qualquer instância',
		fallback: 'Escolha instâncias da lista'
	},
	day: { fallback: 'Escolha um dia da semana' },
	time: { fallback: 'Escolha um horário de 30 em 30 minutos' }
};

/** Erros da API como `{campo: mensagem}`, com a primeira mensagem de cada campo. */
export function talentFieldMessages(fields: FieldError[]): Partial<Record<string, string>> {
	const out: Partial<Record<string, string>> = {};
	for (const { field, code } of fields) {
		const m = MESSAGES[field as TalentField];
		out[field] ??= m?.[code] ?? m?.fallback ?? 'Valor inválido';
	}
	return out;
}
