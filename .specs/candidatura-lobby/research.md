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

## Retomada — 2026-10-06
A spec foi aprovada em 2026-09-29, antes de qualquer código. Desde então entraram
login, personagens e lobbies, e algumas premissas mudaram:
- **P-01 (horário):** o lobby não tem fim; o conflito usa a janela fixa de 2 h a partir
  do início (RN-10 e D-03 da `lobbies`). As RN-11 e CA-02.10 a CA-02.12 precisam usar
  essa janela.
- **Nível mínimo:** o lobby já tem nível mínimo (RN-07 da `lobbies`). A seção "Fora de
  escopo" ainda diz que requisito de nível fica para depois.
- **Travas do personagem:** a RN-21 da `lobbies` já trava função e exclusão do
  personagem dono de lobby aberto. O usuário decidiu (2026-10-06) que um personagem
  registrado num grupo também não pode editar o nível.
- **Expiração:** o estado iniciado do lobby vem do horário (RN-13 da `lobbies`), sem
  job agendado. As pendências podem expirar pela mesma regra, na leitura.
- **Tela:** o usuário pediu (2026-10-06) que, no detalhe do lobby, clicar no card de uma
  pessoa mostre os detalhes dela no painel da direita.
- **Código pronto para reuso:** a contagem `occupied` do lobby (D-01 da `lobbies`), a
  trava no Usuário (D-05), a consulta de conflito, os diálogos com motivo (cancelar
  lobby) e o Discord falso com várias contas no ponta a ponta.
- **Tamanho:** 8 US e mais de 60 critérios. Pode ser entregue em duas partes.
