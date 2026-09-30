# Research — Fundação

- Feature: `fundacao`
- Data: 2026-09-29

## Objetivo
Deixar o repositório pronto para receber a primeira feature de produto: esqueleto do
backend e do web, banco local, migrações, testes e CI, seguindo os ADRs já aprovados.

## Achados
- Decisões aceitas: SvelteKit (ADR-01), `net/http` (ADR-02), monorepo com
  `backend/`, `web/`, `docs/adr/` e `.specs/` (ADR-03). Também está decidido: Go no
  backend, PostgreSQL com datas em UTC e login com Discord.
- Ainda não há código, `.gitignore`, CI nem configuração de ferramentas.
- A spec de candidatura (`.specs/candidatura-lobby/`) está aprovada e já na `main`.
  Ela pede, no RNF-04, teste automatizado para toda regra citando o ID do critério,
  e, no RNF-03, operações atômicas no banco.
- Hospedagem e mobile continuam pendentes e não bloqueiam a fundação. A única
  restrição conhecida é que o preview no Discord exige renderização no servidor
  (ADR-01), então o SvelteKit vai precisar de um adaptador de servidor.

## Ambiente local (Windows 11)
- Node 24.14.0 e npm 11.9.0 estão instalados.
- **Go não está instalado.**
- **Docker e PostgreSQL não estão instalados.**
- Por isso, a fundação precisa documentar o setup local, e é preciso decidir se o
  Postgres local roda por Docker ou por instalação nativa.

Atualização de 2026-09-29, durante a implementação: Go 1.27.0 e Docker Desktop 4.93.0
(engine 29.8.1, WSL 2) instalados. O Smart App Control foi desligado porque
bloqueava os binários de teste do Go (ver "Descobertas" em `tasks.md`).

## Decisões técnicas que ainda faltam (candidatas a ADR)
- ADR-04 — Acesso ao banco e migrações (proposta: `pgx` + `sqlc` + `goose`).
- ADR-05 — Contrato da API (proposta: OpenAPI, com cliente TypeScript gerado).

## Classificação: G (Grande)
É uma entrega técnica, sem regra de negócio, mas com decisões de arquitetura
(ADR-04 e ADR-05), várias camadas (backend, web, banco, CI) e impacto em todas as
features seguintes.
