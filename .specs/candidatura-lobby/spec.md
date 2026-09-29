# Spec — Candidatura a lobby

- Feature: `candidatura-lobby` · Nível: G · Status: Aprovada
- Notion: [Épico](https://app.notion.com/p/3ead4a3a5eff810bb7f4eae87f4ab3b2)
- Última revisão: 2026-09-29 — aprovada pelo usuário; perguntas em aberto resolvidas: retirada do pedido de troca,
  visibilidade, bloqueio na remoção

## 1. Contexto
Montar grupo para instância difícil depende de conseguir as funções certas (1 Tank,
de 2 a 4 Suportes, o resto Dano). Hoje isso é feito na conversa, sem controle de vagas.
Esta feature define como um jogador se candidata a um lobby com um dos seus
personagens, como o dono do lobby aceita ou recusa (a recusa sempre com justificativa)
e como a composição muda depois disso: saída, remoção e troca de personagem.

### Premissas vindas de outras features (ainda sem spec)
Esta spec não define estas regras. Ela só depende de que existam:
- **P-01 (Lobby)** — O lobby tem um dono (o usuário que criou), horário de início e
  de fim em UTC e um número de vagas por função, com no máximo 12 vagas no total.
- **P-02 (Lobby)** — O dono escolhe um dos seus personagens ao criar o lobby, e esse
  personagem ocupa uma vaga da função dele desde a criação.
- **P-03 (Lobby)** — O lobby está *aberto* (antes do início e não cancelado),
  *iniciado* (o horário de início já passou) ou *cancelado*.
- **P-04 (Personagem)** — O personagem pertence a um único usuário e tem uma função
  (Tank, Suporte ou Dano).

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Jogador | Usuário logado com Discord que não é dono do lobby | Candidatar-se, retirar candidatura, sair do grupo, pedir troca de personagem |
| Membro | Jogador com candidatura aceita no lobby | Sair do grupo, pedir troca de personagem |
| Dono | Usuário que criou o lobby | Aceitar ou recusar candidaturas, remover membros, trocar o próprio personagem, decidir pedidos de troca |
| Sistema | Processos automáticos | Expirar pendências, cancelar pendências de personagem excluído |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como jogador, quero me candidatar a um lobby com um dos meus personagens, para participar do grupo. |
| US-02 | P1 | Como dono, quero aceitar ou recusar candidaturas, justificando a recusa, para montar a composição que preciso. |
| US-03 | P1 | Como jogador, quero acompanhar e retirar minhas candidaturas pendentes, para saber onde estou e desistir se precisar. |
| US-04 | P1 | Como jogador, quero que as pendências expirem quando o lobby começa ou é cancelado, para que nada fique pendente sem motivo. |
| US-05 | P2 | Como membro, quero sair do grupo antes do início, para liberar a vaga se eu não puder ir. |
| US-06 | P2 | Como dono, quero remover um membro com justificativa, para ajustar a composição antes do início. |
| US-07 | P2 | Como dono, quero trocar meu personagem no lobby, para cobrir uma função que ninguém preencheu. |
| US-08 | P2 | Como membro, quero pedir ao dono a troca do meu personagem, informando o motivo, para cobrir uma função sem perder a vaga. |

## 4. Regras de domínio

### Candidatura
- **RN-01** — O candidato só se candidata com um personagem que pertence a ele.
- **RN-02** — Um usuário tem no máximo uma candidatura ativa (pendente ou aceita) por
  lobby, com qualquer personagem.
- **RN-03** — O dono não se candidata ao próprio lobby.
- **RN-04** — A função da candidatura é a função do personagem no momento da
  candidatura. O candidato não escolhe outra função.
- **RN-05** — Só é possível se candidatar a um lobby aberto em que a função do
  personagem tenha pelo menos uma vaga livre. Vagas livres da função = vagas da função
  − ocupantes da função (dono e membros aceitos).
- **RN-06** — A mensagem do candidato é opcional e tem no máximo 250 caracteres.
- **RN-07** — Quem teve candidatura recusada num lobby, ou foi removido dele com
  bloqueio (RN-15), não pode se candidatar de novo a esse lobby, com nenhum
  personagem. Quem foi removido sem bloqueio pode se candidatar de novo.

### Decisão do dono
- **RN-08** — Só o dono aceita ou recusa, e só candidaturas pendentes.
- **RN-09** — Toda justificativa (recusa, remoção, motivo e recusa de troca) é
  obrigatória e tem de 10 a 250 caracteres, contados depois de remover os espaços
  nas pontas.
- **RN-10** — O aceite só acontece se a função ainda tiver vaga livre no momento do
  aceite. Se não tiver, o aceite falha e a candidatura continua pendente.
- **RN-11** — O aceite é bloqueado se o personagem ocupar vaga (como dono ou membro)
  em outro lobby não cancelado cujo intervalo [início, fim) se sobrepõe ao deste.
  Ter candidatura pendente em lobbies sobrepostos é permitido.
- **RN-12** — Dois aceites simultâneos nunca deixam uma função com mais ocupantes do
  que vagas.

### Ciclo de vida
- **RN-13** — O candidato pode retirar a própria candidatura pendente a qualquer
  momento enquanto o lobby estiver aberto.
- **RN-14** — O membro pode sair do grupo enquanto o lobby estiver aberto. A vaga
  volta a ficar livre.
- **RN-15** — O dono pode remover um membro enquanto o lobby estiver aberto, com
  justificativa (RN-09). A vaga volta a ficar livre. Na remoção, o dono escolhe se
  bloqueia o usuário neste lobby (por exemplo, por comportamento desrespeitoso). O
  bloqueio vale só para este lobby.
- **RN-16** — Quando o lobby inicia ou é cancelado, todas as candidaturas e pedidos
  de troca pendentes dele passam para *expirada*.
- **RN-17** — Os estados da candidatura são: *pendente*, *aceita*, *recusada*,
  *retirada*, *expirada*, *saiu*, *removida* e *cancelada*. As únicas transições
  permitidas são:
  - pendente → aceita | recusada | retirada | expirada | cancelada
  - aceita → saiu | removida | cancelada
  - recusada, retirada, expirada, saiu, removida e cancelada são finais.
- **RN-18** — Toda transição registra quem fez, quando (UTC) e a justificativa ou
  mensagem, quando houver.

### Troca de personagem
- **RN-19** — O dono pode trocar o próprio personagem no lobby, sem aprovação,
  enquanto o lobby estiver aberto. O personagem novo precisa ser dele, ter vaga
  livre na sua função (a vaga que o dono deixa conta como livre) e não pode ter
  conflito de horário (RN-11).
- **RN-20** — O membro pode abrir um pedido de troca informando um personagem novo
  (dele) e um motivo (RN-09). Ele tem no máximo um pedido pendente por lobby.
- **RN-21** — Enquanto o pedido está pendente, o membro continua no grupo com o
  personagem atual.
- **RN-22** — Só o dono aceita ou recusa o pedido de troca. A recusa exige
  justificativa (RN-09).
- **RN-23** — O pedido só é aceito se o personagem novo tiver vaga livre na sua
  função (a vaga que o membro deixa conta como livre) e não tiver conflito de horário
  (RN-11). Se alguma dessas condições falhar, o aceite falha e o pedido continua
  pendente. A troca aceita substitui o personagem de uma vez, sem deixar o membro
  fora do grupo em nenhum momento.
- **RN-24** — Os estados do pedido de troca são *pendente*, *aceito*, *recusado*,
  *retirado*, *expirado* e *cancelado*. Todos, menos *pendente*, são finais. Se o
  membro sair ou for removido, o pedido pendente dele passa para *cancelado*.
- **RN-27** — O membro pode retirar o próprio pedido de troca pendente enquanto o
  lobby estiver aberto. Ele continua no grupo com o personagem atual e pode abrir um
  pedido novo depois.

### Visibilidade
- **RN-28** — Os detalhes de uma candidatura pendente (personagem, mensagem) são
  visíveis só para o dono do lobby e para o próprio candidato. Os outros usuários
  veem a composição (ocupantes por função) e a quantidade de candidaturas pendentes
  do lobby.
- **RN-29** — A justificativa de recusa ou remoção é visível só para o dono do lobby
  e para o candidato afetado.

### Personagem
- **RN-25** — Um personagem com candidatura pendente, vaga ocupada num lobby aberto
  ou pedido de troca pendente não pode ter a função alterada.
- **RN-26** — Quando um personagem é excluído, as candidaturas dele (pendentes e
  aceitas) passam para *cancelada*, liberando as vagas que ele ocupava, e os pedidos
  de troca pendentes em que ele aparece passam para *cancelado*.

## 5. Critérios de aceite

### US-01 — Candidatar-se a um lobby
```gherkin
CA-01.1 — Candidatura com vaga livre  [US-01, RN-01, RN-04, RN-05]
Given um lobby aberto com 1 vaga livre de Suporte
  And o jogador tem um personagem com função Suporte
When ele se candidata com esse personagem
Then é criada uma candidatura pendente com a função Suporte
  And a vaga continua livre até o dono aceitar

CA-01.2 — Candidatura com mensagem  [US-01, RN-06]
Given um lobby aberto com vaga livre na função do personagem
When o jogador se candidata com uma mensagem de 250 caracteres
Then a candidatura é criada com a mensagem

CA-01.3 — Mensagem longa demais  [US-01, RN-06]
Given um lobby aberto com vaga livre na função do personagem
When o jogador se candidata com uma mensagem de 251 caracteres
Then a candidatura é rejeitada com o erro "mensagem acima de 250 caracteres"
  And nenhuma candidatura é criada

CA-01.4 — Personagem de outro usuário  [US-01, RN-01]
Given um personagem que pertence a outro usuário
When o jogador tenta se candidatar com ele
Then a candidatura é rejeitada com o erro "personagem não pertence ao usuário"

CA-01.5 — Segunda candidatura no mesmo lobby  [US-01, RN-02]
Given o jogador já tem uma candidatura pendente no lobby com o personagem A
When ele tenta se candidatar ao mesmo lobby com o personagem B
Then a candidatura é rejeitada com o erro "já existe candidatura ativa neste lobby"

CA-01.6 — Dono no próprio lobby  [US-01, RN-03]
Given o usuário é o dono do lobby
When ele tenta se candidatar ao próprio lobby
Then a candidatura é rejeitada com o erro "dono não se candidata ao próprio lobby"

CA-01.7 — Função sem vaga  [US-01, RN-05]
Given um lobby aberto com todas as vagas de Tank ocupadas
When o jogador tenta se candidatar com um personagem Tank
Then a candidatura é rejeitada com o erro "função sem vaga"

CA-01.8 — Lobby que não está aberto  [US-01, RN-05]
Given um lobby iniciado ou cancelado
When o jogador tenta se candidatar
Then a candidatura é rejeitada com o erro "lobby não está aberto"

CA-01.9 — Nova candidatura depois de recusa  [US-01, RN-07]
Given o jogador teve candidatura recusada no lobby
When ele tenta se candidatar de novo ao mesmo lobby com qualquer personagem
Then a candidatura é rejeitada com o erro "candidatura recusada anteriormente neste lobby"

CA-01.10 — Pendência com horário sobreposto é permitida  [US-01, RN-11]
Given o personagem ocupa vaga num lobby das 20:00 às 22:00 UTC
  And outro lobby aberto das 21:00 às 23:00 UTC tem vaga na função dele
When o jogador se candidata ao segundo lobby com esse personagem
Then a candidatura é criada como pendente
```

### US-02 — Aceitar ou recusar candidatura
```gherkin
CA-02.1 — Aceite com vaga livre  [US-02, RN-08, RN-10]
Given uma candidatura pendente para Dano
  And a função Dano tem 1 vaga livre
When o dono aceita a candidatura
Then a candidatura passa para aceita
  And a função Dano fica com 0 vagas livres

CA-02.2 — Recusa com justificativa  [US-02, RN-08, RN-09]
Given uma candidatura pendente
When o dono recusa com a justificativa "precisamos de mais dano mágico"
Then a candidatura passa para recusada
  And a justificativa fica registrada na candidatura

CA-02.3 — Recusa sem justificativa  [US-02, RN-09]
Given uma candidatura pendente
When o dono recusa sem justificativa
Then a recusa é rejeitada com o erro "justificativa obrigatória"
  And a candidatura continua pendente

CA-02.4 — Justificativa fora do tamanho  [US-02, RN-09]
Given uma candidatura pendente
When o dono recusa com uma justificativa que tem 9 caracteres depois de remover os espaços nas pontas
Then a recusa é rejeitada com o erro "justificativa deve ter de 10 a 250 caracteres"
  And a candidatura continua pendente

CA-02.5 — Limites da justificativa  [US-02, RN-09]
Given uma candidatura pendente
When o dono recusa com uma justificativa de exatamente 10 ou exatamente 250 caracteres
Then a candidatura passa para recusada

CA-02.6 — Quem não é dono  [US-02, RN-08]
Given uma candidatura pendente
  And o usuário logado não é o dono do lobby
When ele tenta aceitar ou recusar a candidatura
Then a ação é rejeitada com o erro "apenas o dono pode decidir"
  And a candidatura continua pendente

CA-02.7 — Candidatura que não está pendente  [US-02, RN-08, RN-17]
Given uma candidatura recusada, retirada ou expirada
When o dono tenta aceitá-la
Then o aceite é rejeitado com o erro "candidatura não está pendente"

CA-02.8 — Aceite com função lotada  [US-02, RN-10]
Given duas candidaturas pendentes para Tank
  And a função Tank tem 1 vaga livre
  And o dono já aceitou a primeira
When o dono tenta aceitar a segunda
Then o aceite é rejeitado com o erro "vaga preenchida"
  And a segunda candidatura continua pendente

CA-02.9 — Aceites simultâneos  [US-02, RN-12]
Given duas candidaturas pendentes para Tank
  And a função Tank tem 1 vaga livre
When os dois aceites são processados ao mesmo tempo
Then exatamente um deles é concluído
  And o outro falha com o erro "vaga preenchida"
  And a função Tank tem 1 ocupante

CA-02.10 — Conflito de horário no aceite  [US-02, RN-11]
Given o personagem ocupa vaga num lobby das 20:00 às 22:00 UTC
  And ele tem candidatura pendente noutro lobby das 21:00 às 23:00 UTC
When o dono do segundo lobby aceita a candidatura
Then o aceite é rejeitado com o erro "personagem em outro lobby no mesmo horário"
  And a candidatura continua pendente

CA-02.11 — Horários encostados não conflitam  [US-02, RN-11]
Given o personagem ocupa vaga num lobby das 20:00 às 22:00 UTC
  And ele tem candidatura pendente noutro lobby das 22:00 às 23:00 UTC
When o dono do segundo lobby aceita a candidatura
Then a candidatura passa para aceita

CA-02.12 — Lobby cancelado não gera conflito  [US-02, RN-11]
Given o personagem estava aceito num lobby das 20:00 às 22:00 UTC que foi cancelado
  And ele tem candidatura pendente noutro lobby das 21:00 às 23:00 UTC
When o dono do segundo lobby aceita a candidatura
Then a candidatura passa para aceita
```

### US-03 — Acompanhar e retirar candidatura
```gherkin
CA-03.1 — Lista das próprias candidaturas  [US-03, RN-17]
Given o jogador tem candidaturas em estados diferentes
When ele consulta as próprias candidaturas
Then vê, para cada uma, o lobby, o personagem, o estado atual
  And vê a justificativa quando o estado é recusada ou removida

CA-03.2 — Retirar candidatura pendente  [US-03, RN-13]
Given o jogador tem uma candidatura pendente num lobby aberto
When ele retira a candidatura
Then a candidatura passa para retirada

CA-03.3 — Retirar candidatura que não está pendente  [US-03, RN-13, RN-17]
Given o jogador tem uma candidatura aceita
When ele tenta retirá-la
Then a ação é rejeitada com o erro "candidatura não está pendente"
  And a candidatura continua aceita

CA-03.4 — Retirar candidatura de outra pessoa  [US-03, RN-13]
Given uma candidatura pendente de outro jogador
When o usuário logado tenta retirá-la
Then a ação é rejeitada com o erro "candidatura não pertence ao usuário"

CA-03.5 — Nova candidatura depois de retirar  [US-03, RN-02, RN-07]
Given o jogador retirou a candidatura dele no lobby
  And o lobby continua aberto com vaga na função
When ele se candidata de novo
Then uma nova candidatura pendente é criada

CA-03.6 — Histórico de transições  [US-03, RN-18]
Given uma candidatura que foi criada e depois recusada
When o histórico dela é consultado
Then há um registro por transição com autor, data e hora em UTC e justificativa, quando houver

CA-03.7 — Terceiros veem só a quantidade de pendentes  [US-03, RN-28]
Given um lobby com 3 candidaturas pendentes
  And o usuário logado não é o dono nem um dos candidatos
When ele consulta o lobby
Then vê a composição e a quantidade 3 de candidaturas pendentes
  And não vê personagem nem mensagem de nenhuma candidatura pendente

CA-03.8 — Justificativa restrita  [US-03, RN-29]
Given uma candidatura recusada com justificativa
  And o usuário logado não é o dono do lobby nem o candidato
When ele consulta o lobby
Then não vê a justificativa da recusa
```

### US-04 — Expiração automática
```gherkin
CA-04.1 — Início do lobby expira pendências  [US-04, RN-16]
Given um lobby com candidaturas pendentes e um pedido de troca pendente
When chega o horário de início do lobby
Then todas as candidaturas pendentes passam para expirada
  And o pedido de troca pendente passa para expirado
  And as candidaturas aceitas continuam aceitas

CA-04.2 — Cancelamento do lobby expira pendências  [US-04, RN-16]
Given um lobby aberto com candidaturas pendentes
When o lobby é cancelado
Then todas as candidaturas pendentes passam para expirada

CA-04.3 — Candidatura expirada não é decidida  [US-04, RN-16, RN-17]
Given uma candidatura expirada
When o dono tenta aceitá-la ou recusá-la
Then a ação é rejeitada com o erro "candidatura não está pendente"
```

### US-05 — Membro sai do grupo
```gherkin
CA-05.1 — Saída antes do início  [US-05, RN-14]
Given um membro aceito num lobby aberto
When ele sai do grupo
Then a candidatura dele passa para saiu
  And a vaga da função dele volta a ficar livre

CA-05.2 — Saída depois do início  [US-05, RN-14]
Given um membro aceito num lobby iniciado
When ele tenta sair do grupo
Then a ação é rejeitada com o erro "lobby não está aberto"

CA-05.3 — Saída cancela pedido de troca  [US-05, RN-24]
Given um membro com pedido de troca pendente
When ele sai do grupo
Then o pedido de troca passa para cancelado
```

### US-06 — Dono remove membro
```gherkin
CA-06.1 — Remoção com justificativa  [US-06, RN-15, RN-09]
Given um membro aceito num lobby aberto
When o dono o remove com a justificativa "mudamos o horário da run"
Then a candidatura do membro passa para removida
  And a vaga da função dele volta a ficar livre
  And a justificativa fica registrada

CA-06.2 — Remoção sem justificativa  [US-06, RN-09]
Given um membro aceito num lobby aberto
When o dono tenta removê-lo sem justificativa
Then a remoção é rejeitada com o erro "justificativa obrigatória"
  And o membro continua aceito

CA-06.3 — Remoção por quem não é dono  [US-06, RN-15]
Given um membro aceito
  And o usuário logado não é o dono do lobby
When ele tenta remover o membro
Then a ação é rejeitada com o erro "apenas o dono pode decidir"

CA-06.4 — Remoção depois do início  [US-06, RN-15]
Given um membro aceito num lobby iniciado
When o dono tenta removê-lo
Then a ação é rejeitada com o erro "lobby não está aberto"

CA-06.5 — Removido sem bloqueio se candidata de novo  [US-06, RN-07, RN-15]
Given um jogador removido do lobby sem bloqueio
  And o lobby continua aberto com vaga na função
When ele se candidata de novo ao mesmo lobby
Then uma nova candidatura pendente é criada

CA-06.6 — Removido com bloqueio  [US-06, RN-07, RN-15]
Given um jogador removido do lobby com bloqueio
When ele tenta se candidatar de novo ao mesmo lobby com qualquer personagem
Then a candidatura é rejeitada com o erro "usuário bloqueado neste lobby"

CA-06.7 — Bloqueio vale só para o lobby  [US-06, RN-15]
Given um jogador removido com bloqueio do lobby A
  And o lobby B, do mesmo dono, está aberto com vaga na função
When ele se candidata ao lobby B
Then uma nova candidatura pendente é criada
```

### US-07 — Dono troca o próprio personagem
```gherkin
CA-07.1 — Troca para função com vaga  [US-07, RN-19]
Given o dono ocupa uma vaga de Dano com o personagem A
  And a função Tank tem 1 vaga livre
  And o personagem B do dono é Tank e não tem conflito de horário
When o dono troca para o personagem B
Then o dono passa a ocupar a vaga de Tank com B
  And a vaga de Dano que ele deixou volta a ficar livre

CA-07.2 — Troca dentro da mesma função lotada  [US-07, RN-19]
Given o dono ocupa uma vaga de Suporte com o personagem A
  And todas as vagas de Suporte estão ocupadas
  And o personagem B do dono também é Suporte
When o dono troca para o personagem B
Then o dono passa a ocupar a mesma vaga de Suporte com B

CA-07.3 — Troca para função sem vaga  [US-07, RN-19]
Given todas as vagas de Tank estão ocupadas
When o dono tenta trocar para um personagem Tank
Then a troca é rejeitada com o erro "função sem vaga"
  And o dono continua com o personagem atual

CA-07.4 — Troca com conflito de horário  [US-07, RN-19, RN-11]
Given o personagem B do dono ocupa vaga noutro lobby com horário sobreposto
When o dono tenta trocar para o personagem B
Then a troca é rejeitada com o erro "personagem em outro lobby no mesmo horário"

CA-07.5 — Troca por quem não é dono  [US-07, RN-19]
Given um membro aceito
When ele tenta trocar de personagem diretamente, sem pedido
Then a ação é rejeitada com o erro "apenas o dono pode decidir"
```

### US-08 — Pedido de troca do membro
```gherkin
CA-08.1 — Abrir pedido de troca  [US-08, RN-20, RN-21]
Given um membro aceito como Dano num lobby aberto
When ele pede a troca para o personagem Tank dele, com o motivo "ninguém apareceu de tank"
Then é criado um pedido de troca pendente
  And o membro continua ocupando a vaga de Dano com o personagem atual

CA-08.2 — Pedido sem motivo  [US-08, RN-20, RN-09]
Given um membro aceito
When ele pede a troca sem informar motivo
Then o pedido é rejeitado com o erro "justificativa obrigatória"

CA-08.3 — Segundo pedido pendente  [US-08, RN-20]
Given um membro com um pedido de troca pendente
When ele tenta abrir outro pedido no mesmo lobby
Then o pedido é rejeitado com o erro "já existe pedido de troca pendente"

CA-08.4 — Personagem de outro usuário no pedido  [US-08, RN-20]
Given um membro aceito
When ele pede a troca para um personagem que não é dele
Then o pedido é rejeitado com o erro "personagem não pertence ao usuário"

CA-08.5 — Dono aceita o pedido  [US-08, RN-22, RN-23]
Given um pedido de troca pendente de Dano para Tank
  And a função Tank tem 1 vaga livre
  And o personagem novo não tem conflito de horário
When o dono aceita o pedido
Then o membro passa a ocupar a vaga de Tank com o personagem novo
  And a vaga de Dano volta a ficar livre
  And o pedido passa para aceito

CA-08.6 — Dono aceita com função lotada  [US-08, RN-23]
Given um pedido de troca pendente para Tank
  And todas as vagas de Tank estão ocupadas
When o dono aceita o pedido
Then o aceite é rejeitado com o erro "vaga preenchida"
  And o pedido continua pendente
  And o membro continua com o personagem atual

CA-08.7 — Dono aceita com conflito de horário  [US-08, RN-23, RN-11]
Given um pedido de troca pendente cujo personagem novo ocupa vaga noutro lobby sobreposto
When o dono aceita o pedido
Then o aceite é rejeitado com o erro "personagem em outro lobby no mesmo horário"
  And o pedido continua pendente

CA-08.8 — Dono recusa o pedido  [US-08, RN-22, RN-09]
Given um pedido de troca pendente
When o dono recusa com a justificativa "já achamos um tank"
Then o pedido passa para recusado
  And o membro continua com o personagem atual

CA-08.9 — Recusa do pedido sem justificativa  [US-08, RN-22, RN-09]
Given um pedido de troca pendente
When o dono recusa sem justificativa
Then a recusa é rejeitada com o erro "justificativa obrigatória"
  And o pedido continua pendente

CA-08.10 — Remoção cancela pedido  [US-08, RN-24]
Given um membro com pedido de troca pendente
When o dono remove o membro com justificativa
Then o pedido de troca passa para cancelado

CA-08.11 — Membro retira o pedido  [US-08, RN-27]
Given um membro com pedido de troca pendente num lobby aberto
When ele retira o pedido
Then o pedido passa para retirado
  And o membro continua com o personagem atual
  And ele pode abrir um novo pedido de troca

CA-08.12 — Retirar pedido de outra pessoa  [US-08, RN-27]
Given um pedido de troca pendente de outro membro
When o usuário logado tenta retirá-lo
Then a ação é rejeitada com o erro "pedido não pertence ao usuário"
  And o pedido continua pendente
```

### Regras de personagem (transversais)
```gherkin
CA-09.1 — Função travada com candidatura ativa  [RN-25]
Given um personagem com candidatura pendente ou vaga ocupada num lobby aberto
When o dono do personagem tenta alterar a função dele
Then a alteração é rejeitada com o erro "personagem em uso em lobby"

CA-09.2 — Função liberada sem vínculos  [RN-25]
Given um personagem sem candidaturas pendentes, sem vaga em lobby aberto e sem pedido de troca pendente
When o dono do personagem altera a função
Then a alteração é aceita

CA-09.3 — Exclusão do personagem  [RN-26]
Given um personagem com uma candidatura pendente num lobby e uma aceita noutro
When o personagem é excluído
Then a candidatura pendente passa para cancelada
  And a candidatura aceita passa para cancelada
  And a vaga que ele ocupava volta a ficar livre
```

## 6. Casos de borda
- Dois aceites para a última vaga ao mesmo tempo → só um conclui [RN-12 / CA-02.9]
- Função lotada com candidaturas pendentes → elas continuam pendentes e podem ser
  aceitas se a vaga abrir de novo (por saída, remoção ou troca) [RN-10 / CA-02.8]
- Lobbies encostados (um acaba às 22:00 e o outro começa às 22:00) → não conflitam
  [RN-11 / CA-02.11]
- Lobby cancelado não conta para conflito de horário [RN-11 / CA-02.12]
- Justificativa só com espaços → conta como vazia [RN-09 / CA-02.3]
- Troca do dono dentro da mesma função lotada → permitida, porque a vaga dele conta
  como livre [RN-19 / CA-07.2]
- Candidato que retira pode se candidatar de novo; candidato recusado não pode
  [RN-07 / CA-03.5, CA-01.9]
- Removido sem bloqueio pode se candidatar de novo; com bloqueio, não. O bloqueio
  não se estende a outros lobbies do mesmo dono [RN-07, RN-15 / CA-06.5 a CA-06.7]
- Início do lobby com pedido de troca pendente → o pedido expira e o membro continua
  com o personagem atual [RN-16 / CA-04.1]

## 7. Requisitos não funcionais
- **RNF-01** — Datas e horários ficam em UTC no banco e são convertidos para o fuso
  do usuário na exibição.
- **RNF-02** — As regras de permissão (RN-08, RN-13, RN-15, RN-19, RN-22) são
  verificadas no backend. Esconder botões na interface não basta.
- **RNF-03** — Aceite de candidatura e aceite de troca são atômicos: ou a transição
  inteira acontece, ou nada muda (RN-12, RN-23).
- **RNF-04** — Toda regra de domínio tem teste automatizado, e cada teste cita o ID
  do critério de aceite.

## 8. Fora de escopo
- Escolha do personagem na criação do lobby — pertence à spec de Lobby (P-02).
- Criar, editar e cancelar lobby, e definir as vagas por função — spec de Lobby.
- Notificações fora do app (DM no Discord, e-mail) — o status fica visível dentro do
  app. Pode entrar numa feature de notificações.
- Requisitos de nível ou classe definidos pelo lobby — pode entrar depois, na spec
  de Lobby.
- Lista de espera quando a função está lotada.
- Reputação de jogador.
- Candidaturas de guilda — item futuro no CLAUDE.md.
- Aceite automático de candidaturas.
- Chat entre candidato e dono além da mensagem opcional (RN-06).
- Bloqueio de um usuário em todos os lobbies de um dono (lista de bloqueio global) —
  o bloqueio desta entrega vale só para um lobby. Pode entrar junto com reputação.

## 9. Perguntas em aberto
- Nenhuma. As premissas assumidas foram resolvidas na aprovação (ver seção 10).
- Fica para a spec de Lobby: o que acontece se o personagem que o dono usa num lobby
  aberto for excluído.

## 10. Decisões tomadas na entrevista
- Mais de um personagem do mesmo usuário no mesmo lobby → não, uma candidatura ativa
  por usuário.
- Função da candidatura → herdada do personagem.
- Candidatura para função lotada → bloqueada.
- Conflito de horário → pendência permitida, aceite bloqueado.
- Dono se candidata ao próprio lobby → não. Ele escolhe o personagem ao criar o lobby
  (spec de Lobby) e pode trocar o próprio personagem direto.
- Troca de personagem do membro → por pedido com motivo. O dono aceita ou recusa
  com justificativa, e o membro não perde a vaga enquanto espera.
- Retirar candidatura pendente → a qualquer momento.
- Membro sai do grupo → sim, até o início, liberando a vaga.
- Dono remove membro → sim, com justificativa, até o início.
- Pendências no início ou no cancelamento → expiram.
- Recusado se candidata de novo ao mesmo lobby → não.
- Aceites concorrentes → o primeiro vence, o segundo falha e as outras pendências
  continuam pendentes.
- Personagem com vínculos → função travada. Exclusão cancela os vínculos.
- Justificativa → de 10 a 250 caracteres. Mensagem do candidato → opcional, até 250.
- Fora de escopo → notificações externas, requisitos de nível/classe, lista de
  espera, reputação.
- Membro retira pedido de troca pendente → sim, pode desistir (RN-27).
- Visibilidade das pendentes → terceiros veem só a quantidade; os detalhes ficam com
  o dono e o candidato (RN-28).
- Visibilidade da justificativa → só o dono e o candidato afetado (RN-29).
- Removido se candidata de novo → sim, a menos que o dono bloqueie na remoção, por
  exemplo por comportamento desrespeitoso (RN-07, RN-15).
- Personagem do dono excluído → fica para a spec de Lobby.
