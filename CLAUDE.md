# RO Lobby (nome provisório)

Lobby para jogadores de Ragnarok Online montarem grupos para instâncias difíceis.
É um projeto de fã, sem vínculo com a Gravity.

## Estado atual
Prontos: fundação (backend Go com `GET /healthz`, PostgreSQL, migrações, OpenAPI e CI), a
Home do Claude Design em `/` com os lobbies reais, login com Discord, o perfil com os
personagens em `/perfil`, lobbies (criar, detalhe, editar, cancelar) e a candidatura
completa (candidatar, decidir, sair, remover com bloqueio e trocas de personagem, com
"Minhas candidaturas" em `/candidaturas`) e o compartilhamento de lobby (convite copiado e
preview Open Graph). A pilha completa roda com
`docker compose --profile app up` (ADR-06). Nada em andamento; os próximos passos ficam no
Notion. O setup local e os comandos estão no `README.md`. As decisões de arquitetura ficam
em `docs/adr/`.

## Produto
- Problema: é difícil montar grupos para instâncias difíceis. Um grupo tem até 12
  pessoas, normalmente 1 Tank, de 2 a 4 Suportes e o resto Dano.
- O jogador cadastra personagens (nick, classe, nível, função, link externo, horários,
  instâncias de interesse) e cria grupos ou se candidata a eles. O criador aceita ou
  recusa com justificativa.
- A home mostra os "grupos para hoje" com horário e composição por função. Os outros
  dias ficam num calendário.
- Futuro, fora do foco atual: guildas recebendo candidaturas, comp builder visual,
  reputação de jogador, link de doação.

## Domínio
Usuário, Personagem, Disponibilidade, Instância (catálogo), Lobby, Candidatura.
Futuro: Guilda, Membro de guilda.

## Tecnologia

### Decidido
- Servidor de RO em foco: Ragnarok LATAM (GnJoy Americas).
- Backend em Go.
- Web primeiro, mobile depois.
- Login com Discord (OAuth).
- PostgreSQL, com datas em UTC no banco e convertidas na exibição.
- Framework web: SvelteKit ([ADR-01](docs/adr/0001-framework-web-sveltekit.md)).
- Framework HTTP do backend: biblioteca padrão `net/http` ([ADR-02](docs/adr/0002-framework-http-net-http.md)).
- Estrutura do repositório: monorepo com `backend/`, `web/`, `docs/adr/` e `.specs/` ([ADR-03](docs/adr/0003-estrutura-monorepo.md)).
- Acesso ao banco: `pgx` + `sqlc`, migrações com `goose` ([ADR-04](docs/adr/0004-acesso-banco-pgx-sqlc-goose.md)).
- Contrato da API: `openapi.yaml` como fonte, com `oapi-codegen` (Go) e `openapi-typescript` (web) ([ADR-05](docs/adr/0005-contrato-api-openapi.md)).
- Hospedagem: local, com o perfil `app` do Docker Compose ([ADR-06](docs/adr/0006-hospedagem-local.md)). Para testes fechados, o PC do usuário publica em `rolobby.com.br` pelo Cloudflare Tunnel ([ADR-08](docs/adr/0008-publicacao-tunel-cloudflare.md)).

### Pendente de ADR (fase de Design)
- Mobile: Kotlin ou Flutter.
- Hospedagem definitiva (nuvem ou VPS), na revisão do ADR-08.

### Em aberto
- Nada no momento.

Itens pendentes ou em aberto não são decisões. Não escreva código que dependa deles
antes de haver um ADR.

## Como trabalhamos
- Toda feature nova passa pela skill spec-driven-dev. As specs ficam em
  `.specs/<feature>/` e são versionadas. As regras de negócio ficam na spec, não aqui.
- O backlog fica no Notion, na página "RO Lobby" (Épicos → User Stories → Tasks).
- Uma branch por feature (`feature/<slug>`).
- Merge: PRs só de documentação (spec, ADR) entram com a revisão do usuário. PRs com
  código só entram com o relatório do validador APROVADO.
- Commits em Conventional Commits citando os IDs da spec,
  ex.: `feat: candidatura a lobby [US-02, CA-02.1]`.
- Qualidade faz parte do objetivo, já que o projeto é portfólio de quem vem de QA.
  Toda regra de domínio tem teste, e cada teste cita o ID do critério de aceite.
- Idioma: documentação, specs e commits em português. Código e identificadores em inglês.
- Não usar logos, artes nem sprites da Gravity. Nomes de classe e de instância do jogo
  podem ser usados: seguem a nomenclatura do bROWiki
  ([Classes](https://browiki.org/wiki/Classes),
  [Instâncias](https://browiki.org/wiki/Inst%C3%A2ncias)), que acompanha os dados do
  Ragnarok LATAM.
