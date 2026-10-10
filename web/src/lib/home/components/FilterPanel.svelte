<script lang="ts">
	import Check from '$lib/ui/Check.svelte';
	import { NO_INSTANCE } from '$lib/lobbies/form';
	import Select from '$lib/ui/Select.svelte';
	import {
		LEVEL_OPTIONS,
		TIME_RANGES,
		type Filters,
		type TimeRange,
		type TimeRangeKey
	} from '../filters';
	import { ROLE_ICONS } from '../catalog';
	import { ROLE_LABELS, ROLES, type Role } from '../types';

	interface Props {
		filters: Filters;
		instances: readonly string[];
		roleCounts: Record<Role, number>;
		timeCounts: Record<TimeRangeKey, number>;
		/** RN-12: as faixas visíveis no dia; sem elas, só as fixas. */
		ranges?: readonly TimeRange[];
		/** Nome do grupo de radios, único por painel (barra lateral e gaveta). */
		name: string;
		size?: 'sm' | 'md';
		onchange: (filters: Filters) => void;
	}

	let {
		filters,
		instances,
		roleCounts,
		timeCounts,
		ranges = TIME_RANGES,
		name,
		size = 'sm',
		onchange
	}: Props = $props();

	// RN-05 da lobby-sem-instancia: a opção "Sem instância definida".
	const instanceOptions = $derived([
		{ value: NO_INSTANCE, label: 'Sem instância definida' },
		...instances.map((i) => ({ value: i, label: i }))
	]);

	function toggleRole(role: Role, on: boolean) {
		const roles = on ? [...filters.roles, role] : filters.roles.filter((r) => r !== role);
		onchange({ ...filters, roles: ROLES.filter((r) => roles.includes(r)) });
	}
</script>

<div class="panel">
	<section>
		<h4>Instância</h4>
		<Select
			label="Instância"
			{size}
			placeholder="Todas"
			options={instanceOptions}
			value={filters.instance}
			onchange={(instance) => onchange({ ...filters, instance })}
		/>
	</section>
	<section>
		<h4>Vaga para</h4>
		{#each ROLES as role (role)}
			<Check
				type="checkbox"
				label={ROLE_LABELS[role]}
				icon={ROLE_ICONS[role]}
				iconColor="var(--{role}-400)"
				meta={roleCounts[role]}
				checked={filters.roles.includes(role)}
				onchange={(on) => toggleRole(role, on)}
			/>
		{/each}
	</section>
	<section>
		<h4>Nível mínimo</h4>
		<Select
			label="Nível mínimo"
			{size}
			options={LEVEL_OPTIONS}
			value={filters.maxMinLevel === null ? '' : String(filters.maxMinLevel)}
			onchange={(v) => onchange({ ...filters, maxMinLevel: v === '' ? null : Number(v) })}
		/>
	</section>
	<section class="last">
		<h4>Faixa de horário</h4>
		{#each ranges as range (range.key)}
			<Check
				type="radio"
				{name}
				value={range.key}
				label={range.label}
				meta={timeCounts[range.key] ?? 0}
				checked={filters.timeRange === range.key}
				onchange={(on) => on && onchange({ ...filters, timeRange: range.key })}
			/>
		{/each}
	</section>
</div>

<style>
	.panel {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	section {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding-bottom: 18px;
		border-bottom: 1px solid var(--border-subtle);
	}
	section.last {
		padding-bottom: 0;
		border-bottom: 0;
	}
	h4 {
		margin: 0;
		font: var(--type-overline);
		letter-spacing: var(--tracking-caps);
		text-transform: uppercase;
		color: var(--fg-3);
	}
</style>
