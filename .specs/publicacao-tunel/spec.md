# Spec — Publicação por túnel para testes fechados

- Feature: `publicacao-tunel` · Nível: P · Status: Aprovada (2026-10-09)
- ADR: [ADR-08](../../docs/adr/0008-publicacao-tunel-cloudflare.md)
- Notion: https://app.notion.com/p/3f4d4a3a5eff8186ae1bea4576070020
- Última revisão: 2026-10-09 — RN-02 e CA-01.2/CA-01.3/CA-02.1: o túnel vai para um arquivo
  Compose à parte em vez de um perfil, aprovado pelo usuário na execução

## 1. Contexto
O usuário comprou `rolobby.com.br` e quer mostrar o RO Lobby a amigos próximos, com o PC
dele ligado como servidor, antes de decidir a hospedagem definitiva. A pilha `app`
(ADR-06) já roda como em produção e só expõe o web. Faltam dois ajustes: a origem do web
está fixa em `http://localhost:3000` no `docker-compose.yml`, e não há como alcançar o PC
de fora sem abrir porta no roteador.

## 2. Research
- O `redirect_uri` do Discord sai da origem do web (`web/src/lib/auth/oauth.ts:39`, D-08
  da `login-discord`). Com `ORIGIN=https://rolobby.com.br`, o callback vira
  `https://rolobby.com.br/auth/discord/callback` sem mudar código.
- O cookie de sessão já é `Secure` fora de localhost (`oauth.ts:80`, RN-08 da
  `login-discord`).
- O adapter-node usa `ORIGIN` para a proteção CSRF das actions. Com a origem pública, o
  acesso por `http://localhost:3000` deixa de enviar formulários: quem testa usa o
  domínio.
- O Cloudflare Tunnel (`cloudflared`) abre uma conexão de saída até o Cloudflare; não há
  porta aberta nem IP exposto. A imagem oficial é `cloudflare/cloudflared`, e o túnel
  gerenciado pelo painel precisa só de um token.
- O `.com.br` fica no registro.br; só o DNS vai para o Cloudflare.

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como dono do projeto, quero publicar a pilha do meu PC em `rolobby.com.br`, para meus amigos testarem com login do Discord. |
| US-02 | P1 | Como desenvolvedor, quero que o modo local continue igual sem o túnel, para nada mudar no dia a dia e na CI. |

## 4. Regras
- **RN-01** — A origem do web na pilha `app` vem de `APP_ORIGIN` no `.env`. Sem ela, o
  padrão continua `http://localhost:${APP_WEB_PORT}`.
- **RN-02** — O túnel é um serviço `tunnel` no arquivo `docker-compose.tunnel.yml`, com a
  imagem `cloudflare/cloudflared` numa versão fixa. Ele só existe quando esse arquivo é
  pedido (`-f docker-compose.yml -f docker-compose.tunnel.yml --profile app`), sobe depois
  que o web está saudável e reinicia sozinho. (Revisada na execução: o Compose exige as
  variáveis obrigatórias de todos os serviços do arquivo, mesmo com o perfil desligado; com
  um perfil no arquivo principal, a falta do token quebraria o modo local.)
- **RN-03** — O token do túnel (`CLOUDFLARE_TUNNEL_TOKEN`) só existe no `.env` e no
  ambiente do serviço `tunnel`. O `.env.example` traz a variável comentada, com um valor
  fictício. Subir com o arquivo do túnel sem o token falha com uma mensagem que diz o que
  falta.
- **RN-04** — O túnel aponta para `http://web:3000` pela rede interna do Compose. A API e
  o banco continuam sem porta publicada.
- **RN-05** — O README ganha a seção "Publicar para testes (túnel)" com os passos no
  Cloudflare (DNS, túnel, rota para `http://web:3000`), no Discord (redirect
  `https://rolobby.com.br/auth/discord/callback`) e no `.env`, e com os limites: fica no
  ar só com o PC ligado, sem backup.

## 5. Critérios de aceite
```gherkin
CA-01.1 — Origem configurável  [US-01, RN-01]
Given APP_ORIGIN=https://rolobby.com.br no .env
When a configuração da pilha app é montada (docker compose config)
Then o serviço web recebe ORIGIN=https://rolobby.com.br

CA-01.2 — Túnel no arquivo próprio  [US-01, RN-02, RN-04]
Given o .env com CLOUDFLARE_TUNNEL_TOKEN
When a configuração é montada com docker-compose.tunnel.yml e --profile app
Then existe o serviço tunnel, com cloudflare/cloudflared numa versão fixa, sem porta publicada
  And ele depende do web saudável e usa restart unless-stopped
  And só o web publica porta

CA-01.3 — Token obrigatório com o túnel  [US-01, RN-03]
Given o .env sem CLOUDFLARE_TUNNEL_TOKEN
When a configuração é montada com docker-compose.tunnel.yml
Then o Compose falha com "defina CLOUDFLARE_TUNNEL_TOKEN no .env"

CA-01.4 — No ar pelo domínio  [US-01, RN-01, RN-04] (UAT, manual)
Given o DNS no Cloudflare, o túnel apontando para http://web:3000 e o redirect no Discord
When a pilha sobe com o túnel e alguém de fora abre https://rolobby.com.br
Then a Home abre com os lobbies
  And o login com Discord volta logado para o site

CA-02.1 — Modo local igual  [US-02, RN-01, RN-02]
Given o .env.example copiado sem mudanças
When a pilha sobe com --profile app, sem o arquivo do túnel (como na CI)
Then o serviço tunnel não existe
  And o web recebe ORIGIN=http://localhost:3000
  And o teste de fumaça da pilha passa
```

## 6. Requisitos não funcionais
- **RNF-01** — Nenhum segredo (token do túnel, Client Secret) entra no git.

## 7. Fora de escopo
- Hospedagem definitiva (VPS ou nuvem), backup e monitoramento — revisão futura do
  ADR-08.
- Restringir o acesso com Cloudflare Access — o usuário escolheu deixar aberto, só com o
  link.
- Subdomínio `www` — dá para redirecionar no painel do Cloudflare sem mudar o repositório.
- Criar contas e configurar os painéis do Cloudflare e do Discord — o usuário faz, com o
  passo a passo do README.

## 8. Decisões tomadas na conversa
- Publicar do PC do usuário, com Cloudflare Tunnel e domínio `rolobby.com.br`.
- Acesso aberto, só com o link; login pelo Discord como hoje.
- Esta entrega vem antes da feature `servidor`.
