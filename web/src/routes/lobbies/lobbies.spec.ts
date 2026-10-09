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
		expect(html).toContain('Você está no grupo');
		expect(html).not.toContain('Retirar candidatura');
	});
});

describe('/lobbies/[id] do lado do membro (candidatura-lobby, T-14)', () => {
	// Bia é membro como Dano (Fogo); tem também Brisa (Tank).
	const FOGO = {
		applicationId: 'm1',
		userId: 'u-bia',
		discordName: 'Bia',
		characterId: 'c-fogo',
		nick: 'Fogo',
		classId: 'arquimago',
		level: 200,
		portrait: 'retrato-1',
		link: null,
		role: 'dps',
		message: null,
		createdAt: '2026-10-06T20:10:00Z'
	};
	const BRISA = {
		id: 'c-brisa',
		nick: 'Brisa',
		classId: 'guardiao-real',
		level: 246,
		role: 'tank',
		portrait: 'retrato-3',
		link: null,
		isMain: false,
		createdAt: ''
	};
	const mine = (status: string, over: Record<string, unknown> = {}) => ({
		id: 'm1',
		characterId: 'c-fogo',
		role: 'dps',
		message: null,
		status,
		reason: null,
		blocked: false,
		createdAt: '',
		decidedAt: null,
		swapRequest: null,
		...over
	});
	const renderAs = (myApplication: Record<string, unknown>, lobby: Record<string, unknown> = {}) =>
		render(Detalhe, {
			props: {
				data: {
					user: { id: 'u-bia', username: 'bia', globalName: 'Bia' },
					lobby: {
						...TEMPLE,
						occupied: { tank: 0, support: 1, dps: 1 },
						members: [FOGO],
						myApplication,
						...lobby
					},
					ownerClass: 'Arcebispo',
					classNames: {},
					characters: [BRISA],
					isOwner: false,
					now: '2026-10-06T19:40:00.000Z',
					loginHref: '/'
				},
				form: null,
				params: { id: TEMPLE.id }
			} as never
		}).body.replace(/<!--[\s\S]*?-->/g, ''); // sem os marcadores do Svelte, para ler o texto
	const candidate = />\s*Candidatar\s*</;
	const empty = { members: [], occupied: { tank: 0, support: 1, dps: 0 } };

	it('CA-05.5 / CA-08.14 / RN-38: o membro está no grupo com o personagem e tem Pedir troca e Sair do grupo', () => {
		const html = renderAs(mine('accepted'));
		expect(html).toMatch(/Você está no grupo<\/b>\s*com Fogo \(Dano\)/);
		expect(html).toMatch(/data-testid="my-application"[\s\S]*?Pedir troca[\s\S]*?Sair do grupo/);
		expect(html).not.toMatch(candidate);
	});

	it('RN-14: com o lobby iniciado, o membro não tem mais Pedir troca nem Sair do grupo', () => {
		const html = renderAs(mine('accepted'), { status: 'started' });
		expect(html).toContain('Você está no grupo');
		expect(html).not.toContain('Pedir troca');
		expect(html).not.toContain('Sair do grupo');
	});

	it('CA-08.14 / RN-21 / RN-27: com pedido pendente, o aviso mostra o personagem pedido e Retirar pedido', () => {
		const swap = {
			id: 's1',
			applicationId: 'm1',
			fromCharacterId: 'c-fogo',
			toCharacterId: 'c-brisa',
			toRole: 'tank',
			reason: 'ninguém apareceu de tank',
			status: 'pending',
			decisionReason: null,
			createdAt: '',
			decidedAt: null
		};
		const html = renderAs(mine('accepted', { swapRequest: swap }));
		expect(html).toMatch(/Pedido de troca pendente<\/b>\s*para Brisa \(Tank, Nv 246\)/);
		expect(html).toContain('Você continua com Fogo até o anfitrião decidir.');
		expect(html).toMatch(
			/action="\?\/withdrawSwap"[\s\S]*?name="swapId" value="s1"[\s\S]*?Retirar pedido/
		);
		expect(html).not.toContain('Pedir troca');

		// Pedido já decidido não muda o aviso.
		const decided = renderAs(mine('accepted', { swapRequest: { ...swap, status: 'rejected' } }));
		expect(decided).toContain('Pedir troca');
		expect(decided).not.toContain('Retirar pedido');
	});

	it('CA-08.15 / RN-39: o membro vê a justificativa da recusa do pedido e pode pedir outra troca', () => {
		const swap = {
			id: 's1',
			applicationId: 'm1',
			fromCharacterId: 'c-fogo',
			toCharacterId: 'c-brisa',
			toRole: 'tank',
			reason: 'ninguém apareceu de tank',
			status: 'rejected',
			decisionReason: 'já achamos um tank',
			createdAt: '',
			decidedAt: ''
		};
		const html = renderAs(mine('accepted', { swapRequest: swap }));
		expect(html).toMatch(
			/data-testid="swap-rejected"[^>]*>\s*Seu pedido de troca foi recusado\. Justificativa: já achamos um tank/
		);
		expect(html).toMatch(/Você está no grupo<\/b>\s*com Fogo \(Dano\)/);
		expect(html).toContain('Pedir troca');
		// Pedido retirado ou lobby iniciado não mostram a recusa.
		expect(
			renderAs(mine('accepted', { swapRequest: { ...swap, status: 'withdrawn' } }))
		).not.toContain('swap-rejected');
		expect(renderAs(mine('accepted', { swapRequest: swap }), { status: 'started' })).not.toContain(
			'swap-rejected'
		);
	});

	it('CA-06.8 / RN-15 / RN-38: removido com bloqueio vê a justificativa e não tem Candidatar', () => {
		const html = renderAs(
			mine('removed', { reason: 'mudamos o horário da run', blocked: true }),
			empty
		);
		expect(html).toContain('Você foi removido deste lobby.');
		expect(html).toContain('Justificativa: mudamos o horário da run');
		expect(html).toContain('Você não pode se candidatar a este lobby.');
		expect(html).not.toMatch(candidate);
	});

	it('CA-06.5 / CA-05.4 / RN-37: removido sem bloqueio e quem saiu podem se candidatar de novo', () => {
		const removed = renderAs(mine('removed', { reason: 'mudamos o horário da run' }), empty);
		expect(removed).toContain('Você foi removido deste lobby.');
		expect(removed).not.toContain('Você não pode se candidatar a este lobby.');
		expect(removed).toMatch(candidate);
		const left = renderAs(mine('left'), empty);
		expect(left).toContain('Você saiu do grupo.');
		expect(left).toMatch(candidate);
	});
});

describe('/lobbies/[id] do lado do dono (candidatura-lobby, T-15)', () => {
	const FAISCA = {
		applicationId: 'a1',
		userId: 'u-bia',
		discordName: 'faisca.ro',
		characterId: 'c-faisca',
		nick: 'Faísca',
		classId: 'elementalista',
		level: 255,
		portrait: 'retrato-1',
		link: null,
		role: 'dps',
		message: null,
		createdAt: '2026-10-06T17:30:00Z'
	};
	const SWAP = {
		id: 's1',
		applicationId: 'a1',
		userId: 'u-bia',
		discordName: 'faisca.ro',
		from: { ...FAISCA, characterId: 'c-faisca' },
		to: {
			characterId: 'c-brisa',
			nick: 'Brisa',
			classId: 'guardiao-real',
			level: 246,
			portrait: 'retrato-3',
			role: 'tank'
		},
		reason: 'posso cobrir de tank',
		createdAt: '2026-10-06T18:00:00Z'
	};
	const renderOwner = (lobby: Record<string, unknown>) =>
		render(Detalhe, {
			props: {
				data: {
					user: { id: TEMPLE.owner.userId, username: 'ana', globalName: 'Ana' },
					lobby: {
						...TEMPLE,
						occupied: { tank: 0, support: 1, dps: 1 },
						members: [FAISCA],
						pending: [],
						pendingCount: 0,
						swapRequests: [],
						...lobby
					},
					ownerClass: 'Arcebispo',
					classNames: {},
					characters: [],
					isOwner: true,
					now: '2026-10-06T19:40:00.000Z',
					loginHref: '/'
				},
				form: null,
				params: { id: TEMPLE.id }
			} as never
		}).body.replace(/<!--[\s\S]*?-->/g, '');

	it('CA-08.14 / RN-38 / D-12: o dono vê os pedidos de troca num bloco próprio, com o selo', () => {
		const html = renderOwner({ swapRequests: [SWAP], pendingCount: 1 });
		expect(html).toMatch(/data-testid="pending-badge"[^>]*>\s*1 pendente · 1 troca/);
		expect(html).toMatch(
			/data-testid="swap-list"[\s\S]*?Pedidos de troca[\s\S]*?só você vê[\s\S]*?data-testid="swap-request"[\s\S]*?Faísca → Brisa[\s\S]*?Dano → Tank · Nv 246/
		);
	});

	it('RN-38: sem pedidos, o bloco diz que não há nenhum', () => {
		expect(renderOwner({})).toContain('Nenhum pedido de troca.');
	});

	it('CA-07.1 / RN-19 / RN-38: o painel começa no anfitrião, com Trocar personagem para o dono', () => {
		const html = renderOwner({});
		expect(html).toMatch(/data-testid="player-panel"[\s\S]*?Trocar personagem/);
	});

	it('RN-16 / RN-38: com o lobby iniciado, o dono não tem os pedidos nem as ações', () => {
		const html = renderOwner({ status: 'started', swapRequests: [] });
		expect(html).not.toContain('data-testid="swap-list"');
		expect(html).not.toContain('Trocar personagem');
	});
});

describe('/lobbies/[id] compartilhável (spec compartilhar-lobby)', () => {
	// O lobby do CA-01.3 e do CA-02.1: Glast Heim no sábado 10/10 às 20:00, nível mínimo 160,
	// 1 vaga de Tank e 2 de Dano abertas, Suporte cheio e anfitrião "Brasa".
	const GLAST = {
		instance: { ...TEMPLE.instance, name: 'Glast Heim' },
		startsAt: '2026-10-10T23:00:00Z',
		slots: { tank: 1, support: 2, dps: 3 },
		occupied: { tank: 0, support: 2, dps: 1 },
		owner: { ...TEMPLE.owner, nick: 'Brasa' }
	};
	const renderShared = (lobby: Record<string, unknown>, userId: string | null) => {
		const { head, body } = render(Detalhe, {
			props: {
				data: {
					user: userId ? { id: userId, username: 'x', globalName: null } : null,
					lobby: { ...TEMPLE, ...GLAST, ...lobby },
					ownerClass: 'Arcebispo',
					classNames: {},
					characters: [],
					isOwner: userId === TEMPLE.owner.userId,
					now: '2026-10-06T19:40:00.000Z',
					loginHref: '/',
					origin: 'https://rolobby.com.br'
				},
				form: null,
				params: { id: TEMPLE.id }
			} as never
		});
		return { head, body };
	};
	const meta = (head: string, key: string) =>
		head.match(new RegExp(`<meta (?:property|name)="${key}" content="([^"]*)"`))?.[1];

	it('CA-01.1: visitante, candidato e dono veem "Compartilhar" no lobby aberto', () => {
		for (const userId of [null, 'outra-pessoa', TEMPLE.owner.userId]) {
			expect(renderShared({}, userId).body).toContain('Compartilhar');
		}
	});

	it('CA-01.2: lobby iniciado ou cancelado não mostra "Compartilhar"', () => {
		for (const status of ['started', 'cancelled']) {
			const { body } = renderShared({ status, cancelReason: 'Faltou tank' }, null);
			expect(body).not.toContain('Compartilhar');
		}
	});

	it('CA-02.1: o HTML sem sessão traz as meta tags Open Graph do lobby aberto', () => {
		const { head } = renderShared({}, null);
		expect(meta(head, 'og:title')).toBe('Glast Heim · sábado, 10/10 às 20:00');
		expect(meta(head, 'og:description')).toBe(
			'Vagas: 1 Tank, 2 Dano · Nível mínimo 160 · Anfitrião Brasa'
		);
		expect(meta(head, 'og:url')).toBe(`https://rolobby.com.br/lobbies/${TEMPLE.id}`);
		expect(meta(head, 'og:site_name')).toBe('RO Lobby');
		expect(meta(head, 'og:type')).toBe('website');
		expect(meta(head, 'theme-color')).toBe('#f6bb45');
		expect(head).not.toContain('og:image');
	});

	it('CA-02.2: lobby cancelado avisa o cancelamento no preview', () => {
		const { head } = renderShared({ status: 'cancelled', cancelReason: 'Faltou tank' }, null);
		expect(meta(head, 'og:description')).toBe('Esse grupo foi cancelado.');
	});
});
