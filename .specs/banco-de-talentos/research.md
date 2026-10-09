# Research — Banco de talentos

## Achados
- A spec `personagens` deixou "horários disponíveis e instâncias de interesse" para uma
  feature seguinte (seções 8 e 10). O CLAUDE.md lista a Disponibilidade no domínio.
- `characters` (migração 00003) guarda nick, classe, nível, função, retrato, link e
  principal. Não há tabela de disponibilidade.
- `users` guarda `username` e `global_name` do Discord. É daí que sai o `@username`.
- O conflito de horário já existe: a janela de 2 h (RN-10 da `lobbies`) e a query
  `CharacterInOpenLobby` (dono, candidatura pendente ou aceita, pedido de troca).
- O catálogo de instâncias está em Go (`internal/catalog`) e é servido por
  `GET /instances`.
- Todo horário da interface é de Brasília (RN-09 da `home-local`). O Brasil não tem
  horário de verão desde 2019, então a hora de parede de Brasília é sempre UTC-3.
- Nada no Notion sobre o tema.

## Arquivos-chave
- `backend/migrations/`, `backend/queries/`, `backend/internal/characters`,
  `backend/internal/lobbies`, `openapi.yaml`
- `web/src/routes/perfil`, `web/src/routes/lobbies/[id]`, `web/src/routes/lobbies/novo`,
  `web/src/lib/lobbies/components/LobbyForm.svelte`

## Classificação
**G**: tabela nova, regra de afinidade que cruza personagem, lobby e candidaturas, três
telas (perfil, detalhe e criação do lobby, catálogo novo) e uma decisão técnica sobre como
guardar horário recorrente.
