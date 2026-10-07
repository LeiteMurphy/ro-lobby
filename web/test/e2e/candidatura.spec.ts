import { expect, test } from '@playwright/test';
import { loginAs, newPage, panel } from './pages';
import {
	acceptApplication,
	apiLogin,
	applyTo,
	cancelLobby,
	createCharacter,
	createLobby,
	rand
} from './seed';

// Ponta a ponta da candidatura (spec candidatura-lobby, T-07), com o Discord falso, o banco
// do ponta a ponta e três contas: dono, candidato e visitante. Os lobbies ficam nos dias 6
// a 8, longe dos da Home (1 e 2) e dos lobbies (3 a 5).

test.describe('candidatura', () => {
	test('CA-01.1 / CA-02.1 / CA-02.2 / CA-03.1 / CA-03.2 / CA-03.7 / CA-10.1 a CA-10.5: candidatar, decidir e acompanhar', async ({
		browser,
		request
	}) => {
		const s = rand();
		// Dono com um Suporte nível 250, lobby nível mínimo 200 com 1 Tank, 2 Suportes e 3 Danos.
		const ownerToken = await apiLogin(request, `dono${s}`);
		const host = await createCharacter(request, ownerToken, {
			nick: `Anf${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support'
		});
		const lobby = await createLobby(request, ownerToken, {
			instanceId: 'templo-do-demonio-rei',
			day: 6,
			time: '20:00',
			characterId: host.id,
			minLevel: 200
		});
		const url = `/lobbies/${lobby.id}`;

		// Candidato com um Tank 250 e um Dano 150 (abaixo do mínimo).
		const biaToken = await apiLogin(request, `bia${s}`);
		await createCharacter(request, biaToken, {
			nick: `Bra${s}`,
			classId: 'guardiao-real',
			level: 250,
			role: 'tank'
		});
		await createCharacter(request, biaToken, {
			nick: `Bai${s}`,
			classId: 'arquimago',
			level: 150,
			role: 'dps'
		});

		// CA-01.1 / CA-01.11: o diálogo desabilita quem não pode e diz por quê.
		const bia = await newPage(browser);
		await loginAs(bia, `bia${s}`, url);
		await bia.getByRole('button', { name: 'Candidatar' }).click();
		const apply = bia.getByTestId('apply-dialog');
		await expect(apply.getByRole('radio', { name: new RegExp(`Bai${s}`) })).toBeDisabled();
		await expect(apply).toContainText('abaixo do nível 200');
		await expect(apply.getByRole('radio', { name: new RegExp(`Bra${s}`) })).toBeChecked();
		await apply.getByLabel('Mensagem para o anfitrião (opcional)').fill('tenho buff de ASPD');
		await apply.getByRole('button', { name: 'Enviar candidatura' }).click();
		await expect(bia.getByTestId('my-application')).toContainText('Sua candidatura está pendente.');
		await expect(bia.getByRole('button', { name: 'Candidatar' })).toHaveCount(0);

		// CA-03.7 / CA-10.3: o visitante vê só a quantidade de pendentes.
		const visitor = await newPage(browser);
		await visitor.goto(url);
		await expect(visitor.getByTestId('pending-count')).toContainText('1 candidatura pendente');
		await expect(visitor.getByTestId('pending-list')).toHaveCount(0);
		await expect(visitor.getByText(`Bra${s}`)).toHaveCount(0);
		// RN-32: o painel abre no anfitrião, sem o Discord dele para o visitante.
		await expect(panel(visitor)).toContainText('Anfitrião');
		await expect(panel(visitor)).toContainText('O Discord aparece para quem está no grupo.');
		await expect(panel(visitor)).not.toContainText(`dono${s}`);

		// CA-10.5: o dono vê o selo na Home (dia 6) e no detalhe.
		const owner = await newPage(browser);
		await loginAs(owner, `dono${s}`, '/');
		await owner.getByRole('tablist', { name: 'Dias' }).getByRole('tab').nth(6).click();
		const card = owner.getByTestId('lobby-card').filter({ hasText: `Anf${s}` });
		await expect(card.getByTestId('pending-badge')).toHaveText('1 pendente');
		await owner.goto(url);
		await expect(owner.getByTestId('pending-badge')).toHaveText('1 pendente');

		// CA-10.3: o dono clica no candidato e vê mensagem, Discord, Aceitar e Recusar.
		const candidate = owner.getByTestId('candidate').filter({ hasText: `Bra${s}` });
		await candidate.click();
		await expect(candidate).toHaveAttribute('aria-pressed', 'true');
		await expect(panel(owner)).toContainText('Candidato');
		await expect(panel(owner)).toContainText('tenho buff de ASPD');
		await expect(panel(owner)).toContainText(`bia${s}`);

		// CA-02.1: aceitar; o membro ocupa a vaga de Tank.
		await panel(owner).getByRole('button', { name: 'Aceitar' }).click();
		const member = owner.getByTestId('member-slot').filter({ hasText: `Bra${s}` });
		await expect(member).toBeVisible();
		await expect(owner.getByTestId('pending-list')).toContainText('Nenhuma candidatura pendente.');
		await expect(owner.getByTestId('pending-badge')).toHaveCount(0);

		// CA-10.1 / CA-10.2: o visitante vê o membro sem Discord; o membro vê o Discord.
		await visitor.reload();
		// RN-31: o card do membro escolhe o painel também pelo teclado.
		const memberCard = visitor.getByTestId('member-slot').filter({ hasText: `Bra${s}` });
		await memberCard.focus();
		await visitor.keyboard.press('Enter');
		await expect(memberCard).toHaveAttribute('aria-pressed', 'true');
		await expect(panel(visitor)).toContainText('Membro');
		await expect(panel(visitor)).toContainText('Guardião Real · Nv 250');
		await expect(panel(visitor)).toContainText('O Discord aparece para quem está no grupo.');
		await expect(panel(visitor)).not.toContainText(`bia${s}`);
		await bia.reload();
		await expect(bia.getByTestId('my-application')).toContainText('Você está no grupo');
		await bia
			.getByTestId('member-slot')
			.filter({ hasText: `Bra${s}` })
			.click();
		await expect(panel(bia)).toContainText(`bia${s}`);

		// CA-01.7: outra conta com Tank vê a função sem vaga.
		const caioToken = await apiLogin(request, `caio${s}`);
		await createCharacter(request, caioToken, {
			nick: `Esc${s}`,
			classId: 'guardiao-real',
			level: 250,
			role: 'tank'
		});
		await createCharacter(request, caioToken, {
			nick: `Fog${s}`,
			classId: 'arquimago',
			level: 220,
			role: 'dps'
		});
		const caio = await newPage(browser);
		await loginAs(caio, `caio${s}`, url);
		await caio.getByRole('button', { name: 'Candidatar' }).click();
		const caioApply = caio.getByTestId('apply-dialog');
		await expect(caioApply.getByRole('radio', { name: new RegExp(`Esc${s}`) })).toBeDisabled();
		await expect(caioApply).toContainText('Tank sem vaga');
		await caioApply.getByRole('radio', { name: new RegExp(`Fog${s}`) }).check({ force: true });
		await caioApply.getByRole('button', { name: 'Enviar candidatura' }).click();
		await expect(caio.getByTestId('my-application')).toContainText('pendente');

		// CA-02.2 / CA-02.3: recusar exige justificativa de 10 a 250.
		await owner.reload();
		await owner
			.getByTestId('candidate')
			.filter({ hasText: `Fog${s}` })
			.click();
		await panel(owner).getByRole('button', { name: 'Recusar' }).click();
		const reject = owner.getByTestId('reject-dialog');
		await reject.getByLabel('Justificativa').fill('curta');
		await reject.getByRole('button', { name: 'Recusar' }).click();
		await expect(reject.getByText('Escreva de 10 a 250 caracteres')).toBeVisible();
		await reject.getByLabel('Justificativa').fill('Precisamos de dano físico');
		await reject.getByRole('button', { name: 'Recusar' }).click();
		await expect(owner.getByTestId('candidate')).toHaveCount(0);

		// CA-10.4 / CA-03.1: "Minhas candidaturas" pelo menu, com a justificativa.
		await caio.getByTestId('user-menu').getByRole('button').click();
		await caio.getByRole('menuitem', { name: 'Minhas candidaturas' }).click();
		await expect(caio).toHaveURL('/candidaturas');
		const row = caio.getByTestId('my-application-row').first();
		await expect(row).toContainText('Templo do Demônio Rei');
		await expect(row).toContainText(`com Fog${s} (Dano)`);
		await expect(row).toContainText('Recusada');
		await expect(row).toContainText('Justificativa: Precisamos de dano físico');

		// CA-03.2: retirar a pendente pela página.
		const dudaToken = await apiLogin(request, `duda${s}`);
		const cura = await createCharacter(request, dudaToken, {
			nick: `Cur${s}`,
			classId: 'arcebispo',
			level: 210,
			role: 'support'
		});
		await applyTo(request, dudaToken, lobby.id, cura.id);
		const duda = await newPage(browser);
		await loginAs(duda, `duda${s}`, '/candidaturas');
		const pending = duda.getByTestId('my-application-row').filter({ hasText: 'Pendente' });
		await pending.getByRole('button', { name: 'Retirar' }).click();
		await expect(duda.getByTestId('my-application-row').first()).toContainText('Retirada');

		// CA-09.1: o personagem aceito não muda de nível no perfil.
		await bia.goto('/perfil');
		await bia.getByRole('button', { name: `Ações de Bra${s}` }).click();
		await bia.getByRole('menuitem', { name: 'Editar' }).click();
		const edit = bia.getByTestId('character-dialog');
		await edit.getByLabel('Nível').fill('251');
		await edit.getByRole('button', { name: 'Salvar personagem' }).click();
		await expect(edit.getByRole('alert')).toHaveText(
			'Esse personagem está num lobby aberto. Saia ou cancele antes de mudar nível ou função.'
		);
	});

	test('CA-02.10 / CA-04.2: conflito de horário no aceite e expiração pelo cancelamento', async ({
		browser,
		request
	}) => {
		const s = rand();
		const ana = await apiLogin(request, `ana${s}`);
		const anaChar = await createCharacter(request, ana, {
			nick: `Ana${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support'
		});
		const leo = await apiLogin(request, `leo${s}`);
		const leoChar = await createCharacter(request, leo, {
			nick: `Leo${s}`,
			classId: 'arcebispo',
			level: 250,
			role: 'support'
		});
		const first = await createLobby(request, ana, {
			instanceId: 'templo-do-demonio-rei',
			day: 7,
			time: '20:00',
			characterId: anaChar.id,
			minLevel: 160
		});
		const second = await createLobby(request, leo, {
			instanceId: 'templo-do-demonio-rei',
			day: 7,
			time: '21:30',
			characterId: leoChar.id,
			minLevel: 160
		});
		const bia = await apiLogin(request, `bia${s}`);
		const fogo = await createCharacter(request, bia, {
			nick: `Fog${s}`,
			classId: 'arquimago',
			level: 250,
			role: 'dps'
		});
		const accepted = await applyTo(request, bia, first.id, fogo.id);
		await acceptApplication(request, ana, accepted.id);
		await applyTo(request, bia, second.id, fogo.id);

		// CA-02.10: o dono do segundo tenta aceitar e vê o conflito; continua pendente.
		const leoPage = await newPage(browser);
		await loginAs(leoPage, `leo${s}`, `/lobbies/${second.id}`);
		await leoPage
			.getByTestId('candidate')
			.filter({ hasText: `Fog${s}` })
			.click();
		await panel(leoPage).getByRole('button', { name: 'Aceitar' }).click();
		await expect(leoPage.getByRole('alert')).toHaveText(
			'O personagem já está em outro grupo a menos de 2 h deste horário.'
		);
		await expect(leoPage.getByTestId('candidate')).toHaveCount(1);

		// CA-04.2: o segundo é cancelado; a pendente expira em "Minhas candidaturas".
		await cancelLobby(request, leo, second.id, 'Imprevisto no trabalho');
		const biaPage = await newPage(browser);
		await loginAs(biaPage, `bia${s}`, '/candidaturas');
		const rows = biaPage.getByTestId('my-application-row');
		await expect(rows).toHaveCount(2);
		await expect(rows.filter({ hasText: '21:30' })).toContainText('Expirada');
		await expect(rows.filter({ hasText: '20:00' })).toContainText('Aceita');
	});
});
