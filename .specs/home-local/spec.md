# Spec — Hospedagem local e Home

- Feature: `home-local` · Nível: M · Status: Aprovada
- Design: Claude Design, projeto "RO Lobby Home v2", arquivo `Home v2.dc.html`
- ADR: [ADR-06](../../docs/adr/0006-hospedagem-local.md) (aceito)
- Última revisão: 2026-10-01 — RN-22 revista a pedido do usuário: nomes de classe do bRO
  (bROWiki) passam a ser permitidos; logos, artes e sprites continuam proibidos

## 1. Contexto
A fundação deixou um esqueleto que roda em modo de desenvolvimento, com a página de
status em `/`. O próximo passo é rodar o projeto inteiro em `localhost` como se fosse um
servidor e trocar a página provisória pela Home desenhada no Claude Design.

Ainda não existem login, personagens nem lobbies. Por isso, a Home mostra dados
fictícios, isolados num módulo que a feature de lobby vai trocar pela API, e as ações
que dependem de backend ficam visíveis, mas desabilitadas.

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Desenvolvedor | Quem roda o projeto na máquina | Subir o ambiente completo, rodar testes |
| Visitante | Quem abre o web, sem login | Ver os grupos, trocar o dia, filtrar |
| CI | GitHub Actions | Provar que as imagens compilam |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como desenvolvedor, quero subir banco, migrações, API e web em containers com um comando, para rodar o projeto em `localhost` como num servidor. |
| US-02 | P1 | Como visitante, quero ver os grupos do dia na Home, com horário e composição por função, para achar um grupo para entrar. |
| US-03 | P1 | Como visitante, quero trocar o dia no seletor, para ver os grupos dos próximos dias. |
| US-04 | P1 | Como visitante, quero filtrar por instância, vaga, nível e horário, para achar só os grupos que me servem. |
| US-05 | P2 | Como visitante, quero ver o próximo grupo com vaga em destaque, para decidir rápido. |
| US-06 | P2 | Como visitante no celular, quero abrir os filtros numa gaveta, para usar a Home numa tela pequena. |
| US-07 | P2 | Como desenvolvedor, quero que a CI prove que as imagens Docker compilam, para a hospedagem local não quebrar sem ninguém ver. |

## 4. Regras

### Hospedagem local
- **RN-01** — `docker compose --profile app up` sobe, nesta ordem, o PostgreSQL, as
  migrações, a API e o web, todos com o build de produção.
- **RN-02** — As migrações rodam num serviço que termina ao aplicar tudo. A API só sobe
  depois que ele termina com sucesso. Se ele falhar, a API não sobe.
- **RN-03** — Só o web é exposto na máquina, na porta `APP_WEB_PORT` (padrão 3000). A
  API e o banco da pilha `app` ficam na rede interna do Compose. O web chama a API pelo
  nome do serviço.
- **RN-04** — `docker compose up` sem perfil continua subindo só o PostgreSQL. O modo de
  desenvolvimento da fundação não muda (RN-03 da fundação).
- **RN-05** — As imagens são multi-stage: a imagem final não tem compilador, código-fonte
  nem dependências de desenvolvimento, e o processo roda com usuário não-root.
- **RN-06** — Toda variável nova aparece no `.env.example` com valor fictício, e o README
  descreve como subir, parar e apagar a pilha `app`.
- **RN-07** — A CI faz o build das duas imagens quando `backend/`, `web/`, o
  `docker-compose.yml` ou a própria CI mudam. Ela não publica as imagens.

### Home
- **RN-08** — Os lobbies da Home vêm de um único módulo de dados fictícios do web. O
  resto do web só acessa esses dados por uma função, que a feature de lobby vai trocar
  pela API.
- **RN-09** — "Hoje", os horários e o tempo relativo são calculados no fuso
  `America/Sao_Paulo`, tanto no servidor quanto no navegador.
- **RN-10** — O seletor mostra 14 dias a partir de hoje, cada um com a quantidade de
  grupos, e começa em hoje.
- **RN-11** — A lista do dia fica em ordem crescente de horário.
- **RN-12** — Os filtros se combinam com E, e cada um vale assim:
  - Instância: igual à escolhida.
  - Vaga para: o lobby tem vaga aberta em pelo menos uma das funções marcadas.
  - Nível mínimo: o nível mínimo do lobby é menor ou igual ao escolhido.
  - Faixa de horário: o início fica dentro da faixa, com o início incluído e o fim
    excluído. As faixas são 18h–20h, 20h–22h e 22h–00h.
- **RN-13** — As contagens dos filtros valem para o dia escolhido:
  - Em "Vaga para", cada função mostra quantos lobbies do dia têm vaga nela, sem
    considerar os filtros.
  - Em "Faixa de horário", cada faixa mostra quantos lobbies passam nos outros filtros e
    começam nela.
- **RN-14** — Vagas abertas de uma função = total da função − ocupadas. O lobby está
  lotado quando as vagas ocupadas somam o total.
  - Lotado: o card fica esmaecido, com a borda cinza e o selo "Lotado".
  - Não lotado: a borda do card fica na cor da função com mais vagas abertas. No empate,
    vale a ordem Tank, Suporte, Dano.
- **RN-15** — O tempo relativo só aparece em hoje e para horários futuros, no formato
  "em N min" (menos de 60 min), "em H h" ou "em H h M min". Ele se atualiza a cada
  minuto no navegador.
- **RN-16** — O destaque mostra o primeiro lobby do dia, por horário, que não está lotado
  e, se o dia for hoje, que ainda não começou. Sem lobby assim, não há destaque. O
  destaque ignora os filtros.
- **RN-17** — Estados vazios:
  - Dia sem grupos: "Nenhum grupo neste dia".
  - Dia com grupos, mas nenhum passa nos filtros: "Nenhum grupo com esses filtros", com o
    botão "Limpar filtros".
- **RN-18** — Enquanto não houver login, candidatura e criação de lobby, estes controles
  aparecem desabilitados, com a dica "Disponível em breve": "Criar lobby", "Candidatar",
  "Ver grupo" e "Entrar com Discord".
- **RN-19** — Abaixo de 900 px de largura:
  - os filtros ficam numa gaveta à esquerda;
  - o botão "Filtros (N)" mostra quantos filtros estão ativos;
  - "Criar lobby" vira um botão só com ícone.
- **RN-20** — O HTML que o servidor devolve para `/` já contém os cards de hoje.
- **RN-21** — A página de status da fundação passa para `/status`, com o mesmo
  comportamento (RN-15 da fundação).
- **RN-22** — A Home não usa logos, artes nem sprites do Ragnarok Online. Os nomes de
  classe seguem a nomenclatura do bRO, conforme o [bROWiki](https://browiki.org/wiki/Classes);
  os nomes de instância continuam inventados. Os textos seguem o design system: pt-BR,
  "você", sem exclamação nem emoji.

## 5. Critérios de aceite

### US-01 — Pilha local em containers
```gherkin
CA-01.1 — Subir tudo  [US-01, RN-01, RN-02, RN-03]
Given uma máquina com o Docker Desktop e o repositório com o .env copiado do .env.example
When o desenvolvedor roda docker compose --profile app up -d --wait
Then o PostgreSQL, a API e o web ficam saudáveis
  And o serviço de migração termina com código 0
  And http://localhost:3000/ responde 200 com a Home
  And http://localhost:3000/status mostra "API online"

CA-01.2 — Migração falha, API não sobe  [US-01, RN-02]
Given uma migração que falha
When o desenvolvedor sobe a pilha app
Then o serviço de migração termina com código diferente de 0
  And o container da API não chega a iniciar

CA-01.3 — Só o web exposto  [US-01, RN-03]
Given a pilha app no ar
When se listam as portas publicadas na máquina
Then só a porta do web aparece
  And a API e o banco da pilha app só são alcançáveis pela rede interna

CA-01.4 — Modo de desenvolvimento preservado  [US-01, RN-04]
Given o repositório
When o desenvolvedor roda docker compose up -d sem perfil
Then só o PostgreSQL sobe

CA-01.5 — Imagens enxutas e sem root  [US-01, RN-05]
Given as imagens da API e do web
When se inspeciona cada uma
Then o usuário do processo não é root
  And a imagem final não tem o compilador Go nem o código-fonte do backend
  And a imagem final do web não tem as dependências de desenvolvimento

CA-01.6 — Documentado  [US-01, RN-06]
Given o README e o .env.example
When se levantam as variáveis lidas pelo Compose e pelas imagens
Then todas estão no .env.example com valor fictício
  And o README tem os comandos para subir, parar e apagar a pilha app
```

### US-02 — Grupos do dia
```gherkin
CA-02.1 — Lista de hoje  [US-02, RN-08, RN-10, RN-11]
Given os dados fictícios com lobbies hoje
When o visitante abre /
Then a Home mostra os lobbies de hoje em ordem de horário
  And cada card mostra horário, instância, vagas ocupadas/total, anfitrião, classe, nível mínimo e composição por função

CA-02.2 — Lotado  [US-02, RN-14]
Given um lobby com todas as vagas ocupadas
When ele aparece na lista
Then o card mostra o selo "Lotado", fica esmaecido e tem a borda cinza
  And não mostra o botão "Candidatar"

CA-02.3 — Borda pela função com mais vagas  [US-02, RN-14]
Given um lobby com 0 vagas de Tank, 2 de Suporte e 2 de Dano abertas
When ele aparece na lista
Then a borda do card tem a cor de Suporte

CA-02.4 — Tempo relativo  [US-02, RN-09, RN-15]
Given agora são 16:40 em America/Sao_Paulo e um lobby de hoje às 18:00
When o visitante vê o card
Then o card mostra "em 1 h 20 min"
  And um lobby de hoje às 16:00 não mostra tempo relativo
  And um lobby de amanhã não mostra tempo relativo

CA-02.5 — Ações desabilitadas  [US-02, RN-18]
Given a Home aberta
When o visitante tenta usar "Criar lobby", "Candidatar", "Ver grupo" ou "Entrar com Discord"
Then os controles estão desabilitados
  And cada um mostra a dica "Disponível em breve"

CA-02.6 — Renderização no servidor  [US-02, RN-20]
Given lobbies hoje
When se busca / sem executar JavaScript
Then o HTML recebido já contém os cards de hoje

CA-02.7 — Status em /status  [US-02, RN-21]
Given a API saudável
When o visitante abre /status
Then a página mostra "API online"
```

### US-03 — Troca de dia
```gherkin
CA-03.1 — 14 dias com contagem  [US-03, RN-10]
Given a Home aberta
When o visitante olha o seletor
Then ele mostra 14 dias a partir de hoje, com hoje selecionado
  And cada dia mostra a quantidade de grupos daquele dia

CA-03.2 — Trocar o dia  [US-03, RN-10, RN-11]
Given a Home em hoje
When o visitante escolhe outro dia com grupos
Then a lista mostra os grupos daquele dia, em ordem de horário
  And o título muda de "Grupos para hoje" para "Grupos para <dia>"

CA-03.3 — Dia sem grupos  [US-03, RN-17]
Given um dia sem grupos
When o visitante escolhe esse dia
Then a Home mostra "Nenhum grupo neste dia"
  And não mostra o botão "Limpar filtros"
```

### US-04 — Filtros
```gherkin
CA-04.1 — Instância  [US-04, RN-12]
Given um dia com lobbies em instâncias diferentes
When o visitante escolhe uma instância
Then a lista mostra só os lobbies dessa instância

CA-04.2 — Vaga para (OU entre funções)  [US-04, RN-12]
Given um dia com lobbies com e sem vaga de Tank e de Suporte
When o visitante marca Tank e Suporte
Then a lista mostra os lobbies com vaga aberta de Tank ou de Suporte
  And esconde os que não têm vaga em nenhuma das duas

CA-04.3 — Nível mínimo  [US-04, RN-12]
Given lobbies com nível mínimo 150, 160 e 185
When o visitante escolhe "Até Nv 160"
Then a lista mostra os lobbies de nível mínimo 150 e 160

CA-04.4 — Faixa de horário  [US-04, RN-12]
Given lobbies às 19:30, 20:00 e 22:45
When o visitante escolhe "20h–22h"
Then a lista mostra só o lobby das 20:00

CA-04.5 — Filtros combinados  [US-04, RN-12]
Given um dia com lobbies variados
When o visitante combina instância e faixa de horário
Then a lista mostra só os lobbies que passam nos dois filtros

CA-04.6 — Contagens  [US-04, RN-13]
Given um dia com lobbies e um filtro de instância ativo
When o visitante olha os filtros
Then cada função em "Vaga para" mostra quantos lobbies do dia têm vaga nela, sem considerar o filtro de instância
  And cada faixa de horário mostra quantos lobbies da instância escolhida começam nela

CA-04.7 — Nenhum resultado  [US-04, RN-17]
Given um dia com grupos
When os filtros escolhidos não deixam nenhum grupo
Then a Home mostra "Nenhum grupo com esses filtros" e o botão "Limpar filtros"
  And "Limpar filtros" volta a lista para todos os grupos do dia

CA-04.8 — Subtítulo  [US-04, RN-12]
Given um dia com 5 grupos
When o visitante aplica filtros que deixam 2
Then o subtítulo mostra "2 de 5 grupos"
```

### US-05 — Destaque
```gherkin
CA-05.1 — Próximo com vaga  [US-05, RN-16]
Given hoje com um lobby lotado às 18:00 e um com vaga às 19:00
When o visitante abre a Home
Then o destaque mostra o lobby das 19:00

CA-05.2 — Ignora os que já começaram  [US-05, RN-09, RN-16]
Given agora são 19:10 e hoje há lobbies com vaga às 19:00 e às 21:30
When o visitante abre a Home
Then o destaque mostra o lobby das 21:30

CA-05.3 — Sem destaque  [US-05, RN-16]
Given um dia em que todos os lobbies estão lotados
When o visitante escolhe esse dia
Then a Home não mostra destaque
```

### US-06 — Celular
```gherkin
CA-06.1 — Gaveta de filtros  [US-06, RN-19]
Given a Home numa tela de 390 px de largura
When o visitante toca em "Filtros"
Then os filtros abrem numa gaveta à esquerda
  And a barra lateral não aparece

CA-06.2 — Contagem de filtros ativos  [US-06, RN-19]
Given a Home numa tela de 390 px com 2 filtros ativos
When o visitante olha o botão de filtros
Then ele mostra "Filtros (2)"
```

### US-07 — Imagens na CI
```gherkin
CA-07.1 — Build das imagens  [US-07, RN-07]
Given um PR que muda backend/, web/ ou o docker-compose.yml
When a CI roda
Then as imagens da API e do web são construídas
  And uma falha no build deixa o check CI ok vermelho
```

## 6. Casos de borda
- Porta 3000 ocupada: a porta vem de `APP_WEB_PORT` [RN-03, RN-06]
- Virada do dia com a Home aberta: o "hoje" é recalculado ao recarregar a página, e o
  tempo relativo some quando o horário passa [RN-09, RN-15]
- Visitante em outro fuso: vê os horários de `America/Sao_Paulo`, igual ao servidor
  [RN-09]
- Dia com todos os lobbies lotados: lista com os cards esmaecidos e sem destaque
  [RN-14, RN-16]

## 7. Requisitos não funcionais
- **RNF-01** — Todo controle interativo tem rótulo acessível, dá para usar pelo teclado e
  tem foco visível (anel âmbar do design system).
- **RNF-02** — A Home não carrega nada de fora do próprio servidor em runtime: fontes,
  ícones e imagens são servidos pelo web.
- **RNF-03** — Todo teste cita o ID do critério de aceite (CLAUDE.md).

## 8. Fora de escopo
- Hospedagem na nuvem, domínio, HTTPS e deploy. Ficam para a revisão do ADR-06.
- Login com Discord, personagens, criação de lobby e candidatura. A Home só mostra os
  controles, desabilitados.
- API e tabelas de lobbies. Os dados continuam fictícios.
- Calendário mensal. O seletor de 14 dias cobre "os outros dias" nesta entrega.
- Publicar as imagens num registry.
- As outras telas do design system (detalhe do lobby, personagens, diálogos).

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Pilha local → perfil `app` do Compose com o build de produção, só o web exposto, na
  porta 3000.
- ADR-06 → "local por enquanto", com revisão quando o login com Discord estiver pronto.
- CI → build das duas imagens, sem publicar.
- Dados da Home → fictícios, iguais aos do design, num módulo isolado. Status em `/status`.
- Sem login → "Entrar com Discord" desabilitado; o destaque considera qualquer função.
- Ações sem backend → desabilitadas, com a dica "Disponível em breve".
- Interativos → dias, filtros com contagem, "Limpar filtros", gaveta e estados vazios.
- Fidelidade → tokens, fontes e assets do design system; componentes em Svelte; ícones
  pelo `@lucide/svelte`; fontes pelo `@fontsource`.
