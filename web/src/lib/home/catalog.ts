// Apresentação da Home: capa e ícone por instância, ícone e função por classe. As capas e
// os ícones são arte original do design system do RO Lobby, e os nomes de instância são
// inventados. Os nomes de classe seguem o bRO, conforme o bROWiki (RN-22).
import type { IconName } from '$lib/ui/Icon.svelte';
import type { Role } from './types';

export interface InstanceArt {
	icon: string;
	cover: string;
}

const INSTANCES: Record<string, InstanceArt> = {
	'Caverna de gelo': {
		icon: '/brand/inst-gelo.svg',
		cover:
			'radial-gradient(90% 120% at 85% 30%,rgba(143,182,255,.55) 0%,rgba(143,182,255,0) 60%),linear-gradient(135deg,#1d4a78 0%,#16304f 55%,#0e1a2b 100%)'
	},
	'Torre sem fim': {
		icon: '/brand/inst-torre.svg',
		cover:
			'radial-gradient(90% 120% at 85% 30%,rgba(166,150,255,.5) 0%,rgba(166,150,255,0) 60%),linear-gradient(135deg,#3b3478 0%,#231f4a 55%,#12132a 100%)'
	},
	'Templo submerso': {
		icon: '/brand/inst-templo.svg',
		cover:
			'radial-gradient(90% 120% at 85% 30%,rgba(79,179,161,.5) 0%,rgba(79,179,161,0) 60%),linear-gradient(135deg,#145a58 0%,#113a3e 55%,#0b1d22 100%)'
	},
	'Fortaleza do deserto': {
		icon: '/brand/inst-deserto.svg',
		cover:
			'radial-gradient(90% 120% at 85% 30%,rgba(246,187,69,.5) 0%,rgba(246,187,69,0) 60%),linear-gradient(135deg,#7a4a16 0%,#4a2e12 55%,#1f150c 100%)'
	},
	'Ruínas ao norte': {
		icon: '/brand/inst-ruinas.svg',
		cover:
			'radial-gradient(90% 120% at 85% 30%,rgba(170,190,120,.45) 0%,rgba(170,190,120,0) 60%),linear-gradient(135deg,#3f4d2c 0%,#2a3322 55%,#141a12 100%)'
	}
};

const FALLBACK_INSTANCE: InstanceArt = {
	icon: '/brand/c1-symbol-dark.svg',
	cover: 'var(--gradient-hero)'
};

export function instanceArt(instance: string): InstanceArt {
	return INSTANCES[instance] ?? FALLBACK_INSTANCE;
}

const CLASSES: Record<string, { icon: IconName; role: Role }> = {
	Arcebispo: { icon: 'cross', role: 'support' },
	Paladino: { icon: 'shield-half', role: 'tank' },
	Feiticeiro: { icon: 'wand-sparkles', role: 'dps' },
	Sicário: { icon: 'venetian-mask', role: 'dps' },
	'Guardião Real': { icon: 'shield-plus', role: 'tank' },
	Musa: { icon: 'music', role: 'support' }
};

export function classArt(hostClass: string): { icon: IconName; role: Role } {
	return CLASSES[hostClass] ?? { icon: 'user', role: 'dps' };
}

export const ROLE_ICONS: Record<Role, IconName> = {
	tank: 'shield',
	support: 'heart-pulse',
	dps: 'swords'
};
