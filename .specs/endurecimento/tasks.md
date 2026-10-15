# Tasks — Endurecimento de segurança

- Spec: `./spec.md` · Research: `./research.md` · Notion: https://app.notion.com/p/3fad4a3a5eff81e1a52de5a23728a9ec

## Ambiente e `.env`

### T-01 — Senha própria da produção e verificador do `.env`  [x]
- Cobre: US-01, US-08, RN-01, RN-02, RN-03, RN-04, RN-05, CA-01.1, CA-01.2, CA-01.3, CA-01.4,
  CA-08.1, CA-08.2
- Depende de: —
- Paralelizável: não
- Arquivos: `scripts/check-env.sh`, `docker-compose.yml` (âncora `x-postgres` separada para
  a pilha `app`, `APP_POSTGRES_PASSWORD`, serviço `env-check` antes do `migrate`),
  `.env.example`, `scripts/check-env.test.sh`, `.github/workflows/ci.yml`
- Pronto quando: o `check-env.sh` roda no host (lendo o `.env`) e no serviço `env-check`
  (lendo o ambiente); os testes do script cobrem os CAs com `.env` de exemplo, de origem
  pública com senha de exemplo, curta, Client Secret de exemplo e bot pela metade, e
  confirmam que nenhum valor sai na saída; o smoke da pilha `app` na CI continua verde com
  `cp .env.example .env`
- Commit: `0133b91 feat: senha própria da produção e verificador do .env`

### T-02 — Portas só em 127.0.0.1  [P] [x]
- Cobre: US-02, RN-06, RN-07, CA-02.1, CA-02.2 (parte versionada)
- Depende de: —
- Paralelizável: sim
- Arquivos: `docker-compose.yml`, `scripts/check-compose-tunnel.sh`
- Pronto quando: `docker compose config` mostra `host_ip: 127.0.0.1` no `postgres` e no
  `web`; o `check-compose-tunnel.sh` falha se alguma porta for publicada fora de
  `127.0.0.1` ou se `app-postgres`/`api` publicarem porta
- Commit: `1e074de feat: portas do banco de dev e do web só em 127.0.0.1`

## Dependências

### T-03 — Go 1.27.2 fixo  [P] [x]
- Cobre: US-03, RN-08, CA-03.1
- Depende de: —
- Paralelizável: sim
- Arquivos: `backend/go.mod`, `backend/Dockerfile`, `.github/workflows/ci.yml` (versão do Go)
- Pronto quando: o `govulncheck ./...` não relata vulnerabilidade alcançável; o Dockerfile usa
  `golang:1.27.2-alpine`; os testes do backend passam
- Commit: `7e997d7 fix(backend): Go 1.27.2 fixo, sem vulnerabilidade conhecida da biblioteca padrão`

### T-04 — Verificação de vulnerabilidades na CI e Dependabot  [x]
- Cobre: US-03, RN-09, RN-10, CA-03.2, CA-03.3
- Depende de: T-03
- Paralelizável: não
- Arquivos: `.github/workflows/ci.yml`, `.github/dependabot.yml`
- Pronto quando: os jobs `backend` e `web` rodam `govulncheck` e
  `npm audit --omit=dev --audit-level=high`, e o `ci-ok` depende deles; o `dependabot.yml`
  cobre gomod, npm, github-actions e docker, toda semana
- Commit: `d59d7d3 ci: govulncheck, npm audit e Dependabot semanal`

## Web

### T-05 — Cabeçalhos de segurança  [P] [x]
- Cobre: US-04, RN-11, RN-12, CA-04.1, CA-04.2
- Depende de: —
- Paralelizável: sim
- Arquivos: `web/src/hooks.server.ts`, `web/src/lib/security/headers.ts` (+ spec de teste),
  teste e2e
- Pronto quando: os testes unitários citam CA-04.1/CA-04.2 (com https, com HSTS; com
  localhost, sem HSTS); o e2e confere os cabeçalhos na Home e numa rota 404
- Commit: `06869f1 feat(web): cabeçalhos de segurança em toda resposta`

## API

### T-06 — Limite de corpo e timeouts do servidor  [P] [x]
- Cobre: US-05, RN-13, RN-14, CA-05.1, CA-05.6
- Depende de: —
- Paralelizável: sim
- Arquivos: `backend/internal/server/limits.go` (+ teste), `backend/internal/server/server.go`,
  `backend/cmd/api/main.go`
- Pronto quando: um teste envia 65 KiB e recebe 413 sem gravar nada (CA-05.1); um teste
  confere os quatro timeouts do `http.Server` (CA-05.6)
- Commit: `ff384f3 feat(backend): limite de 64 KiB no corpo e prazos do servidor`

### T-07 — Limite de requisições por IP e por usuário  [x]
- Cobre: US-05, RN-15, RN-16, CA-05.2 (API), CA-05.3, CA-05.4, CA-05.5
- Depende de: T-06
- Paralelizável: não
- Arquivos: `backend/internal/ratelimit/` (janela por chave, em memória, com relógio
  injetável), `backend/internal/server/ratelimit.go`, `backend/internal/config/config.go`
  (`API_RATE_LIMIT=off` só no e2e), `web/src/hooks.server.ts` (`handleFetch` repassa o IP do
  visitante só nas chamadas à API), `web/src/lib/security/client-ip.ts`,
  `web/playwright.config.ts`
- Pronto quando: os testes com relógio falso cobrem 30 ações por usuário com a 31ª em 429 e
  `Retry-After`, outro usuário livre, o login com 10 por IP, a liberação depois da janela e o
  limite geral de 300 por IP; o IP repassado pelo web é o usado na contagem
- Commit: `ec63e48 feat: limite de requisições por IP e por sessão`

### T-08 — Mensagem de "muitas tentativas" no web  [x]
- Cobre: US-05, RN-15, CA-05.2 (web), CA-05.3 (web)
- Depende de: T-07
- Paralelizável: não
- Arquivos: `web/src/lib/api/request.ts`, pontos que exibem erro das actions, retorno do login
- Pronto quando: um 429 da API vira o tipo `rate_limited` e as actions mostram "Muitas
  tentativas. Aguarde um pouco e tente de novo."; um 429 no retorno do login leva ao aviso
  de login com falha; os testes citam os CAs
- Commit: `dd2bb8d feat(web): mensagem de muitas tentativas quando a API responde 429`

## GitHub e operação

### T-09 — Proteções do repositório no GitHub  [x]
- Cobre: US-06, RN-17, CA-06.1, CA-06.2
- Depende de: T-04 (o `ci-ok` precisa estar completo antes de virar obrigatório)
- Paralelizável: não
- Arquivos: nenhum versionado; registro do comando `gh api` aplicado no `tasks.md`
- Pronto quando: `gh api repos/LeiteMurphy/ro-lobby` mostra secret scanning, push protection
  e as atualizações do Dependabot como `enabled`; a proteção da `main` exige o `ci-ok`,
  bloqueia force push e exclusão e não exige revisão
- Commit: — (configuração do GitHub, aplicada em 2026-10-15 com `gh api`):
  - `PATCH repos/LeiteMurphy/ro-lobby` com `secret_scanning` e
    `secret_scanning_push_protection` em `enabled`;
  - `PUT .../vulnerability-alerts` e `PUT .../automated-security-fixes`;
  - `PUT .../branches/main/protection`: check obrigatório `CI ok`, pull request com 0
    aprovações, sem force push, sem exclusão, `enforce_admins` desligado.
  - Conferência: `dependabot_security_updates=enabled secret_scanning=enabled
    secret_scanning_push_protection=enabled`; alertas 204; `checks=["CI ok"] reviews=0
    force_push=false deletions=false`.

### T-10 — Backup com retenção  [P] [x]
- Cobre: US-07, RN-18, CA-07.1
- Depende de: —
- Paralelizável: sim
- Arquivos: `scripts/backup.sh`, `scripts/backup.test.sh`
- Pronto quando: o teste com 10 backups falsos cria um novo e sobra 10, sem o mais antigo; a
  saída mostra só o nome do arquivo e a contagem
- Commit: `dccfdce feat: backup do banco da pilha com retenção dos 10 mais recentes`

### T-11 — Seção de segurança no README  [x]
- Cobre: US-08, RN-05, RN-19, CA-08.3
- Depende de: T-01, T-02, T-10
- Paralelizável: não
- Arquivos: `README.md`
- Pronto quando: o README tem: gerar segredos fortes, rodar o `check-env.sh`, backup antes de
  publicar, trocar a senha de produção (`ALTER ROLE`), o que fazer se um segredo vazar e o
  checklist do Cloudflare
- Commit: `21e177d docs: seção de segurança no README, com segredos, backup, troca de senha e Cloudflare`

### T-12 — Troca da senha de produção e publicação  [ ]
- Cobre: US-01, US-02, RN-19, CA-01.5, CA-02.2 (parte da produção)
- Depende de: merge do PR (memória "publicar só depois do merge")
- Paralelizável: não
- Arquivos: nenhum versionado (`.env` local)
- Pronto quando: o backup foi feito; a senha forte foi gerada no `.env` sem passar pelo chat;
  o `ALTER ROLE` foi aplicado; a pilha subiu pela `main` com o túnel; https://rolobby.com.br
  responde 200 com os mesmos usuários; o web publica em `127.0.0.1:3000`
- Commit: — (operação; fica fora do relatório do validador, conferida no roteiro)

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-01 |
| CA-01.2 | T-01 |
| CA-01.3 | T-01 |
| CA-01.4 | T-01 |
| CA-01.5 | T-12 |
| CA-02.1 | T-02 |
| CA-02.2 | T-02, T-12 |
| CA-03.1 | T-03 |
| CA-03.2 | T-04 |
| CA-03.3 | T-04 |
| CA-04.1 | T-05 |
| CA-04.2 | T-05 |
| CA-05.1 | T-06 |
| CA-05.2 | T-07, T-08 |
| CA-05.3 | T-07, T-08 |
| CA-05.4 | T-07 |
| CA-05.5 | T-07 |
| CA-05.6 | T-06 |
| CA-06.1 | T-09 |
| CA-06.2 | T-09 |
| CA-07.1 | T-10 |
| CA-08.1 | T-01 |
| CA-08.2 | T-01 |
| CA-08.3 | T-11 |

## Descobertas
- 2026-10-15 — O início do login só redireciona para o Discord e não chega à API. O limite
  de login passou a contar a troca do código (RN-15, CA-05.3 ajustados na spec).
- 2026-10-15 — T-01: o `check-env.sh` também recusa `APP_POSTGRES_PASSWORD` com caractere
  fora de letras, números, `-` e `_`, porque a senha vai dentro da `DATABASE_URL` e um `@` ou
  `/` quebraria a conexão. É uma defesa da RN-01/RN-02, sem regra nova de produto.
- 2026-10-15 — T-01: com a pilha `app`, o Compose passa a exigir `APP_POSTGRES_PASSWORD` em
  qualquer comando (inclusive no modo dev), porque ele confere as variáveis de todos os
  serviços do arquivo. O `.env` do operador recebe a variável na T-12.
- 2026-10-15 — T-07: o IP vem do `CF-Connecting-IP` lido pelo próprio web, e não do
  `ADDRESS_HEADER` do adapter-node. Com o `ADDRESS_HEADER`, toda requisição local sem o
  cabeçalho (como o healthcheck do container) quebraria. O web só é alcançado pelo túnel ou
  pelo próprio PC (RN-06), e o Cloudflare sobrescreve o cabeçalho, então ele é confiável.
- 2026-10-15 — T-07: "ações por usuário" contam por sessão (hash do token Bearer), sem ir ao
  banco a cada requisição. Quem está logado em dois aparelhos tem dois limites de 30/min.
- 2026-10-15 — T-07: o e2e desliga o limite com `API_RATE_LIMIT=off`, porque faz centenas de
  requisições do mesmo IP. A API se recusa a subir com essa opção e `APP_ORIGIN` https.
  O logout (`DELETE /session`) e o `/healthz` nunca contam.
- 2026-10-15 — T-07/T-08: o 429 do limite (`{"error":"rate_limited"}`) é uma resposta
  transversal, fora do `openapi.yaml`. A sala do Discord já usava 429 para o intervalo do
  aviso (`cooldown`); o web separa os dois pelo campo `error`.
- 2026-10-15 — T-10: o backup usa `docker exec` direto no container, e não
  `docker compose exec`, para funcionar mesmo antes de o `.env` ter `APP_POSTGRES_PASSWORD`
  (RN-19).
- 2026-10-15 — T-01: o `check-env.sh` também recusa o `DISCORD_CLIENT_ID` do exemplo e
  cobra o `CLOUDFLARE_TUNNEL_TOKEN` com origem pública (só no host, porque o token não vai
  ao `env-check`). As duas checagens decorrem da RN-04 ("variável obrigatória faltando,
  valor de exemplo em uso com origem pública").
- 2026-10-15 — Validação, ciclo 1: os arquivos estáticos (`/_app/*`, `/robots.txt`,
  imagens) saíam sem os cabeçalhos da RN-11, porque o adapter-node entrega esses arquivos
  antes do hook do SvelteKit. Correção na T-05: `web/server.js` envolve o handler do
  adapter-node, e os valores ficam num arquivo só (`web/security-headers.js`), usado pelo
  hook e pelo servidor. O Dockerfile sobe com `node server.js`, e o `smoke-app.sh` confere
  os cabeçalhos num estático da imagem real.
