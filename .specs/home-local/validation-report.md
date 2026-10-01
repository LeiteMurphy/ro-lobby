# Relatório de validação — home-local (ciclo 4)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend `go test -count=1 ./...` ok, integração `-tags=integration -count=1` ok, Vitest 68/68 em 8 arquivos, Playwright 16/16, `scripts/smoke-app.sh` 14/14 verificações) · lint ok (`gofmt -l` vazio, `golangci-lint` 0 issues, Prettier + ESLint ok, `svelte-check` 0 erros/0 avisos em 418 arquivos) · build ok (`go build ./...`, `npm run build`; imagens reconstruídas pelo smoke). `npm run generate` não gerou diff. CI do PR #11 (run 36933901971, head `3ba9024`): Filtro de caminhos, Backend, Web, Ponta a ponta, Pilha app e CI ok, todos verdes.

Intervalo validado: `origin/main..feature/home-local` (18 commits, 8c77cb0..3ba9024). Desde o ciclo 3 (`0007789..3ba9024`) mudaram só arquivos do web e da spec: a revisão da RN-22 em `spec.md` (linha "Última revisão" e texto da regra), uma descoberta em `tasks.md`, o novo catálogo `web/src/lib/catalog/classes.ts` com o teste `classes.spec.ts`, o `classArt` de `web/src/lib/home/catalog.ts` passando a ler o catálogo, as classes dos dados fictícios em `fixtures.ts`, 13 ícones Lucide novos em `Icon.svelte` e o nome do teste RN-22 em `routes/home.spec.ts`. Backend, Compose, Dockerfiles, CI e scripts não mudaram.

Conferi a lista do catálogo contra a página https://browiki.org/wiki/Classes em 2026-10-01: são os mesmos 82 títulos, com a mesma grafia e na mesma ordem que o `BROWIKI_PAGES` do teste. O HTML servido pela pilha `app` em `localhost:3000` mostra as classes novas de hoje (Guardião Real, Musa) e não mostra mais "Guardiã" nem "Andarilho". Também reproduzi o CA-01.2 por conta própria (ver detalhe).

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
| RN-01 | ✅ | `docker-compose.yml`: `app-postgres` → `migrate` (`service_healthy`) → `api` (`service_completed_successfully`) → `web` (`service_healthy`), builds de produção | smoke CA-01.1 (executado); job "Pilha app" verde | — |
| RN-02 | ✅ | `api.depends_on.migrate: service_completed_successfully` | Execução própria: projeto `-p rolobby-ca012` com override do `command` do `migrate`; `migrate` saiu com 1 e o container da API ficou em `Created`, sem iniciar. Projeto removido depois | Sem teste automatizado (ver Observações) |
| RN-03 | ✅ | `ports` só no `web`, por `APP_WEB_PORT`; API e `app-postgres` sem porta; `API_BASE_URL: http://api:8080` | smoke CA-01.3; `docker ps`: só `0.0.0.0:3000` na pilha app | — |
| RN-04 | ✅ | Postgres de dev em `profiles: [dev]` + `COMPOSE_PROFILES=dev` (D-03) | smoke CA-01.4 | — |
| RN-05 | ✅ | `backend/Dockerfile` (distroless nonroot), `web/Dockerfile` (estágio `prod-deps`, usuário `node`) | smoke CA-01.5 (4 verificações) | — |
| RN-06 | ✅ | `.env.example`, `README.md` | Inspeção; sem mudança desde os ciclos anteriores | — |
| RN-07 | ✅ | `.github/workflows/ci.yml`: job `app`, filtro de caminhos, `ci-ok.needs` | Run 36933901971: "Pilha app" e "CI ok" verdes | — |
| RN-08 | ✅ | `fixtures.ts` `getHomeLobbies` (l. 86) é o único acesso; o catálogo de classes é só leitura de apresentação | `home.spec.ts` CA-02.1 | O `fixtures.ts` agora importa `CLASSES`; o isolamento continua, já que o resto do web segue chamando só `getHomeLobbies` |
| RN-09 | ✅ | `time.ts` `TIME_ZONE`, `zonedNow` | `home.spec.ts:43`, `:50` | — |
| RN-10 | ✅ | `days.ts` `buildDays` | `home.spec.ts:54`, `:112`; `routes/home.spec.ts:67` | — |
| RN-11 | ✅ | `lobbies.ts` `lobbiesForDay` | `home.spec.ts:60`, `:70`; e2e CA-02.1, CA-03.2 | — |
| RN-12 | ✅ | `filters.ts` `matches` (l. 51) | `home.spec.ts:170` a `:205` | — |
| RN-13 | ✅ | `roleCounts` (l. 70), `timeRangeCounts` (l. 77) | `home.spec.ts:211`, `:221` | — |
| RN-14 | ✅ | `lobbies.ts` `openSlots`, `isFull`, `edgeRole`; `LobbyCard.svelte` `.card.full` | `home.spec.ts:75`, `:82`, `:86`, `:91` | A cor da borda continua vindo da composição; a função sugerida do catálogo só pinta o ícone da classe |
| RN-15 | ✅ | `time.ts` `relativeLabel` + atualização por minuto | `home.spec.ts:96`, `:100`, `:105` | — |
| RN-16 | ✅ | `lobbies.ts` `featuredLobby` (l. 50) | `home.spec.ts:249`, `:257`, `:262`, `:269` | — |
| RN-17 | ✅ | `filters.ts` `emptyState`, `EmptyState.svelte` | `home.spec.ts:134`, `:225`; e2e CA-03.3, CA-04.7 | — |
| RN-18 | ✅ | prop `soon` no Button/IconButton | `ui.spec.ts`; `routes/home.spec.ts:44`; e2e CA-02.5 | — |
| RN-19 | ✅ | `Drawer.svelte`, CSS < 900 px, `activeFilterCount` | e2e CA-06.1, CA-06.2, RN-19; `home.spec.ts:238` | — |
| RN-20 | ✅ | `+page.server.ts` | `routes/home.spec.ts:22`; e2e CA-02.6 | — |
| RN-21 | ✅ | `web/src/routes/status/` | `status/page.spec.ts`; e2e `status.spec.ts`; smoke "API online" | — |
| RN-22 (revista) | ✅ | Sem logos, artes ou sprites do jogo: assets em `web/static/brand/` são do design system, ícones de classe são Lucide. Classes: `catalog/classes.ts` com os 82 nomes do bROWiki; `fixtures.ts` só usa nomes do catálogo ("Guardiã" → Guardião Real, "Andarilho" → Musa); instâncias continuam inventadas (`catalog.ts`). Texto da tela sem "!" (conferido no HTML servido) | `classes.spec.ts`: "tem exatamente as classes da página do bROWiki" (82, comparado com a lista que conferi na página), "os dados fictícios só usam classes do catálogo" (e mais de 20 classes nos 14 dias), "as classes do design mantêm o ícone e a cor de função"; `routes/home.spec.ts:72` (sem Gravity e nomes de mapa ou monstro) | Ver Observações (CLAUDE.md e servidor LATAM) |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | perfil `app` | smoke CA-01.1 (migrate 0, três serviços saudáveis, `/` com cards, `/status` "API online"); CI "Pilha app" | — |
| CA-01.2 | ✅ | `depends_on … service_completed_successfully` | Execução própria neste ciclo (ver RN-02) | A falha provocada foi por argumento inválido do `migrate`, não por SQL quebrado; o efeito no Compose é o mesmo (saída ≠ 0) |
| CA-01.3 | ✅ | só `web.ports` | smoke CA-01.3; `docker ps` | — |
| CA-01.4 | ✅ | perfil `dev` | smoke CA-01.4 | Indireto (`config --services`) |
| CA-01.5 | ✅ | Dockerfiles multi-stage | smoke CA-01.5 | — |
| CA-01.6 | ✅ | `.env.example`, `README.md` | Inspeção | Nenhuma variável nova neste ciclo |
| CA-02.1 | ✅ | `+page.svelte`, `LobbyCard.svelte`, `HostLine.svelte`, `RoleComposition.svelte` | `home.spec.ts:60`; `routes/home.spec.ts:28`; e2e CA-02.1 | — |
| CA-02.2 | ✅ | `LobbyCard.svelte` `{#if full}`, `.card.full` | `home.spec.ts:75`; `routes/home.spec.ts:36` | O teste de tela confere só o selo |
| CA-02.3 | ✅ | `edgeRole` | `home.spec.ts:82` | — |
| CA-02.4 | ✅ | `relativeLabel`, `zonedNow` | `home.spec.ts:96`, `:100`; `routes/home.spec.ts:40` | — |
| CA-02.5 | ✅ | prop `soon` (`aria-disabled` + dica) | `ui.spec.ts`; `routes/home.spec.ts:44`; e2e CA-02.5 | — |
| CA-02.6 | ✅ | `+page.server.ts` | `routes/home.spec.ts:22`; e2e CA-02.6 | — |
| CA-02.7 | ✅ | `routes/status/` | e2e `status.spec.ts`; smoke | — |
| CA-03.1 | ✅ | `buildDays`, `DaySelector.svelte` | `home.spec.ts:112`; `routes/home.spec.ts:67` | — |
| CA-03.2 | ✅ | `+page.svelte` (título) | e2e CA-03.2; e2e RNF-01 (teclado) | Passou com as classes rotativas dos outros dias |
| CA-03.3 | ✅ | `emptyState`, `EmptyState.svelte` | `home.spec.ts:134`; e2e CA-03.3 | — |
| CA-04.1 | ✅ | `matches` | `home.spec.ts:170`; e2e CA-04.1 | — |
| CA-04.2 | ✅ | `matches` (`some`) | `home.spec.ts:177` | — |
| CA-04.3 | ✅ | `matches` (`<=`) | `home.spec.ts:186` | — |
| CA-04.4 | ✅ | `TIME_RANGES`, faixa com fim excluído | `home.spec.ts:192`, `:196` | — |
| CA-04.5 | ✅ | `matches` | `home.spec.ts:205` | — |
| CA-04.6 | ✅ | `roleCounts`, `timeRangeCounts` | `home.spec.ts:211`, `:221` | — |
| CA-04.7 | ✅ | `EmptyState.svelte`, `NO_FILTERS` | `home.spec.ts:225`; e2e CA-04.7 | — |
| CA-04.8 | ✅ | `countLabel` | `home.spec.ts:232` | — |
| CA-05.1 | ✅ | `featuredLobby` | `home.spec.ts:249`; `routes/home.spec.ts:62` | — |
| CA-05.2 | ✅ | `featuredLobby` | `home.spec.ts:257`, `:262` | — |
| CA-05.3 | ✅ | `{#if featured}` | `home.spec.ts:269` | — |
| CA-06.1 | ✅ | `Drawer.svelte`, CSS < 900 px | e2e CA-06.1 | — |
| CA-06.2 | ✅ | `activeFilterCount` | e2e CA-06.2; `home.spec.ts:238` | — |
| CA-07.1 | ✅ | job `app` + `ci-ok.needs` | Run 36933901971 construiu as duas imagens; CI ok verde | A metade "falha deixa o CI ok vermelho" vem da lógica do `ci-ok`; nenhuma falha real foi provocada |

### Não funcionais
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RNF-01 | ✅ | Anel global em `tokens/base.css`; regras `:focus-visible` no Button, IconButton e DaySelector (sem mudança desde o ciclo 3) | e2e "RNF-01: todo ponto de Tab mostra o anel de foco âmbar" (desktop e celular) e "RNF-01: dá para trocar o dia e filtrar só com o teclado"; `home.spec.ts:127` | O ícone de classe não é interativo; nada novo a focar |
| RNF-02 | ✅ | Os 13 ícones novos vêm de `@lucide/svelte/icons/*`, empacotados (`Icon.svelte`) | `classes.spec.ts` "RNF-02: todo ícone do catálogo está empacotado no build"; e2e RNF-02 | — |
| RNF-03 | ✅ | — | Os 6 testes novos de `classes.spec.ts` citam RN-22 ou RNF-02; os demais citam IDs | `status.spec.ts` cita IDs da fundação, herdados de lá |

## Pendências para correção
- Nenhuma.

## Scope creep
- Nenhum bloqueante. No limite do escopo: o catálogo `classes.ts` traz campos que a Home não usa (`plural`, `tier`, `family`) e a `tasks.md` diz que ele "serve de base para o cadastro de personagens", que está no "Fora de escopo". O catálogo não implementa cadastro nenhum. O `plural` é o que permite o teste contra o bROWiki, `tier` escolhe as classes dos dados fictícios e `family` entra no teste das linhas. O usuário pediu o catálogo. Por isso não marquei 🚫.
- Os itens no limite dos ciclos anteriores continuam aceitáveis: "Entrar com Discord" só com ícone abaixo de 900 px, o filtro `app` da CI que também dispara com `.env.example` e `scripts/smoke-app.sh`, e a edição do `CLAUDE.md` prevista na T-10.

## Observações (não bloqueantes)
- **CLAUDE.md contra a RN-22 revista:** o `CLAUDE.md` (l. 67) ainda diz "Não usar logos, artes nem nomes oficiais da Gravity". Os nomes de classe do bRO são nomes oficiais da localização do jogo, e a RN-22 agora os permite. A spec foi revista a pedido do usuário, então a validação segue a spec. Mesmo assim, as duas fontes se contradizem: vale o usuário ajustar o `CLAUDE.md` (por exemplo, "salvo os nomes de classe, conforme a RN-22 da home-local") para a próxima feature não reprovar ou desfazer isso. O `<meta name="description">` citando "Ragnarok Online" deixou de ser dúvida perante a RN-22, mas também esbarra nessa linha do `CLAUDE.md`.
- **bRO ou LATAM:** o servidor em foco é o Ragnarok LATAM (CLAUDE.md), e a RN-22 adota a nomenclatura do bRO. Se os nomes em português do LATAM forem diferentes de algum nome do bROWiki, o cadastro de personagens vai precisar decidir qual vale. Não afeta esta entrega.
- **"Hoje continua igual ao design":** a descoberta da `tasks.md` e o commit `e57a884` dizem isso, mas hoje não está mais idêntico ao Home v2, porque o `07ef774` trocou "Guardiã" por Guardião Real e "Andarilho" por Musa (ícone e cor de função mantidos, e o teste confere isso). O §10 da spec ("iguais aos do design") cede à RN-22 revista. Vale deixar a frase da `tasks.md` mais exata.
- **Teste do catálogo autorreferente:** a lista `BROWIKI_PAGES` foi copiada à mão para o teste. Neste ciclo eu a conferi contra a página. O teste não pega uma mudança futura no bROWiki, o que é esperado e está documentado no comentário (data de consulta).
- **Classificação de nível:** o teste das linhas espera 3 classes de 4ª para o Arqueiro (Falcão do Vento, Maestro, Diva) e 2 para as outras famílias. Está correto. Bardo e Odalisca aparecem como 2ª separadas, o que também bate com o bROWiki.
- **Ordem dos imports em `Icon.svelte`:** os 13 ícones novos entraram antes do `ChevronDown`, fora da ordem alfabética do resto. É só estética; o lint passou.
- **CA-01.2 sem regressão automatizada:** continua só com execução manual. Um passo no `smoke-app.sh` com `-p` próprio e override do `command` (como fiz neste ciclo) protegeria a RN-02.
- **CA-02.2:** o teste de tela só confere o selo "Lotado"; a ausência de "Candidatar" e a classe `full` não são verificadas na tela.
- **Destaque ignora filtros e "Ver grupo" no teste de foco:** seguem como no ciclo 3. Nenhum teste aplica filtro e confere o destaque, e o e2e de foco depende do horário em que roda.
- **Rastreabilidade:** os 3 commits novos estão em Conventional Commits e citam IDs (`[US-02, RN-22]` e `[US-01..US-07]`). As tasks marcadas como feitas batem com o `git log`.
- **Ambiente:** a pilha `app` e o PostgreSQL de dev continuam no ar. O projeto `rolobby-ca012` que subi para o CA-01.2 foi removido com volume e rede. Os servidores do Playwright foram encerrados (portas 8080 e 4173 livres). Nenhum arquivo versionado foi alterado além deste relatório.
