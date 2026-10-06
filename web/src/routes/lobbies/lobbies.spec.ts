import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import LobbyForm from '$lib/lobbies/components/LobbyForm.svelte';
import { TEMPLE } from '$lib/lobbies/fixtures';
import Detalhe from './[id]/+page.svelte';
import Novo from './novo/+page.svelte';

// Formulário de lobby e página de criação renderizados no servidor (spec lobbies, T-07).

const INSTANCES = [
	{ id: 'torre-da-constelacao', name: 'Torre da Constelação', level: 240, reset: 'three_days' },
	{ id: 'templo-do-demonio-rei', name: 'Templo do Demônio Rei', level: 160, reset: 'daily' },
	{ id: 'sonho-sombrio', name: 'Sonho Sombrio', level: 120, reset: 'daily' },
	{ id: 'vila-dos-porings', name: 'Vila dos Porings', level: 30, reset: 'daily' }
] as const;
const LIRIEN = {
	id: 'c1',
	nick: 'Lirien',
	classId: 'arcebispo',
	level: 178,
	role: 'support',
	portrait: 'retrato-2',
	link: null,
	isMain: true,
	createdAt: ''
} as const;
const DAYS = [
	{ date: '2026-10-06', label: 'Hoje, ter, 6 out' },
	{ date: '2026-10-07', label: 'qua, 7 out' }
];
const VALUES = {
	instanceId: 'templo-do-demonio-rei',
	date: '2026-10-07',
	time: '20:00',
	tank: '1',
	support: '2',
	dps: '3',
	minLevel: '160',
	characterId: 'c1',
	note: ''
};

const renderForm = (props: Record<string, unknown> = {}) =>
	render(LobbyForm, {
		props: {
			mode: 'create',
			instances: INSTANCES,
			characters: [LIRIEN],
			classes: [{ id: 'arcebispo', name: 'Arcebispo', plural: '', tier: 'terceira', family: '' }],
			days: DAYS,
			values: VALUES,
			cancelHref: '/',
			...props
		} as never
	}).body;

describe('formulário de lobby', () => {
	it('CA-06.2 / RN-02: instâncias em dois grupos, 130+ primeiro', () => {
		const html = renderForm();
		const groups = [...html.matchAll(/<optgroup label="([^"]+)"/g)].map((m) => m[1]);
		expect(groups).toEqual(['Nível 130 ou mais', 'Nível menor']);
		const high = html.slice(html.indexOf('Nível 130 ou mais'), html.indexOf('Nível menor'));
		expect(high).toContain('Torre da Constelação');
		expect(high).not.toContain('Sonho Sombrio');
		expect(html.indexOf('Sonho Sombrio')).toBeLessThan(html.indexOf('Vila dos Porings'));
	});

	it('CA-01.1: os padrões aparecem preenchidos e a prévia mostra o card', () => {
		const html = renderForm();
		expect(html).toMatch(/<option value="templo-do-demonio-rei" selected/);
		expect(html).toMatch(/name="tank"[^>]*value="1"|value="1"[^>]*name="tank"/);
		expect(html).toMatch(/value="c1"[^>]*checked|checked[^>]*value="c1"/);
		expect(html).toContain('Como vai aparecer na Home · qua, 7 out');
		expect(html).toContain('data-testid="lobby-card"');
		expect(html).toContain('Arcebispo');
	});

	it('RNF-01: os botões das vagas têm rótulo, e o erro fica ligado ao campo', () => {
		const html = renderForm({
			errors: { startsAt: 'Esse personagem já está num grupo nesse horário' }
		});
		expect(html).toContain('aria-label="Mais uma vaga de Tank"');
		expect(html).toContain('aria-label="Menos uma vaga de Dano"');
		expect(html).toMatch(/aria-describedby="([^"]+-startsAt-err)"/);
		expect(html).toContain('Esse personagem já está num grupo nesse horário');
	});

	it('RN-17: na edição, instância e personagem ficam fixos', () => {
		const html = renderForm({
			mode: 'update',
			instances: [],
			fixedInstance: { name: 'Templo do Demônio Rei', level: 160 }
		});
		expect(html).not.toContain('name="instanceId"');
		expect(html).toContain('Templo do Demônio Rei');
		expect(html).toMatch(/name="characterId"[^>]*disabled|disabled[^>]*name="characterId"/);
		expect(html).toContain('Salvar alterações');
	});
});

describe('/lobbies/novo renderizada no servidor', () => {
	it('CA-01.10 / RN-12: sem personagens, pede para cadastrar um, com link para /perfil', () => {
		const data = {
			user: { id: 'u', username: 'ana', globalName: null },
			characters: [],
			instances: INSTANCES,
			classes: [],
			days: DAYS,
			values: VALUES,
			loadError: null,
			loginHref: '/'
		};
		const html = render(Novo, { props: { data, form: null, params: {} } as never }).body;
		expect(html).toContain('Cadastre um personagem para criar lobbies');
		expect(html).toMatch(/<a href="\/perfil"[^>]*>[\s\S]*?Ir para o perfil/);
		expect(html).not.toContain('name="instanceId"');
	});
});

describe('/lobbies/[id] renderizada no servidor', () => {
	const renderDetail = (lobby: Record<string, unknown>, isOwner: boolean) =>
		render(Detalhe, {
			props: {
				data: {
					user: isOwner ? { id: 'u', username: 'ana', globalName: null } : null,
					lobby: { ...TEMPLE, ...lobby },
					ownerClass: 'Arcebispo',
					isOwner,
					now: '2026-10-06T19:40:00.000Z',
					loginHref: '/'
				},
				form: null,
				params: { id: TEMPLE.id }
			} as never
		}).body;

	it('CA-03.1 / RN-15: instância, dia e hora de Brasília, nível, vagas com o dono e observação', () => {
		const html = renderDetail({ note: 'Chamar no Discord 15 min antes' }, false);
		expect(html).toContain('Templo do Demônio Rei');
		expect(html).toMatch(/qua, 7 out · 20:00/);
		expect(html).toMatch(/Nível mínimo <b[^>]*>160<\/b>/);
		expect(html).toContain('Retorno diário');
		expect(html).toContain('1 de 6');
		expect(html).toMatch(
			/data-testid="owner-slot"[\s\S]*?Lirien[\s\S]*?Arcebispo · Nv 178 · anfitrião/
		);
		expect(html.match(/Vaga aberta/g)).toHaveLength(5);
		expect(html).toContain('Chamar no Discord 15 min antes');
		expect(html).toContain('Grimbold');
	});

	it('CA-03.3 / RN-16: o dono vê Editar e Cancelar; os outros, Candidatar em breve', () => {
		const owner = renderDetail({}, true);
		expect(owner).toContain(`href="/lobbies/${TEMPLE.id}/editar"`);
		expect(owner).toContain('Editar');
		expect(owner).toContain('Cancelar lobby');
		expect(owner).not.toContain('Candidatar');
		const other = renderDetail({}, false);
		expect(other).not.toContain('Cancelar lobby');
		expect(other).toMatch(/aria-disabled="true"[^>]*>[\s\S]*?Candidatar/);
		expect(other).toContain('Disponível em breve');
	});

	it('CA-05.1 / RN-19: cancelado mostra o selo e o motivo, sem ações', () => {
		const html = renderDetail(
			{ status: 'cancelled', cancelReason: 'Metade do grupo não pode' },
			true
		);
		expect(html).toMatch(/data-testid="lobby-status"[^>]*>\s*Cancelado/);
		expect(html).toContain('Motivo do cancelamento');
		expect(html).toContain('Metade do grupo não pode');
		expect(html).not.toContain('Cancelar lobby');
		expect(html).not.toContain('/editar');
	});

	it('D-01: personagem do dono excluído depois do início aparece como "Personagem excluído"', () => {
		const html = renderDetail(
			{
				status: 'started',
				owner: { ...TEMPLE.owner, characterId: null, nick: null, classId: null, level: null }
			},
			false
		);
		expect(html).toContain('Personagem excluído');
		expect(html).toMatch(/data-testid="lobby-status"[^>]*>\s*Iniciado/);
	});
});
