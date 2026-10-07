import { expect, test, type Page } from '@playwright/test';
import { loginAs, newPage, panel } from './pages';
import { acceptApplication, apiLogin, applyTo, createCharacter, createLobby, rand } from './seed';

// Ponta a ponta da Parte 2 da candidatura (spec candidatura-lobby, T-16): sair do grupo,
// remover com bloqueio, troca do dono e pedido de troca do membro, com o Discord falso e o
// banco do ponta a ponta. Os lobbies ficam nos dias 9 e 10, longe dos outros testes.

const notice = (page: Page) => page.getByTestId('my-application');

test.describe('candidatura, Parte 2', () => {
	test('CA-08.14 / CA-08.15 / CA-07.1: o membro pede troca, o dono aceita; o membro retira outro pedido; o dono troca o dele', async ({
		browser,
		request
	}) => {
		const s = rand();
		// Dono com um Suporte e um Dano; lobby com nível mínimo 200, 1 Tank, 2 Suportes e 3 Danos.
		const ownerToken = await apiLogin(request, `dono${s}`);
		const host = await createCharacter(request, ownerToken, {
			nick: `Anf${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support'
		});
		await createCharacter(request, ownerToken, {
			nick: `Lam${s}`,
			classId: 'sicario',
			level: 250,
			role: 'dps'
		});
		const lobby = await createLobby(request, ownerToken, {
			instanceId: 'templo-do-demonio-rei',
			day: 9,
			time: '20:00',
			characterId: host.id,
			minLevel: 200
		});
		const url = `/lobbies/${lobby.id}`;

		// Bia entra como Dano (Fogo) e tem também um Tank (Brisa) e um Dano abaixo do mínimo.
		const biaToken = await apiLogin(request, `bia${s}`);
		const fogo = await createCharacter(request, biaToken, {
			nick: `Fog${s}`,
			classId: 'arquimago',
			level: 250,
			role: 'dps'
		});
		await createCharacter(request, biaToken, {
			nick: `Bri${s}`,
			classId: 'guardiao-real',
			level: 246,
			role: 'tank'
		});
		await createCharacter(request, biaToken, {
			nick: `Bai${s}`,
			classId: 'arquimago',
			level: 150,
			role: 'dps'
		});
		const app = await applyTo(request, biaToken, lobby.id, fogo.id);
		await acceptApplication(request, ownerToken, app.id);

		// CA-08.14: Bia pede a troca para Brisa, com motivo; o de nível baixo fica de fora.
		const bia = await newPage(browser);
		await loginAs(bia, `bia${s}`, url);
		await expect(notice(bia)).toContainText(`Você está no grupo com Fog${s} (Dano)`);
		await notice(bia).getByRole('button', { name: 'Pedir troca' }).click();
		const ask = bia.getByTestId('swap-dialog');
		await expect(ask.getByRole('radio', { name: new RegExp(`Bai${s}`) })).toBeDisabled();
		await expect(ask).toContainText('abaixo do nível 200');
		await expect(ask.getByRole('radio', { name: new RegExp(`Bri${s}`) })).toBeChecked();
		await ask.getByRole('button', { name: 'Enviar pedido' }).click();
		await expect(ask.getByText('Escreva a justificativa')).toBeVisible();
		await ask.getByLabel('Motivo').fill('Ninguém apareceu de tank, posso cobrir');
		await ask.getByRole('button', { name: 'Enviar pedido' }).click();
		await expect(notice(bia)).toContainText(`Pedido de troca pendente para Bri${s} (Tank, Nv 246)`);
		await expect(notice(bia)).toContainText(`Você continua com Fog${s}`);

		// O dono vê o pedido no bloco próprio e, ao escolher pelo teclado, atual → novo e motivo.
		const owner = await newPage(browser);
		await loginAs(owner, `dono${s}`, url);
		await expect(owner.getByTestId('pending-badge')).toHaveText('1 troca');
		const request1 = owner.getByTestId('swap-request').filter({ hasText: `Fog${s} → Bri${s}` });
		await request1.focus();
		await owner.keyboard.press('Enter');
		await expect(request1).toHaveAttribute('aria-pressed', 'true');
		const swapPanel = owner.getByTestId('swap-panel');
		await expect(swapPanel).toContainText('Ninguém apareceu de tank, posso cobrir');
		await expect(swapPanel).toContainText(`bia${s}`);

		// CA-08.5: aceitar; Bia passa para a vaga de Tank com Brisa.
		await swapPanel.getByRole('button', { name: 'Aceitar' }).click();
		await expect(owner.getByTestId('member-slot').filter({ hasText: `Bri${s}` })).toBeVisible();
		await expect(owner.getByTestId('member-slot').filter({ hasText: `Fog${s}` })).toHaveCount(0);
		await expect(owner.getByTestId('swap-list')).toContainText('Nenhum pedido de troca.');

		// CA-08.11: Bia pede a volta para Fogo e retira o pedido; continua com Brisa.
		await bia.reload();
		await expect(notice(bia)).toContainText(`Você está no grupo com Bri${s} (Tank)`);
		await notice(bia).getByRole('button', { name: 'Pedir troca' }).click();
		await ask.getByRole('radio', { name: new RegExp(`Fog${s}`) }).check({ force: true });
		await ask.getByLabel('Motivo').fill('Melhor voltar para o dano');
		await ask.getByRole('button', { name: 'Enviar pedido' }).click();
		await expect(notice(bia)).toContainText(`Pedido de troca pendente para Fog${s}`);
		await notice(bia).getByRole('button', { name: 'Retirar pedido' }).click();
		await expect(notice(bia)).toContainText(`Você está no grupo com Bri${s} (Tank)`);
		await expect(notice(bia).getByRole('button', { name: 'Pedir troca' })).toBeVisible();

		// CA-08.15 / RN-39: o dono recusa um novo pedido; Bia vê a justificativa.
		await notice(bia).getByRole('button', { name: 'Pedir troca' }).click();
		await ask.getByRole('radio', { name: new RegExp(`Fog${s}`) }).check({ force: true });
		await ask.getByLabel('Motivo').fill('Melhor voltar para o dano');
		await ask.getByRole('button', { name: 'Enviar pedido' }).click();
		await expect(notice(bia)).toContainText('Pedido de troca pendente');
		await owner.reload();
		await owner.getByTestId('swap-request').first().click();
		await owner.getByTestId('swap-panel').getByRole('button', { name: 'Recusar' }).click();
		const refuse = owner.getByTestId('reject-dialog');
		await refuse.getByLabel('Justificativa').fill('Precisamos de você no tank');
		await refuse.getByRole('button', { name: 'Recusar' }).click();
		await expect(owner.getByTestId('swap-list')).toContainText('Nenhum pedido de troca.');
		await bia.reload();
		await expect(bia.getByTestId('swap-rejected')).toHaveText(
			'Seu pedido de troca foi recusado. Justificativa: Precisamos de você no tank'
		);
		await expect(notice(bia)).toContainText(`Você está no grupo com Bri${s} (Tank)`);

		// CA-07.1: o dono troca o próprio Suporte pelo Dano, sem aprovação.
		await owner.reload();
		await owner.getByTestId('owner-slot').click();
		await panel(owner).getByRole('button', { name: 'Trocar personagem' }).click();
		const own = owner.getByTestId('swap-dialog');
		await expect(own).toContainText('Trocar seu personagem');
		await expect(own.getByRole('radio', { name: new RegExp(`Lam${s}`) })).toBeChecked();
		await own.getByRole('button', { name: 'Trocar' }).click();
		await expect(owner.getByTestId('owner-slot')).toContainText(`Lam${s}`);
		await expect(owner.getByTestId('owner-slot')).toHaveClass(/slot--dps/);
	});

	test('CA-05.5 / CA-05.4 / CA-06.8: sair pelo aviso e por Minhas candidaturas, e remover com bloqueio', async ({
		browser,
		request
	}) => {
		const s = rand();
		const ownerToken = await apiLogin(request, `dono${s}`);
		const host = await createCharacter(request, ownerToken, {
			nick: `Anf${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support'
		});
		const lobby = await createLobby(request, ownerToken, {
			instanceId: 'templo-do-demonio-rei',
			day: 10,
			time: '20:00',
			characterId: host.id,
			minLevel: 160
		});
		const url = `/lobbies/${lobby.id}`;
		const member = async (name: string, nick: string, role: 'tank' | 'support' | 'dps') => {
			const token = await apiLogin(request, `${name}${s}`);
			const c = await createCharacter(request, token, {
				nick: `${nick}${s}`,
				classId: role === 'tank' ? 'guardiao-real' : role === 'support' ? 'arcebispo' : 'arquimago',
				level: 200,
				role
			});
			const app = await applyTo(request, token, lobby.id, c.id);
			await acceptApplication(request, ownerToken, app.id);
		};
		await member('caio', 'Esc', 'tank');
		await member('duda', 'Fog', 'dps');
		await member('bia', 'Cur', 'support');

		// CA-05.5: Caio sai pelo aviso, com confirmação; CA-05.4: e se candidata de novo.
		const caio = await newPage(browser);
		await loginAs(caio, `caio${s}`, url);
		await notice(caio).getByRole('button', { name: 'Sair do grupo' }).click();
		const leave = caio.getByTestId('leave-dialog');
		await expect(leave).toContainText('Sua vaga de Tank fica livre para outro jogador.');
		await leave.getByRole('button', { name: 'Sair do grupo' }).click();
		await expect(notice(caio)).toContainText('Você saiu do grupo.');
		await expect(caio.getByTestId('member-slot').filter({ hasText: `Esc${s}` })).toHaveCount(0);
		await caio.getByRole('button', { name: 'Candidatar' }).click();
		await caio
			.getByTestId('apply-dialog')
			.getByRole('button', { name: 'Enviar candidatura' })
			.click();
		await expect(notice(caio)).toContainText('Sua candidatura está pendente.');

		// RN-38: Duda sai por "Minhas candidaturas"; a linha passa a "Saiu".
		const duda = await newPage(browser);
		await loginAs(duda, `duda${s}`, '/candidaturas');
		const row = duda.getByTestId('my-application-row').first();
		await expect(row).toContainText('Aceita');
		await row.getByRole('button', { name: 'Sair do grupo' }).click();
		await duda.getByTestId('leave-dialog').getByRole('button', { name: 'Sair do grupo' }).click();
		await expect(duda.getByTestId('my-application-row').first()).toContainText('Saiu');

		// CA-06.8: o dono remove Bia com justificativa e bloqueio.
		const owner = await newPage(browser);
		await loginAs(owner, `dono${s}`, url);
		await owner
			.getByTestId('member-slot')
			.filter({ hasText: `Cur${s}` })
			.click();
		await panel(owner).getByRole('button', { name: 'Remover do grupo' }).click();
		const remove = owner.getByTestId('remove-dialog');
		await remove.getByRole('button', { name: 'Remover' }).click();
		await expect(remove.getByText('Escreva a justificativa')).toBeVisible();
		await remove.getByLabel('Justificativa').fill('Mudamos o horário da run, valeu!');
		await remove.getByRole('checkbox', { name: /Bloquear neste lobby/ }).check();
		await remove.getByRole('button', { name: 'Remover' }).click();
		await expect(owner.getByTestId('member-slot').filter({ hasText: `Cur${s}` })).toHaveCount(0);

		// O removido vê o aviso no detalhe, sem Candidatar, e "Removida" com a justificativa.
		const bia = await newPage(browser);
		await loginAs(bia, `bia${s}`, url);
		await expect(notice(bia)).toContainText('Você foi removido deste lobby.');
		await expect(notice(bia)).toContainText('Justificativa: Mudamos o horário da run, valeu!');
		await expect(notice(bia)).toContainText('Você não pode se candidatar a este lobby.');
		await expect(bia.getByRole('button', { name: 'Candidatar' })).toHaveCount(0);
		await bia.goto('/candidaturas');
		const removed = bia.getByTestId('my-application-row').first();
		await expect(removed).toContainText('Removida');
		await expect(removed).toContainText('Justificativa: Mudamos o horário da run, valeu!');
	});
});
