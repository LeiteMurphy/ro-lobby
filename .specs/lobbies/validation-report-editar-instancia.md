# Relatório de validação — lobbies, revisão de 2026-10-09: a edição troca a instância (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes backend passaram (todos os pacotes `ok`, `-tags=integration -count=1`) · web vitest 292/292 · e2e Playwright 47/47 · lint backend 0 issues · lint web ok (prettier + eslint) · check 0 erros, 0 avisos · build ok

Escopo: só a revisão de 2026-10-09 (RN-17 revisada, CA-04.5, CA-04.6, o caso de borda do membro
abaixo do nível da nova instância) e a não regressão de CA-04.1 a CA-04.4, RN-07 e RN-18 na
edição. Alterações: `main..feature/editar-instancia` (commits `3cb6a6f` e `e782a21`).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 3 | 0 | 0 | 0 |
| Critérios (CA) | 5 | 1 | 0 | 0 |
| Casos de borda | 0 | 1 | 0 | 0 |
| Não funcionais (RNF-03, RNF-04) | 2 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-17 (revisada) | ✅ | `backend/internal/lobbies/lobbies.go` (Update): instância nova só se estiver no catálogo (`FieldInstanceID`/`CodeInvalid`); sem `instanceId` ou com a mesma, vale a gravada, mesmo fora do catálogo; `UpdateLobby` grava `instance_id`, `instance_name` e `instance_level` (`backend/queries/lobbies.sql:68-71`); `openapi.yaml` com `instanceId` opcional em `LobbyUpdate`; handler `backend/internal/server/lobbies.go:126`; web: `LobbyForm.svelte` com o select também em `update`, `onInstanceChange` sugere o nível de entrada; `editar/+page.server.ts` carrega o catálogo e mantém a instância atual se ela saiu dele | `TestUpdate_CA04_5_Instance`, `TestGet_RN01_InstanceLeftCatalog` (edição mantendo instância removida), `TestUpdate_CA04_5_InstanceKeepsMembers`; vitest "CA-04.5 / RN-17: a edição traz o catálogo…", "RN-17: a instância atual que saiu do catálogo continua na lista", "CA-04.5 / RN-17: na edição, a instância muda e o personagem fica fixo"; e2e `lobbies.spec.ts:192` | O personagem continua fixo (`disabled`), verificado no teste de render. Regra garantida na API (RNF-04). |
| RN-07 (na edição) | ✅ | `lobbies.go` Update: `in.MinLevel < int(instanceLevel)` usa o nível da instância nova | `TestUpdate_CA04_6_InstanceAboveOwner` (troca para 240 com mínimo 160 → `CodeInvalid`); `TestUpdate_CA04_2_CA04_3_Limits` (sem regressão) | |
| RN-18 (na edição) | ✅ | `lobbies.go` Update: `CodeAboveOwner` quando o mínimo passa do nível do dono; checagem de vagas por ocupantes inalterada | `TestUpdate_CA04_6_InstanceAboveOwner`, `TestUpdate_CA04_2_CA04_3_Limits` | |
| CA-04.1 | ✅ | sem mudança de comportamento; o corpo do PUT agora inclui `instanceId` igual ao atual | e2e `lobbies.spec.ts:192` (21:00 e 1/8 na Home); vitest `/lobbies/[id]/editar` com o corpo esperado incluindo `instanceId: 'templo-do-demonio-rei'` | |
| CA-04.2 | ✅ | inalterado | `TestUpdate_CA04_2_CA04_3_Limits`; vitest do erro nas vagas | |
| CA-04.3 | ✅ | inalterado; agora compara com o nível da instância resolvida | `TestUpdate_CA04_2_CA04_3_Limits` | |
| CA-04.4 | ✅ | inalterado (`GetOwnLobbyForUpdate` antes de resolver a instância) | `TestUpdate_CA04_4_OtherOrNotOpen` | |
| CA-04.5 | ⚠️ | ver RN-17 | Formulário sugere 120: e2e `selectOption('sonho-sombrio')` → `Nível mínimo` = `120`. Detalhe mostra "Sonho Sombrio": e2e (h1). Membro aceito e pendente continuam: `TestUpdate_CA04_5_InstanceKeepsMembers` (status `accepted`/`pending`, ocupação e pendentes conferidos) | O Then pede que **a Home** também mostre "Sonho Sombrio"; nenhum teste confere isso. Pelo código, a listagem usa `l.InstanceName` gravado (`server/lobbies.go:209`), então deve funcionar, mas falta a prova. US-04 é P2. |
| CA-04.6 | ✅ | `lobbies.go` Update: verificação de `CodeAboveOwner` antes do `UpdateLobby`, dentro da transação | `TestUpdate_CA04_6_InstanceAboveOwner`: erro em `minLevel` (`CodeAboveOwner` com 240, `CodeInvalid` com 160) e `Get` confirma `templo-do-demonio-rei` e mínimo 160 | No web, o erro em `minLevel` usa o mapeamento já existente do CA-04.3. |
| Borda: membro abaixo do nível da nova instância continua | ⚠️ | `Update` não mexe em candidaturas nem compara nível de membros, então o comportamento vale pelo código | Nenhum teste. `TestUpdate_CA04_5_InstanceKeepsMembers` troca de Templo (160) para Sonho Sombrio (120), e o membro "Cura" tem nível 200: o membro fica **acima** da nova instância, então o caso de borda não é exercitado. A parte "o nível mínimo vale para candidaturas novas" também não é testada após a troca. | |
| RNF-03 | ✅ | — | Todos os testes novos citam CA-04.5, CA-04.6 ou RN-17 | |
| RNF-04 | ✅ | catálogo, RN-07 e RN-18 validados no serviço | testes de integração do serviço | |

## Pendências para correção
Nenhuma bloqueante. Itens ⚠️ (US-04 é P2), recomendados:
1. [CA-04.5] Conferir que a Home mostra "Sonho Sombrio" depois da troca (por exemplo, no e2e `web/test/e2e/lobbies.spec.ts`, voltar à Home e checar o card, como já se faz para o CA-04.1).
2. [Caso de borda, RN-17/RN-18] Teste com membro aceito de nível abaixo do de entrada da nova instância (dono com nível suficiente e uma instância nova acima do membro), conferindo que ele continua aceito e que uma candidatura nova abaixo do novo mínimo é recusada.

## Scope creep
- Nenhum. A remoção de `fixedInstance` e da classe `.fixed` em `LobbyForm.svelte` e a troca do texto de ajuda em `editar/+page.svelte` decorrem da RN-17 revisada.

## Observações (não bloqueantes)
- Na edição, o select ainda tem a opção vazia "Escolha a instância". Se o dono a escolher, o web manda `instanceId: ""` e a API mantém a instância atual sem avisar (contrato: "sem ele, a instância continua"). Não fere a spec, mas pode confundir; dá para tirar a opção vazia no modo `update` ou tratar `""` como erro no web.
- Commits citam os IDs (`[RN-17, CA-04.5, CA-04.6]`).

## Correções depois do ciclo 1
As pendências ⚠️ foram tratadas sem mudar o veredito:
1. [CA-04.5] O e2e `lobbies.spec.ts` volta à Home e confere "Sonho Sombrio" no card.
2. [Borda RN-17/RN-18] `TestUpdate_RN17_MemberBelowNewInstance`: o membro de nível 170
   continua aceito depois da troca para uma instância de nível 180, e uma candidatura nova
   de nível 175 é recusada com `below_min_level`.
3. [Observação] Na edição, o seletor não tem mais a opção vazia "Escolha a instância".

Execução depois das correções: integração do backend ok, vitest 292/292, Playwright 47/47,
lint e check ok.
