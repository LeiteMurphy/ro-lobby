# Relatório de validação — retrato-por-classe (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes passaram (backend `go test -tags=integration ./...`: todos os pacotes ok · web vitest 368/368 · Playwright 54/54) · lint ok (golangci-lint 0 issues; prettier + eslint ok) · build/check ok (svelte-check: 0 erros, 0 avisos)

Alterações conferidas: `main..feature/retrato-por-classe` (81185a5..c4a4f03), 62 arquivos.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 6 | 0 | 0 | 0 |
| Critérios (CA) | 8 | 0 | 0 | 0 |
| Não funcionais | 3 | 0 | 0 | 0 |
| Casos de borda (seção 6) | 3 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `backend/internal/catalog/catalog.go` (`arts`, `classArt`, `ArtOf`); `GET /classes` manda `art` em `server/characters.go` (`ListClasses`) | `TestArtOf_CA01_1_CA01_3_CA01_6` (toda classe do catálogo tem arte da lista de 27; casos de linha, família e Aprendiz); `TestClasses_CA07_1_PublicCatalog` (campo `art`) | Conferência mecânica: as 82 classes do catálogo = as 82 chaves de `classArt` = as 82 classes da tabela da seção 6 do `tasks.md`, sem nenhuma divergência de arte. |
| RN-02 | ✅ | `toAPICharacter` usa `catalog.ArtOf(c.ClassID)`; `toInput` ignora `portrait` da entrada; `CharacterDialog.svelte` sem a escolha, com `portraitForClass(classId, classes)` | `TestToAPICharacter_CA01_2_CA01_5_ClassPortrait`; `TestCharacters_CA02_1_Create` (entrada sem portrait); e2e `CA-01.4 / CA-01.2 (retrato-por-classe)` | Regra garantida na API (fonte), não só no web. |
| RN-03 | ✅ | O retrato gravado deixa de ser lido em todas as respostas (`characters.go`, `lobbies.go`, `applications.go`, `swaps.go`, `talents.go`) | `TestToAPICharacter_CA01_2_CA01_5_ClassPortrait` (Cardeal com `retrato-3` gravado → `sacerdote`) | Sem migração de dados; coluna mantida (D-03). |
| RN-04 | ✅ | `ArtOf` devolve `DefaultPortrait` (`retrato-1`); `portraitInfo` no web cai em `retrato-1.svg` | `TestArtOf_...` (`classe-removida` → `retrato-1`); vitest `CA-01.6 / RN-04`; SSR `/perfil` `CA-01.6 (retrato-por-classe)` | |
| RN-05 | ✅ | `portraitSrc`/`portraitAlt` em ApplyDialog, PlayerPanel, SwapDialog, SwapPanel, LobbyForm, candidaturas, detalhe do lobby (dono, membros, pendentes, trocas), TalentCard e CharacterCard (alt = nome da linha) | vitest `CA-01.1 / RN-05`; SSR perfil, lobbies (`CA-02.1`), talentos (`CA-02.2`); e2e perfil por `getByRole('img', { name: 'Linha do …' })` | Nenhum `/portraits/*.svg` fixo sobrou nas telas (grep). A Home não mostra retrato de personagem. |
| RN-06 | ✅ | 27 PNG 64×64 em `web/static/portraits/classes/`; gerador próprio `scripts/class-art/class_art.py` (Pillow, desenho procedural); `image-rendering: pixelated` em todos os componentes que mostram o retrato | Inspeção visual de folha com as 27 artes ampliadas 3× (vizinho mais próximo) | As artes são emblemas (frasco, flecha, adaga, violino, tocha, arco, espada, folha, pergaminho, martelo e engrenagem, máscara, pata, revólver, varinha, bolsa de moedas, katar, sol e lua, punho, shuriken, cruz, chicote, livro, cetro, mochila, faixa preta, escudo e espada) sobre fundo radial. Nenhum logo, sprite ou ilustração da Gravity. |
| CA-01.1 | ✅ | idem RN-01/RN-05 | `TestArtOf_...` (renegado → arruaceiro); SSR `/perfil` `CA-01.1 / CA-01.5` (src `classes/arruaceiro.png`, alt "Linha do Arruaceiro") | |
| CA-01.2 | ✅ | `toAPICharacter`; diálogo com `bind:value={classId}` | `TestToAPICharacter_CA01_2_...` (feiticeiro → sabio); e2e edita Templário para Feiticeiro e o card mostra "Linha do Sábio" | |
| CA-01.3 | ✅ | `classArt` (gatuno, taekwon, aprendiz → superaprendiz); rótulos "Família Gatuno", "Família Taekwon", "Linha do Superaprendiz" em `portraits.ts` | `TestArtOf_...`; vitest verifica "Família Gatuno" e "Linha do Superaprendiz" | O rótulo "Família Taekwon" não tem asserção no web (só a arte no backend); não bloqueia. |
| CA-01.4 | ✅ | `CharacterDialog.svelte`: bloco `data-testid="class-art"`, sem `name="portrait"` | SSR `CA-01.4` (sem rádio, aviso; na edição mostra a arte); vitest `portraitForClass`; e2e: 0 rádios "Retrato", arte muda ao selecionar Paladino e Renegado | |
| CA-01.5 | ✅ | idem RN-03 | `TestToAPICharacter_CA01_2_CA01_5_ClassPortrait` | |
| CA-01.6 | ✅ | idem RN-04 | `TestArtOf_...`; SSR `/perfil` `CA-01.6`; vitest `CA-01.6 / RN-04` | |
| CA-02.1 | ✅ | `toAPILobby` (dono) e `toAPIParticipant` (membros/pendentes) via `classPortrait` | SSR `lobbies.spec.ts` `CA-02.1 (retrato-por-classe)` (dono Sacerdote, membro Sicário → Mercenário, com alt); API: `TestLobbies_RN22_PublicList` (dono arcebispo → sacerdote), `TestGetLobby_D06_Viewer` (membro → templario); `TestArtOf_...` (sicario → mercenario) | Os testes de API que provam o mapeamento no lobby são testes antigos ajustados e citam os IDs da spec original, não o CA-02.1 desta. Ver observações. |
| CA-02.2 | ✅ | `toAPITalents` usa `ArtOf`; `TalentCard` com alt | SSR `talentos.spec.ts` `CA-02.2 (retrato-por-classe)`; `TestListLobbyTalents_CA02_1` (arquimago → bruxo); `TestArtOf_...` (maestro → bardo) | |
| RNF-01 | ✅ | PNG em `web/static`, servidos pelo próprio web | vitest `RNF-01 / RNF-02 / D-04`; e2e `RNF-02` de /perfil, lobbies e Home (nada carregado de fora) | |
| RNF-02 | ✅ | Maior arte: `mestre-taekwon.png`, 2239 bytes | vitest verifica ≤ 4096 bytes para as 27 e que a pasta tem exatamente as 27 | |
| RNF-03 | ✅ | — | Todos os testes novos citam o CA (`TestArtOf_CA01_1_CA01_3_CA01_6`, `TestToAPICharacter_CA01_2_CA01_5_…`, `CA-01.x`/`CA-02.x (retrato-por-classe)` no vitest/SSR/e2e); os ajustados mantêm os IDs que já citavam | O teste vitest "RN-04: os 4 retratos genéricos continuam" cita só a regra, sem CA; aceitável (RN-04 → CA-01.6). |
| Borda: personagem excluído | ✅ | `classPortrait` devolve vazio quando o retrato gravado é vazio | `TestToAPICharacter_...` (`classPortrait("cardeal","")`); `TestGetLobby_D12_...` (`to.portrait` nulo) | |
| Borda: Super Aprendiz EX / Hiperaprendiz | ✅ | `classArt` | `TestArtOf_...` (hiperaprendiz); superaprendiz-ex pela conferência da tabela | |
| Borda: linha do Ninja | ✅ | `classArt` | `TestArtOf_...` (shiranui) + conferência da tabela | |

## Pendências para correção
Nenhuma.

## Scope creep
Nenhum. Tudo no diff se liga a um ID ou decisão da spec/tasks:
- `scripts/class-art/` — D-04 (gerador para refazer as artes).
- Troca de `alt=""` por `portraitAlt` nas listas — RN-05 (registrado em "Descobertas").
- Itens de "Fora de escopo" (arte por classe dentro da linha, por gênero, escolha de outra linha) não foram implementados.

## Observações (não bloqueantes)
- D-03 pede que o campo `portrait` da entrada fique "obsoleto no contrato", mas `CharacterInput.portrait` em `openapi.yaml` não tem `deprecated: true` nem uma descrição que diga que ele é ignorado.
- `web/src/routes/perfil/+page.server.ts` ainda lê `portrait` do formulário e o manda para a API. Como o diálogo já não envia o campo e a API o ignora, é só um caminho morto.
- O serviço `characters` continua validando `in.Portrait` contra os 4 retratos (`HasPortrait`). Hoje é inofensivo, porque o handler nunca preenche o campo, mas um chamador interno que passe uma arte de classe receberia 422.
- CA-02.1: nenhum teste de API cita o CA-02.1 desta spec nem exercita "Sicário no lobby" ponta a ponta. A cobertura vem da combinação do mapeamento comum (`classPortrait`) testado no lobby com o catálogo testado para `sicario`.
- `scripts/class-art/preview_head.html` carrega Google Fonts. É só a prévia de desenvolvimento, fora do que o web serve, então não fere a RNF-01.
