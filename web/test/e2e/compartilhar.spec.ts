import { expect, test } from '@playwright/test';
import { apiLogin, createCharacter, createLobby, rand, spDay } from './seed';

// Ponta a ponta do compartilhamento (spec compartilhar-lobby): o convite copiado e o
// diálogo quando o navegador recusa a área de transferência. O lobby fica no dia 4, longe
// dos da Home.

const WEEKDAYS = ['domingo', 'segunda', 'terça', 'quarta', 'quinta', 'sexta', 'sábado'];

async function seedLobby(request: Parameters<typeof apiLogin>[0]) {
	const s = rand();
	const token = await apiLogin(request, `share${s}`);
	const character = await createCharacter(request, token, {
		nick: `Bra${s}`,
		classId: 'arcebispo',
		level: 178,
		role: 'support'
	});
	return createLobby(request, token, {
		instanceId: 'templo-do-demonio-rei',
		day: 4,
		time: '20:00',
		characterId: character.id,
		minLevel: 160
	});
}

test.describe('compartilhar lobby', () => {
	test('CA-01.3: "Compartilhar" copia o convite de três linhas e avisa', async ({
		page,
		context,
		request,
		baseURL
	}) => {
		const lobby = await seedLobby(request);
		await context.grantPermissions(['clipboard-read', 'clipboard-write']);
		await page.goto(`/lobbies/${lobby.id}`);
		await page.getByRole('button', { name: 'Compartilhar' }).click();
		await expect(page.getByRole('button', { name: 'Convite copiado' })).toBeVisible();

		const [y, m, d] = spDay(4).split('-').map(Number);
		const weekday = WEEKDAYS[new Date(Date.UTC(y, m - 1, d)).getUTCDay()];
		const date = `${String(d).padStart(2, '0')}/${String(m).padStart(2, '0')}`;
		// O anfitrião ocupa uma vaga de Suporte; ficam 1 Tank, 1 Suporte e 3 Dano.
		// A área de transferência do Windows devolve as quebras de linha como CRLF.
		const copied = await page.evaluate(() => navigator.clipboard.readText());
		expect(copied.replace(/\r\n/g, '\n')).toBe(
			[
				`Grupo para Templo do Demônio Rei · ${weekday}, ${date} às 20:00 (horário de Brasília)`,
				'Vagas: 1 Tank, 1 Suporte, 3 Dano · Nível mínimo 160',
				`Candidate-se: ${baseURL}/lobbies/${lobby.id}`
			].join('\n')
		);
	});

	test('CA-01.5: sem permissão de copiar, abre o diálogo com o convite selecionado', async ({
		page,
		request
	}) => {
		const lobby = await seedLobby(request);
		await page.addInitScript(() => {
			navigator.clipboard.writeText = () => Promise.reject(new DOMException('', 'NotAllowedError'));
		});
		await page.goto(`/lobbies/${lobby.id}`);
		await page.getByRole('button', { name: 'Compartilhar' }).click();

		const dialog = page.getByTestId('share-dialog');
		await expect(dialog).toBeVisible();
		const invite = dialog.getByRole('textbox');
		await expect(invite).toHaveValue(/^Grupo para Templo do Demônio Rei · /);
		await expect(invite).toBeFocused();
		const selected = await invite.evaluate(
			(el: HTMLTextAreaElement) => el.selectionEnd - el.selectionStart === el.value.length
		);
		expect(selected).toBe(true);
		await expect(page.getByRole('button', { name: 'Convite copiado' })).toHaveCount(0);
	});
});
