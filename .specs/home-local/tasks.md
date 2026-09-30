# Tasks — Hospedagem local e Home

- Spec: `./spec.md` · ADR: `docs/adr/0006-hospedagem-local.md`
- Notion: [Épico](https://app.notion.com/p/3ebd4a3a5eff8144aac0cdabbdc91913)
- Branch: `feature/home-local`

## Decisões de implementação
- **D-01** — Uma imagem do backend, com três binários (`api`, `migrate` e `healthcheck`), sobre
  `gcr.io/distroless/static-debian12:nonroot`. O serviço de migração do Compose usa a
  mesma imagem com outro comando. A imagem não tem shell nem `curl`, então o `healthcheck`
  (GET `/healthz`, sai com 0 só em 200) é o binário que o Compose usa para saber se a API
  está pronta.
- **D-02** — Imagem do web em dois estágios sobre `node:24-alpine`: o primeiro faz o
  `npm ci` e o build, e o segundo leva só o `build/`, o `package.json` e as dependências
  de produção. Roda com o usuário `node`.
- **D-03** — A pilha `app` tem um PostgreSQL próprio (`app-postgres`), sem porta
  publicada e com volume separado (`app-postgres-data`). Assim só o web fica exposto
  (RN-03) e os dados da pilha não se misturam com os do desenvolvimento. O Postgres de
  dev fica no perfil `dev`, ligado por padrão com `COMPOSE_PROFILES=dev` no `.env`. O
  `--profile app` da linha de comando substitui esse padrão, então a pilha `app` não
  sobe o banco de dev (RN-04 e CA-01.3).
- **D-04** — A CI ganha um job `app` (build das imagens e teste de fumaça da pilha) que
  roda quando `backend/`, `web/`, o `docker-compose.yml` ou a CI mudam, e entra no
  `ci-ok`. O teste de fumaça é um script versionado (`scripts/smoke-app.sh`), que também
  roda na máquina.
- **D-05** — A lógica da Home (dias, filtros, contagens, destaque, tempo relativo) fica em
  funções puras em `web/src/lib/home/`, que recebem o "agora" como parâmetro. Assim os
  testes fixam o horário (CA-02.4, CA-05.2).
- **D-06** — Os dados fictícios ficam em `web/src/lib/home/fixtures.ts`, com os 6 lobbies
  e a contagem por dia do design. Os lobbies são posicionados em relação a "hoje".
- **D-07** — Os componentes do design system viram componentes Svelte em
  `web/src/lib/ui/`. Os tokens CSS e os assets SVG são copiados do projeto do Claude
  Design sem alteração.

## US-01 — Pilha local em containers  (P1)

### T-01 — Imagens da API e do web  [x]
- Cobre: RN-05, CA-01.5, D-01, D-02
- Depende de: —
- Paralelizável: não
- Arquivos: `backend/Dockerfile`, `backend/.dockerignore`, `web/Dockerfile`,
  `web/.dockerignore`
- Pronto quando: `docker build` das duas imagens passa; `docker inspect` mostra usuário
  não-root; a imagem da API não tem `go` nem o código-fonte, e a do web não tem
  `devDependencies`.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff81b6bcbdf677fbd182b2

### T-02 — Perfil `app` do Compose  [x]
- Cobre: RN-01, RN-02, RN-03, RN-04, RN-06, CA-01.1, CA-01.2, CA-01.3, CA-01.4, D-03
- Depende de: T-01
- Paralelizável: não
- Arquivos: `docker-compose.yml`, `.env.example`
- Pronto quando: `docker compose --profile app up -d --wait` deixa banco, API e web
  saudáveis e o `migrate` termina com 0; `docker compose up -d` sem perfil sobe só o
  Postgres de dev; só a porta do web aparece publicada pela pilha `app`; com uma migração
  quebrada de propósito, o `migrate` sai com erro e a API não inicia.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff811db8d7ec2dbc1ec445

## US-02 — Grupos do dia  (P1)

### T-03 — Página de status em `/status`  [P] [x]
- Cobre: RN-21, CA-02.7
- Depende de: —
- Paralelizável: [P] com T-01 e T-02
- Arquivos: `web/src/routes/status/`, `web/test/e2e/status.spec.ts`
- Pronto quando: `/status` mostra os três estados da fundação; os testes Vitest e
  Playwright do status passam no caminho novo.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff813cbb75f86ca7a538b9

### T-04 — Base do design system no web  [P] [ ]
- Cobre: RN-22, RNF-01, RNF-02, D-07
- Depende de: —
- Paralelizável: [P] com T-01 a T-03
- Arquivos: `web/src/lib/ui/` (tokens, Icon, Button, IconButton, Badge, Select,
  Checkbox, Radio, Drawer, Tooltip), `web/static/brand/`, `web/package.json`
  (`@lucide/svelte`, `@fontsource/*`)
- Pronto quando: os componentes têm teste de renderização (rótulo acessível, estado
  desabilitado com a dica); `npm run build` não referencia nenhum domínio externo.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff818d8a8dc8f2c50dbab5

### T-05 — Lógica da Home e dados fictícios  [P] [ ]
- Cobre: RN-08 a RN-17, CA-02.1, CA-02.2, CA-02.3, CA-02.4, CA-03.1, CA-04.1 a CA-04.8,
  CA-05.1, CA-05.2, CA-05.3, D-05, D-06
- Depende de: —
- Paralelizável: [P] com T-01 a T-04
- Arquivos: `web/src/lib/home/` (`fixtures.ts`, `lobbies.ts`, `days.ts`, `filters.ts`,
  `time.ts`) e os `*.spec.ts`
- Pronto quando: os testes Vitest de cada critério listado passam, com o "agora" fixado.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff8149bdcad801f9868b49

### T-06 — Tela da Home  [ ]
- Cobre: RN-18, RN-20, CA-02.1, CA-02.5, CA-02.6, CA-03.2, CA-03.3, CA-04.7
- Depende de: T-03, T-04, T-05
- Paralelizável: não
- Arquivos: `web/src/routes/+page.server.ts`, `web/src/routes/+page.svelte`,
  `web/src/lib/home/components/` (TopBar, DaySelector, FilterPanel, FeaturedLobby,
  LobbyCard, RoleComposition, EmptyState)
- Pronto quando: `/` renderiza no servidor os cards de hoje; troca de dia, filtros,
  "Limpar filtros" e estados vazios funcionam; os controles sem backend estão
  desabilitados com "Disponível em breve"; teste de renderização do SSR passa.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff819f9688df62aa68a5db

## US-06 — Celular  (P2)

### T-07 — Layout abaixo de 900 px  [ ]
- Cobre: RN-19, CA-06.1, CA-06.2
- Depende de: T-06
- Paralelizável: não
- Arquivos: componentes da Home, `web/src/lib/ui/Drawer.svelte`
- Pronto quando: em 390 px a barra lateral some, "Filtros (N)" abre a gaveta e "Criar
  lobby" vira ícone.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff816e8a53cfb9d807712d

## US-02 a US-06 — Ponta a ponta

### T-08 — Playwright da Home  [ ]
- Cobre: CA-02.1, CA-02.5, CA-02.6, CA-03.2, CA-04.1, CA-04.7, CA-06.1, CA-06.2
- Depende de: T-07
- Paralelizável: não
- Arquivos: `web/test/e2e/home.spec.ts`, `web/playwright.config.ts`
- Pronto quando: os testes passam em desktop (1280 px) e celular (390 px).
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff81babf88da228b7ce0aa

## US-07 — Imagens na CI  (P2)

### T-09 — Job `app` na CI e teste de fumaça  [ ]
- Cobre: RN-07, CA-07.1, CA-01.1, CA-01.3, CA-01.5, D-04
- Depende de: T-02, T-06
- Paralelizável: não
- Arquivos: `.github/workflows/ci.yml`, `scripts/smoke-app.sh`
- Pronto quando: o script passa na máquina e no job `app` da CI do PR (sobe a pilha,
  confere a Home em `/`, "API online" em `/status`, portas publicadas e usuários
  não-root), e o `ci-ok` depende do job.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff8114986dd2841c551cb3

## US-01 — Fechamento

### T-10 — README e estado do projeto  [ ]
- Cobre: RN-06, CA-01.6
- Depende de: T-09
- Paralelizável: não
- Arquivos: `README.md`, `CLAUDE.md` (hospedagem decidida pelo ADR-06; estado atual)
- Pronto quando: o README tem subir, parar e apagar a pilha `app`; toda variável do
  Compose e das imagens está no `.env.example`.
- Commit: —
- Notion: https://app.notion.com/p/3ebd4a3a5eff8156b9f9c89b6a1dc609

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-02, T-09 |
| CA-01.2 | T-02 |
| CA-01.3 | T-02, T-09 |
| CA-01.4 | T-02 |
| CA-01.5 | T-01, T-09 |
| CA-01.6 | T-10 |
| CA-02.1 | T-05, T-06, T-08 |
| CA-02.2 | T-05 |
| CA-02.3 | T-05 |
| CA-02.4 | T-05 |
| CA-02.5 | T-06, T-08 |
| CA-02.6 | T-06, T-08 |
| CA-02.7 | T-03 |
| CA-03.1 | T-05 |
| CA-03.2 | T-06, T-08 |
| CA-03.3 | T-06 |
| CA-04.1 | T-05, T-08 |
| CA-04.2 | T-05 |
| CA-04.3 | T-05 |
| CA-04.4 | T-05 |
| CA-04.5 | T-05 |
| CA-04.6 | T-05 |
| CA-04.7 | T-05, T-06, T-08 |
| CA-04.8 | T-05 |
| CA-05.1 | T-05 |
| CA-05.2 | T-05 |
| CA-05.3 | T-05 |
| CA-06.1 | T-07, T-08 |
| CA-06.2 | T-07, T-08 |
| CA-07.1 | T-09 |

Todos os 30 critérios da spec estão cobertos.

## Descobertas
- 2026-09-30 — O `npm prune --omit=dev` deixava 21 pacotes de desenvolvimento no
  `node_modules` da imagem do web. — Resolvido na T-01 com um estágio `prod-deps` que faz
  `npm ci --omit=dev` do zero.
- 2026-09-30 — Serviço sem perfil sobe sempre, então o Postgres de dev subiria junto com a
  pilha `app`, com a porta 5432 publicada (fere o CA-01.3). — Resolvido na T-02 com o
  perfil `dev` padrão via `COMPOSE_PROFILES` (D-03). Quem já tem um `.env` antigo precisa
  acrescentar a variável; a T-10 avisa no README.
