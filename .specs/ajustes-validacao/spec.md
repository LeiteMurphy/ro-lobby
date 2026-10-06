# Spec — Ajustes das observações de validação (login e personagens)

- Feature: `ajustes-validacao` · Nível: P · Status: Em revisão (vai com o PR)
- Origem: observações não bloqueantes de `.specs/login-discord/validation-report.md` e
  `.specs/personagens/validation-report.md`

## Contexto
Os dois validadores aprovaram as features, mas deixaram observações. Esta spec junta as que
valem corrigir agora. Não há regra de negócio nova, exceto o RN-21 da `personagens`, que
escreve na spec um comportamento que o código já tinha e os testes já cobriam sem ID.

## Itens

| ID | Origem | O que muda | Pronto quando |
|---|---|---|---|
| AJ-01 | login, CA-05.1 | O ponta a ponta prova que a pessoa chega **logada** à página de origem | O teste entra por `/auth/discord/login?next=/perfil` e vê "Meus personagens" e o menu do usuário |
| AJ-02 | login, RN-13 | Criar o Usuário e a Sessão numa transação só: se a Sessão falhar, o Usuário não fica no banco | Teste de integração força a falha ao gravar a Sessão e confere que nenhum Usuário foi criado |
| AJ-03 | login, RN-03 | Client ID ou Client Secret errado gera um log de erro na API (sem o segredo) | Teste confere o log no 401 `invalid_client` e a ausência dele no 400 `invalid_grant` |
| AJ-04 | personagens, RN-19 | Fechar e reabrir "Adicionar personagem" não traz os valores e erros da tentativa anterior | Ponta a ponta: erro, fechar, reabrir → campos vazios e sem mensagem |
| AJ-05 | personagens, D-04 | O design descreve os retratos como são: pixel art em faixas | Texto do D-04 atualizado |
| AJ-06 | login e personagens, RNF-03 | Os testes sem ID passam a citar o critério | Nenhum teste das duas features sem ID |
| AJ-07 | personagens, RN-21 | Achado ao dar ID ao teste do 500: o servidor gerado devolvia o texto do erro interno no corpo. A API passa a registrar o erro no log e responder só "erro interno" | Testes do 500 no login e nos personagens conferem que o corpo não traz o erro |

### RN-21 (novo, na spec `personagens`)
Se a API não responder ou falhar, a página de perfil mostra "Não foi possível falar com o
servidor. Tente de novo." e não quebra. Um erro inesperado na API responde 500, sem
detalhes no corpo.

## Fora de escopo (ficam registrados, sem mudança agora)
- **Dados do ponta a ponta no banco de desenvolvimento.** Os testes usam usuários e nicks
  com sufixo aleatório, então não há colisão. Um banco só para o ponta a ponta fica para
  quando a hospedagem (ADR-06) for revista.
- **`Update` de personagem fora da transação com trava.** Hoje está correto; a trava entra
  junto com as regras de candidatura (RN-25 da `candidatura-lobby`), que mexem na edição.
