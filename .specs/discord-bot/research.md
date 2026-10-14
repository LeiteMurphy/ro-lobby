# Research — Bot do Discord no servidor oficial

## Pedido
Quando um lobby fica pronto, o bot do RO Lobby cria no servidor oficial do Discord um canal
privado só para os membros do grupo, menciona todos, lembra antes do horário e apaga o canal
depois. O jogador entra no servidor sozinho pelo escopo OAuth `guilds.join`, por opt-in.
Esta é a opção 2 da conversa de 2026-10-13. As opções 1 (webhook do líder) e 3 (bot nos
servidores das guildas) ficaram para depois.

## O que o projeto já tem
- `backend/internal/discord/client.go`: cliente REST com `net/http` que troca o `code` do
  OAuth pelo perfil (`/users/@me`). O escopo é só `identify` (RN-01 da `login-discord`), e
  o token do Discord é descartado logo depois (RN-04). O Usuário guarda só `discord_id`,
  `username` e `global_name` (RN-06).
- `backend/internal/discordfake` e `cmd/fakediscord`: Discord falso usado nos testes e na
  pilha local. Dá para estender com as rotas do bot.
- O fluxo OAuth começa no web (`web/src/lib/auth/oauth.ts`, com `scope=identify`), passa o
  `code` à API e a API troca o código (ADR-07).
- Lobby (`00004_lobbies.sql`): o estado não é coluna. O lobby está cancelado quando
  `cancelled_at` existe, iniciado quando `starts_at` já passou e aberto nos outros casos.
  **Não existe o estado "pronto/cheio"** nem horário de término.
- Membros do lobby: o dono, pelo `owner_character_id`, e as candidaturas `accepted`. Um
  membro pode sair, ser removido ou trocar de personagem (`applications`,
  `swap_requests`).
- **Não há tarefas em segundo plano.** A expiração de candidaturas é calculada na leitura.
  A API é um único processo (`cmd/api/main.go`) sem agendador.
- Config: `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET` e `DISCORD_API_BASE_URL`. Ainda não
  há token de bot.
- Hospedagem: o PC do usuário, publicado pelo túnel (ADR-06, ADR-08).

## API do Discord (o que importa aqui)
- **Bot**: é o mesmo aplicativo do OAuth, com um usuário bot. As chamadas REST levam
  `Authorization: Bot <token>`. Para criar canal, mandar mensagem e apagar canal, o REST
  basta; o gateway (WebSocket) não é necessário.
- **`guilds.join`**: `PUT /guilds/{guild}/members/{user}` com o token do bot e o
  `access_token` do usuário, que precisa ter o escopo `guilds.join`. O bot precisa estar no
  servidor com a permissão *Create Instant Invite*. O token do usuário só é usado nesse
  momento, então dá para descartá-lo em seguida e manter a regra de não guardar tokens.
- **Canal privado**: `POST /guilds/{guild}/channels` com `permission_overwrites`. Nega
  `VIEW_CHANNEL` para `@everyone` e libera para cada membro. Também dá para editar o
  overwrite de um membro (`PUT /channels/{id}/permissions/{user}`) e apagar o canal
  (`DELETE /channels/{id}`). O bot precisa de *Manage Channels* e *Manage Roles*.
- **Menção**: `<@discord_id>` no texto da mensagem notifica quem está no servidor e tem
  acesso ao canal.
- **Limites**: 500 canais por servidor, 50 por categoria e rate limits por rota (o
  cliente precisa respeitar o `429` com `retry_after`).
- **DM**: só é possível com servidor em comum. Aqui não é necessária.

## Pontos que a spec precisa decidir
1. O que dispara o "lobby pronto": vagas cheias, ação do líder ou o horário chegando.
2. Consentimento: onde e quando o jogador conecta a conta ao servidor, e o que acontece
   com membros que não conectaram.
3. Mudanças no grupo depois do canal criado: entrada, saída, remoção e troca.
4. Ciclo de vida do canal: lembrete, quando apagar (o lobby não tem término) e
   cancelamento.
5. Texto ou voz, e o que a mensagem traz.
6. Falhas: Discord fora do ar, bot sem permissão, jogador que saiu do servidor.

## Classificação: **G (Grande)**
É uma integração externa nova (bot e `guilds.join`), com estado novo (vínculo
lobby↔canal e consentimento do jogador), um agendador em segundo plano que o projeto ainda
não tem e uma decisão de arquitetura que pede ADR (ADR-09) antes do código. Passa por todas
as fases.
