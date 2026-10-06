import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import { DAY_COUNTS, getHomeLobbies } from '$lib/home/fixtures';
import { lobbiesForDay } from '$lib/home/lobbies';
import EmptyState from '$lib/home/components/EmptyState.svelte';
import Page from './+page.svelte';

// 2026-09-30 19:40 UTC = 16:40 em São Paulo, como no design.
const NOW = '2026-09-30T19:40:00.000Z';
const TODAY = '2026-09-30';

type User = { id: string; username: string; globalName: string | null };

function renderHome(user: User | null = null, loginError = false): string {
	const data = {
		now: NOW,
		today: TODAY,
		lobbies: getHomeLobbies(TODAY),
		user,
		loginHref: '/auth/discord/login?next=%2F',
		loginError
	};
	// O PageProps completo inclui `params`, que a Home não usa.
	const props = { data, params: {} } as unknown as Parameters<typeof Page>[1];
	return render(Page, { props }).body;
}

describe('Home renderizada no servidor', () => {
	const html = renderHome();
	const today = lobbiesForDay(getHomeLobbies(TODAY), TODAY);

	it('CA-02.6 / RN-20: o HTML do servidor já traz os cards de hoje', () => {
		const cards = html.match(/data-testid="lobby-card"/g) ?? [];
		expect(cards).toHaveLength(DAY_COUNTS[0]);
		for (const l of today) expect(html).toContain(`${l.instance} às ${l.time}`);
	});

	it('CA-02.1: título, subtítulo e dados de cada card', () => {
		expect(html).toContain('Grupos para hoje');
		expect(html).toContain(`qua, 30 set · ${DAY_COUNTS[0]} grupos · por horário`);
		expect(html).toContain('Lyrae');
		expect(html).toContain('Arcebispo');
		expect(html).toContain('Nv 160+');
	});

	it('CA-02.2: o lobby lotado mostra o selo "Lotado"', () => {
		expect(html).toMatch(/data-edge="full"[\s\S]*?Lotado/);
	});

	it('CA-02.4: tempo relativo no card das 18:00', () => {
		expect(html).toContain('em 1 h 20 min');
	});

	it('CA-02.5 / RN-18: ações sem backend continuam desabilitadas com "Disponível em breve"', () => {
		const buttons = [...html.matchAll(/<button(\s[^>]*)?>([\s\S]*?)<\/button>/g)].map(
			([, attrs, inner]) => ({
				disabled: (attrs ?? '').includes('aria-disabled="true"'),
				text: inner.replace(/<[^>]+>/g, '').trim()
			})
		);
		// Desde a spec lobbies, "Criar lobby" e "Ver grupo" funcionam; "Candidatar" segue em breve.
		for (const label of ['Candidatar']) {
			const matching = buttons.filter((b) => b.text === label);
			expect(matching.length, label).toBeGreaterThan(0);
			expect(
				matching.every((b) => b.disabled),
				label
			).toBe(true);
		}
		expect(html).toContain('Disponível em breve');
	});

	it('CA-02.3 / CA-02.4 / CA-02.5 / RN-23 / RN-24 (lobbies): "Criar lobby" leva à criação no dia escolhido, e "Ver grupo" ao detalhe', () => {
		const create = `/lobbies/novo?dia=${TODAY}`;
		expect(html.split(`<a href="${create}"`).length - 1).toBe(2); // cabeçalho: desktop e celular
		expect(html).toMatch(/<a href="\/lobbies\/novo\?dia=2026-09-30"[^>]*aria-label="Criar lobby"/);
		for (const l of today) {
			expect(html).toContain(`href="/lobbies/${l.id}"`);
			expect(html).toContain(`aria-label="Ver grupo: ${l.instance} às ${l.time}"`);
		}
		expect(html).toMatch(/<a href="\/lobbies\/[^"]+"[^>]*>(?:\s|<!--[^>]*-->)*Ver grupo/);
	});

	it('RN-03 (lobbies): todas as instâncias usam a capa e o ícone padrão', () => {
		expect(html).not.toContain('/brand/inst-');
		expect(html).toContain('/brand/c1-symbol-dark.svg');
	});

	it('CA-05.1: o destaque aparece com o próximo grupo com vaga', () => {
		expect(html).toContain('data-testid="featured-lobby"');
		expect(html).toContain('Próximo grupo com vaga');
	});

	it('CA-03.1: o seletor mostra 14 dias, com hoje selecionado', () => {
		expect(html.match(/role="tab"/g)).toHaveLength(14);
		expect(html).toMatch(/aria-selected="true"[^>]*data-date="2026-09-30"/);
	});

	it('RN-22: sem marca da Gravity e sem nomes oficiais de mapa ou monstro na tela', () => {
		expect(html).not.toMatch(/Gravity|Prontera|Glast Heim|Poring/i);
	});
});

const GRIMBOLD = {
	id: '6f1c2b8e-3a4d-4e5f-9a1b-2c3d4e5f6a7b',
	username: 'grimbold',
	globalName: 'Grimbold'
};

describe('Home com login (spec login-discord)', () => {
	it('RN-14: visitante vê "Entrar com Discord" como link para o login, não desabilitado', () => {
		const html = renderHome();
		expect(html).toMatch(
			/<a href="\/auth\/discord\/login\?next=%2F"[^>]*>[\s\S]*?Entrar com Discord/
		);
		expect(html).not.toMatch(/aria-disabled="true"[^>]*>[\s\S]{0,600}?Entrar com Discord/);
	});

	it('CA-02.1 / CA-02.3: logado, o HTML do servidor já traz a inicial e o nome, sem "Entrar com Discord"', () => {
		const html = renderHome(GRIMBOLD);
		expect(html).toContain('data-testid="user-menu"');
		expect(html).toMatch(/class="tile[^"]*"[^>]*>G</);
		expect(html).toContain('Grimbold');
		expect(html).not.toContain('Entrar com Discord');
	});

	it('CA-06.2 / RN-18 (personagens): o menu do usuário tem "Meu perfil", para /perfil, acima de "Sair"', () => {
		const html = renderHome(GRIMBOLD);
		const menu = html.match(/<div[^>]*role="menu"[^>]*>[\s\S]*?<\/form>/)?.[0] ?? '';
		expect(menu).toMatch(/<div[^>]*hidden/); // fechado até o Usuário abrir
		expect(menu).toMatch(/<a href="\/perfil"[^>]*role="menuitem"[^>]*>[\s\S]*?Meu perfil/);
		expect(menu.indexOf('Meu perfil')).toBeLessThan(menu.indexOf('Sair'));
		expect(menu.match(/role="menuitem"/g)).toHaveLength(2);
	});

	it('CA-02.2: sem nome de exibição, mostra o nome de usuário e a inicial dele', () => {
		const html = renderHome({ ...GRIMBOLD, username: 'mirai.exe', globalName: null });
		expect(html).toMatch(/class="tile[^"]*"[^>]*>M</);
		expect(html).toContain('mirai.exe');
	});

	it('CA-02.4 / RNF-05: logado, nenhuma imagem de fora do servidor', () => {
		const html = renderHome(GRIMBOLD);
		const external = [...html.matchAll(/(?:src|href)="(https?:\/\/[^"]+)"/g)].map((m) => m[1]);
		expect(external).toEqual([]);
		expect(html).not.toContain('cdn.discordapp.com');
	});

	it('CA-04.1 / RN-13: com ?login=erro, a Home mostra a mensagem de falha', () => {
		const html = renderHome(null, true);
		expect(html).toMatch(
			/role="alert"[^>]*>Não foi possível entrar com o Discord\. Tente de novo\.</
		);
		expect(renderHome()).not.toContain('role="alert"');
	});
});

describe('aviso de lista vazia', () => {
	it('CA-02.3 / RN-23 (lobbies): "Criar lobby" funciona no dia vazio e nos filtros sem resultado', () => {
		for (const kind of ['day', 'filters'] as const) {
			const html = render(EmptyState, { props: { kind, onreset: () => {} } }).body;
			expect(html).toMatch(/<a href="\/lobbies\/novo"[^>]*>(?:\s|<[^>]+>)*Criar lobby/);
			expect(html).not.toContain('Disponível em breve');
		}
	});

	it('CA-02.5 / RN-24 (lobbies): com o dia escolhido, "Criar lobby" abre a criação nele', () => {
		const html = render(EmptyState, {
			props: { kind: 'day', onreset: () => {}, createDate: '2026-10-09' }
		}).body;
		expect(html).toContain('href="/lobbies/novo?dia=2026-10-09"');
	});
});
