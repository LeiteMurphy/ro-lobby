# Relatório de validação — lobbies (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** testes passou: backend unitários 14/14 pacotes · backend integração 14/14 pacotes (204 testes e subtestes, 0 falhas) · Vitest 193/193 (19 arquivos) · Playwright 39/39 (inclui os ponta a ponta de `login-discord`, `personagens`, `home-local` e `status`) · lint ok (golangci-lint 0 issues; Prettier e ESLint ok; svelte-check com 0 erros e 0 avisos) · build ok

Intervalo validado: `origin/main..HEAD` (13 commits, de `516685b` a `e5168ff`), na branch `feature/lobbies`.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 22 | 1 | 0 | 0 |
| Critérios (CA) | 29 | 2 | 0 | 0 |
| Não funcionais | 5 | 0 | 0 | 0 |

O motivo da reprovação é um só: o nome de uma instância do catálogo não bate com o bROWiki (RN-01 e CA-06.1, ambos de US-06, que é P1). O resto da feature está implementado, testado e passando.

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ⚠️ | backend/internal/catalog/instances.go:34-93; server/lobbies.go `ListInstances` (sem sessão) | TestInstances_CA06_1_GroupInstancesFromBROWiki, TestInstances_RN01_IDsAreUniqueKebabCase, TestInstances_CA06_1_PublicCatalog, TestGet_RN01_InstanceLeftCatalog | Conferi os 51 nomes com o HTML de https://browiki.org/wiki/Inst%C3%A2ncias (baixado em 2026-10-06): 50 batem exatamente e nenhuma "Solo" entrou (Edda do Quarto Crescente, Torneio de Magia, Salão de Ymir, Palácio das Mágoas e Invasão ao Aeroplano ficaram de fora). Exceção: `instances.go:92` traz "Sarah vs **Fenril**" (id `sarah-vs-fenril`), mas o bROWiki escreve "Sarah vs **Fenrir**" (`/wiki/Sarah_vs_Fenrir`). O teste `instances_test.go:45` grava o mesmo erro. |
| RN-02 | ✅ | catalog/instances.go:106-121; LobbyForm.svelte:66-69 (dois `optgroup`) | TestInstances_CA06_2_ChoiceOrder; lobbies.spec.ts "CA-06.2 / RN-02" | |
| RN-03 | ✅ | web/src/lib/home/catalog.ts `instanceArt` → capa e ícone padrão | home.spec.ts "RN-03 (lobbies)" | Os SVG antigos `/brand/inst-*.svg` ficaram sem uso (ver observações). |
| RN-04 | ✅ | server/lobbies.go `CreateLobby` (`currentUser`, 401); novo/+page.server.ts `toLogin` | TestLobbies_CA06_3_WritesRequireSession; lobbies.server.spec.ts "CA-02.3 / RN-04"; e2e CA-02.3 | |
| RN-05 | ✅ | lobbies.go:438-454 `checkStart` (fuso SP, hoje a hoje+13); web/src/lib/lobbies/time.ts | TestCreate_CA01_2_StartWindow, TestCreate_RN05_DayTurnInSaoPaulo; lobbies.spec.ts "borda de RN-05" | |
| RN-06 | ✅ | lobbies.go:457-467; migração 00004 (CHECK por função e soma); padrão 1/2/3 em novo/+page.server.ts | TestCreate_CA01_3_to_CA01_12_FieldErrors (CA-01.3); TestLobbies_RN06_RN07_RN09_RN19_ChecksRejectInvalidValues; lobbies.server.spec.ts "CA-01.1 / RN-06…" | |
| RN-07 | ✅ | lobbies.go:249-251, 347-349; CHECK `min_level >= instance_level`; LobbyForm `onInstanceChange` | CA-01.4 no TestCreate_CA01_3_to_CA01_12_FieldErrors; TestUpdate_CA04_2_CA04_3_Limits (150 < 160) | |
| RN-08 | ✅ | lobbies.go:266-278 (personagem próprio, nível, vaga da função) | CA-01.5, CA-01.6, CA-01.7 no TestCreate_CA01_3_to_CA01_12_FieldErrors; TestCreate_CA01_1_Defaults (ocupados) | |
| RN-09 | ✅ | lobbies.go:469-475 (trim + 250 runas); CHECK no banco | CA-01.11; TestUpdate_CA04_1 (trim) | |
| RN-10 | ✅ | lobbies.go:478-492; queries/lobbies.sql `HasScheduleConflict` (não cancelados, `IS DISTINCT FROM`) | TestCreate_CA01_8_ScheduleConflict, TestLobbies_RN10_ScheduleConflictWindow, TestCancel_CA05_1_CA05_2 (cancelado não conflita), TestUpdate_CA04_1 (não conflita consigo) | |
| RN-11 | ✅ | lobbies.go:279-285 | TestCreate_CA01_9_LimitOfFive, TestCreate_CA06_5_Concurrent | |
| RN-12 | ✅ | novo/+page.svelte (aviso + link /perfil); messages.ts `NO_CHARACTER_MESSAGE` | lobbies.spec.ts "CA-01.10 / RN-12"; e2e CA-02.3 / CA-01.10 | |
| RN-13 | ✅ | lobbies.go:518-526 (estado derivado); design §4 | TestList_CA02_2_OnlyOpen | |
| RN-14 | ✅ | queries/lobbies.sql `ListOpenLobbies` (`cancelled_at IS NULL AND starts_at > now`) | TestList_CA02_2_OnlyOpen, TestLobbies_RN13_RN14_ListOnlyOpen; e2e (cancelado sai da Home) | |
| RN-15 | ✅ | server `GetLobby` sem sessão, 404; routes/lobbies/[id]/+page.server.ts | TestGetLobby_CA03_2_NotFound, TestGet_CA03_2_NotFound; lobbies.spec.ts "CA-03.1 / RN-15"; e2e CA-03.1 | |
| RN-16 | ✅ | [id]/+page.svelte:86-99 | lobbies.spec.ts "CA-03.3 / RN-16"; e2e CA-03.3 | |
| RN-17 | ✅ | lobbies.go:318-379 (sem instância nem personagem no `UpdateInput`); editar/+page.server.ts | TestUpdate_CA04_1, TestUpdate_CA04_4_OtherOrNotOpen; lobbies.spec.ts "RN-17" | |
| RN-18 | ✅ | lobbies.go:350-355 | TestUpdate_CA04_2_CA04_3_Limits | |
| RN-19 | ✅ | lobbies.go:382-416; CHECK `cancel_reason` 10..250; CancelDialog.svelte | TestCancel_CA05_1_CA05_2, TestLobbies_RN19_CancelNeedsReason; lobbies.server.spec.ts CA-05.1/CA-05.2; e2e | |
| RN-20 | ✅ | lobbies.go:419-434 (`GetOwnLobbyForUpdate` por dono → ErrNotFound) | TestUpdate_CA04_4_OtherOrNotOpen; TestLobbyWrites_CA04_4_CA05_3_NotFoundAndNotOpen; lobbies.server.spec.ts "RN-20" | |
| RN-21 | ✅ | characters.go `checkNotInOpenLobby` em `Update` (só se a função muda) e `Delete`, dentro de `inTx`; server/characters.go 409 | TestOwnerOfOpenLobby_CA06_4_Locked, TestOwnerOfStartedOrCancelledLobby_RN21_Free; TestCharacters_CA06_4_InOpenLobbyIs409; perfil/page.server.spec.ts "CA-06.4"; e2e | |
| RN-22 | ✅ | routes/+page.server.ts (GET /lobbies dos 14 dias); lib/lobbies/toHome.ts | home.server.spec.ts "CA-02.1 / RN-22"; lobbies.spec.ts "CA-02.1"; e2e home e lobbies | |
| RN-23 | ✅ | TopBar.svelte (link /lobbies/novo); LobbyCard.svelte e FeaturedLobby.svelte (link do detalhe); LobbyForm com prévia | home.spec.ts "CA-02.3 / CA-02.4 / RN-23"; e2e CA-02.3, CA-02.4 | |

### Critérios
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | lobbies.go `Create`, `toLobby` | TestCreate_CA01_1_Defaults; lobbies.server.spec.ts (padrões e UTC); e2e CA-01.1 | |
| CA-01.2 | ✅ | `checkStart` | TestCreate_CA01_2_StartWindow (passado, agora, dia 14, vazio; dia 13 às 23:30 aceito) | |
| CA-01.3 | ✅ | `checkSlots` | subtestes "CA-01.3 total 0/13" | |
| CA-01.4 | ✅ | lobbies.go:249 | subteste "CA-01.4" (Torre da Constelação, 200) | |
| CA-01.5 | ✅ | lobbies.go:273 | subteste "CA-01.5" (`characterId/level_too_low`) | |
| CA-01.6 | ✅ | lobbies.go:276 | subteste "CA-01.6" | |
| CA-01.7 | ✅ | lobbies.go:266-268 | subtestes "CA-01.7"; TestCreateLobby_CA01_7_CA01_12_ValidationIs422 | |
| CA-01.8 | ✅ | `checkConflict`; messages.ts | TestCreate_CA01_8_ScheduleConflict (21:30 recusado, 22:00 aceito); e2e CA-01.8 (mensagem junto do campo) | |
| CA-01.9 | ✅ | lobbies.go:283; messages.ts `LOBBY_LIMIT_MESSAGE` | TestCreate_CA01_9_LimitOfFive; TestCreateLobby_CA01_9_LimitIs409; lobbies.server.spec.ts "CA-01.9" | |
| CA-01.10 | ✅ | novo/+page.svelte | lobbies.spec.ts "CA-01.10"; e2e | |
| CA-01.11 | ✅ | `checkNote` | subteste "CA-01.11" | |
| CA-01.12 | ✅ | lobbies.go:240-246 | subteste "CA-01.12"; TestCreateLobby_CA01_7_CA01_12_ValidationIs422 (422 com `instanceId`) | |
| CA-02.1 | ✅ | +page.server.ts, toHome.ts | home.server.spec.ts; e2e home "CA-02.1 (lobbies)" e lobbies "CA-02.1" | |
| CA-02.2 | ✅ | `ListOpenLobbies` | TestList_CA02_2_OnlyOpen (iniciado há 1 min e cancelado fora); e2e (cancelado sai da Home) | |
| CA-02.3 | ✅ | TopBar.svelte; novo/+page.server.ts `toLogin` | home.spec.ts; lobbies.server.spec.ts; e2e CA-02.3 | |
| CA-02.4 | ✅ | LobbyCard.svelte, FeaturedLobby.svelte | home.spec.ts; e2e CA-02.4 | |
| CA-03.1 | ✅ | [id]/+page.svelte | lobbies.spec.ts "CA-03.1"; e2e | |
| CA-03.2 | ✅ | [id]/+page.server.ts `error(404)`; routes/+error.svelte | lobbies.server.spec.ts "CA-03.2"; TestGetLobby_CA03_2_NotFound; e2e (`error-page`) | |
| CA-03.3 | ✅ | [id]/+page.svelte:86-99 | lobbies.spec.ts "CA-03.3"; e2e CA-03.3 (outra conta vê "Candidatar" desabilitado) | |
| CA-04.1 | ⚠️ | `Update`; editar/+page.server.ts | TestUpdate_CA04_1; TestLobbyWrites_CA04_1_CA05_1_PassData; e2e CA-04.1 (detalhe mostra 21:00 e 0 de 5) | O Then pede que "o detalhe **e a Home**" mostrem 21:00 e 5 vagas de Dano; o e2e só confere o detalhe e segue para o cancelamento. A Home usa a mesma API (coberta por CA-02.1), então o risco é baixo. US-04 é P2: não bloqueia. |
| CA-04.2 | ✅ | lobbies.go:353 | TestUpdate_CA04_2_CA04_3_Limits; lobbies.server.spec.ts "CA-04.2" | |
| CA-04.3 | ✅ | lobbies.go:350 | TestUpdate_CA04_2_CA04_3_Limits | |
| CA-04.4 | ✅ | `ownOpenLobby`; server 404/409 | TestUpdate_CA04_4_OtherOrNotOpen (confere que o lobby não mudou); TestLobbyWrites_CA04_4_CA05_3_NotFoundAndNotOpen | |
| CA-05.1 | ✅ | `Cancel`; [id]/+page.svelte (selo e motivo) | TestCancel_CA05_1_CA05_2; lobbies.spec.ts "CA-05.1"; e2e | |
| CA-05.2 | ✅ | lobbies.go:391-397 | TestCancel_CA05_1_CA05_2 ("não dá" → `reason/too_short`); lobbies.server.spec.ts; e2e | |
| CA-05.3 | ✅ | `ownOpenLobby` → 404 | TestUpdate_CA04_4_OtherOrNotOpen; TestLobbyWrites_CA04_4_CA05_3_NotFoundAndNotOpen | |
| CA-06.1 | ⚠️ | catalog/instances.go; `ListInstances` | TestInstances_CA06_1_PublicCatalog (51 itens, sem sessão), TestInstances_CA06_1_GroupInstancesFromBROWiki | Mesma divergência da RN-01: "Sarah vs Fenril" no lugar de "Sarah vs Fenrir". O resto do Then (sem "Solo"; nome, nível e retorno) está comprovado. |
| CA-06.2 | ✅ | LobbyForm.svelte `optgroup` | lobbies.spec.ts "CA-06.2"; TestInstances_CA06_2_ChoiceOrder | |
| CA-06.3 | ✅ | server/lobbies.go (401 nas três escritas) | TestLobbies_CA06_3_WritesRequireSession (sem token e com token desconhecido) | |
| CA-06.4 | ✅ | characters.go; server/characters.go; perfil/messages.ts | TestOwnerOfOpenLobby_CA06_4_Locked; TestCharacters_CA06_4_InOpenLobbyIs409; perfil/page.server.spec.ts; e2e (mensagem exata) | |
| CA-06.5 | ✅ | `inTx` com `LockUser` (D-05) | TestCreate_CA06_5_Concurrent (8 simultâneos → 5 e 3; 4 em conflito → 1) | |

### Não funcionais
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RNF-01 | ✅ | rótulos e `aria-describedby` no LobbyForm; CancelDialog com `aria-labelledby`/`describedby` | lobbies.spec.ts "RNF-01"; e2e "RNF-01: formulário e diálogo de cancelamento pelo teclado" | |
| RNF-02 | ✅ | — | e2e "RNF-02: criação e detalhe não carregam nada de fora do servidor"; home "RNF-02" | |
| RNF-03 | ✅ | — | Todos os testes novos citam IDs | Um ID de teste aponta para outra spec sem dizer (ver observações). |
| RNF-04 | ✅ | form.ts (só formato de dia e hora no web); todas as regras no serviço Go e nos CHECK | lobbies.server.spec.ts "RNF-04"; testes de integração do serviço | |
| RNF-05 | ✅ | routes/+page.server.ts | home.spec.ts (SSR); e2e "CA-02.6: o HTML do servidor já traz a Home" | |

### Design e rastreabilidade
- Contratos: `openapi.yaml` com as 6 rotas, os schemas `Instance`, `Slots`, `LobbyInput`, `LobbyUpdate` e `Lobby` e os códigos de erro novos; TestContract_CA06_3_LobbyRoutesAreDescribed passa. O código gerado está em dia (o build e o lint passam).
- Modelo de dados: a migração `00004_lobbies.sql` bate com o design §4, com um CHECK a mais, `min_level >= instance_level`, coerente com a RN-07.
- D-01 a D-09 respeitadas. D-04: `time/tzdata` embutido (lobbies.go:18). D-05: criar, editar, cancelar, excluir personagem e mudar a função rodam com o Usuário travado.
- Commits: os 13 seguem Conventional Commits e citam IDs. As tasks T-01 a T-10 estão marcadas `[x]` com commits que existem no intervalo.
- Features anteriores: os testes de `login-discord`, `personagens`, `home-local` e `status` (Go, Vitest e Playwright) passam todos.

## Pendências para correção
1. [RN-01, CA-06.1] O catálogo usa o nome "Sarah vs Fenril" (`backend/internal/catalog/instances.go:92`, id `sarah-vs-fenril`), mas a página de Instâncias do bROWiki escreve "Sarah vs Fenrir". A RN-01 manda seguir o bROWiki, e o Then da CA-06.1 pede "as instâncias de grupo do bROWiki". É preciso corrigir o nome e o id (`sarah-vs-fenrir`) no catálogo e no teste `backend/internal/catalog/instances_test.go:45`. Como o id fica gravado em `lobbies.instance_id`, convém decidir se lobbies de desenvolvimento com o id antigo precisam de ajuste (não há dados de produção).

## Scope creep
- Nenhum. As mudanças fora de `lobbies` têm ligação com a spec ou com as tasks:
  - `backend/cmd/migrate` e `internal/migrate/fresh.go`: banco próprio do e2e (T-10, registrado em Descobertas);
  - `web/src/lib/api/request.ts`: extraído de `characters/api.ts` (Descobertas);
  - `web/src/lib/auth/session.ts`: login com volta, RN-04 e RN-23;
  - `+error.svelte`: CA-03.2.

## Observações (não bloqueantes)
- [CA-04.1] Vale acrescentar ao e2e uma conferência na Home depois da edição (21:00 e 5 vagas de Dano), para cobrir o Then inteiro.
- [RNF-03] `TestLobbies_RN21_UnexpectedErrorIs500` (backend/internal/server/lobbies_test.go:224) cita "RN-21", que nesta spec é o personagem travado. O comentário explica que se refere à RN-21 da `personagens`, mas o nome do teste confunde. Sugestão: `TestLobbies_RN21Personagens_…` ou citar só o comentário.
- [RN-18, borda "personagem cai abaixo do nível mínimo"] `Update` recusa com `minLevel/above_owner` qualquer edição enquanto o nível mínimo atual estiver acima do nível novo do dono, mesmo que o dono só mude o horário. É uma leitura possível de "só a edição do nível mínimo passa a respeitar o nível novo", mas na prática obriga a baixar o nível mínimo para editar qualquer coisa. Vale confirmar com o usuário.
- [RN-03] Os arquivos `/brand/inst-*.svg` (arte das instâncias inventadas) ficaram sem uso no web. Podem ser removidos ou guardados para a futura arte por instância.
- [RN-02] O desempate por nome usa `cmp.Compare` em bytes. Com os nomes atuais a ordem sai certa, mas um nome com acento na primeira letra (ex.: "Ó…") iria para o fim. Um `collate` pt-BR evitaria isso no futuro.
- A criação começa pela instância mais alta que o principal alcança, não pela primeira do catálogo (ajuste registrado em Descobertas). Isso não contraria a spec.
