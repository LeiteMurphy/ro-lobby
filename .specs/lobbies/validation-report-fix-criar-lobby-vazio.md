# Relatório de validação — lobbies, correção do Criar lobby (ciclo 1)

**Veredito geral: APROVADO**

- Spec: `.specs/lobbies/spec.md`, itens RN-23, CA-02.3 e RN-04
- Alterações: `main..fix/criar-lobby-vazio`, 1 commit (`35de64f fix(web): "Criar lobby" do aviso de lista vazia leva à criação [RN-23, CA-02.3]`), 3 arquivos, +23/−1
- Nível: P (correção pontual)
- Validado em 2026-10-06

## Execução

Comandos rodados em `web/`, com o resultado real:

| Comando | Resultado |
|---|---|
| `npm test` (vitest) | 19 arquivos, 194 testes, todos passaram |
| `npm run lint` (prettier + eslint) | sem erros ("All matched files use Prettier code style!", eslint sem saída) |
| `npm run check` (svelte-check) | 487 arquivos, 0 erros, 0 avisos |
| `npx playwright test home` | 14 testes, 14 passaram (22,0 s) |

## Itens

| ID | Veredito | Evidência no código | Evidência no teste | Observação |
|---|---|---|---|---|
| RN-23 (todo "Criar lobby" da Home leva a `/lobbies/novo`) | ✅ | `web/src/lib/home/components/EmptyState.svelte:31`: o botão deixou de ser `soon` e virou `<Button iconLeft="plus" href={resolve('/lobbies/novo')}>`, sem condição por `kind`, então vale para `day` e para `filters`. Os outros dois já existiam: `TopBar.svelte:26` (desktop) e `TopBar.svelte:29-34` (`IconButton` no celular), ambos com `href={resolve('/lobbies/novo')}`. Não sobrou nenhum "Criar lobby" com `soon` em `web/src` fora dos exemplos de `ui.spec.ts`. | Unitário `web/src/routes/home.spec.ts:157-164`: renderiza `EmptyState` com `kind` `day` e `filters` e confere `<a href="/lobbies/novo" ...>Criar lobby` e a ausência de "Disponível em breve". E2E `web/test/e2e/home.spec.ts:207-212` (dia vazio): o link do aviso tem `href="/lobbies/novo"`. E2E `home.spec.ts:238-240` (filtros sem resultado): mesma asserção de `href`. Cabeçalho e ícone do celular: `home.spec.ts:73-75` (unitário) e `test/e2e/home.spec.ts:177-179` (já existiam). | O teste unitário renderiza o componente isolado, e o E2E confirma que o aviso aparece de fato na Home nos dois casos. |
| CA-02.3 / RN-04 (visitante vai ao login e volta para a criação) | ✅ | O link aponta para `/lobbies/novo`, cujo `load` (`web/src/routes/lobbies/novo/+page.server.ts:18`) chama `toLogin(cookies, PATH)` sem sessão, com `PATH = '/lobbies/novo'`. O destino é o mesmo do botão do cabeçalho, então o retorno segue o mesmo caminho. | E2E `home.spec.ts:211-212`: clicar no "Criar lobby" do aviso de dia vazio leva a `/oauth2/authorize`. Volta para `/lobbies/novo` depois do login: `web/test/e2e/lobbies.spec.ts:76-85` (CA-02.3, já existia, pelo botão do cabeçalho). | Pelo aviso, o teste só confere a ida ao login. A volta não é testada a partir desse botão, mas a URL e o servidor são os mesmos do teste de `lobbies.spec.ts`, que confere a volta. Não bloqueia. |
| Citação de IDs nos testes | ✅ | — | O nome do teste unitário cita `CA-02.3 / RN-23 (lobbies)`. Os trechos novos do E2E ficam dentro dos testes `CA-03.3` e `CA-04.7` (IDs da `home-local`) e citam `CA-02.3 / RN-23 (lobbies)` e `RN-04` em comentário (`home.spec.ts:207-208`). O commit cita `[RN-23, CA-02.3]`. | No E2E, o ID da lobbies aparece no comentário, não no nome do teste. Isso é aceitável, porque o teste continua sendo do critério da `home-local`. |

## Pendências

Nenhuma.

## Scope creep

Nenhum. O diff mexe só no botão do aviso de lista vazia (1 import e 1 linha trocada, mais o comentário com o ID) e nos testes desse comportamento. Não há mudança de backend, contrato, estilo nem de outros componentes.

## Observações

- Fora do diff e sem efeito no veredito: o comentário em `web/src/routes/perfil/page.spec.ts:111` ainda fala do "aria-disabled de 'Criar lobby'", comportamento que não existe mais desde a RN-23. Vale ajustar numa próxima mexida.
- A suíte E2E `lobbies` não foi rodada neste ciclo, porque o escopo pedia `playwright test home`. A volta do login para `/lobbies/novo` foi confirmada pela leitura do teste existente (`lobbies.spec.ts:76-85`), que não foi alterado.
