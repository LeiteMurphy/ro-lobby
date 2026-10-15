# Spec — Sala do grupo no Discord

- Feature: `discord-bot` · Nível: G · Status: Aprovada (2026-10-14)
- Notion: https://app.notion.com/p/3fad4a3a5eff81468b99db1fa0b99fc5
- Última revisão: 2026-10-14 — servidor do jogo na mensagem de abertura (RN-08, CA-02.2);
  antes, link do Discord do dono (US-07), "Avisar no Discord" com
  intervalo (US-08) e, pelo design, uma categoria por sala (RN-05, RN-22, D-04)

## 1. Contexto
Depois de montado, o grupo ainda precisa se achar no Discord para combinar a run e entrar
em voz, e hoje cada um troca nick na mão. O login já é pelo Discord, então o RO Lobby sabe
o ID de cada jogador. Esta feature cria um servidor oficial do RO Lobby onde um bot abre,
para cada lobby, uma sala privada (texto e voz) só com os membros do grupo. O bot menciona
todos, acompanha quem entra e sai, lembra do horário e apaga a sala depois. O jogador entra
no servidor sozinho, por opt-in, sem precisar de convite.

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Jogador | Usuário logado | Conectar a conta ao servidor oficial |
| Dono | Usuário que criou o lobby | Abrir a sala do próprio lobby, informar o link do próprio Discord e avisar o grupo |
| Membro | Dono ou Usuário com candidatura aceita no lobby | Ver o link da sala e o link do Discord do dono; acessar a sala se estiver conectado |
| Bot | Usuário bot do aplicativo do RO Lobby no Discord | Criar, ajustar e apagar salas e postar mensagens no servidor oficial |
| Operador | Quem roda a pilha (hoje, o dono do projeto) | Criar o servidor oficial e configurar o bot |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como jogador, quero conectar minha conta ao servidor do RO Lobby, para entrar nas salas dos meus grupos sem precisar de convite. |
| US-02 | P1 | Como dono, quero abrir uma sala no Discord para o meu lobby, para o grupo conversar e entrar em voz. |
| US-03 | P1 | Como membro, quero ver no lobby o link da sala e quem já está nela, para saber onde o grupo está. |
| US-04 | P1 | Como membro, quero que a sala acompanhe quem entra e quem sai do grupo, para ela ter sempre as pessoas certas. |
| US-05 | P1 | Como jogador, quero que a sala some sozinha depois do grupo, e que eu seja avisado quando o lobby for cancelado, para o servidor não acumular salas mortas. |
| US-06 | P2 | Como membro, quero ser lembrado na sala 15 minutos antes do início e avisado se o horário mudar, para não perder o grupo. |
| US-07 | P2 | Como dono, quero colocar no lobby o link do meu próprio Discord, visível só para os membros, para reunir o grupo no servidor que eu já uso. |
| US-08 | P2 | Como dono, quero um botão "Avisar no Discord" que faça o bot marcar o grupo na sala, para chamar todo mundo na hora sem marcar um por um. |

## 4. Regras de domínio

### Conexão ao servidor
- **RN-01** — O login continua pedindo só o escopo `identify` (RN-01 da `login-discord`).
  A conexão ao servidor é um passo separado e opcional, feito pelo perfil, com os escopos
  `identify` e `guilds.join`.
- **RN-02** — Na conexão, a API usa o token do Discord só para colocar o Usuário no
  servidor oficial e o descarta em seguida. O RO Lobby não guarda nem registra em log
  tokens do Discord (RN-04 e RN-06 da `login-discord` continuam valendo). Do Usuário, a
  feature guarda só a data da conexão.
- **RN-03** — A conexão só vale se a conta do Discord autorizada for a mesma do Usuário
  logado (mesmo ID do Discord). Com outra conta, a conexão é recusada e nada muda.
- **RN-04** — Conectar de novo é permitido e serve para voltar ao servidor (por exemplo,
  depois de sair dele no Discord). Quem já está no servidor continua nele, e a data da
  conexão é atualizada.

### Sala
- **RN-05** — Cada lobby tem no máximo uma sala. A sala é uma categoria com um canal de texto
  e um canal de voz no servidor oficial, com o título do lobby (instância ou título livre) e um código
  curto do lobby no nome, para dois lobbies iguais não se confundirem.
- **RN-06** — Só o dono abre a sala, e só com o lobby aberto (nem iniciado, nem
  cancelado) e sem sala aberta antes. Não há outro gatilho: lobby lotado não abre a sala
  sozinho.
- **RN-07** — A sala é privada. `@everyone` não vê os canais. Vê e entra quem é membro do
  lobby e está conectado (RN-01), além do bot.
- **RN-08** — Ao abrir, o bot posta no canal de texto uma mensagem com:
  - título, servidor do jogo (RN-08 da `servidor`) e horário do lobby (horário de Brasília,
    RN-09 da `home-local`);
  - a composição por função (Tank, Suporte, Dano), com nick e classe de cada membro;
  - a menção a cada membro conectado;
  - o nick de quem ainda não está na sala, com o aviso de que precisa se conectar pelo
    perfil;
  - o link do lobby.
- **RN-09** — Uma sala tem três estados: aberta, encerrando (lobby cancelado, aguardando
  a remoção) e apagada. Apagada é final: o lobby não abre outra sala.

### Sincronização com o grupo
- **RN-10** — Com a sala aberta, quem vira membro (candidatura aceita) ganha acesso, se
  estiver conectado, e é mencionado numa mensagem "<nick> entrou no grupo".
- **RN-11** — Com a sala aberta, quem deixa de ser membro (saiu ou foi removido) perde o
  acesso, e o bot posta "<nick> saiu do grupo". O motivo de uma remoção não aparece no
  Discord.
- **RN-12** — Trocar o personagem do mesmo Usuário não muda o acesso nem gera mensagem.
- **RN-13** — Quem se conecta (RN-01) ganha acesso às salas abertas de todos os lobbies em
  que já é membro, e é mencionado em cada uma.

### Lembrete, mudança e fim
- **RN-14** — 15 minutos antes do início, o bot posta na sala aberta um lembrete que
  menciona os membros conectados. A sala aberta a menos de 15 minutos do início não recebe
  lembrete, porque a mensagem de abertura já mencionou todos.
- **RN-15** — Se o dono muda o horário com a sala aberta, o bot posta o novo horário
  mencionando os membros conectados, e o lembrete segue o novo horário.
- **RN-16** — A sala aberta é apagada 6 horas depois do início do lobby.
- **RN-17** — Quando o lobby é cancelado com a sala aberta, o bot posta o cancelamento com
  o motivo (RN-19 da `lobbies`), mencionando os membros, e a sala passa a encerrando. Ela é
  apagada 1 hora depois do cancelamento.

### Detalhe do lobby
- **RN-18** — No detalhe do lobby, só os membros veem a seção "Sala no Discord":
  - sem sala, o dono vê "Abrir sala no Discord" (RN-06), e os outros membros veem que a
    sala ainda não foi aberta;
  - com a sala aberta, o link para o canal de texto e a lista dos membros com "na sala" ou
    "não conectado";
  - o membro não conectado vê o atalho para se conectar (RN-01);
  - com a sala encerrando ou apagada, a seção diz que a sala foi encerrada;
  - o link do Discord do dono, se existir (RN-23), mostrado como "Discord do anfitrião",
    junto com a sala ou no lugar dela.
- **RN-19** — Quem não é membro, logado ou não, não vê a seção, o link da sala nem o link
  do Discord do dono, nem na página nem na resposta da API.

### Link do Discord do dono
- **RN-23** — O dono pode informar, ao criar ou editar o lobby aberto, um link opcional do
  próprio Discord. O link vale só com `https://` e um destes formatos: `discord.gg/<código>`,
  `discord.com/invite/<código>` ou `discord.com/channels/<servidor>/<canal>`. Ele tem até 200
  caracteres. Fora disso, o lobby é recusado com "Use um link de convite do Discord". O dono
  pode trocar ou apagar o link enquanto o lobby está aberto.
- **RN-24** — O link do dono não depende do bot nem da conexão ao servidor. Ele aparece aos
  membros mesmo com a feature do bot desligada (RN-22).
- **RN-25** — O link do dono não aparece em nada que circula para fora do grupo: card da
  home, convite copiado e preview Open Graph (`compartilhar-lobby`), banco de talentos e
  mensagens do bot.

### Avisar no Discord
- **RN-26** — Com a sala aberta, o dono pode usar "Avisar no Discord". O bot posta no canal
  de texto "<nick do dono> está chamando o grupo", com a menção aos membros conectados, sem
  o próprio dono, e o link do lobby. O aviso tem texto fixo, sem mensagem livre. Sem sala
  aberta (antes de abrir, encerrando ou apagada), o botão não aparece e a API recusa.
- **RN-27** — O aviso tem um intervalo mínimo de 10 minutos por lobby, contado a partir do
  último aviso aceito. Durante o intervalo, o botão mostra quanto falta, e a API recusa com
  "Aguarde <n> min para avisar de novo" sem chamar o Discord. O lembrete automático
  (RN-14) e as mensagens de entrada e saída não contam para o intervalo.
- **RN-28** — Só o dono avisa. Um aviso que falha no Discord não é refeito depois (o valor
  dele está no momento) e não conta para o intervalo; o dono vê "Não foi possível avisar
  agora. Tente de novo em instantes".

### Falhas e configuração
- **RN-20** — Se o Discord falhar ao abrir a sala, a abertura é recusada com "Não foi
  possível abrir a sala agora. Tente de novo em instantes", e o lobby continua sem sala. Se
  o servidor estiver no limite de canais do Discord, a mensagem é "O servidor do RO Lobby
  está cheio agora. Tente mais tarde".
- **RN-21** — As ações do grupo (aceitar, sair, remover, editar, cancelar, conectar) não
  dependem do Discord. Elas valem mesmo com o Discord fora do ar. O efeito na sala (acesso,
  mensagem, lembrete, remoção) fica pendente e é refeito até dar certo ou até a sala ser
  apagada.
- **RN-22** — Sem a configuração do bot (token e servidor), a feature fica
  desligada: o perfil não oferece a conexão, e o detalhe do lobby não mostra a seção. O
  resto do RO Lobby funciona normalmente.

## 5. Critérios de aceite

### US-01 — Conectar a conta ao servidor
```gherkin
CA-01.1 — Conexão coloca o jogador no servidor  [US-01, RN-01, RN-02]
Given um Usuário logado que nunca se conectou
  And o bot configurado
When ele escolhe "Entrar no servidor do RO Lobby" no perfil e autoriza no Discord
Then o Discord recebe o pedido para incluir esse ID do Discord no servidor oficial
  And o perfil mostra "Conectado ao servidor do RO Lobby" com a data
  And nenhum token do Discord fica gravado no banco nem aparece no log

CA-01.2 — Login não pede guilds.join  [US-01, RN-01]
Given um visitante
When ele começa o login
Then a URL de autorização do Discord tem só o escopo identify

CA-01.3 — Conexão pede identify e guilds.join  [US-01, RN-01]
Given um Usuário logado
When ele começa a conexão ao servidor
Then a URL de autorização do Discord tem os escopos identify e guilds.join
  And tem um state novo, conferido no retorno

CA-01.4 — Conta do Discord diferente  [US-01, RN-03]
Given um Usuário logado com o ID do Discord A
When ele autoriza a conexão com a conta do Discord B
Then a conexão é recusada com "Use a mesma conta do Discord do seu login"
  And o Usuário continua sem conexão
  And o Discord não recebe o pedido de inclusão no servidor

CA-01.5 — Autorização cancelada  [US-01, RN-01]
Given um Usuário logado
When ele cancela a autorização no Discord
Then volta ao perfil com "Conexão cancelada"
  And o Usuário continua sem conexão

CA-01.6 — Reconectar  [US-01, RN-04]
Given um Usuário conectado em 10/10
When ele se conecta de novo em 14/10
Then o Discord recebe de novo o pedido de inclusão no servidor
  And a data da conexão passa a ser 14/10

CA-01.7 — Bot não configurado  [US-01, RN-22]
Given a API sem a configuração do bot
When um Usuário abre o perfil
Then o perfil não mostra "Entrar no servidor do RO Lobby"
  And o início da conexão responde que a feature está desligada
```

### US-02 — Abrir a sala
```gherkin
CA-02.1 — Dono abre a sala  [US-02, RN-05, RN-06, RN-07]
Given um lobby aberto sem sala, com o dono e dois membros aceitos, todos conectados
When o dono escolhe "Abrir sala no Discord"
Then o servidor oficial ganha um canal de texto e um de voz com o título e o código do lobby
  And os dois canais negam a visão a @everyone
  And os dois canais liberam a visão ao bot e aos três membros
  And a sala do lobby fica aberta

CA-02.2 — Mensagem de abertura  [US-02, RN-08]
Given um lobby aberto com o dono (Tank) e dois membros aceitos (Suporte e Dano)
  And o membro Dano não está conectado
When o dono abre a sala
Then o canal de texto recebe uma mensagem com título, servidor, horário de Brasília e composição por função com nick e classe
  And a mensagem menciona o dono e o membro Suporte pelo ID do Discord
  And a mensagem não menciona o membro Dano, mas cita o nick dele como fora da sala
  And a mensagem tem o link do lobby

CA-02.3 — Membro não conectado fica de fora  [US-02, RN-07]
Given um lobby aberto com um membro que não está conectado
When o dono abre a sala
Then os canais não liberam a visão a esse membro

CA-02.4 — Só o dono abre  [US-02, RN-06]
Given um lobby aberto sem sala
When um membro aceito que não é o dono tenta abrir a sala
Then a abertura é recusada
  And o lobby continua sem sala

CA-02.5 — Lobby iniciado ou cancelado  [US-02, RN-06]
Given um lobby iniciado ou cancelado, sem sala
When o dono tenta abrir a sala
Then a abertura é recusada
  And nenhum canal é criado

CA-02.6 — Uma sala por lobby  [US-02, RN-05, RN-06, RN-09]
Given um lobby com sala aberta, encerrando ou apagada
When o dono tenta abrir a sala de novo
Then a abertura é recusada
  And nenhum canal novo é criado

CA-02.7 — Discord indisponível na abertura  [US-02, RN-20]
Given um lobby aberto sem sala
  And o Discord respondendo com erro
When o dono abre a sala
Then a abertura é recusada com "Não foi possível abrir a sala agora. Tente de novo em instantes"
  And o lobby continua sem sala
  And nenhum canal criado nessa tentativa fica para trás no servidor

CA-02.8 — Servidor no limite de canais  [US-02, RN-20]
Given o Discord recusando a criação do canal por limite de canais
When o dono abre a sala
Then a abertura é recusada com "O servidor do RO Lobby está cheio agora. Tente mais tarde"
  And o lobby continua sem sala
```

### US-03 — Sala no detalhe do lobby
```gherkin
CA-03.1 — Membro vê o link e quem está na sala  [US-03, RN-18]
Given um lobby com sala aberta, com um membro conectado e outro não
When um membro abre o detalhe do lobby
Then a seção "Sala no Discord" mostra o link do canal de texto
  And lista o membro conectado como "na sala" e o outro como "não conectado"

CA-03.2 — Membro não conectado vê o atalho  [US-03, RN-18]
Given um lobby com sala aberta
  And um membro que não está conectado
When ele abre o detalhe do lobby
Then a seção mostra o atalho para se conectar ao servidor

CA-03.3 — Dono sem sala vê o botão  [US-03, RN-18]
Given um lobby aberto sem sala
When o dono abre o detalhe
Then a seção mostra "Abrir sala no Discord"
  And um membro aceito vê "A sala ainda não foi aberta" sem o botão

CA-03.4 — Quem não é membro não vê a sala  [US-03, RN-19]
Given um lobby com sala aberta
When um visitante, um Usuário sem candidatura ou um candidato pendente abre o detalhe
Then a página não mostra a seção "Sala no Discord" nem o link do canal
  And a resposta da API para ele não traz o link do canal

CA-03.5 — Sala encerrada  [US-03, RN-18]
Given um lobby com a sala encerrando ou apagada
When um membro abre o detalhe
Then a seção diz "A sala foi encerrada"
  And não mostra o link nem o botão de abrir
```

### US-04 — Sala acompanha o grupo
```gherkin
CA-04.1 — Aceito ganha acesso  [US-04, RN-10]
Given um lobby com sala aberta
  And um candidato pendente conectado
When o dono aceita a candidatura
Then os canais da sala liberam a visão a esse Usuário
  And o canal de texto recebe "<nick> entrou no grupo" com a menção a ele

CA-04.2 — Aceito não conectado  [US-04, RN-10]
Given um lobby com sala aberta
  And um candidato pendente não conectado
When o dono aceita a candidatura
Then os canais não liberam a visão a ele
  And o canal de texto recebe "<nick> entrou no grupo" sem menção

CA-04.3 — Quem sai perde o acesso  [US-04, RN-11]
Given um lobby com sala aberta e um membro aceito conectado
When o membro sai do grupo, ou o dono o remove com justificativa
Then os canais deixam de liberar a visão a ele
  And o canal de texto recebe "<nick> saiu do grupo"
  And a mensagem não traz a justificativa da remoção

CA-04.4 — Troca de personagem não mexe na sala  [US-04, RN-12]
Given um lobby com sala aberta e um membro aceito
When a troca de personagem desse membro é concluída
Then o acesso aos canais não muda
  And o canal de texto não recebe mensagem

CA-04.5 — Conexão tardia  [US-04, RN-13]
Given um Usuário não conectado, membro de dois lobbies com sala aberta e de um com sala apagada
When ele se conecta ao servidor
Then as duas salas abertas liberam a visão a ele e o mencionam
  And a sala apagada não é tocada

CA-04.6 — Discord fora do ar não trava o grupo  [US-04, RN-21]
Given um lobby com sala aberta
  And o Discord respondendo com erro
When o dono aceita um candidato conectado
Then a candidatura fica aceita normalmente
  And, quando o Discord volta, os canais liberam a visão a ele e o mencionam
```

### US-05 — Fim da sala
```gherkin
CA-05.1 — Sala apagada 6 horas depois do início  [US-05, RN-16, RN-09]
Given um lobby com sala aberta que começou às 20:00
When chegam 02:00 do dia seguinte
Then os dois canais são apagados no servidor
  And a sala fica apagada

CA-05.2 — Ainda não é hora de apagar  [US-05, RN-16]
Given um lobby com sala aberta que começou às 20:00
When são 01:59 do dia seguinte
Then os canais continuam no servidor

CA-05.3 — Cancelamento avisa e encerra  [US-05, RN-17]
Given um lobby com sala aberta e membros conectados
When o dono cancela o lobby com o motivo "Faltou o tank, fica para amanhã"
Then o canal de texto recebe o cancelamento com o motivo e a menção aos membros conectados
  And a sala fica encerrando
  And os canais são apagados 1 hora depois do cancelamento

CA-05.4 — Remoção pendente  [US-05, RN-21]
Given uma sala que já devia ter sido apagada
  And o Discord respondendo com erro
When o Discord volta
Then os canais são apagados
  And a sala fica apagada

CA-05.5 — Canal já apagado à mão  [US-05, RN-09]
Given uma sala aberta cujos canais alguém apagou direto no Discord
When chega a hora de apagar a sala
Then a sala fica apagada sem erro
```

### US-06 — Lembrete e mudança de horário
```gherkin
CA-06.1 — Lembrete 15 minutos antes  [US-06, RN-14]
Given um lobby com sala aberta que começa às 20:00
When são 19:45
Then o canal de texto recebe um lembrete que menciona os membros conectados
  And o lembrete é postado uma vez só

CA-06.2 — Sala aberta em cima da hora  [US-06, RN-14]
Given um lobby que começa às 20:00
When o dono abre a sala às 19:50
Then a sala não recebe lembrete

CA-06.3 — Lobby cancelado não lembra  [US-06, RN-14, RN-17]
Given um lobby que começa às 20:00, com a sala aberta, cancelado às 19:30
When são 19:45
Then a sala não recebe lembrete

CA-06.4 — Mudança de horário  [US-06, RN-15]
Given um lobby com sala aberta que começava às 20:00
When o dono muda o início para 21:00
Then o canal de texto recebe o novo horário com a menção aos membros conectados
  And o lembrete passa a ser às 20:45, e não às 19:45
  And a remoção passa a ser às 03:00
```

### US-07 — Link do Discord do dono
```gherkin
CA-07.1 — Dono informa o link  [US-07, RN-23]
Given um Usuário criando um lobby
When ele informa "https://discord.gg/abc123" como link do Discord
Then o lobby é criado com esse link

CA-07.2 — Link inválido  [US-07, RN-23]
Given um dono editando o lobby aberto
When ele informa "https://exemplo.com/discord" ou "http://discord.gg/abc123"
Then a edição é recusada com "Use um link de convite do Discord"
  And o link anterior continua valendo

CA-07.3 — Trocar e apagar o link  [US-07, RN-23]
Given um lobby aberto com link do Discord
When o dono apaga o link na edição
Then o lobby fica sem link
  And a seção dos membros deixa de mostrar "Discord do anfitrião"

CA-07.4 — Membros veem o link  [US-07, RN-18, RN-19]
Given um lobby com link do Discord
When o dono ou um membro aceito abre o detalhe
Then a seção mostra o link como "Discord do anfitrião"

CA-07.5 — Quem não é membro não vê o link  [US-07, RN-19]
Given um lobby com link do Discord
When um visitante, um Usuário sem candidatura ou um candidato pendente abre o detalhe
Then a página não mostra o link
  And a resposta da API para ele não traz o link

CA-07.6 — Link não vaza para fora do grupo  [US-07, RN-25]
Given um lobby com link do Discord
When alguém vê o card na home, copia o convite, cola o link do lobby no Discord ou vê o lobby no banco de talentos
Then o link do Discord do dono não aparece em nenhum deles

CA-07.7 — Link sem o bot  [US-07, RN-24]
Given a API sem a configuração do bot
  And um lobby com link do Discord
When um membro aceito abre o detalhe
Then a seção mostra o link do anfitrião
  And não mostra nada sobre a sala do bot
```

### US-08 — Avisar no Discord
```gherkin
CA-08.1 — Dono avisa o grupo  [US-08, RN-26]
Given um lobby com sala aberta, com o dono e dois membros conectados
When o dono escolhe "Avisar no Discord"
Then o canal de texto recebe "<nick do dono> está chamando o grupo" com o link do lobby
  And a mensagem menciona os dois membros e não menciona o dono

CA-08.2 — Intervalo de 10 minutos  [US-08, RN-27]
Given um lobby cujo dono avisou o grupo às 19:00
When o dono tenta avisar de novo às 19:04
Then o aviso é recusado com "Aguarde 6 min para avisar de novo"
  And o Discord não recebe nenhuma chamada
  And o detalhe mostra o botão desabilitado com o tempo que falta

CA-08.3 — Depois do intervalo  [US-08, RN-27]
Given um lobby cujo dono avisou o grupo às 19:00
When o dono avisa de novo às 19:10
Then o canal de texto recebe o novo aviso

CA-08.4 — Só o dono avisa  [US-08, RN-28]
Given um lobby com sala aberta
When um membro aceito que não é o dono tenta avisar
Then o aviso é recusado
  And o canal de texto não recebe mensagem

CA-08.5 — Sem sala aberta  [US-08, RN-26]
Given um lobby sem sala, ou com a sala encerrando ou apagada
When o dono abre o detalhe
Then o botão "Avisar no Discord" não aparece
  And o aviso pela API é recusado

CA-08.6 — Falha no Discord não consome o intervalo  [US-08, RN-28]
Given um lobby com sala aberta
  And o Discord respondendo com erro
When o dono avisa o grupo às 19:00
Then o dono vê "Não foi possível avisar agora. Tente de novo em instantes"
  And o aviso não é refeito depois
  And, com o Discord de volta, o dono consegue avisar às 19:01

CA-08.7 — Lembrete não conta para o intervalo  [US-08, RN-27]
Given um lobby com sala aberta que começa às 20:00
  And o lembrete das 19:45 já foi postado
When o dono avisa o grupo às 19:46
Then o canal de texto recebe o aviso
```

## 6. Casos de borda
- Jogador sai do servidor pelo próprio Discord → o RO Lobby não percebe; ele continua
  "conectado" e as permissões continuam valendo se voltar. Para voltar, ele reconecta pelo
  perfil [RN-04].
- Dono exclui a conta ou perde o personagem → segue as regras atuais do lobby; a sala segue
  o estado do lobby [RN-16, RN-17].
- Canal renomeado ou mexido à mão no Discord → o bot não corrige; só usa os IDs dos canais
  [RN-05].
- Dois cliques em "Abrir sala" ao mesmo tempo → só uma sala é criada, e a outra tentativa é
  recusada [RN-05].
- A API reinicia com efeitos pendentes ou horários vencidos → eles são feitos quando ela
  volta, uma vez só [RN-14, RN-16, RN-21]. Um lembrete que venceu depois do início do
  lobby é descartado, porque não serve mais [RN-14].
- Lobby lotado ou com vagas → não importa para a sala [RN-06].
- Dois cliques em "Avisar no Discord" ao mesmo tempo → só um aviso é postado [RN-27].
- Link do dono que expirou ou foi revogado no Discord → o RO Lobby não confere; o dono
  troca o link na edição [RN-23].
- Membro removido que tinha o link do dono → já viu o link e pode guardá-lo; o RO Lobby só
  deixa de mostrar. Cabe ao dono revogar o convite no Discord dele [RN-19].

## 7. Requisitos não funcionais
- **RNF-01** — O token do bot existe só no ambiente da API. Ele não aparece no web, no
  repositório nem no log. Os tokens dos usuários seguem a RN-02.
- **RNF-02** — O cliente do Discord respeita o rate limit (`429` com `retry_after`) e tem
  prazo nas chamadas. Na abertura da sala, a resposta ao dono não passa de 10 segundos.
- **RNF-03** — Os testes e o CI usam o Discord falso (`discordfake`), nunca o Discord real.
- **RNF-04** — Os efeitos na sala e os horários (lembrete, remoção) ficam no banco e são
  idempotentes: sobrevivem a um reinício da API sem duplicar mensagens nem canais.
- **RNF-05** — Horários em UTC no banco e exibidos no horário de Brasília (CLAUDE.md).
- **RNF-06** — O `README.md` explica como criar o servidor oficial, ativar o bot no
  aplicativo do Discord, convidá-lo com as permissões necessárias e preencher a
  configuração.

## 8. Fora de escopo
- Comandos de barra (`/lobby`) e qualquer conexão por gateway — o REST basta para esta
  entrega.
- Bot em servidores de guilda e webhook do líder (opções 1 e 3) — ficam para o épico de
  Guildas.
- Mensagem privada (DM) — a menção na sala já notifica.
- Cargos por classe ou função no servidor.
- Desconectar a conta pelo RO Lobby — para sair, o jogador sai do servidor no Discord.
- Detectar quem saiu do servidor pelo Discord — exigiria o gateway.
- Abrir a sala sozinho quando o lobby lota.
- Criar o servidor oficial — é feito à mão pelo operador (RNF-06).
- Avisar no Discord do dono — o bot não está nos servidores dos jogadores; o aviso vale só
  na sala oficial.
- Texto livre no aviso — evita abuso e moderação; o grupo conversa na própria sala.
- Conferir se o link do dono ainda funciona no Discord.

## 9. Perguntas em aberto
- [x] ID do servidor oficial — `1560117933705797643`, informado pelo usuário em 2026-10-14.
  A categoria deixou de ser configuração (D-04). Falta o token do bot, que vai só no `.env`.

## 10. Decisões tomadas na entrevista
- O que abre a sala → o dono, num botão, a qualquer momento antes do início.
- Onde o jogador se conecta → no perfil, por opt-in, uma vez; o login fica só com `identify`.
- Membro não conectado → a sala abre sem ele; o detalhe mostra quem está fora e o convite
  para se conectar; ao conectar, ele entra sozinho.
- Entrada e saída depois da sala aberta → a sala acompanha; troca de personagem não muda nada.
- Conteúdo da sala → canal de texto e de voz, dentro de uma categoria.
- Lembrete → 15 minutos antes, P2.
- Fim → 6 horas depois do início; no cancelamento, aviso com motivo e remoção em 1 hora.
- Fora de escopo → comandos de barra, guildas, DM, webhook, cargos por classe e a criação do
  servidor.
- (Revisão de 2026-10-14) Link do Discord do dono → entra, opcional, visível só para os
  membros, sem depender do bot.
- (Revisão de 2026-10-14) "Avisar no Discord" → entra, só para o dono, na sala oficial, com
  intervalo mínimo entre avisos para evitar abuso.
