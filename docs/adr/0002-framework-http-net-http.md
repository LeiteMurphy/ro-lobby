# ADR-02 — Framework HTTP do backend: biblioteca padrão (`net/http`)

- Status: Aceito
- Data: 2026-09-29

## Contexto
O backend é em Go. A API precisa de rotas com método e parâmetro de caminho, de
middlewares simples (autenticação, log, recuperação de pânico) e de handlers fáceis de
testar, porque toda regra de domínio tem teste. O projeto é solo, então cada
dependência a mais é uma coisa a mais para manter.

## Opções consideradas
- (a) Biblioteca padrão (`net/http`). Desde o Go 1.22, o `ServeMux` aceita método e
  parâmetro na rota (`GET /lobbies/{id}`).
- (b) `chi`: roteador leve, compatível com `net/http`.
- (c) Gin ou Echo: frameworks completos, com contexto próprio.
- (d) Fiber: baseado em `fasthttp`, não é compatível com `net/http`.

## Decisão
(a) Biblioteca padrão (`net/http`).

## Consequências
- \+ Nenhuma dependência externa para roteamento. Os handlers são testados direto
  com `net/http/httptest`.
- \+ Se faltar alguma coisa (grupos de rota, middlewares prontos), o `chi` usa os
  mesmos handlers, então a migração é pequena.
- − Middlewares e agrupamento de rotas são escritos à mão.
- − Validação e binding de JSON não vêm prontos.
