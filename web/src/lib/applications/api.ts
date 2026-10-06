// Chamadas do servidor do web à API de candidaturas, pelos tipos gerados do contrato
// (spec candidatura-lobby, design D-06 e D-07).
import { request, type Call } from '$lib/api/request';
import type { components } from '$lib/api/schema.gen';

export type Application = components['schemas']['Application'];
export type ApplicationStatus = components['schemas']['ApplicationStatus'];
export type MyApplication = components['schemas']['MyApplication'];
export type Participant = components['schemas']['LobbyParticipant'];
export type ViewerApplication = components['schemas']['ViewerApplication'];
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

/** RN-33: as candidaturas do Usuário da sessão. */
export function listMyApplications(call: Call) {
	return request<MyApplication[]>(call, 'GET', '/me/applications');
}
