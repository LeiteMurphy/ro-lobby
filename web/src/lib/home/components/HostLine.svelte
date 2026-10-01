<script lang="ts">
	import Icon from '$lib/ui/Icon.svelte';
	import { classArt } from '../catalog';

	interface Props {
		host: string;
		hostClass: string;
		minLevel: number;
		size?: 'md' | 'lg';
	}

	let { host, hostClass, minLevel, size = 'md' }: Props = $props();
	const art = $derived(classArt(hostClass));
	const lg = $derived(size === 'lg');
</script>

<div class="line" class:lg>
	<span class="who">
		<span
			class="tile"
			style="background:var(--gradient-{art.role});border-color:var(--{art.role}-line)"
		>
			<Icon name={art.icon} size={lg ? 14 : 13} color="var(--{art.role}-300)" />
		</span>
		<span class="host">{host}</span>
		<span class="cls">{hostClass}</span>
	</span>
	<span class="level">
		<Icon name="chevrons-up" size={lg ? 15 : 14} color="var(--fg-3)" />Nv {minLevel}+
	</span>
</div>

<style>
	.line {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
		font: 500 13px/1.35 var(--font-ui);
		color: var(--fg-2);
	}
	.line.lg {
		gap: 16px;
		font-size: 14px;
	}
	.who {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}
	.tile {
		flex: none;
		width: 24px;
		height: 24px;
		border-radius: var(--radius-sm);
		border: 1px solid;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.lg .tile {
		width: 26px;
		height: 26px;
	}
	.host {
		color: var(--fg-1);
		font-weight: 600;
	}
	.cls {
		color: var(--fg-3);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.level {
		display: flex;
		align-items: center;
		gap: 6px;
	}
</style>
