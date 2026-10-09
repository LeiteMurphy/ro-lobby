# Spec — Compartilhar lobby

- Feature: `compartilhar-lobby` · Nível: P · Status: Aprovada (2026-10-09)
- Notion: https://app.notion.com/p/3f4d4a3a5eff81a99154f880727cb5e9
- Última revisão: 2026-10-08 — primeira versão

## 1. Contexto
Os grupos de RO se formam no Discord e no WhatsApp. Hoje, para divulgar um lobby, a pessoa
copia a URL da barra do navegador, e o link colado aparece sem preview. O detalhe do lobby
já é público e leva o visitante ao login e de volta ao lobby para se candidatar (RN-15 da
`lobbies`). Faltam um botão que gere um convite pronto e o preview do link, que foi o
motivo do SSR no ADR-01.

## 2. Research
- `/lobbies/{id}` é público e renderizado no servidor (`web/src/routes/lobbies/[id]`). O
  visitante vê "Entrar para se candidatar", que volta ao lobby depois do login.
- A página só define `<title>`; não há meta tags Open Graph em nenhuma rota.
- A origem dos links sai da origem do web (`ORIGIN` no adapter-node). Com a publicação
  por túnel (`publicacao-tunel`), ela vira `https://rolobby.com.br`; sem ela, o convite
  traz `http://localhost:3000`, que só serve para teste.
- Os assets da marca são SVG (`web/static/brand`). O Discord não mostra SVG em preview.

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como jogador, quero copiar um convite pronto do lobby, para divulgar o grupo no Discord ou no WhatsApp. |
| US-02 | P1 | Como jogador no Discord, quero ver um preview do lobby quando alguém cola o link, para saber do que se trata antes de abrir. |

## 4. Regras
- **RN-01** — O detalhe de um lobby aberto mostra "Compartilhar" para qualquer pessoa,
  logada ou não, ao lado das outras ações. Lobby iniciado ou cancelado não mostra o botão.
- **RN-02** — "Compartilhar" copia para a área de transferência um convite de três linhas:
  ```
  Grupo para <instância> · <dia da semana>, <dd/mm> às <hh:mm> (horário de Brasília)
  <vagas> · Nível mínimo <n>
  Candidate-se: <origem>/lobbies/<id>
  ```
  A data é absoluta (não "hoje" nem "amanhã"), porque o texto continua circulando depois.
  O horário segue a RN-09 da `home-local`.
- **RN-03** — `<vagas>` lista as vagas abertas por função, na ordem Tank, Suporte, Dano,
  só as funções com vaga, no formato "Vagas: 1 Tank, 2 Suporte, 3 Dano". Sem vaga aberta,
  vira "Grupo lotado".
- **RN-04** — Depois de copiar, o botão mostra "Convite copiado" por alguns segundos. Se o
  navegador não deixar copiar, abre um diálogo com o convite selecionado para a pessoa
  copiar na mão.
- **RN-05** — O detalhe do lobby tem meta tags Open Graph, sem imagem:
  - `og:title`: "<instância> · <dia da semana>, <dd/mm> às <hh:mm>";
  - `og:description`: com o lobby aberto, "<vagas da RN-03> · Nível mínimo <n> ·
    Anfitrião <nick>"; iniciado, "Esse grupo já começou."; cancelado, "Esse grupo foi
    cancelado.";
  - `og:url` (link do lobby), `og:site_name` "RO Lobby", `og:type` "website" e
    `theme-color` com a cor de destaque da marca.
- **RN-06** — Nada no convite nem no preview usa logo, arte ou sprite da Gravity
  (CLAUDE.md).

## 5. Critérios de aceite
```gherkin
CA-01.1 — Botão em lobby aberto  [US-01, RN-01]
Given um lobby aberto
When um visitante, um candidato ou o dono abre o detalhe
Then o botão "Compartilhar" aparece

CA-01.2 — Sem botão em lobby encerrado  [US-01, RN-01]
Given um lobby iniciado ou cancelado
When alguém abre o detalhe
Then o botão "Compartilhar" não aparece

CA-01.3 — Convite copiado  [US-01, RN-02, RN-03, RN-04]
Given um lobby aberto de "Glast Heim" no sábado 10/10 às 20:00, nível mínimo 160,
  com 1 vaga de Tank e 2 de Dano abertas e Suporte cheio
When alguém clica em "Compartilhar"
Then a área de transferência recebe:
  "Grupo para Glast Heim · sábado, 10/10 às 20:00 (horário de Brasília)
   Vagas: 1 Tank, 2 Dano · Nível mínimo 160
   Candidate-se: <origem>/lobbies/<id>"
  And o botão mostra "Convite copiado"

CA-01.4 — Grupo lotado  [US-01, RN-03]
Given um lobby aberto sem vaga aberta
When alguém copia o convite
Then a segunda linha começa com "Grupo lotado"

CA-01.5 — Sem permissão de copiar  [US-01, RN-04]
Given um navegador que recusa a escrita na área de transferência
When alguém clica em "Compartilhar"
Then abre um diálogo com o convite selecionado

CA-02.1 — Preview de lobby aberto  [US-02, RN-05]
Given o mesmo lobby do CA-01.3, com anfitrião "Brasa"
When o HTML de /lobbies/<id> é pedido sem sessão
Then ele contém og:title "Glast Heim · sábado, 10/10 às 20:00"
  And og:description "Vagas: 1 Tank, 2 Dano · Nível mínimo 160 · Anfitrião Brasa"
  And og:url, og:site_name "RO Lobby", og:type "website" e theme-color

CA-02.2 — Preview de lobby encerrado  [US-02, RN-05]
Given um lobby cancelado
When o HTML de /lobbies/<id> é pedido
Then og:description é "Esse grupo foi cancelado."

CA-02.3 — Preview no Discord  [US-02, RN-05] (UAT, manual)
Given o site publicado em rolobby.com.br
When alguém cola o link de um lobby no Discord
Then aparece o cartão com título e descrição
```

## 6. Fora de escopo
- Imagem no preview — os assets da marca são SVG, e o Discord pede PNG ou JPG. Pode
  entrar com uma arte própria depois.
- Botão no card da Home — escolhido só no detalhe.
- Compartilhamento nativo do celular (Web Share API) e botões diretos para Discord ou
  WhatsApp.
- Encurtador de link ou link com código próprio — o link é o do detalhe.

## 7. Decisões tomadas na conversa
- Copia um convite pronto com o link.
- Todos veem o botão, só com o lobby aberto.
- Preview com Open Graph nesta entrega.
- Botão só no detalhe do lobby.
