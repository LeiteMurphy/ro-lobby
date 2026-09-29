# Spec — Fundação

- Feature: `fundacao` · Nível: G · Status: Aprovada
- Notion: ainda não publicado
- Última revisão: 2026-09-29 — aprovada pelo usuário, com as premissas da seção 9 confirmadas

## 1. Contexto
O repositório ainda não tem código. Antes da primeira feature de produto (login com
Discord), ele precisa de um esqueleto que já siga os ADRs aprovados:

- backend em Go com `net/http` (ADR-02);
- web em SvelteKit (ADR-01);
- monorepo (ADR-03);
- banco com `pgx`, `sqlc` e `goose` (ADR-04);
- contrato OpenAPI (ADR-05);
- testes e CI bloqueando o merge.

A entrega mínima que prova que tudo se conecta é um health check no backend e uma
página provisória no web que mostra o status da API.

## 2. Atores
| Ator | Descrição | Permissões principais |
|---|---|---|
| Desenvolvedor | Quem clona e trabalha no repositório | Subir o ambiente local, rodar testes, lint e geração de código |
| CI | GitHub Actions | Rodar lint, testes, build e as verificações de código gerado em cada PR |
| Visitante | Quem abre o web | Ver a página de status |

## 3. User Stories
| ID | Prioridade | História |
|---|---|---|
| US-01 | P1 | Como desenvolvedor, quero subir o ambiente local seguindo um passo a passo documentado, para começar a trabalhar sem adivinhar dependências. |
| US-02 | P1 | Como desenvolvedor, quero um health check no backend que reflita o estado do banco, para saber se a API está saudável. |
| US-03 | P1 | Como desenvolvedor, quero aplicar e reverter migrações versionadas, para evoluir o banco de forma controlada. |
| US-04 | P1 | Como desenvolvedor, quero que o contrato OpenAPI gere os tipos do Go e do TypeScript, para que backend e web não se desencontrem. |
| US-05 | P1 | Como visitante, quero ver uma página com o status da API, para confirmar que o web conversa com o backend. |
| US-06 | P1 | Como desenvolvedor, quero que a CI rode lint, testes e build em cada PR e bloqueie o merge se algo falhar, para manter a qualidade da `main`. |
| US-07 | P2 | Como desenvolvedor, quero um teste ponta a ponta da página de status, para provar o fluxo completo navegador → web → API → banco. |
| US-08 | P2 | Como desenvolvedor, quero ver o relatório de cobertura em cada PR, para acompanhar o que está testado. |

## 4. Regras
Nesta feature, as regras são invariantes técnicos, não de domínio.

### Configuração e ambiente
- **RN-01** — Toda configuração vem de variáveis de ambiente. Nenhum segredo é
  versionado.
- **RN-02** — O `.env.example` é versionado e lista todas as variáveis que backend,
  web e Compose leem, com valores fictícios.
- **RN-03** — O `docker-compose.yml` sobe o PostgreSQL local, com os dados guardados
  num volume nomeado. O backend e o web rodam direto na máquina.
- **RN-04** — As versões do Go e do Node são fixadas no repositório (`go.mod`,
  `.nvmrc` ou `engines`), e a CI usa as mesmas versões.
- **RN-05** — O README tem o passo a passo do setup local no Windows, incluindo a
  instalação do Go e do Docker Desktop, e os comandos para subir o banco, migrar,
  rodar o backend, rodar o web e rodar os testes.

### Backend e banco
- **RN-06** — `GET /healthz` responde 200 com `{"status":"ok","database":"ok"}`
  quando o banco responde a um ping em até 2 segundos.
- **RN-07** — Quando o banco não responde em até 2 segundos, `GET /healthz`
  responde 503 com `{"status":"degraded","database":"unavailable"}`.
- **RN-08** — A conexão com o banco usa o fuso UTC.
- **RN-09** — As migrações são arquivos SQL versionados no `backend/`, cada um com
  `up` e `down`. Aplicar e reverter todas as migrações num banco vazio termina sem
  erro.
- **RN-10** — Os testes unitários rodam sem banco. Os testes de integração rodam
  contra um PostgreSQL real, com um comando próprio.

### Contrato e código gerado
- **RN-11** — O `openapi.yaml` é a fonte do contrato. A rota `/healthz` está
  descrita nele.
- **RN-12** — O código gerado (`oapi-codegen`, `openapi-typescript`, `sqlc`) é
  versionado. A CI gera de novo e falha se houver qualquer diferença com o que foi
  commitado.
- **RN-13** — O web só acessa a API por meio dos tipos gerados do contrato.

### Web
- **RN-14** — O web usa o adaptador de servidor (Node), não o estático, porque o
  preview no Discord exige renderização no servidor (ADR-01).
- **RN-15** — A página de status é renderizada no servidor e mostra "API online"
  quando `/healthz` responde 200, "API com problema" quando responde 503 e "API
  indisponível" quando não responde.
- **RN-16** — O TypeScript roda em modo estrito.

### Qualidade e CI
- **RN-17** — A CI roda em todo PR para a `main`:
  - o job do backend quando `backend/**` muda;
  - o job do web quando `web/**` muda;
  - os dois quando o `openapi.yaml` ou a própria CI mudam.
- **RN-18** — O job do backend roda `gofmt`, `golangci-lint`, os testes unitários,
  os de integração (com Postgres como container de serviço), o build e a verificação
  de código gerado.
- **RN-19** — O job do web roda ESLint, Prettier (verificação), `svelte-check`,
  Vitest, o build e a verificação de código gerado.
- **RN-20** — Qualquer falha em lint, teste, build ou código gerado deixa o job
  vermelho, e o merge fica bloqueado.
- **RN-21** — A CI publica o relatório de cobertura do backend e do web, sem
  percentual mínimo.
- **RN-22** — O teste ponta a ponta (Playwright) abre a página de status com o
  backend e o banco no ar e confere o texto "API online".

## 5. Critérios de aceite

### US-01 — Ambiente local
```gherkin
CA-01.1 — Setup do zero  [US-01, RN-03, RN-05]
Given uma máquina Windows com Go, Node e Docker Desktop instalados conforme o README
  And o repositório recém-clonado
When o desenvolvedor segue o passo a passo do README
Then o PostgreSQL sobe pelo Compose
  And as migrações são aplicadas
  And o backend responde em GET /healthz com 200
  And o web abre a página de status mostrando "API online"

CA-01.2 — Variáveis documentadas  [US-01, RN-01, RN-02]
Given o código do backend, do web e o docker-compose.yml
When se levantam todas as variáveis de ambiente que eles leem
Then todas aparecem no .env.example com valor fictício
  And nenhum arquivo versionado contém senha ou token real

CA-01.3 — Dados persistem  [US-01, RN-03]
Given o banco local com migrações aplicadas
When o desenvolvedor para e sobe o Compose de novo
Then as migrações continuam aplicadas

CA-01.4 — Versões fixadas  [US-01, RN-04]
Given o repositório
When se compara a versão do Go e do Node declaradas no repositório com as usadas na CI
Then as versões são as mesmas
```

### US-02 — Health check
```gherkin
CA-02.1 — Banco disponível  [US-02, RN-06]
Given o backend rodando e o banco respondendo
When se faz GET /healthz
Then a resposta é 200 com {"status":"ok","database":"ok"}

CA-02.2 — Banco fora do ar  [US-02, RN-07]
Given o backend rodando e o banco parado
When se faz GET /healthz
Then a resposta é 503 com {"status":"degraded","database":"unavailable"}
  And a resposta chega em até 3 segundos

CA-02.3 — Banco lento  [US-02, RN-07]
Given um banco que demora mais de 2 segundos para responder ao ping
When se faz GET /healthz
Then a resposta é 503 com {"status":"degraded","database":"unavailable"}

CA-02.4 — Método não permitido  [US-02, RN-11]
Given o backend rodando
When se faz POST /healthz
Then a resposta é 405

CA-02.5 — Fuso da conexão  [US-02, RN-08]
Given o backend conectado ao banco
When se consulta o fuso da sessão (SHOW TimeZone)
Then o valor é UTC
```

### US-03 — Migrações
```gherkin
CA-03.1 — Aplicar em banco vazio  [US-03, RN-09]
Given um banco PostgreSQL vazio
When o desenvolvedor aplica todas as migrações
Then o comando termina sem erro
  And a versão registrada é a da última migração

CA-03.2 — Reverter tudo  [US-03, RN-09]
Given um banco com todas as migrações aplicadas
When o desenvolvedor reverte todas as migrações
Then o comando termina sem erro
  And o banco volta a não ter tabelas do projeto

CA-03.3 — Toda migração tem down  [US-03, RN-09]
Given os arquivos de migração do backend
When se verifica cada arquivo
Then todos têm uma seção up e uma seção down
```

### US-04 — Contrato e código gerado
```gherkin
CA-04.1 — Rota descrita no contrato  [US-04, RN-11]
Given o openapi.yaml
When se lista as rotas descritas
Then GET /healthz está descrita com as respostas 200 e 503 e seus corpos

CA-04.2 — Código gerado em dia  [US-04, RN-12]
Given um PR em que o código gerado bate com o openapi.yaml e as queries do sqlc
When a CI roda a verificação de código gerado
Then a verificação passa

CA-04.3 — Código gerado desatualizado  [US-04, RN-12, RN-20]
Given um PR que muda o openapi.yaml sem gerar o código de novo
When a CI roda a verificação de código gerado
Then a verificação falha
  And o merge fica bloqueado

CA-04.4 — Web usa os tipos gerados  [US-04, RN-13]
Given o código do web que chama a API
When se inspeciona o tipo da resposta de /healthz usado no web
Then ele vem do arquivo gerado a partir do openapi.yaml
```

### US-05 — Página de status
```gherkin
CA-05.1 — API saudável  [US-05, RN-15]
Given /healthz respondendo 200
When o visitante abre a página de status
Then a página mostra "API online"

CA-05.2 — API com problema  [US-05, RN-15]
Given /healthz respondendo 503
When o visitante abre a página de status
Then a página mostra "API com problema"

CA-05.3 — API fora do ar  [US-05, RN-15]
Given o backend parado
When o visitante abre a página de status
Then a página mostra "API indisponível"
  And a página não mostra erro nem stack trace

CA-05.4 — Renderização no servidor  [US-05, RN-14, RN-15]
Given /healthz respondendo 200
When se busca a página de status sem executar JavaScript
Then o HTML recebido já contém "API online"
```

### US-06 — CI
```gherkin
CA-06.1 — PR só de backend  [US-06, RN-17]
Given um PR que muda apenas arquivos em backend/
When a CI roda
Then o job do backend roda
  And o job do web não roda

CA-06.2 — PR só de web  [US-06, RN-17]
Given um PR que muda apenas arquivos em web/
When a CI roda
Then o job do web roda
  And o job do backend não roda

CA-06.3 — PR que muda o contrato  [US-06, RN-17]
Given um PR que muda o openapi.yaml
When a CI roda
Then os jobs do backend e do web rodam

CA-06.4 — Lint quebrado  [US-06, RN-18, RN-19, RN-20]
Given um PR com código fora do padrão do gofmt, golangci-lint, ESLint ou Prettier
When a CI roda
Then o job correspondente falha

CA-06.5 — Teste quebrado  [US-06, RN-18, RN-19, RN-20]
Given um PR com um teste falhando no backend ou no web
When a CI roda
Then o job correspondente falha

CA-06.6 — Testes de integração na CI  [US-06, RN-10, RN-18]
Given um PR que muda backend/
When a CI roda
Then os testes de integração rodam contra um PostgreSQL em container de serviço

CA-06.7 — Testes unitários sem banco  [US-06, RN-10]
Given nenhum banco disponível
When o desenvolvedor roda só os testes unitários do backend
Then todos rodam sem tentar conectar ao banco

CA-06.8 — TypeScript estrito  [US-06, RN-16, RN-19]
Given um PR com código TypeScript que só compila fora do modo estrito
When a CI roda
Then o job do web falha
```

### US-07 — Ponta a ponta
```gherkin
CA-07.1 — Fluxo completo  [US-07, RN-22]
Given o banco, o backend e o web no ar
When o teste Playwright abre a página de status
Then encontra o texto "API online"
```

### US-08 — Cobertura
```gherkin
CA-08.1 — Relatório publicado  [US-08, RN-21]
Given um PR que muda backend/ ou web/
When a CI termina
Then o relatório de cobertura do job correspondente fica disponível no PR
  And nenhum percentual mínimo faz o job falhar
```

## 6. Casos de borda
- Banco parado → `/healthz` responde 503 rápido, sem travar a requisição
  [RN-07 / CA-02.2]
- Backend parado → a página de status mostra "API indisponível", sem erro na tela
  [RN-15 / CA-05.3]
- Contrato alterado sem gerar o código → a CI falha [RN-12 / CA-04.3]
- PR que muda só documentação (`docs/`, `.specs/`) → nenhum job de código roda
  [RN-17]
- Porta ocupada na máquina local → as portas vêm do `.env`, e o README diz como
  trocar [RN-01, RN-05]

## 7. Requisitos não funcionais
- **RNF-01** — Datas em UTC no banco (CLAUDE.md), garantidas pela RN-08.
- **RNF-02** — A CI de um PR de uma pasta só termina em até 10 minutos.
- **RNF-03** — Todo teste cita o ID do critério de aceite que ele cobre (CLAUDE.md).

## 8. Fora de escopo
- Login com Discord — é a próxima feature de produto, com spec própria.
- Tabelas de domínio (Usuário, Personagem, Lobby, Candidatura) — entram nas specs
  de cada feature.
- A Home de verdade — vem do Claude Design, sem o design system da organização.
- Hospedagem e deploy — ADR pendente.
- Mobile — ADR pendente.
- Percentual mínimo de cobertura.
- Imagens Docker do backend e do web — o Compose só sobe o banco (RN-03).
- Hooks de pre-commit.

## 9. Perguntas em aberto
- Nenhuma. As premissas assumidas foram confirmadas na aprovação (ver seção 10).

## 10. Decisões tomadas na entrevista
- Login com Discord na fundação → não, fica para depois.
- Postgres local → Docker Compose (exige instalar o Docker Desktop).
- CI no GitHub Actions por pasta, bloqueando o merge → sim.
- ADR-04 (`pgx` + `sqlc` + `goose`) → aprovado.
- ADR-05 (OpenAPI + `oapi-codegen` + `openapi-typescript`) → aprovado.
- Testes: unitários e de integração com Postgres real no backend; Vitest e um
  ponta a ponta com Playwright no web → sim.
- Entrega mínima: `/healthz` com o estado do banco e uma página de status provisória
  → sim.
- Lint: `gofmt`, `golangci-lint`, ESLint, Prettier, `svelte-check`, TypeScript estrito
  → sim.
- Configuração por variáveis de ambiente, com `.env.example` → sim.
- Cobertura publicada, sem percentual mínimo → sim.
- Código gerado → versionado, com verificação na CI (RN-12).
- Compose → só o Postgres; backend e web rodam direto na máquina (RN-03).
- Portas padrão → API 8080, web 5173 e Postgres 5432, configuráveis pelo `.env`.
- Proteção da `main` → configurada pelo usuário no GitHub; a fundação só documenta o passo.
