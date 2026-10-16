# Relatório de validação — discord-em-construcao (ciclo 1)

**Veredito geral:** APROVADO
**Execução:** testes passou (465/465, vitest) · e2e passou (68/68, Playwright, chave desligada) · lint ok · tipos ok (0 erros, 0 avisos) · build ok · CI do PR ok (Web, Ponta a ponta, Pilha app, CI ok; Backend pulado pelo filtro de caminhos)
**Compose:** `docker compose --profile app config` resolve `DISCORD_UNDER_CONSTRUCTION=true` no serviço `web` (só leitura, sem mexer nos containers).

## Resumo
| Tipo | ✅ | ⚠️ | ❌ | 🚫 |
|---|---|---|---|---|
| Regras (RN) | 7 | 0 | 0 | 0 |
| Critérios (CA) | 5 | 0 | 0 | 0 |
| Não funcionais | 1 | 0 | 0 | 0 |

## Detalhe por item
| ID | Veredito | Evidência (código) | Evidência (teste) | Observação |
|---|---|---|---|---|
| RN-01 | ✅ | web/src/lib/discord/construction.ts:6 (`=== 'true'`); docker-compose.yml:130 (`${DISCORD_UNDER_CONSTRUCTION:-true}`); .env.example | construcao.spec.ts e construcao.server.spec.ts (chave `true`); e2e sem a chave; `compose config` resolve `true` no web | Dev e e2e ficam sem a chave, como pede a regra. |
| RN-02 | ✅ | web/src/routes/perfil/+page.svelte:125-140; perfil/+page.server.ts:70; messages.ts `UNDER_CONSTRUCTION.profile` | page.spec.ts "CA-01.1 / RN-02" (conectado e não conectado) | Substitui "Entrar..." e "Reconectar". |
| RN-03 | ✅ | DiscordSection.svelte:52-55; lobbies/[id]/+page.server.ts:73; +page.svelte:551 | section.spec.ts "CA-01.2 / RN-03 / RN-04" (sala nula e aberta) | Sala já aberta também fica escondida (caso de borda). |
| RN-04 | ✅ | DiscordSection.svelte:140 (owner-link fora do ramo da sala) | section.spec.ts asserta `data-testid="owner-link"` e o href | |
| RN-05 | ✅ | perfil/discord/+server.ts:22; perfil/discord/retorno/+server.ts:13-16; messages.ts `CONNECT_MESSAGES.construcao` | construcao.spec.ts (dois CA-01.3 + RN-05) | Sem cookie de state criado, sem fetch à API; o state antigo é apagado. |
| RN-06 | ✅ | lobbies/[id]/+page.server.ts:128-130, 135, 153 | construcao.server.spec.ts "CA-01.4" (openRoom e alert, 0 chamadas) | Recusa com 409 e a frase da regra. |
| RN-07 | ✅ | `discordUnderConstruction()` falso mantém os ramos antigos | section.spec.ts e page.spec.ts "CA-01.5 / RN-07"; e2e discord.spec.ts passa sem a chave | |
| CA-01.1 | ✅ | perfil/+page.svelte:125-140 | page.spec.ts: frase, `<button ... disabled>Em construção`, sem `href="/perfil/discord"` | |
| CA-01.2 | ✅ | DiscordSection.svelte:52-55 | section.spec.ts: aviso, botão desabilitado, sem `?/openRoom`, `?/alert`, room-link, "Quem está na sala", `/perfil/discord`; owner-link presente | |
| CA-01.3 | ✅ | +server.ts:22; retorno/+server.ts:13 | construcao.spec.ts: redireciona a `/perfil?discord=construcao`, sem chamadas | O /perfil exibe `discordMessage` fora do bloco `discord.enabled` (+page.svelte:109). |
| CA-01.4 | ✅ | +page.server.ts:135, 153 | construcao.server.spec.ts | |
| CA-01.5 | ✅ | ramos `{:else}` intactos | e2e 68/68 local e no CI com a chave desligada; testes unitários com `false` | |
| RNF-01 | ✅ | Button.svelte usa o atributo nativo `disabled` no `<button>` | page.spec.ts e section.spec.ts assertam `disabled` no `<button>` | Atributo nativo é anunciado como indisponível por leitor de tela. |

## Pendências para correção
Nenhuma.

## Scope creep
- Nenhum. O diff só toca o web, o Compose e o `.env.example`; a API e o bot não mudaram (Fora de escopo respeitado). O commit 385c955 (.specs/endurecimento/tasks.md) é registro da T-12 de outra feature, desconsiderado conforme instrução.

## Observações (não bloqueantes)
- Não há teste unitário direto de `discordUnderConstruction()` para valores como `TRUE`, `1` ou vazio; o comportamento (só `true` liga) está no código e o caso vazio é coberto pelo e2e.
- Não há e2e com a chave ligada; a cobertura vem de testes de renderização SSR e das actions/rotas, o que atende os CAs.
- Em construção, `/perfil/discord` redireciona antes de checar a sessão; um visitante sem login vai ao `/perfil`, que já leva ao login. Comportamento aceitável.
- Commits citam os IDs (US-01, RN-01..RN-07, CA-01.1..CA-01.5).
