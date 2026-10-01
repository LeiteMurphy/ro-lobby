# Relatório de validação — home-local (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend `go test ./...` ok, integração `-tags=integration` ok, Vitest 61/61, Playwright 13/13, `scripts/smoke-app.sh` 14/14 verificações) · lint ok (`gofmt -l` vazio, `golangci-lint` 0 issues, Prettier + ESLint ok, `svelte-check` 0 erros/0 avisos) · build ok (`go build ./...`, `npm run build`, imagens Docker da API e do web). `npm run generate` não gerou diff. CI do PR #11: todos os jobs verdes (Filtro de caminhos, Backend, Web, Ponta a ponta, Pilha app, CI ok).

Intervalo validado: `origin/main..feature/home-local` (11 commits, 8c77cb0..0f2d6bc).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 22 | 0 | 0 | 0 |
| Critérios (CA) | 30 | 0 | 0 | 0 |
| Não funcionais | 1 | 2 | 0 | 0 |

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `docker-compose.yml`: `app-postgres` → `migrate` (`service_healthy`) → `api` (`service_completed_successfully`) → `web` (`service_healthy`); `build: ./backend` e `build: ./web` com Dockerfiles de produção | `scripts/smoke-app.sh` (CA-01.1), job "Pilha app" da CI; executado localmente: ordem confirmada no log do `up --wait` | — |
| RN-02 | ✅ | `docker-compose.yml` `api.depends_on.migrate.condition: service_completed_successfully`, `migrate.restart: "no"`; `backend/cmd/migrate/main.go:24-26` sai com 1 em erro; healthcheck `backend/cmd/healthcheck/main.go` | `TestCheck_RN02_PassesOnlyOn200`, `TestCheck_RN02_FailsWhenAPIIsDown`; execução manual (override com comando inválido no `migrate`): `migrate` exit=1, `api` ficou `created` com `StartedAt=0001-01-01` | O caminho de falha não tem teste automatizado (ver Observações) |
| RN-03 | ✅ | `docker-compose.yml`: só `web.ports` publica `${APP_WEB_PORT:-3000}:3000`; `API_BASE_URL: http://api:8080` | smoke CA-01.3; `docker ps` com a pilha no ar: só `0.0.0.0:3000` publicado; `curl localhost:8080` e `:5432` sem resposta; `/status` mostra "API online" via rede interna | — |
| RN-04 | ✅ | `postgres` no perfil `dev`, `COMPOSE_PROFILES=dev` no `.env.example` (D-03) | smoke CA-01.4 (`compose config --services` = `postgres`); execução: `docker compose up -d --wait` sem perfil subiu só `ro-lobby-postgres-1` | — |
| RN-05 | ✅ | `backend/Dockerfile` (distroless `static-debian12:nonroot`, `USER nonroot`), `web/Dockerfile` (estágio `prod-deps` com `npm ci --omit=dev`, `USER node`) | smoke CA-01.5; conferido: API `nonroot:nonroot`, sem `.go` nem `/usr/local/go`; web `uid=1000(node)`, `/app` só com `build/`, `node_modules/` vazio e `package.json` | — |
| RN-06 | ✅ | `.env.example` tem `COMPOSE_PROFILES`, `APP_WEB_PORT`, `POSTGRES_*`, `DATABASE_URL`, `API_PORT`, `WEB_PORT`, `API_BASE_URL`; `ORIGIN`/`PORT` são montados no Compose e citados no comentário; `README.md` "Pilha completa em containers" com subir, logs, parar e apagar (`down -v`) | Inspeção (item de documentação) | — |
| RN-07 | ✅ | `.github/workflows/ci.yml`: filtro `app` (`backend/**`, `web/**`, `docker-compose.yml`, `*shared` com a própria CI), job `app` sem push/login em registry, `ci-ok.needs` inclui `app` | Job "Pilha app" passou no PR #11 (run 36664106070) | — |
| RN-08 | ✅ | `web/src/lib/home/fixtures.ts` `getHomeLobbies()`; único import fora de testes: `web/src/routes/+page.server.ts:1` | `home.spec.ts` "CA-02.1: dados fictícios de hoje" | — |
| RN-09 | ✅ | `web/src/lib/home/time.ts` `TIME_ZONE`, `Intl.DateTimeFormat` com `timeZone` fixo; usado no servidor (`+page.server.ts`) e no navegador (`+page.svelte`) | `home.spec.ts` "RN-09: fuso" (2 casos); execução com `timezoneId: 'Asia/Tokyo'`: horários e "em 1 h 20 min" iguais a São Paulo | — |
| RN-10 | ✅ | `days.ts` `buildDays` (14 dias, contagem, `today: i === 0`); `+page.svelte` `selectedDate = $derived(data.today)` | `home.spec.ts` CA-03.1; `routes/home.spec.ts` CA-03.1 (14 tabs, hoje `aria-selected`) | — |
| RN-11 | ✅ | `lobbies.ts` `lobbiesForDay` ordena por `toMinutes` | `home.spec.ts` "CA-02.1 / RN-11"; e2e CA-02.1 e CA-03.2 | — |
| RN-12 | ✅ | `filters.ts` `matches` (E entre filtros, OU entre funções, `minLevel <= N`, `[start, end)`) | `home.spec.ts` CA-04.1 a CA-04.5 e "RN-12: o fim da faixa é excluído" | — |
| RN-13 | ✅ | `filters.ts` `roleCounts` (sem filtros) e `timeRangeCounts` (`skipTime`); `+page.svelte:62-63` | `home.spec.ts` CA-04.6 (2 casos); execução: com instância filtrada, "Vaga para" continuou 3/4/3 | — |
| RN-14 | ✅ | `lobbies.ts` `openSlots`, `isFull`, `edgeRole` (desempate na ordem de `ROLES`); `LobbyCard.svelte` (`.full` com opacidade 0,55, borda `--fg-4`, `Badge` "Lotado" no lugar do botão) | `home.spec.ts` CA-02.2, CA-02.3, "RN-14" (2 casos); `routes/home.spec.ts` CA-02.2; execução: card lotado com opacidade 0.55, borda `rgb(90,103,122)`, sem "Candidatar" | — |
| RN-15 | ✅ | `time.ts` `relativeLabel`; `+page.svelte:43-47` `setInterval` de 60 s | `home.spec.ts` CA-02.4 (2 casos) e "RN-15: formatos"; execução com `page.clock`: "em 1 h 20 min" virou "em 1 h 19 min" após 1 min | A atualização por minuto só foi provada por execução |
| RN-16 | ✅ | `lobbies.ts` `featuredLobby`; `+page.svelte:57` usa `dayLobbies` (sem filtros) | `home.spec.ts` CA-05.1, CA-05.2 (2 casos), CA-05.3 | — |
| RN-17 | ✅ | `filters.ts` `emptyState`; `EmptyState.svelte` ("Limpar filtros" só em `kind === 'filters'`) | `home.spec.ts` CA-03.3, CA-04.7; e2e CA-03.3, CA-04.7 | — |
| RN-18 | ✅ | `Button.svelte`/`IconButton.svelte` prop `soon` (`aria-disabled`, clique anulado, tooltip "Disponível em breve"); usada em TopBar, LobbyCard, FeaturedLobby, EmptyState | `ui.spec.ts` RN-18 (2); `routes/home.spec.ts` CA-02.5 (4 rótulos); e2e CA-02.5 | — |
| RN-19 | ✅ | `+page.svelte` `@media (max-width: 899px)`, `Drawer.svelte` (`left: 0`), botão "Filtros (N)" com `activeFilterCount`; `TopBar.svelte` troca por `IconButton` | e2e CA-06.1, CA-06.2, "RN-19: Criar lobby vira ícone"; `home.spec.ts` "CA-06.2 / RN-19"; execução: gaveta em x=0, largura 320 | — |
| RN-20 | ✅ | `+page.server.ts` `load` com os lobbies; cards renderizados no SSR | `routes/home.spec.ts` CA-02.6; e2e CA-02.6 (`request.get('/')`); smoke confere `data-testid="lobby-card"` no HTML | — |
| RN-21 | ✅ | `web/src/routes/status/+page.server.ts` e `+page.svelte` (movidos) | `routes/status/page.spec.ts` (3 estados); e2e `status.spec.ts` em `/status`; smoke "API online" | — |
| RN-22 | ✅ | Assets originais em `web/static/brand/`, instâncias com nomes inventados, textos em pt-BR com "você" e sem exclamação | `routes/home.spec.ts` "RN-22" (lista de bloqueio); comparado com o `Home v2.dc.html`: dados e textos iguais ao design | Ver Observações (nomes de classe e lista de bloqueio curta) |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | `docker-compose.yml` perfil `app` | smoke: migrate 0, `app-postgres`/`api`/`web` healthy, `/` 200 com a Home, `/status` "API online"; CI "Pilha app" | — |
| CA-01.2 | ✅ | `depends_on … service_completed_successfully` | Execução manual com override no scratchpad: migrate exit 1, API nunca iniciou | Sem teste automatizado |
| CA-01.3 | ✅ | só `web.ports` | smoke CA-01.3; `docker ps` e `curl` nas portas 8080/5432 | — |
| CA-01.4 | ✅ | perfil `dev` + `COMPOSE_PROFILES=dev` | smoke (`config --services`); execução real de `docker compose up -d --wait` | — |
| CA-01.5 | ✅ | Dockerfiles multi-stage | smoke CA-01.5; inspeção manual das duas imagens | — |
| CA-01.6 | ✅ | `.env.example`, `README.md` | Inspeção | — |
| CA-02.1 | ✅ | `+page.svelte`, `LobbyCard.svelte`, `HostLine.svelte`, `RoleComposition.svelte` | `home.spec.ts` CA-02.1 (2); `routes/home.spec.ts` CA-02.1; e2e CA-02.1 | — |
| CA-02.2 | ✅ | `LobbyCard.svelte` `{#if full}` Badge, `.card.full`, borda `--fg-4` | `home.spec.ts` CA-02.2; `routes/home.spec.ts` CA-02.2; execução (opacidade, cor da borda, sem "Candidatar") | Os testes automatizados não conferem a ausência de "Candidatar" nem o esmaecimento |
| CA-02.3 | ✅ | `edgeRole`; `LobbyCard.svelte:31` `var(--${edge}-400)` | `home.spec.ts` CA-02.3; execução: card das 19:00 com `data-edge="support"` | — |
| CA-02.4 | ✅ | `relativeLabel`, `zonedNow` | `home.spec.ts` CA-02.4 (3); `routes/home.spec.ts` CA-02.4 | — |
| CA-02.5 | ✅ | prop `soon` | `ui.spec.ts`; `routes/home.spec.ts` CA-02.5; e2e CA-02.5 (hover mostra a dica, clique não navega) | Os controles usam `aria-disabled`, não `disabled`, para seguirem focáveis e mostrarem a dica. É intencional (documentado no Button) |
| CA-02.6 | ✅ | `+page.server.ts` | `routes/home.spec.ts` CA-02.6; e2e CA-02.6 | — |
| CA-02.7 | ✅ | `routes/status/` | `status/page.spec.ts`; e2e `status.spec.ts` | — |
| CA-03.1 | ✅ | `buildDays`, `DaySelector.svelte` | `home.spec.ts` CA-03.1; `routes/home.spec.ts` CA-03.1 | — |
| CA-03.2 | ✅ | `+page.svelte:68` título | e2e CA-03.2; execução pelo teclado (Enter no dia 2 → "Grupos para qui, 1 out") | — |
| CA-03.3 | ✅ | `emptyState`, `EmptyState.svelte` | `home.spec.ts` CA-03.3; e2e CA-03.3 (sem "Limpar filtros") | — |
| CA-04.1 | ✅ | `matches` | `home.spec.ts` CA-04.1; e2e CA-04.1 | — |
| CA-04.2 | ✅ | `matches` (`some`) | `home.spec.ts` CA-04.2 | — |
| CA-04.3 | ✅ | `matches` (`<=`), `LEVEL_OPTIONS` | `home.spec.ts` CA-04.3 | — |
| CA-04.4 | ✅ | `inRange`, `TIME_RANGES` | `home.spec.ts` CA-04.4 | — |
| CA-04.5 | ✅ | `matches` | `home.spec.ts` CA-04.5 | — |
| CA-04.6 | ✅ | `roleCounts`, `timeRangeCounts` | `home.spec.ts` CA-04.6 (2) | — |
| CA-04.7 | ✅ | `EmptyState.svelte`, `resetFilters` | `home.spec.ts` CA-04.7; e2e CA-04.7 (volta a todos os grupos) | — |
| CA-04.8 | ✅ | `countLabel`, `+page.svelte:69-73` | `home.spec.ts` CA-04.8; e2e CA-04.1 confere "N de M grupos" | — |
| CA-05.1 | ✅ | `featuredLobby` | `home.spec.ts` CA-05.1; `routes/home.spec.ts` CA-05.1 | — |
| CA-05.2 | ✅ | `featuredLobby` (`> now.minutes` só em hoje) | `home.spec.ts` CA-05.2 (2) | — |
| CA-05.3 | ✅ | `{#if featured}` | `home.spec.ts` CA-05.3 | — |
| CA-06.1 | ✅ | `Drawer.svelte`, CSS `< 900 px` | e2e CA-06.1; execução (x=0, barra lateral oculta, Esc fecha) | — |
| CA-06.2 | ✅ | `activeFilterCount`, `+page.svelte:132` | e2e CA-06.2; `home.spec.ts` "CA-06.2 / RN-19" | — |
| CA-07.1 | ✅ | job `app` + `ci-ok.needs` | CI do PR #11 construiu as duas imagens e passou | A metade "falha no build deixa o CI ok vermelho" vem da lógica do `ci-ok` (qualquer resultado diferente de success/skipped falha); não foi provocada uma falha real |

### Não funcionais
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RNF-01 | ⚠️ | `aria-label` em IconButton/Select/tabs/Drawer; `:focus-visible` com `--focus-ring` em `tokens/base.css:31`; `Check.svelte:84` | `ui.spec.ts` (rótulos, Drawer inerte quando fechada); execução: Tab percorre marca, Criar lobby, Entrar com Discord, setas, 14 dias, filtros, todos com anel | Só os rótulos têm teste automatizado; teclado e foco visível foram provados só por execução. O Select usa `border-color: var(--accent)` + anel `--accent-soft` no contêiner em vez do `--focus-ring` padrão |
| RNF-02 | ✅ | Fontes via `@fontsource` (`tokens/index.css`), ícones Lucide empacotados (`Icon.svelte`), assets em `static/brand/` | e2e "RNF-02" (nenhuma origem externa); `ui.spec.ts` RNF-02; `build/client` sem URLs externas carregáveis (só `svelte.dev` em mensagens de erro, namespaces `w3.org` e `c2pa.org` nos metadados dos SVG) | — |
| RNF-03 | ⚠️ | — | Todos os testes novos citam ID, exceto `home.spec.ts:54` "addDays atravessa o fim do mês" | Teste auxiliar sem ID de critério |

## Pendências para correção
Nenhuma bloqueante. Ajustes sugeridos (P2/não funcionais):
1. [RNF-03] `web/src/lib/home/home.spec.ts:54`: o teste "addDays atravessa o fim do mês" não cita ID. Associar a RN-10 ou CA-03.1, já que sustenta os 14 dias.
2. [RNF-01] Não há teste automatizado de teclado nem de foco visível. Um e2e curto com Tab e o anel âmbar fecharia o item. O foco do Select usa um anel diferente do `--focus-ring` do design system.

## Scope creep
- Nenhum. Todas as alterações se ligam a IDs da spec ou às decisões D-01 a D-07. Itens próximos do limite, todos aceitáveis:
  - A versão só com ícone de "Entrar com Discord" abaixo de 900 px (`TopBar.svelte`) não está na RN-19, que cita só "Criar lobby". Ela deriva da RN-18 e do layout do design.
  - O filtro `app` da CI também dispara com `.env.example` e `scripts/smoke-app.sh`. Isso vai além da RN-07, mas coerente com ela.
  - A edição do `CLAUDE.md` (estado atual, ADR-06 decidido) está prevista na T-10.
- Nada toca "Fora de escopo": não há deploy na nuvem, push de imagem, API de lobbies, calendário mensal nem outras telas.

## Observações (não bloqueantes)
- **CA-01.2 sem regressão automatizada:** o comportamento foi confirmado à mão, mas nem o smoke nem a CI provocam uma migração que falha. Um passo extra no `smoke-app.sh` (override com `command` inválido, como o usado nesta validação) protegeria a RN-02.
- **Smoke CA-01.4:** o script confere com `docker compose config --services`, sem subir o modo dev de fato. O resultado bate com a execução real, mas é uma verificação indireta.
- **Texto acessível dos dias sem grupos:** `DaySelector.svelte` gera `aria-label` "sex, 2 out: nenhum grupos" (plural errado). Deveria ser "nenhum grupo".
- **RN-22:** as classes (Arcebispo, Paladino, Feiticeiro, Sicário, Andarilho) são nomes de classe do jogo, e o `<meta name="description">` cita "Ragnarok Online". Ambos vêm do design aprovado e são uso descritivo, não marca, mas o comentário de `catalog.ts` ("com nomes inventados") só vale para as instâncias. O teste RN-22 bloqueia só uma lista curta (Gravity, Prontera, Glast Heim, Poring).
- **Tablist dos dias:** usa `role="tab"` sem `tabpanel` associado nem navegação por setas. Cada dia é um Tab stop (14 paradas). Funciona pelo teclado, mas não segue o padrão ARIA de abas.
- **Destaque ignora filtros:** o código passa `dayLobbies` (sem filtros), mas nenhum teste filtra e confere que o destaque continua.
- **Rastreabilidade:** os commits citam os IDs em Conventional Commits. As tasks estão todas `[x]` e batem com o código, mas o campo "Commit:" de cada task ficou "—".
- **Ambiente:** a pilha `app` foi derrubada sem `-v` (o volume `app-postgres-data` foi mantido). O PostgreSQL de dev, que já estava no ar antes da validação, continua no ar. Nenhum arquivo versionado foi alterado.
