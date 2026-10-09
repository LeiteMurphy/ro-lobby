# Relatório de validação — grupo-livre (ciclo 2)

**Veredito geral:** REPROVADO
**Execução:** backend testes passou (todos os pacotes, `-tags=integration`, inclusive `TestMigration00008_RNF01_ExistingLobby`) · golangci-lint ok (0 issues) · web vitest passou (343/343, 30 arquivos) · web lint ok (prettier + eslint) · svelte-check ok (0 erros, 0 avisos) · build ok · e2e passou (52/52)

Motivo da reprovação: as quatro pendências do ciclo 1 foram resolvidas, mas a varredura por pontos onde o grupo livre ainda é tratado por vaga de função achou mais um na tela do dono, ligado à RN-05 (P1, US-02). O `PlayerPanel` abre por padrão no anfitrião e diz ao dono de um grupo livre que a troca exige "vaga na função". A RN-05 diz o contrário: a troca não depende de vaga. Há também dois textos menores com o mesmo problema (pedido de troca e banco de talentos).

## Pendências do ciclo 1
| # | Pendência | Situação | Evidência |
|---|---|---|---|
| 1 | `SwapPanel` falava de vaga de função no livre | Resolvida | `SwapPanel.svelte:20-21,54,76` (`isFree` → `sameRole`; rótulo "Vaga"; aviso escondido). Teste `free.spec.ts` "CA-02.4 / RN-05: no grupo livre cheio, o painel da troca não fala de vaga de função" (confere também que o lobby por função continua com "Vaga de Tank") |
| 2 | `LeaveDialog` "Sua vaga de {função}" no livre | Resolvida | `LeaveDialog.svelte:53`; `routes/lobbies/[id]/+page.svelte:606` passa `null` no livre. Teste `free.spec.ts` "CA-02.1 / RN-04: sair do grupo livre libera \"Sua vaga\", sem função" |
| 3 | Teste da migração 00008 sobre lobby antigo | Resolvida | `backend/internal/migrate/formation_integration_test.go` `TestMigration00008_RNF01_ExistingLobby`: sobe até a 7, grava um lobby, aplica a 8 (`formation='roles'`, `free_slots` nulo), reverte e confere o CHECK antigo. Passou |
| 4 | Testes sem ID de CA | Quase toda resolvida | Renomeados `TestHasRoom_CA02_1_CA02_2`, `TestCount_CA05_2_Free`, `TestCountTalents_CA05_2_Free` e os `it` de `free.spec.ts`, `home/free.spec.ts`, `share.spec.ts`. Restam 2 (ver RNF-04) |

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 11 | 2 | 0 | 0 |
| Critérios (CA) | 15 | 0 | 0 | 0 |
| Casos de borda | 5 | 0 | 0 | 0 |
| Não funcionais | 3 | 1 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `migrations/00008_lobby_formation.sql` (DEFAULT 'roles', CHECK); `lobbies.go:579-590` `checkFormation` (vazio → roles); web `novo/+page.server.ts` | `TestCheckFormation_CA01_2`, `TestCreate_CA01_1_Free`, `TestLobbies_RNF01_DefaultFormation`, `TestMigration00008_RNF01_ExistingLobby`, `free.spec.ts` "CA-01.3" | |
| RN-02 | ✅ | `checkFormation` 2..12; CHECK no banco; dono ocupa vaga (`applications.go:591` +1) | `TestCreate_CA01_2_FreeLimits`, `TestCreate_CA01_1_Free` | |
| RN-03 | ✅ | `applications.go:589-596` `noRoom` (livre ignora função); `lobbies.go:569` `HasRoom`; web `seats.ts:35` `roomFor` | `TestFree_CA02_1_AnyRole`, `TestHasRoom_CA02_1_CA02_2`, `free.spec.ts` "CA-02.1 / CA-02.2" | |
| RN-04 | ✅ | `noRoom` em Apply (l.203) e Accept (l.241) → `group_full`; `LockLobby`; `LeaveDialog` sem função | `TestFree_CA02_2_Full`, `TestFree_CA02_3_Concurrent`, `free.spec.ts` "CA-02.1 / RN-04: sair…" | |
| RN-05 | ⚠️ | API: `swaps.go:328` pula a vaga no livre (troca do dono l.73 e do membro l.194 passam por `checkSwap`); web: `detail.ts:127` `swapEligibility` → "mesma vaga"; `SwapPanel` corrigido | `TestFree_CA02_4_SwapWithoutRoleSlot`; `free.spec.ts` "CA-02.4 / RN-05" (dois testes) | **Texto errado na tela do dono**: `web/src/lib/applications/components/PlayerPanel.svelte:107-110` mostra, sempre que o dono vê o próprio cartão (o painel abre no anfitrião por padrão), "Troque sem aprovação, para um personagem seu com o nível mínimo e vaga na função." O `PlayerPanel` não recebe o lobby nem a formação, então mostra o mesmo texto no grupo livre, onde a troca do dono (RN-19) não depende de vaga. Menor: `SwapDialog.svelte:125-126` (modo `request`) diz "O pedido pode esperar a vaga abrir.", e no livre o pedido nunca espera vaga. |
| RN-06 | ✅ | `lobbies.go:420-423` `in.FreeSlots < occupied.total()` → `freeSlots/below_occupied` | `TestFree_CA04_1_BelowOccupied` | |
| RN-07 | ✅ | `lobbies.go:404-416` `formation/locked` com aceito ou pendente, na transação; ida a "Por função" passa pelo laço `below_occupied` | `TestUpdate_CA04_2_SwitchFormationAlone`, `TestUpdate_CA04_3_FormationLocked` | Ver observação sobre o erro da ida a "Por função" sem vaga para o dono. |
| RN-08 | ✅ | `LobbyCard.svelte:28,71-74` (`edge='free'`, `FreeComposition`); `FeaturedLobby.svelte:38-41`; `toHome.ts` `free` | `home/free.spec.ts` "CA-03.1"; e2e `grupo-livre.spec.ts` | Ver observação sobre a legenda da Home. |
| RN-09 | ✅ | `home/filters.ts:58,71` com `roomFor` (`home/lobbies.ts:54-56`) | `home/free.spec.ts` "CA-03.2 / RN-09" | |
| RN-10 | ✅ | `routes/lobbies/[id]/+page.svelte:330-380` (`free-places`; blocos por função escondidos no livre); `detail.ts:94` `freeComposition` | `free.spec.ts` "CA-01.1 / RN-10"; `lobbies.spec.ts`; e2e | |
| RN-11 | ✅ | `detail.ts:146-165` `eligibility` (livre: nível e total), `swapEligibility` | `free.spec.ts` "CA-02.5 / RN-11", "CA-02.4 / RN-05"; e2e | |
| RN-12 | ✅ | `share.ts:22-36` `openSlotsLabel` | `share.spec.ts` "CA-05.1" (2 testes) | |
| RN-13 | ⚠️ | `talents/affinity.go` (`HasRoom`); `talents/catalog.go` `Count` (FreeSlots-1); `server/talents.go`; web `disponiveis/+server.ts`, `LobbyForm.svelte` | `TestForLobby_CA05_2_Free`, `TestCount_CA05_2_Free`, `TestCountTalents_CA05_2_Free`, `disponiveis.spec.ts` | Comportamento certo. O texto do painel "Jogadores disponíveis" (`routes/lobbies/[id]/+page.svelte:551-553`) diz "com a instância, o horário, o nível e uma função com vaga" também no grupo livre. Pela RN-13, no livre a condição é "o grupo tem vaga livre". |
| CA-01.1 | ✅ | RN-01/02/10 | `TestCreate_CA01_1_Free`; e2e "1 de 12" e 11 "Vaga aberta" | |
| CA-01.2 | ✅ | `checkFormation` | `TestCreate_CA01_2_FreeLimits`, `TestCheckFormation_CA01_2` | |
| CA-01.3 | ✅ | `novo/+page.server.ts` | `free.spec.ts` "CA-01.3"; e2e `toBeChecked` | |
| CA-02.1 | ✅ | `noRoom` | `TestFree_CA02_1_AnyRole` | |
| CA-02.2 | ✅ | `noRoom` → `group_full`; `applications/messages.ts` | `TestFree_CA02_2_Full`; `free.spec.ts` | |
| CA-02.3 | ✅ | `LockLobby` | `TestFree_CA02_3_Concurrent` | |
| CA-02.4 | ✅ | `checkSwap`; `SwapPanel` | `TestFree_CA02_4_SwapWithoutRoleSlot`; `free.spec.ts` SSR do `SwapPanel` | |
| CA-02.5 | ✅ | `eligibility` | `free.spec.ts` "CA-02.5"; e2e | |
| CA-03.1 | ✅ | `LobbyCard`, `FreeComposition` | `home/free.spec.ts`; e2e | |
| CA-03.2 | ✅ | `filters.ts` | `home/free.spec.ts` "CA-03.2" | Só unitário. |
| CA-04.1 | ✅ | Update `below_occupied` | `TestFree_CA04_1_BelowOccupied` | |
| CA-04.2 | ✅ | Update | `TestUpdate_CA04_2_SwitchFormationAlone` | |
| CA-04.3 | ✅ | Update `locked`; `lobbies/messages.ts` | `TestUpdate_CA04_3_FormationLocked`; `free.spec.ts` "CA-04.3" | |
| CA-05.1 | ✅ | `share.ts` | `share.spec.ts` "CA-05.1" | |
| CA-05.2 | ✅ | `affinity.go` | `TestForLobby_CA05_2_Free` | |
| Borda: grupo de 2 vagas | ✅ | `checkFormation` | `TestFree_CA02_2_Full`, `TestFree_CA02_3_Concurrent` | |
| Borda: sair/remover libera vaga sem função | ✅ | ocupação = soma dos aceitos; `LeaveDialog` sem função | `free.spec.ts` "CA-02.1 / RN-04: sair…" | |
| Borda: lobby por função existente | ✅ | DEFAULT 'roles' | `TestMigration00008_RNF01_ExistingLobby`; suítes antigas verdes | |
| Borda: troca pendente no livre | ✅ | `checkSwap`; `SwapPanel` | `TestFree_CA02_4_…`; `free.spec.ts` SSR | |
| Borda: Minhas candidaturas mostra função | ✅ | sem mudança | suíte existente | |
| RNF-01 | ✅ | migração com DEFAULT e CHECK compatível | `TestMigration00008_RNF01_ExistingLobby`, `TestLobbies_RNF01_DefaultFormation` | |
| RNF-02 | ✅ | regras no serviço Go e CHECK no banco; web só avisa | testes de integração | |
| RNF-03 | ✅ | `LobbyForm.svelte` fieldset "Formação" com rádios rotulados | e2e (foco + ArrowRight) | |
| RNF-04 | ⚠️ | — | — | Ainda sem ID de CA: `backend/internal/db/formation_integration_test.go:33` `TestLobbies_RN01_RN02_FormationChecks` (cobre CA-01.2) e `web/src/routes/lobbies/novo/disponiveis/disponiveis.spec.ts:53` "RN-13 da grupo-livre: …" (cobre CA-05.2). Os testes que citam só RNF-01 estão certos, porque o RNF não tem CA. |

## Não regressão do lobby "Por função"
- Backend: os ramos por função ficam iguais atrás de `Formation != free` (`noRoom` → `role_full`, `checkSwap` → `role_full`, Update com o laço por função). As suítes antigas de lobbies, candidatura, trocas, saída e talentos passam (`TestApply_CA01_7_RoleFull`, `TestAccept_CA02_8_RoleFull`, `TestAccept_CA02_9_Concurrent` etc.).
- Web: `SwapPanel` continua com "Vaga de Tank" no lobby por função (teste SSR em `free.spec.ts`). `LeaveDialog` recebe o rótulo da função fora do livre. `share.ts`, `eligibility` e `swapEligibility` mantêm o ramo por função. As 52 e2e passam, entre elas `lobbies.spec.ts`, `candidatura*.spec.ts`, `home.spec.ts` e `talentos.spec.ts`.

## Pendências para correção
1. [RN-05, P1/US-02] `web/src/lib/applications/components/PlayerPanel.svelte:107-110`: no grupo livre, o texto do botão "Trocar personagem" do anfitrião não pode exigir "vaga na função". Hoje o componente não sabe a formação e mostra "…com o nível mínimo e vaga na função." O texto é o padrão do painel do dono no detalhe. No livre ele deve dizer só o nível mínimo (e, se quiser, que o personagem novo fica com a mesma vaga). Cobrir com teste SSR citando CA-02.4, que também confira que o lobby por função mantém o texto atual.
2. [RN-05] `web/src/lib/applications/components/SwapDialog.svelte:125-126` (modo `request`): "O pedido pode esperar a vaga abrir." não vale no grupo livre, onde a troca não depende de vaga. Ajustar o texto no livre e testar.
3. [RN-13] `web/src/routes/lobbies/[id]/+page.svelte:551-553`: o texto do painel "Jogadores disponíveis" diz "uma função com vaga". No grupo livre a condição é "o grupo tem vaga livre". Ajustar e testar citando CA-05.2.
4. [RNF-04] Citar o CA em `TestLobbies_RN01_RN02_FormationChecks` (`backend/internal/db/formation_integration_test.go:33`) e no `it` de `disponiveis.spec.ts:53`.

## Scope creep
- Nenhum. O commit de correção (b494c6d) só toca os itens das pendências do ciclo 1.

## Observações (não bloqueantes)
- Legenda da Home (`web/src/routes/+page.svelte:135`): "Borda: função com mais vagas" com as três cores não explica a borda neutra do grupo livre (RN-08). Vale incluir uma chave "Grupo livre".
- `routes/lobbies/novo/+page.svelte:30` (sem personagem): "…que ocupa a vaga da função dele." Aparece antes de escolher a formação, então é aceitável, mas no grupo livre não é exato.
- Continua do ciclo 1: a borda usa `var(--fg-2)`, e não o token `--free-line` do design (`design.md:74`).
- Continua do ciclo 1: a ida a "Por função" sem vaga para a função do dono sai como `slots/below_occupied`, e não como `slots/invalid` da RN-08 da `lobbies`. O teste só confere `err != nil`.
- Continua do ciclo 1: `openSlotsLabel` com `formation` opcional; o Down da 00008 apaga os grupos livres (não documentado no design); a troca do dono no livre não tem teste de integração próprio (o caminho `checkSwap` é o mesmo da troca do membro).
- Rastreabilidade ok: b494c6d cita RN-05, RN-04, RNF-01, RNF-04 e CA-02.4. As tasks T-01..T-06 [x] batem com o código.
