# Relatório de validação — banco-de-talentos e candidatura-lobby, revisão de 2026-10-09: removidos e desbloqueio (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes backend passou (todos os pacotes, `go test -count=1 -tags=integration ./...`) · web vitest passou (323/323, 28 arquivos) · e2e passou (51/51) · lint ok (golangci-lint 0 issues; prettier + eslint ok) · check ok (0 erros, 0 avisos) · build ok · código gerado em dia (`go generate ./...` e `npm run generate` sem diferença no `git status`)

Escopo: `main..feature/talentos-removidos` (commits 2b1375c, 1071863, 86970bc). Itens validados: RN-13, CA-02.5 e CA-02.6 da `banco-de-talentos`; RN-15 revisada e CA-06.9 da `candidatura-lobby`; não regressão do resto das duas features.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 2 | 0 | 0 | 0 |
| Critérios (CA) | 2 | 1 | 0 | 0 |
| Não funcionais | 3 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-13 (banco-de-talentos) | ✅ | `backend/queries/talents.sql` (colunas `removed` e `blocked` por `ap.user_id = c.user_id`, ou seja, por pessoa); `backend/internal/talents/affinity.go:56,93,124` (`Probe.Lobby` só no `ForLobby`); `backend/internal/server/talents.go:148`; `openapi.yaml` (`Talent.removed`/`blocked`); `web/src/lib/talents/components/TalentCard.svelte` (selo e botão só com `canUnblock && talent.blocked`); `web/src/routes/lobbies/[id]/+page.svelte:503-507` (`canUnblock` só no painel do dono) | `TestForLobby_CA02_5_RemovedWithoutBlock`, `TestForLobby_CA02_6_RemovedBlockedAndUnblock`, `TestListLobbyTalents_CA02_1` (corpo com `removed`/`blocked`), vitest RN-13 em `lobbies.spec.ts`, e2e `CA-02.5 / CA-02.6 / CA-06.9` | Na criação e no catálogo `lobby_id` é nulo, e os dois campos ficam `false`. |
| CA-02.5 | ✅ | idem RN-13 | `TestForLobby_CA02_5_RemovedWithoutBlock` (Brasa e Cinza aparecem com `removed=true, blocked=false`); vitest "CA-02.5 / RN-13" (texto "Removido deste grupo" uma vez, sem "Desbloquear") | Sem e2e próprio para o caso sem bloqueio; o estado "Removido deste grupo" aparece no e2e depois do desbloqueio. |
| CA-02.6 | ✅ | idem RN-13; desbloqueio em `applications.go:384-419` | `TestForLobby_CA02_6_RemovedBlockedAndUnblock` (dois personagens bloqueados, desbloqueio por Cinza muda os dois, nova candidatura aceita); vitest "CA-02.6 / RN-13" (selo e form `?/unblock` com o `characterId`); e2e `talentos.spec.ts` (dois cards "Removido · bloqueado", clique em "Desbloquear", os dois viram "Removido deste grupo", botão some, nova candidatura pela API) | Cobertura completa do Then. |
| RN-15 revisada (candidatura-lobby) | ✅ | `backend/internal/applications/applications.go:384-419` (`GetOwnLobbyForUpdate` por dono, `isOpen`, `UnblockUserInLobby`); `backend/queries/applications.sql` `UnblockUserInLobby` (só zera `blocked`, por usuário do personagem, só neste lobby); `backend/internal/server/swaps.go:107-131` (401/404/409/204); `web/src/routes/lobbies/[id]/+page.server.ts:276-291` | `TestUnblock_CA06_9`, `TestUnblock_CA06_9_NotOpen`, `TestUnblockInLobby_CA06_9`, vitest "desbloqueio pelo painel", e2e | Permissão verificada no backend (RNF-02 da candidatura-lobby): quem não é dono recebe `ErrNotFound`. |
| CA-06.9 | ⚠️ | idem RN-15 | `TestUnblock_CA06_9` (não dono → 404, desbloqueio com outro personagem, status continua `removed`, nova candidatura pendente, repetir é no-op); `TestUnblock_CA06_9_NotOpen` (lobby cancelado → `CodeNotOpen`); handler 409; web 409 com aviso | Dois casos do Then sem asserção: (1) "continua no histórico **com a justificativa**" — o teste só confere o status, não o `reason`; (2) "lobby **iniciado**" — só o cancelado é testado. O código cobre os dois (`UPDATE` só toca `blocked`; `isOpen` exige `StartsAt.After(now)`). US-06 é P2, então não bloqueia. |
| RNF-02 (banco-de-talentos) | ✅ | regras na API (`talents.sql`, `applications.go`); o web só exibe | testes de integração acima | |
| RNF-04 (ambas) | ✅ | — | todos os testes novos citam CA-02.5, CA-02.6 ou CA-06.9 no nome ou comentário | |
| RNF-05 (banco-de-talentos) | ✅ | `toAPITalents(list, withDiscord)` inalterado | `TestListLobbyTalents_CA02_1` e e2e de catálogo seguem passando | Sem regressão. |

### Não regressão
- Todos os testes de integração do backend (applications, lobbies, talents, server etc.) passaram sem cache, incluindo CA-06.5 a CA-06.8 da `candidatura-lobby` e CA-01.x a CA-04.x da `banco-de-talentos`.
- A query `Affinity` ganhou o parâmetro `$1` (lobby) e renumerou os demais; o código gerado por sqlc bate com a query (`go generate` sem diferença) e os testes de afinidade (CA-02.1 a CA-02.4, CA-03.1) passam.
- `Talent` no OpenAPI passou a exigir `removed` e `blocked`; o catálogo `/talentos` e a contagem preenchem `false`. Vitest, svelte-check e e2e do catálogo passaram.

## Pendências para correção
Nenhuma bloqueante.

1. [CA-06.9, não bloqueante, P2] Em `backend/internal/applications/unblock_integration_test.go`, `TestUnblock_CA06_9` não confere que a justificativa ("mudamos o horário da run") continua na candidatura removida, e `TestUnblock_CA06_9_NotOpen` cobre só o lobby cancelado, não o iniciado. O Then pede as duas coisas.

## Scope creep
- Nenhum. Todas as alterações se ligam a RN-13, CA-02.5, CA-02.6, RN-15 ou CA-06.9. Nada toca a seção "Fora de escopo" (notificação, convite no site etc.).

## Observações (não bloqueantes)
- `POST /lobbies/{id}/unblock` com `characterId` vazio ou inválido devolve 404 "lobby não encontrado" (`applications.go:395-397`), o que confunde lobby com personagem. Personagem válido de quem não está bloqueado devolve 204 sem mudar nada, como documentado no OpenAPI.
- Quem foi removido sem bloqueio, voltou a se candidatar e depois foi recusado aparece com "Removido deste grupo", mas não pode se candidatar de novo (RN-07, recusa). A spec não trata esse caso; o selo não avisa da recusa.
- O CA-02.5 não tem e2e próprio; a cobertura vem da integração e do vitest, o que basta para o critério.

## Correções depois do ciclo 1
- [CA-06.9] `TestUnblock_CA06_9` confere que a candidatura removida continua com a justificativa; `TestUnblock_CA06_9_NotOpen` cobre também o lobby iniciado.
