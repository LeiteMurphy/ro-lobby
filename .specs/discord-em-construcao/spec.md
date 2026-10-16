# Spec — Sala no Discord em construção

- Feature: `discord-em-construcao` · Nível: P · Status: Aprovada (2026-10-15)
- Notion: —
- Última revisão: 2026-10-15 — primeira versão, com as respostas do usuário

## 1. Contexto
O servidor oficial do RO Lobby no Discord ainda está sendo configurado pelo dono do
projeto, mas a sala do grupo (spec `discord-bot`) já está no ar em `rolobby.com.br`, e o
botão "Entrar no servidor do RO Lobby" leva os jogadores a um servidor incompleto. Até a
configuração terminar, a entrada no servidor e a sala do grupo ficam "em construção" no
site. O bot e a API não mudam: quando a configuração terminar, basta desligar a chave.

## 2. Research
- O botão aparece em dois lugares: o card "Servidor do RO Lobby no Discord" do `/perfil`
  (`web/src/routes/perfil/+page.svelte`) e o atalho da seção "Sala no Discord" do detalhe
  do lobby, para membro não conectado (`web/src/lib/discord/components/DiscordSection.svelte`).
- Os dois levam a `/perfil/discord`, que redireciona ao Discord (`identify guilds.join`), e
  o retorno é `/perfil/discord/retorno`.
- "Abrir sala" e "Avisar no Discord" são actions do detalhe do lobby (`?/openRoom` e
  `?/alert`) que chamam a API.
- O link do Discord do dono (RN-23 da discord-bot) não depende do bot e continua.

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como dono do projeto, quero que a entrada no servidor e a sala do grupo apareçam como "Em construção", para os jogadores não entrarem num servidor que ainda não está pronto. |

## 4. Regras
- **RN-01** — Uma chave no ambiente do web, `DISCORD_UNDER_CONSTRUCTION`, liga o modo em
  construção quando vale `true`. A pilha `app` liga por padrão; o modo de desenvolvimento e
  o e2e da sala ficam como hoje.
- **RN-02** — Em construção, o card do `/perfil` mostra o botão desabilitado com o texto
  "Em construção" e a frase "O servidor do RO Lobby no Discord está em construção. Em breve
  você entra por aqui.", no lugar de "Entrar no servidor do RO Lobby" ou "Reconectar".
- **RN-03** — Em construção, a seção "Sala no Discord" do detalhe do lobby mostra, para os
  membros, "A sala no Discord está em construção." e um botão desabilitado "Em
  construção", no lugar de abrir a sala, do link da sala, da lista de quem está na sala, do
  atalho para entrar no servidor e do "Avisar no Discord".
- **RN-04** — O link do Discord do dono continua aparecendo aos membros como hoje
  (RN-23 a RN-25 da discord-bot).
- **RN-05** — Em construção, quem abre `/perfil/discord` ou `/perfil/discord/retorno`
  volta ao `/perfil` com "O servidor do RO Lobby no Discord está em construção.", sem ir ao
  Discord e sem chamar a API.
- **RN-06** — Em construção, as actions "Abrir sala" e "Avisar no Discord" recusam com "A
  sala no Discord está em construção.", sem chamar a API.
- **RN-07** — Com a chave desligada, tudo volta a ser como na spec `discord-bot`.

## 5. Critérios de aceite
```gherkin
CA-01.1 — Perfil em construção  [US-01, RN-01, RN-02]
Given a sala no Discord ligada e DISCORD_UNDER_CONSTRUCTION=true
When um usuário logado abre o /perfil
Then o card do servidor mostra o botão "Em construção" desabilitado e a frase da RN-02
  And não há link para /perfil/discord

CA-01.2 — Detalhe do lobby em construção  [US-01, RN-03, RN-04]
Given um lobby com o link do Discord do dono, a chave ligada e um membro logado
When ele abre o detalhe do lobby
Then a seção "Sala no Discord" mostra "A sala no Discord está em construção." e o botão
     "Em construção" desabilitado
  And não aparecem "Abrir sala no Discord", o link da sala, a lista da sala, o atalho para
      entrar no servidor nem "Avisar no Discord"
  And o "Discord do anfitrião" continua aparecendo

CA-01.3 — Endereço da conexão fechado  [US-01, RN-05]
Given a chave ligada
When um usuário logado abre /perfil/discord ou /perfil/discord/retorno
Then ele volta ao /perfil com "O servidor do RO Lobby no Discord está em construção."
  And nenhuma chamada vai ao Discord nem à API

CA-01.4 — Actions recusadas  [US-01, RN-06]
Given a chave ligada e o dono de um lobby
When ele envia a action de abrir a sala ou de avisar
Then a resposta traz "A sala no Discord está em construção." e a API não é chamada

CA-01.5 — Chave desligada volta ao normal  [US-01, RN-07]
Given DISCORD_UNDER_CONSTRUCTION vazia ou diferente de true
When o usuário abre o /perfil e o detalhe do lobby
Then o botão "Entrar no servidor do RO Lobby" e a seção da sala funcionam como na
     spec discord-bot (o e2e atual da sala continua passando)
```

## 6. Casos de borda
- Sala que já estava aberta antes da chave → o site não mostra o link nem a lista; a sala
  continua existindo no Discord e é apagada no prazo de sempre pelo bot (RN-16 e RN-17 da
  discord-bot) [RN-03].
- Bot desligado (sem token) → a seção e o card continuam escondidos, como hoje; a chave não
  muda nada [RN-07].

## 7. Requisitos não funcionais
- **RNF-01** — O botão desabilitado é anunciado como desabilitado por leitor de tela.

## 8. Fora de escopo
- Mudar a API ou o bot: a abertura da sala continua possível pela API, que não é exposta
  (RN-07 da endurecimento).
- Configuração do servidor oficial no Discord: é trabalho do dono, fora do código.
- Painel do Cloudflare: adiado pelo usuário.

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Onde fica em construção → nos dois lugares (perfil e detalhe do lobby).
- A sala inteira (abrir, avisar) → tudo em construção; o link do dono continua.
- O endereço `/perfil/discord` → fechado também.
