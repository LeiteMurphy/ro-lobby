# Tasks — Sala do grupo no Discord

- Spec: `./spec.md` · Design: `./design.md` (aprovado) · ADR: `docs/adr/0009-bot-discord-rest-outbox.md`
- Notion: [Épico](https://app.notion.com/p/3fad4a3a5eff81468b99db1fa0b99fc5)
- Branch: `feature/discord-bot`

## Backend

### T-01 — Migração e configuração do bot  [x]
- Cobre: RN-05 (uma sala por lobby), RN-09 (estados), RN-22, RNF-01, RNF-04 (modelo), D-02, D-04
- Depende de: —
- Paralelizável: não
- Arquivos:
  - `backend/migrations/00011_discord.sql`, com `users.discord_connected_at`,
    `lobbies.discord_link`, `discord_rooms` e `discord_jobs`;
  - `backend/internal/config` (`DISCORD_BOT_TOKEN` e `DISCORD_GUILD_ID`);
  - `docker-compose.yml`, `.env.example`.
- Pronto quando:
  - a migração aplica e reverte com dados antigos;
  - os CHECK recusam um estado inválido, duas salas no mesmo lobby e um link com mais
    de 200 caracteres;
  - os testes da config provam que:
    - sem nenhuma das duas variáveis, a feature fica desligada;
    - com as duas, ela liga;
    - com só uma delas, a subida dá erro;
  - o token não aparece no log nem no `String()` da config.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff81f2afedc879dfab7577

### T-02 — Cliente do bot e Discord falso  [P] [x]
- Cobre: RN-07, RN-20 (erros do Discord), RNF-02, RNF-03, D-01, D-04, D-06, CA-02.7
  (compensação), CA-02.8 (limite), CA-05.5 (404 conta como sucesso)
- Depende de: —
- Paralelizável: [P] com T-01 (pacotes diferentes)
- Arquivos: `backend/internal/discordbot/`, `backend/internal/discordfake/` (rotas do bot e
  registro do que ele fez), `backend/cmd/fakediscord`
- Pronto quando: os testes contra o Discord falso provam que:
  - `CreateRoom` cria a categoria, o texto e a voz com os overwrites (nega `@everyone`,
    libera o bot e os IDs);
  - uma falha no meio apaga o que já foi criado;
  - o `429` é respeitado uma vez e depois vira `ErrUnavailable`;
  - o código `30013` vira `ErrGuildFull`;
  - apagar um canal que já não existe é sucesso;
  - toda mensagem vai com `allowed_mentions.parse = []` e só os IDs listados.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff81679476ed56596d8689

### T-03 — Trabalhador: fila, novas tentativas, lembrete e remoção  [x]
- Cobre: US-05, US-06, RN-09, RN-14, RN-16, RN-21, RNF-04, D-02, CA-05.1, CA-05.2, CA-05.4,
  CA-05.5, CA-06.1, CA-06.2, CA-06.3
- Depende de: T-01, T-02
- Paralelizável: não
- Arquivos: `backend/internal/discordroom/worker.go`, `backend/queries/discord.sql`,
  `backend/cmd/api/main.go` (sobe o trabalhador com a feature ligada)
- Pronto quando: os testes de integração, com relógio fixo e o Discord falso, provam:
  - o job executado uma vez só, e de novo depois de uma falha, com espera crescente;
  - o job descartado quando a sala foi apagada;
  - o lembrete 15 min antes, uma vez só, que não sai com a sala aberta em cima da hora,
    com o lobby cancelado ou depois do início;
  - a remoção 6 h depois do início, e não antes;
  - a sala encerrando apagada 1 h depois do cancelamento;
  - a remoção refeita quando o Discord volta;
  - um canal apagado à mão não trava a sala.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff818bbf78e9ca8e171fc2

### T-04 — Abrir a sala e o bloco `discord` do detalhe (API)  [x]
- Cobre: US-02, US-03, RN-05, RN-06, RN-07, RN-08, RN-18, RN-19, RN-20, D-05, CA-02.1 a
  CA-02.8, CA-03.1 a CA-03.5 (API)
- Depende de: T-03
- Paralelizável: não
- Arquivos: `backend/internal/discordroom/` (`Open`, `View`), `backend/internal/lobbies/detail.go`,
  `backend/internal/server/` (`POST /lobbies/{id}/discord-room`, bloco `discord` em
  `GET /lobbies/{id}`), `openapi.yaml`
- Pronto quando: os testes de integração provam:
  - a sala aberta pelo dono, com os canais e as permissões;
  - a mensagem de abertura (composição com nick e classe, menções só dos conectados,
    nick de quem está fora e link do lobby);
  - as recusas `not_owner`, `not_open` e `room_exists`;
  - `discord_unavailable` e `guild_full` sem sala gravada nem canal sobrando;
  - dois cliques ao mesmo tempo, com uma sala só;
  - o bloco `discord` só para os membros, com o link apenas com a sala aberta.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff812d902cdcc622553c98

### T-05 — Conexão ao servidor (API)  [x]
- Cobre: US-01, RN-01, RN-02, RN-03, RN-04, RN-22, CA-01.1 (API), CA-01.4, CA-01.6,
  CA-01.7 (API)
- Depende de: T-01, T-02
- Paralelizável: [P] com T-04 (arquivos diferentes, exceto o `openapi.yaml`, que fica
  para quem commitar depois)
- Arquivos: `backend/internal/discord/client.go` (`JoinGuildFromCode`), `backend/internal/discordroom/`
  (`Connect`), `backend/internal/server/` (`POST /me/discord-connection`, campos novos de
  `GET /me`), `openapi.yaml`
- Pronto quando: os testes provam:
  - a inclusão no servidor com o token do usuário e a data gravada;
  - o token sem ir para o banco nem para o log;
  - `wrong_account` sem chamar a inclusão;
  - reconectar atualizando a data;
  - `discord_disabled` com a feature desligada.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff81d38ed4dbbae8e1d813

### T-06 — A sala acompanha o grupo  [ ]
- Cobre: US-04, US-05, US-06, RN-10, RN-11, RN-12, RN-13, RN-15, RN-17, RN-21, CA-04.1 a
  CA-04.6, CA-05.3, CA-06.4
- Depende de: T-04, T-05
- Paralelizável: não
- Arquivos: `backend/internal/discordroom/` (`OnMemberJoined`, `OnMemberLeft`,
  `OnStartChanged`, `OnCancelled`), `backend/internal/applications/`, `backend/internal/lobbies/`
- Pronto quando: os testes de integração provam:
  - aceite com e sem conexão;
  - saída e remoção sem o motivo no Discord;
  - a troca de personagem sem efeito;
  - a conexão tardia liberando só as salas abertas;
  - a mudança de horário avisando e movendo o lembrete e a remoção;
  - o cancelamento com motivo e sala encerrando;
  - as ações do grupo funcionando com o Discord fora do ar, com os efeitos refeitos
    depois.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff818291abc786b9174adb

### T-07 — Link do Discord do dono (API)  [P] [ ]
- Cobre: US-07, RN-23, RN-24, RN-25, CA-07.1 a CA-07.7 (API)
- Depende de: T-04
- Paralelizável: [P] com T-05 e T-06 (só `lobbies` e o bloco `discord`)
- Arquivos: `backend/internal/lobbies/` (validação e gravação), `backend/internal/discordroom/`
  (`View`), `openapi.yaml` (`discordLink` em criar e editar)
- Pronto quando: os testes provam:
  - os três formatos aceitos, e os recusados (`http://`, outro host, mais de 200
    caracteres);
  - trocar e apagar o link;
  - o link só para os membros;
  - o link ausente da lista da Home, do talento e do convite;
  - o link presente com a feature desligada.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff813c8196d9e6e9ee3b21

### T-08 — Avisar no Discord (API)  [ ]
- Cobre: US-08, RN-26, RN-27, RN-28, CA-08.1 a CA-08.7 (API)
- Depende de: T-04
- Paralelizável: não
- Arquivos: `backend/internal/discordroom/` (`Alert`), `backend/internal/server/`
  (`POST /lobbies/{id}/discord-alert`, `alertAvailableAt`), `openapi.yaml`
- Pronto quando: os testes provam:
  - o aviso com as menções sem o dono;
  - o `429` com `retryAfterSeconds` e sem chamada ao Discord;
  - o aviso de novo depois de 10 min;
  - só o dono avisa, e só com a sala aberta;
  - a falha que não conta para o intervalo e não é refeita;
  - o lembrete que não conta para o intervalo;
  - dois avisos simultâneos postando um só.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff814baa21ff1d35ef8e12

## Web

### T-09 — Perfil: entrar no servidor do RO Lobby  [ ]
- Cobre: US-01, RN-01, RN-02, RN-04, RN-22, CA-01.1 a CA-01.7 (tela)
- Depende de: T-05
- Paralelizável: [P] com T-10 e T-11
- Arquivos: `web/src/routes/perfil/discord/` (início e retorno do OAuth, com `state`),
  `web/src/routes/perfil/`, `web/src/lib/auth/oauth.ts`
- Pronto quando: os testes de SSR e o e2e (com o Discord falso) provam:
  - a URL com `identify guilds.join` e o `state`;
  - "Conectado ao servidor do RO Lobby" com a data;
  - as mensagens de conta diferente e de conexão cancelada;
  - "Reconectar";
  - nada disso aparece com a feature desligada;
  - o login continua só com `identify`.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff81e7b183df1ef6ce1f88

### T-10 — Detalhe do lobby: seção "Sala no Discord"  [ ]
- Cobre: US-02, US-03, US-07, US-08, RN-18, RN-19, RN-24, RN-26, RN-27, CA-02.1, CA-02.7,
  CA-02.8 (mensagens), CA-03.1 a CA-03.5, CA-07.4, CA-07.5, CA-07.7, CA-08.2 (contagem),
  CA-08.5
- Depende de: T-04, T-07, T-08
- Paralelizável: [P] com T-09 e T-11
- Arquivos: `web/src/lib/discord/` (`DiscordSection.svelte`, mensagens),
  `web/src/routes/lobbies/[id]/` (actions `openRoom` e `alert`)
- Pronto quando: os testes de SSR e o e2e provam:
  - os estados da RN-18 para dono, membro conectado e membro não conectado;
  - nada para quem não é membro;
  - o "Abrir sala no Discord";
  - o "Avisar no Discord" com o tempo que falta;
  - o "Discord do anfitrião";
  - as mensagens de erro em pt-BR.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff816da953f02c75cfaff5

### T-11 — Link do Discord no formulário do lobby  [P] [ ]
- Cobre: US-07, RN-23, CA-07.1, CA-07.2, CA-07.3 (tela)
- Depende de: T-07
- Paralelizável: [P] com T-09 e T-10
- Arquivos: `web/src/lib/lobbies/components/LobbyForm.svelte`, `form.ts`, `messages.ts`
- Pronto quando: os testes de SSR e o e2e provam:
  - o campo opcional com a ajuda "Só os membros aceitos veem";
  - o erro "Use um link de convite do Discord" junto do campo;
  - apagar o link na edição.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff8102ab4bdd8dcb3d7cb9

### T-12 — Operação: README e pilha  [ ]
- Cobre: RNF-01, RNF-06
- Depende de: T-01 a T-11
- Paralelizável: não
- Arquivos: `README.md`, `.env.example`, `docker-compose.yml`
- Pronto quando: o README explica:
  - criar o servidor oficial;
  - ativar o bot e gerar o token;
  - o link de convite com as permissões (Manage Channels, Manage Roles, Create Instant
    Invite, Send Messages, View Channel);
  - registrar o redirect `…/perfil/discord/retorno`;
  - preencher `DISCORD_BOT_TOKEN` e `DISCORD_GUILD_ID`;
  - regenerar o token se vazar.

  A pilha `app` sobe com e sem as variáveis.
- Commit:
- Notion: https://app.notion.com/p/3fad4a3a5eff81f1a25bcb2abbd03f41

## Matriz de cobertura
| Critério | Tasks |
|---|---|
| CA-01.1 | T-05, T-09 |
| CA-01.2 | T-09 |
| CA-01.3 | T-09 |
| CA-01.4 | T-05, T-09 |
| CA-01.5 | T-09 |
| CA-01.6 | T-05, T-09 |
| CA-01.7 | T-01, T-05, T-09 |
| CA-02.1 | T-04, T-10 |
| CA-02.2 | T-04 |
| CA-02.3 | T-04 |
| CA-02.4 | T-04 |
| CA-02.5 | T-04 |
| CA-02.6 | T-04 |
| CA-02.7 | T-02, T-04, T-10 |
| CA-02.8 | T-02, T-04, T-10 |
| CA-03.1 | T-04, T-10 |
| CA-03.2 | T-04, T-10 |
| CA-03.3 | T-04, T-10 |
| CA-03.4 | T-04, T-10 |
| CA-03.5 | T-04, T-10 |
| CA-04.1 | T-06 |
| CA-04.2 | T-06 |
| CA-04.3 | T-06 |
| CA-04.4 | T-06 |
| CA-04.5 | T-06 |
| CA-04.6 | T-06 |
| CA-05.1 | T-03 |
| CA-05.2 | T-03 |
| CA-05.3 | T-06 |
| CA-05.4 | T-03 |
| CA-05.5 | T-02, T-03 |
| CA-06.1 | T-03 |
| CA-06.2 | T-03 |
| CA-06.3 | T-03 |
| CA-06.4 | T-06 |
| CA-07.1 | T-07, T-11 |
| CA-07.2 | T-07, T-11 |
| CA-07.3 | T-07, T-11 |
| CA-07.4 | T-07, T-10 |
| CA-07.5 | T-07, T-10 |
| CA-07.6 | T-07 |
| CA-07.7 | T-07, T-10 |
| CA-08.1 | T-08 |
| CA-08.2 | T-08, T-10 |
| CA-08.3 | T-08 |
| CA-08.4 | T-08 |
| CA-08.5 | T-08, T-10 |
| CA-08.6 | T-08 |
| CA-08.7 | T-08 |

Sem lacunas: todos os 49 critérios da spec têm task.

## Descobertas
- 2026-10-14 — A migração passa de `00010` (design) para `00011`, porque a `00010` foi do
  servidor do jogo.
- 2026-10-14 — O teste real com o Discord (fora do Discord falso) depende do servidor
  oficial e dos IDs, pendência da spec. Ele entra no roteiro de UAT depois da T-12.
- 2026-10-14 — O estado da conexão saiu de `GET /me` (design) para `GET /me/discord-connection`,
  para não mudar o esquema `User`, que o web inteiro usa. O `POST` fica na mesma rota.
- 2026-10-14 — A API ganhou `APP_ORIGIN` (a mesma origem do web), para o link do lobby na
  mensagem de abertura (RN-08).
