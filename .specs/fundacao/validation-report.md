# Relatório de validação — fundacao (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes passaram (backend unitários 16/16 com o banco parado e `DATABASE_URL` vazia · backend com `-tags=integration` 21/21 · Vitest 14/14 · Playwright 2/2) · lint ok (`gofmt -l .` vazio, `golangci-lint` 0 issues, `go vet` com e sem a tag, `npm run lint` e `npm run check` sem erros) · build ok (`go build ./...`, `npm run build` com `adapter-node`) · código gerado sem diferença (`go generate ./...` e `npm run generate`, `git status --porcelain` vazio) · CI do PR #4 (run 36654210727, commit `785966c`) verde nos 5 jobs, em 5m03s

Intervalo validado: `origin/main..feature/fundacao` (16 commits, 67 arquivos). O commit novo desde o ciclo 1 é o `785966c`. Ambiente: Windows 11, Go 1.27.0, Node 24.14.0, PostgreSQL 18 pelo Compose.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 23 | 0 | 0 | 0 |
| Critérios (CA) | 31 | 0 | 0 | 0 |
| Não funcionais | 2 | 1 | 0 | 0 |

A pendência do ciclo 1 (CA-01.6 e RN-23) está corrigida. O ⚠️ que sobra (RNF-03) é de rastreabilidade de nomes de teste, não é P1 e não bloqueia.

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `backend/internal/config/config.go:21-37`, `config/dotenv.go:15-37`, `web/src/routes/+page.server.ts:9`, `web/vite.config.ts:10,26` | `TestLoad_RN01_*` (4), `TestLoadDotEnv_RN01_*` (2), `TestPoolConfig_RN01_RejectsInvalidURL` | `git grep` por senha, token, chave e prefixos conhecidos: só os valores fictícios `ro_lobby_dev` e `ro_lobby_ci`. Nenhum `.env`, `.pem` ou `.key` versionado. |
| RN-02 | ✅ | `.env.example:5-17` | `TestParseDotEnv_RN02_*` (2); levantamento das leituras (CA-01.2) | |
| RN-03 | ✅ | `docker-compose.yml` (só `postgres`, volume nomeado `postgres-data`, healthcheck) | Execução: `stop`/`up` e `down`/`up -d --wait`; volume `ro-lobby_postgres-data` mantido, `migrate status` = `applied` | |
| RN-04 | ✅ | `backend/go.mod:3`, `.nvmrc`, `web/package.json:43-45`, `web/.npmrc`, `ci.yml:81,166,245,250` | `config.spec.ts` "RN-04" | |
| RN-05 | ✅ | `README.md:12-128` | Segui o README num clone novo (ver CA-01.1) | |
| RN-06 | ✅ | `backend/internal/health/health.go:12,34-45` | `TestHealthz_CA02_1_*`, `TestHealthzIntegration_CA02_1_*`; `curl` na API real: 200 `{"database":"ok","status":"ok"}` | |
| RN-07 | ✅ | `health.go:35-40,49-62` | `TestHealthz_CA02_2_*`, `TestHealthz_CA02_3_*`, `TestHealthzIntegration_CA02_2_*`; execução: container pausado → 503 em 2,001 s; parado → 503 em 0,005 s | |
| RN-08 | ✅ | `backend/internal/database/database.go:17`; `migrations/00001_baseline.sql:6-10` | `TestPoolConfig_RN08_ForcesUTC`, `TestConnection_CA02_5_SessionTimeZoneIsUTC` | |
| RN-09 | ✅ | `migrations/00001_baseline.sql`, `migrations/embed.go`, `internal/migrate/migrate.go`, `cmd/migrate/main.go` | `TestMigrations_CA03_1_*`, `CA03_2_*`, `CA03_3_*` | |
| RN-10 | ✅ | `//go:build integration` em `testdb.go` e `*_integration_test.go`; `server_test.go:16-18` (`fakePinger`) | `go test ./...` com o Postgres parado: 16/16; `-tags=integration`: 21/21 | |
| RN-11 | ✅ | `openapi.yaml:10-48`; `internal/server/server.go:12-15` | `TestContract_CA04_1_HealthzIsDescribed`, `TestHealthz_CA02_4_MethodNotAllowed` | |
| RN-12 | ✅ | `internal/api/generate.go:5`, `internal/db/generate.go:5`, `web/package.json:17`, `ci.yml:105-111,179-185` | Geração sem diferença; mutação do `summary` no `openapi.yaml` refeita neste ciclo → `api.gen.go` e `schema.gen.ts` modificados | |
| RN-13 | ✅ | `web/src/lib/api/health.ts:1-6,40` | `health.spec.ts` "CA-04.4" (`expectTypeOf` contra `components['schemas']['Health']`) | |
| RN-14 | ✅ | `web/vite.config.ts:3,21` | `config.spec.ts` "RN-14"; `npm run build` → "Using @sveltejs/adapter-node" | |
| RN-15 | ✅ | `+page.server.ts:8-10`, `health.ts:24-46`, `+page.svelte:14-16` | `health.spec.ts`, `page.server.spec.ts`, `page.spec.ts`; execução do build com API saudável, com banco pausado/parado e com API parada | Outros códigos (500, 404) viram "API com problema", conforme a descoberta de `tasks.md`. |
| RN-16 | ✅ | `web/tsconfig.json:12` | `config.spec.ts` "RN-16 / CA-06.8"; mutação com `any` implícito → `npm run check` sai com 1 | |
| RN-17 | ✅ | `ci.yml:6-10,36-49,54,154,223` | Runs 36652806330 (só backend: Web `skipped`), 36652813215 (só web: Backend `skipped`), 36654210727 (muda a CI: todos rodaram) | Caso "PR só de documentação" conferido só pela leitura do filtro. |
| RN-18 | ✅ | `ci.yml:85-126` | Run 36654210727: todos os passos do Backend em `success`; o log da integração mostra `internal/migrate` com 80% (pacote que só tem teste com a tag) | |
| RN-19 | ✅ | `ci.yml:170-198` | Run 36654210727: todos os passos do Web em `success` | |
| RN-20 | ✅ | `ci.yml:277-294` (`ci-ok` falha com qualquer resultado fora de `success`/`skipped`) | Runs 36652806330 e 36652813215: `CI ok` = `failure` com job cancelado; mutações locais de gofmt, golangci-lint, Prettier, Vitest, `go test` e tipos saíram com 1 | O bloqueio do merge depende da proteção da `main` (ver Observações). |
| RN-21 | ✅ | `ci.yml:129-149,201-217`; `vite.config.ts:32-38` sem `thresholds` | Run 36654210727: artefatos `coverage-backend` (11 KB) e `coverage-web` (25 KB) | |
| RN-22 | ✅ | `web/playwright.config.ts`, `web/test/e2e/status.spec.ts:4-8`, `ci.yml:219-275` | `npm run test:e2e` local 2/2; "Ponta a ponta" verde no run 36654210727 | |
| RN-23 | ✅ | `.gitattributes:2`; `.gitignore:1-26` (agora com `coverage*.out` e `coverage*.html`) | `git ls-files --eol`: 74 `i/lf`, 1 `i/none` (`favicon.svg`, sem quebra de linha), nenhum CRLF; clone novo sem nenhum `w/crlf`; `git check-ignore` (ver CA-01.6) | Corrigido no `785966c`. |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | `README.md`, `docker-compose.yml`, `cmd/migrate`, `cmd/api`, `web/` | Execução num `git clone` novo no scratchpad: `Copy-Item .env.example .env` → `docker compose up -d --wait` → `migrate up` → `npm ci` → API com `/healthz` 200 → `vite dev` na 5173 com "API online"; `npm run test:e2e` 2/2 | O Compose do clone reusa o mesmo volume (o `name: ro-lobby` é fixo), então o `migrate up` respondeu "nada a fazer". A aplicação em banco vazio está provada pelo CA-03.1. |
| CA-01.2 | ✅ | `.env.example` | `git grep`: backend lê `DATABASE_URL` e `API_PORT`; web lê `API_BASE_URL` e `WEB_PORT`; Compose lê `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` e `POSTGRES_PORT`. Todas estão no `.env.example` com valor fictício | O Playwright também lê `CI`, definido pelo GitHub Actions. Não é configuração do projeto. |
| CA-01.3 | ✅ | `docker-compose.yml:15-16,23-24` | Execução: `docker compose stop`/`up` e `down`/`up -d --wait`; `migrate status` = `applied 00001_baseline.sql` e `/healthz` 200 depois | |
| CA-01.4 | ✅ | `go.mod:3`, `.nvmrc`, `ci.yml:81,166,245,250` | `config.spec.ts` "RN-04"; a CI lê as versões dos mesmos arquivos | |
| CA-01.5 | ✅ | `.gitattributes:2` | `git ls-files --eol` sem CRLF; `gofmt -l .` vazio no Windows; passo "Quebra de linha LF" verde | |
| CA-01.6 | ✅ | `.gitignore:2-4,7,10-13,16-21` | Gerei os artefatos com os comandos da CI (`coverage-unit.out`, `coverage.out`, `coverage.html`, `bin/`), mais `.env`, `.env.local`, `node_modules/` na raiz, `web/build`, `web/coverage`, `web/test-results`: `git status --porcelain` vazio, todos em `!!` no `--ignored`; `.env.example` continua em `git ls-files` | Corrigido no `785966c`. |
| CA-02.1 | ✅ | `health.go:41-44` | `TestHealthz_CA02_1_*` (status, `Content-Type`, corpo exato), `TestHealthzIntegration_CA02_1_*`; `curl` | |
| CA-02.2 | ✅ | `health.go:35-40`; `database.go:21-30` | `TestHealthz_CA02_2_*`, `TestHealthzIntegration_CA02_2_*` (corpo e prazo ≤ 3 s); `docker compose stop` → 503 em 0,005 s | |
| CA-02.3 | ✅ | `health.go:49-62` | `TestHealthz_CA02_3_SlowDatabase` (Pinger que ignora o contexto; 503, corpo e tempo entre 2 s e 3 s); `docker compose pause` → 503 em 2,001 s | |
| CA-02.4 | ✅ | rota gerada em `api.gen.go` (`GET /healthz` no `ServeMux`) | `TestHealthz_CA02_4_MethodNotAllowed` (405 e o banco não é consultado); `curl -X POST` → 405 | |
| CA-02.5 | ✅ | `database.go:17`; `queries/session.sql`; `internal/db/session.sql.go` | `TestConnection_CA02_5_SessionTimeZoneIsUTC` (banco em `America/Sao_Paulo`; `SHOW TimeZone` e a query do sqlc devolvem `UTC`) | |
| CA-03.1 | ✅ | `internal/migrate/migrate.go`, `cmd/migrate/main.go:58-59` | `TestMigrations_CA03_1_UpOnEmptyDatabase` (banco novo, sem tabelas; versão = última) | |
| CA-03.2 | ✅ | `cmd/migrate/main.go:66-67`; Down de `00001_baseline.sql` | `TestMigrations_CA03_2_DownToZero` (versão 0, sem tabela além de `goose_db_version`, configuração de fuso desfeita) | |
| CA-03.3 | ✅ | `migrations/embed.go` | `TestMigrations_CA03_3_AllHaveUpAndDown` (Up e Down presentes, na ordem, Down não vazio) | |
| CA-04.1 | ✅ | `openapi.yaml:10-48` | `TestContract_CA04_1_HealthzIsDescribed` (valida o documento, 200 e 503, campos obrigatórios, corpos contra o schema) | |
| CA-04.2 | ✅ | `ci.yml:105-111,179-185` | Geração local sem diferença; passos "Código gerado em dia" verdes no run 36654210727 | |
| CA-04.3 | ✅ | idem | Mutação refeita neste ciclo: `summary` alterado sem gerar → `git status --porcelain -- .` lista `api.gen.go` no backend e `schema.gen.ts` no web, e o passo sairia com 1; `ci-ok` falha com qualquer job em falha | Prova local, com os mesmos comandos da CI. O bloqueio do merge depende da proteção da `main`. |
| CA-04.4 | ✅ | `health.ts:1-6` | `health.spec.ts` "CA-04.4" | |
| CA-05.1 | ✅ | `health.ts:39-41`, `+page.svelte:15` | `health.spec.ts`, `page.server.spec.ts`, `page.spec.ts` "CA-05.1"; build real e `vite dev` com "API online" | |
| CA-05.2 | ✅ | `health.ts:43-45` | `health.spec.ts` e `page.spec.ts` "CA-05.2"; build real com o banco pausado e com o banco parado → "API com problema" | |
| CA-05.3 | ✅ | `health.ts:30-37` | `health.spec.ts` (rede recusada e prazo), `page.server.spec.ts` e `page.spec.ts` "CA-05.3"; build real com a API parada → "API indisponível" | A única ocorrência de "error" no HTML é o `error: null` do script de hidratação. Não aparece na tela. |
| CA-05.4 | ✅ | `+page.server.ts` | `status.spec.ts` "CA-05.4" (`request.get('/')` sem navegador); `curl` no build | |
| CA-06.1 | ✅ | `ci.yml:43-49,54,154` | Run 36652806330 (PR #5, só backend): Web `skipped` | Run cancelado na fila, mas a decisão do filtro ficou registrada. |
| CA-06.2 | ✅ | idem | Run 36652813215 (PR #6, só web): Backend `skipped`, Web `success` | |
| CA-06.3 | ✅ | `ci.yml:40-42` | Run 36652820138 (PR #7, só `openapi.yaml`, registrado no ciclo 1): os dois iniciaram; run 36654210727 (muda a CI): os dois rodaram | |
| CA-06.4 | ✅ | `ci.yml:91-102,187-188` | Run 36652825897 (PR #8): gofmt e ESLint/Prettier em `failure`; mutações refeitas neste ciclo: gofmt lista o arquivo, `golangci-lint` sai com 1 (errcheck), Prettier sai com 1 | |
| CA-06.5 | ✅ | `ci.yml:113-114,194-195` | Mutações refeitas: `t.Fatal` → `go test` sai com 1; `expect(1).toBe(2)` → `vitest` sai com 1 | Prova local. |
| CA-06.6 | ✅ | `ci.yml:60-74,122-123` | Run 36654210727: "Testes de integração" verde contra o serviço `postgres` | |
| CA-06.7 | ✅ | build tags; `fakePinger` | `go test ./...` com o container parado e `DATABASE_URL=` vazia: 16/16 | |
| CA-06.8 | ✅ | `tsconfig.json:12`; `ci.yml:191-192` | Mutação refeita: `export function f(x)` → `npm run check` sai com 1 | O ESLint não pega esse caso (saiu com 0), mas o passo `svelte-check` do mesmo job pega. |
| CA-07.1 | ✅ | `playwright.config.ts`, `status.spec.ts:4-8` | `npm run test:e2e` local 2/2; "Ponta a ponta" verde no run 36654210727 | |
| CA-08.1 | ✅ | `ci.yml:129-149,201-217` | Run 36654210727: artefatos `coverage-backend` e `coverage-web`; resumos em `$GITHUB_STEP_SUMMARY`; nenhum limite configurado | |

### Não funcionais
| ID | Veredito | Evidência | Observação |
|---|---|---|---|
| RNF-01 | ✅ | Ver RN-08 e CA-02.5 | |
| RNF-02 | ✅ | `timeout-minutes: 10` nos jobs; run 36654210727 (backend, web e ponta a ponta) em 5m03s; o job Backend levou 4m48s | |
| RNF-03 | ⚠️ | Todos os testes citam agora um ID da spec (os quatro apontados no ciclo 1 foram renomeados) | Alguns citam só a RN, não o CA que cobrem: `config.spec.ts` "RN-04" cobre o CA-01.4 e "RN-14" faz parte do CA-05.4. `TestLoad_RN01_*` e `TestParseDotEnv_RN02_*` testam regras sem CA próprio, então citar a RN é razoável. Não bloqueia. |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. O commit novo (`785966c`) só mexe em `.gitignore`, nomes e comentários de teste, o tipo `HealthBody | null` em `health.ts`, `tasks.md` e o relatório do ciclo 1. Tudo se liga à RN-23, ao CA-01.6, à RNF-03 ou às observações do ciclo 1. O restante do diff segue como no ciclo 1: nada toca "Fora de escopo" (sem login, tabelas de domínio, imagens Docker do backend ou do web, hooks de pre-commit ou percentual mínimo de cobertura). Os extras pequenos (`migrate down`/`reset`/`status`, `.nvmrc` no filtro do web, `robots.txt` e favicon do scaffold) ficam dentro das US.

## Observações (não bloqueantes)
- **Proteção da `main` indisponível.** `gh api .../branches/main/protection` continua respondendo HTTP 403 ("Upgrade to GitHub Pro or make this repository public"), e o repositório é privado. O "merge fica bloqueado" da RN-20 e do CA-04.3 vale hoje só como regra de conduta. A spec deixa essa configuração com o usuário (seção 10), mas vale registrar no README que o passo de proteção exige repositório público ou plano pago.
- RNF-03: acrescentar o CA-01.4 ao nome do teste "RN-04" em `web/src/config.spec.ts` deixa a rastreabilidade completa.
- As tasks estão todas com `Commit: —`. Os commits citam os IDs, mas o hash de cada task não foi anotado.
- O caso de borda "PR só de documentação → nenhum job de código roda" foi conferido só pela leitura do filtro. Nenhum run exercitou esse caso.
- O Compose usa `name: ro-lobby` fixo, então dois clones na mesma máquina dividem o mesmo container e volume. Não afeta a spec.
- O `adapter-node` em produção lê `PORT` e `HOST`, não `WEB_PORT`. Vale lembrar na ADR de hospedagem.
- A mensagem para códigos fora de 200 e 503 ("API com problema") segue a descoberta de `tasks.md`, que ainda pede confirmação do usuário.
