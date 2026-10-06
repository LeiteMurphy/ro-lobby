# Relatório de validação — personagens (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend unitário: 13 pacotes ok · backend integração: 13 pacotes ok, 0 falhas · Vitest 148/148 · Playwright 32/32) · lint ok (golangci-lint 0 issues; prettier + eslint ok; svelte-check 0 erros, 0 avisos) · build ok (`npm run build`, adapter-node)

Alterações analisadas: `2ea8bf0..HEAD` (14 commits, 55 arquivos).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 20 | 0 | 0 | 0 |
| Critérios (CA) | 26 | 0 | 0 | 0 |
| Não funcionais | 3 | 1 | 0 | 0 |

## Detalhe por item

### Regras
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `server/characters.go:36` (`currentUser` em toda rota); consultas filtram `user_id` (`queries/characters.sql`); FK `ON DELETE CASCADE` (`00003_characters.sql`) | `TestCharacters_CA07_2_RequireSession`, `TestList_CA01_3_RN17_OnlyOwnMainFirst`, `TestCharacters_RN01_DeletedWithUser`, `TestCreate_RN01_UnknownUser`, page.server.spec "RN-01: sem cookie, nenhuma action chama a API" | |
| RN-02 | ✅ | `UpdateCharacter`/`DeleteCharacter`/`SetMain` com `WHERE id AND user_id`; `characters.go:182,213,235` (id malformado = `ErrNotFound`); rotas → 404 | `TestOtherUsersCharacter_CA03_3_CA04_4_NotFound` (inclui id inexistente e malformado, e confere que nada mudou), `TestCharacters_RN02_OtherUsersCharacterIsUntouched`, `TestCharactersFlowIntegration` | |
| RN-03 | ✅ | `routes/perfil/+page.server.ts:31-38` | page.server.spec "CA-06.1 / RN-03", e2e "CA-06.1" | |
| RN-04 | ✅ | `characters.go:260-268` (trim, 1–24 runas, controle); CHECK no banco | `TestValidate_FieldErrors`, `TestValidate_RN04_TrimsNick`, `TestValidate_AcceptsLimits` (24 × "ã"), `TestCharacters_RN04_RN07_RN08_RN09_ChecksRejectInvalidValues` | |
| RN-05 | ✅ | índice `characters_nick_key` em `lower(nick COLLATE pg_c_utf8)`; `translate` → `nick/taken` (`characters.go:351`) | `TestCharacters_CA07_3_NickUniqueIgnoringCase` (Brasa/brasa, FAÍSCA/faísca), `TestCharacters_RN05_AccentMakesADifferentNick`, `TestCreate_CA02_3_NickTakenByAnotherUser`, `TestUpdate_RN05_NickOfAnotherCharacter` | |
| RN-06 | ✅ | `characters.go:270-274` + `catalog.ClassByID` | `TestValidate_FieldErrors` (CA-02.4), `TestClassByID_RN06`, `TestUpdate_RN06_RemovedClassAsksForValidOne` | |
| RN-07 | ✅ | `characters.go:276`; CHECK 1..275 | `TestValidate_FieldErrors` (0 e 276), `TestValidate_AcceptsLimits` (1 e 275), CHECK no banco | |
| RN-08 | ✅ | `characters.go:280-285`; CHECK `IN ('tank','support','dps')` | `TestValidate_FieldErrors`, `TestContractEnums_D04_MatchCatalog` | |
| RN-09 | ✅ | `characters.go:293-316` (D-09); `CharacterCard.svelte` `target="_blank" rel="noopener noreferrer"` | `TestValidate_FieldErrors` (http, javascript, sem host, com usuário, 301 chars), page.spec "CA-01.4", e2e CA-01.4 | |
| RN-10 | ✅ | `characters.go:287-291` (padrão `retrato-1`); `static/portraits/*.svg`; `portraits.ts` | `TestPortraits_RN10`, `TestCreate_CA02_1_CA02_2_FirstIsMainWithDefaultPortrait`, characters.spec "D-04: cada retrato do contrato tem um arquivo", e2e CA-02.2 | |
| RN-11 | ✅ | `characters.go:152` dentro de `inTx` com `LockUser FOR UPDATE` | `TestCreate_CA02_8_LimitOfTen`, `TestCreate_CA07_3_Concurrent` (15 simultâneos → 10 criados, 5 recusados) | |
| RN-12 | ✅ | `IsMain: count == 0` (`characters.go:163`); índice parcial `characters_one_main_per_user` | `TestCharacters_RN12_OneMainPerUser`, `TestDelete_RN12_OnlyCharacter`, concorrência com 1 principal | |
| RN-13 | ✅ | `SetMain` (`characters.go:229-252`) com `ClearMain` + `SetMain` na transação | `TestSetMain_CA05_1`, e2e CA-05.1 | |
| RN-14 | ✅ | `DeleteCharacter ... RETURNING is_main` + `PromoteOldest` (`created_at, seq`) | `TestDelete_CA04_3_OldestBecomesMain`, `TestDelete_RN14_NonMainKeepsMain`, `TestCharacters_RN17_RN14_OrderAndOldest` | |
| RN-15 | ✅ | `Update` reaplica `Validate`; `ConfirmDialog.svelte` | `TestUpdate_CA03_1_CA03_2`, e2e CA-03.1, CA-04.1, CA-04.2 | |
| RN-16 | ✅ | `CharacterCard.svelte`, `CardMenu.svelte` (`{#if !isMain}` para "Tornar principal") | page.spec "CA-01.1 / RN-16 / RN-17", e2e "O principal não oferece 'Tornar principal'" | |
| RN-17 | ✅ | `ListCharacters ORDER BY is_main DESC, created_at, seq` | `TestList_CA01_3_RN17_OnlyOwnMainFirst`, `TestCharacters_RN17_RN14_OrderAndOldest`, e2e ordem dos headings | |
| RN-18 | ✅ | `UserMenu.svelte` (link `/perfil` antes do form de logout) | home.spec "CA-06.2 / RN-18", e2e CA-06.2 | |
| RN-19 | ✅ | `+page.server.ts:93-111` (`fail(422, {values, errors})`), `messages.ts`, `CharacterDialog.svelte` (`aria-describedby` no campo) | page.server.spec "CA-02.9 / RN-19", page.spec "CA-02.9 / RN-19", e2e CA-02.9 (valores e mensagem conferidos), `TestValidate_RN19_ReportsAllFields` | |
| RN-20 | ✅ | `ListClasses` (`server/characters.go:48`) sem `security` no contrato; diálogo usa `data.classes` | `TestClasses_CA07_1_PublicCatalog`, `TestClasses_CA07_1_MatchBROWiki`, `TestContract_CA07_CharacterRoutesAreDescribed`, page.spec "RN-06 / RN-20: o select tem as 82 classes" | |

### Critérios de aceite
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| CA-01.1 | ✅ | `+page.svelte`, `CharacterCard.svelte` | page.spec CA-01.1 (retrato, nick, classe, nível, função, selo); e2e CA-01.1 | |
| CA-01.2 | ✅ | `+page.svelte:81-92` | page.spec CA-01.2; e2e (texto do perfil vazio) | |
| CA-01.3 | ✅ | `ListCharacters` por `user_id` | `TestList_CA01_3_...`, `TestCharactersFlowIntegration`, e2e com duas contas | |
| CA-01.4 | ✅ | `CharacterCard.svelte:279-288` | page.spec CA-01.4; e2e confere href, target e rel | |
| CA-02.1 | ✅ | `characters.go:163` | `TestCreate_CA02_1_CA02_2_...`, `TestCharactersFlowIntegration`, e2e | |
| CA-02.2 | ✅ | `Validate` (padrão), rádios do diálogo | `TestCreate_CA02_1_CA02_2_...`, page.server.spec CA-02.2, e2e (src `retrato-1` e `retrato-3`) | |
| CA-02.3 | ✅ | índice + `translate`; `messages.ts` "Esse nick já está em uso" | `TestCreate_CA02_3_...`, integração da rota (422), e2e com a mensagem exata | |
| CA-02.4 | ✅ | `characters.go:272` → 422 `classId/invalid` | `TestValidate_FieldErrors` ("Paladino Supremo" e "paladino-supremo"), `TestCharacters_CA02_4_ValidationIs422` | |
| CA-02.5 | ✅ | `characters.go:276` | `TestValidate_FieldErrors`, `TestCreate_CA02_4_to_CA02_7_...` | O 422 na rota é provado pelo mapeamento genérico (`TestCharacters_CA02_4_ValidationIs422`) |
| CA-02.6 | ✅ | `characters.go:261-265` | idem (nick "   " e 25 caracteres) | |
| CA-02.7 | ✅ | `validLink` | idem (`http://exemplo.com`, `javascript:alert(1)`) | |
| CA-02.8 | ✅ | `ErrLimitReached` → 409 `character_limit` → "Você já tem 10 personagens"; botão `disabled` + `aria-describedby` | `TestCreate_CA02_8_LimitOfTen`, `TestCharacters_CA02_8_LimitIs409`, page.server.spec CA-02.8, page.spec CA-02.8, e2e CA-02.8 | |
| CA-02.9 | ✅ | `CharacterDialog.svelte` (`update({ reset: false })`, erros por campo) | e2e confere mensagem ligada ao nick e classe, nível, função e link preservados | |
| CA-03.1 | ✅ | `Update` | `TestUpdate_CA03_1_CA03_2`, e2e (Elementalista, 170) | |
| CA-03.2 | ✅ | `UpdateCharacter` na própria linha não viola o índice | `TestUpdate_CA03_1_CA03_2` (mesmo nick e só mudando a caixa) | |
| CA-03.3 | ✅ | `WHERE id AND user_id` → 404 | `TestOtherUsersCharacter_...`, `TestCharacters_CA03_3_CA04_4_NotFoundIs404`, `TestCharactersFlowIntegration` | "Continua igual" verificado na lista da dona |
| CA-04.1 | ✅ | `Delete` | `TestDelete_CA04_1_FreesNick`, e2e (recadastra o nick em outra caixa) | |
| CA-04.2 | ✅ | `ConfirmDialog.svelte` "Cancelar" | e2e CA-04.2 (3 cartas continuam) | |
| CA-04.3 | ✅ | `PromoteOldest` | `TestDelete_CA04_3_OldestBecomesMain` (cenário exato da spec), e2e | |
| CA-04.4 | ✅ | idem CA-03.3 | `TestOtherUsersCharacter_...` (personagem continua existindo) | |
| CA-05.1 | ✅ | `SetMain` | `TestSetMain_CA05_1`, `TestList_CA01_3_...`, e2e (ordem e selo único) | |
| CA-06.1 | ✅ | `load` → `/auth/discord/login?next=/perfil` | page.server.spec CA-06.1, e2e (volta a `/perfil`) | |
| CA-06.2 | ✅ | `UserMenu.svelte` | home.spec CA-06.2, e2e (itens `['Meu perfil', 'Sair']` e navegação) | |
| CA-07.1 | ✅ | `catalog.Classes()` (82) | `TestClasses_CA07_1_PublicCatalog` (sem sessão, 82), `TestClasses_CA07_1_MatchBROWiki` | |
| CA-07.2 | ✅ | `currentUser` em todas as rotas | `TestCharacters_CA07_2_RequireSession` (sem token e token inválido, nas 5 rotas, serviço não chamado) | |
| CA-07.3 | ✅ | índice único + `LockUser FOR UPDATE` | `TestCreate_CA07_3_Concurrent` (mesmo nick simultâneo por dois Usuários → 1 aceito, 1 `taken`), `TestCharacters_CA07_3_NickUniqueIgnoringCase` | |

### Não funcionais
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RNF-01 | ✅ | `<dialog>` nativo com `showModal`; `CardMenu` com setas/Escape; rótulos em todos os campos; `:focus-visible` global (`tokens/base.css:31`) | e2e "RNF-01: menu '…' e diálogos pelo teclado"; page.spec "RNF-01" | |
| RNF-02 | ✅ | retratos em `static/portraits/`, ícones Lucide empacotados | characters.spec "RNF-02: os SVGs não carregam nada de fora", e2e "RNF-02: /perfil não carrega nada de fora do servidor" | |
| RNF-03 | ⚠️ | — | — | Quase todos os testes citam ID, mas alguns não citam nenhum: characters.spec "monta as rotas de editar, excluir e principal com o id escapado" e "API fora do ar vira unavailable, sem lançar"; page.server.spec "API fora do ar mostra o aviso, sem quebrar a página"; page.spec "API fora do ar mostra o aviso"; `TestCharacters_UnexpectedErrorIs500`; `TestClasses_ReturnsACopy`; `TestValidate_AcceptsValidInput` cita CA-02.2 só no comentário da asserção. Vários citam só RN/D (sem CA), o que o projeto já aceitou nas features anteriores |
| RNF-04 | ✅ | `Validate` no serviço + CHECKs no banco; web só repete `maxlength`, `min`/`max` e `required` (com `novalidate`, a API decide) | page.server.spec "RNF-04: nível vazio ou não inteiro vai como 0, e a API devolve o erro" | |

### Casos de borda (seção 6)
| Borda | Veredito | Teste |
|---|---|---|
| " Brasa " salvo como "Brasa" | ✅ | `TestValidate_RN04_TrimsNick`, `TestCreate_CA02_1_CA02_2_...` |
| "Faísca" ≠ "Faisca" | ✅ | `TestCharacters_RN05_AccentMakesADifferentNick` |
| Excluir o único personagem | ✅ | `TestDelete_RN12_OnlyCharacter` |
| Classe que saiu do catálogo | ✅ | `TestUpdate_RN06_RemovedClassAsksForValidOne`, page.spec "borda de RN-06" |
| Sessão vencida no meio da edição | ✅ | page.server.spec "borda: sessão vencida no meio da edição leva ao login e apaga o cookie" |

### Design
- Endpoints, schemas (`Class`, `Role`, `Portrait`, `CharacterInput`, `Character`, `ValidationError`, `Error` com `not_found` e `character_limit`) e códigos de status batem com o `openapi.yaml` e o código gerado; `TestContract_CA07_CharacterRoutesAreDescribed` confere rotas, respostas e sessão.
- Modelo de dados igual ao item 4 do design, mais a coluna `seq` (desempate da ordem de cadastro), que serve a RN-14/RN-17 e não contradiz o design.
- D-01 a D-11 respeitadas. D-06 com teste de concorrência; D-08 sem consulta separada de existência; D-11 com teste no Discord falso (`TestAuthorize_D11_ChooseAnotherUser`).
- ADR-02/04/05/07 respeitadas (`net/http` strict server, `pgx` + `sqlc` + `goose`, contrato como fonte, token só no servidor do web).

### Rastreabilidade
- Os 14 commits do intervalo seguem Conventional Commits e citam IDs.
- As 8 tasks marcadas `[x]` têm commit e código correspondentes. A parte da T-05 "um ícone para cada linha do catálogo" não foi feita, e isso está registrado em "Descobertas" com a justificativa (a carta 1b não mostra ícone de classe).

## Pendências para correção
Nenhuma bloqueante.

## Scope creep
Nenhum. As mudanças fora de `characters`/`perfil` se ligam a IDs da spec ou do design:
- `backend/internal/discordfake/fake.go` — D-11.
- `backend/internal/api/oapi.cfg.yaml` e renomes em `server/auth.go` e testes — consequência da geração dos enums novos (registrada em "Descobertas").
- `web/src/lib/ui/Button.svelte` (`type`, `describedby`) e `Icon.svelte` (5 ícones) — apoio a CA-02.8, RN-16 e RN-19.
- `UserMenu.svelte` (setas, menu sempre no HTML com `hidden`) e ajuste em `test/e2e/login.spec.ts` — RN-18 e RNF-01.

## Observações (não bloqueantes)
- **RNF-03:** acrescentar um ID aos poucos testes listados acima (ex.: "API fora do ar" → borda de RN-01/RN-03 ou RN-19; rotas escapadas → RN-02).
- **Diálogo com erro antigo:** depois de uma falha em "Adicionar personagem", se o Usuário fechar o diálogo e abrir de novo, o `form` da última action ainda é do modo `create`, então o diálogo reabre com os valores e as mensagens de erro anteriores (`+page.svelte:31-38`). Não fere nenhum critério, mas pode confundir. Uma saída é ignorar o `form` quando o diálogo é aberto por clique.
- **Retratos:** D-04 descreve "fundo em gradiente com um emblema geométrico"; os SVGs são pixel art em faixas com emblema (e `portraits.ts` já diz "pixel art"). Vale alinhar o texto do design.
- **E2E no banco de dev:** os testes deixam usuários e personagens com sufixo aleatório no banco de desenvolvimento a cada execução. O risco de colisão está mitigado; o acúmulo de dados não.
- **`Update` fora da transação com trava:** correto hoje (não mexe em limite nem principal), mas quando entrarem as travas da `candidatura-lobby` (RN-25/RN-26 de lá) a edição provavelmente vai precisar do mesmo `inTx`.
