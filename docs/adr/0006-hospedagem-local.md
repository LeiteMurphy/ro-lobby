# ADR-06 — Hospedagem: local por enquanto

- Status: Aceito, revisado pelo [ADR-08](0008-publicacao-tunel-cloudflare.md) (publicação
  para testes fechados)
- Data: 2026-09-30
- Spec: `.specs/home-local/spec.md` (US-01, US-07)

## Contexto
O CLAUDE.md deixa a hospedagem pendente de ADR. O projeto ainda não tem login, lobbies
nem candidatura, então não há o que publicar para jogadores. O que falta hoje é rodar a
pilha inteira como num servidor (build de produção, migrações antes da API, web com SSR
chamando a API pela rede interna) e ter a Home do Claude Design no ar.

O login com Discord (OAuth) vai exigir uma URL pública de callback. É ele que torna a
hospedagem externa necessária.

## Opções consideradas
- (a) Local: um perfil do Docker Compose sobe PostgreSQL, migrações, API e web com as
  imagens de produção, em `localhost`.
- (b) Nuvem agora, num PaaS com container (API e web) e PostgreSQL gerenciado.
- (c) VPS própria com o mesmo Compose.

## Decisão
(a) Local por enquanto. As imagens e o Compose são os mesmos que (b) ou (c) vão usar
depois, então a decisão de onde publicar fica para quando houver produto.

**Revisão:** quando a spec do login com Discord for aprovada, ou antes, se alguém de fora
precisar acessar.

## Consequências
- \+ Sem custo e sem conta em provedor.
- \+ As imagens multi-stage e a ordem migração → API → web ficam prontas e testadas na CI.
- \+ O modo de desenvolvimento da fundação continua igual (perfil padrão do Compose).
- − Ninguém de fora acessa. O preview no Discord (motivo do SSR no ADR-01) não dá para
  testar ainda.
- − Backup, logs e monitoramento ficam para a revisão.
