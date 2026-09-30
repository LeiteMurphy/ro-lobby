# Tasks — Fundação

- Spec: `./spec.md` · Design: ADR-01 a ADR-05 em `docs/adr/` (sem `design.md` próprio)
- Notion: [Épico](https://app.notion.com/p/3ead4a3a5eff81ca93e6c2681809464e)
- Branch: `feature/fundacao`

## Decisões de implementação
Detalhes que a spec e os ADRs deixam livres. Nenhum muda regra, escopo ou tecnologia.

- **D-01** — O `openapi.yaml` fica na raiz do repositório, porque é compartilhado por
  `backend/` e `web/` (RN-11, RN-17).
- **D-02** — As ferramentas Go (`sqlc`, `goose`, `oapi-codegen`) ficam fixadas como
  `tool` no `backend/go.mod` e rodam com `go tool`. Assim a versão é a mesma na máquina
  e na CI (RN-04, RN-12). O `golangci-lint` roda pela action oficial na CI e tem a
  instalação descrita no README.
- **D-03** — As migrações rodam por um comando próprio (`go run ./cmd/migrate up|down|status`),
  com os arquivos SQL embutidos. Ele lê só a `DATABASE_URL`, e os testes de
  integração usam o mesmo código (RN-09, CA-03.*).
- **D-04** — A primeira migração é uma linha de base sem tabela de domínio: fixa o
  fuso do banco em UTC (`ALTER DATABASE ... SET timezone`), e o `down` desfaz. Ela dá
  ao `goose` uma versão para registrar e reverter sem antecipar as tabelas das
  próximas specs (RN-08, RN-09).
- **D-05** — A única consulta do `sqlc` é a do fuso da sessão
  (`SELECT current_setting('TimeZone')`), usada no teste do CA-02.5 (RN-08, RN-12).
- **D-06** — Testes de integração do Go usam a build tag `integration`:
  `go test ./...` roda só os unitários, sem banco, e
  `go test -tags=integration ./...` roda os de integração (RN-10, CA-06.7).
- **D-07** — Um único `.env` na raiz serve ao Compose, ao backend e ao web. O backend
  carrega o `.env` se ele existir, sem sobrescrever variáveis já definidas, e o Vite
  lê a mesma pasta (`envDir`). Sem Makefile: os comandos do README funcionam no
  PowerShell (RN-01, RN-02, RN-05).
- **D-08** — Filtro de caminhos da CI dentro de um único workflow (`dorny/paths-filter`),
  com jobs condicionais e um job final `ci-ok` como check obrigatório da `main`. Com
  `paths:` no gatilho, um job que não roda fica pendente e trava o merge; job pulado
  dentro do workflow conta como sucesso (RN-17, RN-20).
- **D-09** — O teste ponta a ponta roda num job próprio da CI (`e2e`), com Postgres
  como serviço, quando o backend, o web ou o contrato mudam (RN-22, CA-07.1).
- **D-10** — A cobertura sai como artefato de cada job e como resumo no próprio job
  (`$GITHUB_STEP_SUMMARY`), visível na aba de checks do PR, sem percentual mínimo
  (RN-21, CA-08.1).

## US-01 — Ambiente local  (P1)

### T-01 — Higiene do repositório  [x]
- Cobre: RN-01, RN-02, RN-04, RN-23, CA-01.5, CA-01.6
- Depende de: —
- Paralelizável: não
- Arquivos: `.gitattributes`, `.gitignore`, `.env.example`, `.nvmrc`
- Pronto quando: `git ls-files --eol` mostra `i/lf` em todos os arquivos de texto; um
  `.env`, um `node_modules/` e uma pasta `coverage/` criados de teste não aparecem no
  `git status`; o `.env.example` está versionado.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff8136a1b7f18400b1db23

### T-02 — PostgreSQL pelo Compose  [x]
- Cobre: RN-03, CA-01.3
- Depende de: T-01
- Paralelizável: não
- Arquivos: `docker-compose.yml`, `.env.example`
- Pronto quando: `docker compose up -d` sobe o Postgres com volume nomeado e
  healthcheck; depois de `docker compose down` e `up` de novo, os dados continuam lá.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff819aa998db996b10f18f

## US-04 — Contrato e código gerado  (P1)

### T-03 — Esqueleto do backend  [x]
- Cobre: RN-01, RN-04, RN-08, D-02, D-07
- Depende de: T-01
- Paralelizável: [P] com T-02
- Arquivos: `backend/go.mod`, `backend/cmd/api/main.go`, `backend/internal/config/`,
  `backend/internal/database/`
- Pronto quando: `go build ./...`, `go vet ./...` e `gofmt -l .` passam; a config
  vem só de variáveis de ambiente, com teste unitário; a conexão do `pgx` define
  `timezone=UTC`.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff81f7a599e6ab53f740dd

### T-04 — Contrato OpenAPI e código Go gerado  [x]
- Cobre: RN-11, RN-12, CA-04.1
- Depende de: T-03
- Paralelizável: não
- Arquivos: `openapi.yaml`, `backend/internal/api/oapi.cfg.yaml`,
  `backend/internal/api/api.gen.go`
- Pronto quando: o `openapi.yaml` descreve `GET /healthz` com as respostas 200 e 503 e
  seus corpos; `go generate ./...` gera a interface do servidor para `net/http` sem
  diferença com o que está commitado; teste do CA-04.1 lê o contrato e confere a rota.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff814887cef0e67b3ffab4

## US-02 — Health check  (P1)

### T-05 — Handler do /healthz  [x]
- Cobre: RN-06, RN-07, RN-10, RN-11, CA-02.1, CA-02.2, CA-02.3, CA-02.4, CA-06.7
- Depende de: T-04
- Paralelizável: não
- Arquivos: `backend/internal/health/`, `backend/cmd/api/main.go`
- Pronto quando: testes unitários com um `pinger` falso passam para os CA-02.1 a 02.4
  (banco ok → 200; erro → 503; ping que passa de 2 s → 503 em menos de 3 s; POST → 405),
  sem conectar em banco.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff81f2a65bff42aadcba7a

## US-03 — Migrações  (P1)

### T-06 — Migrações, sqlc e testes de integração  [x]
- Cobre: RN-08, RN-09, RN-10, RN-12, CA-02.1, CA-02.2, CA-02.5, CA-03.1, CA-03.2,
  CA-03.3, D-03, D-04, D-05, D-06
- Depende de: T-02, T-05
- Paralelizável: não
- Arquivos: `backend/migrations/`, `backend/cmd/migrate/`, `backend/sqlc.yaml`,
  `backend/queries/`, `backend/internal/db/` (gerado), testes `*_integration_test.go`
- Pronto quando: contra o Postgres do Compose, `go test -tags=integration ./...` passa
  para os CA-02.1 (banco real → 200), CA-02.2 (banco inacessível → 503 em até 3 s),
  CA-02.5 (sessão em UTC), CA-03.1 (up em banco vazio), CA-03.2 (down até 0, sem tabelas
  do projeto) e CA-03.3 (todo arquivo tem `-- +goose Up` e `-- +goose Down`).
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff815f8ce3d2f1c8fa5f27

## US-05 — Página de status  (P1)

### T-07 — Esqueleto do web  [x]
- Cobre: RN-04, RN-14, RN-16
- Depende de: T-01
- Paralelizável: [P] com T-03 a T-06
- Arquivos: `web/` (SvelteKit com `adapter-node`, TypeScript estrito, ESLint, Prettier,
  Vitest), `web/package.json` com `engines`
- Pronto quando: `npm run lint`, `npm run check`, `npm run test` e `npm run build`
  passam; o `tsconfig` tem `strict: true`; o adaptador é o `adapter-node`.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff818f84c4ddcccf310b3d

### T-08 — Tipos gerados e página de status  [ ]
- Cobre: RN-13, RN-15, CA-04.4, CA-05.1, CA-05.2, CA-05.3, CA-05.4
- Depende de: T-04, T-07
- Paralelizável: não
- Arquivos: `web/src/lib/api/schema.gen.ts`, `web/src/lib/api/`,
  `web/src/routes/+page.server.ts`, `web/src/routes/+page.svelte`
- Pronto quando: `npm run generate` gera os tipos sem diferença com o commitado; a
  chamada ao `/healthz` usa o tipo gerado; testes Vitest cobrem 200 → "API online",
  503 → "API com problema", falha de rede → "API indisponível" sem erro na tela, e o
  HTML do SSR já traz o texto (CA-05.4).
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff819fb067e8024517f34b

## US-07 — Ponta a ponta  (P2)

### T-09 — Teste Playwright da página de status  [ ]
- Cobre: RN-22, CA-07.1, CA-01.1
- Depende de: T-06, T-08
- Paralelizável: não
- Arquivos: `web/playwright.config.ts`, `web/e2e/status.spec.ts`
- Pronto quando: com banco, backend e web no ar, `npm run test:e2e` encontra
  "API online".
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff81908995c676bb60ac75

## US-06 — CI  (P1)

### T-10 — Workflow do GitHub Actions  [ ]
- Cobre: RN-04, RN-12, RN-17, RN-18, RN-19, RN-20, RN-21, CA-01.4, CA-04.2, CA-04.3,
  CA-06.1 a CA-06.8, CA-08.1, D-08, D-09, D-10
- Depende de: T-06, T-08, T-09
- Paralelizável: não
- Arquivos: `.github/workflows/ci.yml`
- Pronto quando: o PR da fundação roda os jobs `backend`, `web`, `e2e` e `ci-ok` em
  verde; Go e Node vêm de `backend/go.mod` e `.nvmrc`; os relatórios de cobertura
  aparecem como artefato e resumo. Os cenários negativos (CA-04.3, CA-06.1, 06.2, 06.4,
  06.5, 06.8) são provados com PRs de demonstração em rascunho, fechados sem merge, com
  os links anotados aqui.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff81e08a5df1132b05a222

## US-01 — Ambiente local (fechamento)

### T-11 — README e estado do projeto  [ ]
- Cobre: RN-02, RN-05, CA-01.1, CA-01.2, CA-01.4
- Depende de: T-10
- Paralelizável: não
- Arquivos: `README.md`, `CLAUDE.md` (seção "Estado atual"), `.specs/fundacao/research.md`
  (ambiente local atualizado)
- Pronto quando: o README tem o passo a passo no Windows (Go, Node, Docker Desktop,
  subir o banco, migrar, backend, web, testes, troca de portas e proteção da `main`);
  seguido do zero numa pasta nova, chega em "API online"; toda variável lida no código
  e no Compose está no `.env.example`.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff81a9a3c5fe51e57d0079

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-09, T-11 |
| CA-01.2 | T-01, T-11 |
| CA-01.3 | T-02 |
| CA-01.4 | T-10, T-11 |
| CA-01.5 | T-01 |
| CA-01.6 | T-01 |
| CA-02.1 | T-05, T-06 |
| CA-02.2 | T-05, T-06 |
| CA-02.3 | T-05 |
| CA-02.4 | T-05 |
| CA-02.5 | T-06 |
| CA-03.1 | T-06 |
| CA-03.2 | T-06 |
| CA-03.3 | T-06 |
| CA-04.1 | T-04 |
| CA-04.2 | T-10 |
| CA-04.3 | T-10 |
| CA-04.4 | T-08 |
| CA-05.1 | T-08 |
| CA-05.2 | T-08 |
| CA-05.3 | T-08 |
| CA-05.4 | T-08 |
| CA-06.1 | T-10 |
| CA-06.2 | T-10 |
| CA-06.3 | T-10 |
| CA-06.4 | T-10 |
| CA-06.5 | T-10 |
| CA-06.6 | T-10 |
| CA-06.7 | T-05, T-10 |
| CA-06.8 | T-10 |
| CA-07.1 | T-09 |
| CA-08.1 | T-10 |

Todos os 31 critérios da spec estão cobertos.

## Descobertas
- 2026-09-29 — O `research.md` ainda diz que Go e Docker não estão instalados. Hoje há
  Go 1.27.0, Node 24.14.0 e Docker Desktop 4.93.0 (engine 29.8.1, WSL 2). — Atualizar
  na T-11.
- 2026-09-29 — O Smart App Control do Windows 11 bloqueia os executáveis sem
  assinatura que o `go test` e o `go build` geram (evento 3118 do Code Integrity no
  `database.test.exe`). — Decisão do usuário: desligar o SAC. A T-11 documenta o passo
  no README como pré-requisito do setup no Windows.
