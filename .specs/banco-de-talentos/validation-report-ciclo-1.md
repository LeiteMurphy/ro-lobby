# Relatório de validação — banco-de-talentos (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** testes backend passou (todos os pacotes `ok`, com `-tags=integration -count=1`) · vitest passou (319/319, 28 arquivos) · e2e passou (50/50) · lint ok (`golangci-lint` 0 issues; prettier + eslint ok) · check ok (svelte-check 0 erros, 0 avisos) · build ok

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 11 | 1 | 0 | 0 |
| Critérios (CA) | 13 | 0 | 0 | 0 |
| Não funcionais | 4 | 1 | 0 | 0 |

A reprovação vem de um único item: a RN-04 (US-01, P1) está ⚠️ porque o caso de borda
da seção 6 ("Instância removida do catálogo") tem comportamento visível diferente do
especificado e não tem teste. O resto da feature está coberto por código e teste.

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `backend/queries/talents.sql` (UpsertAvailability religa com `enabled = true`; DisableAvailability só muda `enabled`); `talents.go` SetAvailability | `TestSetAvailability_CA01_1_CA01_4`, `TestSetAvailability_RN01_DisableNeverEnabled`, `TestAvailability_RN01_UpsertAndDisable`, `TestCatalog_CA04_3_DisabledAndInvalid`, e2e "ligar pelo teclado... desligar e reabrir" | |
| RN-02 | ✅ | `talents.go` validate (days required/invalid); migração `CHECK (days BETWEEN 1 AND 127)` | `TestSetAvailability_CA01_3_FieldErrors` ("sem dia", "dia 7"); `TestAvailability_RN02_RN03_RN04_ChecksRejectInvalidValues` | |
| RN-03 | ✅ | `talents.go` ParseClock (30 em 30), `same_as_start`; `talents.sql` Affinity/Catalog com virada e fim excluído; CHECK na migração | `TestSetAvailability_CA01_3_FieldErrors`; `TestForLobby_CA02_2_OutOfAffinity` (AteVinte = fim excluído; Virada); `TestCatalog_CA01_2_Midnight` | Falta caso de virada sábado→domingo (citado nos riscos do design); a fórmula `(dow + 6) % 7` cobre, mas sem teste. Não bloqueante |
| RN-04 | ⚠️ | `talents.go` validate (required/invalid, dedupe); `inCatalog` em `FromRow`; `affinity.go` toTalent filtra ids fora do catálogo | Validação: `TestSetAvailability_CA01_3_FieldErrors`; CHECK: teste de `db`. **Sem teste** para instância que saiu do catálogo | Caso de borda da seção 6 com defeito: personagem com `anyInstance = false` cujas instâncias saíram todas do catálogo chega ao web com `instances: []`, e `instancesLabel` (`web/src/lib/talents/format.ts`) devolve "Qualquer instância" quando `names.length === 0`. O card passa a dizer "Qualquer instância" para quem, pela mesma regra, saiu da afinidade e não aparece no filtro por instância |
| RN-05 | ✅ | `talents.go` SetAvailability (LockUser + GetOwnCharacter → ErrNotFound); FK `ON DELETE CASCADE` | `TestSetAvailability_CA01_5_OtherUser`; `TestAvailability_RN05_CascadeOnCharacterDelete`; `TestSetAvailability_CA01_3_CA01_5_Errors` (404 na rota) | |
| RN-06 | ✅ | `talents.sql` Affinity (cond. 1 a 5); `affinity.go` ForLobby (funções com vaga = slots > occupied) e `wallClock` em `America/Sao_Paulo` | `TestForLobby_CA02_1`, `TestForLobby_CA02_2_OutOfAffinity`, `TestForLobby_CA02_3_Busy`, `TestForLobby_CA02_4_OwnerAndOpen` (lobby sem vaga → lista vazia) | |
| RN-07 | ✅ | `talents.sql` `c.user_id <> exclude_user_id` | `TestForLobby_CA02_3_Busy` (Alt da dona fica de fora) | |
| RN-08 | ✅ | `ORDER BY c.level DESC, lower(c.nick), c.seq` (Affinity e Catalog) | `TestForLobby_CA02_1` (Fogo antes de Brasa); `TestForLobby_CA02_3_Busy` (Fogo, Livre, Brasa); `TestCatalog_CA01_1_CA04_1_Filters` (Cura, Fogo, Brasa: empate em 200 resolvido por nick) | |
| RN-09 | ✅ | `affinity.go` ForLobby (ErrNotFound para outro dono, ErrNotOpen para não aberto); `server/talents.go` 401/404/409; `web/src/routes/lobbies/[id]/+page.server.ts` só busca com `isOwner && status === 'open'`; `+page.svelte` painel | `TestForLobby_CA02_4_OwnerAndOpen`; `TestListLobbyTalents_CA02_4_Errors`; vitest `CA-02.4 / RN-09` (SSR e load); e2e visitante sem painel | |
| RN-10 | ✅ | `catalog.go` Count (vagas do formulário menos a do dono, mesma Affinity); `routes/lobbies/novo/disponiveis/+server.ts`; `LobbyForm.svelte` (300 ms, só na criação, linha some sem resposta) | `TestCount_CA03_1`; `TestCountTalents_CA03_1`; `disponiveis.spec.ts`; e2e CA-03.1 (1 → 0 ao mudar a hora) | |
| RN-11 | ✅ | `talents.sql` Catalog (filtros opcionais com "E", hora com virada); `catalog.go` validação; `routes/talentos/+page.server.ts` (query string, GET sem JS) | `TestCatalog_CA01_1_CA04_1_Filters`, `TestCatalog_CA01_2_Midnight`, `TestCatalog_CA04_3_DisabledAndInvalid`; `talentos.spec.ts`; e2e catálogo | Filtro só com dia inclui faixas que viram a meia-noite vindas do dia anterior; a spec não detalha esse caso, e a escolha está documentada no SQL |
| RN-12 | ✅ | `TalentCard.svelte` (retrato, nick, classe, nível, função, dias e faixa, instâncias, link, `@username` ou "Entre para ver o Discord"); `server/talents.go` toAPITalents omite `discordUsername` sem sessão | `TestListTalents_CA04_1_CA04_2`; vitest `CA-01.1 / RN-12` e `CA-04.2 / RNF-05`; e2e (HTML do visitante sem `@cat…`; logado vê) | Exibição das instâncias tem o defeito descrito na RN-04 |
| CA-01.1 | ✅ | SetAvailability + Catalog + AvailabilityDialog | `TestSetAvailability_CA01_1_CA01_4`; `TestCatalog_CA01_1_CA04_1_Filters`; e2e segunda a sexta, 19–23, Templo e Sonho Sombrio | |
| CA-01.2 | ✅ | Catalog (virada) | `TestCatalog_CA01_2_Midnight` (sábado 01:00 = 1, sábado 22:00 = 0); e2e `dia=6&hora=01:00` / `22:00` | |
| CA-01.3 | ✅ | validate + rota 422 + `talentFieldMessages` | `TestSetAvailability_CA01_3_FieldErrors` (confere que nada foi gravado); `TestSetAvailability_CA01_3_CA01_5_Errors` (422 com `fields`); vitest; e2e mensagens no diálogo | |
| CA-01.4 | ✅ | DisableAvailability; `characters.List` traz `availability`; diálogo abre com os valores guardados | `TestSetAvailability_CA01_1_CA01_4`; vitest `CA-01.4: reabre...`; e2e desligar e reabrir | |
| CA-01.5 | ✅ | GetOwnCharacter → 404 | `TestSetAvailability_CA01_5_OtherUser` (confere que o de outro não mudou); rota 404; vitest perfil `CA-01.5` | |
| CA-02.1 | ✅ | ForLobby + painel | `TestForLobby_CA02_1` (exatamente Fogo, Brasa; Cura de fora); `TestListLobbyTalents_CA02_1`; vitest; e2e | |
| CA-02.2 | ✅ | Affinity | `TestForLobby_CA02_2_OutOfAffinity` (instância, dia, nível) | |
| CA-02.3 | ✅ | Affinity cond. 5 e RN-07 | `TestForLobby_CA02_3_Busy` (pendente aqui, aceito às 21:00, personagem da dona) | |
| CA-02.4 | ✅ | ForLobby + load/página | `TestForLobby_CA02_4_OwnerAndOpen`; `TestListLobbyTalents_CA02_4_Errors`; vitest load (visitante, outro Usuário, iniciado) e SSR; e2e visitante | |
| CA-03.1 | ✅ | Count + endpoint do web + LobbyForm | `TestCount_CA03_1` (3 às 20:00, 0 às 17:00 com os dados do CA-02.1); e2e no navegador | |
| CA-04.1 | ✅ | Catalog | `TestCatalog_CA01_1_CA04_1_Filters` (Tank → Brasa; Sonho Sombrio → Fogo); e2e | |
| CA-04.2 | ✅ | toAPITalents + TalentCard | `TestListTalents_CA04_1_CA04_2` (corpo sem `discordUsername`); vitest; e2e (HTML do visitante sem o nome) | |
| CA-04.3 | ✅ | `WHERE a.enabled` | `TestCatalog_CA04_3_DisabledAndInvalid`; e2e (`Off…` não aparece) | |
| RNF-01 | ✅ | Migração (smallint de minutos e bitmask); `wallClock` com `lobbies.Location` | Testes da afinidade usam `at()` em São Paulo e conferem dia/hora de Brasília | |
| RNF-02 | ⚠️ | Regras na API (serviço + CHECK) | Testes de serviço e de banco | O web não repete nenhuma regra simples antes de enviar (`novalidate`, sem checagem na action); os erros só vêm da API. Não bloqueante |
| RNF-03 | ✅ | `AvailabilityDialog.svelte` (rótulos, `aria-describedby`, `:focus-visible` nos chips); filtros com `<label for>` | e2e abre o diálogo por Enter e fecha por Escape; vitest `CA-01.3 / RNF-03` | |
| RNF-04 | ✅ | — | Todos os testes novos citam CA/RN no nome ou no comentário | |
| RNF-05 | ✅ | `server/talents.go` ListTalents; `talentos/+page.server.ts` só manda token com `locals.user` | `TestListTalents_CA04_1_CA04_2`; vitest `CA-04.2 / D-05`; e2e lê o HTML | |

Design: as quatro rotas, os schemas (`AvailabilityInput`, `Availability`, `Talent`,
`Character.availability`), a tabela `character_availability` com os CHECK e o índice
parcial, a migração `00007`, D-01 a D-06 e o limite de 100 com aviso batem com o
implementado.

## Pendências para correção
1. [RN-04 / caso de borda da seção 6] Personagem com `anyInstance = false` cujas
   instâncias saíram todas do catálogo aparece no catálogo e no painel como "Qualquer
   instância": `instancesLabel` em `web/src/lib/talents/format.ts` trata
   `names.length === 0` como "Qualquer". A spec diz que ele sai da afinidade e continua
   no banco, e a RN-12 manda mostrar as instâncias de interesse dele. O rótulo precisa
   seguir só o `anyInstance` (com algum texto próprio para lista vazia), com teste
   citando RN-04.
2. [RN-04] Falta teste para "Instância que sair do catálogo some da lista do
   personagem": nenhum teste grava um id fora do catálogo (dá para gravar direto com
   `q.UpsertAvailability`, que não valida o catálogo) e confere que `FromRow`, o
   catálogo e a afinidade o filtram, e que o personagem sem instância restante fica de
   fora da afinidade e continua em `/talentos`.

## Scope creep
- Nenhum. As mudanças fora de `talents` são ajustes de assinatura (`server.New` com o
  serviço novo) e de fixtures (`availability: null`) exigidos pelo campo novo do
  contrato. O link na TopBar e o `research.md` estão no design e no processo.

## Observações (não bloqueantes)
- RNF-02: o web não avisa antes nenhum erro simples (sem dia, início igual ao fim);
  tudo volta da API. Funciona, mas a spec pede o aviso prévio.
- Sem teste da virada sábado→domingo, citada na tabela de riscos do design.
- `GET /talents/count` não tem `minimum` em `minLevel`, `tank`, `support` e `dps` no
  `openapi.yaml`; valores negativos passam e só zeram a contagem.
- A condição 5 da RN-06 não exclui personagens removidos com bloqueio deste lobby
  (spec `candidatura-lobby`). A spec do banco não pede isso, mas o dono pode ver na
  lista alguém que não pode se candidatar de novo.
- Rastreabilidade ok: os commits citam T-xx e os IDs da spec, e as tasks marcadas [x]
  apontam para commits que existem no intervalo.
