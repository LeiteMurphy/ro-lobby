# Tasks — Login com Discord

- Spec: `./spec.md` · ADR: `docs/adr/0007-login-discord-sessao.md`
- Notion: [Épico](https://app.notion.com/p/3f0d4a3a5eff81509987fc63b41ecb4a)
- Branch: `feature/login-discord`

## Decisões de implementação
- **D-01 — Tabelas.** `users`: `id uuid` (gerado no banco), `discord_id text unique`,
  `username text`, `global_name text null`, `created_at` e `last_login_at` em
  `timestamptz`. `sessions`: `token_hash bytea` (chave primária, SHA-256 do token),
  `user_id` com `on delete cascade`, `created_at` e `last_used_at`. O vencimento é
  `last_used_at + 30 dias`, calculado na consulta (RN-05 a RN-07, RN-09).
- **D-02 — Contrato.** Três rotas novas no `openapi.yaml`:
  - `POST /auth/discord` com `{code, redirectUri}`: 201 com `{sessionToken, user}`, 400
    para código recusado, 502 para Discord fora do ar ou lento;
  - `GET /me` com `Authorization: Bearer <token>`: 200 com o usuário, 401 sem sessão;
  - `DELETE /session` com o mesmo header: 204.
- **D-03 — Rotas do web.**
  - `GET /auth/discord/login?next=/caminho`: grava o cookie de `state` (com o destino) e
    redireciona para o Discord.
  - `GET /auth/discord/callback`: confere o `state`, chama a API, grava o cookie de sessão
    e redireciona para o destino, ou para `/?login=erro` em caso de falha.
  - `POST /auth/logout`: chama a API, apaga o cookie e volta para a Home.
  - `hooks.server.ts`: com o cookie de sessão, pergunta `GET /me` à API e põe o usuário
    em `locals`; um 401 apaga o cookie (RN-10).
- **D-04 — Cliente do Discord.** Em Go, com timeout de 5 s. A URL da API do Discord vem
  de `DISCORD_API_BASE_URL` (padrão `https://discord.com/api`) e a URL de autorização,
  usada pelo web, de `DISCORD_AUTHORIZE_URL` (padrão `https://discord.com/oauth2/authorize`)
  (RN-17).
- **D-05 — Discord falso.** Um pacote Go (`internal/discordfake`) imita a autorização, o
  token e o `/users/@me`, e um comando (`cmd/fakediscord`) sobe esse servidor para o
  ponta a ponta. Ele não entra na imagem da API. Os testes de integração usam o pacote
  com `httptest` (RNF-04).
- **D-06 — Token de sessão.** 32 bytes de `crypto/rand` em base64url, enviado à API no
  header `Authorization: Bearer`. O "agora" é injetado no serviço, para os testes de
  vencimento (CA-06.3, CA-06.4) não dependerem do relógio.
- **D-07 — Cookies.** `rol_session` (sessão, 30 dias) e `rol_oauth_state` (10 minutos,
  `state` e destino). Os dois são `HttpOnly`, `SameSite=Lax` e `Path=/`, e são `Secure`
  quando o host não é `localhost` nem `127.0.0.1` (RN-08).
- **D-08 — `redirect_uri`.** O web monta o endereço a partir da própria origem
  (`<origem>/auth/discord/callback`) e repassa à API, que precisa enviar o mesmo valor ao
  Discord. No Developer Portal ficam cadastrados os de `localhost:3000` e `localhost:5173`.
- **D-09 — Variáveis.** `DISCORD_CLIENT_ID` (web e API), `DISCORD_CLIENT_SECRET` (só
  API), `DISCORD_API_BASE_URL` (API) e `DISCORD_AUTHORIZE_URL` (web). O `.env.example`
  traz o Client ID e um secret fictício (RN-03).

## US-06 e US-01 — Base no backend

### T-01 — Tabelas de Usuário e Sessão  [x]
- Cobre: RN-05, RN-06, RN-07, RN-09, CA-01.4, D-01
- Depende de: —
- Paralelizável: não
- Arquivos: `backend/migrations/00002_users_sessions.sql`, `backend/queries/users.sql`,
  `backend/queries/sessions.sql`, `backend/internal/db/` (gerado)
- Pronto quando: a migração aplica e reverte num banco vazio; testes de integração das
  queries passam (upsert pelo `discord_id`, sessão por hash, vencimento); nenhuma coluna
  guarda token em texto nem avatar.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff81bfbb38f9cc7350b37d

### T-02 — Cliente do Discord e Discord falso  [P] [x]
- Cobre: RN-01, RN-04, RN-13, RN-17, RNF-04, D-04, D-05
- Depende de: —
- Paralelizável: [P] com T-01
- Arquivos: `backend/internal/discord/`, `backend/internal/discordfake/`,
  `backend/cmd/fakediscord/`
- Pronto quando: testes do cliente contra o falso passam para código aceito, código
  recusado e Discord lento (erro em até 5 s); o pedido de token vai como
  `application/x-www-form-urlencoded` com o `redirect_uri`.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff8146bc6ddac5e0160af2

### T-03 — Serviço de autenticação  [x]
- Cobre: RN-04, RN-05, RN-07, RN-09, RN-10, RN-11, CA-01.2, CA-03.2, CA-06.3, CA-06.4, D-06
- Depende de: T-01, T-02
- Paralelizável: não
- Arquivos: `backend/internal/auth/`
- Pronto quando: testes de integração passam para primeiro login, login seguinte
  (mesmo Usuário, nome atualizado), sessão renovada pelo uso, sessão vencida, renovação
  gravada no máximo uma vez por hora e sair sem afetar outra sessão.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff81d9a931c29fe5cca219

### T-04 — Contrato e rotas da API  [x]
- Cobre: RN-16, CA-06.1, CA-06.2, CA-04.4, CA-04.5, D-02
- Depende de: T-03
- Paralelizável: não
- Arquivos: `openapi.yaml`, `backend/internal/api/api.gen.go`, `backend/internal/server/`,
  `backend/cmd/api/main.go`, `backend/internal/config/`
- Pronto quando: o contrato descreve as três rotas; o código gerado está em dia; testes
  das rotas passam (201, 400, 502, 200, 401, 204), com o Discord falso.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff810a99fac5e2f7a06c5d

## US-01, US-04 e US-05 — Fluxo no web

### T-05 — Rotas de login, callback e logout no web  [x]
- Cobre: RN-02, RN-08, RN-10, RN-12, RN-13, CA-01.3, CA-04.1, CA-04.2, CA-04.3,
  CA-04.6, CA-05.1, CA-05.2, CA-06.5, D-03, D-07, D-08
- Depende de: T-04
- Paralelizável: não
- Arquivos: `web/src/routes/auth/`, `web/src/hooks.server.ts`, `web/src/lib/auth/`,
  `web/src/lib/api/schema.gen.ts`
- Pronto quando: testes Vitest cobrem a URL de autorização (escopo, `state`), o cookie de
  `state` de uso único, os erros sem chamar a API, o destino só relativo e os atributos
  dos cookies.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff819dadf8f88780fe3403

## US-02 e US-03 — Barra superior

### T-06 — Usuário na barra, menu "Sair" e mensagem de erro  [ ]
- Cobre: RN-14, RN-15, CA-02.1, CA-02.2, CA-02.3, CA-03.1, RNF-02
- Depende de: T-05
- Paralelizável: não
- Arquivos: `web/src/lib/home/components/TopBar.svelte`, `web/src/lib/home/components/UserMenu.svelte`,
  `web/src/routes/+layout.server.ts`, `web/src/routes/+page.svelte`
- Pronto quando: logado, a barra mostra a inicial e o nome (ou o nome de usuário) com o
  menu "Sair"; deslogado, "Entrar com Discord" funciona; `?login=erro` mostra a
  mensagem; testes de renderização no servidor passam.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff810db721c0d2aed94a44

## US-01 — Ambiente e segurança

### T-07 — Variáveis, pilha `app` e README  [ ]
- Cobre: RN-03, CA-01.5, D-09
- Depende de: T-06
- Paralelizável: não
- Arquivos: `.env.example`, `docker-compose.yml`, `README.md`, `scripts/smoke-app.sh`,
  `web/src/lib/auth/*.spec.ts`
- Pronto quando: a pilha `app` sobe com as variáveis do Discord; o README explica o
  aplicativo do Developer Portal e os Redirects; um teste confere que o Client Secret não
  aparece no build do web.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff812cbd5ccdb46338b651

## US-01 a US-05 — Ponta a ponta

### T-08 — Playwright com o Discord falso e CI  [ ]
- Cobre: CA-01.1, CA-01.3, CA-02.1, CA-02.4, CA-03.1, CA-04.1, CA-04.2, CA-05.1, RNF-02,
  RNF-04, RNF-05
- Depende de: T-07
- Paralelizável: não
- Arquivos: `web/playwright.config.ts`, `web/test/e2e/login.spec.ts`,
  `.github/workflows/ci.yml`
- Pronto quando: o Playwright sobe o Discord falso, a API e o web e prova entrar, ver o
  nome, sair, cancelar e voltar para a página de origem; o menu funciona pelo teclado;
  nenhuma requisição sai para fora do servidor; o job da CI roda sem rede externa.
- Commit: —
- Notion: https://app.notion.com/p/3f0d4a3a5eff813d9c3bd81e3bd0826f

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-03, T-08 |
| CA-01.2 | T-03 |
| CA-01.3 | T-05, T-08 |
| CA-01.4 | T-01 |
| CA-01.5 | T-07 |
| CA-02.1 | T-06, T-08 |
| CA-02.2 | T-06 |
| CA-02.3 | T-06 |
| CA-02.4 | T-01, T-08 |
| CA-03.1 | T-06, T-08 |
| CA-03.2 | T-03 |
| CA-04.1 | T-05, T-08 |
| CA-04.2 | T-05, T-08 |
| CA-04.3 | T-05 |
| CA-04.4 | T-04 |
| CA-04.5 | T-02, T-04 |
| CA-04.6 | T-05 |
| CA-05.1 | T-05, T-08 |
| CA-05.2 | T-05 |
| CA-06.1 | T-04 |
| CA-06.2 | T-04, T-05 |
| CA-06.3 | T-03 |
| CA-06.4 | T-03 |
| CA-06.5 | T-05 |

Todos os 24 critérios da spec estão cobertos.

## Descobertas
- 2026-10-05 — O Client Secret do aplicativo do Discord foi colado na conversa. —
  Orientado a gerar um novo no Developer Portal e colocar direto no `.env`.
- 2026-10-05 — A API passou a exigir `DISCORD_CLIENT_ID` e `DISCORD_CLIENT_SECRET` para
  subir (T-04). Para nenhum commit quebrar a pilha `app` nem a CI, as variáveis entraram
  já na T-04 no `.env.example` (fictícias), no serviço `api` do Compose e no job de ponta
  a ponta da CI. A T-07 fica com o web, o README e o teste do secret no build.
