# Relatório de validação — candidatura-lobby, Parte 2 (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend `go test ./...` ok; `go test -tags=integration ./...` ok em todos os pacotes, `internal/applications` em 35 s; web Vitest 278/278 em 24 arquivos; Playwright 45/45) · lint ok (golangci-lint 0 issues; `gofmt -l` vazio; prettier e eslint ok) · build ok (`go build ./...`; `svelte-check` com 509 arquivos, 0 erros e 0 avisos)

Escopo: o mesmo do ciclo 1 (US-05 a US-08, RN-14, RN-15, RN-19 a RN-24, RN-27, RN-36 a RN-38, as partes de remoção, bloqueio e pedido de troca da RN-07, RN-16, RN-18, RN-25 e RN-26; CA-05.1 a CA-08.14 e CA-09.1 a CA-09.4 com pedido pendente; design P2.1 a P2.7, D-08 a D-13). A Parte 1 não foi revalidada.

Alterações desde o ciclo 1: `11d99fe` (fix em `backend/internal/applications/swaps.go` e 2 testes de integração) e `5f6c7ac` (relatório do ciclo 1 e T-17 no `tasks.md`). Nenhuma mudança em web, OpenAPI, queries ou migrações. Os itens que o ciclo 1 deu como ✅ foram mantidos com a mesma evidência, depois de conferir que o fix não os toca e que os testes deles continuam passando.

## Pendência do ciclo 1: resolvida
**[RN-16, D-10] Pedido de troca pendente num lobby cancelado.**
- `RequestSwap` agora lê a candidatura sem trava, trava o lobby (`LockLobby`, `SELECT ... FOR UPDATE`) depois do Usuário, trava a candidatura e só então lê o lobby com `GetLobby` e confere `isOpen`. Como `lobbies.Cancel` trava o mesmo lobby (`GetOwnLobbyForUpdate`) antes de `ExpirePendingSwapsForLobby`, as duas transações ficam em série: ou o pedido é gravado antes e o cancelamento o expira, ou o pedido vê o lobby cancelado e recebe `not_open`. Em READ COMMITTED, o `GetLobby` depois da trava enxerga o cancelamento já confirmado.
- Ordem das travas: `RequestSwap` faz Usuário do membro → lobby; `Cancel` faz Usuário do dono → lobby. Não há ciclo, então não há risco novo de deadlock.
- Defesa: `decideSwap` (aceitar e recusar) e `WithdrawSwap` passaram a recusar com `not_pending` quando o lobby não está aberto; `decideSwap` também recusa se a candidatura não está `accepted` (atende a observação do ciclo 1).
- Testes: `TestRequestSwap_RN16_ConcurrentWithCancel` (5 rodadas; pedido recusado com `not_open` ou gravado e depois `expired`) e `TestSwap_RN16_CancelledLobbyGuards` (pedido pendente num lobby cancelado não é aceito nem retirado, e o personagem da candidatura não muda).
- Conferência extra: rodei os dois testes num worktree temporário com o `swaps.go` de antes do fix. `TestSwap_RN16_CancelledLobbyGuards` falha 6/6, então ele prova as defesas. `TestRequestSwap_RN16_ConcurrentWithCancel` passa mesmo sem o fix (6 × 5 rodadas): a janela da corrida é estreita demais para o teste a acertar. A correção da corrida está garantida pelo código (trava em comum), mas o teste de concorrência não a reproduz (ver observações).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 17 | 0 | 0 | 0 |
| Critérios (CA) | 38 | 0 | 0 | 0 |
| Não funcionais | 4 | 0 | 0 | 0 |

Decisões do design (D-08 a D-13): 6 ✅.

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-07 (bloqueio) | ✅ | `applications.go` Apply: `IsBlocked` → `blocked`; query `IsBlocked` olha só `lobby_id` + `user_id` + `status='removed' AND blocked` | `TestRemove_CA06_6_CA06_7_Blocked`, `TestRemove_CA06_5_ApplyAgainWithoutBlock`, rota em `TestPart2FlowIntegration_*` (409 `blocked`) | |
| RN-14 | ✅ | `Leave` → `endMembership`: `not_yours`, `isOpen`, `status == accepted`, transição para `left` | `TestLeave_CA05_1_BeforeStart`, `TestLeave_CA05_2_NotOpen`, `TestLeave_RN14_OnlyOwnAcceptedApplication` | A vaga volta porque `freeSlots` conta só os aceitos. |
| RN-15 | ✅ | `Remove`: `validReason`, `not_owner`, `isOpen`, `SetApplicationBlocked` quando há bloqueio; coluna `applications.blocked` (migração 00006) | `TestRemove_CA06_1..CA06_7`, `TestApplications_RN15_Blocked` (db) | |
| RN-16 (pedido de troca) | ✅ | No início, o pedido é lido como expirado (`effectiveStatus`). No cancelamento, `lobbies.Cancel` grava `expired` com evento (`ExpirePendingSwapsForLobby`). Ciclo 2: `RequestSwap` trava o lobby (`LockLobby`, `swaps.go:127`) depois do Usuário, na ordem do D-05, e relê o lobby depois da trava; `decideSwap` (`swaps.go:251`) e `WithdrawSwap` (`swaps.go:308`) recusam com `not_pending` se o lobby não está aberto | `TestSwap_D12_Expire`, `TestSwapRequests_D12_ListExpireAndApply`, `TestRequestSwap_RN16_ConcurrentWithCancel`, `TestSwap_RN16_CancelledLobbyGuards` | Pendência do ciclo 1 resolvida (T-17, commit 11d99fe). |
| RN-18 (pedido de troca) | ✅ | `transitionSwap` grava o estado e o `swap_request_events` na mesma transação; a abertura grava o evento `→ pending` | `TestAcceptSwap_CA08_5` (2 eventos com autor), `TestLeave_CA05_3_CancelsSwap`, `TestSwapRequests_RN18_Events` | A troca do dono e a mudança de personagem no aceite não geram `application_event` (ver observações). |
| RN-19 | ✅ | `SwapOwnerCharacter`: `not_owner`, `isOpen`, `ownCharacter`, `checkSwap` (nível, vaga com a vaga atual contada como livre, conflito sem contar o próprio lobby) | `TestOwnerSwap_CA07_1..CA07_6`, `TestOwnerSwap_RN19_Guards` | |
| RN-20 | ✅ | `RequestSwap`: `validReason`, `ownCharacter`, `GetPendingSwapRequest` → `swap_pending`; índice único parcial `swap_requests_one_pending`, que `wrapSwap` converte | `TestRequestSwap_CA08_1..CA08_4`, `TestRequestSwap_RN20_Guards`, `TestSwapRequests_RN20_OnePending` (db) | |
| RN-21 | ✅ | `RequestSwap` não toca em `applications` | `TestRequestSwap_CA08_1_Pending` (personagem e ocupantes iguais) | |
| RN-22 | ✅ | `decideSwap`: `OwnerID != uid` → `not_owner`; `RejectSwap` com `validReason` | `TestDecideSwap_RN22_Guards`, `TestRejectSwap_CA08_8`, `TestRejectSwap_CA08_9_ReasonRequired` | |
| RN-23 | ✅ | `AcceptSwap`: `checkSwap` e depois `SetApplicationCharacter` + `transitionSwap` na mesma transação, com o lobby travado | `TestAcceptSwap_CA08_5..CA08_7`, `TestAcceptSwap_D10_ConcurrentLastSlot` | |
| RN-24 | ✅ | `CHECK` dos 6 estados; `cancelPendingSwap` em `endMembership`, antes da transição | `TestLeave_CA05_3_CancelsSwap`, `TestRemove_CA08_10_CancelsSwap`, `TestSwapRequests_RN20_RN22_RN24_Checks` | |
| RN-25 (pedido de troca) | ✅ | `CharacterInOpenLobby` conta também o `to_character_id` de pedido pendente em lobby aberto | `TestSwapTarget_CA09_1_CA09_3_Locked` | |
| RN-26 (pedido de troca) | ✅ | Mesma query, usada no `Delete` | `TestSwapTarget_CA09_1_CA09_3_Locked`, `TestSwapTarget_CA09_2_CA09_4_Free` | |
| RN-27 | ✅ | `WithdrawSwap`: `not_yours`, `not_pending`, transição para `withdrawn` | `TestWithdrawSwap_CA08_11`, `TestWithdrawSwap_CA08_12_NotYours` | |
| RN-36 | ✅ | `checkSwap` (dono e aceite) e `RequestSwap` comparam o nível com `MinLevel` | `TestOwnerSwap_CA07_6_BelowMinLevel`, `TestRequestSwap_CA08_13_BelowMinLevel`, `TestAcceptSwap_RN36_MinLevelAtAccept` | |
| RN-37 | ✅ | `left` não bloqueia: `HasActiveApplication` vê só as ativas e `WasRejected`/`IsBlocked` não contam `left` | `TestLeave_CA05_4_ApplyAgain`; e2e `candidatura-parte2.spec.ts` (Caio sai e se candidata de novo) | |
| RN-38 | ✅ | `lobbies/[id]/+page.svelte`: aviso do membro com "Pedir troca", "Sair do grupo", pedido pendente e "Retirar pedido"; bloco "Pedidos de troca"; `SwapPanel`; `PlayerPanel` com "Remover do grupo" e "Trocar personagem"; `applyState` → `blocked` esconde "Candidatar"; `/candidaturas` com "Sair do grupo" | `lobbies.spec.ts` (T-14/T-15), `owner.spec.ts`, `membership.spec.ts`, `candidaturas.spec.ts`; e2e `candidatura-parte2.spec.ts` (2 testes) | |

### Critérios
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-04.1 (pedido de troca) | ✅ | leitura como `expired` no início | `TestSwap_D12_Expire` (aceitar e retirar no início → `not_pending`) | |
| CA-05.1 | ✅ | `Leave` | `TestLeave_CA05_1_BeforeStart` (estado, ocupantes, evento); rota em `TestPart2FlowIntegration_*` | |
| CA-05.2 | ✅ | `isOpen` | `TestLeave_CA05_2_NotOpen` (iniciado e cancelado) | |
| CA-05.3 | ✅ | `cancelPendingSwap` | `TestLeave_CA05_3_CancelsSwap` | |
| CA-05.4 | ✅ | Apply | `TestLeave_CA05_4_ApplyAgain`; e2e | |
| CA-05.5 | ✅ | `LeaveDialog` + action `leave` | `membership.spec.ts`; e2e (o aviso muda e o card some) | |
| CA-06.1 | ✅ | `Remove` | `TestRemove_CA06_1_WithReason` (justificativa aparada, ocupantes, evento) | |
| CA-06.2 | ✅ | `validReason` | `TestRemove_CA06_2_ReasonRequired` | Mensagem por código (D-07), como na Parte 1. |
| CA-06.3 | ✅ | `not_owner` | `TestRemove_CA06_3_NotOwner`; rota 409 | |
| CA-06.4 | ✅ | `isOpen` | `TestRemove_CA06_4_NotOpen` | |
| CA-06.5 | ✅ | Apply | `TestRemove_CA06_5_ApplyAgainWithoutBlock` | |
| CA-06.6 | ✅ | `IsBlocked` | `TestRemove_CA06_6_CA06_7_Blocked` (3 personagens) | |
| CA-06.7 | ✅ | `IsBlocked` olha só o lobby | idem (outro lobby do mesmo dono → pendente) | |
| CA-06.8 | ✅ | `RemoveDialog`, aviso do removido, `/candidaturas` | e2e (remover com bloqueio; o removido não vê "Candidatar"; "Removida" com a justificativa) | |
| CA-07.1 | ✅ | `SwapOwnerCharacter` | `TestOwnerSwap_CA07_1_RoleWithSlot`; e2e | |
| CA-07.2 | ✅ | `checkSwap` (mesma função) | `TestOwnerSwap_CA07_2_SameRoleFull` | |
| CA-07.3 | ✅ | `role_full` | `TestOwnerSwap_CA07_3_RoleFull` (o dono continua com o atual) | |
| CA-07.4 | ✅ | `HasScheduleConflict` | `TestOwnerSwap_CA07_4_ScheduleConflict` | |
| CA-07.5 | ✅ | `not_owner` (409, como diz a revisão do P2.5) | `TestOwnerSwap_CA07_5_NotOwner`; rota em `TestPart2FlowIntegration_*` | |
| CA-07.6 | ✅ | `below_min_level` | `TestOwnerSwap_CA07_6_BelowMinLevel` | |
| CA-08.1 | ✅ | `RequestSwap` | `TestRequestSwap_CA08_1_Pending`; rota 201 | |
| CA-08.2 | ✅ | `validReason` | `TestRequestSwap_CA08_2_ReasonRequired`; e2e ("Escreva a justificativa") | |
| CA-08.3 | ✅ | `swap_pending` | `TestRequestSwap_CA08_3_SecondPending`; rota 409 | |
| CA-08.4 | ✅ | `ownCharacter` → 422 `characterId` | `TestRequestSwap_CA08_4_OtherUsersCharacter` | 422 por D-07, como a CA-01.4. |
| CA-08.5 | ✅ | `AcceptSwap` | `TestAcceptSwap_CA08_5`; rota e e2e | |
| CA-08.6 | ✅ | `role_full`, o pedido continua pendente | `TestAcceptSwap_CA08_6_RoleFull` | |
| CA-08.7 | ✅ | `schedule_conflict` | `TestAcceptSwap_CA08_7_ScheduleConflict` | |
| CA-08.8 | ✅ | `RejectSwap` | `TestRejectSwap_CA08_8` | |
| CA-08.9 | ✅ | `validReason` | `TestRejectSwap_CA08_9_ReasonRequired` | |
| CA-08.10 | ✅ | `cancelPendingSwap` | `TestRemove_CA08_10_CancelsSwap` | |
| CA-08.11 | ✅ | `WithdrawSwap` | `TestWithdrawSwap_CA08_11` (abre outro depois); e2e | |
| CA-08.12 | ✅ | `not_yours` + `swapRuleMessage` ("Esse pedido de troca não é seu.") | `TestWithdrawSwap_CA08_12_NotYours`; `membership.spec.ts` | |
| CA-08.13 | ✅ | `below_min_level` no pedido | `TestRequestSwap_CA08_13_BelowMinLevel` | |
| CA-08.14 | ✅ | aviso do membro, bloco "Pedidos de troca", `SwapPanel` | e2e (pedido, aviso pendente, bloco, painel pelo teclado com o motivo, aceitar); `owner.spec.ts` (atual → novo, Aceitar, Recusar) | |
| CA-09.1 (pedido) | ✅ | `CharacterInOpenLobby` | `TestSwapTarget_CA09_1_CA09_3_Locked` (nível e função travados; nick livre) | |
| CA-09.2 (pedido) | ✅ | idem | `TestSwapTarget_CA09_2_CA09_4_Free` (recusado, retirado, expirado, cancelado) | |
| CA-09.3 (pedido) | ✅ | idem | `TestSwapTarget_CA09_1_CA09_3_Locked` | |
| CA-09.4 (pedido) | ✅ | idem | `TestSwapTarget_CA09_2_CA09_4_Free` ("lobby iniciado") | |

### Não funcionais
| ID | Veredito | Evidência | Observação |
|---|---|---|---|
| RNF-01 | ✅ | `now := s.Now().UTC()`; colunas `timestamptz` | |
| RNF-02 | ✅ | `not_owner` e `not_yours` no serviço para remover, trocar do dono, decidir e retirar; `TestPart2FlowIntegration_*` chama as rotas direto | |
| RNF-03 | ✅ | `AcceptSwap` numa transação só; `TestAcceptSwap_D10_ConcurrentLastSlot` (5 rodadas, sempre 1 Tank); pedir troca e cancelar em paralelo em `TestRequestSwap_RN16_ConcurrentWithCancel` | |
| RNF-04 | ✅ | Os testes novos citam CA ou RN no nome ou no comentário | |

### Design
| ID | Veredito | Observação |
|---|---|---|
| D-08 | ✅ | `swap_requests` e `swap_request_events` como no P2.3 (FKs, CHECKs, índice parcial). |
| D-09 | ✅ | `applications.blocked`; `IsBlocked` restrito ao lobby. |
| D-10 | ✅ | Sair, remover, aceitar e agora pedir troca seguem a ordem Usuário → lobby → linha. `Cancel` trava Usuário do dono → lobby; `RequestSwap` trava Usuário do membro → lobby: as duas disputam a trava do lobby, sem ciclo de travas. Testes: `TestLeaveRemove_D10_Concurrent`, `TestAcceptSwap_D10_ConcurrentLastSlot`, `TestRequestSwap_RN16_ConcurrentWithCancel`. |
| D-11 | ✅ | `checkSwap`; `swapEligibility` na tela (o pedido pode esperar a vaga, a troca do dono não). |
| D-12 | ✅ | `swapRequests` só para o dono e só com o lobby aberto; `myApplication.swapRequest` e `blocked`. Testado em `TestGetLobby_D12_*` e no fluxo integrado, com dono, membro, outro membro, visitante e sem sessão. |
| D-13 | ✅ | `TestPart2_D13_Errors`: as 7 rotas × 9 códigos → 409, 404 e 422. |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum item 🚫.
- O selo do dono no detalhe ("1 pendente · 1 troca", `lobbies/[id]/+page.svelte`, `ownerBadge`), que o ciclo 1 pediu para confirmar, está na tela 2c do design aprovado. Tem teste em `web/src/routes/lobbies/lobbies.spec.ts` (`1 pendente · 1 troca`). Não é scope creep.
- "Ver grupo" ao lado de "Sair do grupo" em "Minhas candidaturas" continua como extensão pequena da tela, ligada à RN-38.

## Observações (não bloqueantes)
- **Teste de concorrência fraco:** `TestRequestSwap_RN16_ConcurrentWithCancel` não falha com o código antigo, porque as duas goroutines raramente se cruzam na janela crítica. Ele prova que o resultado final é coerente, mas não que a trava é necessária. Para reproduzir a corrida de forma confiável, seria preciso segurar uma das transações (por exemplo, abrir uma transação que trava o lobby e só liberar depois de disparar o pedido). O ciclo 1 da Parte 1 aceitou testes de concorrência do mesmo tipo.
- **Parte 1 sem regressão:** os testes da Parte 1 (Go, Vitest e e2e) passam. O fix só toca `swaps.go`, que é da Parte 2.
- **Códigos de erro novos no aceite e na retirada:** com o lobby cancelado, aceitar, recusar e retirar um pedido agora respondem `not_pending` (409), igual ao que já acontecia com o lobby iniciado. Está coerente com o D-12 e o D-13.
- Seguem valendo as observações do ciclo 1 que não foram tratadas e que a spec não exige: a troca do dono e a troca aceita não geram `application_event` (RN-18 fala de transições de estado); `Remove` valida a justificativa antes do dono (422 em vez de 409 para quem não é dono com justificativa vazia); a justificativa da recusa do pedido de troca sai na API, mas não aparece para o membro (descoberta da T-15, para o usuário decidir).
- **Rastreabilidade:** os commits do ciclo citam os IDs (`[RN-16, RN-24, D-05, D-10]`); a T-17 está marcada como feita no `tasks.md`, com o commit `11d99fe` e os arquivos certos. A rastreabilidade da T-16 que o ciclo 1 apontou foi corrigida (`candidatura-parte2.spec.ts` e `pages.ts`).
