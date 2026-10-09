# Tasks — Banco de talentos

- Spec: `./spec.md` · Design: `./design.md`
- Notion: [Épico](https://app.notion.com/p/3f4d4a3a5eff817fbd09f468784d45a5)
- Branch: `feature/banco-de-talentos`

## US-01 — Pôr o personagem no banco  (P1)

### T-01 — Tabela de disponibilidade e contrato da API  [ ]
- Cobre: RN-01 a RN-04, D-01, D-02, D-03, RNF-01
- Depende de: —
- Paralelizável: não
- Arquivos: `backend/migrations/00007_character_availability.sql`,
  `backend/queries/talents.sql`, `backend/internal/db/` (gerado), `openapi.yaml`,
  `backend/internal/api/api.gen.go` e `web/src/lib/api/schema.gen.ts` (gerados)
- Pronto quando: a migração aplica e reverte; testes de integração provam os `CHECK`
  (dias 1..127, minutos de 30 em 30, início diferente do fim, instância ou "Qualquer") e o
  CASCADE na exclusão do personagem; o contrato tem as quatro rotas e os schemas do design.
- Commit:
- Notion: https://app.notion.com/p/3f4d4a3a5eff81e8a9dfc4f1a2f80f02

### T-02 — Serviço e rota de disponibilidade  [ ]
- Cobre: RN-01 a RN-05, CA-01.3, CA-01.4, CA-01.5, D-06
- Depende de: T-01
- Paralelizável: não
- Arquivos: `backend/internal/talents/`, `backend/internal/server/talents.go`,
  `backend/internal/characters/` (campo `availability` no `Character`)
- Pronto quando: testes provam o upsert, os 422 por campo, o 404 para personagem de outro
  Usuário, que desligar guarda os dados e que `GET /characters` traz a disponibilidade.
- Commit:
- Notion: https://app.notion.com/p/3f4d4a3a5eff8156a2f2fc7f0b6776eb

## US-02 — Afinidade no detalhe do lobby  (P1)

### T-03 — Consulta de afinidade e rota do lobby  [ ]
- Cobre: RN-06, RN-07, RN-08, RN-09, CA-02.1, CA-02.2, CA-02.3, CA-02.4 (API), D-04
- Depende de: T-02
- Paralelizável: não
- Arquivos: `backend/queries/talents.sql`, `backend/internal/talents/`,
  `backend/internal/server/talents.go`
- Pronto quando: testes de integração provam cada condição da RN-06 (instância e
  "Qualquer", dia e faixa com virada da meia-noite e fim excluído, nível, função com vaga,
  livre na janela de 2 h, inclusive pendente), a exclusão dos personagens do dono, a ordem
  e as respostas 404 e 409 de `GET /lobbies/{id}/talents`.
- Commit:
- Notion: https://app.notion.com/p/3f4d4a3a5eff8169aa5dc932e7b800b3

### T-04 — Catálogo e contagem na API  [P] [ ]
- Cobre: RN-10, RN-11, RN-12, CA-01.1, CA-01.2, CA-03.1 (API), CA-04.1, CA-04.2 (API),
  CA-04.3, D-05, RNF-05
- Depende de: T-03
- Paralelizável: não (mesmo pacote da T-03); fica antes das tasks de tela
- Arquivos: `backend/internal/talents/`, `backend/internal/server/talents.go`
- Pronto quando: testes provam os filtros de `GET /talents` (instância, função, dia, hora
  com virada), o limite de 100, o `discordUsername` só com sessão e a contagem de
  `GET /talents/count` com as vagas do formulário menos a do dono.
- Commit:
- Notion: https://app.notion.com/p/3f4d4a3a5eff819cbc57eb0383e6d674

## Telas

### T-05 — Disponibilidade no perfil  [P] [ ]
- Cobre: US-01, CA-01.1, CA-01.3, CA-01.4, RNF-03
- Depende de: T-02
- Paralelizável: [P] com T-06 e T-07 (rotas diferentes)
- Arquivos: `web/src/routes/perfil/`, `web/src/lib/characters/` (diálogo e mensagens),
  `web/test/e2e/talentos.spec.ts`
- Pronto quando: testes de SSR e e2e provam ligar com dias, faixa e instâncias, os erros
  por campo, desligar e reabrir com os dados guardados, tudo pelo teclado.
- Commit:
- Notion: https://app.notion.com/p/3f4d4a3a5eff8139b828f74e62de3055

### T-06 — Painel "Jogadores disponíveis" e prévia da criação  [P] [ ]
- Cobre: US-02, US-03, CA-02.1, CA-02.4, CA-03.1
- Depende de: T-03, T-04
- Paralelizável: [P] com T-05 e T-07
- Arquivos: `web/src/routes/lobbies/[id]/`, `web/src/routes/lobbies/novo/`
  (`disponiveis/+server.ts`), `web/src/lib/lobbies/components/LobbyForm.svelte`,
  `web/src/lib/talents/`
- Pronto quando: testes provam o painel só para o dono do lobby aberto, com a ordem da
  API e o aviso de lista vazia, e a contagem da prévia que muda com a hora.
- Commit:
- Notion: https://app.notion.com/p/3f4d4a3a5eff81dc99bad7727d2e2d70

### T-07 — Página /talentos  [P] [ ]
- Cobre: US-04, CA-01.2, CA-04.1, CA-04.2, CA-04.3, RNF-03, RNF-05
- Depende de: T-04
- Paralelizável: [P] com T-05 e T-06
- Arquivos: `web/src/routes/talentos/`, `web/src/lib/talents/`,
  `web/src/lib/home/components/TopBar.svelte`
- Pronto quando: testes de SSR e e2e provam os filtros pela query string, o
  `@username` só para logado ("Entre para ver o Discord" para visitante, sem o nome no
  HTML), o aviso dos 100 primeiros e o link na TopBar.
- Commit:
- Notion: https://app.notion.com/p/3f4d4a3a5eff816e88eff592a441a67c

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-04, T-05 |
| CA-01.2 | T-04, T-07 |
| CA-01.3 | T-02, T-05 |
| CA-01.4 | T-02, T-05 |
| CA-01.5 | T-02 |
| CA-02.1 | T-03, T-06 |
| CA-02.2 | T-03 |
| CA-02.3 | T-03 |
| CA-02.4 | T-03, T-06 |
| CA-03.1 | T-04, T-06 |
| CA-04.1 | T-04, T-07 |
| CA-04.2 | T-04, T-07 |
| CA-04.3 | T-04, T-07 |

## Descobertas
- (nenhuma)
