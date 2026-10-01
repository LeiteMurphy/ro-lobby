# Relatório de validação — home-local (ciclo 2)

**Veredito geral:** REPROVADO
**Execução:** testes passou (backend `go test ./...` ok, integração `-tags=integration` ok, Vitest 62/62, Playwright 14/14, `scripts/smoke-app.sh` 14/14 verificações) · lint ok (`gofmt -l` vazio, `golangci-lint` 0 issues, Prettier + ESLint ok, `svelte-check` 0 erros/0 avisos) · build ok (`go build ./...`, `npm run build`; imagens construídas pelo smoke). `npm run generate` não gerou diff. CI do PR #11 (run 36802776997, head `373dbdb`): todos os jobs verdes, CI ok incluído.

Intervalo validado: `origin/main..feature/home-local` (13 commits, 8c77cb0..373dbdb). Desde o ciclo 1 (`fa9d46d..373dbdb`) mudaram só `days.ts` (`dayAriaLabel`), `DaySelector.svelte`, o comentário do `catalog.ts`, dois testes Vitest, um e2e de teclado (RNF-01), a descoberta em `tasks.md` e o nome do relatório do ciclo 1. Backend, Compose, Dockerfiles e CI não mudaram.

O motivo da reprovação é o RNF-01. Passei Tab por toda a Home e medi o estilo calculado de cada controle com foco e sem foco. Dois controles não mudam nada ao receber foco pelo teclado. O ciclo 1 tinha registrado "todos com anel", e isso não se confirma.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 22 | 0 | 0 | 0 |
| Critérios (CA) | 30 | 0 | 0 | 0 |
| Não funcionais | 2 | 0 | 1 | 0 |

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `docker-compose.yml`: `app-postgres` → `migrate` (`service_healthy`) → `api` (`service_completed_successfully`) → `web` (`service_healthy`); `build: ./backend` e `./web` | `scripts/smoke-app.sh` CA-01.1 (executado: 14/14); job "Pilha app" verde | — |
| RN-02 | ✅ | `api.depends_on.migrate.condition: service_completed_successfully`, `migrate.restart: "no"` | `TestCheck_RN02_*` (healthcheck). Executado num projeto Compose isolado (`-p rolobbyval`, override `command: ["bogus"]`): `migrate` exit=1, `api` ficou `created` com `StartedAt=0001-01-01`. Projeto apagado com `down -v` | O caminho de falha segue sem regressão automatizada (ver Observações) |
| RN-03 | ✅ | Só `web.ports` publica `${APP_WEB_PORT:-3000}:3000`; `API_BASE_URL: http://api:8080` | smoke CA-01.3; `docker ps`: `api` e `app-postgres` sem porta publicada | — |
| RN-04 | ✅ | `postgres` no perfil `dev`; `COMPOSE_PROFILES=dev` no `.env.example` (D-03) | smoke CA-01.4 (`config --services` = `postgres`) | Verificação indireta, igual ao ciclo 1 |
| RN-05 | ✅ | `backend/Dockerfile` distroless `nonroot`; `web/Dockerfile` com `prod-deps` (`npm ci --omit=dev`) e `USER node` | smoke CA-01.5 (usuário, sem Go, sem fonte, sem devDependencies) | — |
| RN-06 | ✅ | `.env.example` e `README.md` "Pilha completa em containers" (subir, logs, parar, apagar) | Inspeção | Não mudou desde o ciclo 1 |
| RN-07 | ✅ | `.github/workflows/ci.yml`: filtro `app`, job `app` sem push, `ci-ok.needs` inclui `app` | Job "Pilha app" verde no run 36802776997 | — |
| RN-08 | ✅ | `fixtures.ts` `getHomeLobbies()`; único consumidor fora de testes: `routes/+page.server.ts` | `home.spec.ts` CA-02.1 | — |
| RN-09 | ✅ | `time.ts` (`TIME_ZONE`, `Intl.DateTimeFormat` com `timeZone`) | `home.spec.ts` "RN-09" (2). Executado com `timezoneId: 'Asia/Tokyo'` e relógio fixo em 16:40 de São Paulo: horários 18:00…22:45 e "em 1 h 20 min" | — |
| RN-10 | ✅ | `days.ts` `buildDays`; `+page.svelte` começa em `data.today` | `home.spec.ts` CA-03.1 e "RN-10 / CA-03.1: os 14 dias atravessam o fim do mês"; `routes/home.spec.ts` CA-03.1 | — |
| RN-11 | ✅ | `lobbies.ts` `lobbiesForDay` (ordena por `toMinutes`) | `home.spec.ts` CA-02.1 / RN-11; e2e CA-02.1, CA-03.2 | — |
| RN-12 | ✅ | `filters.ts` `matches` (E; OU entre funções; `minLevel <= N`; `[start, end)`) | `home.spec.ts` CA-04.1 a CA-04.5 e "RN-12: o fim da faixa é excluído" | — |
| RN-13 | ✅ | `filters.ts` `roleCounts` (sem filtros), `timeRangeCounts` (`skipTime`) | `home.spec.ts` CA-04.6 (2) | — |
| RN-14 | ✅ | `lobbies.ts` `openSlots`, `isFull`, `edgeRole`; `LobbyCard.svelte` (`class:full`, borda `--fg-4`, `Badge` "Lotado" no lugar de "Candidatar") | `home.spec.ts` CA-02.2, CA-02.3, RN-14; executado no SSR de `localhost:3000`: o único card `data-edge="full"` tem "Lotado" e não tem "Candidatar"; os outros 5 têm "Candidatar" e não têm "Lotado" | — |
| RN-15 | ✅ | `time.ts` `relativeLabel`; `+page.svelte:34` `setInterval` de 60 s | `home.spec.ts` CA-02.4 e "RN-15: formatos". Executado com `page.clock`: "em 1 h 20 min" → "em 1 h 19 min" após 61 s | A atualização por minuto só tem prova por execução |
| RN-16 | ✅ | `lobbies.ts` `featuredLobby`; `+page.svelte:46` usa `dayLobbies` (sem filtros) | `home.spec.ts` CA-05.1, CA-05.2 (2), CA-05.3 | Nenhum teste aplica filtro e confere o destaque |
| RN-17 | ✅ | `filters.ts` `emptyState`; `EmptyState.svelte` | `home.spec.ts` CA-03.3, CA-04.7; e2e CA-03.3, CA-04.7 | — |
| RN-18 | ✅ | `Button.svelte` / `IconButton.svelte` prop `soon` (`aria-disabled`, tooltip "Disponível em breve") | `ui.spec.ts` RN-18; `routes/home.spec.ts` CA-02.5; e2e CA-02.5 | — |
| RN-19 | ✅ | `+page.svelte` `@media (max-width: 899px)`, `Drawer.svelte`, `activeFilterCount`, `TopBar.svelte` | e2e CA-06.1, CA-06.2, "RN-19: Criar lobby vira ícone"; `home.spec.ts` CA-06.2 / RN-19 | — |
| RN-20 | ✅ | `+page.server.ts` `load` | `routes/home.spec.ts` CA-02.6; e2e CA-02.6; smoke confere `data-testid="lobby-card"` no HTML | — |
| RN-21 | ✅ | `routes/status/` | `routes/status/page.spec.ts`; e2e `status.spec.ts`; smoke "API online" | — |
| RN-22 | ✅ | Assets em `web/static/brand/`, instâncias com nomes inventados; comentário do `catalog.ts` corrigido | `routes/home.spec.ts` RN-22 (lista de bloqueio) | Ver Observações (classes e meta description) |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | perfil `app` | smoke CA-01.1 (migrate 0, três serviços saudáveis, `/` com cards, `/status` "API online"); CI "Pilha app" | — |
| CA-01.2 | ✅ | `depends_on … service_completed_successfully` | Execução em projeto isolado: migrate exit 1, API nunca iniciou | Sem teste automatizado |
| CA-01.3 | ✅ | só `web.ports` | smoke CA-01.3; `docker ps` | — |
| CA-01.4 | ✅ | perfil `dev` | smoke CA-01.4 | Indireto (`config --services`) |
| CA-01.5 | ✅ | Dockerfiles multi-stage | smoke CA-01.5 | — |
| CA-01.6 | ✅ | `.env.example`, `README.md` | Inspeção | — |
| CA-02.1 | ✅ | `+page.svelte`, `LobbyCard.svelte`, `HostLine.svelte`, `RoleComposition.svelte` | `home.spec.ts` CA-02.1; `routes/home.spec.ts` CA-02.1; e2e CA-02.1 | — |
| CA-02.2 | ✅ | `LobbyCard.svelte` `{#if full}`, `.card.full`, borda `--fg-4` | `home.spec.ts` CA-02.2; `routes/home.spec.ts` CA-02.2 (só o selo); execução no SSR: sem "Candidatar" no card lotado | O teste automatizado confere só o selo. A ausência de "Candidatar" e o esmaecimento têm prova só por execução |
| CA-02.3 | ✅ | `edgeRole`; `LobbyCard.svelte:31` | `home.spec.ts` CA-02.3; SSR: card das 19:00 com `data-edge="support"` | — |
| CA-02.4 | ✅ | `relativeLabel`, `zonedNow` | `home.spec.ts` CA-02.4; `routes/home.spec.ts` CA-02.4 | — |
| CA-02.5 | ✅ | prop `soon` | `ui.spec.ts`; `routes/home.spec.ts` CA-02.5; e2e CA-02.5 | `aria-disabled` em vez de `disabled`, de propósito, para manter o foco e a dica |
| CA-02.6 | ✅ | `+page.server.ts` | `routes/home.spec.ts` CA-02.6; e2e CA-02.6 | — |
| CA-02.7 | ✅ | `routes/status/` | `status/page.spec.ts`; e2e `status.spec.ts` | — |
| CA-03.1 | ✅ | `buildDays`, `DaySelector.svelte` | `home.spec.ts` CA-03.1; `routes/home.spec.ts` CA-03.1 | — |
| CA-03.2 | ✅ | `+page.svelte` (título) | e2e CA-03.2; e2e RNF-01 (troca pelo teclado) | — |
| CA-03.3 | ✅ | `emptyState`, `EmptyState.svelte` | `home.spec.ts` CA-03.3; e2e CA-03.3 | — |
| CA-04.1 | ✅ | `matches` | `home.spec.ts` CA-04.1; e2e CA-04.1 | — |
| CA-04.2 | ✅ | `matches` (`some`) | `home.spec.ts` CA-04.2 | — |
| CA-04.3 | ✅ | `matches` (`<=`), `LEVEL_OPTIONS` | `home.spec.ts` CA-04.3 | — |
| CA-04.4 | ✅ | `inRange`, `TIME_RANGES` | `home.spec.ts` CA-04.4 | — |
| CA-04.5 | ✅ | `matches` | `home.spec.ts` CA-04.5 | — |
| CA-04.6 | ✅ | `roleCounts`, `timeRangeCounts` | `home.spec.ts` CA-04.6 (2) | — |
| CA-04.7 | ✅ | `EmptyState.svelte`, `resetFilters` | `home.spec.ts` CA-04.7; e2e CA-04.7 | — |
| CA-04.8 | ✅ | `countLabel` | `home.spec.ts` CA-04.8; e2e CA-04.1 | — |
| CA-05.1 | ✅ | `featuredLobby` | `home.spec.ts` CA-05.1; `routes/home.spec.ts` CA-05.1 | — |
| CA-05.2 | ✅ | `featuredLobby` (`> now.minutes` em hoje) | `home.spec.ts` CA-05.2 (2) | — |
| CA-05.3 | ✅ | `{#if featured}` | `home.spec.ts` CA-05.3 | — |
| CA-06.1 | ✅ | `Drawer.svelte`, CSS < 900 px | e2e CA-06.1 | — |
| CA-06.2 | ✅ | `activeFilterCount` | e2e CA-06.2; `home.spec.ts` CA-06.2 / RN-19 | — |
| CA-07.1 | ✅ | job `app` + `ci-ok.needs` | Run 36802776997 construiu as duas imagens; CI ok verde | A metade "falha deixa o CI ok vermelho" vem da lógica do `ci-ok`; nenhuma falha real foi provocada |

### Não funcionais
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RNF-01 | ❌ | Anel global em `web/src/lib/ui/tokens/base.css:31` (`:focus-visible { box-shadow: var(--focus-ring), … }`). Dois controles o perdem por especificidade: `web/src/lib/ui/Button.svelte:117-121` (`.btn--primary { box-shadow: inset … }`) e `web/src/lib/home/components/DaySelector.svelte:103-107` (`.day.sel { box-shadow: var(--glow-gold) }`). As regras com escopo do Svelte (0,2,0) vencem o `:focus-visible` global (0,1,0) | e2e `web/test/e2e/home.spec.ts:106` "RNF-01" passa, mas só foca o 2º dia (não selecionado) e um checkbox. Execução: medi `box-shadow`, `outline` e `border-color` com foco e sem foco, Tab a Tab, em 1280 px e 390 px | **"Criar lobby" (desktop):** com foco = sem foco (`inset 0 1px 0 rgba(255,255,255,.3)`, outline `none`), sem nenhum indicador. **Dia selecionado ("Hoje, qua, 30 set"):** com foco = sem foco (`glow-gold`), em 1280 px e 390 px. É o primeiro Tab stop do seletor e o dia que sempre começa selecionado. Os outros controles mostram o anel âmbar: link da marca, "Entrar com Discord", setas, dias não selecionados, checkboxes, radios, "Candidatar" e o ícone "Criar lobby" em 390 px. Os Selects mostram borda âmbar e halo de 3 px, o que a `tasks.md` documenta como regra do design system para campos. Os rótulos acessíveis estão ok |
| RNF-02 | ✅ | `@fontsource`, Lucide empacotado, `static/brand/` | e2e "RNF-02"; `ui.spec.ts` RNF-02 | — |
| RNF-03 | ✅ | — | Todos os testes novos e alterados citam ID; o teste do ciclo 1 sem ID virou "RN-10 / CA-03.1: os 14 dias atravessam o fim do mês" | Os nomes de `status.spec.ts` citam IDs da fundação (CA-07.1, CA-05.4), herdados de lá |

## Pendências para correção
1. [RNF-01] "Criar lobby" no desktop não mostra foco visível. Em `web/src/lib/ui/Button.svelte`, a variante primária precisa de `.btn--primary:focus-visible { box-shadow: var(--focus-ring), … }`, ou outra forma de não perder o anel. O RNF-01 exige o anel âmbar do design system em todo controle interativo.
2. [RNF-01] O dia selecionado no seletor não mostra foco visível. Em `web/src/lib/home/components/DaySelector.svelte`, `.day.sel` precisa compor o `--focus-ring` com o `glow-gold` quando estiver em `:focus-visible`.
3. [RNF-01] O e2e `home.spec.ts:106` não pegou esses dois casos. Ele deve cobrir o botão primário e o dia selecionado e comparar o estilo com foco contra o estilo sem foco, ou conferir a cor do anel. Hoje ele só confere `boxShadow !== 'none'`.

## Scope creep
- Nenhum. As alterações desde o ciclo 1 se ligam ao RNF-01 e ao RNF-03 e às observações do ciclo 1. Os itens no limite do escopo já registrados no ciclo 1 continuam aceitáveis: o "Entrar com Discord" só com ícone abaixo de 900 px, o filtro `app` da CI que também dispara com `.env.example` e `scripts/smoke-app.sh`, e a edição do `CLAUDE.md` prevista na T-10.
- Nada toca o "Fora de escopo".

## Observações (não bloqueantes)
- **CA-01.2 sem regressão automatizada:** confirmei de novo à mão, num projeto Compose separado, mas nem o smoke nem a CI provocam uma migração que falha. Um passo no `smoke-app.sh` com `-p` próprio e override do `command` protegeria a RN-02 sem mexer na pilha principal.
- **CA-02.2:** o teste automatizado só confere o selo "Lotado". Vale conferir também a ausência de "Candidatar" no card lotado e a classe `full`.
- **RN-22:** as classes (Arcebispo, Paladino, Feiticeiro, Sicário, Andarilho) são nomes de classe do jogo, e o `<meta name="description">` em `+page.svelte:71` cita "Ragnarok Online". Os dois vêm do design aprovado e são uso descritivo, e o comentário do `catalog.ts` agora diz isso. Mesmo assim, vale o usuário confirmar que a RN-22 os admite. A lista de bloqueio do teste segue curta.
- **Tablist dos dias:** `role="tab"` sem `tabpanel` e sem navegação por setas, com 14 Tab stops. Funciona pelo teclado, mas não segue o padrão ARIA de abas.
- **Destaque ignora filtros:** o código está certo, mas nenhum teste filtra e confere que o destaque continua.
- **Rastreabilidade:** os commits estão em Conventional Commits e citam IDs. O campo "Commit:" das tasks agora está preenchido e bate com o `git log`.
- **Ambiente:** a pilha `app` e o PostgreSQL de dev continuam no ar. O projeto isolado `rolobbyval` foi apagado com `down -v`. Os servidores do Playwright foram encerrados (portas 8080 e 4173 livres). Nenhum arquivo versionado foi alterado além deste relatório.
