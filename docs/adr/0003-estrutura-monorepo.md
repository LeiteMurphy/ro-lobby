# ADR-03 — Estrutura do repositório: monorepo

- Status: Aceito
- Data: 2026-09-29

## Contexto
O projeto tem backend em Go (ADR-02), web em SvelteKit (ADR-01) e, no futuro, um app
mobile. O desenvolvimento é solo e orientado a spec: uma feature como a candidatura
muda backend, web e spec juntos, e o validador confere tudo numa branch só.

## Opções consideradas
- (a) Monorepo, com uma pasta por aplicação.
- (b) Um repositório por aplicação (backend, web, mobile).

## Decisão
(a) Monorepo, com esta estrutura:

```
backend/     módulo Go (API)
web/         SvelteKit
docs/adr/    decisões de arquitetura
.specs/      specs por feature
```

A pasta `mobile/` entra quando o mobile for decidido.

## Consequências
- \+ Uma feature é uma branch, um PR e um relatório de validação, com commits atômicos
  que cruzam backend, web e spec.
- \+ Specs, ADRs e código ficam versionados juntos.
- − A CI precisa rodar só o que mudou (por caminho) para não ficar lenta.
- − As ferramentas de cada linguagem convivem na raiz (`go.mod` em `backend/`,
  `package.json` em `web/`).
