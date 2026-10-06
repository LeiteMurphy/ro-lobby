# Relatório de validação — ajustes-validacao (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend unitários ok; backend integração ok, com `-count=1`; web vitest 148/148; ponta a ponta Playwright 33/33) · lint ok (golangci-lint 0 issues; prettier e eslint sem erros; svelte-check 0 erros e 0 avisos) · build ok

Intervalo: `origin/feature/login-discord..HEAD` (8 commits, cbe4b97..eb4a2b0). Desde o ciclo 1 entrou só o `1f9c6ca` (renomeia e comenta testes; nenhum código de produção mudou) e o relatório do ciclo 1.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Itens de ajuste (AJ) | 7 | 0 | 0 | 0 |
| Regras (RN-21, nova em personagens) | 1 | 0 | 0 | 0 |
| Não funcionais (RNF-03 das duas features) | 2 | 0 | 0 | 0 |
| Regressão (critérios de login-discord e personagens) | todos os testes passam | | | |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| AJ-01 | ✅ | — (só teste) | `web/test/e2e/login.spec.ts:86` "CA-05.1: depois de entrar, volta logado para a página de origem" | Entra por `/auth/discord/login?next=%2Fperfil`, confere URL `/perfil`, o heading "Meus personagens" e o menu com "Grimbold". Passou |
| AJ-02 | ✅ | `backend/internal/auth/auth.go:70-99` (`pgx.BeginFunc` com `UpsertUserByDiscordID` e `CreateSession` no mesmo `tx`) | `TestLogin_RN13_SessionFailureCreatesNoUser` (`auth/auth_integration_test.go:97`): trigger recusa o INSERT em `sessions`, login falha, `users` fica vazio | Passou na integração |
| AJ-03 | ✅ | `backend/internal/discord/client.go:105-113` (401 → `log.Error` com `client_id`, sem o segredo; 400 → só `ErrInvalidCode`) | `TestProfileFromCode_RN03_WrongSecret` (`discord/client_test.go:83`) e `TestProfileFromCode_RN03_InvalidCodeDoesNotLogSecretWarning` (`:101`) | Falso responde 401 `invalid_client` e 400 `invalid_grant`, como pede o "Pronto quando" |
| AJ-04 | ✅ | `web/src/routes/perfil/+page.svelte:25-42` (`staleForm` guardado ao abrir; `dialogForm` ignora o `form` anterior) | `web/test/e2e/perfil.spec.ts:175` "RN-19 / AJ-04: reabrir..." | Passou |
| AJ-05 | ✅ | `.specs/personagens/design.md` (D-04) | Conferido contra `web/static/portraits/` no ciclo 1; texto não mudou desde então | Item só de texto |
| AJ-06 | ✅ | `1f9c6ca`: `TestValidate_CA02_2_AcceptsValidInputWithDefaultPortrait` (com comentário "CA-02.2 / RN-10" acima), `it.each(...)('RN-01 / RN-02 / CA-02.8 / RN-21: status %i vira %o')`, e mais 4 nomes com ID | Varredura de todos os `func Test*` do backend (114) e de todos os `it(`/`it.each`/`test(` dos `*.spec.ts` do web e do e2e: nenhum sem ID no nome ou na linha de comentário logo acima | Os IDs citados existem nas specs (RN-01, RN-02, RN-04, RN-07, RN-09, RN-10, RN-13, RN-14, RN-21, CA-02.2..CA-02.8; D-03, D-04, D-11 do design de personagens) e batem com o que cada teste confere |
| AJ-07 | ✅ | `backend/internal/server/server.go:23-41` (`ResponseErrorHandlerFunc: internalError`: log com `slog.ErrorContext`, corpo "erro interno") | `TestAuthDiscord_RN13_UnexpectedErrorIs500` (`server/auth_test.go:106`) e `TestCharacters_RN21_UnexpectedErrorIs500` (`server/characters_test.go:193`): 500 e corpo sem o texto do erro | |
| RN-21 (personagens) | ✅ | Web: `lib/characters` converte falha de rede/500 em `unavailable`; `/perfil` mostra o aviso. API: `server.go:37-41` | `characters.spec.ts` "RN-21: API fora do ar vira unavailable, sem lançar"; `page.server.spec.ts` e `page.spec.ts` "RN-21: ..."; `TestCharacters_RN21_UnexpectedErrorIs500` | Texto em `.specs/personagens/spec.md` |
| RNF-03 (login-discord) | ✅ | — | Varredura: todo teste cita ID | |
| RNF-03 (personagens) | ✅ | — | Varredura: todo teste cita ID | As duas pendências do ciclo 1 foram resolvidas |
| Regressão login-discord | ✅ | — | Go unitário e integração, vitest, Playwright `login.spec.ts` 8/8 | |
| Regressão personagens | ✅ | — | Go unitário e integração, vitest, Playwright `perfil.spec.ts` 10/10 | |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. O commit novo só renomeia testes e acrescenta comentários com IDs [AJ-06]. O `RequestErrorHandlerFunc` explícito em `server.go:26-28` repete o padrão do código gerado (ver ciclo 1).

## Observações (não bloqueantes)
- AJ-07: o registro no log do `internalError` continua sem teste; dá para cobrir com um logger injetável, como no AJ-03.
- `TestValidate_CA02_2_AcceptsValidInputWithDefaultPortrait` confere só o retrato padrão; a parte "escolhe outro retrato" do CA-02.2 é coberta em outros testes da feature personagens, não aqui.
- `TestCharactersFlowIntegration_CA02_3_CA03_3_CA04_3` traz no nome só parte dos IDs; a lista completa (CA-01.3, CA-02.1, CA-02.3, CA-03.3, CA-04.3, CA-07.2) está no comentário acima, o que atende ao RNF-03.
- Rastreabilidade: os 8 commits seguem Conventional Commits e citam os IDs. Não há `tasks.md` (nível P).
- Fora de escopo respeitado: o `Update` de personagem continua fora da transação com trava e o ponta a ponta continua no banco de desenvolvimento.
