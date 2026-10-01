# Relatório de validação — home-local (ciclo 3)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend `go test ./...` ok, integração `-tags=integration` ok, Vitest 62/62, Playwright 16/16, `scripts/smoke-app.sh` 14/14 verificações) · lint ok (`gofmt -l` vazio, `golangci-lint` 0 issues, Prettier + ESLint ok, `svelte-check` 0 erros/0 avisos em 403 arquivos) · build ok (`go build ./...`, `npm run build`; imagens reconstruídas pelo smoke). `npm run generate` não gerou diff. CI do PR #11 (run 36803558816, head `c260186`): Filtro de caminhos, Backend, Web, Ponta a ponta, Pilha app e CI ok, todos verdes.

Intervalo validado: `origin/main..feature/home-local` (15 commits, 8c77cb0..c260186). Desde o ciclo 2 (`373dbdb..c260186`) mudaram só `Button.svelte`, `IconButton.svelte` e `DaySelector.svelte` (regras de `:focus-visible`), o e2e `web/test/e2e/home.spec.ts` (teste de foco em todo ponto de Tab, desktop e celular), a descoberta em `tasks.md` e o relatório do ciclo 2. Backend, Compose, Dockerfiles, CI e scripts não mudaram (`git diff --stat` vazio nesses caminhos).

A pendência do ciclo 2 (RNF-01) foi corrigida. Medi por conta própria, com um script Playwright contra a pilha `app` em `localhost:3000`, o `box-shadow` de cada ponto de Tab com foco e sem foco. Todos os 30 pontos em 1280 px e os 23 em 390 px mostram o anel âmbar (`rgb(246, 187, 69) 0 0 0 4px`), incluindo "Criar lobby" no desktop e o dia selecionado ("Hoje, qua, 30 set"), que no ciclo 2 não mudavam. Os dois Selects mostram a borda âmbar com o halo de 3 px, regra do design system para campos já registrada na `tasks.md`.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 22 | 0 | 0 | 0 |
| Critérios (CA) | 30 | 0 | 0 | 0 |
| Não funcionais | 3 | 0 | 0 | 0 |

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `docker-compose.yml`: `app-postgres` → `migrate` (`service_healthy`, l. 54) → `api` (`service_completed_successfully`, l. 66) → `web` (`service_healthy`, l. 87); builds de produção | `smoke-app.sh` CA-01.1 (executado); job "Pilha app" verde | — |
| RN-02 | ✅ | `api.depends_on.migrate: service_completed_successfully` | Execução manual nos ciclos 1 e 2 (migrate exit 1, API não iniciou); Compose não mudou desde então | Sem teste automatizado (ver Observações) |
| RN-03 | ✅ | `ports` só no `web` (l. 83), `APP_WEB_PORT`; API e `app-postgres` sem porta | `smoke-app.sh` CA-01.3; `docker ps` mostra só `0.0.0.0:3000` na pilha app | — |
| RN-04 | ✅ | Postgres de dev em `profiles: [dev]` + `COMPOSE_PROFILES=dev` (D-03) | `smoke-app.sh` CA-01.4 | — |
| RN-05 | ✅ | `backend/Dockerfile` (distroless nonroot), `web/Dockerfile` (estágio `prod-deps`, usuário `node`) | `smoke-app.sh` CA-01.5 (4 verificações) | — |
| RN-06 | ✅ | `.env.example`, `README.md` | Inspeção (ciclos anteriores; arquivos não mudaram) | — |
| RN-07 | ✅ | `.github/workflows/ci.yml` job `app` + filtro de caminhos + `ci-ok.needs` | Run 36803558816: "Pilha app" verde, CI ok verde | — |
| RN-08 | ✅ | `web/src/lib/home/fixtures.ts`, acesso por função em `catalog.ts` | `home.spec.ts` CA-02.1 | — |
| RN-09 | ✅ | `time.ts` (`zonedNow`, `America/Sao_Paulo`) | `home.spec.ts` CA-02.4 ("19:40 UTC é 16:40 em São Paulo") | — |
| RN-10 | ✅ | `days.ts` `buildDays` | `home.spec.ts` CA-03.1 e virada de mês; medição: 14 dias, hoje selecionado | — |
| RN-11 | ✅ | ordenação por horário em `lobbies.ts` | `home.spec.ts` CA-02.1; e2e CA-02.1 e CA-03.2 | — |
| RN-12 | ✅ | `filters.ts` `matches`, `inRange` | `home.spec.ts` CA-04.1 a CA-04.5 | — |
| RN-13 | ✅ | `roleCounts`, `timeRangeCounts` | `home.spec.ts` CA-04.6 (2 testes) | — |
| RN-14 | ✅ | `edgeRole`, `LobbyCard.svelte` (`.card.full`) | `home.spec.ts` CA-02.2, CA-02.3 | — |
| RN-15 | ✅ | `relativeLabel` + atualização por minuto | `home.spec.ts` CA-02.4 (2 testes) | — |
| RN-16 | ✅ | `featuredLobby` | `home.spec.ts` CA-05.1, CA-05.2, CA-05.3 | Às 22:58 de hoje não havia destaque, como esperado (todos os lobbies de hoje já começaram) |
| RN-17 | ✅ | `emptyState`, `EmptyState.svelte` | `home.spec.ts` CA-03.3, CA-04.7; e2e CA-03.3 e CA-04.7 | — |
| RN-18 | ✅ | prop `soon` no Button/IconButton | `ui.spec.ts`; `routes/home.spec.ts` CA-02.5; e2e CA-02.5 | — |
| RN-19 | ✅ | `Drawer.svelte`, CSS < 900 px, `activeFilterCount` | e2e CA-06.1, CA-06.2, RN-19 | — |
| RN-20 | ✅ | `+page.server.ts` | `routes/home.spec.ts` CA-02.6; e2e CA-02.6 | — |
| RN-21 | ✅ | `web/src/routes/status/` | `status/page.spec.ts`; e2e `status.spec.ts`; smoke `/status` "API online" | — |
| RN-22 | ✅ | Assets em `web/static/brand/`, instâncias com nomes inventados | `routes/home.spec.ts` RN-22 (lista de bloqueio) | Ver Observações (classes e meta description) |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | perfil `app` | smoke CA-01.1 (migrate 0, três serviços saudáveis, `/` com cards, `/status` "API online"); CI "Pilha app" | — |
| CA-01.2 | ✅ | `depends_on … service_completed_successfully` | Execução manual nos ciclos anteriores; sem mudança no Compose | Sem regressão automatizada |
| CA-01.3 | ✅ | só `web.ports` | smoke CA-01.3; `docker ps` | — |
| CA-01.4 | ✅ | perfil `dev` | smoke CA-01.4 | Indireto (`config --services`) |
| CA-01.5 | ✅ | Dockerfiles multi-stage | smoke CA-01.5 | — |
| CA-01.6 | ✅ | `.env.example`, `README.md` | Inspeção | — |
| CA-02.1 | ✅ | `+page.svelte`, `LobbyCard.svelte`, `HostLine.svelte`, `RoleComposition.svelte` | `home.spec.ts` CA-02.1; `routes/home.spec.ts` CA-02.1; e2e CA-02.1 | — |
| CA-02.2 | ✅ | `LobbyCard.svelte` `{#if full}`, `.card.full` | `home.spec.ts:75` CA-02.2; `routes/home.spec.ts:36` (selo) | O teste de tela confere só o selo |
| CA-02.3 | ✅ | `edgeRole` | `home.spec.ts` CA-02.3 | — |
| CA-02.4 | ✅ | `relativeLabel`, `zonedNow` | `home.spec.ts:96`, `:100`; `routes/home.spec.ts:40` | — |
| CA-02.5 | ✅ | prop `soon` (`aria-disabled` + dica) | `ui.spec.ts`; `routes/home.spec.ts` CA-02.5; e2e CA-02.5 | — |
| CA-02.6 | ✅ | `+page.server.ts` | `routes/home.spec.ts` CA-02.6; e2e CA-02.6 | — |
| CA-02.7 | ✅ | `routes/status/` | e2e `status.spec.ts`; smoke | — |
| CA-03.1 | ✅ | `buildDays`, `DaySelector.svelte` | `home.spec.ts` CA-03.1 | — |
| CA-03.2 | ✅ | `+page.svelte` (título) | e2e CA-03.2; e2e RNF-01 (troca pelo teclado) | — |
| CA-03.3 | ✅ | `emptyState`, `EmptyState.svelte` | `home.spec.ts` CA-03.3; e2e CA-03.3 | — |
| CA-04.1 | ✅ | `matches` | `home.spec.ts` CA-04.1; e2e CA-04.1 | — |
| CA-04.2 | ✅ | `matches` (`some`) | `home.spec.ts` CA-04.2 | — |
| CA-04.3 | ✅ | `matches` (`<=`) | `home.spec.ts` CA-04.3 | — |
| CA-04.4 | ✅ | `inRange`, `TIME_RANGES` | `home.spec.ts` CA-04.4 | — |
| CA-04.5 | ✅ | `matches` | `home.spec.ts` CA-04.5 | — |
| CA-04.6 | ✅ | `roleCounts`, `timeRangeCounts` | `home.spec.ts:211`, `:221` | — |
| CA-04.7 | ✅ | `EmptyState.svelte`, `resetFilters` | `home.spec.ts` CA-04.7; e2e CA-04.7 | — |
| CA-04.8 | ✅ | `countLabel` | `home.spec.ts` CA-04.8 | — |
| CA-05.1 | ✅ | `featuredLobby` | `home.spec.ts` CA-05.1 | — |
| CA-05.2 | ✅ | `featuredLobby` | `home.spec.ts` CA-05.2 | — |
| CA-05.3 | ✅ | `{#if featured}` | `home.spec.ts:269` CA-05.3 | — |
| CA-06.1 | ✅ | `Drawer.svelte`, CSS < 900 px | e2e CA-06.1 | — |
| CA-06.2 | ✅ | `activeFilterCount` | e2e CA-06.2 | — |
| CA-07.1 | ✅ | job `app` + `ci-ok.needs` | Run 36803558816 construiu as duas imagens; CI ok verde | A metade "falha deixa o CI ok vermelho" vem da lógica do `ci-ok`; nenhuma falha real foi provocada |

### Não funcionais
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RNF-01 | ✅ | Anel global em `web/src/lib/ui/tokens/base.css:31`. Correções: `Button.svelte` `.btn.btn:focus-visible`, `IconButton.svelte` `.ib.ib:focus-visible`, `DaySelector.svelte` `.day.day:focus-visible` e `.day.sel.sel:focus-visible` (anel + `glow-gold`). Com o escopo do Svelte, essas regras vencem as variantes e o `:hover` | e2e `home.spec.ts` "RNF-01: todo ponto de Tab mostra o anel de foco âmbar" (desktop e celular): percorre a página com Tab e exige estilo com foco diferente do sem foco e com o âmbar; e2e "RNF-01: dá para trocar o dia e filtrar só com o teclado". Medição própria: 30/30 pontos em 1280 px e 23/23 em 390 px com anel (Selects com borda e halo âmbar) | O teste aceita o âmbar translúcido do halo dos campos; para "Criar lobby" e o dia selecionado ele pega a regressão, porque o estilo sem foco desses dois não muda com o foco sem a correção |
| RNF-02 | ✅ | `@fontsource`, Lucide empacotado, `static/brand/` | e2e "RNF-02"; `ui.spec.ts` RNF-02 | — |
| RNF-03 | ✅ | — | Os testes novos do ciclo citam RNF-01; os demais citam IDs | Os nomes em `status.spec.ts` citam IDs da fundação (CA-07.1, CA-05.4), herdados de lá |

## Pendências para correção
- Nenhuma.

## Scope creep
- Nenhum. As alterações desde o ciclo 2 se ligam ao RNF-01. Os itens no limite do escopo dos ciclos anteriores continuam aceitáveis: "Entrar com Discord" só com ícone abaixo de 900 px, o filtro `app` da CI que também dispara com `.env.example` e `scripts/smoke-app.sh`, e a edição do `CLAUDE.md` prevista na T-10.
- Nada toca o "Fora de escopo".

## Observações (não bloqueantes)
- **"Ver grupo" fora da medição de foco:** o destaque só aparece quando há lobby futuro com vaga, e às 22:58 não havia nenhum. Por isso o botão secundário "Ver grupo" não entrou no teste de Tab nem na minha medição. A regra `.btn.btn:focus-visible` vale para todas as variantes, então o risco é baixo. Mesmo assim, o teste depende do horário em que roda; fixar o relógio no e2e deixaria a cobertura estável.
- **Gaveta aberta:** os controles dentro da gaveta em 390 px não passaram pelo teste de Tab (ela fica fechada). São os mesmos componentes da barra lateral, que passaram.
- **CA-01.2 sem regressão automatizada:** nem o smoke nem a CI provocam uma migração que falha. Um passo no `smoke-app.sh` com `-p` próprio e override do `command` protegeria a RN-02.
- **CA-02.2:** o teste de tela só confere o selo "Lotado". Vale conferir também a ausência de "Candidatar" no card lotado e a classe `full`.
- **RN-22:** as classes (Arcebispo, Paladino, Feiticeiro, Sicário, Andarilho) são nomes de classe do jogo, e o `<meta name="description">` de `+page.svelte` cita "Ragnarok Online". Vêm do design aprovado e são uso descritivo, mas vale o usuário confirmar que a RN-22 os admite.
- **Tablist dos dias:** `role="tab"` sem `tabpanel` e sem navegação por setas, com 14 pontos de Tab. Funciona pelo teclado, mas não segue o padrão ARIA de abas.
- **Destaque ignora filtros:** o código está certo, mas nenhum teste aplica filtro e confere que o destaque continua.
- **`go test` em cache:** os pacotes Go saíram `(cached)`, o que é esperado, já que o backend não mudou desde o ciclo 2; a CI rodou o job Backend do zero e passou.
- **Rastreabilidade:** os commits estão em Conventional Commits e citam IDs (`c260186` cita RNF-01). As tasks marcadas como feitas batem com o `git log`.
- **Ambiente:** a pilha `app` e o PostgreSQL de dev continuam no ar. Os servidores do Playwright foram encerrados (portas 8080 e 4173 livres). Nenhum arquivo versionado foi alterado além deste relatório.
