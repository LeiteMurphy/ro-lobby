# Spec — Retrato pela classe

- Feature: `retrato-por-classe` · Nível: M · Status: Rascunho
- Notion: (a criar na fase de Tasks)
- Última revisão: 2026-10-09 — primeira versão
- Artes aprovadas: prévia "Artes de classe" (versão 3), 27 PNG em 64×64

## 1. Contexto
Hoje o jogador escolhe um de 4 retratos genéricos para o personagem (RN-10 da
`personagens`). O usuário quer que o retrato mostre a classe. Ele pediu artes no estilo
das ilustrações oficiais, mas o projeto não pode usar arte da Gravity (CLAUDE.md). Por
isso o caminho aprovado é uma arte própria, em pixel art, por linha de classe: o emblema
do arquétipo (arma, cor e tema), sem personagem desenhado. As 27 artes já foram aprovadas
na prévia.

## 2. Research
- O retrato é um campo do personagem (`portrait`, enum `retrato-1` a `retrato-4`) e
  aparece em vários lugares:
  - no card e no diálogo do personagem em `/perfil`;
  - na composição e no painel do jogador no detalhe do lobby;
  - nos pedidos de troca e nos diálogos de candidatura e de troca;
  - em "Minhas candidaturas" e no banco de talentos.
- O catálogo de classes (82 classes, com a família) vive em Go (`internal/catalog`) e é
  servido por `GET /classes`. Ele ainda não diz a linha de cada classe.
- Os retratos atuais são SVG em `web/static/portraits`.

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como jogador, quero que o retrato do meu personagem mostre a linha da classe dele, para reconhecer o personagem de relance. |
| US-02 | P2 | Como jogador olhando um lobby ou o banco de talentos, quero ver o retrato da classe de cada personagem, para entender a composição do grupo. |

## 4. Regras
- **RN-01** — Cada classe do catálogo tem uma arte:
  - a da linha dela: são 20 linhas, como Cavaleiro, Templário, Arruaceiro, Bardo e
    Odalisca;
  - ou a da família, para a primeira classe, que ainda não escolheu a linha
    (Espadachim, Mago, Gatuno, Mercador, Noviço, Arqueiro e Taekwon).
  O Aprendiz usa a arte da linha do Superaprendiz. A tabela de classe para arte fica no
  catálogo (seção 6 do `tasks.md`).
- **RN-02** — O retrato do personagem é a arte da classe dele, sempre. Trocar a classe
  troca o retrato. O jogador não escolhe mais retrato, e os 4 retratos genéricos saem da
  tela de personagem (revisa a RN-10 da `personagens`).
- **RN-03** — Personagens que já existem passam a mostrar a arte da classe sem nenhuma
  ação do jogador.
- **RN-04** — Classe fora do catálogo (borda da RN-06 da `personagens`) mostra uma arte
  neutra, o Retrato 1 de hoje.
- **RN-05** — O retrato aparece em todos os lugares onde hoje aparece o retrato do
  personagem (seção 2), com o nome da linha como texto alternativo, por exemplo "Linha do
  Arruaceiro".
- **RN-06** — As artes são originais do RO Lobby, sem logo, arte nem sprite da Gravity
  (CLAUDE.md), e ficam ampliadas sem borrar (pixel art).

## 5. Critérios de aceite

### US-01 — Retrato do personagem
```gherkin
CA-01.1 — Retrato pela linha  [US-01, RN-01, RN-02, RN-05]
Given o jogador com um personagem Renegado
When ele abre /perfil
Then o card mostra a arte "Linha do Arruaceiro"

CA-01.2 — Trocar a classe troca o retrato  [US-01, RN-02]
Given um personagem Templário
When o jogador edita a classe para Feiticeiro
Then o card passa a mostrar a arte "Linha do Sábio"

CA-01.3 — Primeira classe e Aprendiz  [US-01, RN-01]
Given personagens Gatuno, Taekwon e Aprendiz
Then eles mostram, nesta ordem, "Família Gatuno", "Família Taekwon" e "Linha do
  Superaprendiz"

CA-01.4 — Sem escolha de retrato  [US-01, RN-02]
Given o diálogo de adicionar ou editar personagem
Then ele não tem mais a escolha entre os 4 retratos
  And mostra a arte da classe escolhida assim que a classe é selecionada

CA-01.5 — Personagem que já existia  [US-01, RN-03]
Given um personagem Cardeal criado antes desta mudança, com o Retrato 3
When o jogador abre /perfil
Then o card mostra "Linha do Sacerdote"

CA-01.6 — Classe fora do catálogo  [US-01, RN-04]
Given um personagem com uma classe que saiu do catálogo
Then o card mostra o Retrato 1
```

### US-02 — Retrato em lobbies e no banco de talentos
```gherkin
CA-02.1 — Composição do lobby  [US-02, RN-05]
Given um lobby com o dono Arcebispo e um membro Sicário
When alguém abre o detalhe
Then o dono aparece com "Linha do Sacerdote" e o membro com "Linha do Mercenário"

CA-02.2 — Banco de talentos  [US-02, RN-05]
Given no banco um personagem Maestro
When alguém abre /talentos
Then o card dele mostra "Linha do Bardo"
```

## 6. Casos de borda
- Personagem excluído num lobby (sem dados): continua sem retrato, como hoje [RN-05]
- Super Aprendiz EX e Hiperaprendiz: "Linha do Superaprendiz" [RN-01]
- Kagerou, Oboro, Shinkiro e Shiranui: "Linha do Ninja" [RN-01]

## 7. Requisitos não funcionais
- **RNF-01** — As artes são servidas pelo próprio web, sem recurso externo.
- **RNF-02** — Cada arte tem no máximo 4 KB.
- **RNF-03** — Todo teste cita o ID do critério de aceite.

## 8. Fora de escopo
- Arte diferente por classe dentro da mesma linha (ex.: Renegado e Mandraque iguais).
- Arte por gênero do personagem.
- Escolher uma arte de outra linha.

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na conversa
- Não usar as artes de classe do bROWiki (são da Gravity). Arte própria por linha.
- Pixel art em 64×64, emblema do arquétipo sobre fundo da cor da linha (versão 3 aprovada).
- Retrato automático pela classe; os 4 genéricos deixam de ser escolhidos.
- Aprendiz usa a linha do Superaprendiz; primeiras classes usam a arte da família.
- Ferreiro com martelo perpendicular ao cabo; Superaprendiz como mochila de iniciante.
