// Chamadas do servidor do web à API de candidaturas, pelos tipos gerados do contrato
// (spec candidatura-lobby, design D-06 e D-07).
import { request, type Call } from '$lib/api/request';
import type { components } from '$lib/api/schema.gen';

export type Application = components['schemas']['Application'];
export type ApplicationStatus = components['schemas']['ApplicationStatus'];
export type MyApplication = components['schemas']['MyApplication'];
export type Participant = components['schemas']['LobbyParticipant'];
export type ViewerApplication = components['schemas']['ViewerApplication'];
export type SwapRequest = components['schemas']['SwapRequest'];
export type LobbySwapRequest = components['schemas']['LobbySwapRequest'];
export type { Result } from '$lib/api/request';

const id = (s: string) => encodeURIComponent(s);

/** RN-01 a RN-07, RN-30: candidatura com um dos próprios personagens. */
export function applyToLobby(call: Call, lobbyId: string, characterId: string, message: string) {
	return request<Application>(call, 'POST', `/lobbies/${id(lobbyId)}/applications`, {
		characterId,
		message: message || undefined
	});
}

/** RN-08, RN-10 a RN-12: o dono aceita. */
export function acceptApplication(call: Call, applicationId: string) {
	return request<Application>(call, 'POST', `/applications/${id(applicationId)}/accept`);
}

/** RN-08, RN-09: o dono recusa com justificativa. */
export function rejectApplication(call: Call, applicationId: string, reason: string) {
	return request<Application>(call, 'POST', `/applications/${id(applicationId)}/reject`, {
		reason
	});
}

/** RN-13: o candidato retira a própria pendente. */
export function withdrawApplication(call: Call, applicationId: string) {
	return request<Application>(call, 'POST', `/applications/${id(applicationId)}/withdraw`);
}

/** RN-14: o membro sai do grupo. */
export function leaveLobby(call: Call, applicationId: string) {
	return request<Application>(call, 'POST', `/applications/${id(applicationId)}/leave`);
}

/** RN-15: o dono remove um membro, com justificativa e, se quiser, bloqueio. */
export function removeMember(call: Call, applicationId: string, reason: string, block: boolean) {
	return request<Application>(call, 'POST', `/applications/${id(applicationId)}/remove`, {
		reason,
		block
	});
}

/** RN-19: o dono troca o próprio personagem; volta o detalhe do lobby. */
export function swapOwnerCharacter(call: Call, lobbyId: string, characterId: string) {
	return request<unknown>(call, 'PUT', `/lobbies/${id(lobbyId)}/owner-character`, {
		characterId
	});
}

/** RN-20: o membro pede a troca do personagem, com motivo. */
export function requestSwap(
	call: Call,
	applicationId: string,
	characterId: string,
	reason: string
) {
	return request<SwapRequest>(call, 'POST', `/applications/${id(applicationId)}/swap-requests`, {
		characterId,
		reason
	});
}

/** RN-22, RN-23: o dono aceita o pedido de troca. */
export function acceptSwap(call: Call, swapId: string) {
	return request<SwapRequest>(call, 'POST', `/swap-requests/${id(swapId)}/accept`);
}

/** RN-22: o dono recusa o pedido de troca com justificativa. */
export function rejectSwap(call: Call, swapId: string, reason: string) {
	return request<SwapRequest>(call, 'POST', `/swap-requests/${id(swapId)}/reject`, { reason });
}

/** RN-27: o membro retira o próprio pedido de troca. */
export function withdrawSwap(call: Call, swapId: string) {
	return request<SwapRequest>(call, 'POST', `/swap-requests/${id(swapId)}/withdraw`);
}

/** RN-33: as candidaturas do Usuário da sessão. */
export function listMyApplications(call: Call) {
	return request<MyApplication[]>(call, 'GET', '/me/applications');
}
