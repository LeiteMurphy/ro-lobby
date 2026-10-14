# Tasks — Servidor do jogo

- Spec: `./spec.md` (aprovada em 2026-10-14) · Design: não há (nível M)
- Notion: [Épico](https://app.notion.com/p/3f4d4a3a5eff813abe61c566699df094)
- Branch: `feature/servidor`

## Backend

### T-01 — Servidor no personagem: migração, catálogo e regras  [x]
- Cobre: US-01, RN-01, RN-02, RN-03, RN-04, RN-15, CA-01.1, CA-01.3 (API), CA-01.4,
  CA-01.5 (API), CA-01.6, CA-01.7, CA-01.8 (API), CA-06.7, RNF-01, RNF-02, RNF-03
- Depende de: —
- Paralelizável: não
- Arquivos:
  - `backend/migrations/00010_server.sql`, com `server` em `characters` e `lobbies`,
    `characters.server_to_confirm` e o índice único (servidor, nick);
  - `backend/internal/catalog` (servidores);
  - `backend/queries/characters.sql`, `backend/internal/characters/`;
  - `backend/internal/server/characters.go` (`GET /servers`);
  - `openapi.yaml` e os arquivos gerados.
- Pronto quando:
  - a migração aplica e reverte com personagens e lobbies antigos, deixando todos em
    Freya e os personagens a conferir;
  - os testes de integração provam:
    - `GET /servers` sem sessão;
    - servidor obrigatório e dentro do catálogo;
    - mesmo nick em servidores diferentes e `taken` no mesmo servidor, inclusive ao mudar
      de servidor;
    - a edição do servidor, e a recusa dela com vínculo ativo.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff812c9aebe6ecfafeeefd

### T-02 — Conferir o servidor dos personagens migrados (API)  [P] [ ]
- Cobre: US-06, RN-15, RN-16, CA-06.2 (API), CA-06.3 (API), CA-06.4 (API), CA-06.5,
  CA-06.6
- Depende de: T-01
- Paralelizável: [P] com T-03 e T-04 (só mexe em `characters`)
- Arquivos: `backend/internal/characters/`, `backend/internal/server/characters.go`
  (`POST /characters/{id}/confirm-server` e o campo `serverToConfirm`), `openapi.yaml`
- Pronto quando: os testes de integração provam que:
  - o personagem novo nasce conferido;
  - "Está certo" e salvar uma edição tiram o personagem da conferência;
  - o personagem de outro Usuário responde 404;
  - um personagem a conferir se candidata, cria lobby e entra no banco de talentos
    normalmente.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff818380f0c27db01b6443

### T-03 — Lobby no servidor do dono e grupo no mesmo servidor  [ ]
- Cobre: US-02, US-03, RN-06, RN-07, RN-09, RN-10, CA-02.1, CA-02.3, CA-03.1 a CA-03.5
- Depende de: T-01
- Paralelizável: não
- Arquivos: `backend/queries/lobbies.sql`, `applications.sql`, `swap_requests.sql`,
  `backend/internal/lobbies/`, `backend/internal/applications/`, `backend/internal/server/`
  (código `wrong_server`), `openapi.yaml` (`server` no `Lobby`)
- Pronto quando: os testes de integração provam que:
  - o lobby grava o servidor do personagem do dono e não o muda na edição, na troca do
    dono nem na exclusão do personagem;
  - candidatar, trocar o personagem do dono e pedir troca com personagem de outro
    servidor respondem 409 `wrong_server`, sem mudar nada;
  - o mesmo servidor segue as regras de sempre.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff8136be09f31656dfa3bf

### T-04 — Banco de talentos por servidor (API)  [ ]
- Cobre: US-05, RN-17, RN-18, RN-19, CA-05.1, CA-05.2 (API), CA-05.3 (API), CA-05.4 (API)
- Depende de: T-03
- Paralelizável: não
- Arquivos: `backend/queries/talents.sql`, `backend/internal/talents/`,
  `backend/internal/server/talents.go` (`server` em `/talents` e `/talents/count`),
  `openapi.yaml`
- Pronto quando: os testes de integração provam que:
  - a afinidade de um lobby em Freya ignora personagens de Nidhogg;
  - a contagem da criação usa o servidor do personagem do dono informado;
  - `/talents?server=nidhogg` filtra;
  - cada item traz o servidor.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff8151a767ca7842ea8d29

## Web

### T-05 — Perfil: servidor do personagem e aviso para conferir  [ ]
- Cobre: US-01, US-06, RN-02, RN-03, RN-04, RN-05, RN-16, CA-01.2, CA-01.3 (formulário),
  CA-01.5 (mensagem), CA-01.8 (mensagem), CA-06.1, CA-06.2, CA-06.3, CA-06.4
- Depende de: T-01, T-02
- Paralelizável: [P] com T-06, T-07 e T-08 (só mexe em `characters` e `perfil`)
- Arquivos: `web/src/lib/characters/` (formulário, carteirinha, `messages.ts`),
  `web/src/routes/perfil/`
- Pronto quando: os testes de SSR e o e2e provam:
  - a escolha do servidor no formulário e o "Escolha o servidor" sem ela;
  - o servidor na carteirinha;
  - as mensagens "Esse nick já está em uso nesse servidor" e a da trava com "servidor,
    nível ou função";
  - o aviso com "Está certo" e "Editar", que some quando não sobra personagem a conferir.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff81ec8656e940177df6d9

### T-06 — Lobby: criação, detalhe, escolha de personagem, convite e preview  [P] [ ]
- Cobre: US-02, US-03, RN-06, RN-08, RN-11, RN-20, CA-02.2, CA-02.4 (detalhe), CA-02.5,
  CA-03.6
- Depende de: T-03
- Paralelizável: [P] com T-05, T-07 e T-08
- Arquivos: `web/src/lib/lobbies/` (formulário e prévia), `web/src/lib/applications/`
  (escolha de personagem), `web/src/routes/lobbies/[id]/` (detalhe, convite e
  `og:title`), `web/src/routes/lobbies/novo/`
- Pronto quando: os testes de SSR e o e2e provam:
  - o servidor na criação, mudando com o personagem;
  - o servidor no detalhe;
  - "Outro servidor (Nidhogg)" desabilitado na candidatura e no pedido de troca;
  - o servidor no convite copiado e no `og:title`.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff8113b8b5f4c5b2ff73a1

### T-07 — Home: servidor no card e filtro "Servidor"  [P] [ ]
- Cobre: US-04, RN-08, RN-12, RN-13, RN-14, CA-02.4 (Home), CA-04.1 a CA-04.5
- Depende de: T-03
- Paralelizável: [P] com T-05, T-06 e T-08
- Arquivos: `web/src/lib/home/` (`filters.ts`, card, destaque), `web/src/routes/+page.*`
- Pronto quando: os testes de unidade e o e2e provam:
  - o filtro em Todos com as contagens por servidor;
  - a filtragem combinada com os outros filtros, inclusive no destaque e no estado vazio;
  - o "Limpar filtros";
  - que a escolha não é guardada;
  - o servidor no card e no destaque.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff81ec874df18b4cca1385

### T-08 — Banco de talentos: filtro, servidor nos itens e prévia  [P] [ ]
- Cobre: US-05, RN-17, RN-18, RN-19, CA-05.2 (prévia), CA-05.3, CA-05.4
- Depende de: T-04
- Paralelizável: [P] com T-05, T-06 e T-07
- Arquivos: `web/src/lib/talents/`, `web/src/routes/talentos/`, a prévia da criação de
  lobby (`web/src/routes/lobbies/novo/disponiveis`)
- Pronto quando: os testes de SSR e o e2e provam:
  - o filtro "Servidor" em `/talentos`, que não é guardado;
  - o servidor em cada item e no painel "Jogadores disponíveis";
  - a prévia da criação mudando com o servidor do personagem escolhido.
- Commit:
- Notion: https://app.notion.com/p/3f9d4a3a5eff812b89edf8ecce6efeef

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-01 |
| CA-01.2 | T-05 |
| CA-01.3 | T-01, T-05 |
| CA-01.4 | T-01 |
| CA-01.5 | T-01, T-05 |
| CA-01.6 | T-01 |
| CA-01.7 | T-01 |
| CA-01.8 | T-01, T-05 |
| CA-02.1 | T-03 |
| CA-02.2 | T-06 |
| CA-02.3 | T-03 |
| CA-02.4 | T-06, T-07 |
| CA-02.5 | T-06 |
| CA-03.1 | T-03 |
| CA-03.2 | T-03 |
| CA-03.3 | T-03 |
| CA-03.4 | T-03 |
| CA-03.5 | T-03 |
| CA-03.6 | T-06 |
| CA-04.1 | T-07 |
| CA-04.2 | T-07 |
| CA-04.3 | T-07 |
| CA-04.4 | T-07 |
| CA-04.5 | T-07 |
| CA-05.1 | T-04 |
| CA-05.2 | T-04, T-08 |
| CA-05.3 | T-04, T-08 |
| CA-05.4 | T-04, T-08 |
| CA-06.1 | T-05 |
| CA-06.2 | T-02, T-05 |
| CA-06.3 | T-02, T-05 |
| CA-06.4 | T-02, T-05 |
| CA-06.5 | T-02 |
| CA-06.6 | T-02 |
| CA-06.7 | T-01 |

Sem lacunas: todos os 35 critérios da spec têm task.

## Descobertas
- 2026-10-14 — A spec dizia 400 nos erros de campo (CA-01.3, CA-01.5, CA-01.6), mas a API
  responde 422 em toda validação desde a `personagens`. A spec foi corrigida para 422; o
  comportamento não muda.
- 2026-10-14 — T-01 já grava o servidor do personagem do dono no lobby criado, porque a
  coluna `lobbies.server` nasce obrigatória na mesma migração. Os testes da RN-06 e da
  RN-07 ficam na T-03.
- 2026-10-14 — Entre a T-01 e a T-05, o formulário do perfil ainda não tem o campo
  servidor, então criar personagem pela tela responde "servidor obrigatório". O e2e do
  perfil volta a passar na T-05. O seed do e2e já manda Freya.
- 2026-10-14 — A migração passa de `00007` (rascunho) para `00010`; a `discord-bot` fica
  com a `00011`.
