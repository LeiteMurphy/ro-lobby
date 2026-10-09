# Tasks — Grupo livre

- Spec: `./spec.md` · Design: `./design.md`
- Notion: [Épico](https://app.notion.com/p/3f4d4a3a5eff81e4b026ec70d84af725)
- Branch: `feature/grupo-livre`

## Backend

### T-01 — Formação no lobby: migração, contrato e criar/editar  [ ]
- Cobre: US-01, US-04, RN-01, RN-02, RN-06, RN-07, CA-01.1, CA-01.2, CA-04.1, CA-04.2,
  CA-04.3, RNF-01, D-01, D-02, D-04
- Depende de: —
- Paralelizável: não
- Arquivos: `backend/migrations/00008_lobby_formation.sql`, `backend/queries/lobbies.sql`,
  `backend/internal/lobbies/`, `backend/internal/server/lobbies.go`, `openapi.yaml`,
  gerados
- Pronto quando: a migração aplica e reverte com um lobby antigo no banco; testes de
  integração provam criar livre com o padrão e com 2 a 12 vagas, recusar 1 e 13, editar
  sem ficar abaixo dos ocupantes, trocar a formação sozinho e recusar com aceito ou
  pendente; `HasRoom` coberto por teste.
- Commit:

### T-02 — Candidatura e trocas no grupo livre  [ ]
- Cobre: US-02, RN-03, RN-04, RN-05, CA-02.1, CA-02.2, CA-02.3, CA-02.4, D-03
- Depende de: T-01
- Paralelizável: não
- Arquivos: `backend/internal/applications/`, `backend/internal/server/` (código
  `group_full`), `openapi.yaml`
- Pronto quando: testes de integração provam três Suportes no mesmo grupo livre,
  `group_full` ao candidatar e ao aceitar com o grupo cheio, aceites simultâneos na
  última vaga e troca de função com o grupo cheio.
- Commit:

### T-03 — Banco de talentos no grupo livre  [P] [ ]
- Cobre: RN-13, CA-05.2 (API)
- Depende de: T-01
- Paralelizável: [P] com T-02 (pacote diferente)
- Arquivos: `backend/internal/talents/`, `backend/internal/server/talents.go`,
  `openapi.yaml` (`/talents/count`)
- Pronto quando: testes provam que um grupo livre com vaga traz Tank, Suporte e Dano;
  que um grupo livre cheio traz ninguém; e que a contagem da criação usa as vagas do
  formulário menos a do dono.
- Commit:

## Web

### T-04 — Formulário, detalhe e diálogos  [ ]
- Cobre: US-01, US-02, US-04, CA-01.1, CA-01.3, CA-02.5, CA-04.3 (mensagem), RN-10,
  RN-11, RNF-03
- Depende de: T-01, T-02
- Paralelizável: não
- Arquivos: `web/src/lib/lobbies/seats.ts`, `LobbyForm.svelte`, `form.ts`, `messages.ts`,
  `web/src/lib/applications/detail.ts`, `messages.ts`, `routes/lobbies/[id]/+page.svelte`,
  `routes/lobbies/novo`, `routes/lobbies/[id]/editar`
- Pronto quando: testes de SSR e e2e provam a escolha da formação pelo teclado, o
  padrão por função, a lista única de lugares no detalhe, a candidatura de qualquer
  função e as mensagens `group_full` e `formation/locked`.
- Commit:

### T-05 — Home: card, destaque e filtro  [P] [ ]
- Cobre: US-03, RN-08, RN-09, CA-03.1, CA-03.2
- Depende de: T-01
- Paralelizável: [P] com T-04 (pastas diferentes)
- Arquivos: `web/src/lib/home/`, `web/src/lib/lobbies/toHome.ts`, `routes/+page.svelte`
- Pronto quando: testes provam o card e o destaque com o selo, "X de N" e a barra; o
  filtro "Vaga para" e as contagens com o grupo livre; e a borda neutra.
- Commit:

### T-06 — Convite, preview e prévia da criação  [P] [ ]
- Cobre: RN-12, RN-13 (web), CA-05.1, CA-05.2 (tela)
- Depende de: T-03, T-04
- Paralelizável: [P] com T-05
- Arquivos: `web/src/lib/lobbies/share.ts`, `LobbyForm.svelte` (contagem),
  `routes/lobbies/novo/disponiveis/+server.ts`, `web/src/lib/talents/api.ts`
- Pronto quando: testes provam "Vagas: 8 livres", "1 livre" e "Grupo lotado" no convite e
  no preview, a contagem da criação no grupo livre e o painel com as três funções.
- Commit:

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-01, T-04 |
| CA-01.2 | T-01 |
| CA-01.3 | T-04 |
| CA-02.1 | T-02 |
| CA-02.2 | T-02, T-04 |
| CA-02.3 | T-02 |
| CA-02.4 | T-02 |
| CA-02.5 | T-04 |
| CA-03.1 | T-05 |
| CA-03.2 | T-05 |
| CA-04.1 | T-01 |
| CA-04.2 | T-01 |
| CA-04.3 | T-01, T-04 |
| CA-05.1 | T-06 |
| CA-05.2 | T-03, T-06 |

## Descobertas
- (nenhuma)
