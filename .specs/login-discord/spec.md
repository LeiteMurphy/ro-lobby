# Spec — Login com Discord

- Feature: `login-discord` · Nível: G · Status: Aprovada
- ADR: [ADR-07](../../docs/adr/0007-login-discord-sessao.md) (aceito)
- Última revisão: 2026-10-03 — sem avatar do Discord: a barra mostra a inicial do nome, para não
  carregar nada de fora do servidor (decisão do usuário)

## 1. Contexto
O RO Lobby ainda não tem usuários. Personagens, lobbies e candidaturas pertencem a um
Usuário, e o login decidido é com o Discord (CLAUDE.md). Esta feature cria o Usuário a
partir da conta do Discord, abre e encerra a sessão e mostra quem está logado na barra
superior. É a base de todas as features de produto seguintes.

A arquitetura atual expõe só o web; o navegador não fala com a API (RN-03 da
`home-local`). O web recebe o retorno do Discord e a API faz a troca do código, guarda o
Usuário e a Sessão (ADR-07).

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Visitante | Quem abre o web sem sessão | Ver a Home, entrar com o Discord |
| Usuário | Quem entrou com o Discord | Tudo do visitante, ver o próprio perfil na barra, sair |
| Discord | Provedor de identidade (OAuth2) | Autenticar a pessoa e devolver o código |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como visitante, quero entrar com a minha conta do Discord, para usar o RO Lobby sem criar senha. |
| US-02 | P1 | Como usuário, quero ver meu nome do Discord na barra superior, para saber que estou logado. |
| US-03 | P1 | Como usuário, quero sair, para encerrar a sessão neste navegador. |
| US-04 | P1 | Como visitante, quero uma mensagem clara quando o login falhar, para saber que posso tentar de novo. |
| US-05 | P2 | Como visitante, quero voltar para a página onde estava depois de entrar, para não perder o que fazia. |
| US-06 | P1 | Como desenvolvedor, quero que a API identifique o usuário da sessão, para as próximas features saberem quem está agindo. |

## 4. Regras

### OAuth com o Discord
- **RN-01** — O login usa o fluxo *authorization code* do Discord com o escopo
  `identify` e nenhum outro.
- **RN-02** — Cada início de login gera um `state` aleatório de pelo menos 128 bits,
  guardado num cookie `HttpOnly` de vida curta (10 minutos). O retorno só é aceito se o
  `state` recebido for igual ao do cookie, e o cookie é apagado no retorno, aceito ou não.
- **RN-03** — O Client Secret existe só no ambiente da API. Ele não aparece no web, no
  navegador, em log nem em arquivo versionado.
- **RN-04** — A troca do código pelo perfil acontece na API: ela troca o `code` no
  endpoint de token do Discord, lê `GET /users/@me` e descarta os tokens do Discord. Os
  tokens do Discord não são guardados.

### Usuário
- **RN-05** — O Usuário é identificado pelo ID do Discord, que é único. O primeiro login
  cria o Usuário; os seguintes atualizam o nome de usuário e o nome de exibição.
  Mudar o nome no Discord não cria outro Usuário.
- **RN-06** — O Usuário guarda só: ID do Discord, nome de usuário, nome de exibição (pode
  ser vazio), data de criação e data do último login, em UTC. O avatar do Discord não é
  guardado nem exibido.

### Sessão
- **RN-07** — Cada login bem-sucedido cria uma Sessão com um token aleatório de 256 bits.
  O banco guarda só o hash SHA-256 do token; o token em si vai só para o cookie.
- **RN-08** — O cookie de sessão é `HttpOnly`, `SameSite=Lax`, `Path=/` e `Secure` quando
  o web não está em `localhost`.
- **RN-09** — A Sessão vale 30 dias a partir do último uso. Uma requisição com a Sessão
  válida empurra o vencimento; o banco só é atualizado se o último uso registrado tiver
  mais de 1 hora, para não escrever a cada requisição.
- **RN-10** — Sessão vencida, desconhecida ou de um Usuário apagado vale como visitante:
  o web mostra "Entrar com Discord" e apaga o cookie.
- **RN-11** — Sair apaga a Sessão no banco e o cookie no navegador. Outras sessões do
  mesmo Usuário (outros navegadores) continuam valendo.

### Navegação e erros
- **RN-12** — Depois de entrar, a pessoa volta para a página de onde saiu. Só caminhos
  relativos do próprio site são aceitos (começando com `/` e não com `//`); qualquer
  outro valor volta para a Home.
- **RN-13** — Se a pessoa cancelar no Discord, o `state` não bater, o código for
  inválido ou o Discord não responder em até 5 segundos, ela volta para a Home com a
  mensagem "Não foi possível entrar com o Discord. Tente de novo." e nenhum Usuário ou
  Sessão é criado.
- **RN-14** — O botão "Entrar com Discord" da barra superior passa a funcionar. Os outros
  controles "Disponível em breve" continuam desabilitados (RN-18 da `home-local`).
- **RN-15** — Logado, a barra superior mostra a inicial e o nome de exibição (ou o nome de
  usuário, se não houver nome de exibição), com um menu que tem "Sair". A inicial fica num
  quadrado no estilo do design system; nenhuma imagem é carregada de fora do servidor
  (RNF-02 da `home-local`).

### API
- **RN-16** — O contrato (`openapi.yaml`) descreve as rotas novas: a troca do código, o
  usuário da sessão e o encerramento da sessão. A rota do usuário da sessão responde 401
  sem sessão válida.
- **RN-17** — As URLs do Discord (autorização e API) vêm de variáveis de ambiente, com os
  endereços oficiais como padrão. Os testes de integração e o ponta a ponta usam um
  Discord falso local, sem rede externa.

## 5. Critérios de aceite

### US-01 — Entrar
```gherkin
CA-01.1 — Primeiro login cria o Usuário  [US-01, RN-01, RN-04, RN-05, RN-07]
Given um visitante sem Usuário no RO Lobby
When ele clica em "Entrar com Discord" e autoriza no Discord
Then um Usuário é criado com o ID, o nome de usuário e o nome de exibição do Discord
  And uma Sessão é criada e o navegador recebe o cookie de sessão
  And a barra superior mostra o nome de exibição dele

CA-01.2 — Login seguinte atualiza o Usuário  [US-01, RN-05]
Given um Usuário que já entrou antes com o nome "Grimbold"
  And ele mudou o nome de exibição no Discord para "Grimbold, o Sábio"
When ele entra de novo
Then continua existindo um único Usuário com aquele ID do Discord
  And o nome de exibição passa a ser "Grimbold, o Sábio"

CA-01.3 — Só o escopo identify  [US-01, RN-01]
Given um visitante na Home
When ele clica em "Entrar com Discord"
Then o navegador vai para a URL de autorização do Discord com scope=identify e response_type=code
  And a URL tem um parâmetro state

CA-01.4 — O banco não guarda o token  [US-01, RN-04, RN-07]
Given um login concluído
When se inspecionam as tabelas de Usuário e Sessão
Then nenhuma coluna guarda o token de sessão em texto, nem tokens do Discord

CA-01.5 — Client Secret fora do web  [US-01, RN-03]
Given o build do web e as respostas do servidor do web
When se procura o valor do Client Secret
Then ele não aparece em nenhum arquivo do build nem em nenhuma resposta
```

### US-02 — Perfil na barra
```gherkin
CA-02.1 — Inicial e nome  [US-02, RN-15]
Given um Usuário logado com o nome de exibição "Grimbold"
When ele abre a Home
Then a barra superior mostra "G" e "Grimbold"
  And não mostra "Entrar com Discord"

CA-02.2 — Sem nome de exibição  [US-02, RN-15]
Given um Usuário sem nome de exibição, com o nome de usuário "mirai.exe"
When ele abre a Home
Then a barra mostra "M" e "mirai.exe"

CA-02.4 — Nada carregado de fora  [US-02, RN-06, RN-15]
Given um Usuário logado com avatar no Discord
When ele abre a Home
Then nenhuma requisição sai do navegador para fora do servidor do web
  And o banco não guarda o avatar

CA-02.3 — Renderização no servidor  [US-02, RN-15]
Given um Usuário logado
When se busca a Home sem executar JavaScript, com o cookie de sessão
Then o HTML já traz o nome dele na barra superior
```

### US-03 — Sair
```gherkin
CA-03.1 — Sair encerra a sessão  [US-03, RN-11]
Given um Usuário logado
When ele escolhe "Sair" no menu da barra
Then a Sessão é apagada no banco e o cookie é removido
  And a barra volta a mostrar "Entrar com Discord"

CA-03.2 — Outras sessões continuam  [US-03, RN-11]
Given um Usuário logado em dois navegadores
When ele sai em um deles
Then a sessão do outro navegador continua válida
```

### US-04 — Erros
```gherkin
CA-04.1 — Cancelou no Discord  [US-04, RN-13]
Given um visitante que clicou em "Entrar com Discord"
When ele cancela a autorização no Discord
Then ele volta para a Home com a mensagem "Não foi possível entrar com o Discord. Tente de novo."
  And nenhum Usuário nem Sessão é criado

CA-04.2 — state diferente  [US-04, RN-02, RN-13]
Given um retorno do Discord com um state diferente do cookie
When o web recebe o retorno
Then ele mostra a mensagem de erro e não chama a API
  And nenhum Usuário nem Sessão é criado

CA-04.3 — Retorno sem cookie de state  [US-04, RN-02, RN-13]
Given um retorno do Discord sem o cookie de state (por exemplo, depois de 10 minutos)
When o web recebe o retorno
Then ele mostra a mensagem de erro
  And nenhum Usuário nem Sessão é criado

CA-04.4 — Código inválido  [US-04, RN-13]
Given um retorno com um code que o Discord recusa
When a API tenta trocar o código
Then a pessoa vê a mensagem de erro
  And nenhum Usuário nem Sessão é criado

CA-04.5 — Discord fora do ar  [US-04, RN-13]
Given o Discord não responde à troca do código
When a pessoa volta do Discord
Then em até 6 segundos ela vê a mensagem de erro
  And nenhum Usuário nem Sessão é criado

CA-04.6 — state de uso único  [US-04, RN-02]
Given um retorno aceito
When o mesmo retorno (mesmo code e state) é repetido
Then o segundo é recusado com a mensagem de erro
```

### US-05 — Voltar para onde estava
```gherkin
CA-05.1 — Volta para a página de origem  [US-05, RN-12]
Given um visitante em /status
When ele entra com o Discord
Then ele volta para /status, logado

CA-05.2 — Endereço externo vira Home  [US-05, RN-12]
Given um pedido de login com destino "https://exemplo.com" ou "//exemplo.com"
When o login termina
Then a pessoa vai para a Home
```

### US-06 — Usuário da sessão na API
```gherkin
CA-06.1 — Sessão válida  [US-06, RN-16]
Given uma Sessão válida
When o web pede o usuário da sessão à API
Then a API responde 200 com o ID do Usuário, o nome de exibição e o nome de usuário

CA-06.2 — Sem sessão  [US-06, RN-10, RN-16]
Given um token de sessão desconhecido, vencido ou ausente
When o web pede o usuário da sessão à API
Then a API responde 401
  And o web trata a pessoa como visitante e apaga o cookie

CA-06.3 — Vencimento a partir do último uso  [US-06, RN-09]
Given uma Sessão usada pela última vez há 29 dias
When ela é usada de novo
Then ela continua válida e passa a vencer 30 dias depois deste uso

CA-06.4 — Sessão vencida  [US-06, RN-09, RN-10]
Given uma Sessão usada pela última vez há 31 dias
When ela é usada
Then a API responde 401

CA-06.5 — Cookie de sessão  [US-06, RN-08]
Given um login concluído em localhost
When se inspeciona o cookie de sessão
Then ele é HttpOnly, SameSite=Lax e Path=/
  And com o web fora de localhost ele também é Secure
```

## 6. Casos de borda
- Duas abas iniciando o login ao mesmo tempo: o segundo `state` substitui o primeiro, e
  o retorno da primeira aba falha com a mensagem de erro [RN-02, RN-13]
- Usuário apaga a conta do Discord: a Sessão continua até vencer; o próximo login com
  outra conta cria outro Usuário [RN-05, RN-09]
- Relógio: datas da Sessão e do Usuário em UTC no banco (CLAUDE.md) [RN-06, RN-09]

## 7. Requisitos não funcionais
- **RNF-01** — Segurança: `state` contra CSRF, token de sessão só como hash, cookie
  `HttpOnly`, Client Secret só na API, nenhum segredo em log ou arquivo versionado.
- **RNF-02** — Acessibilidade: o menu do usuário funciona pelo teclado, com foco visível
  (mesmo critério do RNF-01 da `home-local`).
- **RNF-05** — Nenhum recurso carregado de fora do servidor do web, mantendo o RNF-02 da
  `home-local`.
- **RNF-03** — Todo teste cita o ID do critério de aceite (CLAUDE.md).
- **RNF-04** — Os testes não dependem da rede: o Discord é falso nos testes de integração,
  no ponta a ponta e na CI.

## 8. Fora de escopo
- Cadastro de personagens, criação de lobby e candidatura.
- Revisão da hospedagem (ADR-06) e o `redirect_uri` de produção.
- "Sair de todos os dispositivos", painel de sessões e exclusão de conta.
- E-mail, servidores (guilds) e qualquer escopo além de `identify`.
- Login por outro provedor ou por senha.
- Foto de perfil. Se entrar no futuro, o servidor baixa o avatar e serve do próprio
  domínio, sem o navegador falar com a CDN do Discord.

## 9. Perguntas em aberto
- Nenhuma.

## 10. Decisões tomadas na entrevista
- OAuth → o web recebe o retorno e a API troca o código e guarda Usuário e Sessão (ADR-07).
- Sessão → token aleatório em cookie `HttpOnly`/`SameSite=Lax`, hash no banco, 30 dias
  renovados pelo uso.
- Dados do Discord → só `identify`: ID, nome de usuário e nome de exibição. Sem avatar,
  sem e-mail e sem guardar tokens do Discord.
- Perfil → atualizado a cada login, chave pelo ID do Discord.
- Depois de entrar → volta para a página de origem; barra com a inicial e o nome.
- Avatar → não usado: carregar da CDN do Discord fere o RNF-02 da `home-local`. A barra
  mostra a inicial do nome.
- Sair → menu no avatar, encerra só a sessão atual.
- Erros → volta para a Home com mensagem curta, sem criar Usuário.
- Fora de escopo → personagem e revisão da hospedagem.
