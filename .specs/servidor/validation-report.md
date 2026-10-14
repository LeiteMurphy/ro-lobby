# Relatório de validação — servidor (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend unitário ok; backend integração ok em todos os pacotes, rodado com `-count=1`; web vitest 396/396; Playwright 60/60) · lint ok (gofmt sem saída, golangci-lint 0 issues, prettier + eslint ok, svelte-check 0 erros e 0 avisos) · build ok (`npm run build`, saída 0) · `go generate ./...` sem diferença no `git status`

## Pendências do ciclo 1
| # | Item | Situação |
|---|---|---|
| 1 | CA-05.3 / RN-18: "recarregar" volta a Todos | Resolvida. A spec foi revisada (commit 7f4381e; Descobertas registra a aprovação do usuário em 2026-10-14). Agora diz que o filtro fica no link e que abrir `/talentos` de novo começa em Todos. O e2e `servidor.spec.ts` "CA-05.3 / CA-05.4" faz exatamente isso (`page.goto('/talentos')` → `Servidor` com valor `''`) |
| 2 | RN-14: seletor de dias sem teste | Resolvida. `web/src/lib/home/server.spec.ts` "RN-14: o seletor de dias conta só os lobbies do servidor escolhido" usa `buildDays` sobre `byServer`, a mesma composição de `web/src/routes/+page.svelte:47-48` (`scoped` → `buildDays(data.today, scoped)`), e confere as contagens por dia em Todos, Freya e Nidhogg |
| 3 | RNF-03: testes antigos de nick sem o critério novo | Resolvida. `TestCreate_CA02_3_CA01_5_NickTakenByAnotherUser`, `TestUpdate_RN05_CA01_5_NickOfAnotherCharacter` e o e2e "CA-02.3 / CA-02.9 / CA-01.3, CA-01.5 (servidor)" citam o CA-01.5 |

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 20 | 0 | 0 | 0 |
| Critérios (CA) | 35 | 0 | 0 | 0 |
| Não funcionais | 3 | 0 | 0 | 0 |

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | backend/internal/catalog/servers.go; server/characters.go `ListServers`; openapi `/servers` sem `security` | TestServers_CA01_1_FreyaAndNidhoggInOrder; TestServers_CA01_1_CatalogWithoutSession; web catalog/servers.spec.ts | |
| RN-02 | ✅ | characters/characters.go:383-386 (`required` / `invalid` no campo `server`); CHECK na 00010 | validate_test.go "CA-01.3 (servidor)"; e2e perfil "CA-01.3 (servidor)" | Validado no serviço |
| RN-03 | ✅ | 00010_server.sql `characters_server_nick_key (server, lower(nick))`; `translate` → `taken` | TestServer_CA01_4/CA01_5/CA01_6; TestMigration00010_CA06_7; e2e perfil | |
| RN-04 | ✅ | characters.go:221 (mudança de `Server` aciona `checkNotInOpenLobby`); messages.ts `IN_OPEN_LOBBY_MESSAGE` | TestServer_CA01_8_LockedByActiveLink; TestLobbyServer_CA03_4_CA03_5_MemberSwap; page.server.spec "CA-01.8 / RN-04" | |
| RN-05 | ✅ | CharacterCard.svelte | page.spec "CA-01.2 / RN-05"; e2e perfil `character-server` | |
| RN-06 | ✅ | lobbies.go `Create` grava `character.Server`; LobbyForm.svelte nota e prévia | TestLobbyServer_CA02_1_InheritsOwnerServer; lobbies.spec "CA-02.2 / RN-06"; e2e servidor CA-02.2 | Sem campo de servidor no formulário |
| RN-07 | ✅ | coluna `lobbies.server` própria; edição e troca do dono não a tocam | TestLobbyServer_CA02_3_NeverChanges | |
| RN-08 | ✅ | LobbyCard.svelte, FeaturedLobby.svelte, lobbies/[id]/+page.svelte | server.spec "CA-02.4 / RN-08"; lobbies.spec; e2e servidor | |
| RN-09 | ✅ | applications.go:203 `CodeWrongServer` | TestLobbyServer_CA03_1_CA03_2_Apply | |
| RN-10 | ✅ | swaps.go:160 (pedido) e swaps.go:328 (`checkSwap`: troca do dono e aceite) | TestLobbyServer_CA03_3_OwnerSwap; TestLobbyServer_CA03_4_CA03_5_MemberSwap | |
| RN-11 | ✅ | applications/detail.ts `otherServer` | applications.spec; membership.spec; e2e servidor CA-03.6 | |
| RN-12 | ✅ | home/filters.ts; FilterPanel.svelte | server.spec CA-04.1/04.2; e2e servidor CA-04.1 a CA-04.5 | |
| RN-13 | ✅ | filters.ts `serverCounts`; +page.svelte:56 `serverCounts(allDayLobbies)` | server.spec CA-04.1; e2e servidor | |
| RN-14 | ✅ | +page.svelte:47-51 (`scoped` alimenta dias, lista, contagens e destaque); `resetFilters` | server.spec "RN-14: o seletor de dias…", CA-04.3 e borda de estado vazio; e2e (destaque, Limpar) | Pendência 2 do ciclo 1 resolvida |
| RN-15 | ✅ | 00010_server.sql (default `freya`; `server_to_confirm` true nos antigos e default false depois; índice criado dentro da migração) | TestMigration00010_CA06_7; TestConfirmServer_CA06_4_NewIsConfirmed | A falha com nicks duplicados não tem teste; o índice antigo torna o estado impossível |
| RN-16 | ✅ | queries/characters.sql `UpdateCharacter` zera a flag; `ConfirmCharacterServer`; perfil +page.svelte aviso; action `confirm` | server_confirm_integration_test.go (CA-06.2/06.3/06.5/06.6); page.spec; page.server.spec | |
| RN-17 | ✅ | talents.sql:45 `c.server = @server`; affinity.go:93; catalog.go:151 | TestServer_CA05_1; TestServer_CA05_2; TestServer_RN17_LobbyInNidhogg | |
| RN-18 | ✅ | talents.sql:80 filtro opcional; catalog.go:59-63 valida; talentos/+page.server.ts lê `servidor` do link | TestServer_CA05_3_CatalogFilter; talentos.spec; e2e servidor "CA-05.3 / CA-05.4" | Comportamento bate com a spec revisada |
| RN-19 | ✅ | TalentCard.svelte `talent-server`; `server` em `Talent` | talents_test; talentos.spec; e2e servidor | |
| RN-20 | ✅ | lobbies/share.ts `inviteText` e `shareMeta.title` | share.spec "CA-02.5 / RN-20"; lobbies.spec og:title; e2e compartilhar | |

### Critérios
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | `ListServers` | TestServers_CA01_1_CatalogWithoutSession | |
| CA-01.2 | ✅ | CharacterDialog.svelte; CharacterCard | TestServer_CA01_2_SavedWithServer; page.spec; e2e perfil | |
| CA-01.3 | ✅ | `Validate`; messages.ts | validate_test; page.spec; e2e perfil "CA-01.3 (servidor)" | |
| CA-01.4 | ✅ | índice (server, nick) | TestServer_CA01_4_SameNickOtherServer; e2e perfil | |
| CA-01.5 | ✅ | `translate` → `taken`; "Esse nick já está em uso nesse servidor" | TestServer_CA01_5_SameNickSameServer; TestCreate_CA02_3_CA01_5; TestUpdate_RN05_CA01_5; characters.spec; e2e perfil | |
| CA-01.6 | ✅ | idem | TestServer_CA01_6_MoveCollides | |
| CA-01.7 | ✅ | `UpdateCharacter` | TestServer_CA01_7_Edit; e2e perfil | |
| CA-01.8 | ✅ | `checkNotInOpenLobby` | TestServer_CA01_8_LockedByActiveLink; page.server.spec | |
| CA-02.1 | ✅ | lobbies.go `Create` | TestLobbyServer_CA02_1_InheritsOwnerServer | |
| CA-02.2 | ✅ | LobbyForm.svelte | lobbies.spec "CA-02.2 / RN-06"; e2e servidor CA-02.2 | |
| CA-02.3 | ✅ | — | TestLobbyServer_CA02_3_NeverChanges | |
| CA-02.4 | ✅ | card, destaque, detalhe | server.spec; lobbies.spec; e2e servidor | |
| CA-02.5 | ✅ | share.ts | share.spec; lobbies.spec; e2e compartilhar | |
| CA-03.1 | ✅ | applications.go:203 | TestLobbyServer_CA03_1_CA03_2_Apply | |
| CA-03.2 | ✅ | — | idem; e2e servidor | |
| CA-03.3 | ✅ | `checkSwap` | TestLobbyServer_CA03_3_OwnerSwap | |
| CA-03.4 | ✅ | swaps.go:160 | TestLobbyServer_CA03_4_CA03_5_MemberSwap | |
| CA-03.5 | ✅ | `checkSwap` no aceite + trava RN-04 | idem | |
| CA-03.6 | ✅ | detail.ts | applications.spec; membership.spec; e2e servidor | |
| CA-04.1 | ✅ | filters.ts; FilterPanel | server.spec; e2e servidor | |
| CA-04.2 | ✅ | `matches` | server.spec; e2e servidor | |
| CA-04.3 | ✅ | `byServer` + `applyFilters` + `emptyState` | server.spec CA-04.3; e2e servidor | |
| CA-04.4 | ✅ | `resetFilters` | e2e servidor | |
| CA-04.5 | ✅ | `$state(NO_FILTERS)`, sem storage | e2e servidor (recarregar volta a Todos) | |
| CA-05.1 | ✅ | talents.sql | TestServer_CA05_1_AffinityOnlySameServer | |
| CA-05.2 | ✅ | catalog.go `Count` | TestServer_CA05_2_CountUsesOwnerServer | |
| CA-05.3 | ✅ | talentos/+page.server.ts; talentos/+page.svelte | TestServer_CA05_3_CatalogFilter; talentos.spec; e2e servidor (escolhe Nidhogg, só o de Nidhogg aparece; abrir `/talentos` de novo volta a Todos) | Pendência 1 do ciclo 1 resolvida pela revisão da spec |
| CA-05.4 | ✅ | TalentCard; API | talents_test; talentos.spec; e2e servidor | |
| CA-06.1 | ✅ | perfil +page.svelte | page.spec CA-06.1 | Sem e2e (Descobertas; fica no UAT) |
| CA-06.2 | ✅ | `ConfirmCharacterServer` | TestConfirmServer_CA06_2_ConfirmsOnlyThatOne; page.server.spec | |
| CA-06.3 | ✅ | `UpdateCharacter` zera a flag | TestConfirmServer_CA06_3_EditConfirms; page.spec | |
| CA-06.4 | ✅ | default false após a migração | TestConfirmServer_CA06_4_NewIsConfirmed | |
| CA-06.5 | ✅ | nenhuma regra lê `server_to_confirm` | TestConfirmServer_CA06_5_WorksNormally | |
| CA-06.6 | ✅ | `ConfirmServer` → `ErrNotFound` | TestConfirmServer_CA06_6_OnlyOwner | |
| CA-06.7 | ✅ | 00010_server.sql | TestMigration00010_CA06_7_ExistingDataGoesToFreya | |

### Não funcionais
| ID | Veredito | Evidência | Observação |
|---|---|---|---|
| RNF-01 | ✅ | `characters_server_nick_key` em (server, lower(nick)) | Testado na migração e nos testes de serviço |
| RNF-02 | ✅ | CHECK `server IN ('freya','nidhogg')` em `characters` e `lobbies` | O CHECK do lobby é conferido só pela leitura do SQL |
| RNF-03 | ✅ | Testes novos citam os IDs; os antigos de nick único citam o CA-01.5 | Pendência 3 do ciclo 1 resolvida |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. Os commits após o ciclo 1 (7f4381e, bd448ba, 450dd48) só mexem na spec, no tasks.md, em testes e no relatório. Os itens de "Fora de escopo" seguem intocados: o filtro da Home não tem memória, "Minhas candidaturas" não mostra o servidor, o lobby não tem campo de servidor e não há troca de servidor de lobby.

## Observações (não bloqueantes)
- O campo `Commit:` das tasks T-01 a T-08 em `tasks.md` continua vazio, embora as tasks estejam marcadas `[x]`. Os commits existem e citam os IDs (3b60359, 43a7bed, c845330, 167a7f4, b7af046, 1d1fb66, e294e15, be355dc).
- O parâmetro `server` de `GET /talents` segue como `type: string` no openapi, sem `$ref: ServerId`. Um valor fora do catálogo volta 422 pelo serviço (catalog.go:59-61), como nos outros filtros.
- O aviso "Confira o servidor" (CA-06.1 a CA-06.3) não tem e2e, porque só a migração cria personagem a conferir. Vale incluir no roteiro de UAT.
- O commit bd448ba cita CA-05.3, mas não altera nada ligado a ele. Isso não afeta a rastreabilidade, porque o ajuste de CA-05.3 está no 7f4381e.
