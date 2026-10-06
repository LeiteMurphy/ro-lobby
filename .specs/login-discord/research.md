# Research — Login com Discord

- Feature: `login-discord`
- Data: 2026-10-01

## Objetivo
Permitir que o visitante entre no RO Lobby com a conta do Discord, criando o Usuário do
domínio. É o pré-requisito de personagens, lobbies e candidatura.

## Achados
- **Discord OAuth2** (docs.discord.com/developers/topics/oauth2):
  - Fluxo *authorization code*. O usuário vai para `https://discord.com/oauth2/authorize`
    e volta para o `redirect_uri` com um `code`. O servidor troca o código em
    `https://discord.com/api/oauth2/token` (corpo `application/x-www-form-urlencoded`).
  - O parâmetro `state` é recomendado pela documentação contra CSRF. A documentação não
    fala de PKCE.
  - Escopo mínimo: `identify`, que dá `id`, `username`, `global_name` e `avatar` em
    `GET /users/@me`. O e-mail exige o escopo `email`.
  - O token de acesso dura 7 dias (604800 s) e vem com refresh token. Revogação em
    `https://discord.com/api/oauth2/token/revoke`.
  - Os `redirect_uri` precisam estar cadastrados no aplicativo do Discord Developer
    Portal, que fornece o Client ID e o Client Secret.
- **Projeto:**
  - Não há usuário, sessão nem autenticação no código. Não há tabelas de domínio.
  - A spec `candidatura-lobby` assume um "usuário logado com Discord" e um Usuário dono de
    Personagens e Lobbies.
  - Na pilha `app`, só o web é exposto (RN-03 da `home-local`); o navegador não fala com
    a API. O web chama a API no servidor (`API_BASE_URL`).
  - O botão "Entrar com Discord" já existe na Home, desabilitado (RN-18 da `home-local`).
  - O ADR-06 marca a revisão da hospedagem para quando esta spec for aprovada: em
    produção, o `redirect_uri` precisa ser uma URL pública. Em `localhost` dá para
    desenvolver e testar.
  - O contrato é o `openapi.yaml` (ADR-05); o banco usa `pgx`, `sqlc` e `goose` (ADR-04).
- **Decisão técnica que falta (candidata a ADR-07):** onde ficam o fluxo OAuth e a
  sessão.
  - (a) O web recebe o retorno do Discord e repassa o `code` para a API. A API troca o
    código, cria ou atualiza o Usuário e abre a sessão. O web guarda só o token de
    sessão num cookie HttpOnly. O Client Secret fica só na API.
  - (b) O web faz o OAuth inteiro e só manda os dados do usuário para a API.
  - (c) A API fica exposta e faz o OAuth direto com o navegador.

## Classificação: G (Grande)
Novo modelo de dados (Usuário e Sessão), integração externa, segurança (CSRF, cookie,
segredo) e uma decisão de arquitetura (ADR-07). Impacta todas as features seguintes.
