# Relatório de validação — discord-bot (ciclo 1)

**Veredito geral:** APROVADO

**Execução (2026-10-15, branch `feature/discord-bot`, `main..HEAD` = 19 commits, 78 arquivos):**
- backend: `go test ./...` passou · `go test -tags=integration ./...` (`-count=1`, PostgreSQL de dev) passou em todos os pacotes · `gofmt -l .` vazio · `golangci-lint run ./...` sem problemas (0 issues) · `go generate ./...` sem diferença no `git status`
- web: `npm run lint` (prettier e eslint) passou · `npm run check` com 0 erros e 0 avisos (556 arquivos) · `npm test` com 39 arquivos e 435 testes passando · `npm run build` passou
- ponta a ponta: `npx playwright test` com **66 testes passando** (47 s), entre eles os 5 de `test/e2e/discord.spec.ts`
- A pilha `app` (produção pelo túnel) não foi tocada. A API dela não publica a porta 8080 no host, então o e2e rodou sem conflito.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| User Stories (US-01..US-08) | 8 | 0 | 0 | 0 |
| Regras de domínio (RN-01..RN-28) | 27 | 1 | 0 | 0 |
| Critérios de aceite (49) | 48 | 1 | 0 | 0 |
| Requisitos não funcionais (RNF-01..06) | 6 | 0 | 0 | 0 |
| Decisões do design (D-01..D-06) | 6 | 0 | 0 | 0 |
| Fora de escopo (11 itens respeitados) | 11 | 0 | 0 | 0 |

Os dois ⚠️ (RN-25 e CA-07.6) são da US-07, que é P2. Todos os itens P1 estão ✅.

## Detalhe por item

Abreviações: `room.go`, `hooks.go`, `worker.go`, `alert.go`, `connect.go` e `messages.go` ficam em `backend/internal/discordroom/`. `srv/` é `backend/internal/server/`. `DIT` é `srv/discord_integration_test.go`, `SIT` é `srv/discord_sync_integration_test.go`, `AIT` é `srv/discord_alert_integration_test.go`, `LIT` é `srv/discord_link_integration_test.go` e `WIT` é `discordroom/worker_integration_test.go`.

### Regras de domínio
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `web/src/lib/auth/oauth.ts` (`CONNECT_SCOPE`, login continua com `identify`), `web/src/routes/perfil/discord/+server.ts` | `perfil/discord/discord.spec.ts` (CA-01.2, CA-01.3), e2e CA-01.1/01.3/01.6 | |
| RN-02 | ✅ | `discord/client.go` `JoinFromCode` (o token só vai para `join`), `discordbot/client.go` `AddMember`, migração só com `users.discord_connected_at` | `db/queries_integration_test.go` `TestSchema_CA01_4_NoTokensNorAvatar` (colunas de users), `TestDiscordConnect_CA01_1_JoinsGuild` | Log: por construção, nenhum `slog` do caminho recebe token, e os erros só levam método e caminho. Ver Observações. |
| RN-03 | ✅ | `discord/client.go` (`ErrWrongAccount` antes de `join`), `connect.go` | `DIT` `TestDiscordConnect_CA01_4_WrongAccount`, e2e CA-01.4 | |
| RN-04 | ✅ | `queries/discord.sql` `SetDiscordConnected`; o falso devolve 204 se o membro já existe | `DIT` `TestDiscordConnect_CA01_6_Reconnect` | |
| RN-05 | ✅ | PK `discord_rooms.lobby_id`; `room.go` `roomName` (título + 4 caracteres do id); `CreateRoom` (categoria, texto e voz) | `TestCreateRoom_RN05_RN07_...`, `TestRoomName_RN05_...`, `DIT` `TestDiscordRoom_RN05_ConcurrentOpen`, `TestMigration00011_RN05_RN09_...` | |
| RN-06 | ✅ | `room.go` `Open`: `FOR UPDATE`, dono, não cancelado, `starts_at > now`, sem sala | `DIT` CA-02.4, CA-02.5, CA-02.6 | Não existe outro gatilho de abertura (busca por `CreateRoom`). |
| RN-07 | ✅ | `discordbot/client.go` `overwrites` (nega VIEW a `@everyone`, libera o bot e os IDs); `connectedIDs` | `TestCreateRoom_RN05_RN07_...` (bits de deny e allow), `DIT` CA-02.1/CA-02.3 | |
| RN-08 | ✅ | `messages.go` `openingMessage` (título, servidor, horário de Brasília, composição com nick e classe, menções, quem está fora, link) | `DIT` `TestDiscordRoom_CA02_2_OpeningMessage` | |
| RN-09 | ✅ | CHECKs de status na migração; `MarkRoomDeleted`; `Open` recusa qualquer sala existente | `TestMigration00011_...`, `DIT` CA-02.6 (aberta, encerrando e apagada) | |
| RN-10 | ✅ | `applications.go:329` → `hooks.go` `OnMemberJoined` | `SIT` `TestDiscordSync_CA04_1_CA04_2_Accept` | |
| RN-11 | ✅ | `applications.go:474` (`endMembership`, que serve a sair e a remover) → `OnMemberLeft`, `leftMessage` sem o motivo | `SIT` `TestDiscordSync_CA04_3_LeaveAndRemove` | |
| RN-12 | ✅ | `swaps.go` e a troca do dono não chamam ganchos (a busca por `discordroom` só encontra `decide`, `endMembership`, `Update` e `Cancel`) | `SIT` `TestDiscordSync_CA04_4_SwapDoesNothing` | O teste cobre a troca do dono. A troca do membro aceito não tem gancho (conferido no código). |
| RN-13 | ✅ | `hooks.go` `onConnected` + `OpenRoomsOfMember` (só `status='open'`) | `SIT` `TestDiscordSync_CA04_5_LateConnection` | |
| RN-14 | ✅ | `DueReminders` (janela, não cancelado, `opened_at <= início-15min`, `starts_at > now`), `MarkReminderSent` | `WIT` CA-06.1, CA-06.2, CA-06.3 | |
| RN-15 | ✅ | `lobbies.go:509-514` → `OnStartChanged` (`ResetDiscordReminder` + post); a remoção já vem de `l.starts_at` | `SIT` `TestDiscordSync_CA06_4_StartChanged` | |
| RN-16 | ✅ | `DueRoomDeletions` (`starts_at <= now-6h`), `deleteDueRooms` | `WIT` `TestDeletion_CA05_1_CA05_2_...` | |
| RN-17 | ✅ | `lobbies.go:584` → `OnCancelled` (post com motivo, `CloseDiscordRoom`); `closing_at <= now-1h` | `SIT` CA-05.3, `WIT` `TestDeletion_RN17_ClosingAfterOneHour` | |
| RN-18 | ✅ | `room.go` `View`; `web/src/lib/discord/components/DiscordSection.svelte` | `DIT` CA-03.1/03.3/03.5, `section.spec.ts`, e2e | |
| RN-19 | ✅ | `srv/discord.go` `viewerIsMember` (dono ou aceito) antes de `rooms.View` em `srv/lobbies.go` `GetLobby` | `DIT` `TestDiscordRoom_CA03_1_CA03_4_DetailBlock` (visitante, pendente e outro, no JSON bruto), `LIT` CA-07.5, e2e | |
| RN-20 | ✅ | `discordbot` `ErrGuildFull` (código 30013), desfaz o que criou; `room.go` desfaz se a gravação falhar; mensagens em `web/src/lib/discord/messages.ts` | `TestCreateRoom_CA02_7_...`, `_CA02_8_...`, `DIT` `TestDiscordRoom_CA02_7_CA02_8_DiscordFails`, `section.spec.ts` | |
| RN-21 | ✅ | Ganchos só enfileiram na transação; `worker.go` com novas tentativas (2ⁿ s, até 5 min) e descarte com a sala apagada | `SIT` CA-04.6, `WIT` `TestJobs_RN21_*`, CA-05.4 | |
| RN-22 | ✅ | `config.go` (as duas variáveis juntas, só uma é erro), `main.go` (bot nulo), `noRooms` | `TestLoad_RN22_BotOnlyWithTokenAndGuild`, `DIT` `TestDiscordRoom_RN22_Disabled`, CA-01.7, `page.spec.ts` | |
| RN-23 | ✅ | `lobbies.go` `CheckDiscordLink` (regex dos 3 formatos, `https`, 200 caracteres), CHECK na migração, `LobbyForm.svelte` | `discord_link_test.go`, `LIT` CA-07.2/07.3, `discord-link.spec.ts`, e2e | |
| RN-24 | ✅ | `View` preenche `OwnerLink` antes de olhar `Enabled` | `LIT` `TestOwnerLink_CA07_7_WithoutBot`, `section.spec.ts` | |
| RN-25 | ⚠️ | Por construção: o link só sai em `LobbyDiscord.ownerLink`; `share.ts`, a lista e os talentos não usam o campo (busca por `ownerLink`/`discord_link`) | `LIT` `TestOwnerLink_CA07_5_CA07_6_...` cobre o detalhe para quem não é membro (a fonte do Open Graph) e a lista da Home | Faltam testes do convite copiado, do banco de talentos e das mensagens do bot. A task T-07 prometia "do talento e do convite". P2, não bloqueia. |
| RN-26 | ✅ | `alert.go` (sala aberta, menção sem o dono, link), botão só com `room.status==='open'` e dono | `AIT` CA-08.1, CA-08.4/08.5, `section.spec.ts` | |
| RN-27 | ✅ | `alert.go` (`CooldownError` antes de chamar o Discord, `SetLastAlert` só no sucesso, `FOR UPDATE`), `alertAvailableAt` só para o dono | `AIT` CA-08.2/08.3, CA-08.7, `TestDiscordAlert_RN27_Concurrent`, e2e | |
| RN-28 | ✅ | `alert.go` (dono; falha desfaz a transação; não enfileira) | `AIT` CA-08.4, CA-08.6, `discord.server.spec.ts` | |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | `connect.go`, `discord/client.go`, `perfil/+page.svelte` ("Conectado ao servidor do RO Lobby em <data>") | `DIT` `TestDiscordConnect_CA01_1_JoinsGuild` (o falso confere o `access_token`), teste de schema, `page.spec.ts`, e2e | A parte "não aparece no log" está garantida por construção e não tem teste explícito para o token do jogador. |
| CA-01.2 | ✅ | `oauth.ts` (padrão `identify`) | `discord.spec.ts` "CA-01.2" | |
| CA-01.3 | ✅ | `perfil/discord/+server.ts`, `retorno/+server.ts` (`sameState`) | `discord.spec.ts` CA-01.3 e "state diferente", e2e | |
| CA-01.4 | ✅ | `ErrWrongAccount` → 409 `wrong_account` → `?discord=conta` | `DIT` CA-01.4 (sem membro no falso, sem data), `discord.spec.ts`, e2e | |
| CA-01.5 | ✅ | `retorno/+server.ts` (`error` → `cancelado`, sem API) | `discord.spec.ts`, e2e | |
| CA-01.6 | ✅ | `SetDiscordConnected` | `DIT` CA-01.6 (10/10 → 14/10), e2e | |
| CA-01.7 | ✅ | `noRooms`, `Connect` → `ErrDisabled` → 404 `discord_disabled`; o perfil esconde a seção | `DIT` CA-01.7, `page.spec.ts` CA-01.7, `discord.spec.ts` | |
| CA-02.1 | ✅ | `Open` + `CreateRoom` | `DIT` CA-02.1, `TestCreateRoom_RN05_RN07` (bot), e2e | |
| CA-02.2 | ✅ | `openingMessage` | `DIT` CA-02.2 (servidor Freya, 20:00 em Brasília, composição, menções 302/301, Fogo fora, link) | |
| CA-02.3 | ✅ | `connectedIDs` | `DIT` CA-02.1/02.3 (303 sem overwrite) | |
| CA-02.4 | ✅ | `not_owner` | `DIT` CA-02.4 (0 canais) | |
| CA-02.5 | ✅ | `not_open` | `DIT` CA-02.5 (iniciado e cancelado, 0 canais) | |
| CA-02.6 | ✅ | `room_exists` | `DIT` CA-02.6 (3 canais no total) | |
| CA-02.7 | ✅ | desfazer em `CreateRoom` e em `Open` | `DIT` CA-02.7 (503, 0 salas, 0 canais), `TestCreateRoom_CA02_7`, `section.spec.ts` | |
| CA-02.8 | ✅ | `ErrGuildFull` → 503 `guild_full` → mensagem | `DIT`, `TestCreateRoom_CA02_8`, `messages.ts` | |
| CA-03.1 | ✅ | `View`, `DiscordSection` | `DIT` CA-03.1, `section.spec.ts`, e2e | |
| CA-03.2 | ✅ | `viewerConnected` → atalho `/perfil/discord` | `section.spec.ts` CA-03.2, e2e | |
| CA-03.3 | ✅ | `DiscordSection` (dono e lobby aberto, ou "A sala ainda não foi aberta.") | `DIT` CA-03.3, `section.spec.ts`, e2e | |
| CA-03.4 | ✅ | `viewerIsMember` | `DIT` (corpo sem `discord.com/channels`), e2e (visitante) | |
| CA-03.5 | ✅ | URL só com `open`; "A sala foi encerrada." | `DIT` CA-03.5, `section.spec.ts` | |
| CA-04.1 | ✅ | `OnMemberJoined` + `sync_access` | `SIT` CA-04.1 | |
| CA-04.2 | ✅ | `joinedMessage` sem menção | `SIT` CA-04.2 (sem acesso, `AllowedMentions.Users` vazio) | |
| CA-04.3 | ✅ | `OnMemberLeft` | `SIT` CA-04.3 (sair e remover, sem o motivo) | |
| CA-04.4 | ✅ | sem gancho nas trocas | `SIT` CA-04.4 | Cobre a troca do dono (ver RN-12). |
| CA-04.5 | ✅ | `onConnected` | `SIT` CA-04.5 (2 abertas liberadas e mencionadas, a apagada intacta) | |
| CA-04.6 | ✅ | fila + novas tentativas | `SIT` CA-04.6 | |
| CA-05.1 | ✅ | `DueRoomDeletions` | `WIT` CA-05.1 | |
| CA-05.2 | ✅ | idem | `WIT` (01:59 continua aberta, com 3 canais) | |
| CA-05.3 | ✅ | `OnCancelled` + remoção 1 h depois | `SIT` CA-05.3, `WIT` RN-17 | |
| CA-05.4 | ✅ | `deleteDueRooms` deixa para a próxima volta | `WIT` CA-05.4 | |
| CA-05.5 | ✅ | 404 conta como sucesso em `DeleteChannel`, `Allow` e `Revoke` | `WIT` CA-05.5, `TestDeleteChannel_CA05_5` | |
| CA-06.1 | ✅ | `sendReminders` | `WIT` CA-06.1 (19:44 nada, 19:45 um só, menções só dos conectados) | |
| CA-06.2 | ✅ | `opened_at <= início-15min` | `WIT` CA-06.2 | |
| CA-06.3 | ✅ | `cancelled_at IS NULL`, `starts_at > now` | `WIT` CA-06.3 | |
| CA-06.4 | ✅ | `OnStartChanged` | `SIT` CA-06.4 (mensagem 21:00, lembrete e remoção no horário novo, não no antigo) | |
| CA-07.1 | ✅ | `Create` + `CheckDiscordLink` | `discord_link_test.go`, `LIT` CA-07.1, `discord-link.spec.ts`, e2e | |
| CA-07.2 | ✅ | `ValidationError` `discordLink`; mensagem "Use um link de convite do Discord" | `LIT` CA-07.2 (link anterior mantido), `discord-link.spec.ts`, e2e | |
| CA-07.3 | ✅ | `Update` com link vazio → NULL; a edição carrega o link atual | `LIT` CA-07.3, `discord-link.spec.ts`, e2e | |
| CA-07.4 | ✅ | `View.OwnerLink`, "Discord do anfitrião" | `LIT` CA-07.4, `section.spec.ts`, e2e | |
| CA-07.5 | ✅ | `viewerIsMember` | `LIT` CA-07.5 (visitante, pendente e outro), e2e | |
| CA-07.6 | ⚠️ | por construção (D-05) | `LIT` cobre a lista da Home e o detalhe para quem não é membro | Sem teste do convite copiado nem do banco de talentos (ver RN-25). |
| CA-07.7 | ✅ | `View` com `Enabled=false` | `LIT` CA-07.7, `section.spec.ts` | |
| CA-08.1 | ✅ | `alertMessage` | `AIT` CA-08.1 (menções 302,303, sem o 301), e2e | |
| CA-08.2 | ✅ | `CooldownError` → 429 `retryAfterSeconds`; `cooldownMessage` | `AIT` (360 s, sem chamadas ao bot), `discord.server.spec.ts`, `section.spec.ts`, e2e (botão desabilitado) | |
| CA-08.3 | ✅ | `now.Before(next)` | `AIT` (aos 10 min, 2 avisos) | |
| CA-08.4 | ✅ | `not_owner` | `AIT` CA-08.4 | |
| CA-08.5 | ✅ | `no_room`; botão só com a sala aberta | `AIT` CA-08.5 (sem sala e encerrando), `section.spec.ts` | Apagada cai no mesmo `status != open` (código). |
| CA-08.6 | ✅ | transação desfeita, sem job | `AIT` CA-08.6, `discord.server.spec.ts` | |
| CA-08.7 | ✅ | o lembrete não grava `last_alert_at` | `AIT` CA-08.7 | |

### User Stories
| ID | Veredito | Observação |
|---|---|---|
| US-01 (P1) | ✅ | CA-01.1..01.7 ✅ |
| US-02 (P1) | ✅ | CA-02.1..02.8 ✅ |
| US-03 (P1) | ✅ | CA-03.1..03.5 ✅ |
| US-04 (P1) | ✅ | CA-04.1..04.6 ✅ |
| US-05 (P1) | ✅ | CA-05.1..05.5 ✅ |
| US-06 (P2) | ✅ | CA-06.1..06.4 ✅ |
| US-07 (P2) | ✅ | CA-07.6 ⚠️, só na cobertura de testes |
| US-08 (P2) | ✅ | CA-08.1..08.7 ✅ |

### Requisitos não funcionais
| ID | Veredito | Evidência | Observação |
|---|---|---|---|
| RNF-01 | ✅ | `config.Bot` `String`/`GoString`/`LogValue`; o token só no `.env`/compose da API | `TestBot_RNF01_TokenHidden`, `TestDo_RNF01_BadTokenNotLogged` |
| RNF-02 | ✅ | `discordbot.do` (prazo de 5 s por chamada, 429 respeitado uma vez se couber no prazo), `OpenTimeout` de 10 s | `TestDo_RNF02_RateLimit`. Ver Observações sobre o pior caso acima de 10 s. |
| RNF-03 | ✅ | `discordfake/bot.go`, `cmd/fakediscord -bot-token`, `playwright.config.ts` | Todos os testes usam o falso |
| RNF-04 | ✅ | `discord_jobs` com `FOR UPDATE SKIP LOCKED`, lembrete e remoção derivados do estado | `TestJobs_RN21_RetryWithBackoffAndRunOnce`, CA-06.1 (uma vez só) |
| RNF-05 | ✅ | `timestamptz`, `messages.go` `when`/`clock` em America/Sao_Paulo, perfil formata em Brasília | CA-02.2, CA-06.1 |
| RNF-06 | ✅ | `README.md` § "Sala do grupo no Discord" (servidor, bot, permissões, redirects, variáveis, trocar o token) | Revisão manual |

### Decisões do design
| ID | Veredito | Evidência |
|---|---|---|
| D-01 | ✅ | `discordbot` só em `net/http` REST, sem gateway nem `discordgo` |
| D-02 | ✅ | `discord_jobs` gravado na transação dos ganchos; lembrete e remoção deduzidos (`DueReminders`, `DueRoomDeletions`); `sync_access` decide na hora (`TestJobs_D02_SyncAccessDecidesAtRunTime`) |
| D-03 | ✅ | `JoinFromCode` entrega o token só para `AddMember`; schema sem token |
| D-04 | ✅ | categoria + texto + voz com overwrites nos três; config sem ID de categoria |
| D-05 | ✅ | bloco `discord` só para dono ou aceito. Foi montado no handler (`srv/lobbies.go`) e não em `Detail`, o que dá na mesma. Testado no JSON bruto. |
| D-06 | ✅ | `allowed_mentions.parse=[]` + IDs (`TestPost_D06_OnlyExplicitMentions`); `escape` do texto do usuário (`TestEscape_D06_...`) |

### Fora de escopo (respeitado)
Comandos de barra e gateway: ausentes. Guildas e webhook: ausentes. DM: ausente. Cargos por classe: ausentes. Desconectar pelo RO Lobby: ausente. Detectar saída do servidor: ausente. Abrir a sala ao lotar: ausente. Criar o servidor: manual (README). Avisar no Discord do dono: ausente. Texto livre no aviso: ausente (texto fixo). Conferir o link do dono: ausente.

## Pendências para correção
Nenhuma bloqueante. Sugestão para um próximo ciclo (P2):
1. **RN-25 / CA-07.6:** testes que provem que o link do dono não aparece no texto do convite (`inviteText`/`shareMeta`), na resposta de `/talentos` e nas mensagens do bot. A task T-07 prometia "do talento e do convite", e hoje só a lista da Home e o detalhe de quem não é membro estão testados.

## Scope creep
- `web/test/e2e/home.spec.ts` (commit `db9d956`, `test:`): troca a regex do título da Home por causa do "sáb". É defeito antigo de outra spec, registrado em "Descobertas" e feito num commit separado porque derrubava a CI. Aceitável, sem efeito no produto.
- `CLAUDE.md`: linha da ADR-09 em "Decidido". É coerente com o ADR aceito.
- `APP_ORIGIN` (config, compose, README) e `GET /me/discord-connection` no lugar de campos em `GET /me`: desvios do design registrados em "Descobertas". Atendem a RN-08 e a RN-22 sem mudar o esquema `User`.
- Nenhuma alteração toca o "Fora de escopo".

## Observações (não bloqueantes)
1. **CA-01.1 / RN-02 (log):** nenhum `slog` do caminho da conexão recebe o token do jogador, e os erros só levam método e caminho. Mesmo assim, não há teste que capture o log durante `POST /me/discord-connection` e confira a ausência do `access_token`, como já existe para o token do bot. Vale um teste.
2. **RNF-02, pior caso:** se a criação de um canal estoura o prazo de 10 s, o `undo` de `CreateRoom` roda com um prazo novo de até 5 s antes de responder. Em tese, a resposta ao dono pode passar de 10 s.
3. **Contagem do intervalo (RN-27):** o tempo que falta é calculado na renderização (`data.now`) e não é atualizado ao vivo. Depois de 10 min, o botão continua desabilitado até recarregar a página.
4. **Código curto do lobby (RN-05):** são 4 caracteres hexadecimais do UUID. Dois lobbies da mesma instância podem colidir (1 em 65 536). Só afeta o nome, porque o bot usa os IDs.
5. **Conta excluída:** a exclusão do Usuário apaga as candidaturas em cascata sem passar por `OnMemberLeft`, então a permissão de quem excluiu a conta continua nos canais até a sala ser apagada. A spec não cobre esse caso (os casos de borda só falam do dono).
6. **`PUT /lobbies/{id}` sem `discordLink` apaga o link** (semântica de PUT). O web sempre reenvia o link atual. Um cliente que omita o campo apaga o link sem querer.
7. **CA-04.4:** o teste cobre a troca do dono. A troca do membro aceito (`swaps.go`) não tem gancho, conferido no código, mas não tem teste próprio.
