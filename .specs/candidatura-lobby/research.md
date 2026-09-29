# Research — Candidatura a lobby

- Feature: `candidatura-lobby`
- Data: 2026-09-29

## Achados
- O repositório está na fase de fundação e não tem código (só `CLAUDE.md` e `README.md`).
  Não há padrões de arquitetura, nomes ou testes para reaproveitar.
- Não existe nenhuma spec anterior em `.specs/`. Esta é a primeira.
- A página "RO Lobby" no Notion existe, mas está vazia (sem Épicos, US ou Tasks).
- O `CLAUDE.md` define o domínio: Usuário, Personagem, Disponibilidade, Instância,
  Lobby, Candidatura. Um grupo tem até 12 pessoas, normalmente 1 Tank, de 2 a 4
  Suportes e o resto Dano. O criador aceita ou recusa com justificativa.
- Decidido: Go no backend, PostgreSQL com datas em UTC, login com Discord.
  Framework HTTP, framework web e estrutura do repositório ainda dependem de ADR.
  Esta spec não depende de nenhum desses itens.

## Dependências
- **Lobby** e **Personagem** ainda não têm spec. A candidatura depende dos dois
  (vagas, horário, criador, função e classe do personagem). Esta spec vai assumir
  o mínimo que precisa deles e deixar isso explícito como premissa, para não
  decidir regras que pertencem a outras features.

## Classificação: G (Grande)
- Entidade nova (Candidatura) com ciclo de vida e transições de estado.
- Várias regras de domínio: vagas por função, lotação, conflito de horário,
  permissões do criador, justificativa obrigatória.
- Depende de entidades que ainda não foram especificadas.

Escopo combinado com o usuário: parar na spec aprovada, sem Design, Tasks nem código.
