# Relatório de validação — compartilhar-lobby (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes unitários passou (290/290, 25 arquivos) · e2e passou (47/47, incluindo os 2 de `compartilhar.spec.ts`) · lint ok (prettier + eslint) · check ok (0 erros, 0 avisos) · build ok (exit 0)

Intervalo validado: `main..feature/compartilhar-lobby` (commits `38ca6b5` e `d5da5f8`).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 6 | 0 | 0 | 0 |
| Critérios (CA) | 7 | 0 | 0 | 0 |
| Não funcionais | n/a | | | |

O CA-02.3 é UAT manual e fica pendente de UAT, fora da contagem (ver abaixo). A spec não
tem requisitos não funcionais próprios.

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | `web/src/routes/lobbies/[id]/+page.svelte:205-208` (`ShareButton` dentro de `{#if lobby.status === 'open'}` e antes dos ramos por `viewerState`) | `lobbies.spec.ts` "CA-01.1…", "CA-01.2…" | O botão não depende de sessão nem de papel. |
| RN-02 | ✅ | `web/src/lib/lobbies/share.ts:11-16` (`whenLabel`, data absoluta em São Paulo via `fromUtcIso`), `:33-39` (`inviteText`), origem de `url.origin` em `+page.server.ts:63` | `share.spec.ts` "CA-01.3: copia o convite…", "CA-01.3 / RN-02: a data é absoluta e em Brasília…"; e2e `compartilhar.spec.ts` CA-01.3 | Virada de dia UTC → SP coberta (02:30 UTC de domingo = sábado 23:30). |
| RN-03 | ✅ | `share.ts:19-25` (`openSlotsLabel`, ordem tank/support/dps, filtra n > 0, "Grupo lotado") | `share.spec.ts` "CA-01.3 / RN-03…", "CA-01.4…"; e2e CA-01.3 ("1 Tank, 1 Suporte, 3 Dano") | |
| RN-04 | ✅ | `ShareButton.svelte:20-32` (`clipboard.writeText`, "Convite copiado" por 3 s, fallback em `<dialog>` com `textarea.select()`) | e2e CA-01.3 (botão vira "Convite copiado"), e2e CA-01.5 (diálogo, foco e seleção completa) | A volta para "Compartilhar" depois dos 3 s não é testada (não bloqueante). |
| RN-05 | ✅ | `+page.svelte:166-175` (og:title, og:description, og:url, og:site_name, og:type, theme-color); `share.ts:51-62` (`shareMeta`, três descrições por status); `THEME_COLOR = '#f6bb45'` = `--gold-400` em `src/lib/ui/tokens/colors.css:21` | `lobbies.spec.ts` "CA-02.1…", "CA-02.2…"; `share.spec.ts` CA-02.1, CA-02.2, "CA-02.2 / RN-05: lobby iniciado…" | Sem `og:image`, conforme a spec (asserção `not.toContain('og:image')`). |
| RN-06 | ✅ | Nenhuma imagem no convite nem no preview; o único ícone novo é o Lucide `share-2` (`Icon.svelte`) | `lobbies.spec.ts` CA-02.1 (sem `og:image`) | |
| CA-01.1 | ✅ | `+page.svelte:207` | `lobbies.spec.ts` "CA-01.1: visitante, candidato e dono veem…" (renderiza com `null`, outra pessoa e dono) | O "candidato" do teste é um usuário logado sem candidatura; o código garante o caso porque o botão fica fora dos ramos de `viewerState`. |
| CA-01.2 | ✅ | `+page.svelte:205` | `lobbies.spec.ts` "CA-01.2…" (started e cancelled, `not.toContain('Compartilhar')`) | |
| CA-01.3 | ✅ | `share.ts:33-39`, `ShareButton.svelte:22-25` | `share.spec.ts` com o exemplo exato da spec (Glast Heim, sábado 10/10 20:00, 160, "Vagas: 1 Tank, 2 Dano"); e2e lê a área de transferência real e confere o texto e "Convite copiado" | |
| CA-01.4 | ✅ | `share.ts:24` | `share.spec.ts` "CA-01.4…" (segunda linha = "Grupo lotado · Nível mínimo 160") | |
| CA-01.5 | ✅ | `ShareButton.svelte:26-31`, `:43-67` | e2e "CA-01.5…" (`writeText` rejeita com NotAllowedError; diálogo visível, texto do convite, textarea focada e toda selecionada) | |
| CA-02.1 | ✅ | `+page.svelte:166-175`, `share.ts:51-62` | `lobbies.spec.ts` "CA-02.1: o HTML sem sessão traz as meta tags…" (SSR com `user: null`, confere os seis valores) | |
| CA-02.2 | ✅ | `share.ts:53-54` | `lobbies.spec.ts` "CA-02.2…" (SSR, og:description = "Esse grupo foi cancelado.") | |
| CA-02.3 | Pendente de UAT | — | — | UAT manual com o Discord real em `rolobby.com.br`; não reprova este ciclo. |

## Pendências para correção
Nenhuma.

Pendente de UAT (não bloqueante):
1. [CA-02.3] Colar o link de um lobby no Discord com o site publicado em `rolobby.com.br` e conferir o cartão com título e descrição.

## Scope creep
- Nenhum. `Icon.svelte` (ícone `share-2`) apoia o botão da RN-01, e as mudanças em
  `lobbies.server.spec.ts` só adaptam os testes existentes ao novo parâmetro `url` do
  `load`. Nada toca "Fora de escopo" (sem imagem no preview, sem botão no card da Home,
  sem Web Share API, sem encurtador).

## Observações (não bloqueantes)
- RN-04: a volta do rótulo para "Compartilhar" depois de 3 s não tem teste; um teste com
  fake timers fecharia o caso.
- CA-01.1: o teste não monta um candidato de verdade (com candidatura); a garantia vem da
  posição do botão no template.
- O e2e de CA-01.3 usa outro lobby (Templo do Demônio Rei, dia 4) e não o exemplo literal
  da spec; o exemplo literal está coberto no unitário `share.spec.ts`.
- `og:url` e o link do convite usam `url.origin`, que no adapter-node vem de `ORIGIN`. Sem
  a publicação por túnel, o convite sai com `http://localhost:3000`, como a spec prevê.
- Rastreabilidade: os commits citam US e CA em Conventional Commits; os testes citam os IDs.
