# Relatório de validação — talentos-filtros (ciclo 1)
**Veredito geral:** REPROVADO
**Execução:** testes unitários 398/398 passaram (34 arquivos) · ponta a ponta 61/61 passaram (na primeira rodada, 1 falha em `perfil.spec.ts:178` por `net::ERR_NO_BUFFER_SPACE` no `page.goto`, erro de rede do Windows sem relação com a mudança; a segunda rodada passou inteira) · lint OK (prettier + eslint) · tipos OK (svelte-check: 0 erros, 0 avisos) · build OK (saída 0)

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| User Stories | 1 | 1 | 0 | 0 |
| Regras (RN) | 1 | 1 | 0 | 0 |
| Critérios (CA) | 2 | 1 | 0 | 0 |
| Fora de escopo | 2 respeitados | — | — | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| US-01 (P1) | ✅ | ver RN-01 | ver CA-01.1, CA-01.2 | |
| US-02 (P1) | ⚠️ | ver RN-02 | ver CA-02.1 | Falta teste do texto visível do diálogo. |
| RN-01 | ✅ | `web/src/routes/talentos/+page.svelte:40-45`: `onchange={(e) => e.currentTarget.requestSubmit()}` no `<form method="GET">`; o evento `change` sobe de qualquer `<select>` (instância, função, servidor, dia, hora). O botão "Filtrar" (`type="submit"`) segue na linha 98. | `web/test/e2e/talentos.spec.ts:216-218` (função, sem clicar em Filtrar, URL com `funcao=tank`); `talentos.spec.ts:309-339` (dia e hora); `servidor.spec.ts:188-189` (servidor, URL com `servidor=nidhogg`); `src/routes/talentos/talentos.spec.ts:243-246` (GET + Filtrar no SSR) | Não há teste que troque só a "Instância" sem Filtrar. Como o handler fica no `<form>`, ela tem o mesmo comportamento. |
| RN-02 | ⚠️ | `web/src/lib/talents/format.ts:7-16` (`WEEKDAYS` 'Dom'…'Sáb', `WEEKDAY_NAMES` 'Domingo'…'Sábado') e `:33` ('Todos os dias'). Usado em `TalentCard.svelte:50` (cards de `/talentos` e de "Jogadores disponíveis"), em `talentos/+page.svelte:84` (filtro Dia) e em `AvailabilityDialog.svelte:127` (chips do diálogo). As regras de montagem não mudaram. | Cards e filtro cobertos (ver CA-02.1). Diálogo sem teste do texto visível. | |
| CA-01.1 | ✅ | como em RN-01 | e2e `talentos.spec.ts:216-219`: escolhe Tank sem Filtrar, depois `toHaveURL(/funcao=tank/)` e o card de Dano some; e2e `talentos.spec.ts:333-338`: dia e hora sem Filtrar, a URL muda e o card some. A mensagem "Ninguém no banco com esses filtros." (lista filtrada vazia) está coberta em `src/routes/talentos/talentos.spec.ts:172`. | O e2e não confere a mensagem vazia porque o banco do e2e é compartilhado. Ela vem de um teste unitário que já existia. |
| CA-01.2 | ✅ | O handler só existe depois da hidratação. O HTML do servidor mantém `<form method="GET">` com o botão submit "Filtrar" (`+page.svelte:40-45, 98`). | `src/routes/talentos/talentos.spec.ts:243-246` (form GET e botão Filtrar no SSR) | A prova é estrutural: não há e2e com `javaScriptEnabled: false`. Não bloqueia, porque o envio de um form GET nativo é comportamento do navegador. |
| CA-02.1 | ⚠️ | `format.ts:7-16, 33`; `AvailabilityDialog.svelte:125-127` | Cards: `src/lib/talents/talents.spec.ts:32-38` ('Todos os dias', 'Seg a Sex', 'Seg, Qua, Sex'), `src/routes/lobbies/lobbies.spec.ts:667-669` ("Jogadores disponíveis"), e2e `talentos.spec.ts:330`. Filtro Dia: `src/routes/talentos/talentos.spec.ts:248-252` (Segunda…Domingo, nessa ordem) e e2e `talentos.spec.ts:331`. **Diálogo: nenhum teste confere "Seg" a "Dom".** | O e2e `talentos.spec.ts:66, 85-86` usa `getByLabel('Segunda')`, que lê o `aria-label` (`WEEKDAY_NAMES`), não o texto visível do chip (`<span aria-hidden>{WEEKDAYS[d]}</span>`). Os testes do diálogo em `talents.spec.ts:70+` também não conferem os rótulos. O código está certo, mas falta o teste do terceiro "And". |

## Pendências para correção
1. **CA-02.1 / RN-02 (diálogo de disponibilidade):** falta um teste citando CA-02.1 que confira os rótulos visíveis dos chips, "Seg", "Ter", "Qua", "Qui", "Sex", "Sáb", "Dom", nessa ordem. Por exemplo, em `web/src/lib/talents/talents.spec.ts`, renderizar `AvailabilityDialog` e extrair os `<span aria-hidden="true">` dentro de `.chips`. O código já cumpre a regra; falta só a evidência.

## Scope creep
Nenhum.
- `src/routes/lobbies/lobbies.spec.ts` e `test/e2e/servidor.spec.ts` mudaram por consequência direta: "Jogadores disponíveis" usa `daysLabel` (RN-02), e o filtro de servidor passou a aplicar sem Filtrar (RN-01).
- Fora de escopo respeitado: `src/lib/lobbies/share.ts` (convite e preview) continua em minúscula, e `src/lib/home/days.ts` (criação) não mudou. A API e o backend não mudaram.

## Observações (não bloqueantes)
- Rastreabilidade: o commit `a9b169d` cita US-01, US-02, RN-01, RN-02, CA-01.1, CA-01.2 e CA-02.1. O commit `8a2bc57` é só de documentação. Os testes novos e os alterados citam o CA ou a RN.
- `requestSubmit()` dispara uma navegação a cada mudança. Com um único select isso não é problema. Escolher dia e hora em sequência gera duas navegações, mas o e2e mostra que a URL final reúne os dois filtros.
- A falha de rede do e2e (`ERR_NO_BUFFER_SPACE`) foi intermitente e do ambiente, e não se repetiu.
