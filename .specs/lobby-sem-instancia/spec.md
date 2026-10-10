# Spec — Lobby sem instância

- Feature: `lobby-sem-instancia` · Nível: M · Status: Aprovada (2026-10-09)
- Notion: https://app.notion.com/p/3f5d4a3a5eff81bb924bcb3af470ab83
- Última revisão: 2026-10-09 — primeira versão

## 1. Contexto
Todo lobby hoje é de uma instância do catálogo. Ela dá o título do card, o nível mínimo
padrão, o filtro da Home e a afinidade do banco de talentos. Muitos grupos não são de
uma instância: caça a MVP, farm, quest, evento. O dono quer abrir o lobby sem prender a
uma instância, com um título dele. É independente da formação: um lobby sem instância
pode ser "Por função" ou "Grupo livre" (spec `grupo-livre`).

## 2. Research
- Hoje a instância é obrigatória em toda a pilha:
  - no banco, `instance_id`, `instance_name` e `instance_level` são NOT NULL, com CHECK
    `min_level >= instance_level`;
  - a API valida a instância no catálogo;
  - o web usa o nome no card, no detalhe, no convite e no preview do link.
- O nome da instância já é copiado para o lobby (D-02 da `lobbies`). O título livre pode
  ocupar esse lugar.
- A Home filtra por nome de instância (RN-12 da `home-local`). O banco de talentos filtra
  pela instância de interesse (RN-06.1 da `banco-de-talentos`).

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como dono, quero criar um lobby sem instância, com um título meu, para montar grupos que não são de uma instância. |
| US-02 | P2 | Como jogador na Home, quero achar os lobbies sem instância pelo filtro, para ver esses grupos separados. |
| US-03 | P2 | Como dono, quero trocar entre instância e "sem instância" na edição, para corrigir a escolha. |

## 4. Regras
- **RN-01** — Na criação, a instância é uma do catálogo ou "Sem instância definida". O
  padrão continua sendo uma instância do catálogo (RN-02 da `lobbies`).
- **RN-02** — Sem instância, o dono pode escrever um título de até 40 caracteres, sem
  espaço nas pontas. O título é opcional. Vazio, o lobby aparece como "Qualquer
  instância".
- **RN-03** — Sem instância, o nível mínimo vai de 1 a 275, com padrão 1. Continua
  valendo que o personagem do dono tem o nível mínimo (RN-08 e RN-18 da `lobbies`).
- **RN-04** — O título (ou "Qualquer instância") ocupa o lugar do nome da instância no
  card, no destaque, no detalhe, no convite e no preview do link (RN-02 e RN-05 da
  `compartilhar-lobby`). O card e o detalhe usam a capa e o ícone genéricos (RN-03 da
  `lobbies`) e não mostram o tipo de retorno.
- **RN-05** — Na Home, o filtro de instância ganha a opção "Sem instância definida", que
  traz só os lobbies sem instância. "Todas" traz todos. Filtrando por uma instância do
  catálogo, os lobbies sem instância não aparecem.
- **RN-06** — Na edição, o dono troca entre uma instância e "Sem instância definida" e
  edita o título (RN-17 da `lobbies`). Indo para uma instância, o nível mínimo sugerido
  volta a ser o de entrada dela. Indo para "sem instância", o título começa vazio.
- **RN-07** — Banco de talentos: no lobby sem instância, a condição de instância da
  afinidade (RN-06.1 da `banco-de-talentos`) vale para todo personagem do banco, porque
  o lobby aceita qualquer instância. As outras condições continuam: dia e faixa, nível,
  função ou vaga, livre no horário. Vale também para a prévia da criação (RN-10).
- **RN-08** — Os outros pontos do lobby continuam iguais: horário, vagas, formação,
  candidatura, trocas, cancelamento e compartilhar.

## 5. Critérios de aceite

### US-01 — Criar sem instância
```gherkin
CA-01.1 — Com título  [US-01, RN-01, RN-02, RN-04]
Given o dono na criação de lobby
When ele escolhe "Sem instância definida", escreve "Caça ao MVP" e cria
Then o detalhe e o card da Home mostram "Caça ao MVP", com a capa genérica
  And o convite começa com "Grupo para Caça ao MVP"

CA-01.2 — Sem título  [US-01, RN-02, RN-04]
Given o dono criando um lobby sem instância
When ele deixa o título vazio
Then o lobby aparece como "Qualquer instância"

CA-01.3 — Título longo  [US-01, RN-02]
Given o dono criando um lobby sem instância
When ele escreve um título de 41 caracteres
Then a criação é recusada com o erro no título

CA-01.4 — Nível mínimo livre  [US-01, RN-03]
Given o dono com um personagem de nível 120
When ele cria um lobby sem instância
Then o nível mínimo vem 1
  And aceita qualquer valor de 1 até 120, e recusa acima do nível do personagem

CA-01.5 — Padrão continua com instância  [US-01, RN-01]
Given o dono abrindo a criação
Then o campo de instância vem com uma instância do catálogo, como hoje
```

### US-02 — Filtro da Home
```gherkin
CA-02.1 — Opção própria no filtro  [US-02, RN-05]
Given no dia um lobby sem instância e um de "Templo do Demônio Rei"
When o visitante escolhe "Sem instância definida"
Then só o lobby sem instância aparece
  And escolhendo "Templo do Demônio Rei", só o do Templo
  And em "Todas", os dois
```

### US-03 — Edição
```gherkin
CA-03.1 — Trocar para sem instância e de volta  [US-03, RN-06]
Given um lobby aberto de "Templo do Demônio Rei"
When o dono troca para "Sem instância definida" com o título "Farm" e salva
Then o lobby aparece como "Farm"
  And ao trocar de volta para "Sonho Sombrio", o nível mínimo sugerido é 120
```

### Banco de talentos
```gherkin
CA-04.1 — Afinidade sem instância  [RN-07]
Given um lobby sem instância com vaga, quarta às 20:00
  And no banco um personagem só com "Sonho Sombrio", disponível na quarta 19–23
When o dono abre o detalhe
Then o personagem aparece em "Jogadores disponíveis"
```

## 6. Casos de borda
- Título só com espaços → fica vazio e o lobby aparece como "Qualquer instância" [RN-02]
- Instância que sair do catálogo depois → continua como hoje (D-02 da `lobbies`), não
  vira "sem instância" [RN-08]
- Lobby sem instância e "Grupo livre" ao mesmo tempo → os dois valem juntos [RN-08]

## 7. Requisitos não funcionais
- **RNF-01** — Lobbies existentes continuam com a instância deles, sem perder dado.
- **RNF-02** — As regras valem na API; o web repete as simples só para avisar antes.
- **RNF-03** — A escolha e o título funcionam pelo teclado, com rótulo.
- **RNF-04** — Todo teste cita o ID do critério de aceite.

## 8. Fora de escopo
- Lista de atividades pré-definidas (MVP, farm, quest) — o título livre resolve por ora.
- Arte por atividade.
- Mais de uma instância no mesmo lobby.

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Título → livre, até 40 caracteres; vazio aparece como "Qualquer instância".
- Nível mínimo → de 1 a 275, padrão 1.
- Filtro da Home → opção própria "Sem instância definida"; não aparece filtrando por uma
  instância.
- Ordem → depois das faixas de horário dinâmicas.
- Banco de talentos (proposta desta spec) → todo personagem do banco tem a condição de
  instância cumprida.
