# Relatório de validação — login-discord (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes passaram (Go unitários: todos os pacotes ok · Go integração: todos os pacotes ok · Vitest 98/98 · Playwright 24/24 · smoke da pilha app ok) · lint ok (`gofmt -l` vazio, `golangci-lint` 0 issues com e sem `-tags=integration`, ESLint/Prettier ok, `svelte-check` 0 erros e 0 warnings) · build ok (`go build ./...`, `npm run build`) · código gerado em dia (`go generate` e `npm run generate` sem diferença no `git status`) · CI do PR #13 verde no HEAD `c34c049` (Backend, Web, Ponta a ponta, Pilha app)

Intervalo validado: `origin/main..feature/login-discord` (9 commits, 64 arquivos).

Além das suítes, executei o fluxo à mão com o Discord falso (`cmd/fakediscord` em 127.0.0.1:18090, API em :18080 e `vite preview` em :14173, todos encerrados no fim), com `curl` e sem JavaScript: login → callback → Home com cookie → repetição do callback → logout → `/me` depois de sair → `next` externo. Também rodei um teste Playwright temporário (apagado depois) para medir o anel de foco do menu do usuário.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 17 | 0 | 0 | 0 |
| Critérios (CA) | 23 | 1 | 0 | 0 |
| Não funcionais | 4 | 1 | 0 | 0 |

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `web/src/lib/auth/oauth.ts` `authorizeUrl` (scope=identify, response_type=code) | `oauth.spec.ts` "CA-01.3 / RN-01"; `auth.spec.ts` "CA-01.3 / RN-02"; e2e `login.spec.ts` CA-01.3; `discordfake` `TestAuthorize_CA01_3_RecordsScopeAndState` | Executado: Location com `scope=identify&response_type=code&state=…` |
| RN-02 | ✅ | `oauth.ts` `newState` (256 bits), `cookieOptions`, `STATE_MAX_AGE=600`; `callback/+server.ts` apaga o cookie antes de qualquer decisão; `sameState` | `oauth.spec.ts` "RN-02: state aleatório…", "comparação de state", "cookie de state…"; `auth.spec.ts` CA-04.2, CA-04.3, CA-04.6 | Executado: `rol_oauth_state` com `Max-Age=600; HttpOnly; SameSite=Lax` e `Max-Age=0` no retorno |
| RN-03 | ✅ | `config.go` (secret só na API); `web/src/lib/auth/config.ts` não lê o secret; `docker-compose.yml` não passa o secret ao web | `secret.spec.ts`; passo da CI "Client Secret fora do build"; `smoke-app.sh` (env do container do web e respostas); `TestLoad_RN03_RequiresDiscordCredentials` | Executado: marcador do secret ausente do build, das respostas do web e dos logs |
| RN-04 | ✅ | `discord/client.go` `ProfileFromCode` (token local, descartado); `auth.Login` | `TestProfileFromCode_RN04_ValidCode`, `TestProfileFromCode_RN04_RedirectURIMustMatch`, `TestSchema_CA01_4_NoTokensNorAvatar` | |
| RN-05 | ✅ | `queries/users.sql` `UpsertUserByDiscordID` (ON CONFLICT discord_id); `discord_id UNIQUE` | `TestUsers_RN05_UpsertByDiscordID`; `TestLogin_CA01_2_NextLoginUpdatesSameUser` | |
| RN-06 | ✅ | `migrations/00002_users_sessions.sql` (só id, discord_id, username, global_name, created_at, last_login_at em `timestamptz`); `auth.Login` usa `Now().UTC()` | `TestSchema_CA01_4_NoTokensNorAvatar` (lista exata de colunas); `TestProfileFromCode_RN06_NoGlobalName` | O falso devolve `avatar` e ele é ignorado |
| RN-07 | ✅ | `auth.newToken` (32 bytes `crypto/rand`), `HashToken` SHA-256; `CHECK (octet_length(token_hash)=32)` | `TestLogin_CA01_4_StoresOnlyTokenHash`; `TestSessions_RN07_OnlyHashIsAccepted` | Executado: cookie com token de 43 caracteres base64url; banco só com hash hex |
| RN-08 | ✅ | `oauth.ts` `cookieOptions` (Secure fora de localhost/127.0.0.1) | `oauth.spec.ts` "CA-06.5 / RN-08" (localhost e host externo); `auth.spec.ts` CA-06.5; e2e CA-06.5 | |
| RN-09 | ✅ | `queries/sessions.sql` `GetActiveSession` (`last_used_at > agora-30d`), `TouchSession` (`< agora-1h`); `auth.Authenticate`; `hooks.server.ts` renova o cookie | `TestAuthenticate_CA06_3_SlidingExpiry`, `TestAuthenticate_CA06_4_ExpiredAfter30Days`, `TestAuthenticate_RN09_TouchAtMostHourly`, `TestSessions_RN09_*` | |
| RN-10 | ✅ | `auth.ErrNoSession`; `ON DELETE CASCADE`; `hooks.server.ts` apaga o cookie no 401 | `TestAuthenticate_CA06_2_UnknownOrMalformedToken`; `TestSessions_RN10_DeletedUserEndsSessions`; `auth.spec.ts` "CA-06.2 / RN-10" | |
| RN-11 | ✅ | `auth.Logout` → `DeleteSession` por hash; `logout/+server.ts` apaga o cookie | `TestLogout_CA03_2_OtherSessionsSurvive`; `TestSessions_RN11_DeleteOnlyThisSession`; `TestAuthFlowIntegration_CA06_1_LoginMeLogout`; e2e CA-03.1 | Executado: `/me` responde 401 depois de sair |
| RN-12 | ✅ | `oauth.ts` `safeNext` (só `/…`, sem `//`, `/\` e caracteres de controle); `decodeStateCookie` reaplica `safeNext`; `display.ts` `loginHref` | `oauth.spec.ts` CA-05.1 e CA-05.2; `display.spec.ts` CA-05.1 | Executado: `next=https://exemplo.com` e `next=//exemplo.com` voltam para `/` |
| RN-13 | ✅ | `callback/+server.ts` (erro, state, code ausente → `/?login=erro` sem chamar a API); `discord.DefaultTimeout=5s`; `+page.svelte` mostra `LOGIN_ERROR_MESSAGE` | `auth.spec.ts` CA-04.1..CA-04.6; `TestLogin_CA04_4_*`, `TestLogin_CA04_5_*`; `TestDefaultTimeout_RN13_IsFiveSeconds`; `home.spec.ts` CA-04.1; e2e CA-04.1 e CA-04.2 | |
| RN-14 | ✅ | `TopBar.svelte` (link "Entrar com Discord"; "Criar lobby" segue `soon`); `Button`/`IconButton` com `href` | `home.spec.ts` "RN-14" e CA-02.5; e2e `home.spec.ts` CA-02.5 | |
| RN-15 | ✅ | `UserMenu.svelte` (`.tile` com a inicial, nome, menu "Sair"); `display.ts` | `display.spec.ts`; `home.spec.ts` CA-02.1/CA-02.2/CA-02.4; e2e CA-01.1/CA-02.1 e CA-02.4 | |
| RN-16 | ✅ | `openapi.yaml` (`POST /auth/discord`, `GET /me`, `DELETE /session`, `securitySchemes.sessionToken`); `server/auth.go` | `TestContract_RN16_AuthRoutesAreDescribed`; `server/auth_test.go` (201, 400, 502, 200, 401, 204, 405) | |
| RN-17 | ✅ | `config.go` `DISCORD_API_BASE_URL` com padrão oficial; `web/src/lib/auth/config.ts` `DISCORD_AUTHORIZE_URL` com padrão oficial; `discordfake` e `cmd/fakediscord` | `TestLoad_RN17_DiscordDefaults`; `auth.spec.ts` (URL de autorização mockada); integração e e2e usam o falso | |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | `auth.Login`, `callback/+server.ts` | `TestLogin_CA01_1_FirstLoginCreatesUserAndSession`; `TestAuthDiscord_CA01_1_Created`; `auth.spec.ts` CA-01.1; e2e CA-01.1 (nome na barra) | |
| CA-01.2 | ✅ | `UpsertUserByDiscordID` | `TestLogin_CA01_2_NextLoginUpdatesSameUser` (mesmo ID, 1 linha, "Grimbold, o Sábio") | |
| CA-01.3 | ✅ | `login/+server.ts`, `oauth.authorizeUrl` | e2e CA-01.3 confere scope, response_type e state na URL real | |
| CA-01.4 | ✅ | migração 00002 | `TestSchema_CA01_4_NoTokensNorAvatar`; `TestLogin_CA01_4_StoresOnlyTokenHash` | |
| CA-01.5 | ✅ | ver RN-03 | `secret.spec.ts`; CI (build com marcador); `smoke-app.sh` | Executado localmente: marcador ausente de `build/`, `.svelte-kit/output` e das respostas de `/`, `/status`, `/?login=erro` e `/auth/discord/login` |
| CA-02.1 | ✅ | `UserMenu.svelte` | `home.spec.ts` CA-02.1; e2e CA-02.1 ("G" e "Grimbold", sem link de login) | |
| CA-02.2 | ✅ | `display.displayName` | `display.spec.ts` CA-02.2; `home.spec.ts` CA-02.2 ("M" e "mirai.exe") | Executado: usuário sem nome de exibição → tile "M" e "mirai.exe" |
| CA-02.3 | ✅ | `hooks.server.ts` → `locals.user` → `+layout.server.ts` → `TopBar` | `home.spec.ts` CA-02.3 (render no servidor); `auth.spec.ts` CA-06.1 (hook) | Executado: `curl` da Home com o cookie traz `data-testid="user-menu"`, a inicial e o nome, sem "Entrar com Discord" |
| CA-02.4 | ✅ | sem coluna de avatar; nenhuma `<img>` externa | e2e "CA-02.4 / RNF-05" (só a origem do web após recarregar logado); `TestSchema_CA01_4_NoTokensNorAvatar` | |
| CA-03.1 | ✅ | `UserMenu.svelte` (form POST `/auth/logout`), `logout/+server.ts`, `DeleteSession` | e2e CA-03.1 (cookie removido, barra volta, persiste após reload); `auth.spec.ts` CA-03.1; `TestAuthFlowIntegration_CA06_1_LoginMeLogout` (401 depois) | |
| CA-03.2 | ✅ | `DeleteSession` por hash | `TestLogout_CA03_2_OtherSessionsSurvive` | |
| CA-04.1 | ✅ | `callback/+server.ts` (`error` → erro) | e2e CA-04.1 (mensagem exata, sem cookie); `auth.spec.ts` CA-04.1 (API não chamada) | |
| CA-04.2 | ✅ | `sameState` | `auth.spec.ts` CA-04.2 (API não chamada); e2e CA-04.2 | |
| CA-04.3 | ✅ | `decodeStateCookie` → null | `auth.spec.ts` CA-04.3 | |
| CA-04.4 | ✅ | `discord.exchange` 400/401 → `ErrInvalidCode` → 400 | `TestLogin_CA04_4_InvalidCodeCreatesNothing`; `TestAuthDiscord_CA04_4_InvalidCode`; `auth.spec.ts` CA-04.4 | |
| CA-04.5 | ✅ | prazo de 5 s no cliente; `exchangeCode` aborta em 6 s | `TestProfileFromCode_CA04_5_SlowDiscord` (erro no prazo), `TestDefaultTimeout_RN13_IsFiveSeconds`, `TestLogin_CA04_5_SlowDiscordCreatesNothing`, `TestAuthDiscord_CA04_5_DiscordUnavailable`, `auth.spec.ts` CA-04.5 | O limite de 6 s vem da composição 5 s (API) + redirect; não há teste de ponta a ponta com o Discord lento (o `cmd/fakediscord` não tem flag de atraso) |
| CA-04.6 | ✅ | cookie de state apagado no primeiro retorno | `auth.spec.ts` CA-04.6; `TestProfileFromCode_CA04_6_CodeIsSingleUse` | Executado: repetir o callback → `/?login=erro` |
| CA-05.1 | ⚠️ | `safeNext`, `loginHref`, `callback` redireciona para `saved.next` | e2e CA-05.1; `oauth.spec.ts` / `display.spec.ts` CA-05.1 | O e2e entra direto por `/auth/discord/login?next=/status` e só confere "API online", que aparece também para visitante; não prova o "logado". A página `/status` não tem barra superior nem botão de login, então o Given "visitante em /status… clica em entrar" não é alcançável pela interface. Executado com `curl`: o callback devolve `Location: /status` junto com o `Set-Cookie: rol_session`. P2 |
| CA-05.2 | ✅ | `safeNext` | `oauth.spec.ts` CA-05.2 (https, `//`, `/\`, controle) | Executado com os dois valores do critério |
| CA-06.1 | ✅ | `server/auth.go` `GetMe` | `TestMe_CA06_1_ValidSession`; `TestAuthFlowIntegration_CA06_1_LoginMeLogout` | |
| CA-06.2 | ✅ | `GetMe` 401; hook apaga cookie | `TestMe_CA06_2_NoSession`; `TestAuthenticate_CA06_2_*`; `TestAuthenticate_CA06_4_*` (vencido); `auth.spec.ts` CA-06.2 | |
| CA-06.3 | ✅ | `Authenticate` + `TouchSession` | `TestAuthenticate_CA06_3_SlidingExpiry` (válida aos 29 dias e de novo 29 dias depois do último uso) | Não testa o limite superior (vencer 30 dias depois do novo uso), mas a consulta é a mesma do CA-06.4 |
| CA-06.4 | ✅ | `GetActiveSession` | `TestAuthenticate_CA06_4_ExpiredAfter30Days`; `TestSessions_RN09_ActiveUntil30DaysAfterLastUse` | |
| CA-06.5 | ✅ | `cookieOptions` | e2e CA-06.5 (localhost, `secure:false`); `oauth.spec.ts` (Secure fora de localhost) | Executado: `HttpOnly; SameSite=Lax; Path=/; Max-Age=2592000` |

### Não funcionais
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RNF-01 | ✅ | state + cookie HttpOnly, hash do token, secret só na API | ver RN-02, RN-03, RN-07, RN-08 | |
| RNF-02 | ✅ | `UserMenu.svelte` (Enter abre, foco vai para "Sair", Escape fecha e devolve o foco); `:focus-visible` global em `tokens/base.css` | e2e "RNF-02: o menu do usuário funciona pelo teclado" | O teste do repositório não confere o anel de foco. Executei um Playwright temporário: chegando por Tab, o gatilho e o "Sair" têm o `box-shadow` âmbar do `--focus-ring` |
| RNF-03 | ⚠️ | — | — | 4 testes novos não citam nenhum ID da spec: `TestAuthorize_AllowRedirectsWithCode` e `TestToken_RejectsJSON` (`discordfake/fake_test.go`), `TestAuthDiscord_UnexpectedErrorIs500` (`server/auth_test.go`) e "sem cookie, não chama a API" (`routes/auth/auth.spec.ts`). Vários outros citam só RN/D (`D-08: redirect_uri…`, `RN-02…`), o que segue a convenção que o repositório já usa |
| RNF-04 | ✅ | `discordfake`, `cmd/fakediscord`, `playwright.config.ts` com `reuseExistingServer: false` e URLs do falso; CI com credenciais fictícias | integração e e2e verdes sem rede externa; CI verde | |
| RNF-05 | ✅ | nenhuma imagem/recurso externo | e2e "CA-02.4 / RNF-05"; `home.spec.ts` CA-02.4 | |

## Pendências para correção
Nenhuma bloqueante. Recomendadas:
1. [CA-05.1] (P2) O e2e não comprova o "logado" depois de voltar para `/status`: conferir o cookie `rol_session` (ou outro sinal de sessão) após o redirect. Registrar também que `/status` não tem botão de login, então o retorno à página de origem só é alcançável pela interface a partir da Home (com os filtros na query).
2. [RNF-03] Pôr um ID da spec nos 4 testes sem ID listados acima, ou removê-los se não se ligarem a nenhum critério.

## Scope creep
- Nenhum. Todas as alterações se ligam a IDs da spec ou a decisões D-01..D-09:
  - `Button.svelte` e `IconButton.svelte` com `href`: necessários para o link "Entrar com Discord" (RN-14).
  - `health.go` (asserção parcial da interface) e `server.go`: composição das rotas com o contrato novo (RN-16).
  - `sqlc.yaml` (`timestamptz` → `time.Time`), `go.mod` (`uuid`, `oapi-codegen/runtime`): suporte a D-01/D-02.
  - `ci.yml`, `smoke-app.sh`, `docker-compose.yml`, `.env.example`, `README.md`: CA-01.5, RNF-04 e D-09.
  - `research.md`: documentação da spec.
- Nada toca "Fora de escopo": não há avatar, e-mail, guilds, "sair de todos", painel de sessões nem `redirect_uri` de produção.
- `cmd/fakediscord` fica fora da imagem da API: o `backend/Dockerfile` compila só `cmd/api`, `cmd/migrate` e `cmd/healthcheck` (D-05).

## Observações (não bloqueantes)
- **Secret errado vira "código recusado".** `discord/client.go` trata o 401 do endpoint de token (`invalid_client`, Client Secret errado ou trocado) como `ErrInvalidCode` → 400, igual a um código recusado. Um secret mal configurado vai aparecer para a pessoa como falha comum de login, sem nenhum sinal no log da API. Vale tratar como 502/500 e registrar no log, sem expor o secret.
- **Login fora de transação.** `auth.Login` faz o upsert do Usuário e depois cria a Sessão sem transação. Se a criação da Sessão falhar, o Usuário fica criado ou atualizado. A RN-13 só exige isso para falhas do Discord, então não viola a spec.
- **O e2e acumula dados no banco de dev.** Encontrei 20 sessões de execuções anteriores. Na minha execução manual, o Discord falso usa o mesmo ID padrão do e2e (`100000000000000001`), então atualizei esse Usuário e depois apaguei ele (as sessões caíram junto, em cascata). O próximo `npm run test:e2e` recria o Usuário.
- **CA-04.5 sem ponta a ponta com atraso.** O `cmd/fakediscord` não tem flag de atraso. O limite de 6 s está garantido por composição (5 s na API e 6 s de abort no web) e por testes de unidade e integração, não pelo fluxo completo.
- **Rastreabilidade em `tasks.md`.** As 8 tasks estão marcadas `[x]` e correspondem a código existente, mas o campo "Commit:" de todas continua "—". Os 9 commits seguem Conventional Commits e citam os IDs.
- **Uma ida à API por requisição.** O hook chama `GET /me` e regrava o cookie de sessão em toda requisição com cookie. É a consequência prevista no ADR-07.
- **`Error.enum` compartilhado.** O schema `Error` no contrato junta `invalid_code`, `discord_unavailable` e `no_session` num único enum para todas as rotas. Funciona, mas o contrato fica menos preciso por rota.
