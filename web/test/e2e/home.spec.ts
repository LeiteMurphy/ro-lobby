import { expect, test, type Page } from '@playwright/test';
import { apiLogin, createCharacter, createLobby, rand, spDay } from './seed';

// Ponta a ponta da Home (spec home-local). As datas dependem do relógio real, então os
// testes leem dias e contagens da própria tela.

const cards = (page: Page) => page.getByTestId('lobby-card');
const dayTabs = (page: Page) => page.getByRole('tablist', { name: 'Dias' }).getByRole('tab');

async function dayCount(page: Page, index: number): Promise<number> {
	const text = (await dayTabs(page).nth(index).locator('span').last().innerText()).trim();
	return text === '—' ? 0 : Number(text);
}

async function cardTimes(page: Page): Promise<string[]> {
	await expect(cards(page).first()).toBeVisible();
	return cards(page).evaluateAll((els) =>
		els.map((el) => (el.getAttribute('aria-label') ?? '').split(' às ')[1])
	);
}

/**
 * RNF-01: percorre a página com Tab e, em cada ponto, exige que o estilo com foco seja
 * diferente do estilo sem foco e tenha o âmbar do design system (#f6bb45, opaco no anel ou
 * translúcido no halo dos campos). Para checkbox e radio, o
 * anel fica na caixa desenhada ao lado do input; para o select, no campo em volta.
 */
async function expectFocusRingOnEveryTabStop(page: Page) {
	// Mede o estado final de cada controle, sem pegar o meio de uma transição.
	await page.addStyleTag({ content: '*,*::before,*::after{transition:none!important}' });
	await page.locator('body').focus();
	const stops: { name: string; focused: string; idx: number }[] = [];
	for (let i = 0; i < 80; i++) {
		await page.keyboard.press('Tab');
		const stop = await page.evaluate((idx) => {
			const el = document.activeElement as HTMLElement | null;
			if (!el || el === document.body || el.dataset.tabStop) return null;
			el.dataset.tabStop = String(idx);
			const target =
				el instanceof HTMLInputElement && ['checkbox', 'radio'].includes(el.type)
					? (el.nextElementSibling as HTMLElement)
					: el instanceof HTMLSelectElement
						? (el.parentElement as HTMLElement)
						: el;
			const name =
				el.getAttribute('aria-label') ||
				el.closest('label')?.textContent?.trim() ||
				el.textContent?.trim() ||
				el.tagName;
			return { name, focused: getComputedStyle(target).boxShadow, idx };
		}, i);
		if (!stop) break;
		stops.push(stop);
	}
	expect(stops.length).toBeGreaterThan(10);

	await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
	await page.mouse.move(0, 0);
	const unfocused = await page.evaluate(() =>
		Object.fromEntries(
			[...document.querySelectorAll<HTMLElement>('[data-tab-stop]')].map((el) => {
				const target =
					el instanceof HTMLInputElement && ['checkbox', 'radio'].includes(el.type)
						? (el.nextElementSibling as HTMLElement)
						: el instanceof HTMLSelectElement
							? (el.parentElement as HTMLElement)
							: el;
				return [el.dataset.tabStop, getComputedStyle(target).boxShadow];
			})
		)
	);

	const missing = stops.filter(
		// O âmbar pode vir opaco (anel) ou translúcido (halo dos campos, como no design system).
		(s) => s.focused === unfocused[s.idx] || !/rgba?\(246, 187, 69/.test(s.focused)
	);
	expect(missing.map((s) => s.name)).toEqual([]);
}

// Spec lobbies (T-10): a Home mostra os lobbies reais. Antes dos testes, uma conta cria
// lobbies conhecidos amanhã (dia 1) e depois de amanhã (dia 2), direto na API; o banco do
// ponta a ponta começa vazio. Outros arquivos podem criar lobbies em paralelo, então os
// testes conferem os lobbies desta conta pelo anfitrião, e as contagens pela própria tela.
const s = rand();
const host = { tank: `Tk${s}`, support: `Sp${s}`, dps: `Dp${s}` };

test.beforeAll(async ({ request }) => {
	const token = await apiLogin(request, `home${s}`);
	const tank = await createCharacter(request, token, {
		nick: host.tank,
		classId: 'guardiao-real',
		level: 200,
		role: 'tank'
	});
	const support = await createCharacter(request, token, {
		nick: host.support,
		classId: 'arcebispo',
		level: 200,
		role: 'support'
	});
	const dps = await createCharacter(request, token, {
		nick: host.dps,
		classId: 'feiticeiro',
		level: 200,
		role: 'dps'
	});
	const lobbies = [
		{ instanceId: 'vila-dos-porings', day: 1, time: '18:00', characterId: tank.id, minLevel: 30 },
		{
			instanceId: 'mansao-da-desilusao',
			day: 1,
			time: '19:00',
			characterId: support.id,
			minLevel: 200
		},
		{
			instanceId: 'templo-do-demonio-rei',
			day: 1,
			time: '21:30',
			characterId: dps.id,
			minLevel: 160
		},
		// Lotado: só a vaga do dono.
		{
			instanceId: 'caverna-de-mors',
			day: 1,
			time: '23:00',
			characterId: support.id,
			minLevel: 160,
			slots: { tank: 0, support: 1, dps: 0 }
		},
		{ instanceId: 'sala-final', day: 2, time: '20:00', characterId: tank.id, minLevel: 150 }
	];
	for (const l of lobbies) await createLobby(request, token, l);
});

const ownCards = (page: Page) =>
	cards(page).filter({ hasText: new RegExp(`${host.tank}|${host.support}|${host.dps}`) });

test.describe('Home no desktop', () => {
	test.use({ viewport: { width: 1280, height: 900 } });

	test('CA-02.1 / CA-02.1 (lobbies): os grupos do dia em ordem de horário, com os dados do card', async ({
		page
	}) => {
		await page.goto('/');
		await expect(page.getByRole('heading', { level: 1 })).toHaveText('Grupos para hoje');
		await dayTabs(page).nth(1).click();
		await expect(cards(page)).toHaveCount(await dayCount(page, 1));
		const times = await cardTimes(page);
		expect(times).toEqual([...times].sort());
		await expect(ownCards(page)).toHaveCount(4);
		const vila = cards(page).filter({ hasText: 'Vila dos Porings' }).filter({ hasText: host.tank });
		await expect(vila).toContainText('18:00');
		await expect(vila).toContainText('Guardião Real');
		await expect(vila).toContainText('Nv 30+');
		await expect(vila).toContainText('1/6');
		// RN-14 (home-local): o lobby só com a vaga do dono está lotado.
		await expect(
			cards(page).filter({ hasText: 'Caverna de Mors' }).filter({ hasText: host.support })
		).toHaveAttribute('data-edge', 'full');
	});

	test('CA-02.6: o HTML do servidor já traz a Home com o seletor de dias', async ({ request }) => {
		const html = await (await request.get('/')).text();
		expect(html).toContain('Grupos para hoje');
		expect(html.match(/role="tab"/g)).toHaveLength(14);
	});

	test('CA-02.5 / CA-10.6 (candidatura-lobby): "Candidatar" do card abre o lobby', async ({
		page
	}) => {
		await page.goto('/');
		await dayTabs(page).nth(1).click();
		// RN-14 (login-discord) e RN-23 (lobbies): "Entrar com Discord" e "Criar lobby" já funcionam.
		await expect(
			page.getByRole('banner').getByRole('link', { name: 'Entrar com Discord' })
		).toHaveAttribute('href', '/auth/discord/login?next=%2F');
		await expect(
			page.getByRole('banner').getByRole('link', { name: 'Criar lobby' })
		).toHaveAttribute('href', `/lobbies/novo?dia=${spDay(1)}`);
		// RN-35: a candidatura acontece no detalhe do lobby.
		const apply = ownCards(page).getByRole('link', { name: 'Candidatar' }).first();
		const href = await apply.getAttribute('href');
		expect(href).toMatch(/^\/lobbies\/[0-9a-f-]{36}$/);
		await apply.click();
		await expect(page).toHaveURL(href!);
		await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
	});

	test('CA-03.2: trocar o dia mostra os grupos daquele dia', async ({ page }) => {
		await page.goto('/');
		const count = await dayCount(page, 2);
		expect(count).toBeGreaterThan(0);
		await dayTabs(page).nth(2).click();
		await expect(dayTabs(page).nth(2)).toHaveAttribute('aria-selected', 'true');
		await expect(page.getByRole('heading', { level: 1 })).toHaveText(
			/^Grupos para \w{3}, \d+ \w{3}$/
		);
		await expect(cards(page)).toHaveCount(count);
		await expect(ownCards(page)).toHaveCount(1);
	});

	test('CA-03.3: dia sem grupos mostra "Nenhum grupo neste dia"', async ({ page }) => {
		await page.goto('/');
		let index = -1;
		for (let i = 13; i > 0 && index < 0; i--) if ((await dayCount(page, i)) === 0) index = i;
		expect(index, 'algum dia sem grupos no banco do ponta a ponta').toBeGreaterThan(0);
		await dayTabs(page).nth(index).click();
		await expect(page.getByTestId('empty-state')).toContainText('Nenhum grupo neste dia');
		await expect(page.getByRole('button', { name: 'Limpar filtros' })).toHaveCount(0);
		// CA-02.3 / RN-23 (lobbies): o "Criar lobby" do aviso leva à criação (o visitante passa
		// pelo login antes, RN-04).
		const create = page.getByTestId('empty-state').getByRole('link', { name: 'Criar lobby' });
		await expect(create).toHaveAttribute('href', `/lobbies/novo?dia=${spDay(index)}`);
		await create.click();
		await expect(page).toHaveURL(/\/oauth2\/authorize/);
	});

	test('CA-04.1: filtro de instância', async ({ page }) => {
		await page.goto('/');
		await dayTabs(page).nth(1).click();
		const sidebar = page.getByRole('complementary', { name: 'Filtros' });
		await sidebar.getByRole('combobox', { name: 'Instância' }).selectOption('Vila dos Porings');
		await expect(cards(page).first()).toBeVisible();
		for (const label of await cards(page).evaluateAll((els) =>
			els.map((e) => e.getAttribute('aria-label'))
		)) {
			expect(label).toMatch(/^Vila dos Porings às /);
		}
		await expect(page.getByTestId('list-subheading')).toContainText(/\d+ de \d+ grupos?/);
	});

	test('CA-04.7: filtros sem resultado e "Limpar filtros"', async ({ page }) => {
		await page.goto('/');
		await dayTabs(page).nth(1).click();
		const total = await dayCount(page, 1);
		const sidebar = page.getByRole('complementary', { name: 'Filtros' });
		// Os lobbies de Vila dos Porings do ponta a ponta são às 18:00, fora da faixa 22h–00h.
		await sidebar.getByRole('combobox', { name: 'Instância' }).selectOption('Vila dos Porings');
		await sidebar.getByText('22h–00h').click();
		await expect(page.getByTestId('empty-state')).toContainText('Nenhum grupo com esses filtros');
		await expect(
			page.getByTestId('empty-state').getByRole('link', { name: 'Criar lobby' })
		).toHaveAttribute('href', `/lobbies/novo?dia=${spDay(1)}`);
		await page.getByRole('button', { name: 'Limpar filtros' }).click();
		await expect(cards(page)).toHaveCount(total);
	});

	test('RNF-01: dá para trocar o dia e filtrar só com o teclado', async ({ page }) => {
		await page.goto('/');
		await dayTabs(page).first().focus();
		await page.keyboard.press('Tab');
		await expect(dayTabs(page).nth(1)).toBeFocused();
		await page.keyboard.press('Enter');
		await expect(dayTabs(page).nth(1)).toHaveAttribute('aria-selected', 'true');
		await expect(page.getByRole('heading', { level: 1 })).not.toHaveText('Grupos para hoje');

		const sidebar = page.getByRole('complementary', { name: 'Filtros' });
		const tank = sidebar.getByRole('checkbox', { name: 'Tank' });
		await sidebar.getByRole('combobox', { name: 'Instância' }).focus();
		await page.keyboard.press('Tab');
		await expect(tank).toBeFocused();
		await page.keyboard.press('Space');
		await expect(tank).toBeChecked();
	});

	test('RNF-01: todo ponto de Tab mostra o anel de foco âmbar', async ({ page }) => {
		await page.goto('/');
		await dayTabs(page).nth(1).click();
		await expectFocusRingOnEveryTabStop(page);
	});

	test('RNF-02: a Home não carrega nada de fora do próprio servidor', async ({ page, baseURL }) => {
		const origins = new Set<string>();
		page.on('request', (req) => origins.add(new URL(req.url()).origin));
		await page.goto('/');
		await page.waitForLoadState('networkidle');
		expect(
			[...origins].filter((o) => o !== new URL(baseURL!).origin && !o.startsWith('data:'))
		).toEqual([]);
	});
});

test.describe('Home no celular', () => {
	test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });

	test('CA-06.1: filtros numa gaveta, sem a barra lateral', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByRole('complementary', { name: 'Filtros' })).toBeHidden();
		await page.getByRole('button', { name: /^Filtros/ }).click();
		const drawer = page.getByRole('dialog', { name: 'Filtros' });
		await expect(drawer).toBeVisible();
		await expect(drawer.getByRole('combobox', { name: 'Instância' })).toBeVisible();
		await drawer.getByRole('button', { name: 'Fechar' }).click();
		await expect(drawer).toBeHidden();
	});

	test('CA-06.2: botão mostra "Filtros (2)" com 2 filtros ativos', async ({ page }) => {
		await page.goto('/');
		await dayTabs(page).nth(1).click();
		await page.getByRole('button', { name: /^Filtros/ }).click();
		const drawer = page.getByRole('dialog', { name: 'Filtros' });
		await drawer.getByText('Tank', { exact: true }).click();
		await drawer.getByText('18h–20h').click();
		await drawer.getByRole('button', { name: /^Ver \d+ grupos?$/ }).click();
		await expect(page.getByRole('button', { name: /^Filtros/ })).toHaveText('Filtros (2)');
	});

	test('RNF-01: todo ponto de Tab mostra o anel de foco âmbar no celular', async ({ page }) => {
		await page.goto('/');
		await expectFocusRingOnEveryTabStop(page);
	});

	test('RN-19: "Criar lobby" vira botão só com ícone', async ({ page }) => {
		await page.goto('/');
		const create = page.getByRole('banner').getByRole('link', { name: 'Criar lobby' });
		// A versão com texto fica com display:none e sai da árvore de acessibilidade.
		await expect(create).toHaveCount(1);
		await expect(create).toBeVisible();
		await expect(create).toHaveText('');
	});
});
