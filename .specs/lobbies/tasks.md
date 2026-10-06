# Tasks — Lobbies

- Spec: `./spec.md` · Design: `./design.md`
- Notion: [Épico](https://app.notion.com/p/3f1d4a3a5eff81c99cb1fe60c2ac1ac2)
- Branch: `feature/lobbies`

## US-06 — Base no backend

### T-01 — Catálogo de instâncias em Go  [P] [ ]
- Cobre: RN-01, RN-02, CA-06.1, CA-06.2, D-02
- Depende de: —
- Paralelizável: [P] com T-02
- Arquivos: `backend/internal/catalog/`
- Pronto quando: testes unitários passam para as instâncias de grupo do bROWiki (nenhuma
  "Solo"), ids únicos em kebab-case, nível e retorno de cada uma, `InstanceByID` válido e
  inválido, e a ordem da RN-02 (130+ primeiro, nível decrescente, nome no empate).
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff815c82d2fc2db4ee8fc0

### T-02 — Tabela de lobbies  [P] [ ]
- Cobre: RN-06, RN-07, RN-09, RN-13, RN-19, D-01, D-02
- Depende de: —
- Paralelizável: [P] com T-01
- Arquivos: `backend/migrations/00004_lobbies.sql`, `backend/queries/lobbies.sql`,
  `backend/internal/db/` (gerado)
- Pronto quando: a migração aplica e reverte; testes de integração provam que o banco
  recusa vagas fora de 0–12 ou soma fora de 1–12, nível fora de 1–275, função
  desconhecida, observação acima de 250 e cancelamento sem motivo (ou motivo fora de
  10–250); o personagem excluído vira nulo no lobby; o lobby sai junto com o Usuário.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff810291ddf2384b6e633d

### T-03 — Serviço de lobbies  [ ]
- Cobre: RN-04 a RN-11, RN-13, RN-14, RN-17 a RN-20, CA-01.1 a CA-01.9, CA-01.11,
  CA-01.12, CA-02.2, CA-04.1 a CA-04.4, CA-05.1, CA-05.2, CA-06.5, D-03, D-04, D-05, D-08
- Depende de: T-01, T-02
- Paralelizável: não
- Arquivos: `backend/internal/lobbies/`
- Pronto quando: testes de integração passam para:
  - criação com os padrões e o personagem do dono na vaga;
  - cada erro de campo (horário, vagas, nível, personagem, observação, instância);
  - conflito de 2 h (21:30 recusado, 22:00 aceito);
  - limite de 5;
  - virada do dia em São Paulo (23:30);
  - listagem só de abertos, por início, num intervalo de datas;
  - edição com vagas abaixo dos ocupantes e nível acima do dono;
  - 404 para lobby de outro e 409 para iniciado ou cancelado;
  - cancelamento com motivo curto e válido;
  - criações simultâneas que não passam do limite nem criam conflito.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff81b1b895cc1ffba31154

### T-04 — Personagem do dono travado  [ ]
- Cobre: RN-21, CA-06.4, D-05
- Depende de: T-02
- Paralelizável: [P] com T-03 (pacote diferente)
- Arquivos: `backend/internal/characters/`, `backend/queries/characters.sql`
- Pronto quando: testes de integração passam para excluir e mudar a função do personagem
  dono de um lobby aberto (recusado com `ErrInOpenLobby`), mudar outro campo (aceito), e
  excluir depois do início ou do cancelamento (aceito).
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff816dab61cbc3178a8dcb

### T-05 — Contrato e rotas da API  [ ]
- Cobre: RN-01, RN-04, RN-15, RN-20, RN-21, CA-01.7, CA-01.12, CA-03.2, CA-04.4,
  CA-05.3, CA-06.1, CA-06.3, D-06, D-07, D-08
- Depende de: T-03, T-04
- Paralelizável: não
- Arquivos: `openapi.yaml`, `backend/internal/api/api.gen.go`, `backend/internal/server/`,
  `backend/cmd/api/main.go`
- Pronto quando:
  - o contrato descreve as rotas de instâncias e de lobbies, e o código gerado está em
    dia;
  - testes das rotas passam para leituras sem sessão, 401 nas escritas sem sessão, 201,
    200, 404, 409 `lobby_limit`, 409 `lobby_not_open` e 422 com os campos;
  - excluir ou mudar a função do personagem dono de lobby aberto responde 409
    `character_in_open_lobby`.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff81509526f6238d068a0d

## US-02 — Home com dados reais

### T-06 — Cliente da API, horário de Brasília e Home com os lobbies reais  [ ]
- Cobre: RN-03, RN-13, RN-14, RN-22, RN-23, RNF-05, CA-02.1, CA-02.2, CA-02.3, CA-02.4,
  D-04, D-09
- Depende de: T-05
- Paralelizável: não
- Arquivos: `web/src/lib/api/schema.gen.ts`, `web/src/lib/lobbies/`,
  `web/src/routes/+page.server.ts`, `web/src/lib/home/`
- Pronto quando: testes Vitest passam para:
  - a conversão de Brasília para UTC e o contrário, com a virada do dia;
  - os 14 dias;
  - a conversão do `Lobby` da API para o da Home;
  - o `load` da Home com a API, e com a API fora do ar;
  - "Criar lobby" virando link para `/lobbies/novo` e "Ver grupo" para o detalhe;
  - a capa padrão para todas as instâncias.

  Os testes de tela da Home continuam passando com os dados de teste.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff81b18493dc21671ed80f

## US-01 e US-04 — Criar e editar

### T-07 — Páginas de criar e editar lobby  [ ]
- Cobre: RN-02, RN-04, RN-05 a RN-12, RN-17, RN-18, RNF-01, RNF-04, CA-01.1, CA-01.2,
  CA-01.10, CA-04.1, CA-06.2, D-07
- Depende de: T-06
- Paralelizável: [P] com T-08
- Arquivos: `web/src/routes/lobbies/novo/`, `web/src/routes/lobbies/[id]/editar/`,
  `web/src/lib/lobbies/components/LobbyForm.svelte`, `web/src/lib/lobbies/messages.ts`
- Pronto quando:
  - visitante é levado ao login com volta para `/lobbies/novo`;
  - sem personagem, aparece o aviso com link para `/perfil`;
  - o select de instâncias tem os dois grupos na ordem certa;
  - os padrões são 1/2/3, o nível da instância e o principal;
  - a prévia do card acompanha o formulário;
  - os erros aparecem junto do campo, sem perder os valores;
  - a edição só abre para o dono com o lobby aberto, sem instância nem personagem;
  - os testes de renderização no servidor e das actions passam.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff81fab6faf0f911165027

## US-03 e US-05 — Detalhe e cancelamento

### T-08 — Página de detalhe e cancelamento  [ ]
- Cobre: RN-15, RN-16, RN-19, RNF-01, CA-03.1, CA-03.2, CA-03.3, CA-05.1, CA-05.2
- Depende de: T-06
- Paralelizável: [P] com T-07
- Arquivos: `web/src/routes/lobbies/[id]/`, `web/src/lib/lobbies/components/`
- Pronto quando:
  - o detalhe mostra instância, dia e hora, estado, nível mínimo, observação, composição
    com o dono na vaga dele e o anfitrião;
  - um lobby inexistente mostra a página de não encontrado;
  - o dono vê "Editar" e "Cancelar lobby";
  - os outros veem "Candidatar" em breve;
  - o cancelamento pede um motivo de 10 a 250 caracteres;
  - o lobby cancelado mostra o selo e o motivo;
  - os testes de renderização e da action passam.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff8128aeedec46dec52cc1

### T-09 — Mensagem do personagem travado no perfil  [P] [ ]
- Cobre: RN-21, CA-06.4
- Depende de: T-05
- Paralelizável: [P] com T-06 a T-08
- Arquivos: `web/src/lib/characters/`, `web/src/routes/perfil/`
- Pronto quando: excluir ou mudar a função do personagem dono de lobby aberto mostra
  "Esse personagem está num lobby aberto. Cancele o lobby antes." no perfil, com teste das
  actions.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff81c9b5d0f9d9e5b05278

## US-01 a US-06 — Ponta a ponta

### T-10 — Ponta a ponta dos lobbies  [ ]
- Cobre: CA-01.1, CA-01.8, CA-01.10, CA-02.1, CA-02.3, CA-02.4, CA-03.1, CA-03.3,
  CA-04.1, CA-05.1, CA-06.4, RNF-01, RNF-02
- Depende de: T-07, T-08, T-09
- Paralelizável: não
- Arquivos: `web/test/e2e/lobbies.spec.ts`, `web/test/e2e/home.spec.ts`
- Pronto quando: o Playwright prova, com duas contas:
  - criar pela Home e ver o card;
  - o conflito de horário;
  - abrir o detalhe pelo "Ver grupo", e o que o dono e o outro jogador veem;
  - editar e cancelar;
  - o perfil travando o personagem do dono;
  - o teclado no formulário e no diálogo;
  - nada carregado de fora do servidor.

  O ponta a ponta da Home deixa de depender dos dados fictícios.
- Commit:
- Notion: https://app.notion.com/p/3f1d4a3a5eff813a826af52e8712af56

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-03, T-07, T-10 |
| CA-01.2 | T-03, T-07 |
| CA-01.3 | T-03 |
| CA-01.4 | T-03 |
| CA-01.5 | T-03 |
| CA-01.6 | T-03 |
| CA-01.7 | T-03, T-05 |
| CA-01.8 | T-03, T-10 |
| CA-01.9 | T-03 |
| CA-01.10 | T-07, T-10 |
| CA-01.11 | T-03 |
| CA-01.12 | T-03, T-05 |
| CA-02.1 | T-06, T-10 |
| CA-02.2 | T-03, T-06 |
| CA-02.3 | T-06, T-10 |
| CA-02.4 | T-06, T-10 |
| CA-03.1 | T-08, T-10 |
| CA-03.2 | T-05, T-08 |
| CA-03.3 | T-08, T-10 |
| CA-04.1 | T-03, T-07, T-10 |
| CA-04.2 | T-03 |
| CA-04.3 | T-03 |
| CA-04.4 | T-03, T-05 |
| CA-05.1 | T-03, T-08, T-10 |
| CA-05.2 | T-03, T-08 |
| CA-05.3 | T-05 |
| CA-06.1 | T-01, T-05 |
| CA-06.2 | T-01, T-07 |
| CA-06.3 | T-05 |
| CA-06.4 | T-04, T-09, T-10 |
| CA-06.5 | T-03 |

Todos os 31 critérios da spec estão cobertos.

## Descobertas
- Nenhuma até agora.
