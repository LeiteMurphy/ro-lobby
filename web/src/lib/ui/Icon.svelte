<script lang="ts" module>
	import Check from '@lucide/svelte/icons/check';
	import Crosshair from '@lucide/svelte/icons/crosshair';
	import Drama from '@lucide/svelte/icons/drama';
	import FlaskConical from '@lucide/svelte/icons/flask-conical';
	import Hammer from '@lucide/svelte/icons/hammer';
	import HandFist from '@lucide/svelte/icons/hand-fist';
	import Leaf from '@lucide/svelte/icons/leaf';
	import Moon from '@lucide/svelte/icons/moon';
	import PawPrint from '@lucide/svelte/icons/paw-print';
	import Sparkles from '@lucide/svelte/icons/sparkles';
	import Star from '@lucide/svelte/icons/star';
	import Sword from '@lucide/svelte/icons/sword';
	import Target from '@lucide/svelte/icons/target';
	import Zap from '@lucide/svelte/icons/zap';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import ChevronsUp from '@lucide/svelte/icons/chevrons-up';
	import Cross from '@lucide/svelte/icons/cross';
	import HeartPulse from '@lucide/svelte/icons/heart-pulse';
	import Lock from '@lucide/svelte/icons/lock';
	import LogIn from '@lucide/svelte/icons/log-in';
	import Music from '@lucide/svelte/icons/music';
	import Plus from '@lucide/svelte/icons/plus';
	import Shield from '@lucide/svelte/icons/shield';
	import ShieldHalf from '@lucide/svelte/icons/shield-half';
	import ShieldPlus from '@lucide/svelte/icons/shield-plus';
	import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
	import Swords from '@lucide/svelte/icons/swords';
	import User from '@lucide/svelte/icons/user';
	import UserPlus from '@lucide/svelte/icons/user-plus';
	import Users from '@lucide/svelte/icons/users';
	import VenetianMask from '@lucide/svelte/icons/venetian-mask';
	import WandSparkles from '@lucide/svelte/icons/wand-sparkles';
	import X from '@lucide/svelte/icons/x';
	import CircleX from '@lucide/svelte/icons/circle-x';
	import Ellipsis from '@lucide/svelte/icons/ellipsis';
	import ExternalLink from '@lucide/svelte/icons/external-link';
	import Pencil from '@lucide/svelte/icons/pencil';
	import Trash2 from '@lucide/svelte/icons/trash-2';
	import Share2 from '@lucide/svelte/icons/share-2';

	// Só os ícones Lucide que a interface usa, empacotados no build (RNF-02). O design
	// system carregava cada ícone de uma CDN.
	export const ICONS = {
		crosshair: Crosshair,
		drama: Drama,
		'flask-conical': FlaskConical,
		hammer: Hammer,
		'hand-fist': HandFist,
		leaf: Leaf,
		moon: Moon,
		'paw-print': PawPrint,
		sparkles: Sparkles,
		star: Star,
		sword: Sword,
		target: Target,
		zap: Zap,
		check: Check,
		'chevron-down': ChevronDown,
		'chevron-left': ChevronLeft,
		'chevron-right': ChevronRight,
		'chevrons-up': ChevronsUp,
		cross: Cross,
		'heart-pulse': HeartPulse,
		lock: Lock,
		'log-in': LogIn,
		music: Music,
		plus: Plus,
		shield: Shield,
		'shield-half': ShieldHalf,
		'shield-plus': ShieldPlus,
		'sliders-horizontal': SlidersHorizontal,
		swords: Swords,
		user: User,
		'user-plus': UserPlus,
		users: Users,
		'venetian-mask': VenetianMask,
		'wand-sparkles': WandSparkles,
		x: X,
		'circle-x': CircleX,
		ellipsis: Ellipsis,
		'external-link': ExternalLink,
		pencil: Pencil,
		'trash-2': Trash2,
		'share-2': Share2
	} as const;

	export type IconName = keyof typeof ICONS;
</script>

<script lang="ts">
	interface Props {
		name: IconName;
		size?: number;
		color?: string;
		/** Texto para leitores de tela. Sem ele, o ícone é decorativo. */
		title?: string;
		class?: string;
	}

	let { name, size = 16, color = 'currentColor', title, class: className }: Props = $props();
	const Component = $derived(ICONS[name]);
</script>

<Component
	{size}
	{color}
	strokeWidth={2}
	class={className}
	aria-hidden={title ? undefined : 'true'}
	aria-label={title}
	role={title ? 'img' : undefined}
	style="flex:none"
/>
