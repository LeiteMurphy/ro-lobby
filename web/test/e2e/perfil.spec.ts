import { expect, test, type Page } from '@playwright/test';

// Ponta a ponta do perfil e dos personagens (spec personagens, T-08), contra a API e o
// Discord falso que o playwright.config.ts sobe. Cada teste entra com um usuário novo do
// Discord falso e usa nicks com sufixo aleatório, para não colidir com execuções antigas
// no banco de desenvolvimento (D-11).

const rand = () => Math.random().toString(36).slice(2, 7);

async function loginAs(page: Page, username: string, from = '/perfil') {
	await page.goto(from);
	if (!page.url().includes('/oauth2/authorize')) {
		await page.getByRole('banner').getByRole('link', { name: 'Entrar com Discord' }).click();
	}
	await expect(page.getByRole('heading', { name: 'Discord falso' })).toBeVisible();
	await page.getByLabel('Entrar como (opcional)').fill(username);
	await page.getByRole('button', { name: 'Autorizar' }).click();
	await page.waitForURL(
		(url) => !url.href.includes('/oauth2/') && !url.pathname.startsWith('/auth/')
	);
}

const cards = (page: Page) => page.getByTestId('character-card');
const card = (page: Page, nick: string) =>
	cards(page).filter({ has: page.getByRole('heading', { name: nick, exact: true }) });
const dialog = (page: Page) => page.getByTestId('character-dialog');

interface NewCharacter {
	nick: string;
	classId?: string;
	level?: number;
	role?: 'Tank' | 'Suporte' | 'Dano';
	portrait?: string;
	link?: string;
}

async function fillDialog(page: Page, c: NewCharacter) {
	const d = dialog(page);
	await expect(d).toBeVisible();
	if (c.portrait)
		await d.getByRole('radio', { name: new RegExp(`^${c.portrait}`) }).check({ force: true });
	await d.getByLabel('Nick').fill(c.nick);
	await d.getByLabel('Classe').selectOption(c.classId ?? 'guardiao-real');
	await d.getByLabel('Nível').fill(String(c.level ?? 172));
	await d.getByRole('radio', { name: c.role ?? 'Tank' }).check();
	if (c.link) await d.getByLabel('Link externo (opcional)').fill(c.link);
}

async function addCharacter(page: Page, c: NewCharacter) {
	await page
		.getByRole('main')
		.getByRole('button', { name: 'Adicionar personagem' })
		.first()
		.click();
	await fillDialog(page, c);
	await dialog(page).getByRole('button', { name: 'Salvar personagem' }).click();
	await expect(dialog(page)).toHaveCount(0);
	await expect(card(page, c.nick)).toBeVisible();
}

async function openMenu(page: Page, nick: string) {
	await page.getByRole('button', { name: `Ações de ${nick}` }).click();
	return page.getByRole('menu', { name: `Ações de ${nick}` });
}

test.describe('perfil e personagens', () => {
	test.use({ viewport: { width: 1280, height: 900 } });

	test('CA-06.1: visitante que abre /perfil passa pelo login e volta para /perfil', async ({
		page
	}) => {
		await page.goto('/perfil');
		await expect(page).toHaveURL(/\/oauth2\/authorize/);
		await page.getByLabel('Entrar como (opcional)').fill(`v${rand()}`);
		await page.getByRole('button', { name: 'Autorizar' }).click();
		await expect(page).toHaveURL('/perfil');
		await expect(page.getByRole('heading', { name: 'Meus personagens' })).toBeVisible();
	});

	test('CA-06.2: "Meu perfil" no menu do usuário leva a /perfil', async ({ page }) => {
		await loginAs(page, `m${rand()}`, '/');
		await page.getByRole('button', { name: /^Menu de / }).click();
		const items = page.getByRole('menuitem');
		await expect(items).toHaveText(['Meu perfil', 'Sair']);
		await page.getByRole('menuitem', { name: 'Meu perfil' }).click();
		await expect(page).toHaveURL('/perfil');
	});

	test('CA-01.2 / CA-02.1 / CA-02.2 / CA-01.1 / CA-01.4: do perfil vazio às cartas', async ({
		page
	}) => {
		const s = rand();
		await loginAs(page, `a${s}`);
		await expect(page.getByText('Você ainda não tem personagens')).toBeVisible();

		// CA-02.1 e CA-02.2: o primeiro vira principal e fica com o retrato padrão.
		await addCharacter(page, { nick: `Brasa${s}` });
		const brasa = card(page, `Brasa${s}`);
		await expect(brasa).toContainText('Principal');
		await expect(brasa.getByRole('img').first()).toHaveAttribute('src', '/portraits/retrato-1.svg');

		await addCharacter(page, {
			nick: `Lirien${s}`,
			classId: 'arcebispo',
			level: 178,
			role: 'Suporte',
			portrait: 'Retrato 3',
			link: 'https://exemplo.com/char/123'
		});
		const lirien = card(page, `Lirien${s}`);
		await expect(lirien.getByRole('img').first()).toHaveAttribute(
			'src',
			'/portraits/retrato-3.svg'
		);
		await expect(lirien).not.toContainText('Principal');

		// CA-01.1: principal primeiro; cada carta com classe, nível e função.
		await expect(cards(page).getByRole('heading')).toHaveText([`Brasa${s}`, `Lirien${s}`]);
		await expect(brasa).toContainText('Guardião Real');
		await expect(brasa).toContainText('172');
		await expect(brasa).toContainText('Tank');
		await expect(lirien).toContainText('Arcebispo');
		await expect(lirien).toContainText('Suporte');

		// CA-01.4: link em outra aba, sem passar a origem.
		const link = lirien.getByRole('link', { name: /Perfil externo/ });
		await expect(link).toHaveAttribute('href', 'https://exemplo.com/char/123');
		await expect(link).toHaveAttribute('target', '_blank');
		await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
		await expect(page.getByText('2 de 10 personagens')).toBeVisible();
	});

	test('CA-02.3 / CA-02.9 / CA-01.3: o nick de outra conta é recusado junto do campo', async ({
		browser
	}) => {
		const s = rand();
		const ana = await (await browser.newContext()).newPage();
		await loginAs(ana, `ana${s}`);
		await addCharacter(ana, { nick: `Brasa${s}` });

		const bia = await (await browser.newContext()).newPage();
		await loginAs(bia, `bia${s}`);
		// CA-01.3: a bia não vê o personagem da ana.
		await expect(bia.getByText('Você ainda não tem personagens')).toBeVisible();

		await bia
			.getByRole('main')
			.getByRole('button', { name: 'Adicionar personagem' })
			.first()
			.click();
		await fillDialog(bia, {
			nick: `brasa${s}`,
			classId: 'paladino',
			level: 99,
			role: 'Dano',
			link: 'https://exemplo.com/bia'
		});
		await dialog(bia).getByRole('button', { name: 'Salvar personagem' }).click();

		const nick = dialog(bia).getByLabel('Nick');
		await expect(nick).toHaveAttribute('aria-invalid', 'true');
		const errorId = await nick.getAttribute('aria-describedby');
		await expect(bia.locator(`[id="${errorId}"]`)).toHaveText('Esse nick já está em uso');
		// RN-19: o resto continua preenchido.
		await expect(nick).toHaveValue(`brasa${s}`);
		await expect(dialog(bia).getByLabel('Classe')).toHaveValue('paladino');
		await expect(dialog(bia).getByLabel('Nível')).toHaveValue('99');
		await expect(dialog(bia).getByRole('radio', { name: 'Dano' })).toBeChecked();
		await expect(dialog(bia).getByLabel('Link externo (opcional)')).toHaveValue(
			'https://exemplo.com/bia'
		);
		await expect(cards(bia)).toHaveCount(0);
	});

	test('CA-03.1 / CA-05.1 / CA-04.2 / CA-04.3 / CA-04.1: editar, trocar o principal e excluir', async ({
		browser
	}) => {
		const s = rand();
		const page = await (await browser.newContext()).newPage();
		await loginAs(page, `e${s}`);
		await addCharacter(page, { nick: `Brasa${s}` });
		await addCharacter(page, {
			nick: `Lirien${s}`,
			classId: 'arcebispo',
			level: 178,
			role: 'Suporte'
		});
		await addCharacter(page, {
			nick: `Faisca${s}`,
			classId: 'feiticeiro',
			level: 165,
			role: 'Dano'
		});

		// CA-03.1: editar classe e nível.
		await (await openMenu(page, `Faisca${s}`)).getByRole('menuitem', { name: 'Editar' }).click();
		await expect(dialog(page).getByLabel('Nick')).toHaveValue(`Faisca${s}`);
		await dialog(page).getByLabel('Classe').selectOption('elementalista');
		await dialog(page).getByLabel('Nível').fill('170');
		await dialog(page).getByRole('button', { name: 'Salvar personagem' }).click();
		await expect(dialog(page)).toHaveCount(0);
		await expect(card(page, `Faisca${s}`)).toContainText('Elementalista');
		await expect(card(page, `Faisca${s}`)).toContainText('170');

		// CA-05.1: Lirien vira o principal e vai para o começo.
		await (
			await openMenu(page, `Lirien${s}`)
		)
			.getByRole('menuitem', { name: 'Tornar principal' })
			.click();
		await expect(cards(page).getByRole('heading')).toHaveText([
			`Lirien${s}`,
			`Brasa${s}`,
			`Faisca${s}`
		]);
		await expect(card(page, `Lirien${s}`)).toContainText('Principal');
		await expect(card(page, `Brasa${s}`)).not.toContainText('Principal');
		// O principal não oferece "Tornar principal".
		const menu = await openMenu(page, `Lirien${s}`);
		await expect(menu.getByRole('menuitem')).toHaveText(['Editar', 'Excluir']);
		await page.keyboard.press('Escape');

		// CA-04.2: cancelar a exclusão não muda nada.
		await (await openMenu(page, `Lirien${s}`)).getByRole('menuitem', { name: 'Excluir' }).click();
		const confirm = page.getByTestId('confirm-dialog');
		await expect(confirm).toContainText(`Excluir Lirien${s}?`);
		await confirm.getByRole('button', { name: 'Cancelar' }).click();
		await expect(confirm).toHaveCount(0);
		await expect(cards(page)).toHaveCount(3);

		// CA-04.3: excluir o principal passa a vez ao mais antigo (Brasa).
		await (await openMenu(page, `Lirien${s}`)).getByRole('menuitem', { name: 'Excluir' }).click();
		await page.getByTestId('confirm-dialog').getByRole('button', { name: 'Excluir' }).click();
		await expect(cards(page)).toHaveCount(2);
		await expect(cards(page).getByRole('heading')).toHaveText([`Brasa${s}`, `Faisca${s}`]);
		await expect(card(page, `Brasa${s}`)).toContainText('Principal');

		// CA-04.1: o nick excluído fica livre.
		await addCharacter(page, { nick: `lirien${s}` });
	});

	test('CA-02.8: com 10 personagens, "Adicionar personagem" fica desabilitado', async ({
		page
	}) => {
		test.slow();
		const s = rand();
		await loginAs(page, `l${s}`);
		for (let i = 0; i < 10; i++) await addCharacter(page, { nick: `L${s}n${i}` });
		const add = page.getByRole('main').getByRole('button', { name: 'Adicionar personagem' });
		await expect(add).toBeDisabled();
		await expect(page.getByText('Você já tem 10 personagens')).toBeVisible();
		await expect(page.getByText('10 de 10 personagens')).toBeVisible();
	});

	test('RNF-01: menu "…" e diálogos pelo teclado', async ({ page }) => {
		const s = rand();
		await loginAs(page, `k${s}`);
		await addCharacter(page, { nick: `Brasa${s}` });
		await addCharacter(page, { nick: `Lirien${s}` });

		const trigger = page.getByRole('button', { name: `Ações de Lirien${s}` });
		await trigger.focus();
		await page.keyboard.press('Enter');
		await expect(page.getByRole('menuitem', { name: 'Tornar principal' })).toBeFocused();
		await page.keyboard.press('ArrowDown');
		await expect(page.getByRole('menuitem', { name: 'Editar' })).toBeFocused();
		await page.keyboard.press('Escape');
		await expect(page.getByRole('menu')).toHaveCount(0);
		await expect(trigger).toBeFocused();

		// O diálogo abre pelo teclado, prende o foco e fecha com Escape.
		await page.keyboard.press('Enter');
		await page.keyboard.press('ArrowDown');
		await page.keyboard.press('Enter');
		await expect(dialog(page)).toBeVisible();
		await expect(dialog(page).getByLabel('Nick')).toHaveValue(`Lirien${s}`);
		await page.keyboard.press('Escape');
		await expect(dialog(page)).toHaveCount(0);
	});

	test('RNF-02: /perfil não carrega nada de fora do servidor', async ({ page, baseURL }) => {
		const s = rand();
		await loginAs(page, `o${s}`);
		await addCharacter(page, { nick: `Brasa${s}`, link: 'https://exemplo.com/x' });
		const origins = new Set<string>();
		page.on('request', (req) => origins.add(new URL(req.url()).origin));
		await page.reload();
		await page.waitForLoadState('networkidle');
		expect([...origins].filter((o) => o !== new URL(baseURL!).origin)).toEqual([]);
	});
});
