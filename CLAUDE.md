# RO Lobby (nome provisório)

Lobby para jogadores de Ragnarok Online montarem grupos para instâncias difíceis.
É um projeto de fã, sem vínculo com a Gravity.

## Estado atual
Fase de fundação: o repositório ainda não tem código. A estrutura de pastas será
definida por ADR na fase de Design.

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

### Pendente de ADR (fase de Design)
- Framework HTTP do backend.
- Framework web: Next.js ou SvelteKit.
- Mobile: Kotlin ou Flutter.
- Hospedagem.
- Estrutura do repositório.

### Em aberto
- Nada no momento.

Itens pendentes ou em aberto não são decisões. Não escreva código que dependa deles
antes de haver um ADR.

## Como trabalhamos
- Toda feature nova passa pela skill spec-driven-dev. As specs ficam em
  `.specs/<feature>/` e são versionadas. As regras de negócio ficam na spec, não aqui.
- O backlog fica no Notion, na página "RO Lobby" (Épicos → User Stories → Tasks).
- Uma branch por feature (`feature/<slug>`). Merge só com relatório do validador APROVADO.
- Commits em Conventional Commits citando os IDs da spec,
  ex.: `feat: candidatura a lobby [US-02, CA-02.1]`.
- Qualidade faz parte do objetivo, já que o projeto é portfólio de quem vem de QA.
  Toda regra de domínio tem teste, e cada teste cita o ID do critério de aceite.
- Idioma: documentação, specs e commits em português. Código e identificadores em inglês.
- Não usar logos, artes nem nomes oficiais da Gravity.
