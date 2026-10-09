# Spec — Lobbies

- Feature: `lobbies` · Nível: G · Status: Aprovada
- Design: Claude Design, `Lobby.dc.html`: criação como página `/lobbies/novo` com prévia do
  card (1b), detalhe (1c) e cancelamento (1d)
- Última revisão: 2026-10-09 — a edição passa a trocar a instância (RN-17, CA-04.5,
  CA-04.6). Antes, 2026-10-06 — teste do usuário no localhost: card inteiro clicável; criar
  lobby a partir do dia escolhido na Home (RN-24)
  (RN-23), retrato do dono nas vagas (RN-15) e detalhe com o fundo da Home

## 1. Contexto
Com login e personagens prontos, o Usuário já pode montar grupos. Esta feature cria o
lobby (instância, horário, vagas por função, nível mínimo e o personagem do dono), a
página de detalhe, a edição e o cancelamento, e troca os dados fictícios da Home pelos
lobbies reais.

Ela cumpre as premissas da spec `candidatura-lobby` (aprovada): P-01 (dono, horário, vagas
por função, até 12), P-02 (o personagem do dono ocupa uma vaga desde a criação) e P-03
(estados aberto, iniciado e cancelado). A candidatura em si fica para a feature seguinte.

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Visitante | Quem não entrou | Ver a Home e o detalhe dos lobbies |
| Usuário | Quem entrou com o Discord | Tudo do visitante e criar lobby |
| Dono | Usuário que criou o lobby | Editar e cancelar o próprio lobby |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como usuário, quero criar um lobby para uma instância, num horário, com as vagas por função, para montar meu grupo. |
| US-02 | P1 | Como visitante, quero ver na Home os lobbies reais dos próximos dias, para achar um grupo. |
| US-03 | P1 | Como visitante, quero abrir o detalhe de um lobby, para ver a composição e as condições. |
| US-04 | P2 | Como dono, quero editar o meu lobby enquanto ele está aberto, para ajustar horário, vagas e nível. |
| US-05 | P2 | Como dono, quero cancelar o meu lobby com um motivo, para avisar que o grupo não vai mais acontecer. |
| US-06 | P1 | Como desenvolvedor, quero que a API valide os lobbies e sirva o catálogo de instâncias, para as regras valerem em qualquer cliente. |

## 4. Regras

### Catálogo de instâncias
- **RN-01** — O catálogo segue a página
  [Instâncias do bROWiki](https://browiki.org/wiki/Inst%C3%A2ncias) (consultada em
  2026-10-06) e tem só as instâncias que aceitam grupo: as marcadas como "Solo" ficam de
  fora. Cada instância tem nome, nível de entrada e tipo de retorno (diário, 3 dias, em
  horas ou semanal). O catálogo vive na API; `GET /instances` não exige sessão.
- **RN-02** — Na escolha da instância, as de nível de entrada 130 ou mais vêm primeiro e
  as de nível menor vêm depois de um separador. Dentro de cada parte, a ordem é por nível
  de entrada decrescente e, no empate, por nome.
- **RN-03** — Ainda não há arte por instância. Cards e detalhe usam a capa e o ícone
  padrão do RO Lobby, sem nada da Gravity.

### Criação
- **RN-04** — Criar lobby exige login. Quem cria é o dono.
- **RN-05** — O início é informado como data e hora no fuso `America/Sao_Paulo` (RN-09 da
  `home-local`) e guardado em UTC. Ele precisa estar no futuro e cair em um dos 14 dias do
  seletor da Home (hoje e os 13 seguintes).
- **RN-06** — Vagas por função: Tank, Suporte e Dano, cada uma de 0 a 12, com total de 1 a
  12 (P-01). O padrão é 1 Tank, 2 Suportes e 3 Danos.
- **RN-07** — Nível mínimo: inteiro do nível de entrada da instância até 275. O padrão é o
  nível de entrada.
- **RN-08** — O dono escolhe um dos próprios personagens (o principal vem marcado). O
  personagem precisa ter o nível mínimo, e a função dele precisa ter pelo menos 1 vaga. Ele
  ocupa essa vaga desde a criação (P-02).
- **RN-09** — Observação: opcional, até 250 caracteres depois de tirar os espaços das pontas.
- **RN-10** — O lobby não tem duração. Para conflito de horário, cada lobby ocupa a janela
  [início, início + 2 h). Um personagem não ocupa vaga em dois lobbies não cancelados cujas
  janelas se sobrepõem. A mesma janela vale para a RN-11 da `candidatura-lobby`.
- **RN-11** — Cada Usuário tem no máximo 5 lobbies abertos como dono.
- **RN-12** — Usuário sem personagem não cria lobby: a tela de criação pede para cadastrar
  um, com link para `/perfil`.

### Estado e listagem
- **RN-13** — O lobby está *aberto* antes do início e se não foi cancelado, *iniciado*
  quando o início já passou, e *cancelado* quando o dono cancelou (P-03). Iniciado vem do
  horário; cancelado é final.
- **RN-14** — A Home e a lista da API mostram só lobbies abertos: o grupo sai da listagem
  no horário de início ou quando é cancelado.

### Detalhe
- **RN-15** — `/lobbies/{id}` é público. Mostra instância, data e hora, estado, nível
  mínimo, observação e as vagas por função com os ocupantes (por enquanto, só o personagem
  do dono: retrato, nick, classe, nível e o selo "Anfitrião"). Lobby inexistente responde
  404.
- **RN-16** — No detalhe, o dono vê "Editar" e "Cancelar lobby" enquanto o lobby está
  aberto. "Candidatar" continua "Disponível em breve" para os outros.

### Edição e cancelamento
- **RN-17** — Só o dono edita, e só com o lobby aberto. Ele muda instância, início, vagas,
  nível mínimo e observação, com as mesmas regras da criação (RN-05 a RN-10). O personagem
  do dono não muda por aqui (RN-19 da `candidatura-lobby`). Ao trocar a instância, o
  formulário sugere o nível de entrada da nova como nível mínimo, e o dono pode ajustar.
  Membros aceitos e candidaturas pendentes continuam no lobby, como na mudança do nível
  mínimo.
- **RN-18** — As vagas de uma função não ficam abaixo dos ocupantes dela, e o nível mínimo
  não fica acima do nível do personagem do dono.
- **RN-19** — Só o dono cancela, e só com o lobby aberto. Cancelar pede confirmação e um
  motivo de 10 a 250 caracteres depois de tirar os espaços das pontas (RN-09 da
  `candidatura-lobby`). O motivo aparece no detalhe do lobby cancelado.
- **RN-20** — Lobby de outro Usuário: editar e cancelar respondem 404 na API, sem revelar
  nada além do que o detalhe público já mostra.

### Personagem do dono
- **RN-21** — O personagem que é dono de um lobby aberto não pode ser excluído nem ter a
  função alterada (RN-25 da `candidatura-lobby`). A tela de perfil explica: "Esse
  personagem está num lobby aberto. Cancele o lobby antes."

### Home
- **RN-22** — A Home passa a buscar os lobbies na API (troca o módulo fictício, RN-08 da
  `home-local`). O card mostra horário, instância, vagas ocupadas/total, anfitrião (nick do
  personagem do dono), classe, nível mínimo e composição por função. Filtros, contagens,
  destaque e estados vazios seguem a `home-local`.
- **RN-23** — "Criar lobby" passa a funcionar e leva à página `/lobbies/novo`, que mostra
  ao lado uma prévia do card como vai aparecer na Home. Visitante vai ao login e volta para
  `/lobbies/novo`. "Ver grupo" abre o detalhe, e clicar em qualquer ponto do card também.
- **RN-24** — "Criar lobby" na Home (cabeçalho e aviso de lista vazia) abre a criação com o
  dia escolhido no seletor já preenchido (a Home abre em hoje). Fora da Home, ou com um dia
  fora dos 14 do seletor, o padrão é amanhã. A hora
  padrão é 20:00; se o dia for hoje e 20:00 já passou, é a próxima hora cheia. Se não sobra
  hora cheia hoje (depois das 23:00), o padrão vira amanhã às 20:00. O visitante passa pelo
  login e volta com o dia preenchido.

## 5. Critérios de aceite

### US-01 — Criar lobby
```gherkin
CA-01.1 — Criação com os padrões  [US-01, RN-04, RN-06, RN-07, RN-08]
Given um Usuário com o personagem principal "Lirien", Arcebispo, nível 178, Suporte
When ele cria um lobby de "Templo do Demônio Rei" amanhã às 20:00 sem mudar vagas e nível
Then o lobby fica aberto com 1 Tank, 2 Suportes e 3 Danos e nível mínimo 160
  And "Lirien" ocupa 1 vaga de Suporte

CA-01.2 — Horário no passado ou fora dos 14 dias  [US-01, RN-05]
Given o formulário de criação
When o Usuário escolhe um horário que já passou, ou um dia depois dos 14 do seletor
Then a criação é recusada com o erro no campo de data e hora

CA-01.3 — Vagas fora da faixa  [US-01, RN-06]
Given o formulário de criação
When o total de vagas é 0 ou 13
Then a criação é recusada com o erro nas vagas

CA-01.4 — Nível mínimo abaixo da instância  [US-01, RN-07]
Given a instância "Torre da Constelação", de nível 240
When o Usuário informa nível mínimo 200
Then a criação é recusada com o erro no nível mínimo

CA-01.5 — Personagem abaixo do nível mínimo  [US-01, RN-08]
Given o personagem "Brasa", nível 172
When o Usuário cria um lobby com nível mínimo 180 usando "Brasa"
Then a criação é recusada com o erro no personagem

CA-01.6 — Função do dono sem vaga  [US-01, RN-08]
Given o personagem "Brasa", Tank
When o Usuário cria um lobby com 0 vagas de Tank usando "Brasa"
Then a criação é recusada com o erro nas vagas

CA-01.7 — Personagem de outro Usuário  [US-01, RN-08]
Given um personagem de outro Usuário
When alguém tenta criar um lobby com ele pela API
Then a criação é recusada com o erro no personagem

CA-01.8 — Conflito de horário  [US-01, RN-10]
Given "Lirien" é dono de um lobby aberto hoje às 20:00
When o Usuário cria outro lobby às 21:30 com "Lirien"
Then a criação é recusada com o erro "Esse personagem já está num grupo nesse horário"
  And às 22:00 a criação é aceita

CA-01.9 — Limite de 5 lobbies abertos  [US-01, RN-11]
Given um Usuário dono de 5 lobbies abertos
When ele tenta criar mais um
Then a criação é recusada com a mensagem "Você já tem 5 lobbies abertos"

CA-01.10 — Sem personagem  [US-01, RN-12]
Given um Usuário sem personagens
When ele abre a criação de lobby
Then vê "Cadastre um personagem para criar lobbies" com o link para /perfil

CA-01.11 — Observação longa demais  [US-01, RN-09]
Given o formulário de criação
When a observação tem 251 caracteres
Then a criação é recusada com o erro na observação

CA-01.12 — Instância fora do catálogo  [US-01, RN-01]
Given uma tentativa de criação com uma instância que não está no catálogo
When a API recebe o pedido
Then responde 422 com o erro no campo instância
```

### US-02 — Lobbies na Home
```gherkin
CA-02.1 — Lobby criado aparece na Home  [US-02, RN-22]
Given um lobby aberto hoje às 20:00 de "Templo do Demônio Rei", com "Lirien" de dono
When um visitante abre a Home
Then vê o card às 20:00 com a instância, "Lirien", Arcebispo, nível mínimo e a composição

CA-02.2 — Iniciado e cancelado saem da listagem  [US-02, RN-13, RN-14]
Given um lobby que começou há 1 minuto e um lobby cancelado
When um visitante abre a Home
Then nenhum dos dois aparece

CA-02.3 — Criar lobby pela Home  [US-02, RN-23]
Given um visitante
When ele usa "Criar lobby"
Then vai ao login e volta para a criação de lobby

CA-02.4 — Ver grupo  [US-02, RN-23]
Given um card na Home
When o visitante usa "Ver grupo"
Then abre o detalhe daquele lobby

CA-02.5 — Criar lobby no dia escolhido  [US-02, RN-24]
Given o Usuário escolheu sexta, 9 out no seletor de dias da Home
When ele usa "Criar lobby"
Then a criação abre com o dia sexta, 9 out e a hora 20:00 preenchidos

CA-02.6 — Hora padrão de hoje  [US-02, RN-24]
Given hoje são 20:40 em Brasília e o Usuário escolheu hoje na Home
When ele usa "Criar lobby"
Then a criação abre com hoje e 21:00 preenchidos
  And às 23:10 abre com amanhã e 20:00
```

### US-03 — Detalhe
```gherkin
CA-03.1 — Detalhe público  [US-03, RN-15]
Given um lobby aberto
When um visitante abre /lobbies/{id}
Then vê instância, data e hora, nível mínimo, observação e as vagas por função com o personagem do dono na vaga dele

CA-03.2 — Lobby inexistente  [US-03, RN-15]
Given um id que não existe
When alguém abre /lobbies/{id}
Then vê a página de não encontrado

CA-03.3 — Ações do dono  [US-03, RN-16]
Given um lobby aberto
When o dono abre o detalhe
Then vê "Editar" e "Cancelar lobby"
  And outro Usuário vê "Candidatar" desabilitado com "Disponível em breve"
```

### US-04 — Editar
```gherkin
CA-04.1 — Editar horário e vagas  [US-04, RN-17]
Given um lobby aberto do Usuário
When ele muda o horário para 21:00 e as vagas de Dano para 5
Then o detalhe e a Home mostram 21:00 e 5 vagas de Dano

CA-04.2 — Vagas abaixo dos ocupantes  [US-04, RN-18]
Given um lobby em que o personagem do dono ocupa a vaga de Suporte
When o dono tenta deixar 0 vagas de Suporte
Then a edição é recusada com o erro nas vagas

CA-04.3 — Nível mínimo acima do dono  [US-04, RN-18]
Given um lobby com o personagem do dono no nível 178
When o dono tenta mudar o nível mínimo para 180
Then a edição é recusada com o erro no nível mínimo

CA-04.4 — Lobby de outro Usuário ou iniciado  [US-04, RN-17, RN-20]
Given um lobby de outro Usuário, ou um lobby já iniciado
When alguém tenta editá-lo pela API
Then a API responde 404 para o lobby de outro e 409 para o iniciado
  And o lobby continua igual

CA-04.5 — Trocar a instância  [US-04, RN-17, RN-07]
Given um lobby aberto do Usuário em "Templo do Demônio Rei", com um membro aceito e uma
  candidatura pendente
When o dono troca a instância para "Sonho Sombrio"
Then o formulário sugere 120, o nível de entrada da nova instância, como nível mínimo
  And depois de salvar, o detalhe e a Home mostram "Sonho Sombrio"
  And o membro aceito e a candidatura pendente continuam no lobby

CA-04.6 — Nova instância acima do personagem do dono  [US-04, RN-07, RN-18]
Given um lobby com o personagem do dono no nível 178
When o dono troca a instância para uma de nível de entrada 240
Then a edição é recusada com o erro no nível mínimo
  And o lobby continua com a instância antiga
```

### US-05 — Cancelar
```gherkin
CA-05.1 — Cancelar com motivo  [US-05, RN-19]
Given um lobby aberto do Usuário
When ele cancela com o motivo "Metade do grupo não pode hoje"
Then o lobby fica cancelado, sai da Home, e o detalhe mostra "Cancelado" com o motivo

CA-05.2 — Motivo curto  [US-05, RN-19]
Given um lobby aberto do Usuário
When ele tenta cancelar com o motivo "não dá"
Then o cancelamento é recusado com o erro no motivo

CA-05.3 — Cancelar lobby de outro Usuário  [US-05, RN-20]
Given um lobby de outro Usuário
When alguém tenta cancelá-lo pela API
Then a API responde 404
```

### US-06 — API e catálogo
```gherkin
CA-06.1 — Catálogo de instâncias  [US-06, RN-01, RN-02]
Given a API no ar
When se pede GET /instances sem sessão
Then a resposta tem as instâncias de grupo do bROWiki, sem as "Solo"
  And cada uma tem nome, nível de entrada e tipo de retorno

CA-06.2 — Separador na escolha  [US-06, RN-02]
Given o formulário de criação
When o Usuário abre a lista de instâncias
Then as de nível 130 ou mais vêm primeiro, em ordem de nível decrescente
  And as de nível menor vêm depois de um separador

CA-06.3 — Rotas do dono exigem sessão  [US-06, RN-04]
Given nenhuma sessão
When se pede para criar, editar ou cancelar um lobby
Then a API responde 401

CA-06.4 — Personagem do dono travado  [US-06, RN-21]
Given "Lirien" é dono de um lobby aberto
When o Usuário tenta excluir "Lirien" ou mudar a função dela
Then a ação é recusada com "Esse personagem está num lobby aberto. Cancele o lobby antes."

CA-06.5 — Limite e conflito sob concorrência  [US-06, RN-10, RN-11]
Given dois pedidos simultâneos de criação que, juntos, passariam do limite ou criariam conflito de horário
When os dois chegam à API
Then só um é aceito
```

## 6. Casos de borda
- Lobby criado para daqui a poucos minutos: aceito; some da Home no horário de início
  [RN-05, RN-14]
- Virada do dia em São Paulo: um lobby às 23:30 de hoje está no dia de hoje, mesmo sendo
  amanhã em UTC [RN-05]
- Instância que sair do catálogo no futuro: os lobbies abertos continuam aparecendo, com o
  nome guardado [RN-01]
- Personagem do dono que sobe de nível: nada muda no lobby [RN-08]
- Personagem do dono que cai abaixo do nível mínimo editando o perfil: o lobby continua;
  só a edição do nível mínimo passa a respeitar o nível novo [RN-18]
- Lobby que começa enquanto o dono está editando: a edição é recusada com 409 [RN-17]
- Membro aceito com nível abaixo do de entrada da nova instância: continua no grupo; o
  nível mínimo vale para candidaturas novas [RN-17, RN-18]

## 7. Requisitos não funcionais
- **RNF-01** — Acessibilidade: formulário, detalhe e diálogo de cancelamento funcionam
  pelo teclado, com foco visível e rótulos.
- **RNF-02** — Nenhum recurso carregado de fora do servidor do web.
- **RNF-03** — Todo teste cita o ID do critério de aceite.
- **RNF-04** — As regras valem na API; o web repete as simples só para avisar antes.
- **RNF-05** — A Home continua renderizada no servidor com os lobbies de hoje (RN-20 da
  `home-local`).

## 8. Fora de escopo
- Candidatura, aceite, recusa, saída e remoção de membros (spec `candidatura-lobby`).
- Arte (capa e ícone) por instância.
- Duração do lobby, recorrência e lobbies de guilda.
- Notificações (Discord ou e-mail).
- Trocar o personagem do dono depois de criado (RN-19 da `candidatura-lobby`).

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Escopo → criar, detalhe, editar, cancelar e Home com dados reais; candidatura depois.
- Tela de criação → página `/lobbies/novo` com prévia do card (variação 1b).
- Instâncias → catálogo do bROWiki na API, só as de grupo; as de nível 130+ primeiro,
  separador, depois as de nível menor; sem arte por instância por enquanto.
- Duração → não existe; o grupo sai da listagem no início. Para conflito, janela fixa de
  2 h (RN-10).
- Campos → instância, início, vagas por função (padrão 1/2/3), nível mínimo (padrão o da
  instância), personagem do dono (padrão o principal), observação.
- Criação → início no futuro e em até 14 dias; personagem com o nível; sem conflito;
  até 5 lobbies abertos.
- Editar e cancelar → só o dono, só aberto; personagem fixo; vagas não abaixo
  dos ocupantes; cancelar com confirmação e motivo de 10 a 250.
- Trocar a instância na edição (revisão de 2026-10-09) → permitido; sugere o nível de
  entrada da nova; membros e pendentes continuam.
