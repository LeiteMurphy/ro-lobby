# Spec — Banco de talentos

- Feature: `banco-de-talentos` · Nível: G · Status: Aprovada (2026-10-09)
- Notion: https://app.notion.com/p/3f4d4a3a5eff817fbd09f468784d45a5
- Última revisão: 2026-10-09 — primeira versão

## 1. Contexto
Hoje o dono de um lobby espera que os jogadores achem o grupo na Home. Quem está disposto
a jogar não tem como avisar que está disponível. O banco de talentos deixa o jogador
oferecer um personagem, com os dias, a faixa de horário e as instâncias de interesse.
Assim o dono vê quem combina com o grupo e chama a pessoa pelo Discord. A spec
`personagens` deixou horários e instâncias de interesse para uma feature seguinte, que é
esta. A notificação pelo Discord e o convite dentro do site ficam para depois (seção 8).

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Visitante | Sem sessão | Vê o catálogo `/talentos` sem o nome no Discord |
| Jogador | Usuário logado com personagens | Põe e tira os próprios personagens do banco; vê o catálogo completo |
| Dono | Jogador que criou o lobby | Vê os personagens com afinidade na criação e no detalhe do lobby aberto |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como jogador, quero pôr um personagem no banco de talentos com dias, faixa de horário e instâncias de interesse, para ser achado por quem monta grupos. |
| US-02 | P1 | Como dono de um lobby aberto, quero ver os personagens com afinidade com as vagas abertas, para chamar quem combina com o grupo. |
| US-03 | P2 | Como dono criando um lobby, quero ver quantos personagens têm afinidade com o que estou montando, para escolher dia e horário com mais gente. |
| US-04 | P2 | Como jogador, quero navegar pelo banco com filtros, para achar gente para jogar mesmo sem lobby criado. |

## 4. Regras de domínio

### Disponibilidade
- **RN-01** — A disponibilidade é de cada personagem. O personagem entra no banco quando o
  dono liga "Disponível no banco de talentos" e preenche dias, faixa e instâncias. Ao
  desligar, ele sai do banco, e os dados ficam guardados para quando ligar de novo.
- **RN-02** — Dias: um ou mais dias da semana (domingo a sábado).
- **RN-03** — Faixa: hora de início e de fim em horário de Brasília, de 30 em 30 minutos,
  com início diferente do fim. Se o fim for menor que o início, a faixa passa da
  meia-noite e termina no dia seguinte (ex.: sexta, 22:00–02:00 vai até sábado 02:00).
  O início entra na faixa e o fim não.
- **RN-04** — Instâncias: "Qualquer instância" ou uma ou mais instâncias do catálogo
  (`GET /instances`). Instância que sair do catálogo some da lista do personagem.
- **RN-05** — Só o dono do personagem liga, desliga e edita a disponibilidade. Personagem
  excluído sai do banco.

### Afinidade
- **RN-06** — Um personagem do banco tem afinidade com um lobby quando:
  1. a instância do lobby está entre as de interesse dele, ou ele marcou "Qualquer
     instância";
  2. o início do lobby, em horário de Brasília, cai num dos dias e dentro da faixa dele
     (RN-03);
  3. o nível dele é maior ou igual ao nível mínimo do lobby;
  4. a função dele tem pelo menos uma vaga aberta no lobby;
  5. ele está livre: não é o personagem do dono e não tem candidatura pendente ou aceita
     neste lobby nem em outro lobby não cancelado cuja janela se sobrepõe (RN-10 da
     `lobbies`, RN-11 da `candidatura-lobby`).
- **RN-07** — Personagens de quem é dono do lobby não aparecem na afinidade desse lobby.
- **RN-08** — A lista de afinidade vem ordenada por nível, do maior para o menor, e por
  nick em caso de empate.

### Onde aparece
- **RN-09** — No detalhe do lobby aberto, só o dono vê "Jogadores disponíveis", com os
  personagens com afinidade (RN-06). Lobby iniciado ou cancelado não mostra a lista.
- **RN-10** — Na criação do lobby, a prévia mostra "N jogadores disponíveis" para a
  instância, o dia, a hora, o nível mínimo e as vagas escolhidos. A contagem acompanha o
  formulário. A condição 5 da RN-06 vale para o horário escolhido.
- **RN-11** — A página `/talentos` lista os personagens do banco, com filtros de
  instância, função, dia da semana e hora. Cada filtro é opcional e eles se combinam com
  "E". O filtro de hora mostra quem tem aquela hora dentro da faixa no dia escolhido (ou
  em qualquer dia, sem dia escolhido). A ordem segue a RN-08.
- **RN-12** — Cada personagem na lista mostra retrato, nick, classe, nível, função, dias e
  faixa, instâncias de interesse e o link externo, se houver. Para quem está logado, mostra
  também o nome no Discord (`@username`) do dono do personagem. Para visitante, mostra
  "Entre para ver o Discord".

## 5. Critérios de aceite

### US-01 — Pôr o personagem no banco
```gherkin
CA-01.1 — Entrar no banco  [US-01, RN-01, RN-02, RN-03, RN-04]
Given o Jogador com o personagem "Lirien" (Arcebispo, 178, Suporte)
When ele liga "Disponível no banco de talentos" em Lirien com segunda a sexta,
  das 19:00 às 23:00, e as instâncias "Templo do Demônio Rei" e "Sonho Sombrio"
Then Lirien aparece em /talentos com esses dias, essa faixa e essas instâncias

CA-01.2 — Faixa que passa da meia-noite  [US-01, RN-03]
Given Lirien no banco às sextas, das 22:00 às 02:00
When alguém filtra /talentos por sábado à 01:00
Then Lirien aparece
  And no filtro de sábado às 22:00 ela não aparece

CA-01.3 — Campos obrigatórios  [US-01, RN-02, RN-03, RN-04]
Given o Jogador ligando a disponibilidade de Lirien
When ele salva sem dia, ou com início igual ao fim, ou sem instância e sem "Qualquer"
Then a API responde 422 com o erro no campo
  And Lirien continua fora do banco

CA-01.4 — Sair do banco guarda os dados  [US-01, RN-01]
Given Lirien no banco
When o Jogador desliga a disponibilidade
Then Lirien some de /talentos
  And ao ligar de novo, o formulário vem com os dias, a faixa e as instâncias de antes

CA-01.5 — Personagem de outro Usuário  [US-01, RN-05]
Given o personagem de outro Usuário
When alguém tenta mudar a disponibilidade dele pela API
Then a API responde 404 e nada muda
```

### US-02 — Afinidade no detalhe do lobby
```gherkin
CA-02.1 — Lista do dono  [US-02, RN-06, RN-08, RN-09, RN-12]
Given um lobby aberto de "Templo do Demônio Rei" na quarta às 20:00, nível mínimo 160,
  com vaga aberta de Tank e de Dano e o Suporte cheio
  And no banco: "Brasa" (Tank, 172, quarta 19:00–23:00, Templo),
  "Fogo" (Dano, 200, qualquer dia e "Qualquer instância", 18:00–00:00)
  e "Cura" (Suporte, 200, quarta 19:00–23:00, Templo)
When o dono abre o detalhe
Then "Jogadores disponíveis" mostra Fogo e depois Brasa
  And Cura não aparece, porque o Suporte está cheio

CA-02.2 — Fora da afinidade  [US-02, RN-06]
Given o mesmo lobby
When um personagem do banco não tem a instância, ou não tem a quarta às 20:00,
  ou está abaixo do nível 160
Then ele não aparece na lista

CA-02.3 — Ocupado no horário  [US-02, RN-06, RN-07]
Given "Brasa" com candidatura pendente neste lobby, ou aceito num lobby que começa às 21:00
  do mesmo dia, ou do mesmo Usuário que o dono
When o dono abre o detalhe
Then Brasa não aparece na lista

CA-02.4 — Só o dono e só aberto  [US-02, RN-09]
Given um lobby com personagens com afinidade
When um visitante, um candidato ou o dono de um lobby iniciado ou cancelado abre o detalhe
Then a lista "Jogadores disponíveis" não aparece
```

### US-03 — Prévia na criação
```gherkin
CA-03.1 — Contagem na criação  [US-03, RN-10]
Given os personagens do CA-02.1 no banco
When o dono preenche "Templo do Demônio Rei", quarta às 20:00, nível 160, vagas 1 Tank,
  2 Suporte e 3 Dano, com o próprio personagem de Suporte
Then a prévia mostra "3 jogadores disponíveis"
  And ao mudar a hora para 17:00, a prévia mostra "0 jogadores disponíveis"
```

### US-04 — Catálogo
```gherkin
CA-04.1 — Filtros  [US-04, RN-11]
Given Brasa, Fogo e Cura no banco como no CA-02.1
When alguém filtra /talentos por função Tank
Then só Brasa aparece
  And ao filtrar por instância "Sonho Sombrio", só Fogo aparece

CA-04.2 — Discord só para logado  [US-04, RN-12]
Given Brasa no banco, de um Usuário com o username "brasa_ro"
When um visitante abre /talentos
Then ele vê Brasa sem "@brasa_ro" e com "Entre para ver o Discord"
  And um Jogador logado vê "@brasa_ro"

CA-04.3 — Fora do banco  [US-04, RN-01]
Given um personagem com a disponibilidade desligada
When alguém abre /talentos
Then ele não aparece
```

## 6. Casos de borda
- Faixa 00:00–00:00 → recusada, porque início e fim são iguais [RN-03]
- Lobby às 23:30 e faixa até 23:30 → sem afinidade, porque o fim não entra [RN-03]
- Personagem que sobe ou cai de nível → a afinidade usa o nível atual [RN-06]
- Personagem que muda de função → a afinidade usa a função atual [RN-06]
- Lobby sem vaga aberta → a lista fica vazia, com "Nenhum jogador disponível agora" [RN-06]
- Instância removida do catálogo → some das instâncias do personagem; se não sobrar
  nenhuma e ele não marcou "Qualquer", ele sai da afinidade, mas continua no banco [RN-04]

## 7. Requisitos não funcionais
- **RNF-01** — Dias e faixas são horário de parede de Brasília, recorrentes, e ficam no
  banco como dia da semana e minutos desde a meia-noite. Não são instantes, então não
  ficam em UTC (decisão no `design.md`). O início do lobby, em UTC, é convertido para
  Brasília antes de comparar.
- **RNF-02** — As regras valem na API; o web repete as simples só para avisar antes.
- **RNF-03** — Acessibilidade: o formulário de disponibilidade e os filtros funcionam pelo
  teclado, com rótulos e foco visível.
- **RNF-04** — Todo teste cita o ID do critério de aceite.
- **RNF-05** — O nome no Discord não vai no HTML nem na resposta da API para quem não está
  logado.

## 8. Fora de escopo
- Notificação pelo Discord (convite, aceite) — integração futura.
- Convite dentro do site — junto com a notificação.
- Mais de uma faixa por personagem ou faixas diferentes por dia — pode entrar depois.
- Disponibilidade em datas específicas (ex.: "só no sábado 12/10").
- Reputação, avaliações e histórico de grupos.

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Onde fica → por personagem, com liga/desliga.
- Horários → dias da semana e uma faixa de hora, que pode passar da meia-noite.
- Onde o dono vê → no detalhe do lobby, na criação e numa página de catálogo.
- Ação → ver e chamar por fora (nome no Discord e link). O convite no site fica com a
  integração do Discord.
- Privacidade → o nome no Discord só para quem está logado.
- Afinidade → instância e horário, nível mínimo, função com vaga e livre no horário.
- Instâncias → "Qualquer instância" ou uma ou mais do catálogo.
