# Relatório de validação — candidatura-lobby, Parte 1 (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend: todos os pacotes ok com `-tags=integration`; web Vitest 231/231 em 22 arquivos; Playwright 41/41) · lint ok (golangci-lint 0 issues; prettier + eslint ok) · build ok (`go build ./...`, `svelte-check` 501 arquivos, 0 erros e 0 avisos)

Conferências extras: `oapi-codegen` e `openapi-typescript` regerados a partir do `openapi.yaml` batem com `api.gen.go` e `schema.gen.ts`; `sqlc diff` sem diferenças.

Escopo validado: US-01 a US-04, US-10, RN-25, RN-26, RN-30 a RN-35 e CA-09.1 a CA-09.4. A RN-35 e a CA-10.6 vêm das decisões de 2026-10-06 depois da implementação. US-05 a US-08 (e as partes da RN-07, RN-16, RN-24, RN-25 e RN-29 que tratam de remoção, bloqueio ou pedido de troca) ficam fora, como diz a seção "Entrega em duas partes".

O que mudou desde o ciclo 1 (commits 1a5fde6, 6bd4604 e 0ff835b):
- RN-32: o Discord do anfitrião some do detalhe para quem não é dono nem membro aceito, e da lista pública `GET /lobbies`. O contrato passou `LobbyOwner.discordName` a `nullable`.
- RN-35 / CA-10.6 (decisão nova): "Candidatar" do card e do destaque da Home leva ao detalhe do lobby.
- D-05: o Apply passou a travar o lobby depois do Usuário. Isso fecha a corrida com o cancelamento apontada no ciclo 1.
- O teste `TestApply_CA01_10_PendingNearbyAllowed` agora monta o Given da CA-01.10.
- O e2e cita RN-31, e não mais RNF-01, no passo do teclado.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 26 | 0 | 0 | 0 |
| Critérios (CA) | 44 | 0 | 0 | 0 |
| Não funcionais | 4 | 0 | 0 | 0 |

Decisões do design (D-01 a D-07): 7 ✅.

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `internal/applications/applications.go` Apply: `GetOwnCharacter(ID, UserID)` → 422 `characterId/invalid` | `TestApply_CA01_4_OtherUsersCharacter`; rota 422 em `TestApplicationsFlowIntegration_*` | |
| RN-02 | ✅ | Apply: `lockUser` + `HasActiveApplication`; índice único parcial `applications_one_active_per_user` (migração 00005); `wrap` converte a violação em `already_active` | `TestApply_CA01_5_AlreadyActive`, `TestApplications_RN02_OneActivePerUser` (db) | |
| RN-03 | ✅ | Apply: `lobby.OwnerID == uid` → `own_lobby` | `TestApply_CA01_6_OwnLobby` | |
| RN-04 | ✅ | Apply: `Role: character.Role`; não há campo de função na entrada | `TestApply_CA01_1_Pending` (role = support) | |
| RN-05 | ✅ | `isOpen` + `freeSlots` (vagas − dono − aceitos) | `TestApply_CA01_7_RoleFull`, `TestApply_CA01_8_NotOpen` | |
| RN-30 | ✅ | Apply e Accept comparam `character.Level < MinLevel` | `TestApply_CA01_11_BelowMinLevel`, `TestAccept_RN30_MinLevelAtAccept` | |
| RN-06 | ✅ | Apply: `utf8.RuneCountInString > 250`; CHECK `char_length(message) <= 250` | `TestApply_CA01_2_CA01_3_Message`, `TestApplications_RN06_RN09_RN17_Checks` | A mensagem é aparada antes da contagem. |
| RN-07 | ✅ | `WasRejected` → `rejected_before` | `TestApply_CA01_9_RejectedBefore`; `TestWithdraw_CA03_2_to_CA03_5` (retirada pode voltar) | Parte do bloqueio na remoção é da Parte 2. |
| RN-08 | ✅ | `decide`: `OwnerID != uid` → `not_owner`; `effectiveStatus != pending` → `not_pending` | `TestDecide_CA02_6_NotOwner`, `TestDecide_CA02_7_CA04_3_NotPending` | |
| RN-09 | ✅ | `Reject`: `TrimSpace` e 10..250 runas; CHECK `char_length(btrim(reason)) BETWEEN 10 AND 250` | `TestReject_CA02_2_to_CA02_5` | |
| RN-10 | ✅ | Accept: `freeSlots < 1` → `role_full` depois da trava do lobby | `TestAccept_CA02_8_RoleFull` | |
| RN-11 | ✅ | `HasScheduleConflict` (janela aberta de ±2 h, dono ou aceito, lobby não cancelado) só no Accept | `TestAccept_CA02_10_to_CA02_12_Schedule`, `TestApplications_D04_ConflictCountsMembers` | |
| RN-12 | ✅ | D-05: `LockUser` do candidato → `LockLobby FOR UPDATE` → `GetApplicationForUpdate` | `TestAccept_CA02_9_Concurrent` (6 aceites simultâneos, 1 conclui) | |
| RN-13 | ✅ | `Withdraw`: `app.UserID != uid` → `not_yours`; `not_pending` | `TestWithdraw_CA03_2_to_CA03_5`, `TestExpire_CA04_1_ByStart` | |
| RN-16 | ✅ | Início: `lobbies.ApplicationStatus` (D-02); cancelamento: `ExpirePendingForLobby` + evento na mesma transação (`lobbies.Cancel`) | `TestExpire_CA04_1_ByStart`, `TestExpire_CA04_2_ByCancel` | Pedido de troca é da Parte 2. |
| RN-17 | ✅ | CHECK de estados na migração; transições só a partir de `pending` no serviço | `TestApplications_RN06_RN09_RN17_Checks`, `TestDecide_CA02_7_CA04_3_NotPending`, `TestWithdraw_*` (aceita não é retirada) | |
| RN-18 | ✅ | `transition` grava estado + `application_events`; criação e cancelamento também gravam evento | `TestHistory_CA03_6`, `TestApplications_RN18_Events`, `TestExpire_CA04_2_ByCancel` | Expiração pelo início não gera evento, como o D-02 aceita. |
| RN-28 | ✅ | `lobbies/detail.go`: `Pending` só para o dono, `Message` zerada para quem não é dono, `PendingCount` público | `TestApplicationsFlowIntegration_*` (sem sessão, terceiro, dono, candidato); `TestGetLobby_D06_Viewer`; e2e visitante | |
| RN-29 | ✅ | `myApplication` (com `reason`) só para quem se candidatou; dono vê a recusa no histórico da própria ação | `TestApplicationsFlowIntegration_*` (Caio vê, Duda não) | Parte da remoção é da Parte 2. |
| RN-25 | ✅ | `characters.Update`: trava quando muda função **ou** nível; `CharacterInOpenLobby` inclui pendente/aceita em lobby aberto | `TestApplicant_CA09_1_CA09_3_Locked`, `TestApplicant_CA09_2_*`; web `page.server.spec.ts`; e2e CA-09.1 | |
| RN-26 | ✅ | `characters.Delete` usa a mesma `checkNotInOpenLobby`; FK `ON DELETE SET NULL` | `TestApplicant_CA09_1_CA09_3_Locked`, `TestApplicant_CA09_4_DeleteAfterLobby` | |
| RN-31 | ✅ | `routes/lobbies/[id]/+page.svelte`: `selectedKey = HOST_KEY`, cards são `<button aria-pressed>`; `PlayerPanel.svelte` | `applications.spec.ts` (RN-31, CA-10.1); `lobbies.spec.ts`; e2e (Enter no card do membro) | |
| RN-32 | ✅ | `lobbies/detail.go:90-93` zera `Owner.DiscordName` quando quem olha não é dono nem membro aceito, e `:104-106` faz o mesmo nos membros; `lobbies.go:215-216` zera na lista pública; `openapi.yaml` `LobbyOwner.discordName` passa a `nullable`; `PlayerPanel.svelte:51-64` mostra o nome ou o aviso "O Discord aparece para quem está no grupo."; mensagem e Aceitar/Recusar só com `canDecide` | `TestApplicationsFlowIntegration_CA03_7_CA03_8_CA10_2_CA10_3` (anfitrião: sem sessão, Duda e Caio recusado → `null`; Bia membro e Ana dono → "Ana"; `GET /lobbies` → `null`); `applications.spec.ts` "RN-32: o anfitrião sem Discord…"; e2e `candidatura.spec.ts` (visitante não vê `dono…` no painel) | Corrigido no ciclo 2 (6bd4604). As outras respostas com `toAPILobby` (criar, editar, cancelar) só saem para o próprio dono. |
| RN-33 | ✅ | `routes/candidaturas/` + item no `UserMenu.svelte` abaixo de "Meu perfil"; `GET /me/applications` | `candidaturas.spec.ts`; `TestListMine_CA03_1` (serviço e rota); e2e login (ordem do menu) e candidatura | |
| RN-34 | ✅ | `home/lobbies.ts` `pendingFor` (só se `viewerId === ownerId`); `LobbyCard`/`FeaturedLobby` | `pending.spec.ts`; e2e CA-10.5 | |
| RN-35 | ✅ | `LobbyCard.svelte` e `FeaturedLobby.svelte` levam ao detalhe | ver CA-10.6 | Acrescentada nas decisões de 2026-10-06. |
| CA-01.1 | ✅ | Apply | `TestApply_CA01_1_Pending` (pendente, Suporte, vaga ainda livre); e2e | |
| CA-01.2 | ✅ | Apply | `TestApply_CA01_2_CA01_3_Message` (250 "é") | |
| CA-01.3 | ✅ | Apply + `applicationFieldMessage` | idem (251 → 422, nada criado) | Texto do erro é "Use até 250 caracteres" (D-07). |
| CA-01.4 | ✅ | Apply | `TestApply_CA01_4_OtherUsersCharacter` | Responde 422 `characterId/invalid` (D-07). |
| CA-01.5 | ✅ | Apply | `TestApply_CA01_5_AlreadyActive`; rota 409 `already_active` | |
| CA-01.6 | ✅ | Apply | `TestApply_CA01_6_OwnLobby` | |
| CA-01.7 | ✅ | Apply | `TestApply_CA01_7_RoleFull`; e2e "Tank sem vaga" | |
| CA-01.8 | ✅ | Apply | `TestApply_CA01_8_NotOpen` (cancelado, iniciado, inexistente) | |
| CA-01.9 | ✅ | Apply | `TestApply_CA01_9_RejectedBefore` | |
| CA-01.10 | ✅ | Apply não consulta conflito | `TestApply_CA01_10_PendingNearbyAllowed` agora aceita a Cura no lobby das 20:00 e candidata às 21:00 → pendente; também `TestAccept_CA02_10_to_CA02_12_Schedule` e o e2e | Given da spec montado (ciclo 2). |
| CA-01.11 | ✅ | Apply | `TestApply_CA01_11_BelowMinLevel` (199 × 200); e2e | |
| CA-02.1 | ✅ | Accept | `TestAccept_CA02_1` (Dano 0 vagas livres); rota e e2e | |
| CA-02.2 | ✅ | Reject | `TestReject_CA02_2_to_CA02_5`; e2e | |
| CA-02.3 | ✅ | Reject | idem ("" e só espaços → `required`; continua pendente) | |
| CA-02.4 | ✅ | Reject | idem ("  123456789  " → `too_short`) | |
| CA-02.5 | ✅ | Reject | idem (10 e 250 exatos) | |
| CA-02.6 | ✅ | decide | `TestDecide_CA02_6_NotOwner`; rota 409 `not_owner` | |
| CA-02.7 | ✅ | decide | `TestDecide_CA02_7_CA04_3_NotPending` (recusada, retirada, expirada) | |
| CA-02.8 | ✅ | Accept | `TestAccept_CA02_8_RoleFull` | |
| CA-02.9 | ✅ | D-05 | `TestAccept_CA02_9_Concurrent` | |
| CA-02.10 | ✅ | Accept | `TestAccept_CA02_10_to_CA02_12_Schedule`; e2e 2º teste | |
| CA-02.11 | ✅ | `starts_at < window_end` (exclusivo) | idem (22:00 aceito) | |
| CA-02.12 | ✅ | `cancelled_at IS NULL` | idem | |
| CA-03.1 | ✅ | `ListMine`, `/candidaturas` | `TestListMine_CA03_1` (serviço e rota); `candidaturas.spec.ts`; e2e | |
| CA-03.2 | ✅ | Withdraw | `TestWithdraw_CA03_2_to_CA03_5`; e2e | |
| CA-03.3 | ✅ | Withdraw | idem (aceita → `not_pending`, continua aceita) | |
| CA-03.4 | ✅ | Withdraw | idem (`not_yours`) | |
| CA-03.5 | ✅ | Apply | idem (nova pendente depois de retirar) | |
| CA-03.6 | ✅ | `transition`, `InsertApplicationEvent` | `TestHistory_CA03_6` (autor, `at` em UTC, justificativa) | Consulta só no banco; não há rota de histórico, e a spec não pede. |
| CA-03.7 | ✅ | `Detail` | `TestApplicationsFlowIntegration_*` (`pendingCount` 2, sem `pending`); e2e visitante | |
| CA-03.8 | ✅ | `Detail` | idem (Duda sem `myApplication`) | |
| CA-04.1 | ✅ | `ApplicationStatus` (D-02) | `TestExpire_CA04_1_ByStart` (pendente expirada, aceita mantida) | Pedido de troca é da Parte 2. |
| CA-04.2 | ✅ | `lobbies.Cancel` | `TestExpire_CA04_2_ByCancel`; e2e "Expirada" | |
| CA-04.3 | ✅ | decide | `TestDecide_CA02_7_CA04_3_NotPending` | |
| CA-09.1 | ✅ | `characters.Update` | `TestApplicant_CA09_1_CA09_3_Locked` (pendente e aceita; nick, classe, retrato e link livres); e2e com a mensagem exata | |
| CA-09.2 | ✅ | idem | `TestApplicant_CA09_2_FreeWithoutActiveLinks` | |
| CA-09.3 | ✅ | `characters.Delete` | `TestApplicant_CA09_1_CA09_3_Locked`; `page.server.spec.ts` (mesma mensagem) | |
| CA-09.4 | ✅ | `CharacterInOpenLobby` exige lobby aberto | `TestApplicant_CA09_4_DeleteAfterLobby` | |
| CA-10.1 | ✅ | `PlayerPanel` | `applications.spec.ts` CA-10.1; e2e (visitante, card destacado com `aria-pressed`) | |
| CA-10.2 | ✅ | `Detail` zera o Discord do membro e do anfitrião para quem não é dono nem membro | `TestApplicationsFlowIntegration_*` (5 papéis, membro e anfitrião); e2e | |
| CA-10.3 | ✅ | `Detail` + `PlayerPanel` (`canDecide`) | `applications.spec.ts` CA-10.3; e2e | |
| CA-10.4 | ✅ | `/candidaturas` | `candidaturas.spec.ts`; e2e | |
| CA-10.5 | ✅ | `pendingFor` | `pending.spec.ts`; e2e | |
| CA-10.6 | ✅ | `LobbyCard.svelte:73-79` e `FeaturedLobby.svelte:39-42`: "Candidatar" vira link para `/lobbies/[id]` (não mais `soon`) | `home.spec.ts` "CA-10.6 / RN-35…" (cada lobby com vaga tem `<a href="/lobbies/{id}">Candidatar`, nenhum desabilitado); e2e `home.spec.ts` "CA-02.5 / CA-10.6…" (clica e chega ao detalhe) | Novo no ciclo 2. |
| RNF-01 | ✅ | `timestamptz`, `.UTC()` na API, `fromUtcIso` no web | e2e mostra 20:00/21:30 de Brasília a partir de UTC; `TestHistory_CA03_6` | |
| RNF-02 | ✅ | Sessão em todas as rotas; dono e autoria checados no serviço | `TestApplications_RNF02_RequireSession`, `TestDecide_CA02_6_NotOwner`, `TestWithdraw_*` | |
| RNF-03 | ✅ | `pgx.BeginFunc` com estado + evento na mesma transação | `TestAccept_CA02_9_Concurrent`, `TestAccept_CA02_8_RoleFull` (falha mantém pendente) | |
| RNF-04 | ✅ | — | Todos os testes novos citam IDs | Ver observação sobre IDs trocados. |
| D-01 | ✅ | migração 00005 | testes db | |
| D-02 | ✅ | `ApplicationStatus`, `Cancel` | `TestExpire_*` | |
| D-03 | ✅ | `GetLobby`/`ListOpenLobbies` com contagens; `toLobby` | `TestApplications_D03_*`, `TestUpdate_RN18_SlotsBelowMembers` | |
| D-04 | ✅ | `HasScheduleConflict` | `TestApplications_D04_*` | |
| D-05 | ✅ | Apply e Accept: Usuário → `LockLobby FOR UPDATE`; Cancel: Usuário dono → `GetOwnLobbyForUpdate` (mesma ordem) | `TestAccept_CA02_9_Concurrent` | No ciclo 2 o Apply também trava o lobby (6bd4604), o que fecha a corrida com o cancelamento. Essa corrida não tem teste próprio. | |
| D-06 | ✅ | `Detail` + `toAPIDetail` | `TestGetLobby_D06_Viewer`, fluxo integrado | |
| D-07 | ✅ | `RuleError` → 409 `{error, code}`; 422 nos campos; `messages.ts` | `TestApply_D07_Errors`, `TestDecide_D07_Errors` | |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum item fora de escopo foi implementado. Os estados da Parte 2 (`left`, `removed`, `cancelled`) existem só no CHECK e nos rótulos, sem transição, como o design prevê.
- Continua do ciclo 1: o selo "N pendentes" também aparece para o dono no detalhe do lobby (`+page.svelte`, `pill--pend`), e não só na Home. É pequeno, ligado à RN-34 e não bloqueia.

## Observações (não bloqueantes)
- Os erros da spec ("função sem vaga", "mensagem acima de 250 caracteres" etc.) são códigos com textos próprios em pt-BR, como o D-07 decidiu. O texto literal dos Then é diferente.
- Na Home, o card de um lobby do próprio Usuário também mostra "Candidatar" (`LobbyCard.svelte` não olha o dono). O botão só leva ao detalhe, e a API recusa a candidatura do dono (RN-03), mas o rótulo pode confundir.
- A trava do lobby no Apply (D-05) não tem teste de concorrência próprio contra o cancelamento. A regra de ordem Usuário → lobby está no código e o teste de aceites simultâneos segue passando.
- O commit `cfa7707 docs: links do Notion nas tasks da candidatura` não cita IDs. É só documentação.
