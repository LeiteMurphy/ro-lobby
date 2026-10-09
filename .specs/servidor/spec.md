# Spec — Servidor do jogo

- Feature: `servidor` · Nível: M · Status: Rascunho
- Notion: (criado na fase de Tasks)
- Última revisão: 2026-10-08 — primeira versão

## 1. Contexto
O Ragnarok LATAM tem dois servidores, Freya e Nidhogg. Cada personagem vive em um deles, e
um grupo só joga junto se todos estiverem no mesmo servidor. Hoje o RO Lobby não sabe o
servidor de ninguém: um jogador de Freya pode se candidatar a um grupo de Nidhogg e só
descobre o problema na hora de jogar. Esta feature grava o servidor no personagem e no
lobby, impede grupos com servidores misturados e deixa a Home filtrar por servidor.

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Visitante | Pessoa sem login | Vê o servidor nos cards e no detalhe; filtra a Home |
| Usuário | Pessoa logada com Discord | Informa o servidor dos seus personagens |
| Dono | Usuário que criou o lobby | O lobby fica no servidor do personagem dele |
| Candidato / Membro | Usuário que se candidata ou já está no grupo | Só usa personagens do servidor do lobby |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como usuário, quero informar o servidor de cada personagem, para o RO Lobby saber onde ele joga. |
| US-02 | P1 | Como dono, quero que o lobby fique no servidor do meu personagem, para os candidatos saberem onde o grupo vai jogar. |
| US-03 | P1 | Como dono, quero que só personagens do servidor do lobby entrem no grupo, para não montar um grupo que não consegue jogar junto. |
| US-04 | P1 | Como visitante, quero filtrar a Home por servidor, para ver só os grupos de onde eu jogo. |

## 4. Regras de domínio

### Catálogo
- **RN-01** — Os servidores são Freya e Nidhogg, nesta ordem. A API usa os códigos
  `freya` e `nidhogg`. O catálogo vive na API, como as classes e as instâncias, e
  `GET /servers` não exige sessão.

### Personagem
- **RN-02** — Servidor é campo obrigatório do personagem, um dos valores da RN-01. A API
  recusa servidor vazio ou fora do catálogo, com o erro no campo.
- **RN-03** — O nick é único por servidor, sem diferenciar maiúsculas de minúsculas.
  "Brasa" em Freya e "brasa" em Nidhogg podem existir ao mesmo tempo, inclusive de
  Usuários diferentes; dois "Brasa" em Freya, não. (Substitui a RN-05 da `personagens`,
  que valia para todo o RO Lobby.)
- **RN-04** — Editar pode mudar o servidor, com as mesmas validações. Um personagem com
  candidatura pendente, vaga ocupada num lobby aberto (como dono ou membro) ou pedido de
  troca pendente não pode ter o servidor alterado, como já vale para função e nível
  (amplia a RN-25 da `candidatura-lobby`). A mensagem no perfil passa a ser "Esse
  personagem está num lobby aberto. Saia ou cancele antes de mudar servidor, nível ou
  função."
- **RN-05** — A carteirinha do perfil mostra o servidor junto da classe e do nível.

### Lobby
- **RN-06** — O lobby fica no servidor do personagem do dono no momento da criação. Não há
  campo de servidor no formulário: a criação e a prévia do card mostram o servidor do
  personagem escolhido e mudam quando ele muda.
- **RN-07** — O servidor fica gravado no lobby e não muda: nem na edição, nem quando o
  personagem do dono é trocado, editado ou excluído.
- **RN-08** — O card da Home, o destaque e o detalhe do lobby mostram o servidor.

### Grupo no mesmo servidor
- **RN-09** — Só personagem do servidor do lobby se candidata a ele (amplia a RN-01 da
  `candidatura-lobby`). A API recusa com o código `wrong_server`.
- **RN-10** — Na troca de personagem, o personagem novo precisa ser do servidor do lobby:
  vale para a troca do dono (RN-19 da `candidatura-lobby`) e para o pedido de troca do
  membro (RN-20), no pedido e no aceite (RN-23). A API recusa com `wrong_server`.
- **RN-11** — Na escolha de personagem para candidatar ou pedir troca, os personagens de
  outro servidor aparecem desabilitados, com o texto "Outro servidor (<nome>)".

### Home
- **RN-12** — A Home ganha o filtro "Servidor", com as opções Todos, Freya e Nidhogg. Ele
  abre em Todos e não guarda a escolha. Com Freya ou Nidhogg, só aparecem os lobbies
  daquele servidor. Ele se combina com os outros filtros com E (amplia a RN-12 da
  `home-local`).
- **RN-13** — Cada opção do filtro mostra quantos lobbies do dia escolhido são daquele
  servidor, sem considerar os outros filtros (Todos mostra o total do dia).
- **RN-14** — O seletor de dias, as contagens de "Vaga para" e de "Faixa de horário", o
  destaque e os estados vazios respeitam o servidor escolhido como respeitam os outros
  filtros hoje. O botão "Limpar filtros" volta o servidor para Todos.

### Dados existentes
- **RN-15** — Na migração, todo personagem e todo lobby que já existe recebe Freya. Se a
  migração encontrar dois nicks iguais (o que a RN-05 antiga já impedia), ela falha sem
  mudar nada.

## 5. Critérios de aceite

### US-01 — Servidor do personagem
```gherkin
CA-01.1 — Catálogo de servidores  [US-01, RN-01]
Given um visitante sem sessão
When a API recebe GET /servers
Then a resposta é 200 com Freya e Nidhogg, nesta ordem, e os códigos freya e nidhogg

CA-01.2 — Cadastrar com servidor  [US-01, RN-02, RN-05]
Given um Usuário logado em /perfil
When ele adiciona um personagem escolhendo Nidhogg
Then o personagem é salvo com o servidor Nidhogg
  And a carteirinha mostra "Nidhogg" junto da classe e do nível

CA-01.3 — Servidor obrigatório  [US-01, RN-02]
Given um Usuário logado
When ele cria um personagem sem servidor, ou com um código fora do catálogo
Then a API responde 400 com o erro no campo server
  And o formulário mostra "Escolha o servidor" junto do campo

CA-01.4 — Mesmo nick em servidores diferentes  [US-01, RN-03]
Given um personagem "Brasa" em Freya de outro Usuário
When o Usuário cria "brasa" em Nidhogg
Then o personagem é criado

CA-01.5 — Nick repetido no mesmo servidor  [US-01, RN-03]
Given um personagem "Brasa" em Freya de qualquer Usuário
When alguém cria ou edita outro personagem para "brasa" em Freya
Then a API responde 400 com o erro taken no campo nick
  And o formulário mostra "Esse nick já está em uso nesse servidor"

CA-01.6 — Mudar de servidor e colidir com um nick  [US-01, RN-03, RN-04]
Given "Brasa" em Freya e outro "Brasa" em Nidhogg
When o dono do de Freya edita o servidor para Nidhogg
Then a API responde 400 com o erro taken no campo nick
  And o personagem continua em Freya

CA-01.7 — Editar o servidor  [US-01, RN-04]
Given um personagem em Freya sem candidatura, vaga ou pedido de troca ativos
When o dono edita o servidor para Nidhogg
Then o personagem passa a ser de Nidhogg

CA-01.8 — Servidor travado por vínculo ativo  [US-01, RN-04]
Given um personagem com candidatura pendente, vaga num lobby aberto ou pedido de troca pendente
When o dono tenta mudar o servidor
Then a API recusa a edição como já recusa para função e nível
  And o perfil mostra "Esse personagem está num lobby aberto. Saia ou cancele antes de mudar servidor, nível ou função."
  And o servidor não muda
```

### US-02 — Servidor do lobby
```gherkin
CA-02.1 — Lobby herda o servidor do dono  [US-02, RN-06]
Given um Usuário com o personagem "Brasa" em Nidhogg
When ele cria um lobby com "Brasa"
Then o lobby fica em Nidhogg

CA-02.2 — Prévia acompanha o personagem  [US-02, RN-06]
Given um Usuário com um personagem em Freya e outro em Nidhogg, em /lobbies/novo
When ele troca o personagem escolhido de um para o outro
Then o servidor mostrado na criação e na prévia do card muda junto

CA-02.3 — Servidor não muda depois  [US-02, RN-07]
Given um lobby aberto em Nidhogg
When o dono edita o lobby, ou troca de personagem, ou o personagem do dono é excluído depois que o lobby inicia
Then o lobby continua em Nidhogg

CA-02.4 — Servidor na tela  [US-02, RN-08]
Given um lobby aberto em Freya
When alguém vê a Home e o detalhe do lobby
Then o card, o destaque (quando for ele) e o detalhe mostram "Freya"
```

### US-03 — Grupo no mesmo servidor
```gherkin
CA-03.1 — Candidatura de outro servidor  [US-03, RN-09]
Given um lobby aberto em Freya
  And um candidato com personagem em Nidhogg
When ele se candidata com esse personagem
Then a API responde 409 com o código wrong_server
  And nenhuma candidatura é criada

CA-03.2 — Candidatura do mesmo servidor  [US-03, RN-09]
Given um lobby aberto em Freya
  And um candidato com personagem em Freya que cumpre as outras regras
When ele se candidata
Then a candidatura fica pendente

CA-03.3 — Troca do dono para outro servidor  [US-03, RN-10]
Given um lobby aberto em Freya
When o dono tenta trocar para um personagem dele em Nidhogg
Then a API responde 409 com o código wrong_server
  And o personagem do dono não muda

CA-03.4 — Pedido de troca para outro servidor  [US-03, RN-10]
Given um membro de um lobby aberto em Freya
When ele pede troca para um personagem em Nidhogg
Then a API responde 409 com o código wrong_server
  And nenhum pedido é criado

CA-03.5 — Aceite de troca com servidor mudado  [US-03, RN-04, RN-10]
Given um pedido de troca pendente para um personagem em Freya, num lobby em Freya
When o dono aceita o pedido
Then o aceite segue as regras de sempre, porque o personagem do pedido não pode mudar de servidor enquanto o pedido está pendente (RN-04)

CA-03.6 — Personagens de outro servidor desabilitados  [US-03, RN-11]
Given um Usuário com um personagem em Freya e outro em Nidhogg
  And um lobby aberto em Freya
When ele abre a escolha de personagem para se candidatar ou pedir troca
Then o personagem de Nidhogg aparece desabilitado com "Outro servidor (Nidhogg)"
  And o de Freya pode ser escolhido
```

### US-04 — Filtro da Home
```gherkin
CA-04.1 — Filtro começa em Todos  [US-04, RN-12, RN-13]
Given 3 lobbies hoje em Freya e 2 em Nidhogg
When alguém abre a Home
Then o filtro "Servidor" está em Todos
  And as opções mostram Todos 5, Freya 3 e Nidhogg 2
  And a lista mostra os 5 lobbies

CA-04.2 — Filtrar por servidor  [US-04, RN-12]
Given 3 lobbies hoje em Freya e 2 em Nidhogg
When alguém escolhe Nidhogg
Then a lista mostra só os 2 de Nidhogg

CA-04.3 — Combina com os outros filtros  [US-04, RN-12, RN-14]
Given lobbies de Freya e de Nidhogg hoje, com vagas de Tank só em um de Freya
When alguém escolhe Nidhogg e marca "Vaga para: Tank"
Then a lista fica vazia, com o estado vazio de filtros da home-local
  And o destaque não mostra lobby de Freya

CA-04.4 — Limpar filtros  [US-04, RN-14]
Given o filtro "Servidor" em Freya
When alguém usa "Limpar filtros"
Then o filtro volta para Todos

CA-04.5 — Não guarda a escolha  [US-04, RN-12]
Given o filtro "Servidor" em Freya
When a pessoa recarrega a Home
Then o filtro está em Todos
```

### Migração
```gherkin
CA-05.1 — Dados existentes recebem Freya  [RN-15]
Given personagens e lobbies criados antes desta feature
When a migração roda
Then todos ficam em Freya
  And a migração volta atrás (down) sem perder os outros dados
```

## 6. Casos de borda
- Dono exclui o personagem depois que o lobby inicia → o lobby continua com o servidor
  gravado [RN-07 / CA-02.3].
- Candidato muda o servidor do personagem depois que a candidatura deixa de estar ativa
  (recusada, retirada, lobby iniciado) → permitido, porque a trava só vale com vínculo
  ativo [RN-04].
- Usuário só tem personagens de outro servidor → todos aparecem desabilitados na escolha
  e não há como candidatar [RN-11].
- Filtro em um servidor sem lobbies no dia → estado vazio de filtros [RN-14].
- Duas pessoas criando "Brasa" em Freya ao mesmo tempo → só uma consegue; a outra recebe
  `taken` [RN-03].

## 7. Requisitos não funcionais
- **RNF-01** — O índice único do nick passa a ser (servidor, nick sem diferenciar
  maiúsculas), no banco, como última defesa da RN-03.
- **RNF-02** — O banco tem CHECK do servidor nas tabelas de personagem e de lobby, como os
  CHECK de função.
- **RNF-03** — Todo teste novo cita o ID do critério; os testes antigos que mudarem de
  comportamento (nick único) passam a citar também o critério desta spec.

## 8. Fora de escopo
- Lembrar o filtro de servidor ou abrir no servidor do principal — escolhido "Todos, sem
  memória" na entrevista; pode virar ajuste depois.
- Servidor em "Minhas candidaturas" — não escolhido na entrevista.
- Escolher o servidor do lobby por campo próprio — o servidor vem do personagem do dono.
- Mudar o servidor de um lobby existente.
- Servidores de outras regiões ou novos servidores — o catálogo tem só os dois atuais;
  um novo entra por mudança no catálogo.

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- Grafia → Freya e Nidhogg.
- Nick único → por servidor.
- Servidor do lobby → vem do personagem do dono, gravado no lobby e fixo.
- Dados existentes → Freya.
- Edição do servidor do personagem → permitida sem vínculo ativo (mesma trava de função e
  nível).
- Filtro da Home → Todos, com contagem, sem lembrar a escolha.
- Personagens de outro servidor na candidatura e na troca → desabilitados com motivo.
- Onde o servidor aparece → carteirinha do perfil, card da Home e destaque, detalhe do
  lobby.
