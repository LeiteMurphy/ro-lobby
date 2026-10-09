# Relatório de validação — publicacao-tunel (ciclo 1)

**Veredito geral:** APROVADO (CA-01.4 pendente de UAT manual)
**Execução:** testes `bash scripts/check-compose-tunnel.sh` 11/11 ok, saída 0 · `bash scripts/smoke-app.sh --down` todos os checks ok, saída 0 ("Pilha app ok em http://localhost:3000") · lint `bash -n` nos dois scripts sem erro (não há outro lint aplicável) · build feito pelo smoke, sem erro

Verificações extras feitas pelo validador:
- `docker manifest inspect cloudflare/cloudflared:2026.9.2`: a tag existe no Docker Hub.
- `docker compose ... -f docker-compose.tunnel.yml --profile app config` com `CLOUDFLARE_TUNNEL_TOKEN=SEGREDO-XYZ`: o valor aparece uma única vez, em `services.tunnel.environment.TUNNEL_TOKEN`. Nenhum serviço usa `env_file`, então o token não vaza para a API, o web, o migrate nem o banco.
- `docker compose --profile dev config --services` com o `.env.example`: só `postgres` (modo dev intacto).
- `.gitignore` ignora `.env` e `.env.*` (exceto `.env.example`); `git log --all -- .env` vazio; o único arquivo `.env*` versionado é o `.env.example`, com valores comentados e fictícios.
- O `.env` local do usuário não define `APP_ORIGIN` nem `CLOUDFLARE_TUNNEL_TOKEN`, então o smoke rodou no modo local padrão.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 5 | 0 | 0 | 0 |
| Critérios (CA) | 4 | 0 (+1 pendente de UAT) | 0 | 0 |
| Não funcionais | 1 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `docker-compose.yml:91` `ORIGIN: ${APP_ORIGIN:-http://localhost:${APP_WEB_PORT:-3000}}` | `check-compose-tunnel.sh` blocos CA-02.1 (padrão `http://localhost:3000`) e CA-01.1 (`https://rolobby.com.br`) | O `redirect_uri` (`web/src/routes/auth/discord/login/+server.ts:24`, `callback/+server.ts:30`) e o `Secure` do cookie (`oauth.ts:80`) derivam de `url.origin`, que segue o `ORIGIN` no adapter-node. Nenhuma mudança de código necessária, como previsto na Research. |
| RN-02 | ✅ | `docker-compose.tunnel.yml:11-21`: serviço `tunnel`, `profiles: [app]`, `image: cloudflare/cloudflared:2026.9.2`, `depends_on.web.condition: service_healthy`, `restart: unless-stopped` | `check-compose-tunnel.sh` bloco CA-01.2 (regex de versão fixa, depends_on, restart) e CA-02.1 (sem `tunnel` sem o arquivo) | Tag conferida no Docker Hub. |
| RN-03 | ✅ | `docker-compose.tunnel.yml:17` `TUNNEL_TOKEN: ${CLOUDFLARE_TUNNEL_TOKEN:?defina CLOUDFLARE_TUNNEL_TOKEN no .env}`; `.env.example:11-16` com a variável comentada e valor fictício | `check-compose-tunnel.sh` blocos CA-01.2 ("token só no tunnel") e CA-01.3 (falha com a mensagem) | Conferido também pelo valor no `config` completo (ver acima). |
| RN-04 | ✅ | `docker-compose.tunnel.yml`: sem `ports`, rota `web:3000` documentada no README; `docker-compose.yml`: só `web` tem `ports` na pilha app | `check-compose-tunnel.sh` ("tunnel sem porta publicada", "só o web publica porta"); `smoke-app.sh` (api e app-postgres sem porta) | O alvo `http://web:3000` é configurado no painel do Cloudflare (túnel gerenciado por token), não no repositório; está no README passo 2. |
| RN-05 | ✅ | `README.md` seção "Publicar para testes (túnel)": DNS (registro.br → Cloudflare), túnel e public hostname `HTTP` + `web:3000`, redirect `https://rolobby.com.br/auth/discord/callback`, `.env`, limites (PC ligado, sem backup), comandos de subir/parar | Revisão manual | Também avisa da restrição de CSRF em `localhost:3000`. |
| CA-01.1 | ✅ | `docker-compose.yml:91` | `check-compose-tunnel.sh` "ORIGIN vem de APP_ORIGIN" (assere o valor exato) | — |
| CA-01.2 | ✅ | `docker-compose.tunnel.yml` | `check-compose-tunnel.sh` bloco CA-01.2: imagem fixa, sem porta, `service_healthy`, `unless-stopped`, só `web` publica porta | — |
| CA-01.3 | ✅ | `docker-compose.tunnel.yml:17` | `check-compose-tunnel.sh` bloco CA-01.3: o `config` falha e a saída contém "defina CLOUDFLARE_TUNNEL_TOKEN no .env" | — |
| CA-01.4 | Pendente de UAT | Configuração e README permitem cumprir: ORIGIN público, redirect e cookie derivados da origem, túnel por token apontando para `web:3000` | Manual (depende das contas no Cloudflare e no Discord) | Não reprova, conforme a spec. |
| CA-02.1 | ✅ | `docker-compose.yml:91` (padrão); túnel só em arquivo à parte | `check-compose-tunnel.sh` bloco CA-02.1 (com cópia do `.env.example`); `smoke-app.sh --down` passou; CI roda os dois em sequência (`.github/workflows/ci.yml`, step "Configuração do túnel" antes do "Teste de fumaça") | — |
| RNF-01 | ✅ | `.gitignore:2-4`; `.env.example` só com placeholder comentado | `git ls-files`, `git log --all -- .env` | Nenhum segredo no diff. |
| Fora de escopo | ✅ | Nada de VPS/nuvem, backup, monitoramento, Cloudflare Access, `www` ou automação dos painéis | Revisão do diff | — |
| ADR-08 (decisões) | ✅ | Túnel em `docker-compose.tunnel.yml`, origem por `.env`, sem mudança de código da aplicação | Diff só toca configuração, scripts, CI e docs | — |

## Pendências para correção
Nenhuma.

## Scope creep
Nenhum. As alterações fora dos arquivos de configuração estão ligadas à feature: `CLAUDE.md` e `docs/adr/0006-hospedagem-local.md` registram o ADR-08 (commit de docs com `[RN-02, RN-03]`), e `.github/workflows/ci.yml` passa a rodar `check-compose-tunnel.sh` (CA-01.1, CA-01.2, CA-01.3, CA-02.1) e a disparar o job quando os novos arquivos mudam.

## Observações (não bloqueantes)
1. Em `scripts/check-compose-tunnel.sh`, o laço "token só no tunnel" confere só a chave `TUNNEL_TOKEN` nos outros serviços, e imprime `ok` mesmo quando algum `fail` foi registrado no laço (o `exit 1` no fim ainda pega a falha). Uma checagem mais forte procuraria o valor do token em todo o `config` (o validador fez isso manualmente e passou). Também a lista de serviços está fixa (`api web migrate app-postgres`); um serviço novo não seria coberto.
2. O teste de configuração depende de `node` no PATH para ler o JSON; na CI isso existe hoje, mas fica implícito.
3. A rastreabilidade dos commits está ok: os três commits citam IDs da spec.
4. Há um diretório não rastreado `.specs/compartilhar-lobby/` no working tree, que não faz parte desta branch nem desta validação.
