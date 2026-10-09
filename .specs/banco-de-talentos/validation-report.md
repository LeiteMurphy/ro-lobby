# Relatório de validação — banco-de-talentos (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes backend passou (todos os pacotes `ok`, `-tags=integration -count=1`) · vitest passou (319/319, 28 arquivos) · e2e passou (50/50) · lint ok (`golangci-lint` 0 issues; prettier + eslint ok) · check ok (svelte-check 0 erros, 0 avisos) · build ok

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 12 | 0 | 0 | 0 |
| Critérios (CA) | 13 | 0 | 0 | 0 |
| Não funcionais | 4 | 1 | 0 | 0 |

As duas pendências do ciclo 1 (RN-04) foram resolvidas no commit `5543a67`. O único ⚠️
restante é a RNF-02, não bloqueante, igual ao ciclo 1.

## Pendências do ciclo 1
| # | Situação | Evidência |
|---|---|---|
| 1. Rótulo "Qualquer instância" para quem ficou sem instância do catálogo | Resolvida | `web/src/lib/talents/format.ts:50` `instancesLabel` segue só o `anyInstance`; lista vazia vira "Nenhuma instância do catálogo". `TalentCard.svelte:37` usa a função. Teste `web/src/lib/talents/talents.spec.ts` "RN-04: ... sem nenhuma no catálogo, diz isso" confere `instancesLabel(false, [])` |
| 2. Falta teste da instância que sai do catálogo | Resolvida | `backend/internal/talents/removed_integration_test.go` `TestRemovedInstance_RN04` grava ids fora do catálogo direto com `q.UpsertAvailability` e confere: catálogo mantém "Misto" (só com o Templo) e "Orfao" (lista vazia, `AnyInstance` falso); afinidade traz só "Misto"; perfil (`chars.List`) não traz o id removido. Passou (execução com `-v`) |

Também entrou `TestMidnight_RN03_SaturdayToSunday` (sábado 23:00–01:00: domingo 00:30 com
afinidade, domingo 23:30 sem), que fecha a observação do ciclo 1 sobre a virada
sábado→domingo.

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `backend/queries/talents.sql` (UpsertAvailability religa com `enabled = true`; DisableAvailability só muda `enabled`); `talents.go` SetAvailability | `TestSetAvailability_CA01_1_CA01_4`, `TestSetAvailability_RN01_DisableNeverEnabled`, `TestAvailability_RN01_UpsertAndDisable`, `TestCatalog_CA04_3_DisabledAndInvalid`; e2e desligar e reabrir | |
| RN-02 | ✅ | `talents.go` validate; CHECK `days BETWEEN 1 AND 127` | `TestSetAvailability_CA01_3_FieldErrors`; `TestAvailability_RN02_RN03_RN04_ChecksRejectInvalidValues` | |
| RN-03 | ✅ | `talents.go` ParseClock; `talents.sql` Affinity/Catalog com virada e fim excluído; CHECK | `TestSetAvailability_CA01_3_FieldErrors`; `TestForLobby_CA02_2_OutOfAffinity`; `TestCatalog_CA01_2_Midnight`; `TestMidnight_RN03_SaturdayToSunday` (novo) | |
| RN-04 | ✅ | `talents.go:265` `inCatalog` em `FromRow`; `affinity.go:118` toTalent filtra pelo catálogo; `talents.sql:37` afinidade só por `any_instance` ou id do lobby; `format.ts:50` rótulo | `TestSetAvailability_CA01_3_FieldErrors`; `TestRemovedInstance_RN04` (novo); vitest `RN-04` (novo caso) | Caso de borda da seção 6 coberto: sai da afinidade, continua no banco |
| RN-05 | ✅ | SetAvailability (LockUser + GetOwnCharacter → ErrNotFound); FK `ON DELETE CASCADE` | `TestSetAvailability_CA01_5_OtherUser`; `TestAvailability_RN05_CascadeOnCharacterDelete`; `TestSetAvailability_CA01_3_CA01_5_Errors` | |
| RN-06 | ✅ | `talents.sql` Affinity (cond. 1 a 5); `affinity.go` ForLobby (funções com vaga) e `wallClock` em São Paulo | `TestForLobby_CA02_1`, `_CA02_2_OutOfAffinity`, `_CA02_3_Busy`, `_CA02_4_OwnerAndOpen` | |
| RN-07 | ✅ | `talents.sql` `c.user_id <> exclude_user_id` | `TestForLobby_CA02_3_Busy` | |
| RN-08 | ✅ | `ORDER BY c.level DESC, lower(c.nick), c.seq` | `TestForLobby_CA02_1`; `TestForLobby_CA02_3_Busy`; `TestCatalog_CA01_1_CA04_1_Filters` | |
| RN-09 | ✅ | ForLobby (ErrNotFound/ErrNotOpen); `server/talents.go` 401/404/409; `lobbies/[id]/+page.server.ts` só com `isOwner && status === 'open'` | `TestForLobby_CA02_4_OwnerAndOpen`; `TestListLobbyTalents_CA02_4_Errors`; vitest `CA-02.4 / RN-09`; e2e visitante | |
| RN-10 | ✅ | `catalog.go` Count; `routes/lobbies/novo/disponiveis/+server.ts`; `LobbyForm.svelte` | `TestCount_CA03_1`; `TestCountTalents_CA03_1`; `disponiveis.spec.ts`; e2e CA-03.1 | |
| RN-11 | ✅ | `talents.sql` Catalog (filtros opcionais com "E"); `catalog.go`; `routes/talentos/+page.server.ts` | `TestCatalog_CA01_1_CA04_1_Filters`, `TestCatalog_CA01_2_Midnight`, `TestCatalog_CA04_3_DisabledAndInvalid`; `talentos.spec.ts`; e2e | |
| RN-12 | ✅ | `TalentCard.svelte`; `server/talents.go` toAPITalents omite `discordUsername` sem sessão | `TestListTalents_CA04_1_CA04_2`; vitest `CA-01.1 / RN-12`, `CA-04.2 / RNF-05`; e2e | O defeito de exibição de instâncias do ciclo 1 foi corrigido |
| CA-01.1 | ✅ | SetAvailability + Catalog + AvailabilityDialog | `TestSetAvailability_CA01_1_CA01_4`; `TestCatalog_CA01_1_CA04_1_Filters`; e2e | |
| CA-01.2 | ✅ | Catalog (virada) | `TestCatalog_CA01_2_Midnight`; e2e `dia=6&hora=01:00` / `22:00` | |
| CA-01.3 | ✅ | validate + 422 + `talentFieldMessages` | `TestSetAvailability_CA01_3_FieldErrors` (nada gravado); `TestSetAvailability_CA01_3_CA01_5_Errors`; vitest; e2e | |
| CA-01.4 | ✅ | DisableAvailability; `characters.List` com `availability` | `TestSetAvailability_CA01_1_CA01_4`; vitest `CA-01.4`; e2e | |
| CA-01.5 | ✅ | GetOwnCharacter → 404 | `TestSetAvailability_CA01_5_OtherUser`; rota 404; vitest perfil | |
| CA-02.1 | ✅ | ForLobby + painel | `TestForLobby_CA02_1`; `TestListLobbyTalents_CA02_1`; vitest; e2e | |
| CA-02.2 | ✅ | Affinity | `TestForLobby_CA02_2_OutOfAffinity` | |
| CA-02.3 | ✅ | Affinity cond. 5 e RN-07 | `TestForLobby_CA02_3_Busy` | |
| CA-02.4 | ✅ | ForLobby + load/página | `TestForLobby_CA02_4_OwnerAndOpen`; `TestListLobbyTalents_CA02_4_Errors`; vitest; e2e | |
| CA-03.1 | ✅ | Count + endpoint web + LobbyForm | `TestCount_CA03_1` (3 às 20:00, 0 às 17:00); e2e | |
| CA-04.1 | ✅ | Catalog | `TestCatalog_CA01_1_CA04_1_Filters`; e2e | |
| CA-04.2 | ✅ | toAPITalents + TalentCard | `TestListTalents_CA04_1_CA04_2`; vitest; e2e (HTML do visitante) | |
| CA-04.3 | ✅ | `WHERE a.enabled` | `TestCatalog_CA04_3_DisabledAndInvalid`; e2e | |
| RNF-01 | ✅ | Migração (minutos e bitmask); `wallClock` com `lobbies.Location` | Testes da afinidade com `at()` em São Paulo | |
| RNF-02 | ⚠️ | Regras na API (serviço + CHECK) | Testes de serviço e de banco | O web não avisa antes nenhum erro simples (`AvailabilityDialog.svelte:82` `novalidate`, erros só vêm de `form.errors` da API). Não bloqueante |
| RNF-03 | ✅ | `AvailabilityDialog.svelte` (rótulos, `aria-describedby`, `aria-invalid`, foco visível); filtros com `<label for>` | e2e teclado (Enter/Escape); vitest `CA-01.3 / RNF-03` | |
| RNF-04 | ✅ | — | Os testes novos do ciclo 2 citam RN-04 e RN-03 no nome | |
| RNF-05 | ✅ | `server/talents.go` ListTalents; `talentos/+page.server.ts` só manda token com `locals.user` | `TestListTalents_CA04_1_CA04_2`; vitest; e2e | |

Design: rotas, schemas, tabela `character_availability`, migração `00007`, D-01 a D-06 e o
limite de 100 continuam batendo com o implementado; o ciclo 2 não mexeu em contrato nem em
modelo de dados.

## Pendências para correção
- Nenhuma.

## Scope creep
- Nenhum. O commit `5543a67` só toca o rótulo da RN-04, testes e documentação da spec.

## Observações (não bloqueantes)
- RNF-02: o web não avisa antes os erros simples (sem dia, início igual ao fim, sem
  instância); tudo volta da API. A spec pede o aviso prévio.
- O rótulo "Nenhuma instância do catálogo" só tem teste unitário de `instancesLabel`; não há
  teste de componente ou e2e do card com lista vazia. O card chama a função diretamente,
  então o risco é baixo.
- No perfil, um personagem com `anyInstance = false` e todas as instâncias removidas reabre o
  diálogo sem nenhuma marcada; ao salvar sem escolher, recebe o 422 do CA-01.3. Coerente com
  a spec, mas sem aviso de por que a lista ficou vazia.
- `GET /talents/count` segue sem `minimum` em `minLevel`, `tank`, `support` e `dps` no
  `openapi.yaml`.
- A afinidade não exclui quem foi removido com bloqueio do lobby (`candidatura-lobby`).
  Registrado no `tasks.md` como pergunta ao usuário; a spec do banco não trata disso.
- Rastreabilidade ok: os commits citam T-xx e IDs da spec; T-01 a T-07 marcadas [x] com
  código correspondente. O relatório do ciclo 1 foi preservado em
  `validation-report-ciclo-1.md`.
