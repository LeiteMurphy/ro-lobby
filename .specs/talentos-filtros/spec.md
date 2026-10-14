# Spec — Filtros do banco de talentos e listagens com maiúscula

- Feature: `talentos-filtros` · Nível: P · Status: Aprovada (2026-10-14)
- Notion: https://app.notion.com/p/3f9d4a3a5eff8100a555f25ef45c2597
- Última revisão: 2026-10-14 — primeira versão

## 1. Contexto
No teste do `rolobby.com.br`, o usuário escolheu "Função: Tank" em `/talentos` e a lista
continuou com Danos e Suportes. A busca estava certa, mas a escolha só valia depois de
"Filtrar" (RN-11 da `banco-de-talentos`, feita para funcionar sem JS), e a tela não dava
sinal disso. O usuário também pediu as listagens com a primeira letra maiúscula.

## 2. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como jogador, quero que a lista do banco de talentos mude assim que eu troco um filtro, para não achar que o filtro não funciona. |
| US-02 | P1 | Como jogador, quero as listagens do banco de talentos com a primeira letra maiúscula, para a tela ficar mais legível. |

## 3. Regras
- **RN-01** — Em `/talentos`, mudar qualquer filtro (instância, função, servidor, dia ou
  hora) aplica a busca na hora, sem clicar em "Filtrar". O link passa a ter o filtro,
  como já acontece com "Filtrar" (RN-18 da `servidor`). Sem JavaScript, "Filtrar" continua
  aplicando.
- **RN-02** — Os dias do banco de talentos começam com maiúscula, em toda listagem:
  - no filtro "Dia" de `/talentos`: "Segunda", "Terça", …, "Domingo";
  - nos cards (`/talentos` e "Jogadores disponíveis"): "Todos os dias", "Seg a Sex" e
    "Seg, Qua, Sex";
  - nos dias do diálogo de disponibilidade: "Seg", "Ter", …
  As regras de montagem do texto (RN-02 da `banco-de-talentos`) não mudam.

## 4. Critérios de aceite
```gherkin
CA-01.1 — Filtro aplica ao mudar  [US-01, RN-01]
Given o banco com personagens de Dano e de Suporte, nenhum Tank
When alguém escolhe "Função: Tank" em /talentos, sem clicar em "Filtrar"
Then a lista mostra "Ninguém no banco com esses filtros."
  And o link passa a ter funcao=tank

CA-01.2 — Sem JavaScript, "Filtrar" continua  [US-01, RN-01]
Given /talentos sem JavaScript
When alguém escolhe um filtro e clica em "Filtrar"
Then a lista vem filtrada

CA-02.1 — Dias com maiúscula  [US-02, RN-02]
Given personagens com todos os dias, com segunda a sexta e com segunda, quarta e sexta
When alguém vê os cards
Then os dias aparecem como "Todos os dias", "Seg a Sex" e "Seg, Qua, Sex"
  And o filtro "Dia" lista "Segunda" a "Domingo"
  And o diálogo de disponibilidade mostra "Seg" a "Dom"
```

## 5. Fora de escopo
- Outras telas com dias em minúscula (convite e preview do lobby, "Hoje, ter, 6 out" da
  criação). Ficam como estão até o usuário pedir; o convite segue a RN-02 da
  `compartilhar-lobby`.
- Mudar quais filtros existem ou como a busca funciona na API.
