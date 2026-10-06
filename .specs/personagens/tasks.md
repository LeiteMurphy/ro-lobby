# Tasks — Perfil e personagens

- Spec: `./spec.md` · Design: `./design.md`
- Notion: [Épico](https://app.notion.com/p/3f1d4a3a5eff810e8ce4dc095c2037dc)
- Branch: `feature/personagens`

## US-07 — Base no backend

### T-01 — Tabela de personagens  [x]
- Cobre: RN-01, RN-04, RN-05, RN-07, RN-08, RN-09, RN-12, CA-07.3, D-02
- Depende de: —
- Paralelizável: não
- Arquivos: `backend/migrations/00003_characters.sql`, `backend/queries/characters.sql`,
  `backend/internal/db/` (gerado)
- Pronto quando: a migração aplica e reverte num banco vazio; testes de integração provam
  que o banco recusa um nick repetido sem diferenciar maiúsculas ("Brasa" e "brasa",
  "FAÍSCA" e "faísca"), aceita "Faísca" ao lado de "Faisca", recusa dois principais do
  mesmo Usuário, nível fora de 1–275, função desconhecida e link sem `https://`, e apaga
  os personagens junto com o Usuário.
- Commit: `353b036`
- Notion: https://app.notion.com/p/3f1d4a3a5eff81948f12f18a21191bfe

### T-02 — Catálogo de classes e retratos em Go  [P] [x]
- Cobre: RN-06, RN-10, RN-20, D-03, D-04
- Depende de: —
- Paralelizável: [P] com T-01
- Arquivos: `backend/internal/catalog/`
- Pronto quando: testes unitários passam para as 82 classes (mesmos nomes do bROWiki que
  o web usa hoje), ids únicos em kebab-case sem acento, `ClassByID` para id válido e
  inválido, e os 4 retratos.
- Commit: `d58945d`
- Notion: https://app.notion.com/p/3f1d4a3a5eff8195ac95f18bf5c75ebd

### T-03 — Serviço de personagens  [x]
- Cobre: RN-01, RN-02, RN-04 a RN-15, CA-02.1, CA-02.3, CA-02.4, CA-02.5, CA-02.6,
  CA-02.7, CA-02.8, CA-03.1, CA-03.2, CA-03.3, CA-04.1, CA-04.3, CA-04.4, CA-05.1,
  CA-07.3, D-06, D-08, D-09
- Depende de: T-01, T-02
- Paralelizável: não
- Arquivos: `backend/internal/characters/`
- Pronto quando: testes de integração passam para:
  - primeiro personagem principal e retrato padrão;
  - cada erro de campo com o código certo;
  - nick com espaços nas pontas;
  - nick de outro Usuário (`taken`) e o próprio nick mantido na edição;
  - limite de 10;
  - troca do principal, e o mais antigo virando principal ao excluir;
  - excluir o único personagem;
  - 404 para personagem de outro Usuário, sem alterar nada;
  - cadastros simultâneos, que não passam de 10 nem repetem nick ou principal.
- Commit: `cd62fd9`
- Notion: https://app.notion.com/p/3f1d4a3a5eff81e0a96cda549838dae7

### T-04 — Contrato e rotas da API  [x]
- Cobre: RN-01, RN-02, RN-20, CA-07.1, CA-07.2, CA-02.4, CA-03.3, CA-04.4, D-01, D-07
- Depende de: T-03
- Paralelizável: não
- Arquivos: `openapi.yaml`, `backend/internal/api/api.gen.go`,
  `backend/internal/server/`, `backend/cmd/api/main.go`
- Pronto quando:
  - o contrato descreve `GET /classes` e as cinco rotas de personagem;
  - o código gerado está em dia;
  - testes das rotas passam para `GET /classes` sem sessão (82 classes), 401 sem sessão
    em todas as rotas de personagem, 201, 200, 204, 404, 409 `character_limit` e 422 com
    os campos.
- Commit: `5922265`
- Notion: https://app.notion.com/p/3f1d4a3a5eff81d1bb49d8393c3dd1d1

## US-01 a US-05 — Tela de perfil

### T-05 — Cliente da API, mensagens e retratos no web  [x]
- Cobre: RN-10, RN-19, RNF-02, RNF-04, D-04, D-07, D-10
- Depende de: T-04
- Paralelizável: não
- Arquivos: `web/src/lib/api/schema.gen.ts`, `web/src/lib/characters/`,
  `web/static/portraits/`, `web/src/lib/catalog/classes.ts`
- Pronto quando: testes Vitest passam para:
  - as chamadas à API (Bearer, 401 virando "sem sessão", erros de campo repassados);
  - as mensagens em pt-BR de cada código;
  - um arquivo em `static/portraits/` para cada retrato do enum;
  - um ícone para cada linha do catálogo, com ícone padrão.
- Commit: `452b1ee`
- Notion: https://app.notion.com/p/3f1d4a3a5eff8134bedacf443d0e4554

### T-06 — Página `/perfil` com cartas, menu e diálogos  [x]
- Cobre: RN-03, RN-09, RN-15, RN-16, RN-17, RN-19, RNF-01, CA-01.1, CA-01.2, CA-01.3,
  CA-01.4, CA-02.2, CA-02.8, CA-02.9, CA-04.2, CA-06.1, D-05
- Depende de: T-05
- Paralelizável: não
- Arquivos: `web/src/routes/perfil/`, `web/src/lib/characters/components/`
  (`CharacterCard`, `CardMenu`, `CharacterDialog`, `ConfirmDialog`)
- Pronto quando:
  - visitante é levado ao login com `next=/perfil`;
  - as cartas aparecem com o principal primeiro e o link com `rel="noopener noreferrer"`;
  - o perfil vazio aparece quando não há personagens;
  - "Adicionar personagem" fica desabilitado com 10;
  - as actions devolvem erros por campo sem perder os valores;
  - num 401 da API, a action apaga o cookie e manda ao login;
  - os testes de renderização no servidor e das actions passam.
- Commit: `bddeb82`
- Notion: https://app.notion.com/p/3f1d4a3a5eff813dbf01c45b0c39ebab

## US-06 — Acesso pelo menu

### T-07 — "Meu perfil" no menu do usuário  [P] [x]
- Cobre: RN-18, CA-06.2
- Depende de: T-05
- Paralelizável: [P] com T-06 (arquivos diferentes)
- Arquivos: `web/src/lib/home/components/UserMenu.svelte`, testes do menu
- Pronto quando: logado, o menu mostra "Meu perfil" (link para `/perfil`) acima de
  "Sair", e o teste de renderização passa.
- Commit: `1647c12`
- Notion: https://app.notion.com/p/3f1d4a3a5eff81c4a044d625759cc6e8

## US-01 a US-06 — Ponta a ponta

### T-08 — Discord falso com dois usuários e ponta a ponta do perfil  [x]
- Cobre: CA-01.1, CA-01.2, CA-01.3, CA-01.4, CA-02.1, CA-02.2, CA-02.3, CA-02.8,
  CA-02.9, CA-03.1, CA-04.1, CA-04.2, CA-04.3, CA-05.1, CA-06.1, CA-06.2, RNF-01, RNF-02,
  D-11
- Depende de: T-06, T-07
- Paralelizável: não
- Arquivos: `backend/internal/discordfake/`, `web/test/e2e/perfil.spec.ts`
- Pronto quando: o Discord falso aceita escolher o usuário, e o Playwright prova:
  - cadastrar, editar, trocar o principal e excluir com confirmação (e cancelar);
  - o nick recusado para a segunda conta;
  - o erro junto do campo;
  - o limite de 10;
  - o menu "…" e os diálogos pelo teclado;
  - nenhuma requisição para fora do servidor em `/perfil`.

  Os nicks levam sufixo aleatório para não colidir entre execuções.
- Commit: `d16d9d0`
- Notion: https://app.notion.com/p/3f1d4a3a5eff81aaa11bfac7a2d9123f

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-06, T-08 |
| CA-01.2 | T-06, T-08 |
| CA-01.3 | T-06, T-08 |
| CA-01.4 | T-06, T-08 |
| CA-02.1 | T-03, T-08 |
| CA-02.2 | T-06, T-08 |
| CA-02.3 | T-03, T-08 |
| CA-02.4 | T-03, T-04 |
| CA-02.5 | T-03 |
| CA-02.6 | T-03 |
| CA-02.7 | T-03 |
| CA-02.8 | T-03, T-06, T-08 |
| CA-02.9 | T-06, T-08 |
| CA-03.1 | T-03, T-08 |
| CA-03.2 | T-03 |
| CA-03.3 | T-03, T-04 |
| CA-04.1 | T-03, T-08 |
| CA-04.2 | T-06, T-08 |
| CA-04.3 | T-03, T-08 |
| CA-04.4 | T-03, T-04 |
| CA-05.1 | T-03, T-08 |
| CA-06.1 | T-06, T-08 |
| CA-06.2 | T-07, T-08 |
| CA-07.1 | T-04 |
| CA-07.2 | T-04 |
| CA-07.3 | T-01, T-03 |

Todos os 26 critérios da spec estão cobertos.

## Descobertas
- 2026-10-06 — Os nomes das constantes de enum geradas pelo oapi-codegen colidiram
  (`Role` e `Portrait` viraram tanto tipo quanto valor de `FieldError.field`). — O
  `oapi.cfg.yaml` passou a prefixar os valores com o tipo (`RoleTank`,
  `ErrorErrorNoSession`), e as rotas do login foram ajustadas aos nomes novos (T-04).
- 2026-10-06 — A carta vertical (1b) mostra só o ícone da função, não o da classe. — O
  teste de "ícone para cada linha do catálogo" previsto na T-05 (D-03) ficou sem uso e
  não foi feito; `classes.ts` continua só para a Home fictícia.
- 2026-10-06 — O ponta a ponta da T-08 mostrou que "Tornar principal" não atualizava a
  lista (o `use:enhance` do menu não chamava `update()`). — Corrigido num commit
  separado (`fix(web)`), antes do commit da T-08.
- 2026-10-06 — Com "Meu perfil" no menu do usuário, o foco ao abrir o menu passou a ir
  para o primeiro item. — O teste de teclado do login (RNF-02 da `login-discord`) foi
  ajustado na T-07: "Meu perfil" recebe o foco e a seta para baixo chega em "Sair".
