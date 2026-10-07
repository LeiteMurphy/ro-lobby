// Mensagens em pt-BR das regras de candidatura que a API devolve por código (spec
// candidatura-lobby, design D-07).
import type { ApplicationRuleCode, FieldError } from '$lib/api/request';
import type { ApplicationStatus } from './api';

const RULES: Record<ApplicationRuleCode, string> = {
	not_open: 'Esse lobby já começou ou foi cancelado.',
	own_lobby: 'Você é o anfitrião deste lobby.',
	already_active: 'Você já tem uma candidatura ativa neste lobby.',
	role_full: 'Essa função não tem mais vaga.',
	rejected_before: 'Sua candidatura a este lobby já foi recusada.',
	below_min_level: 'O personagem está abaixo do nível mínimo do lobby.',
	schedule_conflict: 'O personagem já está em outro grupo a menos de 2 h deste horário.',
	not_pending: 'Essa candidatura não está mais pendente.',
	not_owner: 'Só o anfitrião decide as candidaturas.',
	not_yours: 'Essa candidatura não é sua.'
};

export function ruleMessage(code: ApplicationRuleCode | undefined): string | null {
	return code ? RULES[code] : null;
}

/** Erros de campo da candidatura e da recusa (RN-06, RN-09). */
export function applicationFieldMessage({ field, code }: FieldError): string {
	if (field === 'message') return 'Use até 250 caracteres';
	if (field === 'characterId') return 'Escolha um dos seus personagens';
	if (field === 'reason') {
		return code === 'required' ? 'Escreva a justificativa' : 'Escreva de 10 a 250 caracteres';
	}
	return 'Valor inválido';
}

export function applicationFieldMessages(fields: FieldError[]): Record<string, string> {
	const out: Record<string, string> = {};
	for (const f of fields) out[f.field] ??= applicationFieldMessage(f);
	return out;
}

export const STATUS_LABELS: Record<ApplicationStatus, string> = {
	pending: 'Pendente',
	accepted: 'Aceita',
	rejected: 'Recusada',
	withdrawn: 'Retirada',
	expired: 'Expirada',
	left: 'Saiu',
	removed: 'Removida',
	cancelled: 'Cancelada'
};
