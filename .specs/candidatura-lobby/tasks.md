# Tasks — Candidatura a lobby (Parte 1)

- Spec: `./spec.md` · Design: `./design.md`
- Notion: [Épico](https://app.notion.com/p/3ead4a3a5eff810bb7f4eae87f4ab3b2)
- Branch: `feature/candidatura`

## US-01 a US-04 — Base no backend

### T-01 — Tabelas de candidatura e histórico, e ocupantes do lobby  [x]
- Cobre: RN-02, RN-06, RN-09, RN-17, RN-18, D-01, D-03
- Depende de: —
- Arquivos: `backend/migrations/00005_applications.sql`, `backend/queries/applications.sql`,
  `backend/queries/lobbies.sql`, `backend/internal/db/` (gerado)
- Pronto quando: a migração aplica e reverte; testes de integração provam a candidatura
  ativa única por Usuário e lobby, os `CHECK` de estado, mensagem e justificativa, o
  histórico e as contagens de ocupantes e pendentes do lobby.
- Commit: de97202
- Notion: https://app.notion.com/p/3f1d4a3a5eff81ff98bacdd773c4fcab

### T-02 — Serviço de candidaturas  [x]
- Cobre: RN-01 a RN-13, RN-16, RN-17, RN-18, RN-30, CA-01.1 a CA-01.11, CA-02.1 a
  CA-02.12, CA-03.2 a CA-03.6, CA-04.1 a CA-04.3, D-02, D-04, D-05
- Depende de: T-01
- Arquivos: `backend/internal/applications/`, `backend/internal/lobbies/`
- Pronto quando: testes de integração passam para cada regra de candidatura, aceite,
  recusa e retirada; conflito com a janela de 2 h (dono e membro); aceites simultâneos
  na última vaga; expiração pelo início e pelo cancelamento; edição de lobby com as vagas
  dos membros; histórico de transições.
- Commit: 1670dec
- Notion: https://app.notion.com/p/3f1d4a3a5eff81a59565f3ba1b15babb

### T-03 — Travas do personagem com candidatura  [P] [x]
- Cobre: RN-25, RN-26, CA-09.1 a CA-09.4
- Depende de: T-01
- Paralelizável: [P] com T-02 (pacote diferente)
- Arquivos: `backend/internal/characters/`, `backend/queries/characters.sql`
- Pronto quando: testes de integração passam para nível e função travados (dono,
  pendente, aceito), outros campos livres, exclusão travada e liberada depois do início.
- Commit: 96cf0f8
- Notion: https://app.notion.com/p/3f1d4a3a5eff8116b56bc051fba37594

### T-04 — Contrato e rotas, com o detalhe conforme quem olha  [x]
- Cobre: RN-28, RN-29, RN-32, RNF-02, CA-03.1, CA-03.7, CA-03.8, CA-10.2, CA-10.3, D-06,
  D-07
- Depende de: T-02, T-03
- Arquivos: `openapi.yaml`, `backend/internal/api/`, `backend/internal/server/`
- Pronto quando: o contrato descreve as rotas novas e o `Lobby` com membros e pendentes;
  testes das rotas passam para cada papel (dono, membro, candidato, visitante) e cada
  erro (401, 404, 409 com código, 422).
- Commit: 484240b
- Notion: https://app.notion.com/p/3f1d4a3a5eff81a4ad7fc17a3e00df60

## US-01, US-02, US-10 — Tela do lobby

### T-05 — Detalhe do lobby com candidatura, painel do jogador e decisão do dono  [x]
- Cobre: RN-28, RN-31, RN-32, CA-01.1, CA-02.1, CA-02.2, CA-10.1, CA-10.2, CA-10.3,
  RN-31 (teclado)
- Depende de: T-04
- Arquivos: `web/src/lib/applications/`, `web/src/routes/lobbies/[id]/`
- Pronto quando: testes Vitest passam para o painel (seleção e conteúdo por papel), o
  diálogo de candidatura (personagens habilitados e motivos), aceitar, recusar com
  justificativa, retirar e as mensagens de cada código de erro.
- Commit: 050ae3b
- Notion: https://app.notion.com/p/3f1d4a3a5eff81018cc5ebd9f412b538

## US-03, US-10 — Acompanhamento

### T-06 — Minhas candidaturas, selo de pendentes e aviso no perfil  [x]
- Cobre: RN-25, RN-33, RN-34, CA-03.1, CA-03.2, CA-10.4, CA-10.5
- Depende de: T-04
- Paralelizável: [P] com T-05
- Arquivos: `web/src/routes/candidaturas/`, `web/src/lib/home/components/`,
  `web/src/routes/perfil/`
- Pronto quando: testes Vitest passam para a lista com estados e justificativa, retirar,
  o item no menu, o selo só para o dono e a mensagem nova da trava.
- Commit: 14e677d
- Notion: https://app.notion.com/p/3f1d4a3a5eff8148a136db1b1c93af05

## US-01 a US-04, US-10 — Ponta a ponta

### T-07 — Ponta a ponta da candidatura com três contas  [x]
- Cobre: CA-01.1, CA-01.7, CA-01.11, CA-02.1, CA-02.2, CA-02.10, CA-03.1, CA-03.2,
  CA-03.7, CA-04.2, CA-09.1, CA-10.1 a CA-10.5, RN-31 (teclado)
- Depende de: T-05, T-06
- Arquivos: `web/test/e2e/candidatura.spec.ts`, `web/test/e2e/seed.ts`
- Pronto quando: o Playwright prova, com dono, candidato e visitante: candidatar, ver o
  pendente só como dono, aceitar, recusar com justificativa, retirar, "Minhas
  candidaturas", o selo, o painel com e sem Discord, a trava de nível no perfil e a
  expiração pelo cancelamento.
- Commit: 42095c9
- Notion: https://app.notion.com/p/3f1d4a3a5eff81c19f88f4b7d49eaaad

## Validação, ciclo 1 — ajustes

### T-08 — Discord do anfitrião só para o grupo e "Candidatar" da Home como atalho  [x]
- Cobre: RN-32, RN-35, CA-10.2, CA-10.6, D-05 (trava do lobby na candidatura), CA-01.10
- Depende de: T-07
- Arquivos: `openapi.yaml`, `backend/internal/lobbies/`, `backend/internal/applications/`,
  `backend/internal/server/`, `web/src/lib/home/components/`, `web/test/e2e/`
- Pronto quando: o Discord do anfitrião não sai na lista pública nem no detalhe para quem
  não é dono nem membro; o "Candidatar" do card e do destaque leva ao detalhe; a
  candidatura trava o lobby; o teste da CA-01.10 monta o Given da spec.
- Commit: 6bd4604
- Notion: https://app.notion.com/p/3f1d4a3a5eff81f1983fdf75805acb48

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
| CA-10.6 | T-08 |

Os 44 critérios da Parte 1 estão cobertos.

## Descobertas
- T-02: o dono pode subir o nível mínimo acima do nível de um membro já aceito. Decidido
  com o usuário em 2026-10-06: fica assim; o membro continua no grupo (spec, seção 10).
- T-05: o "Candidatar" do card na Home seguia "Disponível em breve". Decidido com o usuário
  em 2026-10-06: vira atalho para o detalhe (RN-35, CA-10.6, T-08).
- Validação, ciclo 1: o Discord do anfitrião aparecia para todos, contra a RN-32. Corrigido
  na T-08.
- T-11 (2026-10-07): o P2.5 do design dizia 404 para a troca de personagem do dono pedida
  por outro Usuário, mas a CA-07.5 pede "apenas o dono pode decidir". Vale a spec:
  `not_owner` (409); o design foi corrigido.

---

# Parte 2 — Sair, remover e trocas de personagem

- Spec: US-05 a US-08, RN-14, RN-15, RN-19 a RN-24, RN-27, RN-36 a RN-38 · Design: P2.1 a
  P2.7, D-08 a D-13
- Branch: `feature/candidatura-parte2` (rebaseada no `main` depois do #18 e do #19)

## US-05 a US-08 — Base no backend

### T-09 — Tabelas de pedido de troca e bloqueio  [x]
- Cobre: RN-15, RN-18, RN-20 (um pedido pendente), RN-24, D-08, D-09
- Depende de: —
- Arquivos: `backend/migrations/00006_swap_requests.sql`, `backend/queries/applications.sql`,
  `backend/queries/swap_requests.sql`, `backend/internal/db/` (gerado)
- Pronto quando: a migração aplica e reverte; testes de integração provam
  `applications.blocked` com padrão `false`, o pedido pendente único por candidatura, os
  `CHECK` de estado, motivo e justificativa, e o histórico em `swap_request_events`.
- Commit: 0e1afdd
- Notion: https://app.notion.com/p/3f2d4a3a5eff810ab699fec4fc4e659a

### T-10 — Sair do grupo, remover membro e bloqueio  [x]
- Cobre: US-05, US-06, RN-07, RN-09, RN-14, RN-15, RN-37, CA-05.1 a CA-05.4, CA-06.1 a
  CA-06.7, CA-08.10, D-09, D-10
- Depende de: T-09
- Arquivos: `backend/internal/applications/`
- Pronto quando: testes de integração passam para sair e remover (lobby aberto e
  iniciado, quem não é dono ou não é membro, justificativa), a vaga liberada, o pedido
  pendente cancelado na mesma transação, a nova candidatura depois de sair ou de ser
  removido sem bloqueio, o erro `blocked` só no lobby do bloqueio e saída concorrente com
  remoção.
- Commit: 118aee3
- Notion: https://app.notion.com/p/3f2d4a3a5eff812182c3e8648404ec3c

### T-11 — Troca do dono e pedidos de troca do membro  [x]
- Cobre: US-07, US-08, RN-11, RN-16, RN-19 a RN-24, RN-27, RN-36, CA-07.1 a CA-07.6,
  CA-08.1 a CA-08.9, CA-08.11 a CA-08.13, D-10, D-11, D-12
- Depende de: T-10 (mesmo pacote)
- Arquivos: `backend/internal/applications/`, `backend/internal/lobbies/`,
  `backend/queries/lobbies.sql`
- Pronto quando: testes de integração passam para a troca do dono (com vaga, mesma função
  lotada, sem vaga, conflito, nível, quem não é dono); abrir, aceitar, recusar e retirar
  pedido com cada erro (motivo, segundo pendente, personagem de outro, nível no pedido e no
  aceite, vaga e conflito no aceite, pedido de outro); aceite concorrente com outro aceite
  na mesma vaga; pedido expirado pelo início e pelo cancelamento do lobby; histórico.
- Commit: 9dec763
- Notion: https://app.notion.com/p/3f2d4a3a5eff813fbe17e7d00ca41d65

### T-12 — Travas do personagem com pedido de troca  [P] [x]
- Cobre: RN-25, RN-26, CA-09.1 a CA-09.4 (com pedido pendente)
- Depende de: T-09
- Paralelizável: [P] com T-10 e T-11 (pacote `characters` e a consulta da trava)
- Arquivos: `backend/queries/lobbies.sql` (`CharacterInOpenLobby`),
  `backend/internal/characters/`
- Pronto quando: testes de integração provam que o personagem pedido numa troca pendente
  não muda nível nem função e não é excluído, e que fica livre depois do pedido recusado,
  retirado ou expirado.
- Commit: 3f6f517
- Notion: https://app.notion.com/p/3f2d4a3a5eff819b8e5ddd2a0a7cd7a4

### T-13 — Contrato e rotas da Parte 2, e o detalhe com pedidos e bloqueio  [x]
- Cobre: RN-28, RN-32, CA-05.1, CA-06.1, CA-06.3, CA-07.5, CA-08.1, CA-08.5, CA-08.12,
  D-12, D-13 e os endpoints do P2.5
- Depende de: T-11, T-12
- Arquivos: `openapi.yaml`, `backend/internal/api/`, `backend/internal/server/`,
  `web/src/lib/api/` (gerado)
- Pronto quando: o contrato descreve as sete rotas, o `SwapRequest`, `swapRequests` no
  `Lobby` e `myApplication.swapRequest`/`blocked`; testes das rotas passam para cada papel
  (dono, membro, outro membro, visitante) e cada erro (401, 404, 409 com código, 422); os
  pedidos de troca só aparecem para o dono.
- Commit: 9eddae6
- Notion: https://app.notion.com/p/3f2d4a3a5eff810bb6dec80598835a35

## US-05, US-06, US-08 — Tela do membro

### T-14 — Aviso do membro, bloqueado e "Minhas candidaturas"  [x]
- Cobre: RN-38, CA-05.5, CA-06.8 (visão do removido), CA-08.14 (visão do membro)
- Depende de: T-13
- Arquivos: `web/src/lib/applications/`, `web/src/routes/lobbies/[id]/`,
  `web/src/routes/candidaturas/`
- Pronto quando: testes Vitest passam para "Sair do grupo" com confirmação, "Pedir troca"
  (`SwapDialog` no modo pedido, personagens habilitados e motivos), "Pedido de troca
  pendente para X" com "Retirar pedido", o aviso de bloqueado sem "Candidatar", "Sair do
  grupo" e os estados "Saiu" e "Removida" com justificativa em "Minhas candidaturas", e
  as mensagens dos códigos novos.
- Commit: a2867a5
- Notion: https://app.notion.com/p/3f2d4a3a5eff8197bc3cddaf7b39f72e

## US-06, US-07, US-08 — Tela do dono

### T-15 — Remover, trocar o próprio personagem e decidir pedidos de troca  [x]
- Cobre: RN-38, CA-06.8 (visão do dono), CA-07.1 (tela), CA-08.14 (visão do dono)
- Depende de: T-14 (mesmas páginas)
- Arquivos: `web/src/lib/applications/components/`, `web/src/routes/lobbies/[id]/`
- Pronto quando: testes Vitest passam para "Remover" no painel do membro (`RemoveDialog`
  com justificativa e "Bloquear neste lobby"), "Trocar personagem" no próprio card
  (`SwapDialog` no modo dono), o bloco "Pedidos de troca" e o painel do pedido com o
  personagem atual, o novo, o motivo, "Aceitar" e "Recusar".
- Commit: b39784a
- Notion: https://app.notion.com/p/3f2d4a3a5eff816faefefacc83680c30

## US-05 a US-08 — Ponta a ponta

### T-16 — Ponta a ponta da Parte 2 com dono, membro e removido  [x]
- Cobre: CA-05.5, CA-06.8, CA-07.1, CA-08.14, CA-05.4, RN-31 (teclado)
- Depende de: T-15
- Arquivos: `web/test/e2e/candidatura.spec.ts`, `web/test/e2e/seed.ts`
- Pronto quando: o Playwright prova sair do grupo e candidatar de novo; remover com
  bloqueio e o aviso do removido no detalhe e em "Minhas candidaturas"; o dono trocar o
  próprio personagem; o membro pedir troca, o dono ver e aceitar, e o membro retirar outro
  pedido.
- Commit: 3b5d28b
- Notion: https://app.notion.com/p/3f2d4a3a5eff81b39c98d42470aca832

## Matriz de cobertura (Parte 2)
| Critério | Tasks |
|---|---|
| CA-05.1, CA-05.2 | T-10, T-13 (05.1) |
| CA-05.3 | T-10 |
| CA-05.4 | T-10, T-16 |
| CA-05.5 | T-14, T-16 |
| CA-06.1 a CA-06.7 | T-10, T-13 (06.1, 06.3) |
| CA-06.8 | T-14, T-15, T-16 |
| CA-07.1 a CA-07.6 | T-11, T-13 (07.5), T-15 (07.1), T-16 (07.1) |
| CA-08.1 a CA-08.9 | T-11, T-13 (08.1, 08.5) |
| CA-08.10 | T-10 |
| CA-08.11 a CA-08.13 | T-11, T-13 (08.12) |
| CA-08.14 | T-14, T-15, T-16 |
| CA-09.1 a CA-09.4 (pedido pendente) | T-12 |

Os 33 critérios da Parte 2 estão cobertos, mais a extensão da CA-09 ao pedido de troca.

## Descobertas (Parte 2)
- T-11 (2026-10-07): o P2.5 dizia 404 para a troca de personagem do dono pedida por outro
  Usuário; vale a CA-07.5 (`not_owner`, 409). O design foi corrigido (ver Descobertas da
  Parte 1, acima).
- T-15 (2026-10-07): a spec não diz quem vê a justificativa da recusa de um pedido de troca
  (RN-22). Hoje ela fica gravada e sai na API (`myApplication.swapRequest.decisionReason`),
  mas a tela do membro não mostra. Pergunta ao usuário: mostrar no aviso do membro, como a
  recusa da candidatura (RN-29)?
- T-15 (2026-10-07): o Svelte descarta o espaço no começo de um `{#if}` quebrado em linha;
  na Parte 2 isso foi corrigido com expressões. O painel do jogador da Parte 1 tem o mesmo
  padrão em "· pendente desde" (provável "Suporte· pendente desde 17:10"). Fora do escopo;
  pergunta ao usuário se corrige.
