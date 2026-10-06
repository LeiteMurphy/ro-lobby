// Apresentação da Home: capa e ícone das instâncias, ícone e função por classe. A capa e o
// ícone são arte original do design system do RO Lobby. Os nomes de classe e de instância
// seguem o bROWiki (RN-22 da home-local, RN-01 da lobbies).
import { findClass } from '$lib/catalog/classes';
import type { IconName } from '$lib/ui/Icon.svelte';
import type { Role } from './types';

export interface InstanceArt {
	icon: string;
	cover: string;
}

const FALLBACK_INSTANCE: InstanceArt = {
	icon: '/brand/c1-symbol-dark.svg',
	cover: 'var(--gradient-hero)'
};

/**
 * spec lobbies, RN-03: ainda não há arte por instância; todas usam a capa e o ícone padrão
 * do RO Lobby, sem nada da Gravity.
 */
export function instanceArt(instance: string): InstanceArt {
	void instance;
	return FALLBACK_INSTANCE;
}

/** Ícone e função sugerida da classe, a partir do catálogo do bRO (RN-22). */
export function classArt(hostClass: string): { icon: IconName; role: Role } {
	const found = findClass(hostClass);
	return found ? { icon: found.icon, role: found.role } : { icon: 'user', role: 'dps' };
}

export const ROLE_ICONS: Record<Role, IconName> = {
	tank: 'shield',
	support: 'heart-pulse',
	dps: 'swords'
};
