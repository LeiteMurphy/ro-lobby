# Relatório de validação — candidatura-lobby, Parte 2 (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** testes passou (backend `go test ./...` ok e `go test -tags=integration ./...` ok em todos os pacotes; web Vitest 278/278 em 24 arquivos; Playwright 45/45) · lint ok (golangci-lint 0 issues; `gofmt -l` vazio; prettier e eslint ok) · build ok (`go build ./...`; `svelte-check` com 509 arquivos, 0 erros e 0 avisos)

Conferência extra: `go generate ./...` (oapi-codegen e sqlc) e `npm run generate` (openapi-typescript) não mudaram nada no repositório. O código gerado bate com `openapi.yaml` e com as queries.

Escopo validado: US-05 a US-08, RN-14, RN-15, RN-19 a RN-24, RN-27, RN-36 a RN-38, e as partes da RN-07, RN-16, RN-18, RN-25 e RN-26 que tratam de remoção, bloqueio ou pedido de troca. Os critérios são CA-05.1 a CA-08.14, CA-09.1 a CA-09.4 com pedido pendente e a parte de troca da CA-04.1. O design entra pelas seções P2.1 a P2.7 (D-08 a D-13). A Parte 1 não foi revalidada; a checagem de regressão está no fim.

Motivo da reprovação: a parte da RN-16 que trata do pedido de troca fica ⚠️ por uma corrida com o cancelamento do lobby (pendência 1). A RN-16 é da US-04, que é P1, e a regra de aprovação pede todos os itens P1 em ✅. Na Parte 1, o ciclo 1 apontou a mesma corrida na candidatura, e o D-05 a fechou no ciclo 2.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 16 | 1 | 0 | 0 |
| Critérios (CA) | 38 | 0 | 0 | 0 |
| Não funcionais | 4 | 0 | 0 | 0 |

Decisões do design (D-08 a D-13): 5 ✅, 1 ⚠️ (D-10, pela mesma corrida).

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-07 (bloqueio) | ✅ | `applications.go` Apply: `IsBlocked` → `blocked`; query `IsBlocked` olha só `lobby_id` + `user_id` + `status='removed' AND blocked` | `TestRemove_CA06_6_CA06_7_Blocked`, `TestRemove_CA06_5_ApplyAgainWithoutBlock`, rota em `TestPart2FlowIntegration_*` (409 `blocked`) | |
| RN-14 | ✅ | `Leave` → `endMembership`: `not_yours`, `isOpen`, `status == accepted`, transição para `left` | `TestLeave_CA05_1_BeforeStart`, `TestLeave_CA05_2_NotOpen`, `TestLeave_RN14_OnlyOwnAcceptedApplication` | A vaga volta porque `freeSlots` conta só os aceitos. |
| RN-15 | ✅ | `Remove`: `validReason`, `not_owner`, `isOpen`, `SetApplicationBlocked` quando há bloqueio; coluna `applications.blocked` (migração 00006) | `TestRemove_CA06_1..CA06_7`, `TestApplications_RN15_Blocked` (db) | |
| RN-16 (pedido de troca) | ⚠️ | No início do lobby, o pedido é lido como expirado (`effectiveStatus` no `decideSwap`/`WithdrawSwap`, `ApplicationStatus` no detalhe). No cancelamento, `lobbies.Cancel` grava `expired` com evento (`ExpirePendingSwapsForLobby`) | `TestSwap_D12_Expire`, `TestSwapRequests_D12_ListExpireAndApply` | Corrida com o cancelamento: ver a pendência 1. |
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
| RNF-03 | ✅ | `AcceptSwap` numa transação só; `TestAcceptSwap_D10_ConcurrentLastSlot` (5 rodadas, sempre 1 Tank) | |
| RNF-04 | ✅ | Os testes novos citam CA ou RN no nome ou no comentário | |

### Design
| ID | Veredito | Observação |
|---|---|---|
| D-08 | ✅ | `swap_requests` e `swap_request_events` como no P2.3 (FKs, CHECKs, índice parcial). |
| D-09 | ✅ | `applications.blocked`; `IsBlocked` restrito ao lobby. |
| D-10 | ⚠️ | Sair, remover e aceitar a troca seguem a ordem Usuário → lobby → linha; `TestLeaveRemove_D10_Concurrent` e `TestAcceptSwap_D10_ConcurrentLastSlot` cobrem esses casos. A trava que o D-10 dá para abrir pedido (só o Usuário do membro) não segura o cancelamento do lobby: ver a pendência 1. |
| D-11 | ✅ | `checkSwap`; `swapEligibility` na tela (o pedido pode esperar a vaga, a troca do dono não). |
| D-12 | ✅ | `swapRequests` só para o dono e só com o lobby aberto; `myApplication.swapRequest` e `blocked`. Testado em `TestGetLobby_D12_*` e no fluxo integrado, com dono, membro, outro membro, visitante e sem sessão. |
| D-13 | ✅ | `TestPart2_D13_Errors`: as 7 rotas × 9 códigos → 409, 404 e 422. |

## Pendências para correção
1. **[RN-16, D-10] Um pedido de troca aberto durante o cancelamento do lobby pode ficar `pending` num lobby cancelado.**
   - **Onde:** `backend/internal/applications/swaps.go`, `RequestSwap`. Ele trava o Usuário do membro (`lockUser`) e a candidatura (`GetApplicationForUpdate`), mas lê o lobby com `GetLobby`, sem trava. `lobbies.Cancel` trava o Usuário do dono e o lobby (`GetOwnLobbyForUpdate`) e roda `ExpirePendingSwapsForLobby`. As duas transações não disputam nenhuma trava em comum.
   - **Como acontece:** o `RequestSwap` lê o lobby como aberto, o `Cancel` roda o `UPDATE ... WHERE status='pending'` sem ver o pedido ainda não confirmado, e o pedido é gravado depois. Ele fica `pending` num lobby cancelado. A RN-16 exige que passe para *expirado*.
   - **Por que não há rede de segurança:** `decideSwap` e `WithdrawSwap` não conferem `isOpen`. O `effectiveStatus` só olha o horário de início, então o dono ainda poderia aceitar esse pedido pela API e trocar o personagem do membro num lobby cancelado.
   - **Precedente:** é a mesma corrida que o ciclo 1 da Parte 1 apontou no Apply e que o D-05 fechou ("Apply passou a travar o lobby depois do Usuário").
   - **O que a regra exige:** nenhum pedido pendente sobrevive ao cancelamento. Por exemplo, travar o lobby (`LockLobby`) depois do Usuário no `RequestSwap`, na ordem do D-05, e conferir `isOpen` no `decideSwap` e no `WithdrawSwap` como defesa.
   - **Teste:** falta um teste de concorrência entre pedir troca e cancelar o lobby, como os de D-10.

## Scope creep
- Nenhum item 🚫. Duas extensões pequenas da tela, ligadas à RN-38 ou ao desenho 2k:
  - o selo do dono no detalhe do lobby passou a somar "N trocas" a "N pendentes" (`lobbies/[id]/+page.svelte`, `ownerBadge`). A RN-34 fala só de pendentes, e na Home;
  - "Ver grupo" ao lado de "Sair do grupo" em "Minhas candidaturas".

  Vale confirmar com o usuário se o selo de trocas está no desenho aprovado.

## Observações (não bloqueantes)
- **Parte 1 sem regressão:** os testes da Parte 1 (Go, Vitest e e2e) passam. O que mudou nela:
  - a mensagem de `not_owner` virou "Só o anfitrião pode decidir.";
  - o aviso do membro virou "Você está no grupo com X (Função)" (o e2e da Parte 1 foi ajustado só nesse texto);
  - o `load` do detalhe agora carrega os personagens também para o dono (RN-19);
  - os helpers do e2e foram para `test/e2e/pages.ts`.
- **RN-18:** a troca do dono (`SetLobbyOwnerCharacter`) não deixa histórico. No aceite da troca, a mudança de personagem da candidatura só fica no evento do pedido, sem `application_event`. A RN-18 fala de transições de estado, e nenhuma das duas é uma, mas uma auditoria futura pode sentir falta.
- **Pedido aceito sem conferir o estado da candidatura:** `AcceptSwap` não confere `app.Status == accepted`. Hoje isso vale como invariante, porque sair e remover cancelam o pedido na mesma transação. Uma conferência explícita (`not_member`) deixaria o código mais robusto.
- **Ordem dos erros na remoção:** `Remove` valida a justificativa antes de conferir o dono. Quem não é dono e manda a justificativa vazia recebe 422, e não 409 `not_owner`. A spec não fixa a ordem.
- **Descoberta da T-15 em aberto:** a justificativa da recusa do pedido de troca sai na API (`myApplication.swapRequest.decisionReason`), mas não aparece para o membro. A spec não define isso; fica para o usuário decidir.
- **Rastreabilidade da T-16:** o tasks.md lista `web/test/e2e/candidatura.spec.ts` e `seed.ts` nos arquivos. O teste novo está em `web/test/e2e/candidatura-parte2.spec.ts`, e o helper em `pages.ts`.
- **Rastreabilidade geral:** os commits citam os IDs, e as tasks T-09 a T-16 marcadas como feitas batem com os commits e o código.
