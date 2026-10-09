# Relatório de validação — grupo-livre (ciclo 3)

**Veredito geral:** APROVADO
**Execução:** backend testes passou (todos os pacotes, `-tags=integration`) · golangci-lint ok (0 issues) · web vitest passou (346/346, 30 arquivos) · web lint ok (prettier + eslint) · svelte-check ok (0 erros, 0 avisos, 534 arquivos) · build ok · e2e passou (52/52)

As quatro pendências do ciclo 2 foram resolvidas no commit 14c1ba3, com teste. A varredura
de pontos onde o grupo livre ainda seria tratado por vaga de função (web e backend) não
achou nenhum ponto novo com comportamento ou texto de tela errado. O lobby "Por função"
mantém os ramos e os textos de antes, e as suítes antigas passam.

## Pendências do ciclo 2
| # | Pendência | Situação | Evidência |
|---|---|---|---|
| 1 | [RN-05] `PlayerPanel` do anfitrião pedia "vaga na função" no livre | Resolvida | `PlayerPanel.svelte:18-25,110-114` (prop `free`; no livre: "…com o nível mínimo; a vaga continua sua."); `routes/lobbies/[id]/+page.svelte:536` passa `{free}`. Teste `free.spec.ts` "CA-02.4 / RN-05: a troca do dono no grupo livre não fala de vaga na função" (confere o livre sem "vaga na função" e o por função com o texto antigo) |
| 2 | [RN-05] `SwapDialog` (pedido) "O pedido pode esperar a vaga abrir." no livre | Resolvida | `SwapDialog.svelte:27-28,128-131`; `+page.svelte:615` passa `{free}`. Teste `free.spec.ts` "CA-02.4 / RN-05: o pedido de troca no grupo livre não espera vaga" (livre e por função) |
| 3 | [RN-13] Texto do painel "Jogadores disponíveis" | Resolvida | `+page.svelte:552-555` ("o grupo com vaga livre" no livre, "uma função com vaga" por função). Teste `lobbies.spec.ts` "CA-05.2 / RN-13: o painel do dono fala do grupo com vaga livre, não de função" (renderiza o detalhe nas duas formações) |
| 4 | [RNF-04] Dois testes sem ID de CA | Resolvida | `TestLobbies_CA01_2_RN01_RN02_FormationChecks` (`backend/internal/db/formation_integration_test.go:33`); `disponiveis.spec.ts:53` "CA-05.2 / RN-13 da grupo-livre: …" |

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 13 | 0 | 0 | 0 |
| Critérios (CA) | 15 | 0 | 0 | 0 |
| Casos de borda | 5 | 0 | 0 | 0 |
| Não funcionais | 4 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `migrations/00008_lobby_formation.sql` (DEFAULT 'roles', CHECK); `lobbies.go:579-590` `checkFormation` (vazio → roles); `queries/lobbies.sql:7` COALESCE; web `novo/+page.server.ts` | `TestCheckFormation_CA01_2`, `TestCreate_CA01_1_Free`, `TestLobbies_RNF01_DefaultFormation`, `TestMigration00008_RNF01_ExistingLobby`, `free.spec.ts` "CA-01.3" | |
| RN-02 | ✅ | `checkFormation` 2..12 (`lobbies.go:584`); CHECK no banco; dono ocupa vaga (`applications.go:589-596`) | `TestCreate_CA01_2_FreeLimits`, `TestCreate_CA01_1_Free`, `TestLobbies_CA01_2_RN01_RN02_FormationChecks` | |
| RN-03 | ✅ | `applications.go:589-596` `noRoom` (livre ignora função); `lobbies.go:569-574` `HasRoom`; web `seats.ts:35-38` `roomFor` | `TestFree_CA02_1_AnyRole`, `TestHasRoom_CA02_1_CA02_2`, `free.spec.ts` | |
| RN-04 | ✅ | `noRoom` em Apply (l.203) e Accept (l.241) → `group_full`; trava do lobby; `LeaveDialog` sem função | `TestFree_CA02_2_Full`, `TestFree_CA02_3_Concurrent`, `free.spec.ts` "CA-02.1 / RN-04: sair…" | |
| RN-05 | ✅ | API: `swaps.go:328` pula a vaga no livre; troca do dono (`swaps.go:73`) e do membro passam por `checkSwap`. Web: `detail.ts:127` "mesma vaga"; `SwapPanel.svelte:19-20,54,76`; `PlayerPanel.svelte:110-114`; `SwapDialog.svelte:128-131` | `TestFree_CA02_4_SwapWithoutRoleSlot`; `free.spec.ts` "CA-02.4 / RN-05" (quatro testes: painel da troca, diálogo, painel do dono, pedido) | |
| RN-06 | ✅ | `lobbies.go:420-423` → `freeSlots/below_occupied` | `TestFree_CA04_1_BelowOccupied` | |
| RN-07 | ✅ | `lobbies.go:404-416` `formation/locked` com aceito ou pendente, dentro da transação; ida a "Por função" passa pelo laço por função (l.426) | `TestUpdate_CA04_2_SwitchFormationAlone`, `TestUpdate_CA04_3_FormationLocked` | Ver observação sobre o código do erro. |
| RN-08 | ✅ | `LobbyCard.svelte:28-31,71-74` (borda neutra, `FreeComposition`); `FeaturedLobby.svelte`; `toHome.ts` | `home/free.spec.ts` "CA-03.1"; e2e `grupo-livre.spec.ts` | |
| RN-09 | ✅ | `home/filters.ts:58,71` com `roomFor` (`home/lobbies.ts:54-57`) | `home/free.spec.ts` "CA-03.2 / RN-09" | |
| RN-10 | ✅ | `routes/lobbies/[id]/+page.svelte` (lista única `free-places`; blocos por função só fora do livre); `detail.ts:94-99` `freeComposition` | `free.spec.ts` "CA-01.1 / RN-10"; `lobbies.spec.ts`; e2e | |
| RN-11 | ✅ | `detail.ts:146-165` `eligibility`, `detail.ts:113-143` `swapEligibility` | `free.spec.ts` "CA-02.5 / RN-11", "CA-02.4 / RN-05"; e2e | |
| RN-12 | ✅ | `share.ts:22-36` `openSlotsLabel` (usado por `inviteText` e `shareMeta`) | `share.spec.ts` "CA-05.1" | |
| RN-13 | ✅ | `talents/affinity.go:80` (`HasRoom`); `talents/catalog.go:127-128` (FreeSlots−1); `server/talents.go:99`; web `disponiveis/+server.ts`, `LobbyForm.svelte`; texto do painel em `+page.svelte:552-555` | `TestForLobby_CA05_2_Free`, `TestCount_CA05_2_Free`, `TestCountTalents_CA05_2_Free`, `disponiveis.spec.ts` "CA-05.2", `lobbies.spec.ts` "CA-05.2 / RN-13" | |
| CA-01.1 | ✅ | RN-01/02/10 | `TestCreate_CA01_1_Free`; e2e "1 de 12" e 11 "Vaga aberta" | |
| CA-01.2 | ✅ | `checkFormation`; CHECK | `TestCreate_CA01_2_FreeLimits`, `TestCheckFormation_CA01_2`, `TestLobbies_CA01_2_…` | |
| CA-01.3 | ✅ | `novo/+page.server.ts` | `free.spec.ts` "CA-01.3"; e2e | |
| CA-02.1 | ✅ | `noRoom` | `TestFree_CA02_1_AnyRole` | |
| CA-02.2 | ✅ | `noRoom` → `group_full`; `applications/messages.ts:12` | `TestFree_CA02_2_Full`; `free.spec.ts` | |
| CA-02.3 | ✅ | trava do lobby no aceite | `TestFree_CA02_3_Concurrent` | |
| CA-02.4 | ✅ | `checkSwap`; textos de troca | `TestFree_CA02_4_SwapWithoutRoleSlot`; `free.spec.ts` | |
| CA-02.5 | ✅ | `eligibility` | `free.spec.ts` "CA-02.5"; e2e | |
| CA-03.1 | ✅ | `LobbyCard`, `FreeComposition` | `home/free.spec.ts`; e2e | |
| CA-03.2 | ✅ | `filters.ts` | `home/free.spec.ts` "CA-03.2" | Só unitário. |
| CA-04.1 | ✅ | Update `below_occupied` | `TestFree_CA04_1_BelowOccupied` | |
| CA-04.2 | ✅ | Update | `TestUpdate_CA04_2_SwitchFormationAlone` | |
| CA-04.3 | ✅ | Update `locked`; `lobbies/messages.ts` | `TestUpdate_CA04_3_FormationLocked`; `free.spec.ts` "CA-04.3" | |
| CA-05.1 | ✅ | `share.ts` | `share.spec.ts` "CA-05.1" | |
| CA-05.2 | ✅ | `affinity.go`, texto do painel | `TestForLobby_CA05_2_Free`; `lobbies.spec.ts` "CA-05.2" | |
| Borda: grupo de 2 vagas | ✅ | `checkFormation` | `TestFree_CA02_2_Full`, `TestFree_CA02_3_Concurrent` | |
| Borda: sair/remover libera vaga sem função | ✅ | ocupação = soma dos aceitos; `LeaveDialog` sem função | `free.spec.ts` "CA-02.1 / RN-04: sair…" | |
| Borda: lobby por função existente | ✅ | DEFAULT 'roles' | `TestMigration00008_RNF01_ExistingLobby`; suítes antigas verdes | |
| Borda: troca pendente no livre | ✅ | `checkSwap`; `SwapPanel` | `TestFree_CA02_4_…`; `free.spec.ts` | |
| Borda: Minhas candidaturas mostra função | ✅ | sem mudança (`routes/candidaturas` não fala de vaga) | suíte existente | |
| RNF-01 | ✅ | migração com DEFAULT e CHECK compatível | `TestMigration00008_RNF01_ExistingLobby`, `TestLobbies_RNF01_DefaultFormation` | |
| RNF-02 | ✅ | regras no serviço Go e CHECK no banco; web só avisa | testes de integração | |
| RNF-03 | ✅ | `LobbyForm.svelte` fieldset "Formação" com rádios rotulados | e2e (foco + ArrowRight) | |
| RNF-04 | ✅ | — | todos os testes da feature citam CA; os que citam só RNF-01 cobrem um RNF sem CA | |

## Varredura: grupo livre tratado por vaga de função
Procurei no web (`slots[`, `occupied`, `openSlots`, `headcount`, "vaga de", "vaga na função", "função com vaga", "esperar a vaga") e no backend (`role_full`, `Slots.of`, `slots_*`, `freeSlots(`, `HasRoom`, `noRoom`, queries SQL). Todo uso por função está atrás de um teste de formação:
- web: `seats.ts` (`roomFor`, `totalSeats`), `detail.ts` (`composition` só no por função; `eligibility` e `swapEligibility` com `isFree` antes), `SwapPanel` (`sameRole = freeGroup || …`, aviso escondido), `share.ts` (`isFree` antes), `home/lobbies.ts` (`roomFor`, `seats`, `lobbyIsFull` com `free`), `LobbyCard`/`FeaturedLobby` (`FreeComposition`), `PlayerPanel`, `SwapDialog`, `LeaveDialog`, painel de talentos, `editar/+page.server.ts` (valores padrão dos steppers no livre, só para a troca de formação).
- backend: `noRoom` (Apply e Accept), `checkSwap` (troca do dono e do membro), Update (laço por função só fora do livre), `HasRoom` (talentos), `catalog.Count`. Não há query SQL que conte vaga por função fora desses pontos.
Nenhum ponto novo.

## Não regressão do lobby "Por função"
- Backend: os ramos por função ficam iguais atrás de `Formation != free` (`noRoom` → `role_full`, `checkSwap` → `role_full`, Update com o laço por função, criação com a RN-08 da `lobbies` em `lobbies.go:312`). As suítes antigas de lobbies, candidatura, trocas, saída e talentos passam.
- Web: os testes novos conferem os dois lados (texto antigo no por função): `PlayerPanel` ("vaga na função"), `SwapDialog` ("O pedido pode esperar a vaga abrir."), painel de talentos ("uma função com vaga"), `SwapPanel` ("Vaga de Tank"). As 52 e2e passam, entre elas `lobbies.spec.ts`, `candidatura*.spec.ts`, `home.spec.ts` e `talentos.spec.ts`.

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. O 14c1ba3 toca só as pendências do ciclo 2 e o texto de `routes/lobbies/novo/+page.svelte:30` ("ocupa uma das vagas"), que era observação do ciclo 2 ligada à RN-02.

## Observações (não bloqueantes)
- `openapi.yaml:397` (troca do dono) e `:844` (aceite da troca) ainda descrevem a regra só por função ("com vaga na função…"), sem a exceção do grupo livre (RN-05); `:441` diz que o personagem do dono "ocupa a vaga da função dele". O comportamento está certo; só a documentação do contrato ficou atrás.
- Continua do ciclo 2: a legenda da Home (`web/src/routes/+page.svelte:135`, "Borda: função com mais vagas") não explica a borda neutra do grupo livre.
- Continua: a borda usa `var(--fg-2)`, e não o token `--free-line` do design (`design.md`, seção 3).
- Continua: a ida a "Por função" sem vaga para a função do dono sai como `slots/below_occupied`, e não `slots/invalid` da RN-08 da `lobbies`; o teste só confere `err != nil`.
- Continua: `openSlotsLabel` com `formation` opcional; o Down da 00008 apaga os grupos livres (não documentado no design); a troca do dono no livre não tem teste de integração próprio (passa pelo mesmo `checkSwap` da troca do membro).
- Rastreabilidade ok: o 14c1ba3 cita RN-05, RN-13, RNF-04, CA-02.4 e CA-05.2. As tasks T-01..T-06 [x] batem com o código.
