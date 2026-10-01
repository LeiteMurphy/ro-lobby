import { expect, test, type Page } from '@playwright/test';

// Ponta a ponta da Home (spec home-local). Os dados são fictícios e as datas dependem do
// relógio real, então os testes leem dias e contagens da própria tela.

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

test.describe('Home no desktop', () => {
	test.use({ viewport: { width: 1280, height: 900 } });

	test('CA-02.1: grupos de hoje em ordem de horário, com os dados do card', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByRole('heading', { level: 1 })).toHaveText('Grupos para hoje');
		await expect(cards(page)).toHaveCount(await dayCount(page, 0));
		const times = await cardTimes(page);
		expect(times).toEqual([...times].sort());
		const first = cards(page).first();
		await expect(first.getByRole('heading', { level: 3 })).not.toBeEmpty();
		await expect(first).toContainText(/Nv \d+\+/);
		await expect(first).toContainText(/\d+\/\d+/);
	});

	test('CA-02.6: o HTML do servidor já traz os cards de hoje', async ({ request }) => {
		const html = await (await request.get('/')).text();
		expect(html).toContain('data-testid="lobby-card"');
		expect(html).toContain('Grupos para hoje');
	});

	test('CA-02.5: ações sem backend desabilitadas, com "Disponível em breve"', async ({ page }) => {
		await page.goto('/');
		for (const name of ['Criar lobby', 'Entrar com Discord']) {
			await expect(
				page.getByRole('banner').getByRole('button', { name, exact: true })
			).toBeDisabled();
		}
		const apply = cards(page).getByRole('button', { name: 'Candidatar' }).first();
		await expect(apply).toBeDisabled();
		await apply.hover();
		await expect(page.getByRole('tooltip', { name: 'Disponível em breve' }).first()).toBeVisible();
		// Clicar não faz nada: a tela continua igual.
		await apply.click({ force: true });
		await expect(page).toHaveURL('/');
	});

	test('CA-03.2: trocar o dia mostra os grupos daquele dia', async ({ page }) => {
		await page.goto('/');
		const count = await dayCount(page, 1);
		await dayTabs(page).nth(1).click();
		await expect(dayTabs(page).nth(1)).toHaveAttribute('aria-selected', 'true');
		await expect(page.getByRole('heading', { level: 1 })).toHaveText(
			/^Grupos para \w{3}, \d+ \w{3}$/
		);
		await expect(cards(page)).toHaveCount(count);
		const times = await cardTimes(page);
		expect(times).toEqual([...times].sort());
	});

	test('CA-03.3: dia sem grupos mostra "Nenhum grupo neste dia"', async ({ page }) => {
		await page.goto('/');
		let index = -1;
		for (let i = 0; i < 14 && index < 0; i++) if ((await dayCount(page, i)) === 0) index = i;
		expect(index, 'os dados fictícios têm dias sem grupos').toBeGreaterThan(0);
		await dayTabs(page).nth(index).click();
		await expect(page.getByTestId('empty-state')).toContainText('Nenhum grupo neste dia');
		await expect(page.getByRole('button', { name: 'Limpar filtros' })).toHaveCount(0);
	});

	test('CA-04.1: filtro de instância', async ({ page }) => {
		await page.goto('/');
		const sidebar = page.getByRole('complementary', { name: 'Filtros' });
		await sidebar.getByRole('combobox', { name: 'Instância' }).selectOption('Torre sem fim');
		await expect(cards(page).first()).toBeVisible();
		for (const label of await cards(page).evaluateAll((els) =>
			els.map((e) => e.getAttribute('aria-label'))
		)) {
			expect(label).toMatch(/^Torre sem fim às /);
		}
		await expect(page.getByTestId('list-subheading')).toContainText(/\d+ de \d+ grupos/);
	});

	test('CA-04.7: filtros sem resultado e "Limpar filtros"', async ({ page }) => {
		await page.goto('/');
		const total = await dayCount(page, 0);
		const sidebar = page.getByRole('complementary', { name: 'Filtros' });
		// Nos dados fictícios, a Caverna de gelo é às 18:00, fora da faixa 22h–00h.
		await sidebar.getByRole('combobox', { name: 'Instância' }).selectOption('Caverna de gelo');
		await sidebar.getByText('22h–00h').click();
		await expect(page.getByTestId('empty-state')).toContainText('Nenhum grupo com esses filtros');
		await page.getByRole('button', { name: 'Limpar filtros' }).click();
		await expect(cards(page)).toHaveCount(total);
	});

	test('RNF-01: dá para trocar o dia e filtrar só com o teclado, com foco visível', async ({
		page
	}) => {
		await page.goto('/');
		const ring = (selector: string) =>
			page.evaluate((s) => getComputedStyle(document.querySelector(s)!).boxShadow, selector);

		// Dia: Tab chega no primeiro dia, Tab vai para o próximo e Enter seleciona.
		await dayTabs(page).first().focus();
		await page.keyboard.press('Tab');
		await expect(dayTabs(page).nth(1)).toBeFocused();
		expect(await page.evaluate(() => getComputedStyle(document.activeElement!).boxShadow)).not.toBe(
			'none'
		);
		await page.keyboard.press('Enter');
		await expect(dayTabs(page).nth(1)).toHaveAttribute('aria-selected', 'true');
		await expect(page.getByRole('heading', { level: 1 })).not.toHaveText('Grupos para hoje');

		// Filtro: o checkbox recebe foco pelo teclado, mostra o anel e marca com Espaço.
		const sidebar = page.getByRole('complementary', { name: 'Filtros' });
		const tank = sidebar.getByRole('checkbox', { name: 'Tank' });
		await sidebar.getByRole('combobox', { name: 'Instância' }).focus();
		await page.keyboard.press('Tab');
		await expect(tank).toBeFocused();
		expect(await ring('aside.sidebar input:focus-visible + span')).not.toBe('none');
		await page.keyboard.press('Space');
		await expect(tank).toBeChecked();
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
		await page.getByRole('button', { name: /^Filtros/ }).click();
		const drawer = page.getByRole('dialog', { name: 'Filtros' });
		await drawer.getByText('Tank', { exact: true }).click();
		await drawer.getByText('18h–20h').click();
		await drawer.getByRole('button', { name: /^Ver \d+ grupos?$/ }).click();
		await expect(page.getByRole('button', { name: /^Filtros/ })).toHaveText('Filtros (2)');
	});

	test('RN-19: "Criar lobby" vira botão só com ícone', async ({ page }) => {
		await page.goto('/');
		const banner = page.getByRole('banner');
		const create = banner.getByRole('button', { name: 'Criar lobby', exact: true });
		// A versão com texto fica com display:none e sai da árvore de acessibilidade.
		await expect(create).toHaveCount(1);
		await expect(create).toBeVisible();
		await expect(create).toHaveText('');
	});
});
