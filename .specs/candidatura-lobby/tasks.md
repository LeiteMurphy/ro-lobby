# Tasks — Candidatura a lobby (Parte 1)

- Spec: `./spec.md` · Design: `./design.md`
- Notion: <link do Épico>
- Branch: `feature/candidatura`

## US-01 a US-04 — Base no backend

### T-01 — Tabelas de candidatura e histórico, e ocupantes do lobby  [ ]
- Cobre: RN-02, RN-06, RN-09, RN-17, RN-18, D-01, D-03
- Depende de: —
- Arquivos: `backend/migrations/00005_applications.sql`, `backend/queries/applications.sql`,
  `backend/queries/lobbies.sql`, `backend/internal/db/` (gerado)
- Pronto quando: a migração aplica e reverte; testes de integração provam a candidatura
  ativa única por Usuário e lobby, os `CHECK` de estado, mensagem e justificativa, o
  histórico e as contagens de ocupantes e pendentes do lobby.
- Commit:
- Notion:

### T-02 — Serviço de candidaturas  [ ]
- Cobre: RN-01 a RN-13, RN-16, RN-17, RN-18, RN-30, CA-01.1 a CA-01.11, CA-02.1 a
  CA-02.12, CA-03.2 a CA-03.6, CA-04.1 a CA-04.3, D-02, D-04, D-05
- Depende de: T-01
- Arquivos: `backend/internal/applications/`, `backend/internal/lobbies/`
- Pronto quando: testes de integração passam para cada regra de candidatura, aceite,
  recusa e retirada; conflito com a janela de 2 h (dono e membro); aceites simultâneos
  na última vaga; expiração pelo início e pelo cancelamento; edição de lobby com as vagas
  dos membros; histórico de transições.
- Commit:
- Notion:

### T-03 — Travas do personagem com candidatura  [P] [ ]
- Cobre: RN-25, RN-26, CA-09.1 a CA-09.4
- Depende de: T-01
- Paralelizável: [P] com T-02 (pacote diferente)
- Arquivos: `backend/internal/characters/`, `backend/queries/characters.sql`
- Pronto quando: testes de integração passam para nível e função travados (dono,
  pendente, aceito), outros campos livres, exclusão travada e liberada depois do início.
- Commit:
- Notion:

### T-04 — Contrato e rotas, com o detalhe conforme quem olha  [ ]
- Cobre: RN-28, RN-29, RN-32, RNF-02, CA-03.1, CA-03.7, CA-03.8, CA-10.2, CA-10.3, D-06,
  D-07
- Depende de: T-02, T-03
- Arquivos: `openapi.yaml`, `backend/internal/api/`, `backend/internal/server/`
- Pronto quando: o contrato descreve as rotas novas e o `Lobby` com membros e pendentes;
  testes das rotas passam para cada papel (dono, membro, candidato, visitante) e cada
  erro (401, 404, 409 com código, 422).
- Commit:
- Notion:

## US-01, US-02, US-10 — Tela do lobby

### T-05 — Detalhe do lobby com candidatura, painel do jogador e decisão do dono  [ ]
- Cobre: RN-28, RN-31, RN-32, CA-01.1, CA-02.1, CA-02.2, CA-10.1, CA-10.2, CA-10.3,
  RNF-01 (teclado)
- Depende de: T-04
- Arquivos: `web/src/lib/applications/`, `web/src/routes/lobbies/[id]/`
- Pronto quando: testes Vitest passam para o painel (seleção e conteúdo por papel), o
  diálogo de candidatura (personagens habilitados e motivos), aceitar, recusar com
  justificativa, retirar e as mensagens de cada código de erro.
- Commit:
- Notion:

## US-03, US-10 — Acompanhamento

### T-06 — Minhas candidaturas, selo de pendentes e aviso no perfil  [ ]
- Cobre: RN-25, RN-33, RN-34, CA-03.1, CA-03.2, CA-10.4, CA-10.5
- Depende de: T-04
- Paralelizável: [P] com T-05
- Arquivos: `web/src/routes/candidaturas/`, `web/src/lib/home/components/`,
  `web/src/routes/perfil/`
- Pronto quando: testes Vitest passam para a lista com estados e justificativa, retirar,
  o item no menu, o selo só para o dono e a mensagem nova da trava.
- Commit:
- Notion:

## US-01 a US-04, US-10 — Ponta a ponta

### T-07 — Ponta a ponta da candidatura com três contas  [ ]
- Cobre: CA-01.1, CA-01.7, CA-01.11, CA-02.1, CA-02.2, CA-02.10, CA-03.1, CA-03.2,
  CA-03.7, CA-04.2, CA-09.1, CA-10.1 a CA-10.5, RNF-01
- Depende de: T-05, T-06
- Arquivos: `web/test/e2e/candidatura.spec.ts`, `web/test/e2e/seed.ts`
- Pronto quando: o Playwright prova, com dono, candidato e visitante: candidatar, ver o
  pendente só como dono, aceitar, recusar com justificativa, retirar, "Minhas
  candidaturas", o selo, o painel com e sem Discord, a trava de nível no perfil e a
  expiração pelo cancelamento.
- Commit:
- Notion:

## Matriz de cobertura (Parte 1)
| Critério | Tasks |
|---|---|
| CA-01.1 a CA-01.11 | T-02 (todos), T-05 (01.1), T-07 (01.1, 01.7, 01.11) |
| CA-02.1 a CA-02.12 | T-02 (todos), T-05 (02.1, 02.2), T-07 (02.1, 02.2, 02.10) |
| CA-03.1 | T-04, T-06, T-07 |
| CA-03.2 a CA-03.6 | T-02 (todos), T-06 (03.2), T-07 (03.2) |
| CA-03.7, CA-03.8 | T-04, T-07 (03.7) |
| CA-04.1 a CA-04.3 | T-02, T-07 (04.2) |
| CA-09.1 a CA-09.4 | T-03, T-07 (09.1) |
| CA-10.1 a CA-10.3 | T-04 (10.2, 10.3), T-05, T-07 |
| CA-10.4, CA-10.5 | T-06, T-07 |

Os 43 critérios da Parte 1 estão cobertos.

## Descobertas
- Nenhuma até agora.
