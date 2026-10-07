# Relatório de validação — candidatura-lobby, ajustes depois do merge da Parte 2 (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes passou (backend `go test ./...` ok; integração `-tags=integration` ok; Vitest 279/279; Playwright 45/45) · lint ok (`gofmt -l` vazio, golangci-lint 0 issues, Prettier ok, ESLint ok) · build ok (`svelte-kit sync` + `svelte-check --threshold error`: 0 erros, 0 avisos)

Alterações conferidas: `origin/main..feature/ajustes-candidatura` (4 commits: cabb9f9, fe8a878, 59d252b, 170b8ce).
O backend não mudou nesta branch; os resultados de Go saíram do cache, o que é coerente com o diff.

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 2 (RN-39, RN-31) | 0 | 0 | 0 |
| Critérios (CA) | 2 (CA-08.15, CA-10.3) | 0 | 0 | 0 |
| Não funcionais | regressão das Partes 1 e 2: ✅ | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-39 | ✅ | `web/src/routes/lobbies/[id]/+page.svelte:114-121` (`rejectedSwap`: membro `accepted`, `lobby.status === 'open'`, pedido `rejected` com `decisionReason`) e `:255-266` (texto "Seu pedido de troca foi recusado. Justificativa: …" junto de "Você está no grupo", com "Pedir troca" logo abaixo em `:267-271`). "Mais recente": `backend/queries/swap_requests.sql:30-35` (`GetLatestSwapRequest`, `ORDER BY created_at DESC, id DESC LIMIT 1`) usado em `backend/internal/lobbies/detail.go:178-186` só dentro de `MyApplication` (dado do próprio usuário, não sai para terceiros). Um pedido pendente mais novo cai no ramo `pendingSwap`, que vem antes. | Vitest `lobbies.spec.ts` "CA-08.15 / RN-39: o membro vê a justificativa…" (mostra com recusado; não mostra com `withdrawn`; não mostra com lobby `started`; "Pedir troca" presente). Playwright `candidatura-parte2.spec.ts` "CA-08.14 / CA-08.15 / CA-07.1" (pedido novo depois de um aceito, dono recusa com justificativa, membro vê o texto exato e continua com o personagem atual). | Restrição "só dono e membro" garantida pelo backend existente (Parte 2); nenhum teste novo de backend, mas não houve mudança de backend. |
| CA-08.15 | ✅ | mesmo de RN-39 | Playwright: `toHaveText('Seu pedido de troca foi recusado. Justificativa: Precisamos de você no tank')` + `Você está no grupo com Bri… (Tank)`; Vitest confere "Pedir troca" | O passo do Playwright não confere o botão "Pedir troca" depois da recusa (o Vitest confere). |
| RN-31 / CA-10.3 (texto "· pendente desde") | ✅ | `web/src/lib/applications/components/PlayerPanel.svelte:51-55` (expressão em vez de `{#if}` quebrado, o espaço não se perde) | Vitest `applications.spec.ts` "CA-10.3: candidato para o dono…": `toContain('Dano · pendente desde 17:10')` sobre o HTML sem comentários (a regex antiga `[\s\S]{0,20}` aceitava o texto sem espaço) | |
| Regressão Partes 1 e 2 | ✅ | diff só toca os trechos acima | Vitest 279/279, Playwright 45/45, integração Go ok | |
| Design | ✅ | sem mudança de contrato nem de modelo; usa `myApplication.swapRequest.decisionReason` já previsto (design D-08/D-12, coluna `decision_reason`) | — | |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. A mudança do "Estado atual" do `CLAUDE.md` (commit 170b8ce) é só documentação e não toca código nem itens de "Fora de escopo".

## Observações (não bloqueantes)
- Rastreabilidade ok: os 4 commits citam IDs (RN-39, CA-08.15, US-08, RN-31, CA-10.3); T-18 e T-19 marcadas `[x]` com commits que existem no intervalo; as duas Descobertas da Parte 2 foram atualizadas com a decisão do usuário. Os links do Notion não foram conferidos.
- RN-39 diz que a justificativa "aparece só para o dono e para o membro". O dono não tem onde rever a recusa depois de decidir (o pedido sai de "Pedidos de troca"). Lido como restrição de visibilidade, está cumprido; se a intenção era o dono também ver depois, falta deixar isso claro na spec.
- `rejectedSwap` exige `decisionReason` não vazio. Pela RN-22 a recusa sempre tem justificativa, então não muda nada hoje, mas uma recusa sem texto (dado antigo ou erro) não mostraria aviso nenhum.
- Não há teste de tela para o caso "recusado e depois um pedido pendente mais novo"; a garantia vem da query `GetLatestSwapRequest` e da ordem dos ramos no template.
