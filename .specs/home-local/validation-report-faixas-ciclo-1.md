# Relatório de validação — home-local, revisão de 2026-10-09: faixas de horário dinâmicas (ciclo 1)

**Veredito geral:** REPROVADO
**Execução:** testes passou (vitest 350/350; Playwright 52/52) · lint ok (prettier + eslint) · check ok (svelte-check: 0 erros, 0 avisos) · build ok

Escopo: RN-12 revisada, RN-13, CA-04.8 (faixa extra), CA-04.9 e a não regressão de CA-04.4 a CA-04.7.
Alterações: `main..feature/faixas-dinamicas` (commits `62d0270` docs e `d12e93d` feat).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 2 | 0 | 0 | 0 |
| Critérios (CA) | 5 | 1 | 0 | 0 |
| Não funcionais | 0 | 1 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-12 (revisada) | ✅ | `web/src/lib/home/filters.ts`: `TIME_BLOCKS` (12 blocos de 2 h, `start` incluído e `end` excluído), `ALWAYS` com 18-20/20-22/22-24, `visibleTimeRanges(dayLobbies, selected)` filtra blocos por `ALWAYS`, uso no dia ou marcado, na ordem do relógio; `inRange` usa `TIME_BLOCKS`. `web/src/routes/+page.svelte:56` passa `dayLobbies` (sem filtros) e `filters.timeRange`; `FilterPanel.svelte` itera `ranges` | `time-ranges.spec.ts`: "CA-04.8 / RN-12: um lobby às 08:30…", "o início entra e o fim não; a madrugada vira 00h–02h" (01:59 → 00h–02h; 10:00 → 10h–12h), "CA-04.9 / RN-12…"; `home.spec.ts` "RN-12: o fim da faixa é excluído" | "Sem considerar os filtros" é garantido pela assinatura (a função não recebe filtros) |
| RN-13 | ✅ | `filters.ts` `timeRangeCounts`: conta "any" e os 12 blocos sobre os lobbies que passam nos outros filtros (`skipTime`); `FilterPanel.svelte` usa `timeCounts[range.key] ?? 0` | `time-ranges.spec.ts` (`['08-10']` = 1; render com as faixas); `home.spec.ts` CA-04.6 (2 testes) | Contagem das faixas extras com outro filtro ativo não é testada (ver observações) |
| CA-04.4 | ✅ | `filters.ts` `inRange` | `home.spec.ts` "CA-04.4: \"20h–22h\" pega 20:00…" | Sem regressão |
| CA-04.8 (faixa extra) | ✅ | `visibleTimeRanges`, `timeRangeCounts`, `applyFilters` | `time-ranges.spec.ts` "CA-04.8 / RN-12: um lobby às 08:30…" (ordem exata das 5 faixas, contagem 1, filtro devolve só 08:30) e "CA-04.8 / RN-13: o painel mostra as faixas do dia…" (render SSR do `FilterPanel`: 08h–10h antes de 18h–20h, sem 10h–12h) | ID duplicado na spec (ver pendência 1) |
| CA-04.9 | ⚠️ | `visibleTimeRanges` inclui `b.key === selected`; `+page.svelte:42` mantém `filters` ao trocar o dia (não há reset em `selectedDate`); `timeRangeCounts` devolve 0 para blocos sem lobby | `time-ranges.spec.ts` "CA-04.9 / RN-12…": confere só os rótulos (`['Qualquer horário','18h–20h','20h–22h','22h–00h']`) e `labels([], '08-10')` contém 08h–10h | Não há asserção do "mesmo com 0" nem do "visível com 0", e a troca de dia com a faixa marcada não é exercitada (nem unitário de página nem e2e). US-04 é P1 |
| CA-04.5 | ✅ | `filters.ts` `matches` | `home.spec.ts` "CA-04.5: filtros combinados com E" | Sem regressão |
| CA-04.6 | ✅ | `roleCounts`, `timeRangeCounts` | `home.spec.ts` "CA-04.6: \"Vaga para\" ignora os filtros…" (agora `toMatchObject`) e "a contagem das faixas não depende da faixa escolhida" | A troca de `toEqual` para `toMatchObject` era necessária (o objeto tem 13 chaves), mas deixou de conferir as chaves novas |
| CA-04.7 | ✅ | `emptyState`, `resetFilters` | `home.spec.ts` CA-04.7; e2e `home.spec.ts:234` "CA-04.7: filtros sem resultado e \"Limpar filtros\"" | Sem regressão |
| RNF-03 | ⚠️ | — | `home.spec.ts:232` "CA-04.8: subtítulo…" e `time-ranges.spec.ts` "CA-04.8 / …" citam o mesmo ID para critérios diferentes | Rastreabilidade ambígua (pendência 1) |

## Pendências para correção
1. [CA-04.8 / RNF-03] A spec tem dois critérios com o ID CA-04.8: o novo "Faixa extra aparece com grupo" (`.specs/home-local/spec.md:241`) e o já existente "Subtítulo" (`spec.md:271`), citado por `web/src/lib/home/home.spec.ts:232` e por `filters.ts` (`/** CA-04.8: "5 grupos"… */`). O teste e o commit `d12e93d` que citam CA-04.8 ficam ambíguos. É preciso dar um ID livre ao critério novo (ex.: CA-04.10 e CA-04.11, ou renumerar o Subtítulo), atualizar o cabeçalho de revisão, os testes de `time-ranges.spec.ts` e os comentários. Como a spec mudou, isso passa pela revisão do usuário.
2. [CA-04.9] Falta teste para as duas partes do Then que pedem a contagem e a troca de dia: (a) no dia sem lobby antes das 18:00, as faixas fixas aparecem com 0 (conferir `timeRangeCounts` / o `meta` renderizado); (b) com "08h–10h" marcada num dia e o visitante trocando para um dia sem lobby nessa faixa, "08h–10h" continua visível e mostra 0. Hoje o código garante isso (`+page.svelte:42` não reinicia os filtros; `timeRangeCounts` cobre todos os blocos), mas nenhum teste exercita a troca de dia nem confere o 0. Um e2e em `web/test/e2e/home.spec.ts` (marcar a faixa e trocar o dia) ou um teste de render do `FilterPanel` com `timeCounts` de um dia vazio e `ranges = visibleTimeRanges([], '08-10')`, conferindo "08h–10h" com 0, fecha o item.

## Scope creep
- Nenhum. As 6 alterações ligam-se a RN-12, RN-13, CA-04.8 e CA-04.9.

## Observações (não bloqueantes)
- `TimeRangeKey` virou `string` e `inRange` cai em `ANY_TIME` para chave desconhecida: uma chave errada passa a significar "qualquer horário" sem erro de tipo. Um tipo literal (`'any' | \`${string}-${string}\``) ou uma lista de chaves válidas evitaria isso.
- RN-13 com faixa extra e outro filtro ativo (ex.: instância que exclui o lobby das 08:30, então "08h–10h" visível com 0) não tem teste; o comportamento decorre de `timeRangeCounts`, que já é coberto para as faixas fixas.
- O CA-04.6 com `toMatchObject` poderia conferir também uma faixa extra (ex.: `'08-10': 0`) para não perder força.
- A spec diz "Status: Aprovada", e o commit `62d0270` diz "aguardando aprovação". Convém confirmar que a revisão foi aprovada pelo usuário antes do merge.
