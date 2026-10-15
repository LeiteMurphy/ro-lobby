# Relatório de validação — talentos-filtros (ciclo 2)
**Veredito geral:** APROVADO
**Execução:** testes unitários 399/399 passaram (34 arquivos) · ponta a ponta 61/61 passaram (rodada única) · lint OK (prettier + eslint) · tipos OK (svelte-check: 0 erros, 0 avisos) · build OK (saída 0)

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| User Stories | 2 | 0 | 0 | 0 |
| Regras (RN) | 2 | 0 | 0 | 0 |
| Critérios (CA) | 3 | 0 | 0 | 0 |
| Fora de escopo | 2 respeitados | — | — | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| US-01 (P1) | ✅ | ver RN-01 | ver CA-01.1, CA-01.2 | |
| US-02 (P1) | ✅ | ver RN-02 | ver CA-02.1 | A pendência do ciclo 1 foi resolvida. |
| RN-01 | ✅ | `web/src/routes/talentos/+page.svelte:40-45`: `onchange={(e) => e.currentTarget.requestSubmit()}` no `<form method="GET">`. O `change` de qualquer `<select>` (instância, função, servidor, dia, hora) sobe até o form. O botão "Filtrar" (`type="submit"`) segue na linha 99. | e2e `web/test/e2e/talentos.spec.ts:216-219` (função, sem Filtrar, URL `funcao=tank`); `talentos.spec.ts:334-338` (dia e hora); `servidor.spec.ts:188-189` (servidor, URL `servidor=nidhogg`); unitário `src/routes/talentos/talentos.spec.ts:243-246` (GET + Filtrar no SSR) | Nenhum teste troca só a "Instância", mas ela usa o mesmo handler do form. |
| RN-02 | ✅ | `web/src/lib/talents/format.ts:7-16` (`WEEKDAYS` 'Dom'…'Sáb', `WEEKDAY_NAMES` 'Domingo'…'Sábado') e `:33` ('Todos os dias'). Usado em `components/TalentCard.svelte:50` (cards de `/talentos` e "Jogadores disponíveis"), `routes/talentos/+page.svelte:84` (filtro Dia) e `components/AvailabilityDialog.svelte:127` (chips do diálogo). A montagem (faixa de 3+ dias seguidos, lista) não mudou. | ver CA-02.1 | |
| CA-01.1 | ✅ | como em RN-01; mensagem vazia em `+page.svelte:106-108` | e2e `talentos.spec.ts:216-219` (Tank sem Filtrar, `toHaveURL(/funcao=tank/)`, o card de Dano some); e2e `talentos.spec.ts:334-338`; mensagem "Ninguém no banco com esses filtros." em `src/routes/talentos/talentos.spec.ts:172` | O e2e não confere a mensagem vazia porque o banco do e2e é compartilhado; ela vem de teste unitário anterior. |
| CA-01.2 | ✅ | O handler só existe após a hidratação; o HTML do servidor mantém `<form method="GET">` com o submit "Filtrar" (`+page.svelte:40-45, 99`). | `src/routes/talentos/talentos.spec.ts:243-246` | Prova estrutural (sem e2e com `javaScriptEnabled: false`); o envio de form GET é comportamento nativo do navegador. |
| CA-02.1 | ✅ | `format.ts:7-16, 33`; `AvailabilityDialog.svelte:125-127` | Cards: `src/lib/talents/talents.spec.ts:32-38` ('Todos os dias', 'Seg a Sex', 'Seg, Qua, Sex'), `src/routes/lobbies/lobbies.spec.ts:667-669` ("Jogadores disponíveis"), e2e `talentos.spec.ts:330`. Filtro Dia: `src/routes/talentos/talentos.spec.ts:248-252` (Segunda…Domingo, em ordem), e2e `talentos.spec.ts:331`. Diálogo: `src/lib/talents/talents.spec.ts:77-84` extrai os `<span aria-hidden="true">` e confere `['Seg','Ter','Qua','Qui','Sex','Sáb','Dom']`. | Resolvida a pendência do ciclo 1. |

## Pendências para correção
Nenhuma.

## Scope creep
Nenhum.
- `src/routes/lobbies/lobbies.spec.ts` e `test/e2e/servidor.spec.ts` mudaram por consequência direta de RN-02 e RN-01.
- Fora de escopo respeitado: `src/lib/lobbies/share.ts` (convite e preview) continua em minúscula; `src/lib/home/days.ts` (criação) não mudou. A API e o backend não mudaram.

## Observações (não bloqueantes)
- Rastreabilidade: `a9b169d` cita US-01, US-02, RN-01, RN-02, CA-01.1, CA-01.2, CA-02.1; `f0b94ed` cita CA-02.1 e RN-02; `8a2bc57` é só documentação. Os testes novos e alterados citam o CA ou a RN.
- `requestSubmit()` dispara uma navegação por mudança; escolher dia e hora em sequência gera duas navegações, e o e2e mostra que a URL final reúne os dois filtros.
