# Relatório de validação — grupo-livre (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** backend testes passou (todos os pacotes, `-tags=integration`) · golangci-lint ok (0 issues) · web vitest passou (341/341, 30 arquivos) · web lint ok · svelte-check ok (0 erros, 0 avisos) · build ok · e2e passou (52/52)

Motivo da reprovação: no painel do pedido de troca (`SwapPanel.svelte`), o grupo livre ainda é tratado por vaga de função. O dono vê "Vaga de Tank: 0 livres" e o aviso "o aceite falha", o que contradiz a RN-05 (P1, US-02) e o caso de borda "aceitar não depende de vaga".

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 12 | 1 | 0 | 0 |
| Critérios (CA) | 15 | 0 | 0 | 0 |
| Casos de borda | 3 | 1 | 1 | 0 |
| Não funcionais | 2 | 2 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `migrations/00008_lobby_formation.sql` (DEFAULT 'roles', CHECK); `lobbies.go` `checkFormation` (vazio → roles); `queries/lobbies.sql` COALESCE; web `novo/+page.server.ts` formation 'roles' | `TestCheckFormation_CA01_2`, `TestCreate_CA01_1_Free` (RNF-01), `TestLobbies_RN01_RN02_FormationChecks`, `free.spec.ts` "CA-01.3 / RN-01" | |
| RN-02 | ✅ | `checkFormation` (2..12, slots zerados → `freeSlots/invalid`); CHECK no banco; dono ocupa vaga (`noRoom` +1, `toLobby` soma o dono) | `TestCreate_CA01_2_FreeLimits`, `TestLobbies_RN01_RN02_FormationChecks`, `TestCreate_CA01_1_Free` | |
| RN-03 | ✅ | `applications.go` `noRoom` (livre ignora a função); `lobbies.HasRoom` | `TestFree_CA02_1_AnyRole`, `TestHasRoom_RN03_RN04` | |
| RN-04 | ✅ | `noRoom` em `Apply` (l.203) e `Accept` (l.241) → `group_full`; trava `LockLobby` no aceite | `TestFree_CA02_2_Full`, `TestFree_CA02_3_Concurrent` | |
| RN-05 | ⚠️ | API: `swaps.go` `checkSwap` l.327-331 pula a vaga no livre; web: `swapEligibility` retorna "mesma vaga" | `TestFree_CA02_4_SwapWithoutRoleSlot`; `free.spec.ts` "RN-05" | **Web errado no painel de decisão do dono**: `web/src/lib/applications/components/SwapPanel.svelte:19` calcula `free = lobby.slots[to.role] - lobby.occupied[to.role]`. No livre, `slots` vem zerado, então `free ≤ 0`, e o painel mostra "Vaga de Tank: 0 livres" (l.51-52) e o aviso "Se a vaga de Tank não abrir, o aceite falha e o pedido continua pendente." (l.73-76). Troca do dono (RN-19) no livre sem teste de integração próprio (passa pelo mesmo `checkSwap`). |
| RN-06 | ✅ | `lobbies.go` Update: `in.FreeSlots < occupied.total()` → `freeSlots/below_occupied` | `TestFree_CA04_1_BelowOccupied` | |
| RN-07 | ✅ | `lobbies.go` Update: troca com aceito ou pendente → `formation/locked`, dentro da transação com `GetOwnLobbyForUpdate`; ida para "Por função" respeita a vaga do dono pelo laço `below_occupied` | `TestUpdate_CA04_2_SwitchFormationAlone`, `TestUpdate_CA04_3_FormationLocked` | O teste da ida a "Por função" sem vaga de Suporte só confere `err != nil`, sem o campo/código. |
| RN-08 | ✅ | `LobbyCard.svelte` (selo via `FreeComposition`, `data-edge="free"`, borda `var(--fg-2)`); `FeaturedLobby.svelte`; `toHome.ts` `free` | `home/free.spec.ts` "CA-03.1 / RN-08" (card e destaque); e2e `grupo-livre.spec.ts` | O design previa o token `--free-line`; foi usado `--fg-2` (neutro, mas não uma cor "própria"). |
| RN-09 | ✅ | `home/filters.ts` `matches` e `roleCounts` com `roomFor` | `home/free.spec.ts` "CA-03.2 / RN-09" | |
| RN-10 | ✅ | `routes/lobbies/[id]/+page.svelte` bloco `free-places`; `detail.ts` `freeComposition` | `lobbies.spec.ts` "CA-01.1 / RN-10"; `free.spec.ts` "RN-10"; e2e | |
| RN-11 | ✅ | `detail.ts` `eligibility` (livre: só nível/total) e `swapEligibility` (livre: "mesma vaga") | `free.spec.ts` "CA-02.5 / RN-11", "RN-05"; e2e "11 vagas livres" | Conflito de horário segue só na API, como antes. |
| RN-12 | ✅ | `share.ts` `openSlotsLabel` (livre) | `share.spec.ts` "CA-05.1 / RN-12" e "RN-12: 1 livre / Grupo lotado" | |
| RN-13 | ✅ | `talents/affinity.go` (`HasRoom`); `talents/catalog.go` `Count` (FreeSlots-1); `server/talents.go`; web `disponiveis/+server.ts`, `LobbyForm.svelte` | `TestForLobby_CA05_2_Free`, `TestCount_RN13_Free`, `TestCountTalents_RN13_Free`, `disponiveis.spec.ts` "RN-13" | |
| CA-01.1 | ✅ | ver RN-01/RN-02/RN-10 | `TestCreate_CA01_1_Free`; e2e "1 de 12" e 11 "Vaga aberta"; `lobbies.spec.ts` | |
| CA-01.2 | ✅ | `checkFormation` | `TestCreate_CA01_2_FreeLimits`, `TestCheckFormation_CA01_2` | |
| CA-01.3 | ✅ | `novo/+page.server.ts` (roles, 1/2/3) | `free.spec.ts` "CA-01.3 / RNF-03"; e2e `toBeChecked` | |
| CA-02.1 | ✅ | `noRoom` | `TestFree_CA02_1_AnyRole` (3 Suportes, sem vaga restante) | |
| CA-02.2 | ✅ | `noRoom` → `group_full`; `applications/messages.ts` | `TestFree_CA02_2_Full`; `free.spec.ts` `ruleMessage('group_full')` | |
| CA-02.3 | ✅ | `lockApplication` + `LockLobby` | `TestFree_CA02_3_Concurrent` (6 aceites, 1 passa) | |
| CA-02.4 | ✅ | `checkSwap` | `TestFree_CA02_4_SwapWithoutRoleSlot` | Só a API. No web, o painel do dono mostra o aviso errado (ver RN-05). |
| CA-02.5 | ✅ | `eligibility` | `free.spec.ts` "CA-02.5 / RN-11"; e2e | |
| CA-03.1 | ✅ | `LobbyCard`, `FreeComposition` | `home/free.spec.ts`; e2e "Grupo livre" / "2 de 12" | |
| CA-03.2 | ✅ | `filters.ts` | `home/free.spec.ts` "CA-03.2" | Só unitário. |
| CA-04.1 | ✅ | Update `below_occupied` | `TestFree_CA04_1_BelowOccupied` | |
| CA-04.2 | ✅ | Update | `TestUpdate_CA04_2_SwitchFormationAlone` (dono numa vaga de Suporte) | |
| CA-04.3 | ✅ | Update `locked`; `lobbies/messages.ts` | `TestUpdate_CA04_3_FormationLocked` (pendente e aceito); `free.spec.ts` "CA-04.3" | |
| CA-05.1 | ✅ | `share.ts` | `share.spec.ts` "CA-05.1" | |
| CA-05.2 | ✅ | `affinity.go` | `TestForLobby_CA05_2_Free` (Tank, Suporte e Dano; cheio → nenhum) | |
| Borda: grupo de 2 vagas | ✅ | `checkFormation` | `TestFree_CA02_2_Full`, `TestFree_CA02_3_Concurrent` (2 vagas) | |
| Borda: sair/remover libera vaga sem função | ⚠️ | API: ocupação = soma dos aceitos | sem teste específico no livre | `LeaveDialog.svelte:53` diz "Sua vaga de {Dano} fica livre", ou seja, vaga de função no grupo livre. |
| Borda: lobby por função existente | ✅ | DEFAULT 'roles' + CHECK compatível | `TestLobbies_RNF01_DefaultFormation`; suítes antigas verdes | |
| Borda: troca pendente no livre, aceitar não depende de vaga | ❌ | API ok (`checkSwap`) | `TestFree_CA02_4_...` | No web, `SwapPanel.svelte:19,51-52,73-76` informa "0 livres" e "o aceite falha", o oposto do especificado. |
| Borda: Minhas candidaturas mostra função | ✅ | sem mudança, mostra a função | suíte existente | |
| RNF-01 | ⚠️ | migração com DEFAULT e CHECK compatível | `TestLobbies_RNF01_DefaultFormation` (só pela query, com COALESCE) | Falta o teste que T-01 ("Pronto quando") e o design (Riscos) prometem: aplicar a 00008 sobre um lobby antigo já gravado e reverter. |
| RNF-02 | ✅ | regras no serviço Go e CHECK no banco; web só avisa | testes de integração | |
| RNF-03 | ✅ | `LobbyForm.svelte` fieldset "Formação" com rádios rotulados | e2e (foco + ArrowRight); `free.spec.ts` | |
| RNF-04 | ⚠️ | — | — | Alguns testes citam só RN, sem CA: `TestHasRoom_RN03_RN04`, `TestCount_RN13_Free`, `TestCountTalents_RN13_Free`, `free.spec.ts` "RN-05" e "RN-10", `share.spec.ts` "RN-12: uma vaga…". |

## Pendências para correção
1. [RN-05, caso de borda "troca pendente no livre"; P1/US-02] `web/src/lib/applications/components/SwapPanel.svelte`: no grupo livre, o painel do pedido de troca não pode falar de vaga de função. Hoje `free` sai de `lobby.slots[to.role] - lobby.occupied[to.role]` (l.19). Com `slots` zerado, o dono vê "Vaga de Tank: 0 livres" (l.51-52) e "Se a vaga de Tank não abrir, o aceite falha…" (l.73-76). O correto é mostrar que o personagem novo fica com a mesma vaga, sem o aviso (usar `isFree` de `seats.ts`), e cobrir com teste SSR citando CA-02.4.
2. [RN-04, caso de borda "a vaga volta a ser livre, sem função"] `web/src/lib/applications/components/LeaveDialog.svelte:53`: no grupo livre, o texto "Sua vaga de {função} fica livre" trata a vaga como sendo de uma função. Ajustar o texto e testar.
3. [RNF-01] Falta o teste da migração 00008 com um lobby já gravado: subir até a 00007, inserir um lobby, aplicar a 00008, conferir `formation='roles'` e `free_slots` nulo, depois reverter. Foi prometido em T-01 e nos riscos do design.
4. [RNF-04] Os testes listados acima precisam citar o ID do critério de aceite, e não só a RN.

## Scope creep
- Nenhum. Todas as alterações do diff se ligam a IDs da spec ou do design.

## Observações (não bloqueantes)
- RN-08: o design previa o token `--free-line`, mas a borda usa `var(--fg-2)`. Vale criar o token próprio ou atualizar o design.
- RN-07: o teste da volta a "Por função" sem vaga para a função do dono só confere que houve erro. Vale conferir `slots/below_occupied`. O código não usa o `slots/invalid` da RN-08 da `lobbies`: o erro sai pela regra dos ocupantes. Funciona, mas a mensagem é outra.
- Migração Down: `DELETE FROM lobbies WHERE formation = 'free'` apaga os grupos livres (e as candidaturas em cascata) ao reverter. É aceitável, mas não está documentado no design.
- `openSlotsLabel` aceita `formation` opcional. Um chamador que omita a formação cai no ramo por função e mostraria "Grupo lotado" para um livre. Hoje todos os chamadores passam o `ApiLobby` completo.
- A troca do dono (RN-19) no grupo livre não tem teste de integração próprio. Passa pelo mesmo `checkSwap` da troca do membro, que está coberta.
- Rastreabilidade ok: os commits citam T-xx, RN e CA, e as tasks [x] batem com o código (T-01 f545634, T-02 bee97cb/e985f93, T-03 874d1be, T-04 083ac1b, T-05 ef38eac, T-06 33f167b).
- Não regressão por função: as suítes antigas (lobbies, candidatura, home, talentos, compartilhar) passam sem mudança de asserção. Só ganharam `formation: 'roles'` e `freeSlots: nil` nos corpos esperados.
