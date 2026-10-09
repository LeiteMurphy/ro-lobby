# Spec — Grupo livre

- Feature: `grupo-livre` · Nível: G · Status: Rascunho
- Notion: (a criar na fase de Tasks)
- Última revisão: 2026-10-09 — primeira versão

## 1. Contexto
Todo lobby hoje tem formação por função: vagas de Tank, Suporte e Dano, e cada personagem
só entra na vaga da própria função. Muitos grupos do jogo não precisam disso: caça a MVP,
farm, evento da guilda, grupo de amigos. Para esses casos, o dono quer abrir até 12 lugares
e deixar entrar quem chegar, sem montar formação de batalha. O "Grupo livre" é um segundo
tipo de formação do lobby. Ele mexe em regras de várias specs (`lobbies`,
`candidatura-lobby`, `home-local`, `banco-de-talentos`, `compartilhar-lobby`), e esta spec
diz o que muda em cada uma quando o lobby é livre. Para o lobby "Por função", nada muda.

## 2. Research
- A formação por função está em todas as camadas. O banco tem as vagas por função e a
  função do dono no lobby, e a função em candidaturas e pedidos de troca.
- No backend: `checkSlots`, a vaga da função do dono (RN-08 da `lobbies`), as vagas não
  abaixo dos ocupantes (RN-18), o `freeSlots` da candidatura e do aceite, e o `checkSwap`
  das trocas.
- No web: a composição da Home e do detalhe, o filtro "Vaga para", a elegibilidade dos
  diálogos e o texto de vagas do convite.
- A função do personagem continua existindo em todo lugar: é dado do personagem e
  aparece nas listas.

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como dono, quero criar um lobby "Grupo livre" com 2 a 12 vagas, para montar um grupo sem formação de batalha. |
| US-02 | P1 | Como jogador, quero me candidatar a um grupo livre com qualquer personagem, para entrar sem depender da função. |
| US-03 | P2 | Como jogador na Home, quero reconhecer o grupo livre e achá-lo pelo filtro de vaga, para ver todos os grupos em que caibo. |
| US-04 | P2 | Como dono, quero editar o grupo livre e trocar a formação enquanto estou sozinho no grupo, para corrigir a escolha. |

## 4. Regras

### Formação
- **RN-01** — Todo lobby tem uma formação: "Por função" (a de hoje) ou "Grupo livre". Na
  criação, o padrão é "Por função".
- **RN-02** — Grupo livre: o dono escolhe de 2 a 12 vagas, com padrão 12. O personagem do
  dono ocupa uma delas desde a criação (P-02 da `lobbies`), com qualquer função.
- **RN-03** — No grupo livre, qualquer personagem ocupa qualquer vaga. A função do
  personagem continua guardada na candidatura e aparece como informação, mas não decide
  vaga.

### Candidatura e trocas (substituem, no grupo livre, as regras por função da `candidatura-lobby`)
- **RN-04** — Candidatar e aceitar exigem uma vaga livre no total do grupo (no lugar das
  RN-05 e RN-10 da `candidatura-lobby`). Aceites simultâneos não passam do total (RN-12).
  As outras regras da candidatura valem iguais: nível mínimo, conflito de horário,
  bloqueio, uma candidatura ativa por pessoa e justificativas.
- **RN-05** — Trocar o personagem (do dono, RN-19, ou do membro, RN-20 e RN-23) não
  depende de vaga: o personagem novo fica com a mesma vaga. Nível mínimo e conflito de
  horário continuam valendo.

### Edição
- **RN-06** — Na edição de um grupo livre, as vagas não ficam abaixo dos ocupantes (o dono
  e os membros aceitos), no lugar da RN-18 por função.
- **RN-07** — O dono só troca a formação enquanto o grupo tem só ele: sem membro aceito
  e sem candidatura pendente. Indo para "Por função", vale a RN-08 da `lobbies`: a função
  do personagem do dono precisa ter vaga.

### Telas e textos
- **RN-08** — Card da Home e destaque mostram o selo "Grupo livre", "X de N" e uma barra
  de ocupação, no lugar dos chips por função. A borda do card usa uma cor neutra própria,
  e não a da função com mais vagas.
- **RN-09** — No filtro "Vaga para" da Home, um grupo livre com vaga conta para qualquer
  função marcada. A contagem de cada função na lateral também conta o grupo livre com
  vaga (RN-12 e RN-13 da `home-local`).
- **RN-10** — No detalhe, a "Composição" vira uma lista única de N lugares, sem blocos
  por função. Mostra o dono com o selo "Anfitrião", os membros aceitos e "Vaga aberta" nos
  lugares livres. Cada pessoa mostra a função do personagem como informação.
- **RN-11** — Diálogos de candidatura e troca: no grupo livre, nenhum personagem fica de
  fora por falta de vaga na função. Ficam de fora só os abaixo do nível mínimo ou em
  conflito de horário.
- **RN-12** — Convite e preview (RN-03 e RN-05 da `compartilhar-lobby`): no grupo livre,
  `<vagas>` é "Vagas: N livres" (ou "1 livre"), ou "Grupo lotado".
- **RN-13** — Banco de talentos (RN-06.4 e RN-10 da `banco-de-talentos`): no grupo livre,
  a condição de função vira "o grupo tem vaga livre", para qualquer função. Na prévia da
  criação, as vagas livres são as do formulário menos a do dono.

## 5. Critérios de aceite

### US-01 — Criar grupo livre
```gherkin
CA-01.1 — Criar com o padrão  [US-01, RN-01, RN-02]
Given o dono na criação de lobby
When ele escolhe "Grupo livre" e não mexe nas vagas
Then o lobby é criado com 12 vagas, 1 ocupada pelo personagem dele
  And o detalhe mostra 11 lugares com "Vaga aberta"

CA-01.2 — Limites das vagas  [US-01, RN-02]
Given o dono criando um grupo livre
When ele pede 1 vaga, ou 13 vagas
Then a criação é recusada com o erro nas vagas

CA-01.3 — Padrão continua por função  [US-01, RN-01]
Given o dono abrindo a criação
Then a formação vem marcada "Por função", com 1 Tank, 2 Suportes e 3 Danos
```

### US-02 — Candidatura no grupo livre
```gherkin
CA-02.1 — Qualquer função entra  [US-02, RN-03, RN-04]
Given um grupo livre de 3 vagas, com o dono (Suporte) e 1 membro Suporte aceito
When um jogador se candidata com outro Suporte e o dono aceita
Then o grupo fica 3 de 3, com três Suportes

CA-02.2 — Grupo cheio  [US-02, RN-04]
Given um grupo livre com todas as vagas ocupadas
When alguém se candidata, ou o dono tenta aceitar uma pendente
Then a API recusa com "Esse grupo não tem mais vaga."

CA-02.3 — Aceites ao mesmo tempo  [US-02, RN-04]
Given um grupo livre com 1 vaga e 2 candidaturas pendentes
When o dono aceita as duas ao mesmo tempo
Then só uma é aceita

CA-02.4 — Troca sem vaga de função  [US-02, RN-05]
Given um membro Dano num grupo livre cheio
When ele troca para um personagem Tank e o dono aceita a troca
Then a troca é aceita e o grupo continua cheio

CA-02.5 — Diálogo de candidatura  [US-02, RN-11]
Given um grupo livre com vaga e nível mínimo 160
When um jogador com personagens de nível 170 (Tank) e 150 (Dano) abre "Candidatar"
Then o Tank aparece disponível, sem aviso de função
  And o Dano aparece indisponível pelo nível
```

### US-03 — Home
```gherkin
CA-03.1 — Card do grupo livre  [US-03, RN-08]
Given um grupo livre de 12 vagas com 4 ocupadas
When alguém abre a Home no dia dele
Then o card mostra "Grupo livre" e "4 de 12", sem chips por função

CA-03.2 — Filtro de vaga  [US-03, RN-09]
Given um grupo livre com vaga e um lobby por função sem vaga de Tank
When alguém marca "Vaga para: Tank"
Then o grupo livre aparece e o lobby por função não
  And a contagem de Tank na lateral conta o grupo livre
```

### US-04 — Edição
```gherkin
CA-04.1 — Vagas abaixo dos ocupantes  [US-04, RN-06]
Given um grupo livre com 5 ocupantes
When o dono tenta deixar 4 vagas
Then a edição é recusada com o erro nas vagas

CA-04.2 — Trocar a formação sozinho  [US-04, RN-07]
Given um grupo livre só com o dono (Suporte), sem candidatura pendente
When ele troca para "Por função" com 1 Tank, 2 Suportes e 3 Danos
Then o lobby passa a ter vagas por função, com o dono numa vaga de Suporte

CA-04.3 — Formação travada com gente  [US-04, RN-07]
Given um lobby com um membro aceito ou uma candidatura pendente
When o dono tenta trocar a formação
Then a edição é recusada com "Só dá para trocar a formação com o grupo vazio"
```

### Textos
```gherkin
CA-05.1 — Convite do grupo livre  [RN-12]
Given um grupo livre de 12 vagas com 4 ocupadas
When alguém copia o convite
Then a segunda linha começa com "Vagas: 8 livres"

CA-05.2 — Banco de talentos no grupo livre  [RN-13]
Given um grupo livre com vaga
  And no banco um Tank, um Suporte e um Dano com afinidade de instância, horário e nível
When o dono abre o detalhe
Then os três aparecem em "Jogadores disponíveis"
```

## 6. Casos de borda
- Grupo livre de 2 vagas: o dono e mais uma pessoa [RN-02]
- Membro sai ou é removido: a vaga volta a ser livre, sem função [RN-04]
- Lobby por função já existente: continua igual; a formação dele é "Por função" [RN-01]
- Grupo livre com pedido de troca pendente: aceitar não depende de vaga [RN-05]
- "Minhas candidaturas" num grupo livre: mostra a função do personagem, como hoje [RN-03]

## 7. Requisitos não funcionais
- **RNF-01** — Lobbies existentes ficam "Por função" sem perder dado (migração com padrão).
- **RNF-02** — As regras valem na API; o web repete as simples só para avisar antes.
- **RNF-03** — A escolha da formação funciona pelo teclado, com rótulo.
- **RNF-04** — Todo teste cita o ID do critério de aceite.

## 8. Fora de escopo
- Formações mistas (ex.: 1 Tank obrigatório e o resto livre) — pode entrar depois.
- Mais de 12 vagas ou grupos de guilda.
- Mudar a formação com o grupo já montado.
- Notificação pelo Discord.

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Nome → "Grupo livre".
- Vagas → o dono escolhe de 2 a 12, padrão 12; ele ocupa uma.
- Função → só informativa; o filtro de vaga e o banco de talentos tratam o grupo livre
  como vaga para qualquer função.
- Trocar a formação na edição → só com o grupo vazio (só o dono).
