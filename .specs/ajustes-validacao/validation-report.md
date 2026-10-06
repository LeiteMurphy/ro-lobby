# Relatório de validação — ajustes-validacao (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** testes passou (backend unitários ok; backend integração ok, com `-count=1`; web vitest 148/148; ponta a ponta Playwright 33/33) · lint ok (golangci-lint 0 issues; prettier, eslint e svelte-check sem erros) · build ok

Intervalo: `origin/feature/login-discord..HEAD` (6 commits, cbe4b97..bc9ff13).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Itens de ajuste (AJ) | 6 | 1 | 0 | 0 |
| Regras (RN-21, nova em personagens) | 1 | 0 | 0 | 0 |
| Não funcionais (RNF-03 das duas features) | 0 | 1 | 0 | 0 |
| Regressão (critérios de login-discord e personagens) | todos os testes passam | | | |

A spec não dá prioridade aos itens (nível P, uma task só), então todos contam como P1: o ⚠️ do AJ-06 impede a aprovação, porque o "Pronto quando" dele ("nenhum teste das duas features sem ID") não foi cumprido.

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| AJ-01 | ✅ | — (só teste) | `web/test/e2e/login.spec.ts:86` "CA-05.1: depois de entrar, volta logado para a página de origem" | Entra por `/auth/discord/login?next=%2Fperfil`, confere URL `/perfil`, o heading "Meus personagens" e o menu do usuário com "Grimbold". Como `/perfil` exige sessão (CA-06.1 de personagens), a chegada prova o login. Passou na execução |
| AJ-02 | ✅ | `backend/internal/auth/auth.go:70-99` (`pgx.BeginFunc` com `UpsertUserByDiscordID` e `CreateSession` no mesmo `tx`); `cmd/api/main.go:53` passa o pool | `TestLogin_RN13_SessionFailureCreatesNoUser` (`auth/auth_integration_test.go:95`): trigger que recusa o INSERT em `sessions`, login falha, `users` fica com 0 linhas | Banco descartável por teste (`testdb.New`), então o trigger não vaza. Rodado com `-v`: PASS. Sem a transação o teste deixaria 1 usuário |
| AJ-03 | ✅ | `backend/internal/discord/client.go:105-112` (401 → `slog.Error` com `client_id`, sem o segredo; 400 → só `ErrInvalidCode`) | `TestProfileFromCode_RN03_WrongSecret` (confere `level=ERROR`, a menção a `DISCORD_CLIENT_SECRET` e que "errado" não aparece) e `TestProfileFromCode_RN03_InvalidCodeDoesNotLogSecretWarning` (400 `invalid_grant`, log vazio) em `discord/client_test.go` | O falso responde 401 `invalid_client` (fake.go:164) e 400 `invalid_grant` (fake.go:172), como o "Pronto quando" pede |
| AJ-04 | ✅ | `web/src/routes/perfil/+page.svelte:25-42` (`staleForm` guardado ao abrir; `dialogForm` ignora o `form` que já existia) | `web/test/e2e/perfil.spec.ts:175` "RN-19 / AJ-04: reabrir..." (erro de nick, Cancelar, reabrir → Nick e Classe vazios, sem a mensagem) | Passou na execução |
| AJ-05 | ✅ | `.specs/personagens/design.md:155-156` | Conferido em `web/static/portraits/retrato-1.svg`: pixel art 16×16 (`crispEdges`) com fundo em faixas | Item só de texto |
| AJ-06 | ⚠️ | 13 testes ganharam ID no nome ou no comentário (diff de `catalog_test.go`, `discordfake/fake_test.go`, `server/auth_test.go`, `server/characters_test.go`, `characters.spec.ts`, `auth.spec.ts`, `page.server.spec.ts`, `page.spec.ts`) | Varredura de todos os `func Test*`, `it(`, `it.each`, `test(` do backend e do web | Ainda há teste de personagens sem ID: `it.each(...)('status %i vira %o')` em `web/src/lib/characters/characters.spec.ts:62-74` (nem o nome nem o `describe` "cliente da API de personagens" citam ID). E `TestValidate_AcceptsValidInput` (`backend/internal/characters/validate_test.go:23`) segue citando CA-02.2 só no comentário de uma asserção, caso que o relatório de personagens já listava |
| AJ-07 | ✅ | `backend/internal/server/server.go:23-41` (`ResponseErrorHandlerFunc: internalError`, que registra com `slog.ErrorContext` e responde "erro interno") | `TestAuthDiscord_RN13_UnexpectedErrorIs500` (`server/auth_test.go:106`) e `TestCharacters_RN21_UnexpectedErrorIs500` (`server/characters_test.go:193`): 500 e corpo sem "banco caiu" | Não há outro caminho de 500 nos handlers (grep em `internal/server` e `internal/health`). O log não é conferido por teste, mas o "Pronto quando" pede só o corpo |
| RN-21 (personagens) | ✅ | Web: `lib/characters` converte falha de rede/500 em `unavailable`; `/perfil` mostra o aviso. API: `server.go:37-41` | `characters.spec.ts` "RN-21: API fora do ar vira unavailable, sem lançar"; `page.server.spec.ts` "RN-21: API fora do ar mostra o aviso, sem quebrar a página"; `page.spec.ts` "RN-21: API fora do ar mostra o aviso"; `TestCharacters_RN21_UnexpectedErrorIs500` | Texto da regra entrou em `.specs/personagens/spec.md:85-89` |
| RNF-03 (login-discord) | ✅ | — | Varredura: todo teste da feature login cita ID | Os 4 testes apontados no relatório de login (`TestAuthorize_*`, `TestToken_*`, `TestAuthDiscord_*500`, "sem cookie...") agora citam RNF-04, RN-13 e RN-10 |
| RNF-03 (personagens) | ⚠️ | — | Ver AJ-06 | 1 teste sem ID e 1 com ID só no comentário de asserção |
| Regressão login-discord | ✅ | — | Toda a suíte (Go unitário e integração, vitest, Playwright `login.spec.ts` 8/8) passou | CA-05.1 passou a ser provado com `/perfil` em vez de `/status`, mais forte que antes |
| Regressão personagens | ✅ | — | Toda a suíte (Go, vitest, Playwright `perfil.spec.ts` 10/10) passou | `auth.NewService` agora recebe o pool; todas as chamadas foram ajustadas (main, testes de auth e de server) |

## Pendências para correção
1. [AJ-06 / RNF-03 personagens] `web/src/lib/characters/characters.spec.ts:62-74`: o `it.each(...)('status %i vira %o', ...)` não cita nenhum ID. Pôr o ID no nome (o mapeamento 401 → `no_session`, 404 → `not_found`, 409 → `limit`, 500 → `unavailable` se liga a RN-01/RN-03, CA-02.8 e RN-21) ou no `describe`. O "Pronto quando" do AJ-06 exige "nenhum teste das duas features sem ID".
2. [AJ-06 / RNF-03 personagens] `backend/internal/characters/validate_test.go:23` `TestValidate_AcceptsValidInput`: citar o ID (CA-02.2) no nome ou no comentário do teste, não só no comentário de uma asserção. O relatório de personagens já apontava este caso, e o AJ-06 foi criado para resolvê-lo.

## Scope creep
- Nenhum. Todas as alterações se ligam a AJ-01..AJ-07 ou RN-21. O `RequestErrorHandlerFunc` explícito em `server.go:26-28` repete o padrão do código gerado (400 com o texto do erro de decodificação) e só aparece porque o `StrictHTTPServerOptions` passou a ser montado; o comportamento não muda.

## Observações (não bloqueantes)
- AJ-07: o registro no log do `internalError` não tem teste. Dá para cobrir do jeito do AJ-03, com um logger injetável.
- AJ-03: o log usa `slog.Default()` em produção, e o `main` não injeta outro; está correto, só fica registrado que o formato segue o logger padrão da API.
- Rastreabilidade: os 6 commits seguem Conventional Commits e citam os IDs (AJ-* e os RN/CA envolvidos). Não há `tasks.md` (nível P).
- Fora de escopo respeitado: o `Update` de personagem continua fora da transação com trava e o ponta a ponta continua no banco de desenvolvimento, como a spec registra.
