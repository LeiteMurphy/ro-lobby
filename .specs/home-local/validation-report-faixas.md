# Relatório de validação — home-local, revisão de 2026-10-09: faixas de horário dinâmicas (ciclo 2)

**Veredito geral:** APROVADO
**Execução:** testes passou (vitest 352/352; Playwright 52/52) · lint ok (prettier + eslint) · check ok (svelte-check: 0 erros, 0 avisos) · build ok

Escopo: RN-12 revisada, RN-13, CA-04.9, CA-04.10 e a não regressão de CA-04.4 a CA-04.8.
Alterações: `main..feature/faixas-dinamicas` (commits `62d0270` docs, `d12e93d` feat, `9397db3` fix).

## Pendências do ciclo 1
1. [CA-04.8 / RNF-03] ID duplicado — **resolvida.** Os critérios novos passaram a CA-04.9 ("Faixa extra aparece com grupo", `spec.md:241`) e CA-04.10 ("Faixa extra some sem grupo", `spec.md:248`); o cabeçalho de revisão (`spec.md:6-7`) cita RN-12, RN-13, CA-04.9 e CA-04.10. CA-04.8 voltou a ser só o Subtítulo (`spec.md:271`, `home.spec.ts:232`, `filters.ts:144`). O texto dos critérios não mudou, só os IDs.
2. [CA-04.10, antes CA-04.9] Teste do "mesmo com 0" e da faixa marcada — **resolvida.** `time-ranges.spec.ts` ganhou dois testes de render do `FilterPanel` que leem o número do `<span class="meta">` de cada faixa (ver detalhe).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 2 | 0 | 0 | 0 |
| Critérios (CA) | 7 | 0 | 0 | 0 |
| Não funcionais | 1 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-12 (revisada) | ✅ | `web/src/lib/home/filters.ts`: `TIME_BLOCKS` (12 blocos de 2 h, `start` incluído e `end` excluído), `ALWAYS` = 18-20/20-22/22-24, `visibleTimeRanges(dayLobbies, selected)` mantém os blocos em `ALWAYS`, com lobby no dia ou marcados, na ordem do relógio; `inRange` usa `TIME_BLOCKS`. `web/src/routes/+page.svelte:56` passa `dayLobbies` (antes dos filtros) e `filters.timeRange`; as duas instâncias do `FilterPanel` (barra e gaveta) recebem `ranges` | `time-ranges.spec.ts`: "CA-04.9 / RN-12: um lobby às 08:30…", "o início entra e o fim não; a madrugada vira 00h–02h", os três testes de CA-04.10; `home.spec.ts` CA-04.4 | "Sem considerar os filtros" é garantido pela assinatura: a função não recebe filtros |
| RN-13 | ✅ | `timeRangeCounts` conta "any" e os 12 blocos sobre os lobbies que passam nos outros filtros (`skipTime`); `FilterPanel.svelte` usa `timeCounts[range.key] ?? 0` | `time-ranges.spec.ts` (`'08-10'` = 1; render com contagens 0); `home.spec.ts` CA-04.6 | |
| CA-04.4 | ✅ | `inRange` | `home.spec.ts` "CA-04.4: \"20h–22h\" pega 20:00…" | Sem regressão |
| CA-04.5 | ✅ | `matches` | `home.spec.ts` "CA-04.5: filtros combinados com E" | Sem regressão |
| CA-04.6 | ✅ | `roleCounts`, `timeRangeCounts` | `home.spec.ts` "CA-04.6: \"Vaga para\" ignora os filtros…" e "a contagem das faixas não depende da faixa escolhida" | `toMatchObject` não confere as chaves novas (observação) |
| CA-04.7 | ✅ | `emptyState`, `resetFilters` | `home.spec.ts` CA-04.7; e2e "CA-04.7: filtros sem resultado e \"Limpar filtros\"" | Sem regressão |
| CA-04.8 (Subtítulo) | ✅ | `filters.ts:144` | `home.spec.ts:232` "CA-04.8: subtítulo \"2 de 5 grupos\"…" | Sem regressão; o ID agora é único |
| CA-04.9 | ✅ | `visibleTimeRanges`, `timeRangeCounts`, `applyFilters` | `time-ranges.spec.ts` "CA-04.9 / RN-12: um lobby às 08:30…" (ordem exata das 5 faixas, contagem 1, filtro devolve só 08:30) e "CA-04.9 / RN-13: o painel mostra as faixas do dia…" (render: 08h–10h antes de 18h–20h, sem 10h–12h) | |
| CA-04.10 | ✅ | `visibleTimeRanges` inclui `b.key === selected`; `+page.svelte:42/91`: trocar o dia só muda `selectedDate`, `filters` não é reiniciado; `timeRangeCounts` devolve 0 para bloco sem lobby | `time-ranges.spec.ts` "CA-04.10 / RN-12: sem lobby antes das 18h, só as fixas…" (lista exata das 4), "…as fixas aparecem com 0" (meta = `0` em 18h–20h, 20h–22h e 22h–00h; sem 08h–10h), "…com \"08h–10h\" marcada, num dia sem grupo nela, ela fica visível com 0" (meta = `0` e radio `checked`) | A troca de dia é testada pelo estado que ela produz (filtro marcado e dia sem grupo na faixa), não pelo clique; o código não reinicia os filtros ao trocar o dia (observação) |
| RNF-03 | ✅ | — | Todos os testes novos citam CA-04.9 ou CA-04.10, e os IDs são únicos na spec | |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. As alterações ligam-se a RN-12, RN-13, CA-04.9 e CA-04.10; o `validation-report-faixas-ciclo-1.md` é artefato do processo.

## Observações (não bloqueantes)
- CA-04.10 não tem e2e que marque "08h–10h" e troque o dia no seletor. O teste de render cobre o estado final, e `+page.svelte` não reinicia `filters` em `selectedDate`, mas uma mudança futura que reinicie os filtros ao trocar o dia não seria pega por nenhum teste.
- `TimeRangeKey` virou `string`, e `inRange` cai em `ANY_TIME` para chave desconhecida: uma chave errada passa a valer "qualquer horário" sem erro de tipo. (Mantida do ciclo 1.)
- RN-13 com faixa extra e outro filtro ativo (ex.: instância que exclui o lobby das 08:30) não tem teste próprio; decorre de `timeRangeCounts`, que é coberto para as faixas fixas. (Mantida do ciclo 1.)
- CA-04.6 com `toMatchObject` poderia conferir também uma faixa extra (ex.: `'08-10': 0`). (Mantida do ciclo 1.)
- A spec diz "Status: Aprovada" e o commit `62d0270` diz "aguardando aprovação"; o `9397db3` também alterou a spec (renumeração). Como a spec mudou, confirme com o usuário que a revisão e a renumeração estão aprovadas antes do merge.
- Na spec, CA-04.9 e CA-04.10 ficam entre CA-04.4 e CA-04.5. É só a ordem no texto; os IDs são únicos.
