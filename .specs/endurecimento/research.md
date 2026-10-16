# Research — Endurecimento de segurança

## Pedido
O repositório é público e a v1 roda no PC do usuário, publicada em `rolobby.com.br` pelo
Cloudflare Tunnel (ADR-08), com 10 usuários reais. O usuário quer "um .env bom e seguro, com
demais seguranças para os usuários". Ficam fora daqui o e-mail pessoal nos commits antigos e
a página de privacidade com exclusão de conta, que vira uma feature separada.

## O que já está bem
- O `.env` nunca foi commitado. O token do bot, o Client Secret e o token do túnel não
  aparecem em nenhuma revisão do histórico (busca pelo valor exato, sem imprimir).
- O banco da pilha `app` não publica porta. A API (8080) também não. O túnel só alcança
  `web:3000` pela rede interna do Compose.
- Os containers rodam sem root: a API usa `distroless:nonroot` e o web usa `USER node`.
- Os cookies são `HttpOnly`, `SameSite=Lax` e `Secure` fora de localhost. Sessões são
  guardadas como hash, e o token do Discord é descartado depois do uso (RN-02 da
  discord-bot).
- A API não registra nome nem ID de usuário no log. O site não coleta e-mail nem senha.
- O `npm audit --omit=dev` encontrou 0 vulnerabilidades.
- O SvelteKit confere a origem das actions (CSRF), com `ORIGIN` igual a `https://rolobby.com.br`.

## Achados
| # | Achado | Risco | Onde |
|---|---|---|---|
| A-01 | A `POSTGRES_PASSWORD` do `.env` é a mesma do `.env.example` público (`ro_lobby_dev`), e dev e produção usam a mesma variável | Médio: hoje o banco de produção não está exposto, mas qualquer brecha de rede usa a senha conhecida | `.env.example`, `docker-compose.yml` (âncora `x-postgres`) |
| A-02 | O Postgres de dev publica `0.0.0.0:5432`, aberto para a rede local com a senha pública | Médio: qualquer aparelho no mesmo Wi-Fi entra no banco de dev | `docker-compose.yml`, serviço `postgres` |
| A-03 | O web publica `0.0.0.0:3000`, aberto para a rede local, sem passar pelo Cloudflare | Baixo: a rede local acessa a produção por fora do túnel | `docker-compose.yml`, serviço `web` |
| A-04 | A API foi compilada com Go 1.27.0, que tem 11 vulnerabilidades da biblioteca padrão alcançáveis (`net/http`, http2, `crypto/tls`, `html/template`), todas corrigidas no 1.27.2. O Dockerfile usa a tag móvel `golang:1.27-alpine` | Médio: a correção depende de quando a imagem foi baixada | `backend/go.mod`, `backend/Dockerfile` |
| A-05 | O web não envia cabeçalhos de segurança: nem CSP, nem `X-Frame-Options`/`frame-ancestors`, nem `Referrer-Policy`, `X-Content-Type-Options` ou `Permissions-Policy` | Médio: clickjacking, e nenhuma defesa extra contra XSS | `web/src/hooks.server.ts`, `vite.config.ts` |
| A-06 | A API não limita o tamanho do corpo das requisições (não há `MaxBytesReader`) e só define `ReadHeaderTimeout` | Baixo/médio: um corpo gigante consome memória | `backend/cmd/api/main.go` |
| A-07 | Não há limite de requisições nas rotas sensíveis (login, candidatura, avisar) | Baixo/médio: abuso e spam. O Cloudflare pode limitar na borda | — |
| A-08 | O GitHub está com secret scanning, push protection e alertas do Dependabot desligados (são gratuitos em repositório público), e a `main` não tem proteção de branch | Médio: um segredo commitado por engano não seria barrado nem avisado | Configurações do repositório |
| A-09 | A CI não roda `govulncheck` nem `npm audit`, e não existe `dependabot.yml` | Baixo/médio: uma vulnerabilidade nova passa despercebida (o A-04 é o exemplo) | `.github/` |
| A-10 | A aplicação usa o superusuário do Postgres (`POSTGRES_USER`) | Baixo: uma injeção de SQL hipotética teria poder total no banco. O `sqlc` já parametriza as queries | `docker-compose.yml` |
| A-11 | Os backups (`~/ro-lobby-backups/*.sql`) guardam dados dos usuários em texto puro, sem retenção definida | Baixo: o risco é local, mas é dado pessoal (LGPD) | Processo de publicação |

## Classificação: **M (Médio)**
Mexe em várias camadas (Compose, `.env`, web, API, CI e configuração do GitHub), mas não cria
modelo de dados nem regra de domínio. As decisões técnicas são poucas e cabem na spec.
Fases: Specify → Tasks → Execute → Validate. A troca da senha de produção exige um passo
operacional (`ALTER ROLE` no volume existente, com backup antes).
