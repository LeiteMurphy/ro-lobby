# Relatório de validação — lobbies (ciclo 3)

**Veredito geral:** APROVADO
**Execução:** testes passou: backend unitários 11/11 pacotes com teste · backend integração 15/15 pacotes com teste, 0 falhas (`-count=1`, sem cache) · Vitest 193/193 (19 arquivos) · Playwright 39/39 (banco `ro_lobby_e2e` recriado; inclui os ponta a ponta de `login-discord`, `personagens`, `home-local` e `status`) · lint ok (golangci-lint 0 issues; Prettier e ESLint ok; svelte-check com 0 erros e 0 avisos em 487 arquivos) · build ok

Intervalo validado: `origin/main..HEAD` (17 commits, de `516685b` a `c02abdd`), na branch `feature/lobbies`. Desde o ciclo 2 entrou `628da48` (relatório do ciclo 2) e `c02abdd` (ajustes do teste do usuário: `owner.portrait` no contrato, detalhe com o fundo da Home e vagas maiores com retrato, card inteiro clicável, fundo das páginas de perfil, criar, editar e erro). Todos os comandos foram executados de novo neste ciclo.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 23 | 0 | 0 | 0 |
| Critérios (CA) | 31 | 0 | 0 | 0 |
| Não funcionais | 5 | 0 | 0 | 0 |

O commit `c02abdd` cumpre as versões novas da RN-15 (retrato, nick, classe, nível e selo "Anfitrião") e da RN-23 (clicar em qualquer ponto do card abre o detalhe), com testes que conferem os dois comportamentos. Nada do que estava ✅ no ciclo 2 regrediu.

## Detalhe por item

### Itens tocados por `c02abdd`
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-15 | ✅ | queries/lobbies.sql (`c.portrait AS owner_portrait` em `GetLobby` e `ListOpenLobbies`); lobbies.go `toLobby` (`Owner.Portrait` só com personagem); server/lobbies.go:228-231 (`portrait` no corpo, nulo se vazio); openapi.yaml `LobbyOwner.portrait` (enum dos 4 retratos, nullable, required); `[id]/+page.svelte` (vaga do dono com `<img src="/portraits/…svg">`, nick, "Classe · Nv", selo "Anfitrião"; card "Anfitrião" lateral) | TestCreate_CA01_1_Defaults (integração: `Portrait: "retrato-1"` lido do banco); TestLobbies_RN22_PublicList (JSON com `"portrait": "retrato-2"`); lobbies.spec.ts "CA-03.1 / RN-15" (regex: retrato-2.svg → Lirien → "Arcebispo · Nv 178" → "Anfitrião"); e2e CA-03.1 (`owner-slot` com nick, "Arcebispo · Nv 178" e "Anfitrião") | Estado, data e hora, nível mínimo, observação e 404 continuam cobertos como no ciclo 2. Caso nulo sem teste próprio (ver observações). |
| RN-23 | ✅ | LobbyCard.svelte: `.detail::after` com `position:absolute; inset:0; z-index:1` sobre o `.card` (`position:relative`); `.action` com `z-index:2` mantém "Candidatar" por cima; prévia (`preview`) continua sem link | e2e lobbies "CA-02.4": confere o link "Ver grupo: … às 20:00" visível e clica na posição (24, 24) do card, sobre a capa, longe do título → URL do detalhe; home.spec.ts "CA-02.3 / CA-02.4 / RN-23 (lobbies)"; e2e home CA-02.5 ("Candidatar" segue desabilitado no card) passa | O destaque (`FeaturedLobby`) não ficou clicável por inteiro; ele mantém o botão "Ver grupo" (ver observações). |
| CA-02.4 | ✅ | LobbyCard.svelte, FeaturedLobby.svelte | e2e lobbies CA-02.4; home.spec.ts | |
| CA-03.1 | ✅ | `[id]/+page.svelte` | lobbies.spec.ts "CA-03.1 / RN-15"; e2e CA-03.1 | |
| CA-03.2 | ✅ | `+error.svelte` (só perdeu o fundo próprio, agora vem do `body`) | lobbies.server.spec.ts "CA-03.2"; TestGetLobby_CA03_2_NotFound; e2e (`error-page`) | |
| RNF-01 | ✅ | O link do card segue sendo o único ponto focável do título; o `::after` não muda a ordem do Tab | e2e home "RNF-01: todo ponto de Tab mostra o anel de foco âmbar" (desktop e celular) e lobbies "RNF-01" passam | |
| RNF-02 | ✅ | Os retratos vêm de `/portraits/*.svg` do próprio web | e2e lobbies "RNF-02: criação e detalhe não carregam nada de fora do servidor"; home RNF-02 | |

### Demais regras e critérios
Sem mudança desde o ciclo 2; testes reexecutados neste ciclo, todos passando. A evidência por item é a mesma do relatório do ciclo 2 (commit `628da48`):

| ID | Veredito | Evidência (código) | Evidência (teste) |
|---|---|---|---|
| RN-01 | ✅ | catalog/instances.go (51 instâncias de grupo, "Sarah vs Fenrir"); `ListInstances` sem sessão | TestInstances_CA06_1_GroupInstancesFromBROWiki, TestInstances_CA06_1_PublicCatalog, TestGet_RN01_InstanceLeftCatalog |
| RN-02 | ✅ | catalog/instances.go (ordem); LobbyForm.svelte (dois `optgroup`) | TestInstances_CA06_2_ChoiceOrder; lobbies.spec.ts "CA-06.2 / RN-02" |
| RN-03 | ✅ | lib/home/catalog.ts `instanceArt` (capa e ícone padrão) | home.spec.ts "RN-03 (lobbies)" |
| RN-04 | ✅ | server/lobbies.go (401); novo/+page.server.ts `toLogin` | TestLobbies_CA06_3_WritesRequireSession; e2e CA-02.3 |
| RN-05 | ✅ | lobbies.go `checkStart` | TestCreate_CA01_2_StartWindow, TestCreate_RN05_DayTurnInSaoPaulo |
| RN-06 | ✅ | lobbies.go `checkSlots`; CHECK da migração 00004 | TestCreate_CA01_3_to_CA01_12_FieldErrors; TestLobbies_RN06_RN07_RN09_RN19_ChecksRejectInvalidValues |
| RN-07 | ✅ | lobbies.go; CHECK `min_level >= instance_level` | subteste CA-01.4; TestUpdate_CA04_2_CA04_3_Limits |
| RN-08 | ✅ | lobbies.go (personagem próprio, nível, vaga da função) | subtestes CA-01.5..CA-01.7; TestCreate_CA01_1_Defaults |
| RN-09 | ✅ | lobbies.go `checkNote`; CHECK | subteste CA-01.11 |
| RN-10 | ✅ | `checkConflict`; `HasScheduleConflict` | TestCreate_CA01_8_ScheduleConflict, TestLobbies_RN10_ScheduleConflictWindow |
| RN-11 | ✅ | lobbies.go (limite 5 com usuário travado) | TestCreate_CA01_9_LimitOfFive, TestCreate_CA06_5_Concurrent |
| RN-12 | ✅ | novo/+page.svelte | lobbies.spec.ts "CA-01.10"; e2e |
| RN-13 | ✅ | lobbies.go `toLobby` (estado derivado) | TestList_CA02_2_OnlyOpen |
| RN-14 | ✅ | `ListOpenLobbies` | TestList_CA02_2_OnlyOpen, TestLobbies_RN13_RN14_ListOnlyOpen; e2e |
| RN-16 | ✅ | `[id]/+page.svelte` (ações do dono / "Candidatar" em breve) | lobbies.spec.ts "CA-03.3 / RN-16"; e2e CA-03.3 |
| RN-17 | ✅ | lobbies.go `Update` | TestUpdate_CA04_1, TestUpdate_CA04_4_OtherOrNotOpen |
| RN-18 | ✅ | lobbies.go | TestUpdate_CA04_2_CA04_3_Limits |
| RN-19 | ✅ | lobbies.go `Cancel`; CHECK; CancelDialog.svelte | TestCancel_CA05_1_CA05_2, TestLobbies_RN19_CancelNeedsReason; e2e |
| RN-20 | ✅ | `ownOpenLobby` (404) | TestLobbyWrites_CA04_4_CA05_3_NotFoundAndNotOpen |
| RN-21 | ✅ | characters.go `checkNotInOpenLobby` | TestOwnerOfOpenLobby_CA06_4_Locked; TestCharacters_CA06_4_InOpenLobbyIs409; e2e |
| RN-22 | ✅ | routes/+page.server.ts; lib/lobbies/toHome.ts | home.server.spec.ts "CA-02.1 / RN-22"; e2e |
| CA-01.1..CA-01.12 | ✅ | lobbies.go `Create` e validações | TestCreate_CA01_1_Defaults, TestCreate_CA01_2_StartWindow, TestCreate_CA01_3_to_CA01_12_FieldErrors, TestCreate_CA01_8_ScheduleConflict, TestCreate_CA01_9_LimitOfFive, TestCreateLobby_CA01_7_CA01_12_ValidationIs422; lobbies.spec.ts CA-01.10; e2e CA-01.1, CA-01.8, CA-01.10 |
| CA-02.1..CA-02.3 | ✅ | +page.server.ts; `ListOpenLobbies`; TopBar.svelte | home.server.spec.ts; TestList_CA02_2_OnlyOpen; e2e CA-02.1, CA-02.3 |
| CA-03.3 | ✅ | `[id]/+page.svelte` | lobbies.spec.ts; e2e |
| CA-04.1..CA-04.4 | ✅ | lobbies.go `Update`; server 404/409 | TestUpdate_CA04_1, TestUpdate_CA04_2_CA04_3_Limits, TestUpdate_CA04_4_OtherOrNotOpen; e2e (detalhe e Home) |
| CA-05.1..CA-05.3 | ✅ | lobbies.go `Cancel`; `[id]/+page.svelte` | TestCancel_CA05_1_CA05_2; TestLobbyWrites_CA04_4_CA05_3_NotFoundAndNotOpen; e2e |
| CA-06.1..CA-06.5 | ✅ | catalog; server/lobbies.go; characters.go; `inTx` + `LockUser` | TestInstances_CA06_1_*, TestInstances_CA06_2_ChoiceOrder, TestLobbies_CA06_3_WritesRequireSession, TestOwnerOfOpenLobby_CA06_4_Locked, TestCreate_CA06_5_Concurrent |
| RNF-03 | ✅ | — | Os testes alterados em `c02abdd` seguem citando IDs (CA-03.1 / RN-15, CA-02.4, RN-23) |
| RNF-04 | ✅ | Regras no serviço Go e nos CHECK; o web só repete o formato | lobbies.server.spec.ts "RNF-04" |
| RNF-05 | ✅ | routes/+page.server.ts | home.spec.ts (SSR); e2e "CA-02.6" |

### Design e rastreabilidade
- Contrato: `openapi.yaml` ganhou `LobbyOwner.portrait` (enum `retrato-1..4`, nullable, required), e o código gerado em Go (`api.gen.go`) e em TypeScript (`schema.gen.ts`) está em dia (build, lint e svelte-check passam). O enum bate com o catálogo de retratos (TestContract dos retratos em characters_test.go).
- Modelo de dados: sem migração nova; o retrato vem de `characters.portrait` (NOT NULL) pelo `LEFT JOIN` já existente, e fica nulo quando o personagem foi excluído (`ON DELETE SET NULL`).
- Commits: os 17 seguem Conventional Commits e citam IDs; `c02abdd` cita US-02, US-03, RN-15, RN-23, CA-02.4 e CA-03.1. A spec registra a revisão (cabeçalho "Última revisão" e RN-15/RN-23), e a última entrada de "Descobertas" do tasks.md descreve o ajuste.
- Features anteriores: `login-discord`, `personagens`, `home-local` e `status` continuam passando (Go, Vitest e Playwright). A troca de fundo no `/perfil` não quebrou nenhum teste da `personagens`.

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. As mudanças de `c02abdd` fora do detalhe e do card (fundo do `/perfil`, `/lobbies/novo`, `/lobbies/[id]/editar` e `+error.svelte`, que só removem `background: var(--ink-0)` para herdar o céu do `body`) são ajustes visuais pedidos pelo usuário e registrados em "Descobertas". Não tocam itens de "Fora de escopo".

## Observações (não bloqueantes)
- [Design] O `design.md` (§ contratos, linha 69) ainda lista `owner` como `{userId, discordName, characterId, nick, classId, level, role}`, sem `portrait`. Vale atualizar o design para bater com o `openapi.yaml`.
- [RN-15] Falta teste do retrato nulo: `toAPILobby` só preenche `portrait` com valor, mas nenhum teste do servidor confere `"portrait": null` para o personagem excluído. No web, o teste "D-01: personagem do dono excluído" (lobbies.spec.ts:176) espalha `TEMPLE.owner` e mantém `portrait: 'retrato-2'` com `nick: null`, uma combinação que a API não produz; o ramo sem `<img>` do detalhe fica sem teste. Sugestão: pôr `portrait: null` nesse fixture e conferir que não há `/portraits/` no HTML.
- [RN-23] O card do destaque (`FeaturedLobby`) não ficou clicável por inteiro; ele continua abrindo o detalhe só pelo botão "Ver grupo". A spec fala em "card", e o destaque é um bloco à parte na `home-local`, então não conta como falha, mas pode surpreender o usuário. Vale confirmar.
- [RN-23] O `::after` do link cobre a capa, a linha do anfitrião e a composição; hoje nenhum desses tem elemento interativo. Se a composição ou o anfitrião ganharem link ou tooltip no futuro, eles precisam de `position: relative; z-index: 2`, como `.action`.
- Mantidas do ciclo 2, ainda válidas:
  - [CA-04.1] O e2e da Home confere "1/8", não a linha de Dano do card.
  - [RNF-03] `TestLobbies_RN21_UnexpectedErrorIs500` (server/lobbies_test.go:224) cita a RN-21 de outra spec.
  - [RN-18] Com o dono abaixo do nível mínimo atual, qualquer edição é recusada até baixar o nível mínimo.
  - [RN-03] Os arquivos `/brand/inst-*.svg` seguem sem uso.
  - [RN-02] O desempate por nome compara bytes, não usa collation pt-BR.
