# Relatório de validação — endurecimento (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes passaram (backend `go test ./...` ok em todos os pacotes; integração
`-tags=integration` ok; web Vitest 455/455 em 42 arquivos; e2e Playwright 68/68;
`check-env.test.sh`, `backup.test.sh` e `check-compose-tunnel.sh` "Tudo certo") · lint ok
(golangci-lint 0 issues; Prettier + ESLint ok; svelte-check 0 erros, 0 avisos) · build ok
(`go build ./...`; `npm run build`) · govulncheck v1.9.0 "No vulnerabilities found" ·
`npm audit --omit=dev --audit-level=high` 0 vulnerabilidades · CI do PR #35 (head
`08e52b6`, igual ao HEAD local): Filtro de caminhos, Web e Pilha app verdes (o passo "Teste de
fumaça" passou, com a nova conferência dos cabeçalhos num estático da imagem real); Backend e
Ponta a ponta ainda em andamento no momento da validação.

A pendência bloqueante do ciclo 1 (RN-11/RN-12 nos arquivos estáticos) foi corrigida e
conferida executando o servidor de produção (`node server.js`). A RN-15 e o CA-05.2 foram
revisados na spec e agora batem com o código. Os itens de produção (CA-01.5, RN-19 e a parte de
produção do CA-02.2) continuam **pendentes de operação pós-merge (T-12)**, por decisão
registrada no tasks.md; não são falha da implementação e não entram no veredito.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 17 | 2 (RN-15 P2, RN-19*) | 0 | 0 |
| Critérios (CA) | 22 | 2 (CA-01.5*, CA-02.2*) | 0 | 0 |
| Não funcionais | 3 | 0 | 0 | 0 |

\* pendente de operação pós-merge (T-12), não é falha de implementação.

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | docker-compose.yml: `app-postgres.environment.POSTGRES_PASSWORD: ${APP_POSTGRES_PASSWORD:?...}`; `x-app-database-url` com `APP_POSTGRES_PASSWORD`; dev mantém `POSTGRES_PASSWORD` | check-compose-tunnel.sh ("app-postgres usa APP_POSTGRES_PASSWORD", "a API conecta com APP_POSTGRES_PASSWORD", "a senha de dev não entra na pilha app"); check-env.test.sh "RN-01 — APP_POSTGRES_PASSWORD obrigatória" | |
| RN-02 | ✅ | scripts/check-env.sh bloco `public = 1` (< 24 caracteres; igual ao `.env.example`; mensagens só com o nome); serviço `env-check` (sem rede) como dependência de `app-postgres` | check-env.test.sh CA-01.1/01.2/01.3 e "Modo ambiente"; check-compose-tunnel.sh "env-check recusa a senha de exemplo com origem pública, sem mostrar o valor" | Também recusa `DISCORD_CLIENT_ID` de exemplo e cobra `CLOUDFLARE_TUNNEL_TOKEN` no host; agora registrado nas Descobertas |
| RN-03 | ✅ | check-env.sh: checagens de exemplo e tamanho só com `https://` | check-env.test.sh CA-01.4; check-compose-tunnel.sh "env-check aceita o .env.example local"; CI "Pilha app" verde com `cp .env.example .env` | |
| RN-04 | ✅ | check-env.sh (obrigatórias, exemplo, curta, bot pela metade, `exit 1`, imprime só nome + problema) | check-env.test.sh CA-08.1, CA-08.2 (3 problemas, código ≠ 0, nenhum valor na saída) | |
| RN-05 | ✅ | check-env.sh sem eco de valor; backup.sh grava o dump direto em arquivo e só imprime nome/contagem; client-ip.ts só repassa o IP à origem da API | check-env.test.sh `no_values`; backup.test.sh "a saída não mostra dados de usuário"; client-ip.spec.ts | |
| RN-06 | ✅ | docker-compose.yml `postgres.ports: 127.0.0.1:${POSTGRES_PORT:-5432}:5432`; `web.ports: 127.0.0.1:${APP_WEB_PORT:-3000}:3000` | check-compose-tunnel.sh "postgres de dev em 127.0.0.1", "web em 127.0.0.1", "nenhuma porta fora de 127.0.0.1" | Containers em execução só mudam quando recriados (T-12 para a pilha; o postgres de dev pelo operador) |
| RN-07 | ✅ | docker-compose.yml sem `ports` em `app-postgres` e `api` | check-compose-tunnel.sh "app-postgres sem porta publicada", "api sem porta publicada"; smoke-app.sh na CI | |
| RN-08 | ✅ | backend/go.mod `go 1.27.2`; backend/Dockerfile `golang:1.27.2-alpine` | govulncheck executado: "No vulnerabilities found" | |
| RN-09 | ✅ | .github/workflows/ci.yml: passo `govulncheck` no job Backend e `npm audit --omit=dev --audit-level=high` no job Web; `ci-ok` depende dos jobs | Comandos executados localmente; código de saída ≠ 0 derruba o passo | Só rodam quando o filtro de caminhos detecta mudança (ver Observações) |
| RN-10 | ✅ | .github/dependabot.yml: gomod `/backend`, npm `/web`, github-actions `/`, docker `/backend` e `/web`, todos `weekly` | Inspeção (CA-03.3) | |
| RN-11 | ✅ | web/security-headers.js (valores únicos); web/server.js envolve o `handler` do adapter-node com `withHeaders`; web/src/hooks.server.ts `withSecurityHeaders`; web/Dockerfile `CMD ["node", "server.js"]` e copia os dois arquivos | headers.spec.ts (CA-04.1, inclusive "o handler envolvido responde um estático com os cabeçalhos"); e2e seguranca.spec.ts (Home, 404); smoke-app.sh "cabeçalhos em / /robots.txt e num /_app/immutable/*.js" verde na CI; **execução** de `node server.js` (ORIGIN https): `/`, `/robots.txt`, `/_app/immutable/entry/app.*.js`, `/brand/bg-ceu-pixel.svg`, `/nao-existe` (404) e `/lobbies/<id>` (503 sem API) com os quatro cabeçalhos, sem duplicação | Corrigido desde o ciclo 1 |
| RN-12 | ✅ | security-headers.js `headersFor(https)`; server.js usa `ORIGIN` começando com `https://`; o hook usa `event.url.protocol` | headers.spec.ts CA-04.2 (https com HSTS; localhost sem; estático sem https sem HSTS); e2e confere ausência em localhost; **execução**: ORIGIN https → `max-age=31536000` em páginas e estáticos; ORIGIN `http://localhost:3996` → sem HSTS em `/` e `/robots.txt` | |
| RN-13 | ✅ | backend/internal/server/limits.go `limitBody` (ContentLength e `MaxBytesReader`), `requestError` → 413; server.go aplica antes das rotas | limits_test.go TestLimitBody_CA05_1_DeclaredTooLarge, _ChunkedTooLarge, TestLimitBody_RN13_WithinLimitPasses | |
| RN-14 | ✅ | limits.go `NewHTTPServer` (5/15/30/120 s); cmd/api/main.go usa | TestNewHTTPServer_CA05_6_Timeouts | |
| RN-15 | ⚠️ (P2) | backend/internal/server/ratelimit.go: `General` 300/IP (exceto `/healthz`), `Login` 10/IP em `POST /auth/discord`, `Actions` 30 por hash do token Bearer em todo método que não é GET/HEAD/OPTIONS, exceto `/auth/discord` e `/session`; 429 + `Retry-After` + `{"error":"rate_limited"}`; web: request.ts e discord/api.ts → `rate_limited` → `unavailableMessage` nas actions | ratelimit_test.go CA-05.2/05.3/05.4/05.5, RN15_HealthzExempt, RN15_LogoutNotAnAction, RN15_EveryMutationCounts (inclui o "Avisar no Discord"), TestNew_RN15_RateLimitsOption; request.spec.ts; lobbies.server.spec.ts; discord.spec.ts; auth.spec.ts | Agora bate com a spec revisada (por sessão, toda mutação logada). Lacuna menor: quando o **limite geral** barra o carregamento de uma página (GET no SSR), o web mostra a mensagem genérica de indisponibilidade (`UNAVAILABLE_MESSAGE` em `+page.server.ts` da Home, lobby, editar, perfil, talentos, candidaturas), e não "Muitas tentativas...". A regra diz que, passado o limite, o web mostra essa frase. Os CAs só cobrem ações e login, que estão certos. P2, não bloqueia |
| RN-16 | ✅ | web/src/lib/security/client-ip.ts (`CF-Connecting-IP` ou endereço da conexão); hooks.server.ts `handleFetch` só para a origem da API; ratelimit.go `clientIP` (`X-Client-IP` válido ou `RemoteAddr`) | client-ip.spec.ts (CA-05.5, IP forjado substituído, RN-05); TestRateLimit_CA05_5_ForwardedIPIsTheKey; TestClientIP_RN16_FallbackToRemoteAddr | A aceitação só da rede interna é garantida pela topologia (RN-07), como a própria regra prevê |
| RN-17 | ✅ | Configuração do GitHub (fora do diff), registrada na T-09 | `gh api`: `secret_scanning`, `secret_scanning_push_protection`, `dependabot_security_updates` = enabled; `vulnerability-alerts` 204; `automated-security-fixes` `{"enabled":true}`; proteção da main: checks `["CI ok"]`, revisões 0, force push false, exclusão false | `enforce_admins` desligado (ver Observações) |
| RN-18 | ✅ | scripts/backup.sh (`pg_dump` via `docker exec` para arquivo temporário, retenção `keep=10`, saída só com nome e contagem) | backup.test.sh CA-07.1, rótulo, dump com falha | `backup.sh` real não executado (restrição de ambiente) |
| RN-19 | ⚠️* | README.md "Trocar a senha do banco" (backup antes, `\password` = ALTER ROLE, atualizar o `.env`) | — | Pendente de operação pós-merge (T-12) |
| CA-01.1 | ✅ | `env-check` como `depends_on` (`service_completed_successfully`) do `app-postgres`; `migrate` depende do banco | check-compose-tunnel.sh (`docker compose run --rm --no-deps env-check` com APP_ORIGIN https: recusa, cita a variável, sem o valor); check-env.test.sh CA-01.1 | |
| CA-01.2 | ✅ | check-env.sh "tem menos de 24 caracteres" | check-env.test.sh CA-01.2 (23 recusa e cita 24; 24 passa) | |
| CA-01.3 | ✅ | check-env.sh compara `DISCORD_CLIENT_SECRET` com o exemplo | check-env.test.sh CA-01.3 | |
| CA-01.4 | ✅ | check-env.sh sem checagens de exemplo fora de https | check-env.test.sh CA-01.4; CI "Pilha app" (smoke com `cp .env.example .env`) verde | smoke-app.sh não executado localmente (restrição); evidência pela CI |
| CA-01.5 | ⚠️* | — | — | Pendente de operação pós-merge (T-12) |
| CA-02.1 | ✅ | docker-compose.yml `127.0.0.1:${POSTGRES_PORT:-5432}` | check-compose-tunnel.sh "postgres de dev em 127.0.0.1" | O container de dev em execução precisa ser recriado para refletir |
| CA-02.2 | ⚠️* | docker-compose.yml web em 127.0.0.1; app-postgres/api sem porta | check-compose-tunnel.sh (parte versionada ok) | Parte de produção (porta real 127.0.0.1:3000 e https://rolobby.com.br 200) pendente da T-12 |
| CA-03.1 | ✅ | go.mod 1.27.2; Dockerfile `golang:1.27.2-alpine` | govulncheck executado sem vulnerabilidades | |
| CA-03.2 | ✅ | ci.yml: passos nos jobs Backend e Web, que alimentam o `CI ok` | Inspeção do workflow; comandos executados | |
| CA-03.3 | ✅ | .github/dependabot.yml | Inspeção | |
| CA-04.1 | ✅ | hooks.server.ts + server.js + security-headers.js | headers.spec.ts; e2e seguranca.spec.ts (Home, rota inexistente); execução de `node server.js` incluindo uma página de lobby | |
| CA-04.2 | ✅ | security-headers.js / headers.ts | headers.spec.ts; e2e (sem HSTS em localhost); execução com ORIGIN https e com ORIGIN http | |
| CA-05.1 | ✅ | limits.go | TestLimitBody_CA05_1_* (serviços nil provam que nada chega à rota) | |
| CA-05.2 | ✅ | ratelimit.go `actionKey` (hash do token = sessão); request.ts; `unavailableMessage` nas actions | TestRateLimit_CA05_2_ActionsPerSession (31ª = 429, `Retry-After: 60`, rota recebeu só 30); _OtherSessionAndReadsUnaffected; lobbies.server.spec.ts (mensagem exata); request.spec.ts; discord.spec.ts | Bate com o texto revisado ("por sessão") |
| CA-05.3 | ✅ | ratelimit.go Login 10/IP; callback do web → `/?login=erro` | TestRateLimit_CA05_3_LoginPerIP (11ª 429, outro IP passa); auth.spec.ts CA-05.3 | |
| CA-05.4 | ✅ | ratelimit.Limiter (janela fixa) | TestRateLimit_CA05_4_ReleasedAfterWindow; TestAllow_CA05_4_WindowResets | |
| CA-05.5 | ✅ | client-ip.ts; ratelimit.go `clientIP`; api sem porta | TestRateLimit_CA05_5_ForwardedIPIsTheKey; client-ip.spec.ts; check-compose-tunnel.sh "api sem porta publicada" | |
| CA-05.6 | ✅ | limits.go | TestNewHTTPServer_CA05_6_Timeouts | |
| CA-06.1 | ✅ | GitHub | `gh api repos/LeiteMurphy/ro-lobby -q .security_and_analysis` (enabled); `automated-security-fixes` enabled | |
| CA-06.2 | ✅ | GitHub | `gh api .../branches/main/protection` (CI ok obrigatório, 0 revisões, sem force push/exclusão) | |
| CA-07.1 | ✅ | scripts/backup.sh | backup.test.sh | |
| CA-08.1 | ✅ | check-env.sh | check-env.test.sh CA-08.1 | |
| CA-08.2 | ✅ | check-env.sh | check-env.test.sh CA-08.2 | |
| CA-08.3 | ✅ | README.md seção "Segurança" (segredos fortes, check-env, backup antes de publicar, segredo vazado no Discord e no Cloudflare, Full / Always Use HTTPS / regra em `/auth/`) | Inspeção | |
| RNF-01 | ✅ | `.env.example` fictício; e2e e CI sem segredos reais | CI (Web, Pilha app verdes); e2e local 68/68 | |
| RNF-02 | ✅ | backend/internal/ratelimit em memória, sem dependência nova | TestAllow_RNF02_SweepsExpired | |
| RNF-03 | ✅ | Mensagens novas só no 429; um 429 em `/me` mantém o cookie (vira `unavailable`, não `none`) | e2e 68/68 | O e2e roda com `API_RATE_LIMIT=off` |

## Pendências para correção
1. [RN-15, P2, não bloqueante] Quando o limite geral (300/min por IP) barra o carregamento
   de uma página, o web mostra a mensagem genérica de indisponibilidade em vez de "Muitas
   tentativas. Aguarde um pouco e tente de novo.". Pontos: os `loadError`/`error(503, ...)`
   em `web/src/routes/+page.server.ts`, `lobbies/[id]/+page.server.ts`,
   `lobbies/[id]/editar/+page.server.ts`, `lobbies/novo/+page.server.ts`,
   `perfil/+page.server.ts`, `talentos/+page.server.ts` e `candidaturas/+page.server.ts`
   (bastaria usar `unavailableMessage(result)`). Alternativa: registrar na spec que a frase
   vale só para ações.
2. [CA-01.5, RN-19, CA-02.2 parte de produção] Pendente de operação pós-merge (T-12), por
   decisão registrada: backup, senha nova no `.env`, `ALTER ROLE`, subida pela `main` com o
   web em `127.0.0.1:3000` e https://rolobby.com.br respondendo 200 com os mesmos dados.
   Não é falha da implementação.
3. [CI] Os jobs Backend e Ponta a ponta do PR #35 ainda estavam rodando; confirmar o `CI ok`
   verde antes do merge (localmente, tudo passou).

## Scope creep
- Nada que toque "Fora de escopo" (sem CSP, sem papel de banco reduzido, sem LGPD, sem
  reescrita de histórico).
- Extensões pequenas ligadas às regras: `check-env.sh` recusa `DISCORD_CLIENT_ID` de exemplo,
  cobra `CLOUDFLARE_TUNNEL_TOKEN` no host e recusa caractere especial na senha (todas
  registradas nas Descobertas); `backup.sh` aceita rótulo opcional; o README recomenda
  *Bot Fight Mode*. Nenhuma é bloqueante.

## Observações (não bloqueantes)
- Com o `.env` atual do operador (sem `APP_POSTGRES_PASSWORD`), qualquer `docker compose`
  falha até a T-12, inclusive no modo dev (registrado nas Descobertas). Fazer a T-12 logo
  depois do merge, e recriar também o postgres de dev para valer o `127.0.0.1`.
- `web/server.js` só é exercitado pelo smoke da pilha (CI) e pelo teste unitário do
  `withHeaders`; o e2e usa `vite preview`. Uma regressão no `server.js` só aparece no job
  "Pilha app".
- `enforce_admins` desligado na proteção da `main`: o dono (único admin) consegue ignorar o
  `CI ok` num push direto. A spec não pede, mas a US-06 fala em "a main só receber código com
  a CI verde".
- `govulncheck` e `npm audit` só rodam quando `backend/` ou `web/` mudam; uma vulnerabilidade
  nova sem mudança de código só aparece pelos alertas do Dependabot. Um `schedule` semanal na
  CI fecharia isso.
- O Dependabot de docker cobre os Dockerfiles, mas não as imagens do `docker-compose.yml`
  (`postgres:18-alpine`, `cloudflared`).
- O limite geral conta as chamadas do SSR à API (`/me` mais as da página), então várias
  pessoas atrás do mesmo IP chegam ao limite antes de "300 páginas". O e2e roda com o limite
  desligado, então isso só aparece em uso real.
- A API aceita `X-Client-IP` sem conferir a origem da conexão; a proteção depende de a API
  continuar sem porta publicada (RN-07).
