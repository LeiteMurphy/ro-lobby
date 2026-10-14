// Servidores do Ragnarok LATAM (spec servidor, RN-01), na mesma ordem e com os mesmos
// códigos de GET /servers. Fixo como as classes: um servidor novo entra por mudança no
// catálogo da API e daqui.
import type { components } from '$lib/api/schema.gen';

export type ServerId = components['schemas']['ServerId'];

export interface GameServer {
	id: ServerId;
	name: string;
}

export const SERVERS: readonly GameServer[] = [
	{ id: 'freya', name: 'Freya' },
	{ id: 'nidhogg', name: 'Nidhogg' }
];

/** Nome do servidor para a tela; um código fora do catálogo aparece como veio. */
export function serverName(id: string): string {
	return SERVERS.find((s) => s.id === id)?.name ?? id;
}
