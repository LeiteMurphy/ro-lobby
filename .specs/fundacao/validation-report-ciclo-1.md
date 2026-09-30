# Relatório de validação — fundacao (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** testes passaram (backend unitários 19/19 · backend com `-tags=integration` 24/24 · Vitest 14/14 · Playwright 2/2) · lint ok (`gofmt -l .` vazio, `golangci-lint` 0 issues, `npm run lint` e `npm run check` sem erros) · build ok (`go build ./...`, `npm run build`) · código gerado sem diferença (`go generate ./...` e `npm run generate`, `git status --porcelain` vazio) · CI do PR #4 (run 36653194441, commit `0b05c00`) verde nos 5 jobs, em 4m21s

Intervalo validado: `origin/main..feature/fundacao` (15 commits, 66 arquivos). Ambiente: Windows 11, Go 1.27.0, Node 24.14.0, PostgreSQL 18 pelo Compose.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 22 | 1 | 0 | 0 |
| Critérios (CA) | 30 | 1 | 0 | 0 |
| Não funcionais | 2 | 1 | 0 | 0 |

O único motivo da reprovação é o CA-01.6 (US-01, P1), junto com a RN-23: o `.gitignore` não cobre dois artefatos de cobertura que a própria CI gera no backend. A correção é pequena.

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `backend/internal/config/config.go:21-37`, `config/dotenv.go:15-37` (não sobrescreve variável já definida), `web/src/routes/+page.server.ts:9`, `web/vite.config.ts:10,26` | `TestLoad_RN01_*` (4 testes), `TestLoadDotEnv_RN01_DoesNotOverrideEnvironment` | Varredura de segredos (`git grep` por password/secret/token/chaves conhecidas): só os valores fictícios `ro_lobby_dev` e `ro_lobby_ci`. |
| RN-02 | ✅ | `.env.example:5-17` | `TestParseDotEnv_RN02_EnvExampleFormat`; levantamento manual das leituras (ver CA-01.2) | |
| RN-03 | ✅ | `docker-compose.yml:4-24` (só `postgres`, volume nomeado `postgres-data`, healthcheck) | Execução: `docker compose down` e `up -d --wait`, `migrate status` continuou `applied` | |
| RN-04 | ✅ | `backend/go.mod:3` (`go 1.27.0`), `.nvmrc` (`24.14.0`), `web/package.json:43-45` (`engines`), `web/.npmrc` (`engine-strict`), `ci.yml:79-81,164-166,243-251` (`go-version-file`, `node-version-file`) | `config.spec.ts` "RN-04: a versão do Node em engines bate com o .nvmrc" | `golangci-lint` fixado em v2.14.0 na CI e no README. |
| RN-05 | ✅ | `README.md:12-128` (Go, Node, Docker Desktop, WSL, Smart App Control, `.env`, banco, migração, backend, web, testes, troca de portas) | Execução dos passos (ver CA-01.1) | |
| RN-06 | ✅ | `backend/internal/health/health.go:12,34-45` | `TestHealthz_CA02_1_DatabaseAvailable`, `TestHealthzIntegration_CA02_1_DatabaseAvailable`; `curl` na API real: 200 `{"database":"ok","status":"ok"}` | |
| RN-07 | ✅ | `health.go:35-40,49-62` (prazo aplicado mesmo se o `Pinger` ignorar o contexto) | `TestHealthz_CA02_2_*`, `TestHealthz_CA02_3_SlowDatabase`, `TestHealthzIntegration_CA02_2_DatabaseDown`; execução com banco parado e com o container pausado | |
| RN-08 | ✅ | `backend/internal/database/database.go:17`; migração `00001_baseline.sql:6-10` | `TestPoolConfig_RN08_ForcesUTC`, `TestConnection_CA02_5_SessionTimeZoneIsUTC` (banco com fuso `America/Sao_Paulo`, sessão continua UTC) | |
| RN-09 | ✅ | `backend/migrations/00001_baseline.sql` (Up e Down), `migrations/embed.go`, `cmd/migrate/main.go`, `internal/migrate/migrate.go` | `TestMigrations_CA03_1_*`, `TestMigrations_CA03_2_*`, `TestMigrations_CA03_3_*`; execução de `migrate reset`, `status` e `up` sem erro | |
| RN-10 | ✅ | build tag `//go:build integration` em `testdb.go`, `*_integration_test.go`; `server_test.go:16-18` (`fakePinger`) | `go test ./...` com o banco parado e `DATABASE_URL` vazia: todos passaram; `-tags=integration` contra o Postgres do Compose: 24/24 | |
| RN-11 | ✅ | `openapi.yaml:10-48`; `internal/server/server.go:12-15` monta as rotas pela interface gerada | `TestContract_CA04_1_HealthzIsDescribed`, `TestHealthz_CA02_4_MethodNotAllowed` | |
| RN-12 | ✅ | `internal/api/generate.go:5`, `internal/db/generate.go:5`, `web/package.json:17`; `ci.yml:105-111,179-185` | Execução: geração sem diferença; mutação do `summary` no `openapi.yaml` deixou `api.gen.go` e `schema.gen.ts` modificados (os passos da CI sairiam com 1) | |
| RN-13 | ✅ | `web/src/lib/api/health.ts:1-6,40` | `health.spec.ts` "CA-04.4" (`expectTypeOf` contra `components['schemas']['Health']`) | |
| RN-14 | ✅ | `web/vite.config.ts:3,21` (`adapter-node`) | `config.spec.ts` "RN-14"; `npm run build` mostra "Using @sveltejs/adapter-node" | |
| RN-15 | ✅ | `+page.server.ts:8-10`, `health.ts:11-15,24-46`, `+page.svelte:14-16` | `health.spec.ts`, `page.server.spec.ts`, `page.spec.ts`; execução do build com API saudável, degradada e parada | 200 com corpo inesperado e outros códigos (500, 404) viram "API com problema", conforme a descoberta registrada em `tasks.md`. |
| RN-16 | ✅ | `web/tsconfig.json:12` | `config.spec.ts` "RN-16 / CA-06.8"; mutação com `any` implícito fez `npm run check` sair com 1 | |
| RN-17 | ✅ | `ci.yml:6-10,36-49,54,154,223` | CI real dos PRs de demonstração: #5 (só backend) → Web `skipped`; #6 (só web) → Backend `skipped`; #7 (só `openapi.yaml`) → os dois iniciaram; PR #4 (muda a CI) → os dois rodaram | Os PRs #5 a #10 tinham base `feature/fundacao`, com o mesmo workflow. O caso "PR só de documentação" foi conferido só pela leitura do filtro. |
| RN-18 | ✅ | `ci.yml:85-126` (LF, gofmt, golangci-lint, código gerado, unitários, migrações, integração com serviço `postgres:18-alpine`, build) | Run 36653194441: todos os passos em `success`; o log da integração mostra `internal/migrate` (pacote que só tem teste com a tag) | |
| RN-19 | ✅ | `ci.yml:170-198` (npm ci, LF, tipos gerados, ESLint e Prettier, svelte-check, Vitest com cobertura, build) | Run 36653194441, job Web em `success` | |
| RN-20 | ✅ | `ci.yml` usa bash com saída em erro; `ci-ok` (`ci.yml:277-294`) falha com qualquer resultado diferente de `success` ou `skipped` | Run 36652825897 (PR #8): `gofmt` e `ESLint e Prettier` em `failure`, `CI ok` em `failure`; mutações locais de teste e de tipos saíram com 1 | O bloqueio do merge depende da proteção da `main`, que a spec (seção 10) deixa com o usuário. Ver "Observações": a API do GitHub responde 403 para proteção e rulesets neste repositório. |
| RN-21 | ✅ | `ci.yml:129-149,201-217`; `web/vite.config.ts:32-38` (sem `thresholds`) | Run 36653194441: artefatos `coverage-backend` e `coverage-web` publicados; passos de resumo em `success` | |
| RN-22 | ✅ | `web/playwright.config.ts`, `web/test/e2e/status.spec.ts:4-8`; job `e2e` em `ci.yml:219-275` | `npm run test:e2e` local: 2/2; job "Ponta a ponta" verde no PR #4 | |
| RN-23 | ⚠️ | `.gitattributes:2` (`* text=auto eol=lf`); `.gitignore:1-20` | `git ls-files --eol`: 73 `i/lf`, 1 `i/none`, nenhum CRLF; passo "Quebra de linha LF" da CI | O `.gitignore` não cobre `backend/coverage-unit.out` nem `backend/coverage.html`, que a CI gera (`ci.yml:114,139`). Ver CA-01.6. |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | `README.md`, `docker-compose.yml`, `cmd/migrate`, `cmd/api`, `web/` | Execução: `docker compose up -d --wait` → `migrate up` → API com `GET /healthz` 200 → build do web mostrou "API online"; `npm run test:e2e` passou | Não refiz o clone do zero numa pasta nova; segui os comandos no clone atual. |
| CA-01.2 | ✅ | `.env.example` | Levantamento com `git grep`: backend lê `DATABASE_URL` e `API_PORT`; web lê `API_BASE_URL` e `WEB_PORT`; Compose lê `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` e `POSTGRES_PORT`. Todas estão no `.env.example`, com valores fictícios. | O Playwright também lê `CI`, que o GitHub Actions define. O `adapter-node` lê `PORT` e `HOST` em produção. Nenhuma das duas é configuração do projeto. |
| CA-01.3 | ✅ | `docker-compose.yml:15-16,23-24` | Execução: `docker compose down`, volume `ro-lobby_postgres-data` mantido, `up -d --wait`, `migrate status` = `applied 00001_baseline.sql` | |
| CA-01.4 | ✅ | `go.mod:3`, `.nvmrc`, `ci.yml:81,166,245,250` | `config.spec.ts` "RN-04"; a CI lê as versões dos mesmos arquivos | |
| CA-01.5 | ✅ | `.gitattributes:2` | `git ls-files --eol` sem nenhum `i/crlf`; `gofmt -l .` vazio no Windows; passo "Quebra de linha LF" verde na CI | |
| CA-01.6 | ⚠️ | `.gitignore:1-20`, `web/.gitignore` | `git check-ignore`: `.env`, `node_modules/`, `web/build/`, `web/.svelte-kit/`, `backend/bin/`, `web/coverage/`, `backend/coverage.out`, `web/test-results/` e `web/playwright-report/` são ignorados; `.env.example` continua versionado | **Falha:** rodar localmente o comando de cobertura da CI (`go test -coverprofile=coverage-unit.out ./...` e `go tool cover -html=... -o coverage.html`) deixa `?? backend/coverage-unit.out` e `?? backend/coverage.html` no `git status`. O Then pede que nenhum artefato de cobertura apareça como arquivo novo. |
| CA-02.1 | ✅ | `health.go:41-44` | `TestHealthz_CA02_1_DatabaseAvailable` (status, `Content-Type` e corpo exato), `TestHealthzIntegration_CA02_1_DatabaseAvailable`; `curl` na API real | |
| CA-02.2 | ✅ | `health.go:35-40`; `database.go:21-30` (a API sobe com o banco fora do ar) | `TestHealthz_CA02_2_DatabaseDown`, `TestHealthzIntegration_CA02_2_DatabaseDown` (corpo e prazo ≤ 3 s); execução com `docker compose stop`: 503 em 0,008 s | |
| CA-02.3 | ✅ | `health.go:49-62` | `TestHealthz_CA02_3_SlowDatabase` (Pinger que ignora o contexto; confere 503, corpo e tempo entre 2 s e 3 s); execução com `docker compose pause`: 503 em 2,002 s | |
| CA-02.4 | ✅ | rota gerada `api.gen.go` com `GET /healthz` no `ServeMux` | `TestHealthz_CA02_4_MethodNotAllowed` (405, e o banco não é consultado); execução: POST e PUT → 405 | |
| CA-02.5 | ✅ | `database.go:17`; `queries/session.sql`; `internal/db/session.sql.go` | `TestConnection_CA02_5_SessionTimeZoneIsUTC` (`SHOW TimeZone` e a query do sqlc devolvem `UTC`, mesmo com o banco em outro fuso) | |
| CA-03.1 | ✅ | `internal/migrate/migrate.go`, `cmd/migrate/main.go:58-59` | `TestMigrations_CA03_1_UpOnEmptyDatabase` (banco novo e vazio, versão = última migração); `migrate up` na CI e local | |
| CA-03.2 | ✅ | `cmd/migrate/main.go:66-67`; Down de `00001_baseline.sql` | `TestMigrations_CA03_2_DownToZero` (versão 0, nenhuma tabela além de `goose_db_version`, configuração de fuso desfeita); `migrate reset` local sem erro | |
| CA-03.3 | ✅ | `migrations/embed.go` | `TestMigrations_CA03_3_AllHaveUpAndDown` (Up e Down presentes, na ordem certa, Down não vazio) | |
| CA-04.1 | ✅ | `openapi.yaml:10-48` | `TestContract_CA04_1_HealthzIsDescribed` (valida o documento, as respostas 200 e 503, os campos obrigatórios e os corpos contra o schema) | |
| CA-04.2 | ✅ | `ci.yml:105-111,179-185` | Execução local sem diferença; passos "Código gerado em dia" verdes no PR #4 | |
| CA-04.3 | ✅ | idem | Mutação local refeita nesta validação: `api.gen.go` e `schema.gen.ts` ficaram modificados, então `git status --porcelain -- .` não vem vazio e o passo sai com 1; `CI ok` falha em qualquer job com falha | A CI do PR #7 foi cancelada antes desses passos. A prova é local, com os mesmos comandos. O bloqueio do merge depende da proteção da `main` (ver RN-20). |
| CA-04.4 | ✅ | `health.ts:1-6` | `health.spec.ts` "CA-04.4" (igualdade de tipos com `expectTypeOf`, import de `./schema.gen`, cabeçalho de arquivo gerado) | |
| CA-05.1 | ✅ | `health.ts:39-41`, `+page.svelte:15` | `health.spec.ts` "CA-05.1", `page.server.spec.ts` "CA-05.1", `page.spec.ts` "CA-05.1 / CA-05.4" | |
| CA-05.2 | ✅ | `health.ts:43-45` | `health.spec.ts` "CA-05.2", `page.spec.ts` "CA-05.2"; build real com o banco parado mostrou "API com problema" | |
| CA-05.3 | ✅ | `health.ts:30-37` (nunca lança; prazo de 3 s) | `health.spec.ts` (rede recusada e prazo esgotado), `page.server.spec.ts` "CA-05.3", `page.spec.ts` "CA-05.3"; build real com `API_BASE_URL` numa porta sem nada: HTTP 200 com "API indisponível" | A única ocorrência de "error" no HTML é o `error: null` do script de hidratação do SvelteKit. Não aparece na tela. |
| CA-05.4 | ✅ | `+page.server.ts` (load no servidor) | `status.spec.ts` "CA-05.4" (`request.get('/')` sem navegador, o HTML contém "API online"); `curl` no build também | |
| CA-06.1 | ✅ | `ci.yml:43-49,54,154` | Run 36652806330 (PR #5, só `backend/internal/health/health.go`): Backend iniciou, Web `skipped` | O run foi cancelado na fila, mas a decisão do filtro ficou registrada. `tasks.md` ainda diz que este CA não foi provado. |
| CA-06.2 | ✅ | idem | Run 36652813215 (PR #6, só `web/src/lib/api/health.ts`): Web `success`, Backend `skipped` | Mesma nota do CA-06.1. |
| CA-06.3 | ✅ | `ci.yml:40-42` | Run 36652820138 (PR #7, só `openapi.yaml`): Backend e Web iniciaram; PR #4 (muda a CI): os dois rodaram | |
| CA-06.4 | ✅ | `ci.yml:91-102,187-188` | Run 36652825897 (PR #8): passos `gofmt` e `ESLint e Prettier` em `failure`; `golangci-lint` provado localmente (tabela em `tasks.md`) | |
| CA-06.5 | ✅ | `ci.yml:113-114,194-195` | Mutações refeitas nesta validação: teste Go com `t.Fatal` → `go test` sai com 1; teste Vitest com `expect(1).toBe(2)` → `vitest` sai com 1 | Prova local. A CI do PR #9 foi cancelada. |
| CA-06.6 | ✅ | `ci.yml:60-74,122-123` | Run 36653194441: "Testes de integração" verde contra o serviço `postgres`, com `internal/migrate` executado | |
| CA-06.7 | ✅ | build tags; `fakePinger` | `go test ./...` com o container parado e `DATABASE_URL` vazia: todos passaram | |
| CA-06.8 | ✅ | `tsconfig.json:12`; `ci.yml:191-192` | Mutação refeita nesta validação: `export function f(x)` → `npm run check` sai com 1 | Prova local. A CI do PR #10 foi cancelada. |
| CA-07.1 | ✅ | `playwright.config.ts`, `status.spec.ts:4-8` | `npm run test:e2e` local 2/2; job "Ponta a ponta" verde no PR #4 | |
| CA-08.1 | ✅ | `ci.yml:129-149,201-217` | Run 36653194441: artefatos `coverage-backend` (11 KB) e `coverage-web` (25 KB); resumos em `$GITHUB_STEP_SUMMARY`; nenhum limite configurado | |

### Não funcionais
| ID | Veredito | Evidência | Observação |
|---|---|---|---|
| RNF-01 | ✅ | Ver RN-08 e CA-02.5 | |
| RNF-02 | ✅ | `timeout-minutes: 10` em cada job; run do PR #4 (as duas pastas mais o ponta a ponta) terminou em 4m21s | |
| RNF-03 | ⚠️ | A maioria dos testes cita CA ou RN no nome | Quatro testes não citam nenhum ID: `TestParseDotEnv_RejectsLineWithoutEquals`, `TestLoadDotEnv_MissingFileIsNotAnError`, `TestPoolConfig_RejectsInvalidURL` e o Vitest "chama GET /healthz na URL base configurada". Outros citam só RN, não CA (`TestLoad_RN01_*`, `TestPoolConfig_RN08_*`, `config.spec.ts` RN-14 e RN-04). |

## Pendências para correção
1. [CA-01.6, RN-23] Acrescentar ao `.gitignore` os artefatos de cobertura do backend que a CI gera: `backend/coverage-unit.out` e `backend/coverage.html` (`ci.yml:114,139`). Um padrão como `coverage*.out` e `coverage.html` resolve. O Then exige que, com os artefatos de build e de cobertura presentes, nenhum apareça no `git status`. Hoje os dois aparecem como `??`.

## Scope creep
- Nenhum. Todos os arquivos do diff se ligam a um ID ou a uma decisão D-01 a D-10. Nada toca os itens de "Fora de escopo": não há login, tabelas de domínio, imagens Docker do backend ou do web, hooks de pre-commit nem percentual mínimo de cobertura. Os extras são pequenos e ficam dentro das US: `migrate down`, `reset` e `status` (US-03), `.nvmrc` no filtro do web (RN-04), `robots.txt` e favicon do scaffold do SvelteKit.

## Observações (não bloqueantes)
- **Proteção da `main` indisponível no plano atual.** `gh api repos/LeiteMurphy/ro-lobby/branches/main/protection` e `.../rulesets` respondem HTTP 403: "Upgrade to GitHub Pro or make this repository public". Então o passo documentado no `README.md:157-164` não pode ser aplicado hoje. Enquanto isso, o "merge fica bloqueado" da RN-20 e do CA-04.3 fica só como regra de conduta. A spec deixa essa configuração com o usuário (seção 10), por isso não entra como pendência do código. Para a garantia valer, o repositório precisa ficar público ou o plano precisa mudar. Vale registrar isso no README.
- `tasks.md` ainda diz que o CA-06.1 e o CA-06.2 ficam "para o primeiro PR real de uma pasta só", mas os runs 36652806330 e 36652813215 já mostram o filtro funcionando (Web e Backend `skipped`). Vale anotar esses runs como evidência. O CA-06.4 também tem prova na CI real (run 36652825897), não só local.
- D-02 cita o `goose` como `tool` no `go.mod`, mas ele entra como biblioteca, usada pelo `cmd/migrate` (D-03). O `tool (...)` só tem `oapi-codegen` e `sqlc`. O comportamento está certo; só o texto da D-02 está impreciso.
- `health.ts:40` declara `body: HealthBody`, mas o valor pode ser `null` por causa do `.catch(() => null)`. Funciona por causa do `?.`, mas o tipo esconde o `null`. `HealthBody | null` fica mais honesto.
- O `adapter-node` em produção lê `PORT` e `HOST`, e não `WEB_PORT`. Hoje não afeta nada, porque a hospedagem está fora de escopo. Vale lembrar na ADR de hospedagem.
- O caso de borda "PR só de documentação → nenhum job de código roda" foi conferido só pela leitura do filtro. Nenhum run exercitou esse caso.
- A mensagem para códigos fora de 200 e 503 ("API com problema") segue a descoberta de `tasks.md`, que pede confirmação do usuário.
