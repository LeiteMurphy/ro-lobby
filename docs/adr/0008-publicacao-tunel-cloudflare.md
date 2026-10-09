# ADR-08 — Publicação para testes fechados: PC local com Cloudflare Tunnel

- Status: Aceito
- Data: 2026-10-09
- Spec: `.specs/publicacao-tunel/spec.md`
- Revisa: [ADR-06](0006-hospedagem-local.md)

## Contexto
O ADR-06 deixou a hospedagem local e marcou a revisão para quando alguém de fora
precisasse acessar. Esse momento chegou: o usuário comprou `rolobby.com.br` e quer que
amigos próximos testem o RO Lobby, com login do Discord. Ainda não é lançamento para
jogadores, então custo e esforço importam mais que disponibilidade. O login com Discord
exige uma URL pública com HTTPS para o callback.

## Opções consideradas
- (a) A pilha `app` continua no PC do usuário, e um Cloudflare Tunnel (`cloudflared` no
  Compose) publica o web em `rolobby.com.br`. O DNS fica no Cloudflare, que cuida do
  HTTPS.
- (b) Abrir a porta 443 no roteador, com IP dinâmico, DNS dinâmico e certificado no PC.
- (c) VPS com o mesmo Compose (ex.: Hetzner, ~€5/mês), com ou sem o túnel.
- (d) Web no Cloudflare Pages (`adapter-cloudflare`) e API numa VPS.

## Decisão
(a) para a fase de testes fechados. Não tem custo, não abre porta nem expõe o IP de casa
e reaproveita as imagens e o Compose do ADR-06 sem mudar o código da aplicação. A mudança
é só de configuração: a origem do web passa a vir do `.env` e o túnel entra num arquivo
Compose à parte (`docker-compose.tunnel.yml`), que o modo local e a CI não leem.

(b) foi descartada por expor a rede de casa. (d) troca o adapter do web e separa os
domínios por um ganho que os testes não precisam. (c) é o caminho natural depois: o
túnel funciona igual numa VPS.

**Revisão:** antes de divulgar para jogadores fora do grupo de amigos, ou se o PC ligado
virar problema. A revisão escolhe entre (c) e uma nuvem e decide backup e monitoramento.

## Consequências
- \+ `rolobby.com.br` com HTTPS e login do Discord, sem custo.
- \+ Nenhuma porta aberta no roteador; a API e o banco seguem só na rede interna.
- \+ O modo local e a CI não mudam: sem `APP_ORIGIN` e sem o arquivo do túnel, tudo fica
  como antes.
- − O site só fica no ar com o PC ligado, sem hibernar e com o Docker rodando.
- − Os dados ficam no volume Docker do PC, sem backup. Perder o volume perde os testes.
- − Com `APP_ORIGIN` público, os formulários só funcionam pelo domínio, não por
  `localhost:3000`.
- − O tráfego passa pelo Cloudflare, que vira dependência da publicação.
