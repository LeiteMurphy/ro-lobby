# Relatório de validação — lobby-sem-instancia (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend: todos os pacotes `ok` com `-tags=integration`; web: vitest 361/361, e2e Playwright 53/53) · lint ok (golangci-lint 0 issues; prettier + eslint ok; svelte-check 0 erros) · build ok

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 6 | 2 | 0 | 0 |
| Critérios (CA) | 8 | 0 | 0 | 0 |
| Não funcionais | 4 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `backend/internal/lobbies/lobbies.go` Create (ramo `in.AnyInstance` / catálogo); `web/src/lib/lobbies/form.ts` `instance()`; `LobbyForm.svelte` opção `NO_INSTANCE` | `TestCreate_CA01_1_CA01_2_AnyInstance`; vitest "CA-01.1 / RN-01 / RN-02", "CA-01.5 / RN-01"; e2e `sem-instancia.spec.ts` | Padrão continua sendo a 1ª instância (`novo/+page.server.ts:45`). |
| RN-02 | ✅ | `lobbies.go` `anyInstance()` (TrimSpace, 40 runas, vazio → nulo); `toLobby` usa `AnyInstanceName`; CHECK `lobbies_any_instance_check` na migração 00009; `maxlength="40"` no form | `TestCreate_CA01_3_TitleTooLong` (41 recusa, 40 aceita); `TestCreate_CA01_1_CA01_2_AnyInstance` ("" e "   "); `TestMigration00009_RNF01_ExistingLobby` (41, espaço na ponta) | Regra garantida na API e no banco. |
| RN-03 | ✅ | `anyInstance()` devolve `level: 1`; Create checa `MinLevel < target.level`; `LobbyForm.onInstanceChange` põe 1 | `TestCreate_CA01_4_MinLevel` (120 ok, 121 `level_too_low`, 0 `invalid`); e2e confere "Nível mínimo" = 1 | |
| RN-04 | ✅ | `toLobby`: `InstanceName` = título ou "Qualquer instância", `InstanceReset` vazio; `toAPILobby` com `instance.id` nulo; `share.ts` e `[id]/+page.svelte` usam `instance.name`; `instanceArt` sempre devolve a capa genérica (`home/catalog.ts:22`) | `TestToAPILobby_CA01_1_AnyInstance`; vitest "CA-01.1 / RN-04: o convite usa o título"; e2e (h1 do detalhe e card na Home) | Ver observações 1 e 2 sobre textos de tela. |
| RN-05 | ✅ | `home/filters.ts` `instanceMatches` e `instanceOptions` (sem títulos livres); `FilterPanel.svelte` opção "Sem instância definida"; `toHome.ts` marca `anyInstance` | vitest "CA-02.1 / RN-05" (os três filtros e o título igual a uma opção); e2e (filtro `__none__` mostra o card) | |
| RN-06 | ⚠️ | `lobbies.go` Update (`case in.AnyInstance` / outra instância do catálogo); `editar/+page.server.ts` (`__none__` + título); `onInstanceChange` volta ao nível de entrada | `TestUpdate_CA03_1_SwitchAnyInstance`; vitest editar "CA-03.1 / RN-06"; e2e (volta para Sonho Sombrio com 120) | "Indo para sem instância, o título começa vazio" não tem teste; o form guarda o título digitado antes se o dono troca para instância e volta na mesma tela (P2, não bloqueia). |
| RN-07 | ✅ | `queries/talents.sql` (`instance_id = ''` aceita todos); `talents/catalog.go` Count com `AnyInstance`; `affinity.go` passa `InstanceID` vazio do lobby; `novo/disponiveis/+server.ts` manda `anyInstance` | `TestForLobby_CA04_1_AnyInstance` (ForLobby e Count, com o dia fora da faixa excluído); vitest disponiveis "CA-04.1 / RN-07" | |
| RN-08 | ⚠️ | Sem mudança de comportamento nos demais fluxos; `applications.go` `instanceName()` em "Minhas candidaturas" | Suítes completas de lobbies, candidatura, grupo-livre, compartilhar e talentos passam | Sem teste para os casos de borda "sem instância + Grupo livre" e "Minhas candidaturas" de um lobby sem instância (P2, não bloqueia). |
| CA-01.1 | ✅ | ver RN-01/RN-02/RN-04 | `TestCreate_CA01_1_CA01_2_AnyInstance`, `TestToAPILobby_CA01_1_AnyInstance`, vitest convite, e2e detalhe + card | |
| CA-01.2 | ✅ | `toLobby` → `AnyInstanceName` | `TestCreate_CA01_1_CA01_2_AnyInstance`; vitest (prévia com "Qualquer instância") | |
| CA-01.3 | ✅ | `anyInstance()` → `title/too_long`; `messages.ts` "Use até 40 caracteres" | `TestCreate_CA01_3_TitleTooLong`; vitest "CA-01.3 / RN-02" | |
| CA-01.4 | ✅ | ver RN-03 | `TestCreate_CA01_4_MinLevel`; vitest (`min="1"`); e2e | |
| CA-01.5 | ✅ | `novo/+page.server.ts:45` | vitest "CA-01.5 / RNF-03" (instância marcada, sem campo de título); e2e (`not.toHaveValue('__none__')`) | |
| CA-02.1 | ✅ | ver RN-05 | vitest "CA-02.1 / RN-05" cobre os três Then; e2e cobre o primeiro | |
| CA-03.1 | ✅ | ver RN-06 | `TestUpdate_CA03_1_SwitchAnyInstance` (Templo → "Farm" → Sonho Sombrio com 120; 10 recusado); e2e (sem instância → Sonho Sombrio, nível sugerido 120) | O e2e cria já sem instância em vez de trocar a partir do Templo; a troca nesse sentido está provada na API e no `toLobbyUpdate`. |
| CA-04.1 | ✅ | ver RN-07 | `TestForLobby_CA04_1_AnyInstance` | |
| RNF-01 | ✅ | `migrations/00009_lobby_any_instance.sql` (só DROP NOT NULL + CHECK) | `TestMigration00009_RNF01_ExistingLobby` (lobby antigo intacto; up e down) | |
| RNF-02 | ✅ | Validação no serviço e no CHECK do banco; o web só põe `maxlength` e o nível 1 | testes de integração do serviço | |
| RNF-03 | ✅ | `<label for="{uid}-title">Título (opcional)</label>`, `aria-invalid`/`aria-describedby`; `<select>` nativo com rótulo | vitest (rótulo); e2e usa `getByLabel` | |
| RNF-04 | ✅ | — | Todos os testes novos citam CA/RN no nome ou comentário | |

## Pendências para correção
Nenhuma bloqueante.

## Scope creep
- Nenhum. As mudanças em `characters_integration_test.go`, `db/*_integration_test.go` e `free.spec.ts` só acompanham os tipos novos (`pgtype.Text`, campo `title`).

## Observações (não bloqueantes)
1. Texto de tela que ainda supõe instância: no detalhe, a dica de "Jogadores disponíveis" diz "No banco de talentos, com a instância, o horário, o nível…" (`web/src/routes/lobbies/[id]/+page.svelte:553`), também no lobby sem instância.
2. No formulário sem instância, a dica do nível mínimo continua "Mínimo da instância: 1" (`web/src/lib/lobbies/components/LobbyForm.svelte:386`).
3. `instanceName()` de "Minhas candidaturas" (`backend/internal/applications/applications.go:680`) não tem teste com um lobby sem instância (RN-04/RN-08).
4. RN-06: sem teste para "indo para sem instância, o título começa vazio" (o estado `title` do form fica com o que já foi digitado na mesma tela).
5. Caso de borda "sem instância + Grupo livre" sem teste dedicado.
6. O `Lobby` da API não expõe `title`/`anyInstance` à parte; a edição deduz o título comparando `instance.name` com "Qualquer instância". Um título digitado exatamente "Qualquer instância" volta vazio na edição (efeito visual nulo).
7. O `goose down` da 00009 apaga os lobbies sem instância (e o que depender deles). Está documentado e testado, mas é um rollback destrutivo.
