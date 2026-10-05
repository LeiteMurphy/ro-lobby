# ADR-07 — Login com Discord: OAuth na API e sessão por cookie

- Status: Aceito
- Data: 2026-10-01
- Spec: `.specs/login-discord/spec.md`

## Contexto
O login com Discord está decidido (CLAUDE.md). Na pilha `app`, só o web é exposto e o
navegador não fala com a API (RN-03 da `home-local`). O web renderiza no servidor e chama
a API pela rede interna. O fluxo OAuth precisa de um Client Secret, que não pode chegar ao
navegador, e de um lugar para guardar quem está logado.

## Opções consideradas
- (a) O web recebe o retorno do Discord, confere o `state` e repassa o `code` para a API.
  A API troca o código, cria ou atualiza o Usuário e cria a Sessão. O web guarda o token
  de sessão num cookie `HttpOnly` e o envia à API em cada requisição feita no servidor.
- (b) O web faz o OAuth inteiro (com o Client Secret no servidor do web) e só envia os
  dados do usuário para a API.
- (c) Expor a API ao navegador e fazer o OAuth direto entre o navegador e a API.
- Para a sessão: token opaco guardado como hash no banco, ou JWT assinado sem estado.

## Decisão
(a), com sessão por token opaco: o banco guarda o hash SHA-256 do token, e o cookie
`HttpOnly` leva o token. O web envia o token à API num header de autorização. O `state`
do OAuth fica num cookie `HttpOnly` de vida curta no web.

## Consequências
- \+ O Client Secret e as regras de Usuário e Sessão ficam no backend em Go, junto do
  resto do domínio.
- \+ A API continua fora do alcance do navegador (ADR-06, RN-03 da `home-local`).
- \+ Sair encerra a sessão de verdade, apagando a linha no banco, o que um JWT sem estado
  não permite.
- − Toda página que precisa saber quem está logado faz uma chamada à API no servidor do
  web.
- − O web e a API dependem de um contrato de autenticação a mais no `openapi.yaml`.
