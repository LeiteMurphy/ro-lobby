# ADR-09 — Bot do Discord por REST, com outbox no PostgreSQL

- Status: Aceito
- Data: 2026-10-14
- Spec: `.specs/discord-bot/spec.md` (design D-01 a D-04)

## Contexto
A sala do grupo no Discord pede que o RO Lobby crie canais, dê e tire acesso, poste
mensagens e apague canais num servidor oficial. O jogador entra nesse servidor pelo escopo
OAuth `guilds.join`. A API é um processo único que roda no PC do usuário (ADR-06,
ADR-08), e não há tarefas em segundo plano. A `login-discord` proíbe guardar tokens do
Discord. As ações do grupo (aceitar, sair, cancelar) não podem falhar porque o Discord
está fora do ar.

## Opções consideradas
- Conversa com o Discord:
  - (a) REST com `net/http`, como o cliente OAuth;
  - (b) a biblioteca `discordgo` com gateway;
  - (c) um serviço de bot separado.
- Efeitos fora da ação do usuário:
  - (a) chamar o Discord dentro da ação e ignorar as falhas;
  - (b) uma tabela de jobs (outbox) gravada na mesma transação, com um trabalhador na API;
  - (c) uma fila externa.
- Entrada no servidor:
  - (a) pedir `guilds.join` no login e guardar o token;
  - (b) um passo de conexão separado, que inclui o jogador na hora e descarta o token.

## Decisão
- O bot usa só REST com `net/http`, sem gateway, no mesmo processo da API.
- Os efeitos assíncronos (acesso, mensagens) vão para uma outbox no PostgreSQL, gravada na
  transação da ação e executada por um trabalhador na API com novas tentativas. O lembrete
  e a remoção da sala são deduzidos do horário do lobby e do estado da sala, sem
  agendamento.
- A conexão ao servidor é um OAuth separado com `identify` e `guilds.join`. A API inclui o
  jogador no servidor na hora e descarta o token.
- Configuração: `DISCORD_BOT_TOKEN` e `DISCORD_GUILD_ID`. Sem as duas, a feature fica
  desligada.

## Consequências
- \+ Nenhuma infraestrutura ou dependência nova. O teste usa o Discord falso existente.
- \+ As ações do domínio continuam atômicas e independentes do Discord. Os efeitos
  sobrevivem a reinícios.
- \+ Nenhum token de usuário é guardado.
- − O RO Lobby não percebe o que acontece no Discord (saída do servidor, canal apagado à
  mão). Comandos de barra vão exigir um endpoint de interações ou o gateway, com novo ADR.
- − A entrega é "pelo menos uma vez": uma queda entre postar e marcar o job pode repetir
  uma mensagem.
- − O bot só funciona com o PC ligado. Ao voltar, ele recupera os efeitos atrasados.
