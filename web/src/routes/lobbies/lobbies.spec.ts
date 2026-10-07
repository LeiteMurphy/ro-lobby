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
					user: isOwner ? { id: TEMPLE.owner.userId, username: 'ana', globalName: null } : null,
					lobby: { ...TEMPLE, ...lobby },
					ownerClass: 'Arcebispo',
					classNames: { arcebispo: 'Arcebispo' },
					characters: [],
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
			/data-testid="owner-slot"[\s\S]*?\/portraits\/retrato-2\.svg[\s\S]*?Lirien[\s\S]*?Arcebispo · Nv 178[\s\S]*?Anfitrião/
		);
		expect(html.match(/Vaga aberta/g)).toHaveLength(5);
		expect(html).toContain('Chamar no Discord 15 min antes');
		expect(html).toContain('Grimbold');
	});

	it('CA-03.3 / RN-16: o dono vê Editar e Cancelar; sem sessão, entrar para se candidatar', () => {
		const owner = renderDetail({}, true);
		expect(owner).toContain(`href="/lobbies/${TEMPLE.id}/editar"`);
		expect(owner).toContain('Editar');
		expect(owner).toContain('Cancelar lobby');
		expect(owner).not.toContain('Candidatar');
		const other = renderDetail({}, false);
		expect(other).not.toContain('Cancelar lobby');
		expect(other).toContain('Entrar para se candidatar');
		expect(other).not.toContain('Disponível em breve');
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

describe('/lobbies/[id] com candidaturas (candidatura-lobby, T-05)', () => {
	const MEMBER = {
		applicationId: 'a1',
		userId: 'u-bia',
		discordName: null,
		characterId: 'c-brasa',
		nick: 'Brasa',
		classId: 'guardiao-real',
		level: 200,
		portrait: 'retrato-3',
		link: null,
		role: 'tank',
		message: null,
		createdAt: '2026-10-06T20:10:00Z'
	};
	const CANDIDATE = {
		...MEMBER,
		applicationId: 'a2',
		nick: 'Fogo',
		role: 'dps',
		discordName: 'Caio',
		message: 'tenho buff'
	};
	const BASE = {
		occupied: { tank: 1, support: 1, dps: 0 },
		members: [MEMBER],
		pendingCount: 2
	};
	const renderAs = (userId: string | null, lobby: Record<string, unknown>) =>
		render(Detalhe, {
			props: {
				data: {
					user: userId ? { id: userId, username: 'x', globalName: null } : null,
					lobby: { ...TEMPLE, ...BASE, ...lobby },
					ownerClass: 'Arcebispo',
					classNames: { 'guardiao-real': 'Guardião Real' },
					characters: [],
					isOwner: userId === TEMPLE.owner.userId,
					now: '2026-10-06T19:40:00.000Z',
					loginHref: '/'
				},
				form: null,
				params: { id: TEMPLE.id }
			} as never
		}).body;
	const mine = (status: string, reason: string | null = null) => ({
		myApplication: {
			id: 'm1',
			characterId: 'c',
			role: 'dps',
			message: null,
			status,
			reason,
			createdAt: '',
			decidedAt: null
		}
	});

	it('CA-02.1 / CA-10.1: o membro aceito ocupa a vaga e é um botão do painel', () => {
		const html = renderAs(null, {});
		expect(html).toMatch(
			/<button[^>]*data-testid="member-slot"[\s\S]*?Brasa[\s\S]*?Guardião Real · Nv 200/
		);
		expect(html).toMatch(/<button[^>]*aria-pressed="true"[^>]*data-testid="owner-slot"/);
		expect(html.match(/Vaga aberta/g)).toHaveLength(4);
		expect(html).toContain('data-testid="player-panel"');
	});

	it('CA-03.7 / CA-10.3 / RN-28: o terceiro vê só a quantidade; o dono vê os candidatos e o selo', () => {
		const other = renderAs('u-duda', {});
		expect(other).toMatch(/data-testid="pending-count"[^>]*>\s*2\s*candidaturas pendentes/);
		expect(other).not.toContain('data-testid="pending-list"');
		expect(other).not.toContain('data-testid="pending-badge"');
		const owner = renderAs(TEMPLE.owner.userId, { pending: [CANDIDATE], pendingCount: 1 });
		expect(owner).toMatch(/data-testid="pending-badge"[^>]*>\s*1 pendente/);
		expect(owner).toMatch(/data-testid="pending-list"[\s\S]*?data-testid="candidate"[\s\S]*?Fogo/);
		expect(owner).toContain('só você vê');
	});

	it('CA-01.1: com sessão e sem candidatura, o botão Candidatar abre o diálogo', () => {
		const html = renderAs('u-duda', {});
		expect(html).toMatch(/<button[^>]*>[\s\S]*?Candidatar<\/span>|>\s*Candidatar\s*</);
		expect(html).not.toContain('Entrar para se candidatar');
		expect(html).not.toContain('data-testid="my-application"');
	});

	it('CA-03.2 / RN-13: pendente mostra o aviso e o botão de retirar, sem Candidatar', () => {
		const html = renderAs('u-duda', mine('pending'));
		expect(html).toMatch(/data-testid="my-application"[\s\S]*?pendente/);
		expect(html).toMatch(/action="\?\/withdraw"[\s\S]*?value="m1"[\s\S]*?Retirar candidatura/);
		expect(html).not.toMatch(/>\s*Candidatar\s*</);
	});

	it('CA-03.8 / RN-29: recusada mostra a justificativa a quem se candidatou, sem Candidatar', () => {
		const html = renderAs('u-duda', mine('rejected', 'já temos dano'));
		expect(html).toContain('Sua candidatura foi recusada.');
		expect(html).toContain('Justificativa: já temos dano');
		expect(html).not.toMatch(/>\s*Candidatar\s*</);
	});

	it('CA-02.1: aceita mostra que está no grupo', () => {
		const html = renderAs('u-duda', mine('accepted'));
		expect(html).toContain('Você está no grupo.');
		expect(html).not.toContain('Retirar candidatura');
	});
});
