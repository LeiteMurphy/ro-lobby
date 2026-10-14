# Relatório de validação — servidor (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** testes passou (backend unitário ok, backend integração ok em todos os pacotes; web vitest 395/395; Playwright 60/60) · lint ok (gofmt sem saída, golangci-lint 0 issues, prettier + eslint ok, svelte-check 0 erros) · build ok (`npm run build`) · `go generate ./...` sem diferença no `git status`

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 18 | 2 | 0 | 0 |
| Critérios (CA) | 34 | 0 | 1 | 0 |
| Não funcionais | 2 | 1 | 0 | 0 |

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | backend/internal/catalog/servers.go; server/characters.go `ListServers`; openapi `/servers` sem `security` | TestServers_CA01_1_FreyaAndNidhoggInOrder; TestServers_CA01_1_CatalogWithoutSession; web catalog/servers.spec.ts | Web usa cópia fixa do catálogo, com teste de ordem e códigos |
| RN-02 | ✅ | characters/characters.go `Validate` (required / invalid no campo `server`); CHECK na 00010 | validate_test.go casos "CA-01.3 (servidor)"; TestValidate_RN19_ReportsAllFields; e2e perfil "CA-01.3 (servidor)" | |
| RN-03 | ✅ | migrations/00010_server.sql índice `characters_server_nick_key (server, lower(nick))`; `translate` mapeia para `taken` | TestServer_CA01_4/CA01_5/CA01_6; TestMigration00010_CA06_7 (nick repetido no servidor); e2e perfil CA-01.4/CA-01.6 | |
| RN-04 | ✅ | characters.go:221 (`current.Server != in.Server` aciona `checkNotInOpenLobby`); messages.ts `IN_OPEN_LOBBY_MESSAGE` | TestServer_CA01_8_LockedByActiveLink (dono, pendente, aceito, alvo de troca); TestLobbyServer_CA03_4_CA03_5; page.server.spec "CA-01.8 / RN-04" | |
| RN-05 | ✅ | CharacterCard.svelte (classe · servidor; nível em stats) | page.spec "CA-01.2 / RN-05"; e2e perfil `character-server` | |
| RN-06 | ✅ | lobbies.go:372 `Server: character.Server`; LobbyForm.svelte nota + prévia | TestLobbyServer_CA02_1; lobbies.spec "CA-02.2 / RN-06"; e2e servidor CA-02.2 | Sem campo de servidor no formulário |
| RN-07 | ✅ | `UpdateLobby` e troca de dono não tocam `lobbies.server`; coluna própria no lobby | TestLobbyServer_CA02_3_NeverChanges (edição, troca do dono, exclusão após início) | |
| RN-08 | ✅ | LobbyCard.svelte, FeaturedLobby.svelte, lobbies/[id]/+page.svelte | server.spec "CA-02.4 / RN-08"; lobbies.spec "CA-02.4"; e2e servidor (card, destaque, detalhe) | |
| RN-09 | ✅ | applications.go:202 `CodeWrongServer` | TestLobbyServer_CA03_1_CA03_2_Apply | Validado no serviço |
| RN-10 | ✅ | swaps.go:159 (pedido), `checkSwap` (troca do dono e aceite) | TestLobbyServer_CA03_3_OwnerSwap; TestLobbyServer_CA03_4_CA03_5_MemberSwap | Os três caminhos (dono, pedido, aceite) passam pela checagem |
| RN-11 | ✅ | applications/detail.ts `otherServer` em `eligibility` e `swapEligibility` | applications.spec, membership.spec (modos request e owner); e2e servidor CA-03.6 | |
| RN-12 | ✅ | home/filters.ts `server` em `Filters`/`NO_FILTERS`/`matches`; FilterPanel.svelte | server.spec CA-04.1/04.2; e2e servidor CA-04.1 a CA-04.5 | |
| RN-13 | ✅ | filters.ts `serverCounts`; +page.svelte `serverCounts(allDayLobbies)` | server.spec CA-04.1 (contagens); e2e (Todos 3, Freya 2, Nidhogg 1) | |
| RN-14 | ⚠️ | +page.svelte `scoped = byServer(...)` alimenta `buildDays`, `roleCounts`, `timeRangeCounts`, `featuredLobby`; `resetFilters` volta a `NO_FILTERS` | server.spec CA-04.3 (contagem de Tank, destaque, vazio); e2e (destaque, Limpar) | O recorte do **seletor de dias** pelo servidor não tem teste (nem unitário, nem e2e) |
| RN-15 | ✅ | 00010_server.sql (default `freya`, `server_to_confirm` true para os antigos e default false depois) | TestMigration00010_CA06_7; TestConfirmServer_CA06_4 | A falha com nicks duplicados não é testada, mas o índice antigo torna o estado impossível; o CREATE UNIQUE INDEX falharia dentro da transação da migração |
| RN-16 | ✅ | queries/characters.sql `UpdateCharacter` zera `server_to_confirm`, `ConfirmCharacterServer`; perfil +page.svelte aviso; action `confirm` | server_confirm_integration_test.go (CA-06.2/06.3/06.5/06.6); page.spec CA-06.1/06.3; page.server.spec CA-06.2/06.6 | Sem e2e do aviso (registrado em Descobertas; fica no UAT) |
| RN-17 | ✅ | queries/talents.sql `c.server = @server`; affinity.go `ForLobby` usa `l.Server`; catalog.go `Count` usa `character.Server` | TestServer_CA05_1; TestServer_CA05_2; TestServer_RN17_LobbyInNidhogg | |
| RN-18 | ⚠️ | talents.sql filtro opcional; talentos/+page.server.ts `servidor`; +page.svelte select | TestServer_CA05_3_CatalogFilter; talentos.spec; e2e servidor CA-05.3 | O filtro fica na query string; recarregar mantém a escolha (ver CA-05.3) |
| RN-19 | ✅ | TalentCard.svelte `talent-server`; API `server` em `Talent` | talents_test (CA-05.4); talentos.spec CA-05.4; e2e servidor | |
| RN-20 | ✅ | lobbies/share.ts `inviteText` e `shareMeta.title` | share.spec "CA-02.5 / RN-20"; lobbies.spec og:title; e2e compartilhar | |

### Critérios
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | server/characters.go `ListServers` | TestServers_CA01_1_CatalogWithoutSession | |
| CA-01.2 | ✅ | CharacterDialog.svelte rádio Servidor; CharacterCard | TestServer_CA01_2_SavedWithServer; page.spec; e2e perfil "CA-01.2 / CA-01.4 / CA-01.7 (servidor)" | |
| CA-01.3 | ✅ | `Validate`; messages.ts `server.fallback`; diálogo com `novalidate` | validate_test; page.spec; e2e perfil "CA-01.3 (servidor)" (passa pela API e mostra "Escolha o servidor") | Não há teste HTTP isolado do 422 com `server`, mas o e2e prova o caminho real |
| CA-01.4 | ✅ | índice (server, nick) | TestServer_CA01_4_SameNickOtherServer; e2e perfil | |
| CA-01.5 | ✅ | `translate` → `taken`; mensagem "nesse servidor" | TestServer_CA01_5_SameNickSameServer (criar e editar); characters.spec; page.server.spec; e2e perfil CA-02.3 | |
| CA-01.6 | ✅ | idem | TestServer_CA01_6_MoveCollides (confere que fica em Freya); e2e perfil | |
| CA-01.7 | ✅ | `UpdateCharacter` com `server` | TestServer_CA01_7_Edit; e2e perfil | |
| CA-01.8 | ✅ | `checkNotInOpenLobby` | TestServer_CA01_8_LockedByActiveLink; page.server.spec | |
| CA-02.1 | ✅ | lobbies.go `Create` | TestLobbyServer_CA02_1_InheritsOwnerServer (detalhe e lista) | |
| CA-02.2 | ✅ | LobbyForm.svelte nota `lobby-server` e `server` na prévia do LobbyCard | lobbies.spec "CA-02.2 / RN-06"; e2e servidor CA-02.2 | A prévia do card só é provada indiretamente (o LobbyCard mostra o servidor, server.spec CA-02.4); os testes conferem a nota |
| CA-02.3 | ✅ | — | TestLobbyServer_CA02_3_NeverChanges | |
| CA-02.4 | ✅ | card, destaque, detalhe | server.spec; lobbies.spec; e2e servidor | |
| CA-02.5 | ✅ | share.ts | share.spec; lobbies.spec og:title; e2e compartilhar | |
| CA-03.1 | ✅ | applications.go | TestLobbyServer_CA03_1_CA03_2_Apply (409 `wrong_server`, `pendingCount` 0) | |
| CA-03.2 | ✅ | — | idem (201, `pending`); e2e servidor | |
| CA-03.3 | ✅ | `checkSwap` na troca do dono | TestLobbyServer_CA03_3_OwnerSwap (personagem do dono não muda) | |
| CA-03.4 | ✅ | swaps.go:159 | TestLobbyServer_CA03_4_CA03_5_MemberSwap (nenhum pedido) | |
| CA-03.5 | ✅ | `checkSwap` no aceite + trava RN-04 | idem (edição do alvo recusada com 409; aceite 200) | |
| CA-03.6 | ✅ | detail.ts | applications.spec; membership.spec; e2e servidor | |
| CA-04.1 | ✅ | filters.ts; FilterPanel | server.spec; e2e servidor | |
| CA-04.2 | ✅ | `matches` | server.spec; e2e servidor | |
| CA-04.3 | ✅ | `byServer` + `applyFilters` + `emptyState` | server.spec CA-04.3 (vazio de filtros, destaque de Nidhogg); e2e (destaque) | |
| CA-04.4 | ✅ | `resetFilters` | e2e servidor (Limpar volta a Todos) | |
| CA-04.5 | ✅ | `$state(NO_FILTERS)`, sem storage | e2e servidor (reload volta a Todos) | |
| CA-05.1 | ✅ | talents.sql | TestServer_CA05_1_AffinityOnlySameServer | |
| CA-05.2 | ✅ | catalog.go `Count` | TestServer_CA05_2_CountUsesOwnerServer (1 com Nidhogg, 1 com Freya) | Na web a prévia já reconsulta pelo `characterId` (Descobertas); e2e talentos CA-03.1 cobre a prévia em si |
| CA-05.3 | ❌ | talentos/+page.server.ts lê `servidor` da URL | e2e servidor testa "abrir /talentos de novo", não "recarregar" | O Then diz "recarregar a página volta o filtro para Todos". Recarregar `/talentos?servidor=nidhogg` mantém Nidhogg. Desvio registrado em Descobertas como **proposta aguardando o usuário**; a spec aprovada não foi alterada |
| CA-05.4 | ✅ | TalentCard; API | talents_test; talentos.spec; e2e servidor; TestServer_CA05_1 (item traz o servidor) | |
| CA-06.1 | ✅ | perfil +page.svelte | page.spec CA-06.1 (dois a conferir, "Está certo" e "Editar", novo fora) | Sem e2e (Descobertas) |
| CA-06.2 | ✅ | `ConfirmCharacterServer` | TestConfirmServer_CA06_2_ConfirmsOnlyThatOne; page.server.spec | |
| CA-06.3 | ✅ | `UpdateCharacter` zera a flag | TestConfirmServer_CA06_3_EditConfirms; page.spec (sem aviso) | |
| CA-06.4 | ✅ | default false após a migração | TestConfirmServer_CA06_4_NewIsConfirmed; migração | |
| CA-06.5 | ✅ | nenhuma regra lê `server_to_confirm` | TestConfirmServer_CA06_5_WorksNormally | |
| CA-06.6 | ✅ | `ConfirmServer` → `ErrNotFound` com `n == 0` | TestConfirmServer_CA06_6_OnlyOwner (404 e continua a conferir) | |
| CA-06.7 | ✅ | 00010_server.sql | TestMigration00010_CA06_7_ExistingDataGoesToFreya (up, CHECK, índice, down com os dados) | |

### Não funcionais
| ID | Veredito | Evidência | Observação |
|---|---|---|---|
| RNF-01 | ✅ | `characters_server_nick_key` em (server, lower(nick)) | Testado na migração e nos testes de serviço |
| RNF-02 | ✅ | CHECK `server IN ('freya','nidhogg')` em `characters` e `lobbies` | Migração testa "servidor fora do catálogo" no personagem; no lobby só por leitura do SQL |
| RNF-03 | ⚠️ | Testes novos citam os IDs | Testes antigos cujo comportamento mudou não citam o critério desta spec: e2e `perfil.spec.ts` "CA-02.3 / CA-02.9 / CA-01.3" (mensagem passou a "nesse servidor") e `TestCreate_CA02_3_NickTakenByAnotherUser` / `TestUpdate_RN05_NickOfAnotherCharacter` (citam a RN-05 antiga, substituída pela RN-03) |

## Pendências para correção
1. [CA-05.3 / RN-18] Recarregar `/talentos` com o filtro "Servidor" escolhido mantém a escolha (fica em `?servidor=`), e o Then exige que volte para Todos. Há dois caminhos: o usuário aprova o ajuste de spec proposto em Descobertas ("abrir `/talentos` de novo começa em Todos; o filtro fica no link, como os outros"), e aí a spec e o e2e passam a dizer isso; ou o código passa a não manter o servidor no recarregamento. Enquanto a spec aprovada disser "recarregar", o item fica ❌.
2. [RN-14] Falta teste do recorte do seletor de dias pelo servidor escolhido (`buildDays(data.today, scoped)` em `web/src/routes/+page.svelte`). Por exemplo: um dia só com lobby de Freya deve aparecer sem grupos quando o filtro está em Nidhogg.
3. [RNF-03] Os testes antigos de nick único que mudaram devem citar também o critério desta spec (CA-01.5): `web/test/e2e/perfil.spec.ts` (teste "CA-02.3 / CA-02.9 / CA-01.3"), `backend/internal/characters/characters_integration_test.go` (`TestCreate_CA02_3_NickTakenByAnotherUser`, `TestUpdate_RN05_NickOfAnotherCharacter`). Não bloqueia sozinho.

## Scope creep
- Nenhum. Todas as alterações do diff se ligam a IDs da spec. Os itens de "Fora de escopo" não foram tocados: o filtro não tem memória, "Minhas candidaturas" não mostra o servidor, o lobby não tem campo de servidor e não há troca de servidor de lobby.

## Observações (não bloqueantes)
- O campo `Commit:` das tasks T-01 a T-08 em `tasks.md` está vazio, embora as tasks estejam marcadas `[x]`. Os commits existem e citam os IDs (3b60359, 43a7bed, c845330, 167a7f4, b7af046, 1d1fb66, e294e15, be355dc).
- O parâmetro `server` de `GET /talents` está como `type: string` no openapi, e não como `$ref: ServerId`. Um valor fora do catálogo volta 422 pelo serviço (testado), o que é coerente com os outros filtros.
- O aviso "Confira o servidor" (CA-06.1 a CA-06.3) não tem e2e porque só existe personagem a conferir via migração. Vale incluir no roteiro de UAT, como já registrado.
- O texto do subtítulo da Home passou a usar o total do dia sem o recorte de servidor (`allDayLobbies`), o que deixa o "N de M" coerente com o estado vazio de filtros. Não é desvio, mas convém conferir visualmente.
